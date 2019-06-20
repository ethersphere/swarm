// Copyright 2019 The go-ethereum Authors
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

	"github.com/ethersphere/swarm/p2p/protocols"
)

type SwapPeer struct {
	*protocols.Peer
	swap *Swap
}

func NewPeer(p *protocols.Peer, s *Swap) *SwapPeer {
	return &SwapPeer{
		Peer: p,
		swap: s,
	}
}

func (sp *SwapPeer) handleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {

	case *ChequeRequestMsg:
		return sp.handleChequeRequestMsg(ctx, msg)

	case *EmitChequeMsg:
		return sp.handleEmitChequeMsg(ctx, msg)

	case *ErrorMsg:
		return sp.handleErrorMsg(ctx, msg)

	case *ConfirmMsg:
		return sp.handleConfirmMsg(ctx, msg)
	}

	return nil
}

func (sp *SwapPeer) handleChequeRequestMsg(ctx context.Context, msg interface{}) error {
	// emit cheque, send to peer
	return sp.Send(ctx, &EmitChequeMsg{})
}

func (sp *SwapPeer) handleEmitChequeMsg(ctx context.Context, msg interface{}) error {
	// reset balance to zero
	sp.swap.resetBalance(sp.Peer)
	// cash in cheque
	return nil
}

func (sp *SwapPeer) handleErrorMsg(ctx context.Context, msg interface{}) error {
	// maybe balance disagreement
	return nil
}

func (sp *SwapPeer) handleConfirmMsg(ctx context.Context, msg interface{}) error {
	return nil
}
