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
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/p2p/protocols"
)

// ErrDontOwe indictates that no balance is actially owned
var ErrDontOwe = errors.New("no negative balance")

// Peer is a devp2p peer for the Swap protocol
type Peer struct {
	*protocols.Peer
	swap               *Swap
	backend            contract.Backend
	beneficiary        common.Address
	contractAddress    common.Address
	lastReceivedCheque *Cheque
}

// NewPeer creates a new swap Peer instance
func NewPeer(p *protocols.Peer, s *Swap, backend contract.Backend, beneficiary common.Address, contractAddress common.Address) *Peer {
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
		go sp.handleEmitChequeMsg(ctx, msg)
	}
	return nil
}

// handleEmitChequeMsg should be handled by the creditor when it receives
// a cheque from a debitor
// TODO: validate the contract address in the cheque to match the address given at handshake
// TODO: this should not be blocking
func (sp *Peer) handleEmitChequeMsg(ctx context.Context, msg *EmitChequeMsg) error {
	cheque := msg.Cheque
	log.Debug("received emit cheque message from peer", "peer", sp.ID().String())
	actualAmount, err := sp.processAndVerifyCheque(cheque)
	if err != nil {
		log.Error("invalid cheque from peer", "peer", sp.ID().String(), "error", err.Error())
		return err
	}

	// reset balance by amount
	// as this is done by the creditor, receiving the cheque, the amount should be negative,
	// so that updateBalance will calculate balance + amount which result in reducing the peer's balance
	sp.swap.resetBalance(sp.ID(), 0-int64(cheque.Honey))

	// cash in cheque
	opts := bind.NewKeyedTransactor(sp.swap.owner.privateKey)
	opts.Context = ctx

	otherSwap, err := contract.InstanceAt(cheque.Contract, sp.backend)
	if err != nil {
		log.Error("could not get an instance of simpleSwap", "error", err)
		return err
	}

	// submit cheque to the blockchain and cashes it directly
	go func() {
		// blocks here, as we are waiting for the transaction to be mined
		receipt, err := otherSwap.SubmitChequeBeneficiary(opts, sp.backend, big.NewInt(int64(cheque.Serial)), big.NewInt(int64(cheque.Amount)), big.NewInt(int64(cheque.Timeout)), cheque.Signature)
		if err != nil {
			log.Error("error calling submitChequeBeneficiary", "error", err)
			//TODO: do something with the error
			return
		}
		log.Info("submit tx mined", "receipt", receipt)

		receipt, err = otherSwap.CashChequeBeneficiary(opts, sp.backend, sp.swap.owner.Contract, big.NewInt(int64(actualAmount)))
		if err != nil {
			log.Error("Got error when calling cashChequeBeneficiary", "err", err)
			//TODO: do something with the error
			return
		}
		log.Info("cash tx mined", "receipt", receipt)
		//TODO: after the cashCheque is done, we have to watch the blockchain for x amount (25) blocks for reorgs
		//TODO: make sure we make a case where we listen to the possibiliyt of the peer shutting down.
	}()
	return err
}

// processAndVerifyCheque verifies the cheque and compares it with the last received cheque
// if the cheque is valid it will also be saved as the new last cheque
func (sp *Peer) processAndVerifyCheque(cheque *Cheque) (uint64, error) {
	if err := sp.verifyChequeProperties(cheque); err != nil {
		return 0, err
	}

	lastCheque := sp.loadLastReceivedCheque()

	// TODO: there should probably be a lock here?
	expectedAmount, err := sp.swap.oracle.GetPrice(cheque.Honey)
	if err != nil {
		return 0, err
	}

	actualAmount, err := verifyChequeAgainstLast(cheque, lastCheque, expectedAmount)
	if err != nil {
		return 0, err
	}

	if err := sp.saveLastReceivedCheque(cheque); err != nil {
		log.Error("error while saving last received cheque", "peer", sp.ID().String(), "err", err.Error())
		// TODO: what do we do here?
	}

	return actualAmount, nil
}

// verifyChequeProperties verifies the signature and if the cheque fields are appropriate for this peer
// it does not verify anything that requires knowing the previous cheque
func (sp *Peer) verifyChequeProperties(cheque *Cheque) error {
	if cheque.Contract != sp.contractAddress {
		return fmt.Errorf("wrong cheque parameters: expected contract: %x, was: %x", sp.contractAddress, cheque.Contract)
	}

	// the beneficiary is the owner of the counterparty swap contract
	if err := cheque.VerifySig(sp.beneficiary); err != nil {
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
// returns the actual amount received in this cheque
func verifyChequeAgainstLast(cheque *Cheque, lastCheque *Cheque, expectedAmount uint64) (uint64, error) {
	actualAmount := cheque.Amount

	if lastCheque != nil {
		if cheque.Serial <= lastCheque.Serial {
			return 0, fmt.Errorf("wrong cheque parameters: expected serial larger than %d, was: %d", lastCheque.Serial, cheque.Serial)
		}

		if cheque.Amount <= lastCheque.Amount {
			return 0, fmt.Errorf("wrong cheque parameters: expected amount larger than %d, was: %d", lastCheque.Amount, cheque.Amount)
		}

		actualAmount -= lastCheque.Amount
	}

	// TODO: maybe allow some range around expectedAmount?
	if expectedAmount != actualAmount {
		return 0, fmt.Errorf("unexpected amount for honey, expected %d was %d", expectedAmount, actualAmount)
	}

	return actualAmount, nil
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
