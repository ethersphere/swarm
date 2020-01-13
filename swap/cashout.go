// Copyright 2020 The Swarm Authors
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
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p/enode"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/state"
)

// CashChequeBeneficiaryTransactionCost is the expected gas cost of a CashChequeBeneficiary transaction
const CashChequeBeneficiaryTransactionCost = 50000

// CashoutProcessor holds all relevant fields needed for the cashout go routine
type CashoutProcessor struct {
	lock    sync.Mutex           // lock for the loop (used to prevent double-closing of channels)
	queue   chan *CashoutRequest // channel for future cashout requests
	context context.Context      // context used for the loop
	cancel  context.CancelFunc   // cancel function of context
	done    chan bool            // channel for signalling that the loop has terminated
	cashed  chan *CashoutRequest // channel for successful cashouts

	store      state.Store       // state store to save and load the active cashout from
	backend    contract.Backend  // ethereum backend to use
	privateKey *ecdsa.PrivateKey // private key to use
}

// CashoutRequest represents a request for a cashout operation
type CashoutRequest struct {
	Cheque      Cheque         // cheque to be cashed
	Peer        enode.ID       // peer this cheque is from
	Destination common.Address // destination for the payout
}

// ActiveCashout stores the necessary information for a cashout in progess
type ActiveCashout struct {
	Request         CashoutRequest // the request that caused this cashout
	TransactionHash common.Hash    // the hash of the current transaction for this request
}

// newCashoutProcessor creates a new instance of CashoutProcessor
func newCashoutProcessor(store state.Store, backend contract.Backend, privateKey *ecdsa.PrivateKey) *CashoutProcessor {
	context, cancel := context.WithCancel(context.Background())
	return &CashoutProcessor{
		store:      store,
		backend:    backend,
		privateKey: privateKey,
		done:       make(chan bool),
		queue:      make(chan *CashoutRequest, 50),
		cashed:     make(chan *CashoutRequest, 50),
		context:    context,
		cancel:     cancel,
	}
}

// queueRequest attempts to add a request to the queue
func (c *CashoutProcessor) queueRequest(cashoutRequest *CashoutRequest) {
	select {
	case c.queue <- cashoutRequest:
	default:
		swapLog.Warn("attempting to write cashout request to closed queue")
	}
}

// start starts the loop go routine
func (c *CashoutProcessor) start() {
	go c.loop()
}

// stop cancels the loop by cancelling c.context and closing all the channels
func (c *CashoutProcessor) stop() {
	c.lock.Lock()
	select {
	case <-c.context.Done():
		return
	default:
	}
	c.cancel()
	c.lock.Unlock()

	close(c.queue)
	<-c.done
	close(c.done)
}

func (c *CashoutProcessor) loop() {
	if err := c.continueActiveCashout(); err != nil {
		swapLog.Error(err.Error()) // TODO: stop processing for now
		return
	}

	for request := range c.queue {
		ctx, cancel := context.WithTimeout(c.context, DefaultTransactionTimeout)
		defer cancel()

		estimatedPayout, transactionCosts, err := c.estimatePayout(ctx, &request.Cheque)
		if err != nil {
			swapLog.Error(err.Error())
			continue
		}

		// do a payout transaction if we get 2 times the gas costs
		if estimatedPayout > 2*transactionCosts {
			if err = c.cashCheque(ctx, request); err != nil {
				// if sending the transaction fails put the request back into the queue
				swapLog.Error("failed to cash cheque", "err", err)
				c.queueRequest(request)
				continue
			}

			swapLog.Info("finished cashing cheque")
		}
	}

	c.done <- true
}

