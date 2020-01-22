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
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/uint256"
)

// ErrDontOwe indictates that no balance is actially owned
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	lock               sync.RWMutex
	swap               *Swap
	beneficiary        common.Address // address of the peers chequebook owner
	contractAddress    common.Address // address of the peers chequebook
	lastReceivedCheque *Cheque        // last cheque we received from the peer
	lastSentCheque     *Cheque        // last cheque that was sent to peer that was confirmed
	pendingCheque      *Cheque        // last cheque that was sent to peer but is not yet confirmed
	balance            int64          // current balance of the peer
	logger             log.Logger     // logger for swap related messages and audit trail with peer identifier
}

// NewPeer creates a new swap Peer instance
func NewPeer(p *protocols.Peer, s *Swap, beneficiary common.Address, contractAddress common.Address) (peer *Peer, err error) {
	peer = &Peer{
		Peer:            p,
		swap:            s,
		beneficiary:     beneficiary,
		contractAddress: contractAddress,
		logger:          newPeerLogger(s, p.ID()),
	}

	if peer.lastReceivedCheque, err = s.loadLastReceivedCheque(p.ID()); err != nil {
		return nil, err
	}

	if peer.lastSentCheque, err = s.loadLastSentCheque(p.ID()); err != nil {
		return nil, err
	}

	if peer.balance, err = s.loadBalance(p.ID()); err != nil {
		return nil, err
	}

	if peer.pendingCheque, err = s.loadPendingCheque(p.ID()); err != nil {
		return nil, err
	}

	return peer, nil
}

// getLastReceivedCheque returns the last cheque we received for this peer
// the caller is expected to hold p.lock
func (p *Peer) getLastReceivedCheque() *Cheque {
	return p.lastReceivedCheque
}

// getLastSentCheque returns the last cheque we sent and got confirmed for this peer
// the caller is expected to hold p.lock
func (p *Peer) getLastSentCheque() *Cheque {
	return p.lastSentCheque
}

// getPendingCheque returns the last cheque we sent but that is not yet confirmed for this peer
// the caller is expected to hold p.lock
func (p *Peer) getPendingCheque() *Cheque {
	return p.pendingCheque
}

// setLastReceivedCheque sets the given cheque as the last one received from this peer
// the caller is expected to hold p.lock
func (p *Peer) setLastReceivedCheque(cheque *Cheque) error {
	p.lastReceivedCheque = cheque
	return p.swap.saveLastReceivedCheque(p.ID(), cheque)
}

// setLastReceivedCheque sets the given cheque as the last sent cheque for this peer
// the caller is expected to hold p.lock
func (p *Peer) setLastSentCheque(cheque *Cheque) error {
	p.lastSentCheque = cheque
	return p.swap.saveLastSentCheque(p.ID(), cheque)
}

// setLastReceivedCheque sets the given cheque as the pending cheque for this peer
// the caller is expected to hold p.lock
func (p *Peer) setPendingCheque(cheque *Cheque) error {
	p.pendingCheque = cheque
	return p.swap.savePendingCheque(p.ID(), cheque)
}

// getLastSentCumulativePayout returns the cumulative payout of the last sent cheque or 0 if there is none
// the caller is expected to hold p.lock
func (p *Peer) getLastSentCumulativePayout() *uint256.Uint256 {
	lastCheque := p.getLastSentCheque()
	if lastCheque != nil {
		return lastCheque.CumulativePayout
	}
	return uint256.New()
}

// the caller is expected to hold p.lock
func (p *Peer) setBalance(balance int64) error {
	p.balance = balance
	return p.swap.saveBalance(p.ID(), balance)
}

// getBalance returns the current balance for this peer
// the caller is expected to hold p.lock
func (p *Peer) getBalance() int64 {
	return p.balance
}

// the caller is expected to hold p.lock
func (p *Peer) updateBalance(amount int64) error {
	//adjust the balance
	//if amount is negative, it will decrease, otherwise increase
	newBalance := p.getBalance() + amount
	if err := p.setBalance(newBalance); err != nil {
		return err
	}
	p.logger.Debug("updated balance", "balance", strconv.FormatInt(newBalance, 10))
	return nil
}

// createCheque creates a new cheque whose beneficiary will be the peer and
// whose amount is based on the last cheque and current balance for this peer
// The cheque will be signed and point to the issuer's contract
// the caller is expected to hold p.lock
func (p *Peer) createCheque() (*Cheque, error) {
	var cheque *Cheque
	var err error

	if p.getBalance() >= 0 {
		return nil, fmt.Errorf("expected negative balance, found: %d", p.getBalance())
	}
	// the balance should be negative here, we take the absolute value:
	honey := uint64(-p.getBalance())

	oraclePrice, err := p.swap.honeyPriceOracle.GetPrice(honey)
	if err != nil {
		return nil, fmt.Errorf("error getting price from oracle: %v", err)
	}
	price := uint256.FromUint64(oraclePrice)

	cumulativePayout := p.getLastSentCumulativePayout()
	newCumulativePayout, err := uint256.New().Add(cumulativePayout, price)
	if err != nil {
		return nil, err
	}

	cheque = &Cheque{
		ChequeParams: ChequeParams{
			CumulativePayout: newCumulativePayout,
			Contract:         p.swap.GetParams().ContractAddress,
			Beneficiary:      p.beneficiary,
		},
		Honey: honey,
	}
	cheque.Signature, err = cheque.Sign(p.swap.owner.privateKey)

	return cheque, err
}

// sendCheque creates and sends a cheque to peer
// if there is already a pending cheque it will resend that one
// otherwise it will create a new cheque and save it as the pending cheque
// the caller is expected to hold p.lock
func (p *Peer) sendCheque() error {
	if p.getPendingCheque() != nil {
		p.logger.Info("previous cheque still pending, resending cheque", "pending", p.getPendingCheque())
		return p.Send(context.Background(), &EmitChequeMsg{
			Cheque: p.getPendingCheque(),
		})
	}
	cheque, err := p.createCheque()
	if err != nil {
		return fmt.Errorf("error while creating cheque: %v", err)
	}

	err = p.setPendingCheque(cheque)
	if err != nil {
		return fmt.Errorf("error while saving pending cheque: %v", err)
	}

	honeyAmount := int64(cheque.Honey)
	err = p.updateBalance(honeyAmount)
	if err != nil {
		return fmt.Errorf("error while updating balance: %v", err)
	}

	metrics.GetOrRegisterCounter("swap.cheques.emitted.num", nil).Inc(1)
	metrics.GetOrRegisterCounter("swap.cheques.emitted.honey", nil).Inc(honeyAmount)

	p.logger.Info("sending cheque to peer", "cheque", cheque)
	return p.Send(context.Background(), &EmitChequeMsg{
		Cheque: cheque,
	})
}
