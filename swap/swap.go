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
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/swap"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

// ErrInvalidChequeSignature indicates the signature on the cheque was invalid
var ErrInvalidChequeSignature = errors.New("invalid cheque signature")

var auditLog log.Logger // logger for Swap related messages and audit trail

// swapLogLevel indicates filter level of log messages
const swapLogLevel = 3

// Swap represents the Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	store            state.Store        // store is needed in order to keep balances and cheques across sessions
	peers            map[enode.ID]*Peer // map of all swap Peers
	peersLock        sync.RWMutex       // lock for peers map
	backend          contract.Backend   // the backend (blockchain) used
	owner            *Owner             // contract access
	params           *Params            // economic and operational parameters
	contract         swap.Contract      // reference to the smart contract
	honeyPriceOracle HoneyOracle        // oracle which resolves the price of honey (in Wei)
}

// Owner encapsulates information related to accessing the contract
type Owner struct {
	address    common.Address    // owner address
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
}

// Params encapsulates param
type Params struct {
	LogPath             string // optional audit log path
	PaymentThreshold    int64  // honey amount at which a payment is triggered
	DisconnectThreshold int64  // honey amount at which a peer disconnects
}

// newLogger returns a new logger
func newLogger(logpath string) log.Logger {
	swapLogger := log.New("swaplog", "*")
	lh := log.Root().GetHandler()

	if logpath == "" {
		swapLogger.SetHandler(lh)
		return swapLogger
	}

	rfh, err := swapRotatingFileHandler(logpath)

	if err != nil {
		log.Warn("RotatingFileHandler was not initialized", "logdir", logpath, "err", err)
		//sets a fallback logger, it will use the swarm logger.
		swapLogger.SetHandler(lh)
		return swapLogger
	}

	//Filters messages with the correct logLevel for swap
	rfh = log.LvlFilterHandler(log.Lvl(swapLogLevel), rfh)

	//Dispatches the logs to the default swarm log and also the filtered swap file logger.
	swapLogger.SetHandler(log.MultiHandler(lh, rfh))

	return swapLogger
}

// swapRotatingFileHandler returns a RotatingFileHandler this will split the logs into multiple files.
// the files are split based on the limit parameter expressed in bytes
func swapRotatingFileHandler(logdir string) (log.Handler, error) {
	return log.RotatingFileHandler(
		logdir,
		262144,
		log.JSONFormatOrderedEx(false, true),
	)
}

// new - swap constructor without integrity check
func new(stateStore state.Store, owner *Owner, backend contract.Backend, params *Params) *Swap {
	return &Swap{
		store:            stateStore,
		peers:            make(map[enode.ID]*Peer),
		backend:          backend,
		owner:            owner,
		params:           params,
		honeyPriceOracle: NewHoneyPriceOracle(),
	}
}

// New prepares and creates all fields to create a swap instance:
// - sets up a SWAP database;
// - verifies whether the disconnect threshold is higher than the payment threshold;
// - connects to the blockchain backend;
// - verifies that we have not connected SWAP before on a different blockchain backend;
// - starts the chequebook; creates the swap instance
func New(dbPath string, prvkey *ecdsa.PrivateKey, backendURL string, params *Params, chequebookAddressFlag common.Address, initialDepositAmountFlag uint64) (swap *Swap, err error) {
	// auditLog for swap-logging purposes
	auditLog = newLogger(params.LogPath)
	// verify that backendURL is not empty
	if backendURL == "" {
		return nil, errors.New("no backend URL given")
	}
	auditLog.Info("connecting to SWAP API", "url", backendURL)
	// initialize the balances store
	var stateStore state.Store
	if stateStore, err = state.NewDBStore(filepath.Join(dbPath, "swap.db")); err != nil {
		return nil, fmt.Errorf("error while initializing statestore: %v", err)
	}
	if params.DisconnectThreshold <= params.PaymentThreshold {
		return nil, fmt.Errorf("disconnect threshold lower or at payment threshold. DisconnectThreshold: %d, PaymentThreshold: %d", params.DisconnectThreshold, params.PaymentThreshold)
	}
	// connect to the backend
	backend, err := ethclient.Dial(backendURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Ethereum API %s: %v", backendURL, err)
	}
	// get the chainID of the backend
	var chainID *big.Int
	if chainID, err = backend.ChainID(context.TODO()); err != nil {
		return nil, fmt.Errorf("error retrieving chainID from backendURL: %v", err)
	}
	// verify that we have not used SWAP before on a different chainID
	if err := checkChainID(chainID.Uint64(), stateStore); err != nil {
		return nil, err
	}
	auditLog.Info("Using backend network ID", "ID", chainID.Uint64())
	// create the owner of SWAP
	owner := createOwner(prvkey)
	// create the swap instance
	swap = new(
		stateStore,
		owner,
		backend,
		params,
	)
	// start the chequebook
	if swap.contract, err = swap.StartChequebook(chequebookAddressFlag, initialDepositAmountFlag); err != nil {
		return nil, err
	}
	return swap, nil
}

