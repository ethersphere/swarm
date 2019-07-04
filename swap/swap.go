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
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

const (
	deployRetries               = 5
	deployDelay                 = 1 * time.Second // delay between retries
	defaultCashInDelay          = uint64(0)
	DefaultInitialDepositAmount = 100000000 // TODO: deliberate value for now; needs experimentation
)

// SwAP Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	stateStore    state.Store        //stateStore is needed in order to keep balances across sessions
	lock          sync.RWMutex       //lock the balances
	balances      map[enode.ID]int64 //map of balances for each peer
	Service       *SwapService
	owner         *Owner
	params        *SwapParams
	contractProxy *swap.Proxy
}

type Owner struct {
	Contract   common.Address // address of chequebook contract
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
	lock       sync.RWMutex
}

type SwapParams struct {
	InitialDepositAmount uint64 //
}

func NewDefaultSwapParams() *SwapParams {
	return &SwapParams{
		InitialDepositAmount: DefaultInitialDepositAmount,
	}
}

// New - swap constructor
func New(stateStore state.Store, prvkey *ecdsa.PrivateKey, contract common.Address) (swap *Swap) {
	swap = &Swap{
		stateStore: stateStore,
		balances:   make(map[enode.ID]int64),
		params:     NewDefaultSwapParams(),
	}
	swap.contractProxy = swap.createProxy()
	swap.owner = swap.createOwner(prvkey, contract)
	return
}

func (s *Swap) createOwner(prvkey *ecdsa.PrivateKey, contract common.Address) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		privateKey: prvkey,
		publicKey:  pubkey,
		address:    crypto.PubkeyToAddress(*pubkey),
	}
}

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
	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	s.balances[peer.ID()] += amount
	//save the new balance to the state store
	peerBalance := s.balances[peer.ID()]
	err = s.stateStore.Put(peer.ID().String(), &peerBalance)

	log.Debug(fmt.Sprintf("balance for peer %s: %s", peer.ID().String(), strconv.FormatInt(peerBalance, 10)))
	return err
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

//Clean up Swap
func (swap *Swap) Close() {
	swap.stateStore.Close()
}

func (s *Swap) resetBalance(peer *protocols.Peer) {
	s.lock.Lock()
	defer s.lock.Unlock()

	//TODO: reset balance based on actual amount
	s.balances[peer.ID()] = 0
}

func (s *Swap) createProxy() *swap.Proxy {
	return swap.NewProxy()
}

func (s *Swap) GetParams() *swap.Params {
	return s.contractProxy.ContractParams()
}

// SetChequebook wraps the chequebook initialiser and sets up autoDeposit to cover spending.
func (s *Swap) Deploy(ctx context.Context, backend swap.Backend, path string) error {

	//TODO do we need this check?
	_, err := s.contractProxy.ValidateCode(ctx, backend, s.owner.Contract)
	if err != nil {
		return err
	}
	return s.deploy(ctx, backend, path)
}

// deploy deploys the Swap contract
func (s *Swap) deploy(ctx context.Context, backend swap.Backend, path string) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	opts.Value = big.NewInt(int64(s.params.InitialDepositAmount))
	opts.Context = ctx

	log.Info(fmt.Sprintf("Deploying new swap (owner: %v)", opts.From.Hex()))
	address, err := s.deployLoop(opts, backend, s.owner.Contract)
	if err != nil {
		log.Error(fmt.Sprintf("unable to deploy swap: %v", err))
		return err
	}
	log.Info(fmt.Sprintf("swap deployed at %v (owner: %v)", address.Hex(), opts.From.Hex()))

	return err
}

// deployLoop repeatedly tries to deploy the swap contract .
func (s *Swap) deployLoop(opts *bind.TransactOpts, backend swap.Backend, owner common.Address) (addr common.Address, err error) {
	var tx *types.Transaction
	for try := 0; try < deployRetries; try++ {
		if try > 0 {
			time.Sleep(deployDelay)
		}
		if _, tx, err = s.contractProxy.Deploy(opts, backend, owner); err != nil {
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
