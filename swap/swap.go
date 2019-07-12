// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swap

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
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
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

const (
	deployRetries                     = 5
	deployDelay                       = 1 * time.Second // delay between retries
	defaultCashInDelay                = uint64(0)       // Default timeout until cashing in cheques is possible - TODO: deliberate value, experiment // should be non-zero once we implement waivers
	DefaultInitialDepositAmount       = 0               // TODO: deliberate value for now; needs experimentation
	DefaultHarddepositTimeoutDuration = 24 * time.Hour  // this is the amount of time in seconds which an issuer has to wait to decrease the harddeposit of a beneficiary. The smart-contract allows for setting this variable differently per beneficiary
)

// SwAP Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	api                 PublicAPI
	stateStore          state.Store          // stateStore is needed in order to keep balances across sessions
	lock                sync.RWMutex         // lock the balances
	balances            map[enode.ID]int64   // map of balances for each peer
	cheques             map[enode.ID]*Cheque // map of balances for each peer
	backend             cswap.Backend
	owner               *Owner  // contract access
	params              *Params // economic and operational parameters
	contractReference   *swap.Swap
	paymentThreshold    int64 // balance difference required for requesting cheque
	disconnectThreshold int64 // balance difference required for dropping peer
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
	InitialDepositAmount uint64 //
}

func NewDefaultParams() *Params {
	return &Params{
		InitialDepositAmount: DefaultInitialDepositAmount,
	}
}

// New - swap constructor
func New(stateStore state.Store, prvkey *ecdsa.PrivateKey, contract common.Address, backend cswap.Backend) *Swap {
	sw := &Swap{
		stateStore:          stateStore,
		balances:            make(map[enode.ID]int64),
		backend:             backend,
		cheques:             make(map[enode.ID]*Cheque),
		params:              NewDefaultParams(),
		paymentThreshold:    DefaultPaymentThreshold,
		disconnectThreshold: DefaultDisconnectThreshold,
	}
	sw.contractReference = swap.New()
	sw.owner = sw.createOwner(prvkey, contract)
	return sw
}

// createOwner assings keys and addresses
func (s *Swap) createOwner(prvkey *ecdsa.PrivateKey, contract common.Address) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		privateKey: prvkey,
		publicKey:  pubkey,
		Contract:   contract,
		address:    crypto.PubkeyToAddress(*pubkey),
	}
}

// convenience log output
func (s *Swap) DeploySuccess() string {
	return fmt.Sprintf("contract: %s, owner: %s, deposit: %v, signer: %x", s.owner.Contract.Hex(), s.owner.address.Hex(), s.params.InitialDepositAmount, s.owner.publicKey)
}

//Swap implements the protocols.Balance interface
//Add is the (sole) accounting function
func (s *Swap) Add(amount int64, peer *protocols.Peer) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	//load existing balances from the state store
	err = s.loadState(peer)
	if err != nil && err != state.ErrNotFound {
		return
	}

	//check if balance with peer is over the disconnect threshold
	if s.balances[peer.ID()] >= s.disconnectThreshold {
		//if so, return error in order to abort the transfer
		return fmt.Errorf("balance for peer %s went over the disconnect threshold %v", peer.ID().String(), s.disconnectThreshold)
	}

	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	s.balances[peer.ID()] += amount

	//save the new balance to the state store
	peerBalance := s.balances[peer.ID()]
	err = s.stateStore.Put(peer.ID().String(), &peerBalance)
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("balance for peer %s: %s", peer.ID().String(), strconv.FormatInt(peerBalance, 10)))

	//check if balance with peer is over the payment threshold
	if peerBalance >= s.paymentThreshold {
		//if so, send cheque request message
		chequeRequestMessage := &ChequeRequestMsg{
			Beneficiary: s.owner.address,
		}
		return peer.Send(context.TODO(), chequeRequestMessage)
	}

	return
}

//GetPeerBalance returns the balance for a given peer
func (swap *Swap) GetPeerBalance(peer enode.ID) (int64, error) {
	swap.lock.RLock()
	defer swap.lock.RUnlock()
	if p, ok := swap.balances[peer]; ok {
		return p, nil
	}
	return 0, errors.New("Peer not found")
}

func (swap *Swap) GetLastCheque(peer enode.ID) (*Cheque, error) {
	swap.lock.RLock()
	defer swap.lock.RUnlock()

	if lc, ok := swap.cheques[peer]; ok {
		return lc, nil
	}

	return nil, errors.New("Peer not found")
}

//GetAllBalances returns the balances for all known peers
func (swap *Swap) GetAllBalances() map[enode.ID]int64 {
	swap.lock.RLock()
	defer swap.lock.RUnlock()
	return swap.balances
}

