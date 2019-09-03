// Copyright 2019 The Swarm Authors
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
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

// ErrDontOwe indictates that no balance is actially owned
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	lock               sync.RWMutex
	swap               *Swap
	beneficiary        common.Address
	contractAddress    common.Address
	lastReceivedCheque *Cheque
	lastSentCheque     *Cheque
	balance            int64
}

// NewPeer creates a new swap Peer instance
func NewPeer(p *protocols.Peer, s *Swap, beneficiary common.Address, contractAddress common.Address) *Peer {
	peer := &Peer{
		Peer:            p,
		swap:            s,
		beneficiary:     beneficiary,
		contractAddress: contractAddress,
	}

	peer.lastReceivedCheque, _ = s.loadLastReceivedCheque(p.ID())
	peer.lastSentCheque, _ = s.loadLastSentCheque(p.ID())
	peer.balance, _ = s.loadBalance(p.ID())

	return peer
}

func (p *Peer) getLastReceivedCheque() *Cheque {
	return p.lastReceivedCheque
}

func (p *Peer) getLastSentCheque() *Cheque {
	return p.lastSentCheque
}

func (p *Peer) setLastReceivedCheque(cheque *Cheque) error {
	p.lastReceivedCheque = cheque
	return p.swap.saveLastReceivedCheque(p.ID(), cheque)
}

func (p *Peer) setLastSentCheque(cheque *Cheque) error {
	p.lastSentCheque = cheque
	return p.swap.saveLastSentCheque(p.ID(), cheque)
}

func (p *Peer) getLastChequeValues() (total uint64, err error) {
	lastCheque := p.getLastReceivedCheque()
	if lastCheque != nil {
		total = lastCheque.CumulativePayout
	}
	return
}

func (p *Peer) setBalance(balance int64) error {
	p.balance = balance
	return p.swap.saveBalance(p.ID(), balance)
}

func (p *Peer) getBalance() int64 {
	return p.balance
}

// To be called with mutex already held
func (p *Peer) updateBalance(amount int64) error {
	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	newBalance := p.getBalance() + amount
	if err := p.setBalance(newBalance); err != nil {
		return err
	}
	log.Debug("balance for peer after accounting", "peer", p.ID().String(), "balance", strconv.FormatInt(newBalance, 10))
	return nil
}

// createCheque creates a new cheque whose beneficiary will be the peer and
// whose amount is based on the last cheque and current balance for this peer
// The cheque will be signed and point to the issuer's contract
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (p *Peer) createCheque() (*Cheque, error) {
	var cheque *Cheque
	var err error

	beneficiary := p.beneficiary
	peerBalance := p.getBalance()
	// the balance should be negative here, we take the absolute value:
	honey := uint64(-peerBalance)
	var amount uint64

	// TODO: this must probably be locked
	amount, err = p.swap.oracle.GetPrice(honey)
	if err != nil {
		return nil, fmt.Errorf("error getting price from oracle: %s", err.Error())
	}

	// if there is no existing cheque when loading from the store, it means it's the first interaction
	// this is a valid scenario
	total, err := p.getLastChequeValues()
	if err != nil && err != state.ErrNotFound {
		return nil, err
	}

	// TODO: lock
	contract := p.swap.owner.Contract

	cheque = &Cheque{
		ChequeParams: ChequeParams{
			CumulativePayout: total + amount,
			Contract:         contract,
			Beneficiary:      beneficiary,
		},
		Honey: honey,
	}
	cheque.Signature, err = cheque.Sign(p.swap.owner.privateKey)

	return cheque, err
}

// sendCheque sends a cheque to peer
// To be called with mutex already held
// Caller must be careful that the same resources aren't concurrently read and written by multiple routines
func (p *Peer) sendCheque() error {
	cheque, err := p.createCheque()
	if err != nil {
		return fmt.Errorf("error while creating cheque: %s", err.Error())
	}

	log.Info("sending cheque", "honey", cheque.Honey, "cumulativePayout", cheque.ChequeParams.CumulativePayout, "beneficiary", cheque.Beneficiary, "contract", cheque.Contract)

	if err := p.setLastSentCheque(cheque); err != nil {
		return fmt.Errorf("error while storing the last cheque: %s", err.Error())
	}

	emit := &EmitChequeMsg{
		Cheque: cheque,
	}

	if err := p.updateBalance(int64(cheque.Honey)); err != nil {
		return err
	}

	return p.Send(context.Background(), emit)
}
