// Copyright 2018 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package swap

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/swap"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/swap/chain"
	"github.com/ethersphere/swarm/swap/int256"
)

// ErrInvalidChequeSignature indicates the signature on the cheque was invalid
var ErrInvalidChequeSignature = errors.New("invalid cheque signature")

// ErrSkipDeposit indicates that the user has specified an amount to deposit (swap-deposit-amount) but also indicated that depositing should be skipped (swap-skip-deposit)
var ErrSkipDeposit = errors.New("swap-deposit-amount non-zero, but swap-skip-deposit true")

// Swap represents the Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	store             state.Store                // store is needed in order to keep balances and cheques across sessions
	peers             map[enode.ID]*Peer         // map of all swap Peers
	peersLock         sync.RWMutex               // lock for peers map
	owner             *Owner                     // contract access
	backend           chain.Backend              // the backend (blockchain) used
	chainID           uint64                     // id of the chain the backend is connected to
	params            *Params                    // economic and operational parameters
	contract          contract.Contract          // reference to the smart contract
	chequebookFactory contract.SimpleSwapFactory // the chequebook factory used
	honeyPriceOracle  HoneyOracle                // oracle which resolves the price of honey (in Wei)
	cashoutProcessor  *CashoutProcessor          // processor for cashing out
	logger            Logger                     //Swap Logger
}

// Owner encapsulates information related to accessing the contract
type Owner struct {
	address    common.Address    // owner address
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
}

// Params encapsulates economic and operational parameters
type Params struct {
	BaseAddrs           *network.BzzAddr // this node's base address
	LogPath             string           // optional audit log path
	LogLevel            int              // optional indicates audit filter level of swap log messages
	PaymentThreshold    *int256.Uint256  // honey amount at which a payment is triggered
	DisconnectThreshold *int256.Uint256  // honey amount at which a peer disconnects
}

// newSwapInstance is a swap constructor function without integrity checks
func newSwapInstance(stateStore state.Store, owner *Owner, backend chain.Backend, chainID uint64, params *Params, chequebookFactory contract.SimpleSwapFactory, logger Logger) *Swap {
	return &Swap{
		store:             stateStore,
		peers:             make(map[enode.ID]*Peer),
		backend:           backend,
		owner:             owner,
		params:            params,
		chequebookFactory: chequebookFactory,
		honeyPriceOracle:  NewHoneyPriceOracle(),
		chainID:           chainID,
		cashoutProcessor:  newCashoutProcessor(backend, owner.privateKey),
		logger:            logger,
	}
}

// New prepares and creates all fields to create a swap instance:
// - sets up a SWAP database;
// - verifies whether the disconnect threshold is higher than the payment threshold;
// - connects to the blockchain backend;
// - verifies that we have not connected SWAP before on a different blockchain backend;
// - starts the chequebook; creates the swap instance
func New(dbPath string, prvkey *ecdsa.PrivateKey, backendURL string, params *Params, chequebookAddressFlag common.Address, skipDepositFlag bool, depositAmountFlag uint64, factoryAddress common.Address) (swap *Swap, err error) {
	// swap log for auditing purposes
	swapLogger := newSwapLogger(params.LogPath, params.LogLevel, params.BaseAddrs)
	// verify that backendURL is not empty
	if backendURL == "" {
		return nil, errors.New("no backend URL given")
	}
	// verify that depositAmountFlag and skipDeposit are not conflicting
	if depositAmountFlag > 0 && skipDepositFlag {
		return nil, ErrSkipDeposit
	}
	swapLogger.Info(InitAction, "connecting to SWAP API", "url", backendURL)
	// initialize the balances store
	var stateStore state.Store
	if stateStore, err = state.NewDBStore(filepath.Join(dbPath, "swap.db")); err != nil {
		return nil, fmt.Errorf("initializing statestore: %w", err)
	}
	if params.DisconnectThreshold.Cmp(params.PaymentThreshold) < 1 {
		return nil, fmt.Errorf("disconnect threshold lower or at payment threshold. DisconnectThreshold: %v, PaymentThreshold: %v", params.DisconnectThreshold, params.PaymentThreshold)
	}
	// connect to the backend
	backend, err := ethclient.Dial(backendURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to Ethereum API, url %s: %w", backendURL, err)
	}
	// get the chainID of the backend
	var chainID *big.Int
	if chainID, err = backend.ChainID(context.TODO()); err != nil {
		return nil, fmt.Errorf("retrieving chainID from backendURL: %v", err)
	}
	// verify that we have not used SWAP before on a different chainID
	if err := checkChainID(chainID.Uint64(), stateStore, swapLogger); err != nil {
		return nil, err
	}
	swapLogger.Info(InitAction, "Using backend network ID", "ID", chainID.Uint64())

	// create the owner of SWAP
	owner := createOwner(prvkey)

	// initialize the factory
	factory, err := createFactory(factoryAddress, chainID, backend, swapLogger)
	if err != nil {
		return nil, err
	}

	// create the swap instance
	swap = newSwapInstance(
		stateStore,
		owner,
		backend,
		chainID.Uint64(),
		params,
		factory,
		swapLogger,
	)
	// start the chequebook
	if swap.contract, err = swap.StartChequebook(chequebookAddressFlag); err != nil {
		return nil, err
	}

	// deposit money in the chequebook if desired
	if !skipDepositFlag {
		// prompt the user for a depositAmount
		var toDeposit = big.NewInt(int64(depositAmountFlag))
		if toDeposit.Cmp(&big.Int{}) == 0 {
			toDeposit, err = swap.promptDepositAmount()
			if err != nil {
				return nil, err
			}
		}
		// deposit if toDeposit is bigger than zero
		if toDeposit.Cmp(&big.Int{}) > 0 {
			if err := swap.Deposit(context.TODO(), toDeposit); err != nil {
				return nil, err
			}
		} else {
			swapLogger.Info(InitAction, "Skipping deposit")
		}
	}

	return swap, nil
}

