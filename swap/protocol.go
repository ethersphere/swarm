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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethersphere/swarm/p2p/protocols"
)

var (
	// ErrEmptyAddressInSignature is used when the empty address is used for the chequebook in the handshake
	ErrEmptyAddressInSignature = errors.New("empty address in handshake")

	// ErrDifferentChainID is used when the chain id exchanged during the handshake does not match
	ErrDifferentChainID = errors.New("different chain id")

	// ErrInvalidHandshakeMsg is used when the message received during handshake does not conform to the
	// structure of the HandshakeMsg
	ErrInvalidHandshakeMsg = errors.New("invalid handshake message")

	// Spec is the swap protocol specification
	Spec = &protocols.Spec{
		Name:       "swap",
		Version:    1,
		MaxMsgSize: 10 * 1024 * 1024,
		Messages: []interface{}{
			HandshakeMsg{},
			EmitChequeMsg{},
			ConfirmChequeMsg{},
		},
	}
)

// Protocols is a node.Service interface method
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

// Start is a node.Service interface method
func (s *Swap) Start(server *p2p.Server) error {
	log.Info("Swap service started")
	return nil
}

// Stop is a node.Service interface method
func (s *Swap) Stop() error {
	log.Info("Swap service stopping")
	return s.Close()
}

// verifyHandshake verifies the chequebook address and chain id transmitted in the swap handshake
func (s *Swap) verifyHandshake(msg interface{}) error {
	handshake, ok := msg.(*HandshakeMsg)
	if !ok {
		return ErrInvalidHandshakeMsg
	}

	if (handshake.ContractAddress == common.Address{}) {
		return ErrEmptyAddressInSignature
	}

	if handshake.ChainID != s.chainID {
		return ErrDifferentChainID
	}

	return s.chequebookFactory.VerifyContract(handshake.ContractAddress)
}

// run is the actual swap protocol run method
func (s *Swap) run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	protoPeer := protocols.NewPeer(p, rw, Spec)

	handshake, err := protoPeer.Handshake(context.Background(), &HandshakeMsg{
		ContractAddress: s.GetParams().ContractAddress,
		ChainID:         s.chainID,
	}, s.verifyHandshake)
	if err != nil {
		return err
	}

	response, ok := handshake.(*HandshakeMsg)
	if !ok {
		return ErrInvalidHandshakeMsg
	}

	beneficiary, err := s.getContractOwner(context.Background(), response.ContractAddress)
	if err != nil {
		return err
	}

	swapPeer, err := s.addPeer(protoPeer, beneficiary, response.ContractAddress)
	if err != nil {
		return err
	}
	defer s.removePeer(swapPeer)

	return swapPeer.Run(s.handleMsg(swapPeer))
}

func (s *Swap) removePeer(p *Peer) {
	s.peersLock.Lock()
	defer s.peersLock.Unlock()
	delete(s.peers, p.ID())
}

func (s *Swap) addPeer(protoPeer *protocols.Peer, beneficiary common.Address, contractAddress common.Address) (*Peer, error) {
	s.peersLock.Lock()
	defer s.peersLock.Unlock()
	p, err := NewPeer(protoPeer, s, beneficiary, contractAddress)
	if err != nil {
		return nil, err
	}
	s.peers[p.ID()] = p
	return p, nil
}

func (s *Swap) getPeer(id enode.ID) *Peer {
	s.peersLock.RLock()
	defer s.peersLock.RUnlock()
	peer := s.peers[id]
	return peer
}
