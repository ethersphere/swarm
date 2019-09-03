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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/swap"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

// ErrInvalidChequeSignature indicates the signature on the cheque was invalid
var ErrInvalidChequeSignature = errors.New("invalid cheque signature")

// Swap represents the Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	api                 API
	store               state.Store        // store is needed in order to keep balances and cheques across sessions
	accountingLock      sync.RWMutex       // lock for data consistency in accounting-related functions
	storeLock           sync.RWMutex       // lock for store access
	peers               map[enode.ID]*Peer // map of all swap Peers
	peersLock           sync.RWMutex       // lock for peers map
	backend             contract.Backend   // the backend (blockchain) used
	owner               *Owner             // contract access
	params              *Params            // economic and operational parameters
	contract            swap.Contract      // reference to the smart contract
	oracle              PriceOracle        // the oracle providing the ether price for honey
	paymentThreshold    int64              // balance difference required for sending cheque
	disconnectThreshold int64              // balance difference required for dropping peer
}

// Owner encapsulates information related to accessing the contract
type Owner struct {
	Contract   common.Address    // address of swap contract
	address    common.Address    // owner address
	privateKey *ecdsa.PrivateKey // private key
	publicKey  *ecdsa.PublicKey  // public key
}

// Params encapsulates param
type Params struct {
	InitialDepositAmount uint64
}

// NewParams returns a Params struct filled with default values
func NewParams() *Params {
	return &Params{
		InitialDepositAmount: DefaultInitialDepositAmount,
	}
}

// New - swap constructor
func New(stateStore state.Store, prvkey *ecdsa.PrivateKey, backend contract.Backend) *Swap {
	return &Swap{
		store:               stateStore,
		peers:               make(map[enode.ID]*Peer),
		backend:             backend,
		owner:               createOwner(prvkey),
		params:              NewParams(),
		paymentThreshold:    DefaultPaymentThreshold,
		disconnectThreshold: DefaultDisconnectThreshold,
		oracle:              NewPriceOracle(),
	}
}

const (
	balancePrefix        = "balance_"
	sentChequePrefix     = "sent_cheque_"
	receivedChequePrefix = "received_cheque_"
)

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

// DeploySuccess is for convenience log output
func (s *Swap) DeploySuccess() string {
	return fmt.Sprintf("contract: %s, owner: %s, deposit: %v, signer: %x", s.owner.Contract.Hex(), s.owner.address.Hex(), s.params.InitialDepositAmount, s.owner.publicKey)
}

// Add is the (sole) accounting function
// Swap implements the protocols.Balance interface
func (s *Swap) Add(amount int64, peer *protocols.Peer) (err error) {
	s.accountingLock.Lock()
	defer s.accountingLock.Unlock()
	swapPeer, ok := s.getPeer(peer.ID())
	if !ok {
		return fmt.Errorf("peer %s not a swap enabled peer", peer.ID().String())
	}
	swapPeer.lock.Lock()
	defer swapPeer.lock.Unlock()

	// Check if balance with peer is over the disconnect threshold
	balance := swapPeer.getBalance()
	if balance >= s.disconnectThreshold {
		return fmt.Errorf("balance for peer %s is over the disconnect threshold %d, disconnecting", peer.ID().String(), s.disconnectThreshold)
	}

	if err = swapPeer.updateBalance(amount); err != nil {
		return err
	}

	// Check if balance with peer crosses the payment threshold
	// It is the peer with a negative balance who sends a cheque, thus we check
	// that the balance is *below* the threshold
	if swapPeer.getBalance() <= -s.paymentThreshold {
		log.Warn("balance for peer went over the payment threshold, sending cheque", "peer", peer.ID().String(), "payment threshold", s.paymentThreshold)
		if !ok {
			return fmt.Errorf("peer %s not found", peer)
		}
		return s.sendCheque(swapPeer)
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
	log.Info("received cheque from peer", "peer", p.ID().String(), "honey", cheque.Honey)
	_, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		return err
	}

	log.Debug("received cheque processed and verified", "peer", p.ID().String())

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	p.updateBalance(-int64(cheque.Honey))
	if err != nil {
		return err
	}

	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	opts.Context = ctx

	otherSwap, err := contract.InstanceAt(cheque.Contract, s.backend)
	if err != nil {
		return err
	}

	// cash cheque in async, otherwise this blocks here until the TX is mined
	go defaultCashCheque(s, otherSwap, opts, cheque)

	return err
}