const (
	balancePrefix          = "balance_"
	sentChequePrefix       = "sent_cheque_"
	receivedChequePrefix   = "received_cheque_"
	connectedChequebookKey = "connected_chequebook"
	connectedBlockchainKey = "connected_blockchain"
	lastSentChequeKey      = "last_sent_cheque"
	lastReceivedChequeKey  = "last_received_cheque"
)

// checkChainID verifies whether we have initialized SWAP before and ensures that we are on the same backendNetworkID if this is the case
func checkChainID(currentChainID uint64, s state.Store) (err error) {
	var connectedBlockchain uint64
	err = s.Get(connectedBlockchainKey, &connectedBlockchain)
	// error reading from database
	if err != nil && err != state.ErrNotFound {
		return fmt.Errorf("error querying usedBeforeAtNetwork from statestore: %v", err)
	}
	// initialized before, but on a different chainID
	if err != state.ErrNotFound && connectedBlockchain != currentChainID {
		return fmt.Errorf("statestore previously used on different backend network. Used before on network: %d, Attempting to connect on network %d", connectedBlockchain, currentChainID)
	}
	if err == state.ErrNotFound {
		auditLog.Info("First time connected to SWAP. Storing chain ID", "ID", currentChainID)
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

func keyToID(key string, prefix string) enode.ID {
	return enode.HexID(key[len(prefix):])
}

// createOwner assings keys and addresses
func createOwner(prvkey *ecdsa.PrivateKey) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		address:    crypto.PubkeyToAddress(*pubkey),
		privateKey: prvkey,
		publicKey:  pubkey,
	}
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

	// Check if balance with peer is over the disconnect threshold
	balance := swapPeer.getBalance()
	if balance >= s.params.DisconnectThreshold {
		return fmt.Errorf("balance for peer %s is over the disconnect threshold %d, disconnecting", peer.ID().String(), s.params.DisconnectThreshold)
	}

	if err = swapPeer.updateBalance(amount); err != nil {
		return err
	}

	// Check if balance with peer crosses the payment threshold
	// It is the peer with a negative balance who sends a cheque, thus we check
	// that the balance is *below* the threshold
	if swapPeer.getBalance() <= -s.params.PaymentThreshold {
		auditLog.Info("balance for peer went over the payment threshold, sending cheque", "peer", peer.ID().String(), "payment threshold", s.params.PaymentThreshold)
		return swapPeer.sendCheque()
	}

	return nil
}

