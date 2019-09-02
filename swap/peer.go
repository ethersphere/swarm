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
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethersphere/swarm/p2p/protocols"
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
	return p.swap.saveLastReceivedCheque(p, cheque)
}

func (p *Peer) setLastSentCheque(cheque *Cheque) error {
	p.lastSentCheque = cheque
	return p.swap.saveLastSentCheque(p, cheque)
}

func (p *Peer) getLastChequeValues() (total uint64, err error) {
	lastCheque := p.getLastReceivedCheque()
	if lastCheque != nil {
		total = lastCheque.CumulativePayout
	}
	return
}
