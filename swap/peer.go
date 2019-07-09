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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/p2p/protocols"
)

var SigDoesNotMatch = errors.New("signature does not match")
var DontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	swap *Swap
}

func NewPeer(p *protocols.Peer, s *Swap) *Peer {
	return &Peer{
		Peer: p,
		swap: s,
	}
}

// handleMsg is for handling messages when receiving messages
func (sp *Peer) handleMsg(ctx context.Context, msg interface{}) error {
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

// handleChequeRequestMsg runs when a peer receives a `ChequeRequestMsg`
// It is thus run by the debitor
// So the debitor needs to:
//   * check that it indeed owes to the requestor (if not, ignore message)
//   * check serial number
//   * check amount
//   * if all is ok, issue the cheque
func (sp *Peer) handleChequeRequestMsg(ctx context.Context, msg interface{}) (err error) {
	// check we have indeed a negative balance with the peer
	var req *ChequeRequestMsg
	var ok bool
	var peerBalance int64

	if req, ok = msg.(*ChequeRequestMsg); !ok {
		return fmt.Errorf("Unexpected message type: %v", err)
	}

	peer := req.Peer

	sp.swap.lock.Lock()
	defer sp.swap.lock.Unlock() //TODO: Do we really want to block so long?

	if peerBalance, ok = sp.swap.balances[peer]; !ok {
		return fmt.Errorf("No exchanges with peer: %v", peer)
	}
	// do we actually owe to the remote peer?
	if peerBalance >= 0 {
		return DontOwe
	}

	// balance is negative, send a cheque
	// TODO: merge with thresholds; need to check for threshold?
	// if not, any negative balance will result in a cheque at this point

	var cheque *Cheque

	_ = sp.swap.loadCheque(peer)
	lastCheque := sp.swap.cheques[peer]

	amount := 0 - peerBalance

	//TODO; need to have ChequeRequestMsg to contain last cheque and compare?
	// emit cheque, send to peer
	if lastCheque == nil {
		cheque = &Cheque{
			Serial: uint64(0),
			Amount: uint64(amount),
		}
	} else {
		cheque = &Cheque{
			Serial: lastCheque.Serial + 1,
			Amount: lastCheque.Amount + uint64(0-peerBalance),
		}
	}
	cheque.Timeout = defaultCashInDelay
	cheque.Contract = sp.swap.owner.Contract
	pk, err := crypto.UnmarshalPubkey(req.PubKey)
	if err != nil {
		return err
	}
	cheque.Beneficiary = crypto.PubkeyToAddress(*pk)
	cheque.Sig, err = sp.swap.signContent(cheque)
	if err != nil {
		return err
	}

	sp.swap.cheques[peer] = cheque
	err = sp.swap.stateStore.Put(peer.String()+"_cheques", &cheque)

	// TODO: error handling might be quite more complex
	if err != nil {
		return err
	}

	emit := &EmitChequeMsg{
		Cheque: cheque,
	}

	// TODO: reset balance here?
	// if we don't, then multiple ChequeRequestMsg may be sent and multiple
	// cheques will be generated
	// If we do, then if something goes wrong and the remote does not reset the balance,
	// we have issues as well.
	// For now, reset the balance
	sp.swap.resetBalance(peer)

	return sp.Send(ctx, emit)
}

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a creditor
func (sp *Peer) handleEmitChequeMsg(ctx context.Context, msg interface{}) error {
	chequeMsg, ok := msg.(*EmitChequeMsg)
	if !ok {
		return fmt.Errorf("Invalid message type, %v", msg)
	}
	cheque := chequeMsg.Cheque
	// reset balance to zero
	sp.swap.resetBalance(sp.ID())
	// send confirmation
	sp.Send(ctx, &ConfirmMsg{})
	// cash in cheque
	//TODO: input parameter checks?

	opts := bind.NewKeyedTransactor(sp.swap.owner.privateKey)
	//TODO: ??????
	opts.Value = big.NewInt(int64(cheque.Amount))
	opts.Context = ctx
	sp.swap.contractProxy.Wrapper.SubmitChequeBeneficiary(opts, big.NewInt(int64(cheque.Serial)), big.NewInt(int64(cheque.Amount)), big.NewInt(int64(cheque.Timeout)), cheque.Sig)
	return nil
}

// TODO: Error handling
func (sp *Peer) handleErrorMsg(ctx context.Context, msg interface{}) error {
	// maybe balance disagreement
	return nil
}

func (sp *Peer) handleConfirmMsg(ctx context.Context, msg interface{}) error {
	// TODO; correct here?
	sp.swap.resetBalance(sp.ID())
	return nil
}
