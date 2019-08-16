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
	"strconv"
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
	api                 PublicAPI
	store               state.Store          // store is needed in order to keep balances and cheques across sessions
	accountingLock      sync.RWMutex         // lock for data consistency in accounting-related functions
	balances            map[enode.ID]int64   // map of balances for each peer
	balancesLock        sync.RWMutex         // lock for balances map
	cheques             map[enode.ID]*Cheque // map of cheques for each peer
	chequesLock         sync.RWMutex         // lock for cheques map
	peers               map[enode.ID]*Peer   // map of all swap Peers
	peersLock           sync.RWMutex         // lock for peers map
	backend             contract.Backend     // the backend (blockchain) used
	owner               *Owner               // contract access
	params              *Params              // economic and operational parameters
	contract            swap.Contract        // reference to the smart contract
	oracle              PriceOracle          // the oracle providing the ether price for honey
	paymentThreshold    int64                // balance difference required for sending cheque
	disconnectThreshold int64                // balance difference required for dropping peer
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
func New(stateStore state.Store, prvkey *ecdsa.PrivateKey, contract common.Address, backend contract.Backend) *Swap {
	return &Swap{
		store:               stateStore,
		balances:            make(map[enode.ID]int64),
		cheques:             make(map[enode.ID]*Cheque),
		peers:               make(map[enode.ID]*Peer),
		backend:             backend,
		owner:               createOwner(prvkey, contract),
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
func createOwner(prvkey *ecdsa.PrivateKey, contract common.Address) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		privateKey: prvkey,
		publicKey:  pubkey,
		Contract:   contract,
		address:    crypto.PubkeyToAddress(*pubkey),
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

	err = s.loadBalance(peer.ID())
	if err != nil && err != state.ErrNotFound {
		return fmt.Errorf("error while loading balance for peer %s", peer.ID().String())
	}

	// Check if balance with peer is over the disconnect threshold
	balance, exists := s.getBalance(peer.ID())
	if !exists {
		return fmt.Errorf("peer %v does not exist", peer.ID())
	}
	if balance >= s.disconnectThreshold {
		return fmt.Errorf("balance for peer %s is over the disconnect threshold %d, disconnecting", peer.ID().String(), s.disconnectThreshold)
	}

	var newBalance int64
	newBalance, err = s.updateBalance(peer.ID(), amount)
	if err != nil {
		return err
	}

	// Check if balance with peer crosses the payment threshold
	// It is the peer with a negative balance who sends a cheque, thus we check
	// that the balance is *below* the threshold
	if newBalance <= -s.paymentThreshold {
		log.Warn("balance for peer went over the payment threshold, sending cheque", "peer", peer.ID().String(), "payment threshold", s.paymentThreshold)
		swapPeer, ok := s.getPeer(peer.ID())
		if !ok {
			return fmt.Errorf("peer %s not found", peer)
		}
		return s.sendCheque(swapPeer)
	}

	return nil
}

func (s *Swap) getBalance(id enode.ID) (int64, bool) {
	s.balancesLock.RLock()
	defer s.balancesLock.RUnlock()
	peerBalance, exists := s.balances[id]
	return peerBalance, exists
}

func (s *Swap) setBalance(id enode.ID, balance int64) {
	s.balancesLock.Lock()
	defer s.balancesLock.Unlock()
	s.balances[id] = balance
}

func (s *Swap) getCheque(id enode.ID) (*Cheque, bool) {
	s.chequesLock.RLock()
	defer s.chequesLock.RUnlock()
	peerCheque, exists := s.cheques[id]
	return peerCheque, exists
}

func (s *Swap) setCheque(id enode.ID, cheque *Cheque) {
	s.chequesLock.Lock()
	defer s.chequesLock.Unlock()
	s.cheques[id] = cheque
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

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a debitor
func (s *Swap) handleEmitChequeMsg(ctx context.Context, p *Peer, msg *EmitChequeMsg) error {
	cheque := msg.Cheque
	log.Info("received cheque from peer", "peer", p.ID().String())
	actualAmount, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		return err
	}

	log.Debug("received cheque processed and verified", "peer", p.ID().String())

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	s.accountingLock.Lock()
	err = s.resetBalance(p.ID(), 0-int64(cheque.Honey))
	s.accountingLock.Unlock()
	if err != nil {
		return err
	}

	// cash in cheque
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	opts.Context = ctx

	otherSwap, err := contract.InstanceAt(cheque.Contract, s.backend)
	if err != nil {
		return err
	}

	// submit cheque to the blockchain and cashes it directly
	go func() {
		// blocks here, as we are waiting for the transaction to be mined
		receipt, err := otherSwap.SubmitChequeBeneficiary(opts, s.backend, big.NewInt(int64(cheque.Serial)), big.NewInt(int64(cheque.Amount)), big.NewInt(int64(cheque.Timeout)), cheque.Signature)
		if err != nil {
			// TODO: do something with the error
			// and we actually need to log this error as we are in an async routine; nobody is handling this error for now
			log.Error("error submitting cheque", "err", err)
			return
		}
		log.Debug("submit tx mined", "receipt", receipt)

		receipt, err = otherSwap.CashChequeBeneficiary(opts, s.backend, s.owner.Contract, big.NewInt(int64(actualAmount)))
		if err != nil {
			// TODO: do something with the error
			// and we actually need to log this error as we are in an async routine; nobody is handling this error for now
			log.Error("error cashing cheque", "err", err)
			return
		}
		log.Info("Cheque successfully submitted and cashed")
	}()
	return err
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
func (s *Swap) processAndVerifyCheque(cheque *Cheque, p *Peer) (uint64, error) {
	if err := cheque.verifyChequeProperties(p, s.owner.address); err != nil {
		return 0, err
	}

	lastCheque := s.loadLastReceivedCheque(p)

	// TODO: there should probably be a lock here?
	expectedAmount, err := s.oracle.GetPrice(cheque.Honey)
	if err != nil {
		return 0, err
	}

	actualAmount, err := cheque.verifyChequeAgainstLast(lastCheque, expectedAmount)
	if err != nil {
		return 0, err
	}

	if err := s.saveLastReceivedCheque(p, cheque); err != nil {
		log.Error("error while saving last received cheque", "peer", p.ID().String(), "err", err.Error())
		// TODO: what do we do here? Related issue: https://github.com/ethersphere/swarm/issues/1515
	}

	return actualAmount, nil
}

// To be called with mutex already held
// Caller must be careful that the same balance isn't concurrently read and written by multiple routines
func (s *Swap) updateBalance(peer enode.ID, amount int64) (int64, error) {
	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	balance, exists := s.getBalance(peer)
	if !exists {
		return 0, fmt.Errorf("peer %v does not exist", peer)
	}
	newBalance := balance + amount
	s.setBalance(peer, newBalance)
	//save the new balance to the state store
	err := s.store.Put(balanceKey(peer), &newBalance)
	if err != nil {
		return 0, err
	}
	log.Debug("balance for peer after accounting", "peer", peer.String(), "balance", strconv.FormatInt(newBalance, 10))
	return newBalance, err
}

// loadBalance loads balances from the state store (persisted)
// To be called with mutex already held
// Caller must be careful that the same balance isn't concurrently read and written by multiple routines
func (s *Swap) loadBalance(peer enode.ID) (err error) {
	var peerBalance int64
	if _, ok := s.getBalance(peer); !ok {
		err = s.store.Get(balanceKey(peer), &peerBalance)
		s.setBalance(peer, peerBalance)
	}
	return
}

// sendCheque sends a cheque to peer
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (s *Swap) sendCheque(swapPeer *Peer) error {
	peer := swapPeer.ID()
	cheque, err := s.createCheque(swapPeer)
	if err != nil {
		return fmt.Errorf("error while creating cheque: %s", err.Error())
	}

	log.Info("sending cheque", "serial", cheque.ChequeParams.Serial, "amount", cheque.ChequeParams.Amount, "beneficiary", cheque.Beneficiary, "contract", cheque.Contract)
	s.setCheque(peer, cheque)

	err = s.store.Put(sentChequeKey(peer), &cheque)
	if err != nil {
		return fmt.Errorf("error while storing the last cheque: %s", err.Error())
	}

	emit := &EmitChequeMsg{
		Cheque: cheque,
	}

	// reset balance;
	err = s.resetBalance(peer, int64(cheque.Amount))
	if err != nil {
		return err
	}

	return swapPeer.Send(context.Background(), emit)
}

// createCheque creates a new cheque whose beneficiary will be the peer and
// whose serial and amount are set based on the last cheque and current balance for this peer
// The cheque will be signed and point to the issuer's contract
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (s *Swap) createCheque(swapPeer *Peer) (*Cheque, error) {
	var cheque *Cheque
	var err error

	peer := swapPeer.ID()
	beneficiary := swapPeer.beneficiary

	peerBalance, exists := s.getBalance(peer)
	if !exists {
		return nil, fmt.Errorf("peer not found %v: ", peer)
	}
	// the balance should be negative here, we take the absolute value:
	honey := uint64(-peerBalance)

	var amount uint64
	amount, err = s.oracle.GetPrice(honey)
	if err != nil {
		return nil, fmt.Errorf("error getting price from oracle: %s", err.Error())
	}

	// if there is no existing cheque when loading from the store, it means it's the first interaction
	// this is a valid scenario
	err = s.loadLastSentCheque(peer)
	if err != nil && err != state.ErrNotFound {
		return nil, err
	}
	lastCheque, exists := s.getCheque(peer)

	serial := uint64(1)
	if exists {
		cheque = &Cheque{
			ChequeParams: ChequeParams{
				Serial: lastCheque.Serial + serial,
				Amount: lastCheque.Amount + amount,
			},
		}
	} else {
		cheque = &Cheque{
			ChequeParams: ChequeParams{
				Serial: serial,
				Amount: amount,
			},
		}
	}
	cheque.ChequeParams.Timeout = defaultCashInDelay
	cheque.ChequeParams.Contract = s.owner.Contract
	cheque.ChequeParams.Honey = honey
	cheque.Beneficiary = beneficiary

	cheque.Signature, err = cheque.Sign(s.owner.privateKey)

	return cheque, err
}

// Balance returns the balance for a given peer
func (s *Swap) Balance(peer enode.ID) (int64, error) {
	var err error
	peerBalance, ok := s.getBalance(peer)
	if !ok {
		err = s.store.Get(balanceKey(peer), &peerBalance)
	}
	return peerBalance, err
}

// Balances returns the balances for all known SWAP peers
func (s *Swap) Balances() (map[enode.ID]int64, error) {
	balances := make(map[enode.ID]int64)

	s.balancesLock.RLock()
	for peer, peerBalance := range s.balances {
		balances[peer] = peerBalance
	}
	s.balancesLock.RUnlock()

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

// loadLastSentCheque loads the last cheque for a peer from the state store (persisted)
// To be called with mutex already held
// Caller must be careful that the same cheque isn't concurrently read and written by multiple routines
func (s *Swap) loadLastSentCheque(peer enode.ID) (err error) {
	//only load if the current instance doesn't already have this peer's
	//last cheque in memory
	var cheque *Cheque
	if _, ok := s.getCheque(peer); !ok {
		err = s.store.Get(sentChequeKey(peer), &cheque)
		if err == nil {
			s.setCheque(peer, cheque)
		}
	}
	return err
}

// loadLastReceivedCheque gets the last received cheque for the peer
// cheque gets loaded from database if not already in memory
func (s *Swap) loadLastReceivedCheque(p *Peer) (cheque *Cheque) {
	s.accountingLock.Lock()
	defer s.accountingLock.Unlock()
	if p.lastReceivedCheque != nil {
		return p.lastReceivedCheque
	}
	s.store.Get(receivedChequeKey(p.ID()), &cheque)
	return
}

// saveLastReceivedCheque saves cheque as the last received cheque for peer
func (s *Swap) saveLastReceivedCheque(p *Peer, cheque *Cheque) error {
	s.accountingLock.Lock()
	defer s.accountingLock.Unlock()
	p.lastReceivedCheque = cheque
	return s.store.Put(receivedChequeKey(p.ID()), cheque)
}

// Close cleans up swap
func (s *Swap) Close() error {
	return s.store.Close()
}

// resetBalance is called:
// * for the creditor: upon receiving the cheque
// * for the debitor: after sending the cheque
func (s *Swap) resetBalance(peer enode.ID, amount int64) error {
	log.Debug("resetting balance for peer", "peer", peer.String(), "amount", amount)
	_, err := s.updateBalance(peer, amount)
	return err
}

// GetParams returns contract parameters (Bin, ABI) from the contract
func (s *Swap) GetParams() *swap.Params {
	return s.contract.ContractParams()
}

// Deploy deploys a new swap contract
func (s *Swap) Deploy(ctx context.Context, backend swap.Backend, path string) error {
	return s.deploy(ctx, backend, path)
}

// verifyContract checks if the bytecode found at address matches the expected bytecode
func (s *Swap) verifyContract(ctx context.Context, address common.Address) error {
	return contract.ValidateCode(ctx, s.backend, address)
}

// getContractOwner retrieve the owner of the chequebook at address from the blockchain
func (s *Swap) getContractOwner(ctx context.Context, address common.Address) (common.Address, error) {
	contr, err := contract.InstanceAt(address, s.backend)
	if err != nil {
		return common.Address{}, err
	}

	return contr.Issuer(nil)
}

// deploy deploys the Swap contract
func (s *Swap) deploy(ctx context.Context, backend swap.Backend, path string) error {
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
	s.owner.Contract = address
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
