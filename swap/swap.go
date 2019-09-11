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
	"github.com/ethereum/go-ethereum/log"
	l "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/swap"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/mattn/go-colorable"
)

// ErrInvalidChequeSignature indicates the signature on the cheque was invalid
var ErrInvalidChequeSignature = errors.New("invalid cheque signature")

// Swap represents the Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	api                 API
	logger              l.Logger           // logger for Swap related messages and audit trail
	store               state.Store        // store is needed in order to keep balances and cheques across sessions
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

// NewLogger returns a new logger
func NewLogger(h interface{}) log.Logger {
	swapLogger := log.New("swaplog", "*")
	hd, ok := h.(log.Handler)

	//TODO: replace with default handler
	if !ok {
		hd = log.LvlFilterHandler(log.Lvl(2), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true)))
	}

	swapLogger.SetHandler(hd)
	return swapLogger
}

// New - swap constructor
func New(logHandler interface{}, stateStore state.Store, prvkey *ecdsa.PrivateKey, contract common.Address, backend contract.Backend) *Swap {
	return &Swap{
		logger:              NewLogger(logHandler),
		store:               stateStore,
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
	swapPeer := s.getPeer(peer.ID())
	if swapPeer == nil {
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
		s.logger.Warn("balance for peer went over the payment threshold, sending cheque", "peer", peer.ID().String(), "payment threshold", s.paymentThreshold)
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
	s.logger.Info("received cheque from peer", "peer", p.ID().String(), "honey", cheque.Honey)
	_, err := s.processAndVerifyCheque(cheque, p)
	if err != nil {
		return err
	}

	s.logger.Debug("received cheque processed and verified", "peer", p.ID().String())

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	if err := p.updateBalance(-int64(cheque.Honey)); err != nil {
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
		s.logger.Error("error cashing cheque", "err", err)
		return
	}

	if result.Bounced {
		s.logger.Error("cheque bounced", "tx", receipt.TxHash)
		return
		// TODO: do something here
	}

	s.logger.Debug("cash tx mined", "receipt", receipt)
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
		s.logger.Error("error while saving last received cheque", "peer", p.ID().String(), "err", err.Error())
		// TODO: what do we do here? Related issue: https://github.com/ethersphere/swarm/issues/1515
	}

	return actualAmount, nil
}

// Balance returns the balance for a given peer
func (s *Swap) Balance(peer enode.ID) (int64, error) {
	swapPeer := s.getPeer(peer)
	if swapPeer == nil {
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
	var cheque *Cheque
	error := s.store.Get(receivedChequeKey(p), &cheque)
	if error == state.ErrNotFound {
		return nil, nil
	}
	return cheque, error
}

// loadLastSentCheque loads the last sent cheque for the peer from the store
func (s *Swap) loadLastSentCheque(p enode.ID) (*Cheque, error) {
	var cheque *Cheque
	error := s.store.Get(sentChequeKey(p), &cheque)
	if error == state.ErrNotFound {
		return nil, nil
	}
	return cheque, error
}

// loadBalance loads the current balance for the peer from the store
func (s *Swap) loadBalance(p enode.ID) (int64, error) {
	var balance int64
	error := s.store.Get(balanceKey(p), &balance)
	if error == state.ErrNotFound {
		return 0, nil
	}
	return balance, error
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

// GetParams returns contract parameters (Bin, ABI) from the contract
func (s *Swap) GetParams() *swap.Params {
	return s.contract.ContractParams()
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

// Deploy deploys the Swap contract
func (s *Swap) Deploy(ctx context.Context, backend swap.Backend, path string) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(s.params.InitialDepositAmount))
	opts.Context = ctx

	s.logger.Info("deploying new swap", "owner", opts.From.Hex())
	address, err := s.deployLoop(opts, backend, s.owner.address, defaultHarddepositTimeoutDuration)
	if err != nil {
		s.logger.Error("unable to deploy swap", "error", err)
		return err
	}
	s.owner.Contract = address
	s.logger.Info("swap deployed", "address", address.Hex(), "owner", opts.From.Hex())

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
			s.logger.Warn("can't send chequebook deploy tx", "try", try, "error", err)
			continue
		}
		if addr, err = bind.WaitDeployed(opts.Context, backend, tx); err != nil {
			s.logger.Warn("chequebook deploy error", "try", try, "error", err)
			continue
		}
		return addr, nil
	}
	return addr, err
}
