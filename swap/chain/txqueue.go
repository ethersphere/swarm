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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethersphere/swarm/state"
)

// TxQueue is a TxScheduler which sends transactions in sequence
// A new transaction is only sent after the previous one confirmed
// This is done to minimize the chance of wrong nonce use
type TxQueue struct {
	lock        sync.Mutex         // lock for the entire queue
	ctx         context.Context    // context used for all network requests and waiting operations to ensure the queue can be stopped at any point
	cancel      context.CancelFunc // function to cancel the above context
	wg          sync.WaitGroup     // used to ensure that all background go routines have finished before Stop returns
	startedChan chan struct{}      // channel to be closed when the queue has started processing
	started     bool               // bool indicating that the queue has been started. used to ensure it does not run multiple times simultaneously
	errorChan   chan error         // channel to stop the queue in case of errors

	store              state.Store                   // state store to use as the db backend
	prefix             string                        // all keys in the state store are prefixed with this
	requestQueue       *persistentQueue              // queue for all future requests
	handlers           map[string]*TxRequestHandlers // map from handlerIDs to their registered handlers
	notificationQueues map[string]*persistentQueue   // map from handlerIDs to the notification queue of that handler

	backend    TxSchedulerBackend // ethereum backend to use
	privateKey *ecdsa.PrivateKey  // private key used to sign transactions
}

// txRequestData is the metadata the queue saves for every request
// the extra data is stored at a different key
type txRequestData struct {
	ID          uint64             // id of the request
	Request     TxRequest          // the request itself
	HandlerID   string             // the type id of this request
	State       TxRequestState     // the state this request is in
	Transaction *types.Transaction // the generated transaction for this request or nil if not yet signed
}

// notificationQueueItem is the metadata the queue saves for every pending notification
// the actual notification content is stored at a different key
type notificationQueueItem struct {
	NotificationType string // the type of the notification
	RequestID        uint64 // the request this notification is for
}

const (
	txReceiptNotificationType       = "TxReceiptNotification"
	txPendingNotificationType       = "TxPendingNotification"
	txCancelledNotificationType     = "TxCancelledNotification"
	txStatusUnknownNotificationType = "TxStatusUnknownNotification"
)

// NewTxQueue creates a new TxQueue
func NewTxQueue(store state.Store, prefix string, backend TxSchedulerBackend, privateKey *ecdsa.PrivateKey) *TxQueue {
	txq := &TxQueue{
		store:              store,
		prefix:             prefix,
		handlers:           make(map[string]*TxRequestHandlers),
		notificationQueues: make(map[string]*persistentQueue),
		backend:            backend,
		privateKey:         privateKey,
		requestQueue:       newPersistentQueue(store, prefix+"_requestQueue_"),
		errorChan:          make(chan error, 1),
		startedChan:        make(chan struct{}),
	}
	// we create the context here already because handlers can be set before the queue starts
	txq.ctx, txq.cancel = context.WithCancel(context.Background())
	return txq
}

// requestKey returns the database key for the txRequestData for the given id
func (txq *TxQueue) requestKey(id uint64) string {
	return fmt.Sprintf("%s_requests_%d", txq.prefix, id)
}

// extraDataKey returns the database key for the extra data stored alongside the request
func (txq *TxQueue) extraDataKey(id uint64) string {
	return fmt.Sprintf("%s_data", txq.requestKey(id))
}

// activeRequestKey returns the database key used for the currently active request
func (txq *TxQueue) activeRequestKey() string {
	return fmt.Sprintf("%s_active", txq.prefix)
}

// notificationKey returns the database key for a notification
func (txq *TxQueue) notificationKey(key string) string {
	return fmt.Sprintf("%s_notification_%s", txq.prefix, key)
}

// idKey returns the database key for the last used id value
func (txq *TxQueue) idKey() string {
	return fmt.Sprintf("%s_request_id", txq.prefix)
}

// stopWithError sends the error to the error channel
// this is used to stop the queue from notification handlers
func (txq *TxQueue) stopWithError(err error) {
	select {
	case txq.errorChan <- err:
	default:
		log.Error("failed to write error to txqueue error channel", "error", err)
	}
}