// handleMsg is for handling messages when receiving messages
func (s *Swap) handleMsg(p *Peer) func(ctx context.Context, msg interface{}) error {
	return func(ctx context.Context, msg interface{}) error {
		switch msg := msg.(type) {
		case *EmitChequeMsg:
			go s.handleEmitChequeMsg(ctx, p, msg)
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
	auditLog.Info("received cheque from peer", "peer", p.ID().String(), "honey", cheque.Honey)
	_, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		return err
	}

	auditLog.Debug("received cheque processed and verified", "peer", p.ID().String())

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	if err := p.updateBalance(-int64(cheque.Honey)); err != nil {
		return err
	}

	otherSwap, err := contract.InstanceAt(cheque.Contract, s.backend)
	if err != nil {
		return err
	}

	gasPrice, err := s.backend.SuggestGasPrice(context.TODO())
	if err != nil {
		return err
	}
	transactionCosts := gasPrice.Uint64() * 50000 // cashing a cheque is approximately 50000 gas
	paidOut, err := otherSwap.PaidOut(nil, cheque.Beneficiary)
	if err != nil {
		return err
	}
	// do a payout transaction if we get 2 times the gas costs
	if (cheque.CumulativePayout - paidOut.Uint64()) > 2*transactionCosts {
		opts := bind.NewKeyedTransactor(s.owner.privateKey)
		opts.Context = ctx
		// cash cheque in async, otherwise this blocks here until the TX is mined
		go defaultCashCheque(s, otherSwap, opts, cheque)
	}

	return err
}

// cashCheque should be called async as it blocks until the transaction(s) are mined
// The function cashes the cheque by sending it to the blockchain
func cashCheque(s *Swap, otherSwap contract.Contract, opts *bind.TransactOpts, cheque *Cheque) {
	// blocks here, as we are waiting for the transaction to be mined
	result, receipt, err := otherSwap.CashChequeBeneficiary(opts, s.backend, s.GetParams().ContractAddress, big.NewInt(int64(cheque.CumulativePayout)), cheque.Signature)
	if err != nil {
		// TODO: do something with the error
		// and we actually need to log this error as we are in an async routine; nobody is handling this error for now
		auditLog.Error("error cashing cheque", "err", err)
		return
	}

	if result.Bounced {
		auditLog.Warn("cheque bounced", "tx", receipt.TxHash)
		return
		// TODO: do something here
	}

	auditLog.Debug("cash tx mined", "receipt", receipt)
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
func (s *Swap) processAndVerifyCheque(cheque *Cheque, p *Peer) (uint64, error) {
	if err := cheque.verifyChequeProperties(p, s.owner.address); err != nil {
		return 0, err
	}

	lastCheque := p.getLastReceivedCheque()

	// TODO: there should probably be a lock here?
	expectedAmount, err := s.honeyPriceOracle.GetPrice(cheque.Honey)
	if err != nil {
		return 0, err
	}

	actualAmount, err := cheque.verifyChequeAgainstLast(lastCheque, expectedAmount)
	if err != nil {
		return 0, err
	}

	if err := p.setLastReceivedCheque(cheque); err != nil {
		auditLog.Error("error while saving last received cheque", "peer", p.ID().String(), "err", err.Error())
		// TODO: what do we do here? Related issue: https://github.com/ethersphere/swarm/issues/1515
	}

	return actualAmount, nil
}

// Balance returns the balance for a given peer
func (s *Swap) Balance(peer enode.ID) (balance int64, err error) {
	if swapPeer := s.getPeer(peer); swapPeer != nil {
		return swapPeer.getBalance(), nil
	}
	err = s.store.Get(balanceKey(peer), &balance)
	return balance, err
}

// Balances returns the balances for all known SWAP peers
func (s *Swap) Balances() (map[enode.ID]int64, error) {
	balances := make(map[enode.ID]int64)

	// get balances from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		balances[peer] = swapPeer.getBalance()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add disk balances for peers not already present
	balanceIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), balancePrefix)
		if _, peerHasBalance := balances[peer]; !peerHasBalance {
			var peerBalance int64
			err = json.Unmarshal(value, &peerBalance)
			if err == nil {
				balances[peer] = peerBalance
			}
		}
		return stop, err
	}
	err := s.store.Iterate(balancePrefix, balanceIterFunction)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

// Cheques returns all known last sent and received cheques, grouped by peer
func (s *Swap) Cheques() (map[enode.ID]map[string]*Cheque, error) {
	cheques := make(map[enode.ID]map[string]*Cheque)

	// get peer cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		cheques[peer][lastSentChequeKey] = swapPeer.getLastSentCheque()
		cheques[peer][lastReceivedChequeKey] = swapPeer.getLastSentCheque()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add disk cheques for peers not already present
	sentChequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), sentChequePrefix)
		if _, peerHasCheque := cheques[peer]; !peerHasCheque {
			var peerCheque Cheque
			err = json.Unmarshal(value, &peerCheque)
			if err == nil {
				cheques[peer][lastSentChequeKey] = &peerCheque
			}
		}
		return stop, err
	}
	err := s.store.Iterate(sentChequePrefix, sentChequesIterFunction)
	if err != nil {
		return nil, err
	}
	receivedChequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), receivedChequePrefix)
		if _, peerHasCheque := cheques[peer]; !peerHasCheque {
			var peerCheque Cheque
			err = json.Unmarshal(value, &peerCheque)
			if err == nil {
				cheques[peer][lastReceivedChequeKey] = &peerCheque
			}
		}
		return stop, err
	}
	err = s.store.Iterate(receivedChequePrefix, receivedChequesIterFunction)
	if err != nil {
		return nil, err
	}

	return cheques, nil
}

