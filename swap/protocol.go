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
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
)

var ErrEmptyAddressInSignature = errors.New("empty address in signature")

var Spec = &protocols.Spec{
	Name:       "swap",
	Version:    1,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		SwapHandshakeMsg{},
		EmitChequeMsg{},
		ErrorMsg{},
		ConfirmMsg{},
	},
}

func (s *Swap) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    Spec.Name,
			Version: Spec.Version,
			Length:  Spec.Length(),
			Run:     s.run,
		},
	}
}

func (s *Swap) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "swap",
			Version:   "1.0",
			Service:   s.api,
			Public:    false,
		},
	}
}

func (s *Swap) Start(server *p2p.Server) error {
	log.Info("Swap service started")
	return nil
}

func (s *Swap) Stop() error {
	return nil
}

func (s *Swap) verifyHandshake(msg interface{}) error {
	handshake, ok := msg.(*SwapHandshakeMsg)
	var empty common.Address
	if !ok || handshake.Beneficiary == empty || handshake.ContractAddress == empty {
		return ErrEmptyAddressInSignature
	}

	return s.verifyContract(context.TODO(), handshake.ContractAddress)
}

func (s *Swap) run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	protoPeer := protocols.NewPeer(p, rw, Spec)

	answer, err := protoPeer.Handshake(context.TODO(), &SwapHandshakeMsg{
		Beneficiary:     s.owner.address,
		ContractAddress: s.owner.Contract,
	}, s.verifyHandshake)

	if err != nil {
		return err
	}

	swapPeer := NewPeer(protoPeer, s, s.backend, answer.(*SwapHandshakeMsg).Beneficiary, answer.(*SwapHandshakeMsg).ContractAddress)
	s.peers[p.ID()] = swapPeer

	s.logBalance(protoPeer)

	return swapPeer.Run(swapPeer.handleMsg)
}

type PublicAPI struct {
}
