package chain

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TxScheduler represents a central sender for all transactions from a single ethereum account
// its purpose is to ensure there are no nonce issues and that transaction initiators are notified of the result
// notifications are guaranteed to happen even across node restarts and disconnects from the ethereum backend
type TxScheduler interface {
	// SetHandlers registers the handlers for the given requestTypeID
	// This starts the delivery of notifications for this requestTypeID
	SetHandlers(requestTypeID TxRequestTypeID, handlers *TxRequestHandlers) error
	// ScheduleRequest adds a new request to be processed
	// The request is assigned an id which is returned
	ScheduleRequest(requestTypeID TxRequestTypeID, request interface{}) (id uint64, err error)
	// GetRequest load the serialized transaction request from disk and tries to decode it
	GetRequest(id uint64, request interface{}) error
	// Start starts processing transactions if it is not already doing so
	Start()
	// Stop stops processing transactions if it is running
	// It will block until processing has terminated
	Stop()
}

// TxRequestTypeID is a combination of a handler and a request type
// All requests with a given TxRequestTypeID are handled the same
type TxRequestTypeID struct {
	Handler     string
	RequestType string
}

func (rti TxRequestTypeID) String() string {
	return fmt.Sprintf("%s.%s", rti.Handler, rti.RequestType)
}

// TxRequestHandlers holds all the callbacks for a given TxRequestTypeID
// Except for Send, any of the functions may be nil
// Notify functions are called by the transaction queue when a notification for a transaction occurs
// If the handler returns an error the notification will be resent in the future (including across restarts)
type TxRequestHandlers struct {
	// Send should send the transaction using the backend and opts provided
	// opts may be modified, however From, Nonce and Signer must be left untouched
	// If the transaction is sent through other means From, Nonce and Signer must be respected (if Nonce set to nil, the "pending" nonce must be used)
	Send func(id uint64, backend Backend, opts *bind.TransactOpts) (common.Hash, error)
	// NotifyReceipt is called the first time a receipt is observed for a transaction
	NotifyReceipt func(ctx context.Context, id uint64, notification *TxReceiptNotification) error
	// NotifyStateChanged is called every time the transaction status changes
	NotifyStateChanged func(ctx context.Context, id uint64, notification *TxStateChangedNotification) error
}

// TxRequestState is the type used to indicate which state the transaction is in
type TxRequestState uint8

// TxReceiptNotification is the notification emitted when the receipt is available
type TxReceiptNotification struct {
	Receipt types.Receipt // the receipt of the included transaction
}

// TxStateChangedNotification is the notification emitted when the state of the request changes
// Note that by the time the handler processes the notification, the state might have already changed again
type TxStateChangedNotification struct {
	OldState TxRequestState // the state prior to the change
	NewState TxRequestState // the state after the change
}

const (
	// TxRequestStateQueued is the initial state for all requests that enter the queue
	TxRequestStateQueued TxRequestState = 0
	// TxRequestStatePending means the request is no longer in the queue but not yet confirmed
	TxRequestStatePending TxRequestState = 1
	// TxRequestStateConfirmed is entered the first time a confirmation is received. This is a terminal state.
	TxRequestStateConfirmed TxRequestState = 2
	// TxRequestStateStatusUnknown is used for all cases where it is unclear wether the transaction was broadcast or not. This is also used for timed-out transactions.
	TxRequestStateStatusUnknown TxRequestState = 3
	// TxRequestStateCancelled is used for all cases where it is certain the transaction was and never will be sent
	TxRequestStateCancelled TxRequestState = 4
)