// ScheduleRequest adds a new request to be processed
// The request is assigned an id which is returned
func (txq *TxQueue) ScheduleRequest(handlerID string, request TxRequest, extraData interface{}) (id uint64, err error) {
	txq.lock.Lock()
	defer txq.lock.Unlock()

	// get the last id
	err = txq.store.Get(txq.idKey(), &id)
	if err != nil && err != state.ErrNotFound {
		return 0, err
	}
	// increment existing id, starting with an initial value of 1
	id++

	// in a single batch this
	// * stores the request data
	// * stores the request extraData
	// * adds it to the queue
	batch := new(state.StoreBatch)
	err = batch.Put(txq.idKey(), id)
	if err != nil {
		return 0, err
	}

	err = batch.Put(txq.extraDataKey(id), extraData)
	if err != nil {
		return 0, err
	}

	err = batch.Put(txq.requestKey(id), &txRequestData{
		ID:        id,
		Request:   request,
		HandlerID: handlerID,
		State:     TxRequestStateScheduled,
	})
	if err != nil {
		return 0, err
	}

	_, triggerQueue, err := txq.requestQueue.enqueue(batch, id)
	if err != nil {
		return 0, err
	}

	// persist to disk
	err = txq.store.WriteBatch(batch)
	if err != nil {
		return 0, err
	}

	triggerQueue()
	return id, nil
}

// GetExtraData load the serialized extra data for this request from disk and tries to decode it
func (txq *TxQueue) GetExtraData(id uint64, request interface{}) error {
	return txq.store.Get(txq.extraDataKey(id), &request)
}

// GetRequestState gets the state the request is currently in
func (txq *TxQueue) GetRequestState(id uint64) (TxRequestState, error) {
	var requestMetadata *txRequestData
	err := txq.store.Get(txq.requestKey(id), &requestMetadata)
	if err != nil {
		return 0, err
	}
	return requestMetadata.State, nil
}

// Start starts processing transactions if it is not already doing so
func (txq *TxQueue) Start() {
	txq.lock.Lock()
	defer txq.lock.Unlock()

	if txq.started {
		return
	}

	txq.started = true
	txq.wg.Add(2)
	go func() {
		defer txq.wg.Done()
		// run the actual loop
		err := txq.processQueue()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error("transaction queue terminated with an error", "queue", txq.prefix, "error", err)
		}
	}()

	go func() {
		defer txq.wg.Done()
		// listen on the error channel and stop the queue on error
		select {
		case err := <-txq.errorChan:
			log.Error("unrecoverable transaction queue error (transaction processing disabled)", "error", err)
			txq.Stop()
		case <-txq.ctx.Done():
		}
	}()

	close(txq.startedChan)
}

// Stop stops processing transactions if it is running
// It will block until processing has terminated
func (txq *TxQueue) Stop() {
	txq.lock.Lock()

	if !txq.started {
		txq.lock.Unlock()
		return
	}

	// we cancel the context that all long running operations in the queue use
	txq.cancel()
	txq.lock.Unlock()
	// wait until all routines have finished
	txq.wg.Wait()
}

// getNotificationQueue gets the notification queue for a handler
// it initializes the struct if it does not yet exist
// the TxQueue lock must be held
func (txq *TxQueue) getNotificationQueue(handlerID string) *persistentQueue {
	queue, ok := txq.notificationQueues[handlerID]
	if !ok {
		queue = newPersistentQueue(txq.store, fmt.Sprintf("%s_notify_%s", txq.prefix, handlerID))
		txq.notificationQueues[handlerID] = queue
	}
	return queue
}