const (
	balancePrefix          = "balance_"
	sentChequePrefix       = "sent_cheque_"
	receivedChequePrefix   = "received_cheque_"
	pendingChequePrefix    = "pending_cheque_"
	connectedChequebookKey = "connected_chequebook"
	connectedBlockchainKey = "connected_blockchain"
)

// createFactory determines the factory address and returns and error if no factory address has been specified or is unknown for the network
func createFactory(factoryAddress common.Address, chainID *big.Int, backend chain.Backend, logger Logger) (factory swap.SimpleSwapFactory, err error) {
	if (factoryAddress == common.Address{}) {
		if factoryAddress, err = contract.FactoryAddressForNetwork(chainID.Uint64()); err != nil {
			return nil, err
		}
	}
	logger.Info(InitAction, "Using chequebook factory", "address", factoryAddress)
	// instantiate an object representing the factory and verify it's bytecode
	factory, err = contract.FactoryAt(factoryAddress, backend)
	if err != nil {
		return nil, err
	}
	if err := factory.VerifySelf(); err != nil {
		return nil, err
	}
	return factory, nil
}

// checkChainID verifies whether we have initialized SWAP before and ensures that we are on the same backendNetworkID if this is the case
// the logger is being injected from swap New function, this let's us continue logging in the same swap log context
func checkChainID(currentChainID uint64, s state.Store, logger Logger) (err error) {
	var connectedBlockchain uint64
	err = s.Get(connectedBlockchainKey, &connectedBlockchain)
	// error reading from database
	if err != nil && err != state.ErrNotFound {
		return fmt.Errorf("querying usedBeforeAtNetwork from statestore: %w", err)
	}
	// initialized before, but on a different chainID
	if err != state.ErrNotFound && connectedBlockchain != currentChainID {
		return fmt.Errorf("statestore previously used on different backend network. Used before on network: %d, Attempting to connect on network %d", connectedBlockchain, currentChainID)
	}
	if err == state.ErrNotFound {
		logger.Info(InitAction, "First time connected to SWAP. Storing chain ID", "ID", currentChainID)
		return s.Put(connectedBlockchainKey, currentChainID)
	}
	return nil
}

// returns the store key for retrieving a peer's balance
func balanceKey(peer enode.ID) string {
	return balancePrefix + peer.String()
}

// returns the store key for retrieving a peer's last sent cheque
func sentChequeKey(peer enode.ID) string {
	return sentChequePrefix + peer.String()
}

// returns the store key for retrieving a peer's last received cheque
func receivedChequeKey(peer enode.ID) string {
	return receivedChequePrefix + peer.String()
}