// PeerCheques returns the last sent and received cheques for a given peer
func (s *Swap) PeerCheques(peer enode.ID) (map[string]*Cheque, error) {
	sentCheque, err := s.SentCheque(peer)
	if err != nil && err != state.ErrNotFound {
		return nil, err
	}
	receivedCheque, err := s.ReceivedCheque(peer)
	if err != nil && err != state.ErrNotFound {
		return nil, err
	}
	return map[string]*Cheque{lastSentChequeKey: sentCheque, lastReceivedChequeKey: receivedCheque}, nil
}

// SentCheque returns the last sent cheque for a given peer
func (s *Swap) SentCheque(peer enode.ID) (cheque *Cheque, err error) {
	if swapPeer := s.getPeer(peer); swapPeer != nil {
		return swapPeer.getLastSentCheque(), nil
	}
	err = s.store.Get(sentChequeKey(peer), &cheque)
	return cheque, err
}

// SentCheques returns the last sent cheques for all known SWAP peers
func (s *Swap) SentCheques() (map[enode.ID]*Cheque, error) {
	cheques := make(map[enode.ID]*Cheque)

	// get sent cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		cheques[peer] = swapPeer.getLastSentCheque()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add disk sent cheques for peers not already present
	chequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), sentChequePrefix)
		if _, peerHasCheque := cheques[peer]; !peerHasCheque {
			var peerCheque Cheque
			err = json.Unmarshal(value, &peerCheque)
			if err == nil {
				cheques[peer] = &peerCheque
			}
		}
		return stop, err
	}
	err := s.store.Iterate(sentChequePrefix, chequesIterFunction)
	if err != nil {
		return nil, err
	}

	return cheques, nil
}

// ReceivedCheque returns the last received cheque for a given peer
func (s *Swap) ReceivedCheque(peer enode.ID) (cheque *Cheque, err error) {
	if swapPeer := s.getPeer(peer); swapPeer != nil {
		return swapPeer.getLastReceivedCheque(), nil
	}
	err = s.store.Get(receivedChequeKey(peer), &cheque)
	return cheque, err
}

// ReceivedCheques returns the last received cheques for all known SWAP peers
func (s *Swap) ReceivedCheques() (map[enode.ID]*Cheque, error) {
	cheques := make(map[enode.ID]*Cheque)

	// get received cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		cheques[peer] = swapPeer.getLastReceivedCheque()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add disk received cheques for peers not already present
	chequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), receivedChequePrefix)
		if _, peerHasCheque := cheques[peer]; !peerHasCheque {
			var peerCheque Cheque
			err = json.Unmarshal(value, &peerCheque)
			if err == nil {
				cheques[peer] = &peerCheque
			}
		}
		return stop, err
	}
	err := s.store.Iterate(receivedChequePrefix, chequesIterFunction)
	if err != nil {
		return nil, err
	}

	return cheques, nil
}

// loadLastReceivedCheque loads the last received cheque for the peer from the store
// and returns nil when there never was a cheque saved
func (s *Swap) loadLastReceivedCheque(p enode.ID) (cheque *Cheque, err error) {
	err = s.store.Get(receivedChequeKey(p), &cheque)
	if err == state.ErrNotFound {
		return nil, nil
	}
	return cheque, err
}

// loadLastSentCheque loads the last sent cheque for the peer from the store
// and returns nil when there never was a cheque saved
func (s *Swap) loadLastSentCheque(p enode.ID) (cheque *Cheque, err error) {
	err = s.store.Get(sentChequeKey(p), &cheque)
	if err == state.ErrNotFound {
		return nil, nil
	}
	return cheque, err
}

// loadBalance loads the current balance for the peer from the store
// and returns 0 if there was no prior balance saved
func (s *Swap) loadBalance(p enode.ID) (balance int64, err error) {
	err = s.store.Get(balanceKey(p), &balance)
	if err == state.ErrNotFound {
		return 0, nil
	}
	return balance, err
}

// saveLastReceivedCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastReceivedCheque(p enode.ID, cheque *Cheque) error {
	return s.store.Put(receivedChequeKey(p), cheque)
}

// saveLastSentCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastSentCheque(p enode.ID, cheque *Cheque) error {
	return s.store.Put(sentChequeKey(p), cheque)
}

// saveBalance saves balance as the current balance for peer
func (s *Swap) saveBalance(p enode.ID, balance int64) error {
	return s.store.Put(balanceKey(p), balance)
}

// Close cleans up swap
func (s *Swap) Close() error {
	return s.store.Close()
}