// SetHandlers registers the handlers for the given handlerID
// This starts the delivery of notifications for this handlerID
func (txq *TxQueue) SetHandlers(handlerID string, handlers *TxRequestHandlers) error {
	txq.lock.Lock()
	defer txq.lock.Unlock()

	if txq.handlers[handlerID] != nil {
		return fmt.Errorf("handlers for %s already set", handlerID)
	}
	txq.handlers[handlerID] = handlers
	notifyQueue := txq.getNotificationQueue(handlerID)

	// go routine processing the notification queue for this handler
	txq.wg.Add(1)
	go func() {
		defer txq.wg.Done()

		// only start sending notification once the loop started
		select {
		case <-txq.startedChan:
		case <-txq.ctx.Done():
			return
		}

		for {
			var item notificationQueueItem
			// get the next notification item
			key, err := notifyQueue.next(txq.ctx, &item, &txq.lock)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					txq.stopWithError(fmt.Errorf("could not read from notification queue: %v", err))
				}
				return
			}
			// since this is the only function which deletes this item from notifyQueue we can already unlock here
			txq.lock.Unlock()

			// load and decode the notification
			var notification interface{}
			switch item.NotificationType {
			case txReceiptNotificationType:
				notification = &TxReceiptNotification{}
			case txPendingNotificationType:
				notification = &TxPendingNotification{}
			case txCancelledNotificationType:
				notification = &TxCancelledNotification{}
			case txStatusUnknownNotificationType:
				notification = &TxStatusUnknownNotification{}
			}

			err = txq.store.Get(txq.notificationKey(key), notification)
			if err != nil {
				txq.stopWithError(fmt.Errorf("could not read notification: %v", err))
				return
			}

			switch item.NotificationType {
			case txReceiptNotificationType:
				if handlers.NotifyReceipt != nil {
					err = handlers.NotifyReceipt(txq.ctx, item.RequestID, notification.(*TxReceiptNotification))
				}
			case txPendingNotificationType:
				if handlers.NotifyPending != nil {
					err = handlers.NotifyPending(txq.ctx, item.RequestID, notification.(*TxPendingNotification))
				}
			case txCancelledNotificationType:
				if handlers.NotifyCancelled != nil {
					err = handlers.NotifyCancelled(txq.ctx, item.RequestID, notification.(*TxCancelledNotification))
				}
			case txStatusUnknownNotificationType:
				if handlers.NotifyStatusUnknown != nil {
					err = handlers.NotifyStatusUnknown(txq.ctx, item.RequestID, notification.(*TxStatusUnknownNotification))
				}
			}

			// if a handler failed we will try again in 10 seconds
			if err != nil {
				log.Error("transaction request handler failed", "type", item.NotificationType, "request", item.RequestID, "error", err)
				select {
				case <-txq.ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}

			// once the notification was handled delete it from the queue
			txq.lock.Lock()
			batch := new(state.StoreBatch)
			notifyQueue.delete(batch, key)
			err = txq.store.WriteBatch(batch)
			txq.lock.Unlock()
			if err != nil {
				txq.stopWithError(fmt.Errorf("could not delete notification: %v", err))
				return
			}
		}
	}()
	return nil
}

// helper function to trigger a notification
// the returned trigger function must be called once the batch has been written
// must be called with the txqueue lock held
func (txq *TxQueue) notify(batch *state.StoreBatch, id uint64, handlerID string, notificationType string, notification interface{}) (triggerNotifyQueue func(), err error) {
	notifyQueue := txq.getNotificationQueue(handlerID)
	key, triggerNotifyQueue, err := notifyQueue.enqueue(batch, &notificationQueueItem{
		RequestID:        id,
		NotificationType: notificationType,
	})
	if err != nil {
		return nil, fmt.Errorf("could not serialize notification queue item: %v", err)
	}

	err = batch.Put(txq.notificationKey(key), notification)
	if err != nil {
		return nil, fmt.Errorf("could not serialize notification: %v", err)
	}
	return triggerNotifyQueue, nil
}

// waitForNextRequest waits for the next request and sets it as the active request
// the txqueue lock must not be held
func (txq *TxQueue) waitForNextRequest() (requestMetadata *txRequestData, err error) {
	var id uint64
	// get the id of the next request in the queue
	key, err := txq.requestQueue.next(txq.ctx, &id, &txq.lock)
	if err != nil {
		return nil, err
	}
	defer txq.lock.Unlock()

	err = txq.store.Get(txq.requestKey(id), &requestMetadata)
	if err != nil {
		return nil, err
	}

	// if the request was successfully decoded it is removed from the queue and set as the active request
	batch := new(state.StoreBatch)
	err = batch.Put(txq.activeRequestKey(), requestMetadata.ID)
	if err != nil {
		return nil, fmt.Errorf("could not put id write into batch: %v", err)
	}
	txq.requestQueue.delete(batch, key)

	err = txq.store.WriteBatch(batch)
	if err != nil {
		return nil, err
	}

	return requestMetadata, nil
}

