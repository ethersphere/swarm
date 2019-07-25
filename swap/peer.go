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

// ErrDontOwe indictates that no balance is actially owned
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	swap               *Swap
	backend            cswap.Backend
	beneficiary        common.Address
	contractAddress    common.Address
	lastReceivedCheque *Cheque
}

// NewPeer creates a new swap Peer instance
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
		return sp.handleEmitChequeMsg(ctx, msg)

	case *ErrorMsg:
		return sp.handleErrorMsg(ctx, msg)

	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
}

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a debitor
// TODO: validate the contract address in the cheque to match the address given at handshake
// TODO: this should not be blocking
func (sp *Peer) handleEmitChequeMsg(ctx context.Context, msg *EmitChequeMsg) error {
	log.Info("received emit cheque message")

	cheque := msg.Cheque
	if err := sp.processAndVerifyCheque(cheque); err != nil {
		log.Error("error invalid cheque", "from", sp.ID().String(), "err", err.Error())
		return err
	}

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	sp.swap.resetBalance(sp.ID(), 0-int64(cheque.Honey))

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
			log.Error("Got error when calling submitChequeBeneficiary", "err", err)
			//TODO: do something with the error
		}
		log.Info("tx minded", "receipt", receipt)
		//TODO: cashCheque
		//TODO: after the cashCheque is done, we have to watch the blockchain for x amount (25) blocks for reorgs
		//TODO: make sure we make a case where we listen to the possibiliyt of the peer shutting down.
	}()
	return err
}

// TODO: Error handling
// handleErrorMsg is called when an ErrorMsg is received
func (sp *Peer) handleErrorMsg(ctx context.Context, msg interface{}) error {
	log.Info("received error msg")
	// maybe balance disagreement
	return nil
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
func (sp *Peer) processAndVerifyCheque(cheque *Cheque) error {
	if err := sp.verifyChequeProperties(cheque); err != nil {
		return err
	}

	lastCheque := sp.loadLastReceivedCheque()

	// TODO: there should probably be a lock here?
	expectedAmount, err := sp.swap.oracle.GetPrice(cheque.Honey)
	if err != nil {
		return err
	}

	if err := sp.verifyChequeAgainstLast(cheque, lastCheque, expectedAmount); err != nil {
		return err
	}

	if err := sp.saveLastReceivedCheque(cheque); err != nil {
		log.Error("error while saving last received cheque", "peer", sp.ID().String(), "err", err.Error())
		// TODO: what do we do here?
	}

	return nil
}

// verifyChequeProperties verifies the signautre and if the cheque fields are appropiate for this peer
// it does not verify anything that requires knowing the previous cheque
func (sp *Peer) verifyChequeProperties(cheque *Cheque) error {
	if cheque.Contract != sp.contractAddress {
		return fmt.Errorf("wrong cheque parameters: expected contract: %x, was: %x", sp.contractAddress, cheque.Contract)
	}

	// the beneficiary is the owner of the counterparty swap contract
	if err := sp.swap.verifyChequeSig(cheque, sp.beneficiary); err != nil {
		return err
	}

	if cheque.Beneficiary != sp.swap.owner.address {
		return fmt.Errorf("wrong cheque parameters: expected beneficiary: %x, was: %x", sp.swap.owner.address, cheque.Beneficiary)
	}

	if cheque.Timeout != 0 {
		return fmt.Errorf("wrong cheque parameters: expected timeout to be 0, was: %d", cheque.Timeout)
	}

	return nil
}

// verifyChequeAgainstLast verifies that serial and amount are higher than in the previous cheque
// furthermore it cheques that the increase in amount is as expected
func (sp *Peer) verifyChequeAgainstLast(cheque *Cheque, lastCheque *Cheque, expectedAmount uint64) error {
	actualAmount := cheque.Amount

	if lastCheque != nil {
		if cheque.Serial <= lastCheque.Serial {
			return fmt.Errorf("wrong cheque parameters: expected serial larger than %d, was: %d", lastCheque.Serial, cheque.Serial)
		}

		if cheque.Amount <= lastCheque.Amount {
			return fmt.Errorf("wrong cheque parameters: expected amount larger than %d, was: %d", lastCheque.Amount, cheque.Amount)
		}

		actualAmount -= lastCheque.Amount
	}

	// TODO: maybe allow some range around expectedAmount?
	if expectedAmount != actualAmount {
		return fmt.Errorf("unexpected amount for honey, expected %d was %d", expectedAmount, actualAmount)
	}

	return nil
}

// loadLastReceivedCheque gets the last received cheque for this peer
// cheque gets loaded from database if not already in memory
func (sp *Peer) loadLastReceivedCheque() *Cheque {
	if sp.lastReceivedCheque == nil {
		sp.lastReceivedCheque = sp.swap.loadLastReceivedCheque(sp.ID())
	}
	return sp.lastReceivedCheque
}

// saveLastReceivedCheque saves cheque as the last received cheque for this peer
func (sp *Peer) saveLastReceivedCheque(cheque *Cheque) error {
	sp.lastReceivedCheque = cheque
	return sp.swap.saveLastReceivedCheque(sp.ID(), cheque)
}