// GetParams returns contract parameters (Bin, ABI, contractAddress) from the contract
func (s *Swap) GetParams() *swap.Params {
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

func promptInitialDepositAmount() (uint64, error) {
	// need to prompt user for initial deposit amount
	// if 0, can not cash in cheques
	prompter := console.Stdin

	// ask user for input
	input, err := prompter.PromptInput("Please provide the amount in Wei which will deposited to your chequebook upon deployment: ")
	if err != nil {
		return 0, err
	}
	// check input
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		// maybe we should provide a fallback here? A bad input results in stopping the boot
		return 0, fmt.Errorf("Conversion error while reading user input: %v", err)
	}
	return uint64(val), nil
}

// StartChequebook starts the chequebook, taking into account the chequebookAddress passed in by the user and the chequebook addresses saved on the node's database
func (s *Swap) StartChequebook(chequebookAddrFlag common.Address, initialDepositAmount uint64) (contract contract.Contract, err error) {
	previouslyUsedChequebook, err := s.loadChequebook()
	// error reading from disk
	if err != nil && err != state.ErrNotFound {
		return nil, fmt.Errorf("Error reading previously used chequebook: %s", err)
	}
	// read from state, but provided flag is not the same
	if err == nil && (chequebookAddrFlag != common.Address{} && chequebookAddrFlag != previouslyUsedChequebook) {
		return nil, fmt.Errorf("Attempting to connect to provided chequebook, but different chequebook used before")
	}
	// nothing written to state disk before, no flag provided: deploying new chequebook
	if err == state.ErrNotFound && chequebookAddrFlag == (common.Address{}) {
		var toDeposit = initialDepositAmount
		if toDeposit == 0 {
			toDeposit, err = promptInitialDepositAmount()
			if err != nil {
				return nil, err
			}
		}
		if contract, err = s.Deploy(context.TODO(), toDeposit); err != nil {
			return nil, err
		}
		if err := s.saveChequebook(contract.ContractParams().ContractAddress); err != nil {
			return nil, err
		}
		auditLog.Info("Deployed chequebook", "contract address", contract.ContractParams().ContractAddress.Hex(), "deposit", toDeposit, "owner", s.owner.address)
		// first time connecting by deploying a new chequebook
		return contract, nil
	}
	// first time connecting with a chequebookAddress passed in
	if chequebookAddrFlag != (common.Address{}) {
		return bindToContractAt(chequebookAddrFlag, s.backend)
	}
	// reconnecting with contract read from statestore
	return bindToContractAt(previouslyUsedChequebook, s.backend)
}

// BindToContractAt binds to an instance of an already existing chequebook contract at address
func bindToContractAt(address common.Address, backend contract.Backend) (contract.Contract, error) {
	// validate whether address is a chequebook
	if err := contract.ValidateCode(context.Background(), backend, address); err != nil {
		return nil, fmt.Errorf("contract validation for %v failed: %v", address.Hex(), err)
	}
	auditLog.Info("bound to chequebook", "chequebookAddr", address)
	// get the instance
	return contract.InstanceAt(address, backend)
}

// Deploy deploys the Swap contract
func (s *Swap) Deploy(ctx context.Context, initialDepositAmount uint64) (contract.Contract, error) {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(initialDepositAmount))
	opts.Context = ctx
	auditLog.Info("Deploying new swap", "owner", opts.From.Hex(), "deposit", opts.Value)
	return s.deployLoop(opts, defaultHarddepositTimeoutDuration)
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, defaultHarddepositTimeoutDuration time.Duration) (instance contract.Contract, err error) {
	var tx *types.Transaction
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}
		if instance, tx, err = contract.Deploy(opts, s.backend, s.owner.address, defaultHarddepositTimeoutDuration); err != nil {
			auditLog.Warn("can't send chequebook deploy tx", "try", try, "error", err)
			continue
		}
		if _, err := bind.WaitDeployed(opts.Context, s.backend, tx); err != nil {
			auditLog.Warn("chequebook deploy error", "try", try, "error", err)
			continue
		}
		return instance, nil
	}
	return nil, err
}

func (s *Swap) loadChequebook() (common.Address, error) {
	var chequebook common.Address
	err := s.store.Get(connectedChequebookKey, &chequebook)
	return chequebook, err
}

func (s *Swap) saveChequebook(chequebook common.Address) error {
	return s.store.Put(connectedChequebookKey, chequebook)
}
