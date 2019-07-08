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
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/p2p/protocols"
)

var SigDoesNotMatch = errors.New("signature does not match")
var DontOwe = errors.New("no negative balance")

// SwapPeer is a devp2p peer for the Swap protocol
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

// handleMsg is for handling messages when receiving messages
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

// handleChequeRequestMsg runs when a peer receives a `ChequeRequestMsg`
// It is thus run by the debitor
// So the debitor needs to:
//   * check that it indeed owes to the requestor (if not, ignore message)
//   * check serial number
//   * check amount
//   * if all is ok, issue the cheque
func (sp *SwapPeer) handleChequeRequestMsg(ctx context.Context, msg interface{}) (err error) {
	// check we have indeed a negative balance with the peer
	var req *ChequeRequestMsg
	var ok bool
	if req, ok = msg.(*ChequeRequestMsg); !ok {
		return fmt.Errorf("Unexpected message type: %v", err)
	}

	peer := req.Peer

	sp.swap.lock.Lock()
	defer sp.swap.lock.Unlock() // TODO: do we really want to block so long?

	peerBalance, err := sp.swap.GetPeerBalance(peer)
	if err != nil {
		return err
	}
	// do we actually owe to the remote peer?
	if peerBalance >= 0 {
		// TODO: should we send a message to the peer?
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
	cheque.Contract = sp.swap.owner.address
	pk, err := crypto.UnmarshalPubkey(req.PubKey)
	if err != nil {
		return err
	}
	cheque.Beneficiary = crypto.PubkeyToAddress(*pk)
	cheque.Sig, err = sp.signContent(cheque)
	if err != nil {
		return err
	}

	sp.swap.cheques[peer] = cheque
	err = sp.swap.stateStore.Put(peer.String()+"_cheques", &cheque)

	// TODO: error handling might be quite more complex
	if err != nil {
		return err
	}

	// TODO: reset balance here?
	// if we don't, then multiple ChequeRequestMsg may be sent and multiple
	// cheques will be generated
	// If we do, then if something goes wrong and the remote does not reset the balance,
	// we have issues as well.
	return sp.Send(ctx, msg)
}

// signContent signs the cheque
func (sp *SwapPeer) signContent(cheque *Cheque) ([]byte, error) {
	serialBytes := make([]byte, 32)
	amountBytes := make([]byte, 32)
	timeoutBytes := make([]byte, 32)
	input := append(cheque.Contract.Bytes(), cheque.Beneficiary.Bytes()...)
	binary.LittleEndian.PutUint64(serialBytes, cheque.Serial)
	binary.LittleEndian.PutUint64(amountBytes, cheque.Amount)
	binary.LittleEndian.PutUint64(timeoutBytes, cheque.Timeout)
	input = append(input, serialBytes[:]...)
	input = append(input, amountBytes[:]...)
	input = append(input, timeoutBytes[:]...)
	return crypto.Sign(crypto.Keccak256(input), sp.swap.owner.privateKey)
}

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a creditor
func (sp *SwapPeer) handleEmitChequeMsg(ctx context.Context, msg interface{}) error {
	chequeMsg, ok := msg.(*EmitChequeMsg)
	if !ok {
		return fmt.Errorf("Invalid message type, %v", msg)
	}
	cheque := chequeMsg.Cheque
	// reset balance to zero
	sp.swap.resetBalance(sp.Peer)
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
func (sp *SwapPeer) handleErrorMsg(ctx context.Context, msg interface{}) error {
	// maybe balance disagreement
	return nil
}

func (sp *SwapPeer) handleConfirmMsg(ctx context.Context, msg interface{}) error {
	// TODO; correct here?
	sp.swap.resetBalance(sp.Peer)
	return nil
}