func pendingChequeKey(peer enode.ID) string {
	return pendingChequePrefix + peer.String()
}

func keyToID(key string, prefix string) enode.ID {
	return enode.HexID(key[len(prefix):])
}

// createOwner assigns keys and addresses
func createOwner(prvkey *ecdsa.PrivateKey) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		address:    crypto.PubkeyToAddress(*pubkey),
		privateKey: prvkey,
		publicKey:  pubkey,
	}
}

// modifyBalanceOk checks that the amount would not result in crossing the disconnection threshold
func (s *Swap) modifyBalanceOk(amount *int256.Int256, swapPeer *Peer) (err error) {
	// check if balance with peer is over the disconnect threshold and if the message would increase the existing debt
	balance := swapPeer.getBalance()
	if balance.Cmp(s.params.DisconnectThreshold) >= 0 && amount.Cmp(int256.Int256From(0)) > 0 {
		return fmt.Errorf("balance for peer %s is over the disconnect threshold %v and cannot incur more debt, disconnecting", swapPeer.ID().String(), s.params.DisconnectThreshold)
	}

	return nil
}

// Check is called as a *dry run* before applying the actual accounting to an operation.
// It only checks that performing a given accounting operation would not incur in an error.
// If it returns no error, this signals to the caller that the operation is safe
func (s *Swap) Check(amount int64, peer *protocols.Peer) (err error) {
	swapPeer := s.getPeer(peer.ID())
	if swapPeer == nil {
		return fmt.Errorf("peer %s not a swap enabled peer", peer.ID().String())
	}

	swapPeer.lock.Lock()
	defer swapPeer.lock.Unlock()
	// currently this is the only real check needed:
	return s.modifyBalanceOk(int256.Int256From(amount), swapPeer)
}

// Add is the (sole) accounting function
// Swap implements the protocols.Balance interface
func (s *Swap) Add(amount int64, peer *protocols.Peer) (err error) {
	swapPeer := s.getPeer(peer.ID())
	if swapPeer == nil {
		return fmt.Errorf("peer %s not a swap enabled peer", peer.ID().String())
	}

	swapPeer.lock.Lock()
	defer swapPeer.lock.Unlock()
	// we should probably check here again:
	if err = s.modifyBalanceOk(int256.Int256From(amount), swapPeer); err != nil {
		return err
	}

	if err = swapPeer.updateBalance(int256.Int256From(amount)); err != nil {
		return err
	}

	return s.checkPaymentThresholdAndSendCheque(swapPeer)
}

// checkPaymentThresholdAndSendCheque checks if balance with peer crosses the payment threshold and attempts to send a cheque if so
// It is the peer with a negative balance who sends a cheque, thus we check
// that the balance is *below* the threshold
// the caller is expected to hold swapPeer.lock
func (s *Swap) checkPaymentThresholdAndSendCheque(swapPeer *Peer) error {
	balance := swapPeer.getBalance()
	thresholdValue := s.params.PaymentThreshold.Value()
	threshold, err := int256.NewInt256(thresholdValue)
	if err != nil {
		return err
	}

	negativeThreshold, err := new(int256.Int256).Mul(int256.Int256From(-1), threshold)
	if err != nil {
		return err
	}

	if balance.Cmp(negativeThreshold) < 1 {
		swapPeer.logger.Info(SendChequeAction, "balance for peer went over the payment threshold, sending cheque", "payment threshold", s.params.PaymentThreshold)
		return swapPeer.sendCheque()
	}
	return nil
}

// handleMsg is for handling messages when receiving messages
func (s *Swap) handleMsg(p *Peer) func(ctx context.Context, msg interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		switch msg := msg.(type) {
		case *EmitChequeMsg:
			return s.handleEmitChequeMsg(ctx, p, msg)
		case *ConfirmChequeMsg:
			return s.handleConfirmChequeMsg(ctx, p, msg)
		}
		return nil
	}
}

