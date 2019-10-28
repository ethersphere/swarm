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
	"encoding/hex"
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

var swapLog log.Logger // logger for Swap related messages and audit trail
const swapLogLevel = 3 // swapLogLevel indicates filter level of log messages

// Swap represents the Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	store             state.Store                // store is needed in order to keep balances and cheques across sessions
	peers             map[enode.ID]*Peer         // map of all swap Peers
	peersLock         sync.RWMutex               // lock for peers map
	backend           contract.Backend           // the backend (blockchain) used
	owner             *Owner                     // contract access
	params            *Params                    // economic and operational parameters
	contract          contract.Contract          // reference to the smart contract
	chequebookFactory contract.SimpleSwapFactory // the chequebook factory used
	honeyPriceOracle  HoneyOracle                // oracle which resolves the price of honey (in Wei)
	chainID           *big.Int
}

// Owner encapsulates information related to accessing the contract
type Owner struct {
	address    common.Address    // owner address
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
}

// Params encapsulates economic and operational parameters
type Params struct {
	OverlayAddr         []byte // this node's base address
	LogPath             string // optional audit log path
	PaymentThreshold    int64  // honey amount at which a payment is triggered
	DisconnectThreshold int64  // honey amount at which a peer disconnects
}

// newSwapLogger returns a new logger for standard swap logs
func newSwapLogger(logPath string, overlayAddr []byte) log.Logger {
	swapLogger := log.New("swaplog", "*", "base", hex.EncodeToString(overlayAddr)[:16])
	setLoggerHandler(logPath, swapLogger)
	return swapLogger
}

// newPeerLogger returns a new logger for swap logs with peer info
func newPeerLogger(s *Swap, peerID enode.ID) log.Logger {
	peerLogger := log.New("swaplog", "*", "base", hex.EncodeToString(s.params.OverlayAddr)[:16], "peer", peerID.String()[:16])
	setLoggerHandler(s.params.LogPath, peerLogger)
	return peerLogger
}