// helper function to set a request state and remove it as the active request in a single batch
// the txqueue lock must be held
func (txq *TxQueue) finalizeRequest(batch *state.StoreBatch, requestMetadata *txRequestData, state TxRequestState) error {
	requestMetadata.State = state
	err := batch.Put(txq.requestKey(requestMetadata.ID), requestMetadata.ID)
	if err != nil {
		return err
	}
	batch.Delete(txq.activeRequestKey())
	return txq.store.WriteBatch(batch)
}

// helper function to set a request as cancelled and emit the appropriate notification
// the txqueue lock must be held
func (txq *TxQueue) finalizeRequestCancelled(requestMetadata *txRequestData, err error) error {
	batch := new(state.StoreBatch)
	trigger, err := txq.notify(batch, requestMetadata.ID, requestMetadata.HandlerID, "TxCancelledNotification", &TxCancelledNotification{
		Reason: err.Error(),
	})
	if err != nil {
		return err
	}

	err = txq.finalizeRequest(batch, requestMetadata, TxRequestStateCancelled)
	if err != nil {
		return err
	}
	trigger()
	return nil
}

// helper function to set a request as status unknown and emit the appropriate notification
// the txqueue lock must be held
func (txq *TxQueue) finalizeRequestStatusUnknown(requestMetadata *txRequestData, reason string) error {
	batch := new(state.StoreBatch)
	trigger, err := txq.notify(batch, requestMetadata.ID, requestMetadata.HandlerID, txStatusUnknownNotificationType, &TxStatusUnknownNotification{
		Reason: reason,
	})
	if err != nil {
		return err
	}

	err = txq.finalizeRequest(batch, requestMetadata, TxRequestStateStatusUnknown)
	if err != nil {
		return err
	}
	trigger()
	return nil
}

// helper function to set a request as confirmed and emit the appropriate notification
// the txqueue lock must be held
func (txq *TxQueue) finalizeRequestConfirmed(requestMetadata *txRequestData, receipt types.Receipt) error {
	batch := new(state.StoreBatch)
	trigger, err := txq.notify(batch, requestMetadata.ID, requestMetadata.HandlerID, txReceiptNotificationType, &TxReceiptNotification{
		Receipt: receipt,
	})
	if err != nil {
		return err
	}

	err = txq.finalizeRequest(batch, requestMetadata, TxRequestStateConfirmed)
	if err != nil {
		return err
	}
	trigger()
	return nil
}

// processRequest continues processing the provided request
func (txq *TxQueue) processRequest(requestMetadata *txRequestData) error {
	switch requestMetadata.State {
	case TxRequestStateScheduled:
		err := txq.generateTransaction(requestMetadata)
		if err != nil {
			return err
		}
		fallthrough
	case TxRequestStateSigned:
		err := txq.sendTransaction(requestMetadata)
		if err != nil {
			return err
		}
		fallthrough
	case TxRequestStatePending:
		return txq.waitForActiveTransaction(requestMetadata)
	default:
		return fmt.Errorf("trying to process transaction in unknown state: %d", requestMetadata.State)
	}
}

// generateTransaction assigns the nonce, signs the resulting transaction and saves it
func (txq *TxQueue) generateTransaction(requestMetadata *txRequestData) error {
	opts := bind.NewKeyedTransactor(txq.privateKey)
	opts.Context = txq.ctx

	nonce, err := txq.backend.PendingNonceAt(txq.ctx, opts.From)
	if err != nil {
		return txq.finalizeRequestCancelled(requestMetadata, err)
	}

	request := requestMetadata.Request
	if request.GasLimit == 0 {
		gasLimit, err := request.EstimateGas(txq.ctx, txq.backend, opts.From)
		if err != nil {
			return txq.finalizeRequestCancelled(requestMetadata, err)
		}
		request.GasLimit = gasLimit
	}

	if request.GasPrice == nil {
		request.GasPrice, err = txq.backend.SuggestGasPrice(txq.ctx)
		if err != nil {
			return txq.finalizeRequestCancelled(requestMetadata, err)
		}
	}

	tx := types.NewTransaction(
		nonce,
		request.To,
		request.Value,
		request.GasLimit,
		request.GasPrice,
		request.Data,
	)

	requestMetadata.Transaction, err = opts.Signer(&types.HomesteadSigner{}, opts.From, tx)
	if err != nil {
		return txq.finalizeRequestCancelled(requestMetadata, err)
	}
	requestMetadata.State = TxRequestStateSigned
	return txq.store.Put(txq.requestKey(requestMetadata.ID), requestMetadata)
}