var defaultCashCheque = cashCheque

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a debitor
func (s *Swap) handleEmitChequeMsg(ctx context.Context, p *Peer, msg *EmitChequeMsg) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	cheque := msg.Cheque
	p.logger.Info(HandleChequeAction, "received cheque from peer", "honey", cheque.Honey)

	if p.getLastReceivedCheque() != nil && cheque.Equal(p.getLastReceivedCheque()) {
		p.logger.Warn(HandleChequeAction, "cheque sent by peer has already been received in the past", "cumulativePayout", cheque.CumulativePayout)
		return p.Send(ctx, &ConfirmChequeMsg{
			Cheque: cheque,
		})
	}

	_, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		return protocols.Break(fmt.Errorf("processing and verifying received cheque: %w", err))
	}

	p.logger.Debug(HandleChequeAction, "processed and verified received cheque", "beneficiary", cheque.Beneficiary, "cumulative payout", cheque.CumulativePayout)

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	honeyAmount := int256.Int256From(int64(cheque.Honey))
	honey, err := new(int256.Int256).Mul(int256.Int256From(-1), honeyAmount)
	if err != nil {
		return protocols.Break(err)
	}

	err = p.updateBalance(honey)
	if err != nil {
		return protocols.Break(fmt.Errorf("updating balance: %w", err))
	}

	metrics.GetOrRegisterCounter("swap.cheques.received.num", nil).Inc(1)
	metrics.GetOrRegisterCounter("swap.cheques.received.honey", nil).Inc(int64(cheque.Honey))

	err = p.Send(ctx, &ConfirmChequeMsg{
		Cheque: cheque,
	})
	if err != nil {
		return protocols.Break(err)
	}

	expectedPayout, transactionCosts, err := s.cashoutProcessor.estimatePayout(context.TODO(), cheque)
	if err != nil {
		return protocols.Break(err)
	}

	costsMultiplier := int256.Uint256From(2)
	costThreshold, err := new(int256.Uint256).Mul(transactionCosts, costsMultiplier)
	if err != nil {
		return err
	}

	// do a payout transaction if we get 2 times the gas costs
	if expectedPayout.Cmp(costThreshold) == 1 {
		go defaultCashCheque(s, cheque)
	}

	return nil
}

func (s *Swap) handleConfirmChequeMsg(ctx context.Context, p *Peer, msg *ConfirmChequeMsg) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	cheque := msg.Cheque

	if p.getPendingCheque() == nil {
		return fmt.Errorf("ignoring confirm msg, no pending cheque, confirm message cheque %s", cheque)
	}

	if !cheque.Equal(p.getPendingCheque()) {
		return fmt.Errorf("ignoring confirm msg, unexpected cheque, confirm message cheque %s, expected %s", cheque, p.getPendingCheque())
	}

	batch := new(state.StoreBatch)
	err := batch.Put(sentChequeKey(p.ID()), cheque)
	if err != nil {
		return protocols.Break(fmt.Errorf("encoding cheque failed: %w", err))
	}

	err = batch.Put(pendingChequeKey(p.ID()), nil)
	if err != nil {
		return protocols.Break(fmt.Errorf("encoding pending cheque failed: %w", err))
	}

	err = s.store.WriteBatch(batch)
	if err != nil {
		return protocols.Break(fmt.Errorf("could not write cheque to database: %w", err))
	}

	p.lastSentCheque = cheque
	p.pendingCheque = nil

	return nil
}

