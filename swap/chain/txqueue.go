package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethersphere/swarm/state"
)

// TxQueue is a TxScheduler which sends transactions in sequence
// A new transaction is only sent after the previous one confirmed
// This done to minimize the chance of wrong nonce use
type TxQueue struct {
	lock      sync.Mutex         // lock for the entire queue
	ctx       context.Context    // context used for all network requests and waiting operations to ensure the queue can be stopped at any point
	cancel    context.CancelFunc // function to cancel the above context
	wg        sync.WaitGroup     // used to ensure that all background go routines have finished before Stop returns
	started   bool               // bool indicating that the queue has been started. used to ensure it does not run multiple times simultaneously
	errorChan chan error         // channel to stop the queue in case of errors

	store              state.Store                            // state store to use as the backend
	prefix             string                                 // all keys in the state store are prefixed with this
	requestQueue       *PersistentQueue                       // queue for all future requests
	handlers           map[TxRequestTypeID]*TxRequestHandlers // map from request type ids to their registered handlers
	notificationQueues map[TxRequestTypeID]*PersistentQueue   // map from request type ids to the notification queue of that handler

	backend    Backend           // ethereum backend to use
	privateKey *ecdsa.PrivateKey // private key used to sign transactions
}

// TxRequestMetadata is the metadata the queue saves for every request
type TxRequestMetadata struct {
	RequestTypeID TxRequestTypeID // the type id of this request
	State         TxRequestState  // the state this request is in
	Hash          common.Hash     // the hash of the associated transaction (if already sent)
}

// NotificationQueueItem is the metadata the queue saves for every pending notification
type NotificationQueueItem struct {
	NotificationType string // the type of the notification
	RequestID        uint64 // the request this notification is for
}

// ErrNoHandler is the error used if a request cannot be sent because no handler was registered for it
var ErrNoHandler = errors.New("no handler")

// NewTxQueue creates a new TxQueue
func NewTxQueue(store state.Store, prefix string, backend Backend, privateKey *ecdsa.PrivateKey) *TxQueue {
	txq := &TxQueue{
		store:              store,
		prefix:             prefix,
		handlers:           make(map[TxRequestTypeID]*TxRequestHandlers),
		notificationQueues: make(map[TxRequestTypeID]*PersistentQueue),
		backend:            backend,
		privateKey:         privateKey,
		requestQueue:       NewPersistentQueue(store, prefix+"_requestQueue_"),
		errorChan:          make(chan error, 1),
	}
	txq.ctx, txq.cancel = context.WithCancel(context.Background())
	return txq
}

// requestKey returns the database key for the TxRequestMetadata data
func (txq *TxQueue) requestKey(id uint64) string {
	return fmt.Sprintf("%s_requests_%d", txq.prefix, id)
}

// requestDataKey returns the database key for the custom TxRequest
func (txq *TxQueue) requestDataKey(id uint64) string {
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

// stopWithError sends the error to the error channel
func (txq *TxQueue) stopWithError(err error) {
	select {
	case txq.errorChan <- err:
	default:
	}
}

// ScheduleRequest adds a new request to be processed
// The request is assigned an id which is returned
func (txq *TxQueue) ScheduleRequest(requestTypeID TxRequestTypeID, request interface{}) (id uint64, err error) {
	txq.lock.Lock()
	defer txq.lock.Unlock()

	// get the last id
	idKey := txq.prefix + "_request_id"
	err = txq.store.Get(idKey, &id)
	if err != nil && err != state.ErrNotFound {
		return 0, err
	}
	// ids start at 1
	id++

	// in a single batch this
	// * stores the request data
	// * stores the request metadata
	// * adds it to the queue
	batch := new(state.StoreBatch)
	batch.Put(idKey, id)
	err = batch.Put(txq.requestDataKey(id), request)
	if err != nil {
		return 0, err
	}

	err = batch.Put(txq.requestKey(id), &TxRequestMetadata{
		RequestTypeID: requestTypeID,
		State:         TxRequestStateQueued,
	})
	if err != nil {
		return 0, err
	}

	_, triggerQueue, err := txq.requestQueue.Queue(batch, id)
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

// GetRequest load the serialized request data from disk and tries to decode it
func (txq *TxQueue) GetRequest(id uint64, request interface{}) error {
	return txq.store.Get(txq.requestDataKey(id), &request)
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
		err := txq.loop()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error("transaction queue terminated with an error", "queue", txq.prefix, "error", err)
		}
		txq.wg.Done()
	}()

	go func() {
		select {
		case err := <-txq.errorChan:
			log.Error("unrecoverable transaction queue error (transaction processing disabled)", "error", err)
			txq.Stop()
		case <-txq.ctx.Done():
		}
		txq.wg.Done()
	}()
}