// setLoggerHandler will set the logger handle to write logs to the specified path
// or use the default swarm logger in case this isn't specified or an error occurs
func setLoggerHandler(logpath string, logger log.Logger) {
	lh := log.Root().GetHandler()

	if logpath == "" {
		logger.SetHandler(lh)
		return
	}

	rfh, err := swapRotatingFileHandler(logpath)

	if err != nil {
		log.Warn("RotatingFileHandler was not initialized", "logdir", logpath, "err", err)
		// use the default swarm logger as a fallback
		logger.SetHandler(lh)
		return
	}

	// filter messages with the correct log level for swap
	rfh = log.LvlFilterHandler(log.Lvl(swapLogLevel), rfh)

	// dispatch the logs to the default swarm log and also the filtered swap logger
	logger.SetHandler(log.MultiHandler(lh, rfh))
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

// newSwapInstance is a swap constructor function without integrity checks
func newSwapInstance(stateStore state.Store, owner *Owner, backend contract.Backend, chainID *big.Int, params *Params, chequebookFactory contract.SimpleSwapFactory) *Swap {
	return &Swap{
		store:             stateStore,
		peers:             make(map[enode.ID]*Peer),
		backend:           backend,
		owner:             owner,
		params:            params,
		chequebookFactory: chequebookFactory,
		honeyPriceOracle:  NewHoneyPriceOracle(),
		chainID:           chainID,
	}
}

// New prepares and creates all fields to create a swap instance:
// - sets up a SWAP database;
// - verifies whether the disconnect threshold is higher than the payment threshold;
// - connects to the blockchain backend;
// - verifies that we have not connected SWAP before on a different blockchain backend;
// - starts the chequebook; creates the swap instance
func New(dbPath string, prvkey *ecdsa.PrivateKey, backendURL string, params *Params, chequebookAddressFlag common.Address, initialDepositAmountFlag uint64, factoryAddress common.Address) (swap *Swap, err error) {
	// swap log for auditing purposes
	swapLog = newSwapLogger(params.LogPath, params.OverlayAddr)
	// verify that backendURL is not empty
	if backendURL == "" {
		return nil, errors.New("no backend URL given")
	}
	swapLog.Info("connecting to SWAP API", "url", backendURL)
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
	swapLog.Info("Using backend network ID", "ID", chainID.Uint64())

	// create the owner of SWAP
	owner := createOwner(prvkey)
	// initialize the factory
	factory, err := createFactory(factoryAddress, chainID, backend)
	if err != nil {
		return nil, err
	}

	// create the swap instance
	swap = newSwapInstance(
		stateStore,
		owner,
		backend,
		chainID,
		params,
		factory,
	)
	// start the chequebook
	if swap.contract, err = swap.StartChequebook(chequebookAddressFlag, initialDepositAmountFlag); err != nil {
		return nil, err
	}
	availableBalance, err := swap.AvailableBalance()
	if err != nil {
		return nil, err
	}

	swapLog.Info("available balance", "balance", availableBalance)

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
func createFactory(factoryAddress common.Address, chainID *big.Int, backend contract.Backend) (factory swap.SimpleSwapFactory, err error) {
	if (factoryAddress == common.Address{}) {
		if factoryAddress, err = contract.FactoryAddressForNetwork(chainID.Uint64()); err != nil {
			return nil, err
		}
	}
	swapLog.Info("Using chequebook factory", "address", factoryAddress)
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
		swapLog.Info("First time connected to SWAP. Storing chain ID", "ID", currentChainID)
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

	return s.checkPaymentThresholdAndSendCheque(swapPeer)
}

// checkPaymentThresholdAndSendCheque checks if balance with peer crosses the payment threshold and attempts to send a cheque if so
// It is the peer with a negative balance who sends a cheque, thus we check
// that the balance is *below* the threshold
// the caller is expected to hold swapPeer.lock
func (s *Swap) checkPaymentThresholdAndSendCheque(swapPeer *Peer) error {
	if swapPeer.getBalance() <= -s.params.PaymentThreshold {
		swapPeer.logger.Info("balance for peer went over the payment threshold, sending cheque", "payment threshold", s.params.PaymentThreshold)
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
		case *ConfirmChequeMsg:
			go s.handleConfirmChequeMsg(ctx, p, msg)
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
	p.logger.Info("received cheque from peer", "honey", cheque.Honey)

	if p.getLastReceivedCheque() != nil && cheque.Equal(p.getLastReceivedCheque()) {
		p.logger.Warn("cheque sent by peer has already been received in the past", "cumulativePayout", cheque.CumulativePayout)
		return p.Send(ctx, &ConfirmChequeMsg{
			Cheque: cheque,
		})
	}

	_, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		log.Error("error processing and verifying received cheque", "err", err)
		return err
	}

	p.logger.Debug("processed and verified received cheque", "beneficiary", cheque.Beneficiary, "cumulative payout", cheque.CumulativePayout)

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	err = p.updateBalance(-int64(cheque.Honey))
	if err != nil {
		log.Error("error updating balance", "err", err)
		return err
	}

	err = p.Send(ctx, &ConfirmChequeMsg{
		Cheque: cheque,
	})
	if err != nil {
		return err
	}

	otherSwap, err := contract.InstanceAt(cheque.Contract, s.backend)
	if err != nil {
		log.Error("error getting contract", "err", err)
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

func (s *Swap) handleConfirmChequeMsg(ctx context.Context, p *Peer, msg *ConfirmChequeMsg) {
	p.lock.Lock()
	defer p.lock.Unlock()
	cheque := msg.Cheque

	if p.getPendingCheque() == nil {
		p.logger.Warn("ignoring confirm msg, no pending cheque", "confirm message cheque", cheque)
		return
	}

	if !cheque.Equal(p.getPendingCheque()) {
		p.logger.Warn("ignoring confirm msg, unexpected cheque", "confirm message cheque", cheque, "expected", p.getPendingCheque())
		return
	}

	err := p.setLastSentCheque(cheque)
	if err != nil {
		p.Drop(fmt.Sprintf("persistence error: %v", err))
		return
	}

	err = p.setPendingCheque(nil)
	if err != nil {
		p.Drop(fmt.Sprintf("persistence error: %v", err))
		return
	}
}

// cashCheque should be called async as it blocks until the transaction(s) are mined
// The function cashes the cheque by sending it to the blockchain
func cashCheque(s *Swap, otherSwap contract.Contract, opts *bind.TransactOpts, cheque *Cheque) {
	// blocks here, as we are waiting for the transaction to be mined
	result, receipt, err := otherSwap.CashChequeBeneficiary(opts, s.GetParams().ContractAddress, big.NewInt(int64(cheque.CumulativePayout)), cheque.Signature)
	if err != nil {
		// TODO: do something with the error
		// and we actually need to log this error as we are in an async routine; nobody is handling this error for now
		swapLog.Error("error cashing cheque", "err", err)
		return
	}

	if result.Bounced {
		swapLog.Warn("cheque bounced", "tx", receipt.TxHash)
		return
		// TODO: do something here
	}

	swapLog.Debug("cash tx mined", "receipt", receipt)
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
		p.logger.Error("error while saving last received cheque", "err", err.Error())
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

	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		balances[peer] = swapPeer.getBalance()
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// add store balances, if peer was not already added
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
func (s *Swap) Cheques() (map[enode.ID]*PeerCheques, error) {
	cheques := make(map[enode.ID]*PeerCheques)

	// get peer cheques from memory
	s.peersLock.Lock()
	for peer, swapPeer := range s.peers {
		swapPeer.lock.Lock()
		pendingCheque := swapPeer.getPendingCheque()
		sentCheque := swapPeer.getLastSentCheque()
		receivedCheque := swapPeer.getLastReceivedCheque()
		// don't add peer to result if there are no cheques
		if sentCheque != nil || receivedCheque != nil || pendingCheque != nil {
			cheques[peer] = &PeerCheques{pendingCheque, sentCheque, receivedCheque}
		}
		swapPeer.lock.Unlock()
	}
	s.peersLock.Unlock()

	// get peer cheques from store
	err := s.addStoreCheques(pendingChequePrefix, cheques)
	if err != nil {
		return nil, err
	}
	err = s.addStoreCheques(sentChequePrefix, cheques)
	if err != nil {
		return nil, err
	}
	err = s.addStoreCheques(receivedChequePrefix, cheques)
	if err != nil {
		return nil, err
	}

	return cheques, nil
}

// AvailableBalance returns the total balance of the chequebook against which new cheques can be written
func (s *Swap) AvailableBalance() (uint64, error) {
	// get the LiquidBalance of the chequebook
	liquidBalance, err := s.contract.LiquidBalance(nil)
	if err != nil {
		return 0, err
	}

	// get all cheques
	cheques, err := s.Cheques()
	if err != nil {
		return 0, err
	}

	// Compute the total worth of cheques sent and how much of of this is cashed
	var sentChequesWorth uint64
	var cashedChequesWorth uint64
	for _, peerCheques := range cheques {
		var sentCheque *Cheque
		if peerCheques.PendingCheque != nil {
			sentCheque = peerCheques.PendingCheque
		} else if peerCheques.LastSentCheque != nil {
			sentCheque = peerCheques.LastSentCheque
		} else {
			continue
		}
		sentChequesWorth += sentCheque.ChequeParams.CumulativePayout
		paidOut, err := s.contract.PaidOut(nil, sentCheque.ChequeParams.Beneficiary)
		if err != nil {
			return 0, err
		}
		cashedChequesWorth += paidOut.Uint64()
	}
	return liquidBalance.Uint64() + cashedChequesWorth - sentChequesWorth, nil
}

// add cheques from store for peers not already present in given cheques map
func (s *Swap) addStoreCheques(chequePrefix string, cheques map[enode.ID]*PeerCheques) error {
	chequesIterFunction := func(key []byte, value []byte) (stop bool, err error) {
		peer := keyToID(string(key), chequePrefix)
		// create struct if peer has no cheques entry yet
		if peerCheques := cheques[peer]; peerCheques == nil {
			cheques[peer] = &PeerCheques{}
		}

		// add cheque from store if not already in result
		var peerCheque Cheque
		err = json.Unmarshal(value, &peerCheque)
		if err == nil {
			switch chequePrefix {
			case pendingChequePrefix:
				cheques[peer].PendingCheque = &peerCheque
			case sentChequePrefix:
				cheques[peer].LastSentCheque = &peerCheque
			case receivedChequePrefix:
				cheques[peer].LastReceivedCheque = &peerCheque
			default:
				err = fmt.Errorf("unknown type of cheque requested through prefix %s", chequePrefix)
			}
		}
		return stop, err
	}
	return s.store.Iterate(chequePrefix, chequesIterFunction)
}

// PeerCheques contains the last cheque known to have been sent to a peer, as well as the last one received from the peer
type PeerCheques struct {
	PendingCheque      *Cheque
	LastSentCheque     *Cheque
	LastReceivedCheque *Cheque
}

// PeerCheques returns the last sent and received cheques for a given peer
func (s *Swap) PeerCheques(peer enode.ID) (PeerCheques, error) {
	var pendingCheque, sentCheque, receivedCheque *Cheque

	swapPeer := s.getPeer(peer)
	if swapPeer != nil {
		pendingCheque = swapPeer.getPendingCheque()
		sentCheque = swapPeer.getLastSentCheque()
		receivedCheque = swapPeer.getLastReceivedCheque()
	} else {
		errPendingCheque := s.store.Get(pendingChequeKey(peer), &pendingCheque)
		if errPendingCheque != nil && errPendingCheque != state.ErrNotFound {
			return PeerCheques{}, errPendingCheque
		}
		errSentCheque := s.store.Get(sentChequeKey(peer), &sentCheque)
		if errSentCheque != nil && errSentCheque != state.ErrNotFound {
			return PeerCheques{}, errSentCheque
		}
		errReceivedCheque := s.store.Get(receivedChequeKey(peer), &receivedCheque)
		if errReceivedCheque != nil && errReceivedCheque != state.ErrNotFound {
			return PeerCheques{}, errReceivedCheque
		}
	}
	return PeerCheques{pendingCheque, sentCheque, receivedCheque}, nil
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
func (s *Swap) loadBalance(p enode.ID) (balance int64, err error) {
	err = s.store.Get(balanceKey(p), &balance)
	if err == state.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
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
func (s *Swap) saveBalance(p enode.ID, balance int64) error {
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
		swapLog.Info("Deployed chequebook", "contract address", contract.ContractParams().ContractAddress.Hex(), "deposit", toDeposit, "owner", s.owner.address)
		// first time connecting by deploying a new chequebook
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
		return nil, fmt.Errorf("contract validation for %v failed: %v", address.Hex(), err)
	}
	swapLog.Info("bound to chequebook", "chequebookAddr", address)
	// get the instance
	return contract.InstanceAt(address, s.backend)
}

// Deploy deploys the Swap contract
func (s *Swap) Deploy(ctx context.Context, initialDepositAmount uint64) (contract.Contract, error) {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(initialDepositAmount))
	opts.Context = ctx
	swapLog.Info("Deploying new swap", "owner", opts.From.Hex(), "deposit", opts.Value)
	return s.deployLoop(opts, defaultHarddepositTimeoutDuration)
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, defaultHarddepositTimeoutDuration time.Duration) (instance contract.Contract, err error) {
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}

		chequebook, err := s.chequebookFactory.DeploySimpleSwap(opts, s.owner.address, big.NewInt(int64(defaultHarddepositTimeoutDuration)))
		if err != nil {
			swapLog.Warn("chequebook deploy error, retrying...", "try", try, "error", err)
			continue
		}

		return chequebook, nil
	}
	return nil, fmt.Errorf("failed to deploy chequebook: %v", err)
}

func (s *Swap) loadChequebook() (common.Address, error) {
	var chequebook common.Address
	err := s.store.Get(connectedChequebookKey, &chequebook)
	return chequebook, err
}

func (s *Swap) saveChequebook(chequebook common.Address) error {
	return s.store.Put(connectedChequebookKey, chequebook)
}
