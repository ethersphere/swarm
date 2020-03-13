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
package chain

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TxSchedulerBackend is an extension of the normal Backend interface
type TxSchedulerBackend interface {
	Backend
	// SendTransactionWithID is the same as SendTransaction but with the ID of the associated request passed alongside
	// This is primarily used so the backend can react with the expected result during testing
	SendTransactionWithID(ctx context.Context, id uint64, tx *types.Transaction) error
}

// DefaultTxSchedulerBackend is the standard backend that should be used
// It simply wraps another Backend
type DefaultTxSchedulerBackend struct {
	Backend
}

// SendTransactionWithID in the default Backend calls the underlying SendTransaction function
func (b *DefaultTxSchedulerBackend) SendTransactionWithID(ctx context.Context, id uint64, tx *types.Transaction) error {
	return b.Backend.SendTransaction(ctx, tx)
}

// TxRequest describes a request for a transaction that can be scheduled
type TxRequest struct {
	To       common.Address // recipient of the transaction
	Data     []byte         // transaction data
	GasPrice *big.Int       // gas price or nil if suggested gas price should be used
	GasLimit uint64         // gas limit or 0 if it should be estimated
	Value    *big.Int       // amount of wei to send
}

// ToSignedTx returns a signed types.Transaction for the given request and nonce
func (request *TxRequest) ToSignedTx(nonce uint64, opts *bind.TransactOpts) (*types.Transaction, error) {
	tx := types.NewTransaction(
		nonce,
		request.To,
		request.Value,
		request.GasLimit,
		request.GasPrice,
		request.Data,
	)

	return opts.Signer(&types.HomesteadSigner{}, opts.From, tx)
}

// EstimateGas estimates the gas usage if this request was send from the supplied sender
func (request *TxRequest) EstimateGas(ctx context.Context, backend Backend, from common.Address) (uint64, error) {
	gasLimit, err := backend.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &request.To,
		Data: request.Data,
	})
	if err != nil {
		return 0, err
	}
	return gasLimit, nil
}

// TxScheduler represents a central sender for all transactions from a single ethereum account
// its purpose is to ensure there are no nonce issues and that transaction initiators are notified of the result
// notifications are guaranteed to happen even across node restarts and disconnects from the ethereum backend
// the account managed by this scheduler must not be used from anywhere else
type TxScheduler interface {
	// SetHandlers registers the handlers for the given handlerID
	// This starts the delivery of notifications for this handlerID
	SetHandlers(handlerID string, handlers *TxRequestHandlers) error
	// ScheduleRequest adds a new request to be processed
	// The request is assigned an id which is returned
	ScheduleRequest(handlerID string, request TxRequest, requestExtraData interface{}) (id uint64, err error)
	// GetExtraData loads the serialized extra data for this request from disk and tries to decode it
	GetExtraData(id uint64, request interface{}) error
	// GetRequestState gets the state the request is currently in
	GetRequestState(id uint64) (TxRequestState, error)
	// Start starts processing transactions if it is not already doing so
	// This cannot be used to restart the queue once stopped
	Start()
	// Stop stops processing transactions if it is running
	// It will block until processing has terminated
	Stop()
}

// TxRequestHandlers holds all the callbacks for a given string
// Any of the functions may be nil
// Notify functions are called by the transaction queue when a notification for a transaction occurs
// If the handler returns an error the notification will be resent in the future (including across restarts)
type TxRequestHandlers struct {
	// NotifyReceipt is called the first time a receipt is observed for a transaction
	// This happens the first time a transaction was included in a block
	NotifyReceipt func(ctx context.Context, id uint64, notification *TxReceiptNotification) error
	// NotifyPending is called after the transaction was successfully sent to the backend
	NotifyPending func(ctx context.Context, id uint64, notification *TxPendingNotification) error
	// NotifyCancelled is called when it is certain that this transaction will never be sent
	NotifyCancelled func(ctx context.Context, id uint64, notification *TxCancelledNotification) error
	// NotifyStatusUnknown is called if it cannot be determined if the transaction might be confirmed
	NotifyStatusUnknown func(ctx context.Context, id uint64, notification *TxStatusUnknownNotification) error
}

// TxReceiptNotification is the notification emitted when the receipt is available
type TxReceiptNotification struct {
	Receipt types.Receipt // the receipt of the included transaction
}

// TxCancelledNotification is the notification emitted when it is certain that a transaction will never be sent
type TxCancelledNotification struct {
	Reason string // The reason behind the cancellation
}

// TxStatusUnknownNotification is the notification emitted if it cannot be determined if the transaction might be confirmed
type TxStatusUnknownNotification struct {
	Reason string // The reason why it is unknown
}

// TxPendingNotification is the notification emitted after the transaction was successfully sent to the backend
type TxPendingNotification struct {
	Transaction *types.Transaction // The transaction that was sent
}

// TxRequestState is the type used to indicate which state the transaction is in
type TxRequestState uint8

const (
	// TxRequestStateScheduled is the initial state for all requests that enter the queue
	TxRequestStateScheduled TxRequestState = 0
	// TxRequestStateSigned means the transaction has been generated and signed but not yet sent
	TxRequestStateSigned TxRequestState = 1
	// TxRequestStatePending means the transaction has been sent but is not yet confirmed
	TxRequestStatePending TxRequestState = 2
	// TxRequestStateConfirmed is entered the first time a confirmation is received
	TxRequestStateConfirmed TxRequestState = 3
	// TxRequestStateStatusUnknown is used for all cases where it is unclear wether the transaction was broadcast or not. This is also used for timed-out transactions.
	TxRequestStateStatusUnknown TxRequestState = 4
	// TxRequestStateCancelled is used for all cases where it is certain the transaction was and never will be sent
	TxRequestStateCancelled TxRequestState = 5
)