// sendTransaction sends the signed transaction to the ethereum backend
func (txq *TxQueue) sendTransaction(requestMetadata *txRequestData) error {
	err := txq.backend.SendTransactionWithID(txq.ctx, requestMetadata.ID, requestMetadata.Transaction)
	txq.lock.Lock()
	defer txq.lock.Unlock()
	if err != nil {
		// even if SendTransactionRequest returns an error there are still certain rare edge cases where the transaction might still be sent so we mark it as status unknown
		return txq.finalizeRequestStatusUnknown(requestMetadata, err.Error())
	}
	// if we have a hash we mark the transaction as pending
	batch := new(state.StoreBatch)
	requestMetadata.State = TxRequestStatePending
	err = batch.Put(txq.requestKey(requestMetadata.ID), requestMetadata)
	if err != nil {
		return err
	}
	trigger, err := txq.notify(batch, requestMetadata.ID, requestMetadata.HandlerID, txPendingNotificationType, &TxPendingNotification{
		Transaction: requestMetadata.Transaction,
	})
	if err != nil {
		return err
	}
	err = txq.store.WriteBatch(batch)
	if err != nil {
		return err
	}
	trigger()
	return nil
}

// processActiveRequest continues monitoring the active request if there is one
// this is called on startup before the queue begins normal operation
func (txq *TxQueue) processActiveRequest() error {
	// get the stored active request key
	// if nothing is stored id will remain 0 (which is invalid as ids start with 1)
	var id uint64
	err := txq.store.Get(txq.activeRequestKey(), &id)
	if err == state.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	// load the request metadata
	var requestMetadata txRequestData
	err = txq.store.Get(txq.requestKey(id), &requestMetadata)
	if err != nil {
		return err
	}

	// continue processing as regular
	return txq.processRequest(&requestMetadata)

}

// waitForActiveTransaction waits for requestMetadata to be mined and resets the active transaction afterwards
// the transaction will also be considered mined once the notification was queued successfully
// this only returns an error if the encoding fails which is an unrecoverable error
// the txqueue lock must not be held
func (txq *TxQueue) waitForActiveTransaction(requestMetadata *txRequestData) error {
	ctx, cancel := context.WithTimeout(txq.ctx, 20*time.Minute)
	defer cancel()

	// an error here means the context was cancelled
	receipt, err := WaitMined(ctx, txq.backend, requestMetadata.Transaction.Hash())
	txq.lock.Lock()
	defer txq.lock.Unlock()
	if err != nil {
		// if the main context of the TxQueue was cancelled we log and return
		if txq.ctx.Err() != nil {
			log.Info("terminating transaction queue while waiting for a transaction", "hash", requestMetadata.Transaction.Hash())
			return nil
		}

		// if the timeout context expired we mark the transaction status as unknown
		// future versions of the queue (with local nonce-tracking) should keep note of that and reuse the nonce for the next request
		log.Warn("transaction timeout reached", "hash", requestMetadata.Transaction.Hash())
		return txq.finalizeRequestStatusUnknown(requestMetadata, "transaction timed out")
	}

	return txq.finalizeRequestConfirmed(requestMetadata, *receipt)
}

// processQueue is the main transaction processing function of the TxQueue
// first it checks if there already is an active request. If so it processes this first
// then it will take requests from the queue in a loop and execute those
func (txq *TxQueue) processQueue() error {
	err := txq.processActiveRequest()
	if err != nil {
		return err
	}

	for {
		select {
		case <-txq.ctx.Done():
			return nil
		default:
		}

		requestMetadata, err := txq.waitForNextRequest()
		if err != nil {
			return err
		}

		err = txq.processRequest(requestMetadata)
		if err != nil {
			return err
		}
	}
}