//load balances from the state store (persisted)
func (s *Swap) loadState(peer *protocols.Peer) (err error) {
	var peerBalance int64
	peerID := peer.ID()
	//only load if the current instance doesn't already have this peer's
	//balance in memory
	if _, ok := s.balances[peerID]; !ok {
		err = s.stateStore.Get(peerID.String(), &peerBalance)
		s.balances[peerID] = peerBalance
	}
	return
}

//load last cheque for a peer from the state store (persisted)
func (s *Swap) loadCheque(peer enode.ID) (err error) {
	//only load if the current instance doesn't already have this peer's
	//last cheque in memory
	var cheque *Cheque
	if _, ok := s.cheques[peer]; !ok {
		err = s.stateStore.Get(peer.String()+"_cheques", &cheque)
		s.cheques[peer] = cheque
	}
	return
}

//Clean up Swap
func (swap *Swap) Close() {
	swap.stateStore.Close()
}

// resetBalance is called:
// * for the creditor: on cheque receival
// * for the debitor: on confirmation receival
func (s *Swap) resetBalance(peerID enode.ID) {
	//TODO: reset balance based on actual amount
	//Lock() is not applied because it is expected to be done in the handleChequeRequestMsg call
	s.balances[peerID] = 0
}

// encodeCheque encodes the cheque in the format used in the signing procedure
func (s *Swap) encodeCheque(cheque *Cheque) []byte {
	serialBytes := make([]byte, 32)
	amountBytes := make([]byte, 32)
	timeoutBytes := make([]byte, 32)
	// we need to write the last 8 bytes as we write a uint64 into a 32-byte array
	// encoded in BigEndian because EVM uses BigEndian encoding
	binary.BigEndian.PutUint64(serialBytes[24:], cheque.Serial)
	binary.BigEndian.PutUint64(amountBytes[24:], cheque.Amount)
	binary.BigEndian.PutUint64(timeoutBytes[24:], cheque.Timeout)
	// construct the actual cheque
	input := cheque.Contract.Bytes()
	input = append(input, cheque.Beneficiary.Bytes()...)
	input = append(input, serialBytes[:]...)
	input = append(input, amountBytes[:]...)
	input = append(input, timeoutBytes[:]...)

	return input
}

// sigHashCheque hashes the cheque using the prefix that would be added by eth_Sign
func (s *Swap) sigHashCheque(cheque *Cheque) []byte {
	input := crypto.Keccak256(s.encodeCheque(cheque))
	withPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(input), input)
	return crypto.Keccak256([]byte(withPrefix))
}

// signContent signs the cheque
func (s *Swap) signContent(cheque *Cheque) ([]byte, error) {
	sig, err := crypto.Sign(s.sigHashCheque(cheque), s.owner.privateKey)
	if err != nil {
		return nil, err
	}
	// increase the v value by 27 as crypto.Sign produces 0 or 1 but the contract only accepts 27 or 28
	// this is to prevent malleable signatures. while not strictly necessary in this case the ECDSA implementation from Openzeppelin expects it.
	sig[len(sig)-1] += 27
	return sig, nil
}

// GetParams returns contract parameters (Bin, ABI) from the contract
func (s *Swap) GetParams() *swap.Params {
	return s.contractReference.ContractParams()
}

func (s *Swap) Deploy(ctx context.Context, backend swap.Backend, path string) error {
	//TODO do we need this check?
	_, err := s.contractReference.ValidateCode(ctx, backend, s.owner.Contract)
	if err != nil {
		return err
	}

	// TODO: What to do if the contract is already deployed?
	return s.deploy(ctx, backend, path)
}

// deploy deploys the Swap contract
func (s *Swap) deploy(ctx context.Context, backend swap.Backend, path string) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	// initial topup value
	opts.Value = big.NewInt(int64(s.params.InitialDepositAmount))
	opts.Context = ctx

	log.Info(fmt.Sprintf("Deploying new swap (owner: %v)", opts.From.Hex()))
	address, err := s.deployLoop(opts, backend, s.owner.address, big.NewInt(int64(DefaultHarddepositTimeoutDuration)))
	if err != nil {
		log.Error(fmt.Sprintf("unable to deploy swap: %v", err))
		return err
	}
	s.owner.Contract = address
	log.Info(fmt.Sprintf("swap deployed at %v (owner: %v)", address.Hex(), opts.From.Hex()))

	return err
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, backend swap.Backend, owner common.Address, harddepositTimeout *big.Int) (addr common.Address, err error) {
	var tx *types.Transaction
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}
		if _, tx, err = s.contractReference.Deploy(opts, backend, owner, harddepositTimeout); err != nil {
			log.Warn(fmt.Sprintf("can't send chequebook deploy tx (try %d): %v", try, err))
			continue
		}
		if addr, err = bind.WaitDeployed(opts.Context, backend, tx); err != nil {
			log.Warn(fmt.Sprintf("chequebook deploy error (try %d): %v", try, err))
			continue
		}
		return addr, nil
	}
	return addr, err
}