// cashCheque should be called async as it blocks until the transaction(s) are mined
// The function cashes the cheque by sending it to the blockchain
func cashCheque(s *Swap, otherSwap contract.Contract, opts *bind.TransactOpts, cheque *Cheque) {
	// blocks here, as we are waiting for the transaction to be mined
	result, receipt, err := otherSwap.CashChequeBeneficiary(opts, s.backend, s.owner.Contract, big.NewInt(int64(cheque.CumulativePayout)), cheque.Signature)
	if err != nil {
		// TODO: do something with the error
		// and we actually need to log this error as we are in an async routine; nobody is handling this error for now
		log.Error("error cashing cheque", "err", err)
		return
	}

	if result.Bounced {
		log.Error("cheque bounced", "tx", receipt.TxHash)
		return
		// TODO: do something here
	}

	log.Debug("cash tx mined", "receipt", receipt)
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
func (s *Swap) processAndVerifyCheque(cheque *Cheque, p *Peer) (uint64, error) {
	if err := cheque.verifyChequeProperties(p, s.owner.address); err != nil {
		return 0, err
	}

	lastCheque := p.getLastReceivedCheque()

	// TODO: there should probably be a lock here?
	expectedAmount, err := s.oracle.GetPrice(cheque.Honey)
	if err != nil {
		return 0, err
	}

	actualAmount, err := cheque.verifyChequeAgainstLast(lastCheque, expectedAmount)
	if err != nil {
		return 0, err
	}

	if err := p.setLastReceivedCheque(cheque); err != nil {
		log.Error("error while saving last received cheque", "peer", p.ID().String(), "err", err.Error())
		// TODO: what do we do here? Related issue: https://github.com/ethersphere/swarm/issues/1515
	}

	return actualAmount, nil
}

// sendCheque sends a cheque to peer
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (s *Swap) sendCheque(swapPeer *Peer) error {
	cheque, err := s.createCheque(swapPeer)
	if err != nil {
		return fmt.Errorf("error while creating cheque: %s", err.Error())
	}

	log.Info("sending cheque", "honey", cheque.Honey, "cumulativePayout", cheque.ChequeParams.CumulativePayout, "beneficiary", cheque.Beneficiary, "contract", cheque.Contract)

	if err := swapPeer.setLastSentCheque(cheque); err != nil {
		return fmt.Errorf("error while storing the last cheque: %s", err.Error())
	}

	emit := &EmitChequeMsg{
		Cheque: cheque,
	}

	if err := swapPeer.updateBalance(int64(cheque.Honey)); err != nil {
		return err
	}

	return swapPeer.Send(context.Background(), emit)
}

// createCheque creates a new cheque whose beneficiary will be the peer and
// whose amount is based on the last cheque and current balance for this peer
// The cheque will be signed and point to the issuer's contract
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (s *Swap) createCheque(swapPeer *Peer) (*Cheque, error) {
	var cheque *Cheque
	var err error

	beneficiary := swapPeer.beneficiary
	peerBalance := swapPeer.getBalance()
	// the balance should be negative here, we take the absolute value:
	honey := uint64(-peerBalance)
	var amount uint64
	amount, err = s.oracle.GetPrice(honey)
	if err != nil {
		return nil, fmt.Errorf("error getting price from oracle: %s", err.Error())
	}

	// if there is no existing cheque when loading from the store, it means it's the first interaction
	// this is a valid scenario
	total, err := swapPeer.getLastChequeValues()
	if err != nil && err != state.ErrNotFound {
		return nil, err
	}

	cheque = &Cheque{
		ChequeParams: ChequeParams{
			CumulativePayout: total + amount,
			Contract:         s.owner.Contract,
			Beneficiary:      beneficiary,
		},
		Honey: honey,
	}
	cheque.Signature, err = cheque.Sign(s.owner.privateKey)

	return cheque, err
}

// Balance returns the balance for a given peer
func (s *Swap) Balance(peer enode.ID) (int64, error) {
	swapPeer, ok := s.peers[peer]
	if !ok {
		return 0, state.ErrNotFound
	}
	return swapPeer.getBalance(), nil
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

// loadLastReceivedCheque loads the last received cheque for the peer from the store
func (s *Swap) loadLastReceivedCheque(p enode.ID) (*Cheque, error) {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	var cheque *Cheque
	error := s.store.Get(receivedChequeKey(p), &cheque)
	return cheque, error
}

// loadLastSentCheque loads the last sent cheque for the peer from the store
func (s *Swap) loadLastSentCheque(p enode.ID) (*Cheque, error) {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	var cheque *Cheque
	error := s.store.Get(sentChequeKey(p), &cheque)
	return cheque, error
}

// loadBalance loads the current balance for the peer from the store
func (s *Swap) loadBalance(p enode.ID) (int64, error) {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	var balance int64
	error := s.store.Get(balanceKey(p), &balance)
	return balance, error
}

// saveLastReceivedCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastReceivedCheque(p enode.ID, cheque *Cheque) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	return s.store.Put(receivedChequeKey(p), cheque)
}

// saveLastSentCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastSentCheque(p enode.ID, cheque *Cheque) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	return s.store.Put(sentChequeKey(p), cheque)
}

