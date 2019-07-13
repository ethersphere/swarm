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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
)

var ErrSigDoesNotMatch = errors.New("signature does not match")
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	swap            *Swap
	backend         cswap.Backend
	beneficiary     common.Address
	contractAddress common.Address
}

func NewPeer(p *protocols.Peer, s *Swap, backend cswap.Backend, beneficiary common.Address, contractAddress common.Address) *Peer {
	return &Peer{
		Peer:            p,
		swap:            s,
		backend:         backend,
		beneficiary:     beneficiary,
		contractAddress: contractAddress,
	}
}

// handleMsg is for handling messages when receiving messages
func (sp *Peer) handleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {

	case *EmitChequeMsg:
		//return sp.handleEmitChequeMsg(ctx, msg)
		return sp.handleEmitChequeMsg(ctx, msg)

	case *ErrorMsg:
		return sp.handleErrorMsg(ctx, msg)

	case *ConfirmMsg:
		return sp.handleConfirmMsg(ctx, msg)
	}

	return nil
}

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a creditor
// TODO: validate the contract address in the cheque to match the address given at handshake
// TODO: this should not be blocking
func (sp *Peer) handleEmitChequeMsg(ctx context.Context, msg interface{}) error {
	log.Info("received emit cheque message")

	chequeMsg, ok := msg.(*EmitChequeMsg)
	if !ok {
		return fmt.Errorf("Invalid message type, %v", msg)
	}
	cheque := chequeMsg.Cheque
	// reset balance to zero, TODO: fix
	sp.swap.resetBalance(sp.ID())
	// send confirmation
	err := sp.Send(ctx, &ConfirmMsg{})
	if err != nil {
		log.Error(fmt.Sprintf("error while sending confirm msg to peer %s: %s", sp.ID().String(), err.Error()))
	}
	// cash in cheque
	//TODO: input parameter checks?
	opts := bind.NewKeyedTransactor(sp.swap.owner.privateKey)
	opts.Context = ctx

	//TODO: make instanceAt to directly return a swap type

	otherSwap, err := cswap.InstanceAt(cheque.Contract, sp.backend)
	if err != nil {
		log.Error("Could not get an instance of simpleSwap")
		return err
	}

	// submit cheque to the blockchain and cashes it directly
	go func() {
		// blocks here, as we are waiting for the transaction to be mined
		receipt, err := otherSwap.SubmitChequeBeneficiary(opts, sp.backend, big.NewInt(int64(cheque.Serial)), big.NewInt(int64(cheque.Amount)), big.NewInt(int64(cheque.Timeout)), cheque.Sig)
		if err != nil {
			log.Error("Got error when calling submitChequeBeneficiary: ", err)
			//TODO: do something with the error
		}
		log.Info(fmt.Sprintf("tx minded: %v", receipt))
		//TODO: cashCheque
		//TODO: after the cashCheque is done, we have to watch the blockchain for x amount (25) blocks for reorgs
		//TODO: make sure we make a case where we listen to the possibiliyt of the peer shutting down.
	}()
	return err
}

// TODO: Error handling
func (sp *Peer) handleErrorMsg(ctx context.Context, msg interface{}) error {
	log.Info("received error msg")
	// maybe balance disagreement
	return nil
}

func (sp *Peer) handleConfirmMsg(ctx context.Context, msg interface{}) error {
	log.Info("received confirm msg")
	return nil
}
