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
	"github.com/ethereum/go-ethereum/metrics"
	contract "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/swap/chain"
	"github.com/ethersphere/swarm/swap/int256"
)

// CashChequeBeneficiaryTransactionCost is the expected gas cost of a CashChequeBeneficiary transaction
const CashChequeBeneficiaryTransactionCost = 50000

// CashoutProcessor holds all relevant fields needed for processing cashouts
type CashoutProcessor struct {
	backend    chain.Backend     // ethereum backend to use
	privateKey *ecdsa.PrivateKey // private key to use
	Logger     Logger
}

// CashoutRequest represents a request for a cashout operation
type CashoutRequest struct {
	Cheque      Cheque         // cheque to be cashed
	Destination common.Address // destination for the payout
	Logger      Logger
}

// ActiveCashout stores the necessary information for a cashout in progess
type ActiveCashout struct {
	Request         CashoutRequest // the request that caused this cashout
	TransactionHash common.Hash    // the hash of the current transaction for this request
	Logger          Logger
}

// newCashoutProcessor creates a new instance of CashoutProcessor
func newCashoutProcessor(backend chain.Backend, privateKey *ecdsa.PrivateKey) *CashoutProcessor {
	return &CashoutProcessor{
		backend:    backend,
		privateKey: privateKey,
	}
}

// cashCheque tries to cash the cheque specified in the request
// after the transaction is sent it waits on its success
func (c *CashoutProcessor) cashCheque(ctx context.Context, request *CashoutRequest) error {
	cheque := request.Cheque
	opts := bind.NewKeyedTransactor(c.privateKey)
	opts.Context = ctx

	otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
	if err != nil {
		return err
	}

	tx, err := otherSwap.CashChequeBeneficiaryStart(opts, request.Destination, cheque.CumulativePayout, cheque.Signature)
	if err != nil {
		return err
	}

	// this blocks until the cashout has been successfully processed
	return c.waitForAndProcessActiveCashout(&ActiveCashout{
		Request:         *request,
		TransactionHash: tx.Hash(),
		Logger:          request.Logger,
	})
}

// estimatePayout estimates the payout for a given cheque as well as the transaction cost
func (c *CashoutProcessor) estimatePayout(ctx context.Context, cheque *Cheque) (expectedPayout *int256.Uint256, transactionCosts *int256.Uint256, err error) {
	otherSwap, err := contract.InstanceAt(cheque.Contract, c.backend)
	if err != nil {
		return nil, nil, err
	}

	po, err := otherSwap.PaidOut(&bind.CallOpts{Context: ctx}, cheque.Beneficiary)
	if err != nil {
		return nil, nil, err
	}

	paidOut, err := int256.NewUint256(po)
	if err != nil {
		return nil, nil, err
	}

	gp, err := c.backend.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, err
	}

	gasPrice, err := int256.NewUint256(gp)
	if err != nil {
		return nil, nil, err
	}

	transactionCosts, err = new(int256.Uint256).Mul(gasPrice, int256.Uint256From(CashChequeBeneficiaryTransactionCost))
	if err != nil {
		return nil, nil, err
	}

	if paidOut.Cmp(cheque.CumulativePayout) > 0 {
		return new(int256.Uint256), transactionCosts, nil
	}

	expectedPayout, err = new(int256.Uint256).Sub(cheque.CumulativePayout, paidOut)
	if err != nil {
		return nil, nil, err
	}

	return expectedPayout, transactionCosts, nil
}

// waitForAndProcessActiveCashout waits for activeCashout to complete
func (c *CashoutProcessor) waitForAndProcessActiveCashout(activeCashout *ActiveCashout) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTransactionTimeout)
	defer cancel()

	receipt, err := chain.WaitMined(ctx, c.backend, activeCashout.TransactionHash)
	if err != nil {
		return err
	}

	otherSwap, err := contract.InstanceAt(activeCashout.Request.Cheque.Contract, c.backend)
	if err != nil {
		return err
	}

	result := otherSwap.CashChequeBeneficiaryResult(receipt)

	metrics.GetOrRegisterCounter("swap/cheques/cashed/honey", nil).Inc(result.TotalPayout.Int64())

	if result.Bounced {
		metrics.GetOrRegisterCounter("swap/cheques/cashed/bounced", nil).Inc(1)
		activeCashout.Logger.Warn(CashChequeAction, "cheque bounced", "tx", receipt.TxHash)
	}

	activeCashout.Logger.Info(CashChequeAction, "cheque cashed", "honey", activeCashout.Request.Cheque.Honey)
	return nil
}
