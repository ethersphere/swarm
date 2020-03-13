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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/metrics"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/swap/chain"
	"github.com/ethersphere/swarm/uint256"
)

// CashChequeBeneficiaryTransactionCost is the expected gas cost of a CashChequeBeneficiary transaction
const CashChequeBeneficiaryTransactionCost = 50000

// CashoutRequestHandlerID is the handlerID used by the CashoutProcessor for CashoutRequests
const CashoutRequestHandlerID = "CashoutProcessor_CashoutRequest"

// CashoutRequest represents a request for a cashout operation
type CashoutRequest struct {
	Cheque      Cheque         // cheque to be cashed
	Destination common.Address // destination for the payout
}

// CashoutProcessor holds all relevant fields needed for processing cashouts
type CashoutProcessor struct {
	backend              chain.Backend     // ethereum backend to use
	txScheduler          chain.TxScheduler // transaction queue to use
	cashoutResultHandler CashoutResultHandler
	logger               Logger
}

// CashoutResultHandler is an interface which accepts CashChequeResults from a CashoutProcessor
type CashoutResultHandler interface {
	// Called by the CashoutProcessor when a CashoutRequest was successfully executed
	// It will be called again if an error is returned
	HandleCashoutResult(request *CashoutRequest, result *contract.CashChequeResult, receipt *types.Receipt) error
}

// newCashoutProcessor creates a new instance of CashoutProcessor
func newCashoutProcessor(txScheduler chain.TxScheduler, backend chain.Backend, privateKey *ecdsa.PrivateKey, cashoutResultHandler CashoutResultHandler, logger Logger) *CashoutProcessor {
	c := &CashoutProcessor{
		backend:              backend,
		txScheduler:          txScheduler,
		cashoutResultHandler: cashoutResultHandler,
		logger:               logger,
	}

	txScheduler.SetHandlers(CashoutRequestHandlerID, &chain.TxRequestHandlers{
		NotifyReceipt: func(ctx context.Context, id uint64, notification *chain.TxReceiptNotification) error {
			var request *CashoutRequest
			err := c.txScheduler.GetExtraData(id, &request)
			if err != nil {
				return err
			}

			otherSwap, err := contract.InstanceAt(request.Cheque.Contract, c.backend)
			if err != nil {
				return err
			}

			receipt := &notification.Receipt
			if receipt.Status == 0 {
				c.logger.Error(CashChequeAction, "cheque cashing transaction reverted", "tx", receipt.TxHash)
				return nil
			}

			result := otherSwap.CashChequeBeneficiaryResult(receipt)
			return c.cashoutResultHandler.HandleCashoutResult(request, result, receipt)
		},
		NotifyPending: func(ctx context.Context, id uint64, notification *chain.TxPendingNotification) error {
			c.logger.Debug(CashChequeAction, "cheque cashing transaction sent", "hash", notification.Transaction.Hash())
			return nil
		},
		NotifyCancelled: func(ctx context.Context, id uint64, notification *chain.TxCancelledNotification) error {
			c.logger.Warn(CashChequeAction, "cheque cashing transaction cancelled", "reason", notification.Reason)
			return nil
		},
		NotifyStatusUnknown: func(ctx context.Context, id uint64, notification *chain.TxStatusUnknownNotification) error {
			c.logger.Error(CashChequeAction, "cheque cashing transaction status unknown", "reason", notification.Reason)
			return nil
		},
	})
	return c
}

// submitCheque submits a cheque for cashout
// the cheque might not be cashed if it is not deemed profitable
func (c *CashoutProcessor) submitCheque(ctx context.Context, request *CashoutRequest) {
	expectedPayout, transactionCosts, err := c.estimatePayout(ctx, &request.Cheque)
	if err != nil {
		c.logger.Error(CashChequeAction, "could not estimate payout", "error", err)
		return
	}

	costsMultiplier := uint256.FromUint64(2)
	costThreshold, err := uint256.New().Mul(transactionCosts, costsMultiplier)
	if err != nil {
		c.logger.Error(CashChequeAction, "overflow in transaction fee", "error", err)
		return
	}

	// do a payout transaction if we get more than 2 times the gas costs
	if expectedPayout.Cmp(costThreshold) == 1 {
		c.logger.Info(CashChequeAction, "queueing cashout", "cheque", &request.Cheque)

		cheque := request.Cheque
		otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
		if err != nil {
			c.logger.Error(CashChequeAction, "could not get swap instance", "error", err)
			return
		}

		txRequest, err := otherSwap.CashChequeBeneficiaryRequest(cheque.Beneficiary, cheque.CumulativePayout, cheque.Signature)
		if err != nil {
			metrics.GetOrRegisterCounter("swap.cheques.cashed.errors", nil).Inc(1)
			c.logger.Error(CashChequeAction, "cashing cheque:", "error", err)
			return
		}

		_, err = c.txScheduler.ScheduleRequest(CashoutRequestHandlerID, *txRequest, request)
		if err != nil {
			metrics.GetOrRegisterCounter("swap.cheques.cashed.errors", nil).Inc(1)
			c.logger.Error(CashChequeAction, "cashing cheque:", "error", err)
		}
	}
}

// estimatePayout estimates the payout for a given cheque as well as the transaction cost
func (c *CashoutProcessor) estimatePayout(ctx context.Context, cheque *Cheque) (expectedPayout *uint256.Uint256, transactionCosts *uint256.Uint256, err error) {
	otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
	if err != nil {
		return nil, nil, err
	}

	po, err := otherSwap.PaidOut(&bind.CallOpts{Context: ctx}, cheque.Beneficiary)
	if err != nil {
		return nil, nil, err
	}

	paidOut, err := uint256.New().Set(*po)
	if err != nil {
		return nil, nil, err
	}

	gp, err := c.backend.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, err
	}

	gasPrice, err := uint256.New().Set(*gp)
	if err != nil {
		return nil, nil, err
	}

	transactionCosts, err = uint256.New().Mul(gasPrice, uint256.FromUint64(CashChequeBeneficiaryTransactionCost))
	if err != nil {
		return nil, nil, err
	}

	if paidOut.Cmp(cheque.CumulativePayout) > 0 {
		return uint256.New(), transactionCosts, nil
	}

	expectedPayout, err = uint256.New().Sub(cheque.CumulativePayout, paidOut)
	if err != nil {
		return nil, nil, err
	}

	return expectedPayout, transactionCosts, nil
}