// cashCheque tries to cash the cheque specified in the request
// after the transaction is sent it waits on its success
// if the transaction fails it is reentered into the queue
func (c *CashoutProcessor) cashCheque(ctx context.Context, request *CashoutRequest) error {
	cheque := request.Cheque
	opts := bind.NewKeyedTransactor(c.privateKey)
	opts.Context = ctx

	otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
	if err != nil {
		return err
	}

	tx, err := otherSwap.CashChequeBeneficiaryStart(opts, request.Destination, big.NewInt(int64(cheque.CumulativePayout)), cheque.Signature)
	if err != nil {
		return err
	}

	activeCashout := &ActiveCashout{
		Request:         *request,
		TransactionHash: tx.Hash(),
	}

	// before waiting save the request and transaction information to disk
	err = c.saveActiveCashout(activeCashout)
	if err != nil {
		return err
	}

	// this blocks until the cashout has been successfully processed
	err = c.processActiveCashout(activeCashout)
	if err != nil {
		return err
	}

	// delete the request and transaction information to disk
	return c.saveActiveCashout(nil)
}

// estimatePayout estimates the payout for a given cheque as well as the transaction cost
func (c *CashoutProcessor) estimatePayout(ctx context.Context, cheque *Cheque) (expectedPayout uint64, transactionCosts uint64, err error) {
	otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
	if err != nil {
		return 0, 0, err
	}

	paidOut, err := otherSwap.PaidOut(&bind.CallOpts{Context: ctx}, cheque.Beneficiary)
	if err != nil {
		return 0, 0, err
	}

	gasPrice, err := c.backend.SuggestGasPrice(ctx)
	if err != nil {
		return 0, 0, err
	}

	transactionCosts = gasPrice.Uint64() * CashChequeBeneficiaryTransactionCost

	if paidOut.Cmp(big.NewInt(int64(cheque.CumulativePayout))) > 0 {
		return 0, transactionCosts, nil
	}

	expectedPayout = cheque.CumulativePayout - paidOut.Uint64()

	return expectedPayout, transactionCosts, nil
}

// continueActiveCashout checks if an active cashout was saved and if so processes that one
// if the cashout fails it reenters the queue
func (c *CashoutProcessor) continueActiveCashout() error {
	activeCashout, err := c.loadActiveCashout()
	if err != nil {
		return err
	}

	if activeCashout != nil {
		if err = c.processActiveCashout(activeCashout); err != nil {
			c.queueRequest(&activeCashout.Request)
			return err
		}
		if err = c.saveActiveCashout(nil); err != nil {
			return err
		}
	}

	return nil
}

// processActiveCashout waits for activeCashout to complete
func (c *CashoutProcessor) processActiveCashout(activeCashout *ActiveCashout) error {
	ctx, cancel := context.WithTimeout(c.context, DefaultTransactionTimeout)
	defer cancel()

	receipt, err := contract.WaitForTransactionByHash(ctx, c.backend, activeCashout.TransactionHash)
	if err != nil {
		return err
	}

	otherSwap, err := contract.InstanceAt(activeCashout.Request.Cheque.Contract, c.backend)
	if err != nil {
		return err
	}

	result := otherSwap.CashChequeBeneficiaryResult(receipt)

	metrics.GetOrRegisterCounter("swap.cheques.cashed.honey", nil).Inc(result.TotalPayout.Int64())

	if result.Bounced {
		metrics.GetOrRegisterCounter("swap.cheques.cashed.bounced", nil).Inc(1)
		swapLog.Warn("cheque bounced", "tx", receipt.TxHash)
	}

	swapLog.Info("cheque cashed", "honey", activeCashout.Request.Cheque.Honey)

	select {
	case c.cashed <- &activeCashout.Request:
	default:
		log.Error("cashed channel full")
	}

	return nil
}

// loadActiveCashout loads the activeCashout from the store
func (c *CashoutProcessor) loadActiveCashout() (activeCashout *ActiveCashout, err error) {
	err = c.store.Get("cashout_loop_active", &activeCashout)
	if err == state.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return activeCashout, nil
}

// saveActiveCashout saves activeCashout to the store
func (c *CashoutProcessor) saveActiveCashout(activeCashout *ActiveCashout) error {
	return c.store.Put("cashout_loop_active", activeCashout)
}