// cashCheque should be called async as it blocks until the transaction(s) are mined
// The function cashes the cheque by sending it to the blockchain
func cashCheque(s *Swap, cheque *Cheque) {
	err := s.cashoutProcessor.cashCheque(context.Background(), &CashoutRequest{
		Cheque:      *cheque,
		Destination: s.GetParams().ContractAddress,
		Logger:      s.logger,
	})

	if err != nil {
		metrics.GetOrRegisterCounter("swap.cheques.cashed.errors", nil).Inc(1)
		s.logger.Error(CashChequeAction, "cashing cheque:", err)
	}
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
// the caller is expected to hold p.lock
func (s *Swap) processAndVerifyCheque(cheque *Cheque, p *Peer) (*int256.Uint256, error) {
	if err := cheque.verifyChequeProperties(p, s.owner.address); err != nil {
		return nil, err
	}

	lastCheque := p.getLastReceivedCheque()

	// TODO: there should probably be a lock here?
	expectedAmount, err := s.honeyPriceOracle.GetPrice(cheque.Honey)
	if err != nil {
		return nil, err
	}

	actualAmount, err := cheque.verifyChequeAgainstLast(lastCheque, int256.Uint256From(expectedAmount))
	if err != nil {
		return nil, err
	}

	chequeHoney := new(big.Int).SetUint64(cheque.Honey)
	honey, err := int256.NewInt256(*chequeHoney)
	if err != nil {
		return nil, err
	}

	// calculate tentative new balance after cheque is processed
	newBalance, err := new(int256.Int256).Sub(p.getBalance(), honey)
	if err != nil {
		return nil, err
	}

	debtTolerance := big.NewInt(-int64(ChequeDebtTolerance))
	tolerance, err := int256.NewInt256(*debtTolerance)
	if err != nil {
		return nil, err
	}

	// check if this new balance would put creditor into debt
	if newBalance.Cmp(tolerance) == -1 {
		return nil, fmt.Errorf("received cheque would result in balance %v which exceeds tolerance %d and would cause debt", newBalance, ChequeDebtTolerance)
	}

	if err := p.setLastReceivedCheque(cheque); err != nil {
		p.logger.Error(HandleChequeAction, "error while saving last received cheque", "err", err.Error())
		// TODO: what do we do here? Related issue: https://github.com/ethersphere/swarm/issues/1515
	}

	return actualAmount, nil
}

// loadLastReceivedCheque loads the last received cheque for the peer from the store
// and returns nil when there never was a cheque saved
func (s *Swap) loadLastReceivedCheque(p enode.ID) (cheque *Cheque, err error) {
	err = s.store.Get(receivedChequeKey(p), &cheque)
	if err == state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cheque, nil
}

// loadLastSentCheque loads the last sent cheque for the peer from the store
// and returns nil when there never was a cheque saved
func (s *Swap) loadLastSentCheque(p enode.ID) (cheque *Cheque, err error) {
	err = s.store.Get(sentChequeKey(p), &cheque)
	if err == state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cheque, nil
}

// loadPendingCheque loads the current pending cheque for the peer from the store
// and returns nil when there never was a pending cheque saved
func (s *Swap) loadPendingCheque(p enode.ID) (cheque *Cheque, err error) {
	err = s.store.Get(pendingChequeKey(p), &cheque)
	if err == state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cheque, nil
}

// loadBalance loads the current balance for the peer from the store
// and returns 0 if there was no prior balance saved
func (s *Swap) loadBalance(p enode.ID) (balance *int256.Int256, err error) {
	err = s.store.Get(balanceKey(p), &balance)
	if err == state.ErrNotFound {
		return int256.Int256From(0), nil
	}
	if err != nil {
		return int256.Int256From(0), err
	}
	return balance, nil
}

// saveLastReceivedCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastReceivedCheque(p enode.ID, cheque *Cheque) error {
	return s.store.Put(receivedChequeKey(p), cheque)
}

// saveLastSentCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastSentCheque(p enode.ID, cheque *Cheque) error {
	return s.store.Put(sentChequeKey(p), cheque)
}

// savePendingCheque saves cheque as the last pending cheque for peer
func (s *Swap) savePendingCheque(p enode.ID, cheque *Cheque) error {
	return s.store.Put(pendingChequeKey(p), cheque)
}

// saveBalance saves balance as the current balance for peer
func (s *Swap) saveBalance(p enode.ID, balance *int256.Int256) error {
	return s.store.Put(balanceKey(p), balance)
}

// Close cleans up swap
func (s *Swap) Close() error {
	return s.store.Close()
}

// GetParams returns contract parameters (Bin, ABI, contractAddress) from the contract
func (s *Swap) GetParams() *contract.Params {
	return s.contract.ContractParams()
}

// getContractOwner retrieve the owner of the chequebook at address from the blockchain
func (s *Swap) getContractOwner(ctx context.Context, address common.Address) (common.Address, error) {
	contr, err := contract.InstanceAt(address, s.backend)
	if err != nil {
		return common.Address{}, err
	}

	return contr.Issuer(nil)
}