// saveBalance saves balance as the current balance for peer
func (s *Swap) saveBalance(p enode.ID, balance int64) error {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()
	return s.store.Put(balanceKey(p), balance)
}

// Close cleans up swap
func (s *Swap) Close() error {
	return s.store.Close()
}

// GetParams returns contract parameters (Bin, ABI) from the contract
func (s *Swap) GetParams() *swap.Params {
	return s.contract.ContractParams()
}

// setChequebookAddr sets the chequebook address
func (s *Swap) setChequebookAddr(chequebookAddr common.Address) {
	s.owner.Contract = chequebookAddr
}

// getContractOwner retrieve the owner of the chequebook at address from the blockchain
func (s *Swap) getContractOwner(ctx context.Context, address common.Address) (common.Address, error) {
	contr, err := contract.InstanceAt(address, s.backend)
	if err != nil {
		return common.Address{}, err
	}

	return contr.Issuer(nil)
}

// StartChequebook deploys a new instance of a chequebook if chequebookAddr is empty, otherwise it wil bind to an existing instance
func (s *Swap) StartChequebook(chequebookAddr common.Address) error {
	if chequebookAddr != (common.Address{}) {
		if err := s.BindToContractAt(chequebookAddr); err != nil {
			return err
		}
		log.Info("Using the provided chequebook", "chequebookAddr", chequebookAddr)
	} else {
		if err := s.Deploy(context.Background(), s.backend); err != nil {
			return err
		}
		log.Info("New SWAP contract deployed", "contract info", s.DeploySuccess())
	}
	return nil
}

// BindToContractAt binds an instance of an already existing chequebook contract at address and sets chequebookAddr
func (s *Swap) BindToContractAt(address common.Address) (err error) {

	if err := contract.ValidateCode(context.Background(), s.backend, address); err != nil {
		return fmt.Errorf("contract validation for %v failed: %v", address, err)
	}
	s.contract, err = contract.InstanceAt(address, s.backend)
	if err != nil {
		return err
	}
	s.setChequebookAddr(address)
	return nil
}

// Deploy deploys the Swap contract and sets the contract address
func (s *Swap) Deploy(ctx context.Context, backend swap.Backend) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(s.params.InitialDepositAmount))
	opts.Context = ctx

	log.Info("deploying new swap", "owner", opts.From.Hex())
	address, err := s.deployLoop(opts, backend, s.owner.address, defaultHarddepositTimeoutDuration)
	if err != nil {
		log.Error("unable to deploy swap", "error", err)
		return err
	}
	s.setChequebookAddr(address)
	log.Info("swap deployed", "address", address.Hex(), "owner", opts.From.Hex())

	return err
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, backend swap.Backend, owner common.Address, defaultHarddepositTimeoutDuration time.Duration) (addr common.Address, err error) {
	var tx *types.Transaction
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}

		if _, s.contract, tx, err = contract.Deploy(opts, backend, owner, defaultHarddepositTimeoutDuration); err != nil {
			log.Warn("can't send chequebook deploy tx", "try", try, "error", err)
			continue
		}
		if addr, err = bind.WaitDeployed(opts.Context, backend, tx); err != nil {
			log.Warn("chequebook deploy error", "try", try, "error", err)
			continue
		}
		return addr, nil
	}
	return addr, err
}
