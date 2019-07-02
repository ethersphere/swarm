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
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/contracts/chequebook"
	"github.com/ethersphere/swarm/contracts/chequebook/contract"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

const (
	chequebookDeployRetries = 5
	chequebookDeployDelay   = 1 * time.Second // delay between retries
	defaultCashInDelay      = uint64(0)
)

// SwAP Swarm Accounting Protocol
// a peer to peer micropayment system
// A node maintains an individual balance with every peer
// Only messages which have a price will be accounted for
type Swap struct {
	stateStore state.Store        //stateStore is needed in order to keep balances across sessions
	lock       sync.RWMutex       //lock the balances
	balances   map[enode.ID]int64 //map of balances for each peer
	service    *SwapService
	owner      *Owner
}

type Owner struct {
	Contract   common.Address // address of chequebook contract
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
	chbook     *chequebook.Chequebook
	lock       sync.RWMutex
}

// New - swap constructor
func New(stateStore state.Store, contract common.Address, prvkey *ecdsa.PrivateKey) (swap *Swap) {
	swap = &Swap{
		stateStore: stateStore,
		balances:   make(map[enode.ID]int64),
	}
	swap.owner = swap.initOwner(contract, prvkey)
	return
}

func (s *Swap) initOwner(contract common.Address, prvkey *ecdsa.PrivateKey) *Owner {
	pubkey := &prvkey.PublicKey
	return &Owner{
		Contract:   contract,
		privateKey: prvkey,
		publicKey:  pubkey,
		address:    crypto.PubkeyToAddress(*pubkey),
	}
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

	s.balances[peer.ID()] = 0
}

// SetChequebook wraps the chequebook initialiser and sets up autoDeposit to cover spending.
func (s *Swap) SetChequebook(ctx context.Context, backend chequebook.Backend, path string) error {

	valid, err := chequebook.ValidateCode(ctx, backend, s.owner.Contract)
	if err != nil {
		return err
	} else if valid {
		return s.newChequebookFromContract(path, backend)
	}
	return s.deployChequebook(ctx, backend, path)
}

// deployChequebook deploys the localProfile Chequebook
func (s *Swap) deployChequebook(ctx context.Context, backend chequebook.Backend, path string) error {
	opts := bind.NewKeyedTransactor(s.owner.privateKey)
	//TODO: opts.Value = lp.AutoDepositBuffer
	opts.Context = ctx

	log.Info(fmt.Sprintf("Deploying new chequebook (owner: %v)", opts.From.Hex()))
	address, err := deployChequebookLoop(opts, backend)
	if err != nil {
		log.Error(fmt.Sprintf("unable to deploy new chequebook: %v", err))
		return err
	}
	log.Info(fmt.Sprintf("new chequebook deployed at %v (owner: %v)", address.Hex(), opts.From.Hex()))

	// need to save config at this point
	s.owner.lock.Lock()
	s.owner.Contract = address
	err = s.newChequebookFromContract(path, backend)
	s.owner.lock.Unlock()
	if err != nil {
		log.Warn(fmt.Sprintf("error initialising cheque book (owner: %v): %v", opts.From.Hex(), err))
	}
	return err
}

// deployChequebookLoop repeatedly tries to deploy a chequebook.
func deployChequebookLoop(opts *bind.TransactOpts, backend chequebook.Backend) (addr common.Address, err error) {
	var tx *types.Transaction
	for try := 0; try < chequebookDeployRetries; try++ {
		if try > 0 {
			time.Sleep(chequebookDeployDelay)
		}
		if _, tx, _, err = contract.DeployChequebook(opts, backend); err != nil {
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

// newChequebookFromContract - initialise the chequebook from a persisted json file or create a new one
// caller holds the lock
func (s *Swap) newChequebookFromContract(path string, backend chequebook.Backend) error {
	hexkey := common.Bytes2Hex(s.owner.Contract.Bytes())
	err := os.MkdirAll(filepath.Join(path, "chequebooks"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory for chequebooks: %v", err)
	}

	chbookpath := filepath.Join(path, "chequebooks", hexkey+".json")
	s.owner.chbook, err = chequebook.LoadChequebook(chbookpath, s.owner.privateKey, backend, true)

	if err != nil {
		s.owner.chbook, err = chequebook.NewChequebook(chbookpath, s.owner.Contract, s.owner.privateKey, backend)
		if err != nil {
			log.Warn(fmt.Sprintf("unable to initialise chequebook (owner: %v): %v", s.owner.address.Hex(), err))
			return fmt.Errorf("unable to initialise chequebook (owner: %v): %v", s.owner.address.Hex(), err)
		}
	}

	//TODO: initial topup contract
	//TODO: automatic topup
	//TODO: external topup?

	return nil
}