// promptDepositAmount blocks and asks the user how much ERC20 he wants to deposit
func (s *Swap) promptDepositAmount() (*big.Int, error) {
	// retrieve available balance
	availableBalance, err := s.AvailableBalance()
	if err != nil {
		return nil, err
	}
	balance, err := s.contract.BalanceAtTokenContract(nil, s.owner.address)
	if err != nil {
		return nil, err
	}
	// log available balance and ERC20 balance
	s.logger.Info(InitAction, "Balance information", "chequebook available balance", availableBalance, "ERC20 balance", balance)
	promptMessage := fmt.Sprintf("Please provide the amount in HONEY which will deposited to your chequebook (0 for skipping deposit): ")
	// need to prompt user for deposit amount
	prompter := console.Stdin
	// ask user for input
	input, err := prompter.PromptInput(promptMessage)
	if err != nil {
		return &big.Int{}, err
	}
	// check input
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		// maybe we should provide a fallback here? A bad input results in stopping the boot
		return &big.Int{}, fmt.Errorf("conversion failed while reading user input: %w", err)
	}
	return big.NewInt(val), nil
}

// StartChequebook starts the chequebook, taking into account the chequebookAddress passed in by the user and the chequebook addresses saved on the node's database
func (s *Swap) StartChequebook(chequebookAddrFlag common.Address) (contract contract.Contract, err error) {
	previouslyUsedChequebook, err := s.loadChequebook()
	// error reading from disk
	if err != nil && err != state.ErrNotFound {
		return nil, fmt.Errorf("reading previously used chequebook: %w", err)
	}
	// read from state, but provided flag is not the same
	if err == nil && (chequebookAddrFlag != common.Address{} && chequebookAddrFlag != previouslyUsedChequebook) {
		return nil, fmt.Errorf("attempting to connect to provided chequebook, but different chequebook used before")
	}
	// nothing written to state disk before, no flag provided: deploying new chequebook
	if err == state.ErrNotFound && chequebookAddrFlag == (common.Address{}) {

		if contract, err = s.Deploy(context.Background()); err != nil {
			return nil, err
		}
		if err := s.saveChequebook(contract.ContractParams().ContractAddress); err != nil {
			return nil, err
		}
		s.logger.Info(InitAction, "Deployed chequebook", "contract address", contract.ContractParams().ContractAddress.Hex(), "owner", s.owner.address)
		return contract, nil
	}
	// first time connecting with a chequebookAddress passed in
	if chequebookAddrFlag != (common.Address{}) {
		return s.bindToContractAt(chequebookAddrFlag)
	}
	// reconnecting with contract read from statestore
	return s.bindToContractAt(previouslyUsedChequebook)
}

// BindToContractAt binds to an instance of an already existing chequebook contract at address
func (s *Swap) bindToContractAt(address common.Address) (contract.Contract, error) {
	// validate whether address is a chequebook
	if err := s.chequebookFactory.VerifyContract(address); err != nil {
		return nil, fmt.Errorf("contract validation for %v: %w", address.Hex(), err)
	}
	s.logger.Info(InitAction, "bound to chequebook", "chequebookAddr", address)
	// get the instance
	return contract.InstanceAt(address, s.backend)
}

// Deploy deploys the Swap contract
func (s *Swap) Deploy(ctx context.Context) (contract.Contract, error) {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	opts.Context = ctx
	s.logger.Info(DeployChequebookAction, "Deploying new swap", "owner", opts.From.Hex())
	chequebook, err := s.chequebookFactory.DeploySimpleSwap(opts, s.owner.address, big.NewInt(int64(defaultHarddepositTimeoutDuration)))
	if err != nil {
		return nil, fmt.Errorf("failed to deploy chequebook: %w", err)
	}
	return chequebook, nil
}

// Deposit deposits ERC20 into the chequebook contract
func (s *Swap) Deposit(ctx context.Context, amount *big.Int) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	opts.Context = ctx
	s.logger.Info(InitAction, "Depositing ERC20 into chequebook", "amount", amount)
	rec, err := s.contract.Deposit(opts, amount)
	if err != nil {
		return err
	}
	s.logger.Info(InitAction, "Deposited ERC20 into chequebook", "amount", amount, "transaction", rec.TxHash)
	return nil
}

func (s *Swap) loadChequebook() (common.Address, error) {
	var chequebook common.Address
	err := s.store.Get(connectedChequebookKey, &chequebook)
	return chequebook, err
}

func (s *Swap) saveChequebook(chequebook common.Address) error {
	return s.store.Put(connectedChequebookKey, chequebook)
}