// Stop stops processing transactions if it is running
// It will block until processing has terminated
func (txq *TxQueue) Stop() {
	txq.lock.Lock()

	if !txq.started {
		txq.lock.Unlock()
		return
	}

	txq.cancel()
	txq.lock.Unlock()
	// wait until all routines have finished
	txq.wg.Wait()
}

// getNotificationQueue gets the notification queue for a handler
// it initializes the struct if it does not yet exist
// the TxQueue lock must be held
func (txq *TxQueue) getNotificationQueue(requestTypeID TxRequestTypeID) *PersistentQueue {
	queue, ok := txq.notificationQueues[requestTypeID]
	if !ok {
		queue = NewPersistentQueue(txq.store, fmt.Sprintf("%s_notify_%s_%s", txq.prefix, requestTypeID.Handler, requestTypeID.RequestType))
		txq.notificationQueues[requestTypeID] = queue
	}
	return queue
}

// SetHandlers registers the handlers for the given TxRequestTypeID
// This starts the delivery of notifications for this TxRequestTypeID
func (txq *TxQueue) SetHandlers(requestTypeID TxRequestTypeID, handlers *TxRequestHandlers) error {
	txq.lock.Lock()
	defer txq.lock.Unlock()

	if txq.handlers[requestTypeID] != nil {
		return fmt.Errorf("handlers for %v.%v already set", requestTypeID.Handler, requestTypeID.RequestType)
	}
	txq.handlers[requestTypeID] = handlers
	notifyQueue := txq.getNotificationQueue(requestTypeID)

	// go routine processing the notification queue for this handler
	txq.wg.Add(1)
	go func() {
		defer txq.wg.Done()

		for {
			var item NotificationQueueItem
			// get the next notification item
			key, err := notifyQueue.Next(txq.ctx, &item, &txq.lock)
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
			case "TxReceiptNotification":
				notification = &TxReceiptNotification{}
			case "TxStateChangedNotification":
				notification = &TxStateChangedNotification{}
			}

			err = txq.store.Get(txq.notificationKey(key), notification)
			if err != nil {
				txq.stopWithError(fmt.Errorf("could not read notification: %v", err))
				return
			}

			switch item.NotificationType {
			case "TxReceiptNotification":
				if handlers.NotifyReceipt != nil {
					err = handlers.NotifyReceipt(txq.ctx, item.RequestID, notification.(*TxReceiptNotification))
				}
			case "TxStateChangedNotification":
				if handlers.NotifyStateChanged != nil {
					err = handlers.NotifyStateChanged(txq.ctx, item.RequestID, notification.(*TxStateChangedNotification))
				}
			}

			// if a handler failed we will try again in 10 seconds
			if err != nil {
				log.Info("transaction request handler failed", "type", item.NotificationType, "request", item.RequestID, "error", err)
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
			notifyQueue.Delete(batch, key)
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

// updateRequestStatus is a helper function to change the state of a transaction request while also emitting a notification
// in one write batch it
// * adds a TxStateChangedNotification notification to the notification queue
// * stores the corresponding notification
// * updates the state of requestMetadata and persists it
// it returns the trigger function which must be called once the batch was written
// this only returns an error if the encoding fails which is an unrecoverable error
// must be called with the txqueue lock held
func (txq *TxQueue) updateRequestStatus(batch *state.StoreBatch, id uint64, requestMetadata *TxRequestMetadata, newState TxRequestState) (triggerNotifyQueue func(), err error) {
	notifyQueue := txq.getNotificationQueue(requestMetadata.RequestTypeID)
	key, triggerNotifyQueue, err := notifyQueue.Queue(batch, &NotificationQueueItem{
		RequestID:        id,
		NotificationType: "TxStateChangedNotification",
	})
	if err != nil {
		return nil, fmt.Errorf("could not serialize notification queue item: %v", err)
	}

	err = batch.Put(txq.notificationKey(key), &TxStateChangedNotification{
		OldState: requestMetadata.State,
		NewState: newState,
	})
	if err != nil {
		return nil, fmt.Errorf("could not serialize notification: %v", err)
	}
	requestMetadata.State = newState
	batch.Put(txq.requestKey(id), requestMetadata)

	return triggerNotifyQueue, nil
}

// waitForNextRequest waits for the next request and sets it as the active request
// the txqueue lock must not be held
func (txq *TxQueue) waitForNextRequest() (id uint64, requestMetadata *TxRequestMetadata, err error) {
	// get the id of the next request in the queue
	key, err := txq.requestQueue.Next(txq.ctx, &id, &txq.lock)
	if err != nil {
		return 0, nil, err
	}
	defer txq.lock.Unlock()

	err = txq.store.Get(txq.requestKey(id), &requestMetadata)
	if err != nil {
		return 0, nil, err
	}

	handlers := txq.handlers[requestMetadata.RequestTypeID]
	if handlers == nil || handlers.Send == nil {
		// if there is no handler for this handler available we mark the request as cancelled and remove it from the queue
		batch := new(state.StoreBatch)
		triggerNotifyQueue, err := txq.updateRequestStatus(batch, id, requestMetadata, TxRequestStateCancelled)
		if err != nil {
			return 0, nil, err
		}

		txq.requestQueue.Delete(batch, key)
		err = txq.store.WriteBatch(batch)
		if err != nil {
			return 0, nil, err
		}

		triggerNotifyQueue()
		return 0, nil, ErrNoHandler
	}

	// if the request was successfully decoded it is removed from the queue and set as the active request
	batch := new(state.StoreBatch)
	err = batch.Put(txq.activeRequestKey(), id)
	if err != nil {
		return 0, nil, fmt.Errorf("could not put id write into batch: %v", err)
	}

	txq.requestQueue.Delete(batch, key)
	if err = txq.store.WriteBatch(batch); err != nil {
		return 0, nil, err
	}
	return id, requestMetadata, nil
}

// sendRequest sends the request with the given id and TxRequestMetadata
// the txqueue lock must not be held
func (txq *TxQueue) sendRequest(id uint64, requestMetadata *TxRequestMetadata) error {
	// finally we call the handler to send the actual transaction
	opts := bind.NewKeyedTransactor(txq.privateKey)
	opts.Context = txq.ctx
	handlers := txq.handlers[requestMetadata.RequestTypeID]
	hash, err := handlers.Send(id, txq.backend, opts)
	txq.lock.Lock()
	defer txq.lock.Unlock()
	if err != nil {
		// even if SendTransactionRequest returns an error there are still certain rare edge cases where the transaction might still be sent so we mark it as status unknown
		// in the future there should be a special error type to indicate the transaction was never sent
		batch := new(state.StoreBatch)
		triggerNotifyQueue, err := txq.updateRequestStatus(batch, id, requestMetadata, TxRequestStateStatusUnknown)
		if err != nil {
			return fmt.Errorf("failed to write transaction request status to store: %v", err)
		}

		if err = txq.store.WriteBatch(batch); err != nil {
			return err
		}
		triggerNotifyQueue()
		return nil
	}

	// if we have a hash we mark the transaction as pending
	batch := new(state.StoreBatch)
	requestMetadata.Hash = hash
	triggerNotifyQueue, err := txq.updateRequestStatus(batch, id, requestMetadata, TxRequestStatePending)
	if err != nil {
		return fmt.Errorf("failed to write transaction request status to store: %v", err)
	}

	if err = txq.store.WriteBatch(batch); err != nil {
		return err
	}
	triggerNotifyQueue()
	return nil
}

// processActiveRequest continues monitoring the active request if there is one
// this is called on startup before the queue begins normal operation
func (txq *TxQueue) processActiveRequest() error {
	// get the stored active request key
	// if nothing is stored id will remain 0 (which is invalid as ids start with 1)
	var id uint64
	err := txq.store.Get(txq.activeRequestKey(), &id)
	if err != nil && err != state.ErrNotFound {
		return err
	}

	// if there is a non-zero id there is an active request
	if id != 0 {
		// load the request metadata
		var requestMetadata TxRequestMetadata
		err = txq.store.Get(txq.requestKey(id), &requestMetadata)
		if err != nil {
			return err
		}

		log.Info("continuing to wait for previous transaction", "hash", requestMetadata.Hash)
		txq.lock.Lock()
		switch requestMetadata.State {
		// if the transaction is still in the Queued state we cannot know for sure where the process terminated
		// with a very high likelihood the transaction was not yet sent, but we cannot be sure of that
		// the transaction is marked as TransactionStatusUnknown and removed as the active transaction
		// in that rare case nonce issue might arise in subsequent requests
		case TxRequestStateQueued:
			defer txq.lock.Unlock()
			batch := new(state.StoreBatch)
			triggerNotifyQueue, err := txq.updateRequestStatus(batch, id, &requestMetadata, TxRequestStateStatusUnknown)
			// this only returns an error if the encoding fails which is an unrecoverable error
			if err != nil {
				return err
			}
			batch.Delete(txq.activeRequestKey())
			if err := txq.store.WriteBatch(batch); err != nil {
				return err
			}
			triggerNotifyQueue()
		// if the transaction is in the pending state this means we were previously waiting for the transaction
		case TxRequestStatePending:
			txq.lock.Unlock()
			// this only returns an error if the encoding fails which is an unrecoverable error
			if err := txq.waitForActiveTransaction(id, &requestMetadata); err != nil {
				return err
			}
		default:
			defer txq.lock.Unlock()
			// this indicates a client bug
			log.Error("found active transaction in unexpected state", "state", requestMetadata.State)
			if err := txq.store.Delete(txq.activeRequestKey()); err != nil {
				return err
			}
		}
	}
	return nil
}

// waitForActiveTransaction waits for requestMetadata to be mined and resets the active transaction  afterwards
// the transaction will also be considered mine once the notification was queued successfully
// this only returns an error if the encoding fails which is an unrecoverable error
// the txqueue lock must not be held
func (txq *TxQueue) waitForActiveTransaction(id uint64, requestMetadata *TxRequestMetadata) error {
	ctx, cancel := context.WithTimeout(txq.ctx, 20*time.Minute)
	defer cancel()

	// an error here means the context was cancelled
	receipt, err := WaitMined(ctx, txq.backend, requestMetadata.Hash)
	txq.lock.Lock()
	defer txq.lock.Unlock()
	if err != nil {
		// if the main context of the TxQueue was cancelled we log and return
		if txq.ctx.Err() != nil {
			log.Info("terminating transaction queue while waiting for a transaction", "hash", requestMetadata.Hash)
			return nil
		}

		// if the timeout context expired we mark the transaction status as unknown
		// future versions of the queue (with local nonce-tracking) should keep note of that and reuse the nonce for the next request
		log.Warn("transaction timeout reached", "hash", requestMetadata.Hash)
		batch := new(state.StoreBatch)
		triggerNotifyQueue, err := txq.updateRequestStatus(batch, id, requestMetadata, TxRequestStateStatusUnknown)
		if err != nil {
			return err
		}
		err = batch.Put(txq.activeRequestKey(), nil)
		if err != nil {
			return err
		}

		if err = txq.store.WriteBatch(batch); err != nil {
			return err
		}

		triggerNotifyQueue()
		return nil
	}

	// if the transaction is mined we need to
	// * update the request state and emit the corresponding notification
	// * emit a TxReceiptNotification
	// * reset the active request
	notifyQueue := txq.getNotificationQueue(requestMetadata.RequestTypeID)
	batch := new(state.StoreBatch)
	triggerNotifyQueueStateChanged, err := txq.updateRequestStatus(batch, id, requestMetadata, TxRequestStateConfirmed)
	if err != nil {
		return err
	}

	key, triggerNotifyQueueReceipt, err := notifyQueue.Queue(batch, &NotificationQueueItem{
		RequestID:        id,
		NotificationType: "TxReceiptNotification",
	})
	if err != nil {
		return err
	}

	batch.Put(txq.notificationKey(key), &TxReceiptNotification{
		Receipt: *receipt,
	})

	err = batch.Put(txq.activeRequestKey(), nil)
	if err != nil {
		return err
	}

	if err = txq.store.WriteBatch(batch); err != nil {
		return err
	}

	triggerNotifyQueueStateChanged()
	triggerNotifyQueueReceipt()
	return nil
}

// loop is the main transaction processing function of the TxQueue
// first it checks if there already is an active request. If so it processes this first
// then it will take requests from the queue in a loop and execute those
func (txq *TxQueue) loop() error {
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

		id, requestMetadata, err := txq.waitForNextRequest()
		if err == ErrNoHandler {
			continue
		}
		if err != nil {
			return err
		}

		err = txq.sendRequest(id, requestMetadata)
		if err != nil {
			return err
		}

		err = txq.waitForActiveTransaction(id, requestMetadata)
		if err != nil {
			// this only returns an error if the encoding fails which is an unrecoverable error
			return fmt.Errorf("error while waiting for transaction: %v", err)
		}
	}
}
