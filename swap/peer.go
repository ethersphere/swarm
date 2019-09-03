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

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethersphere/swarm/p2p/protocols"
)

// ErrDontOwe indictates that no balance is actially owned
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	swap                   *Swap
	beneficiary            common.Address
	contractAddress        common.Address
	lastReceivedCheque     *Cheque
	honeyOracle            HoneyOracle     // the oracle providing the ether price for honey
	paymentThresholdOracle ThresholdOracle // the oracle providing the payment treshold
	disconnectThreshold    int64           // balance difference required for dropping peer
}

// NewPeer creates a new swap Peer instance
func NewPeer(p *protocols.Peer, s *Swap, beneficiary common.Address, contractAddress common.Address) *Peer {
	return &Peer{
		Peer:                   p,
		swap:                   s,
		beneficiary:            beneficiary,
		contractAddress:        contractAddress,
		paymentThresholdOracle: NewThresholdOracle(),
		disconnectThreshold:    DefaultDisconnectThreshold,
		honeyOracle:            NewHoneyPriceOracle(),
	}
}
