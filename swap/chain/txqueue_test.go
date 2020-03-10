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
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/state"
)

var (
	senderKey, _  = crypto.HexToECDSA("634fb5a872396d9693e5c9f9d7233cfa93f395c093371017ff44aa9ae6564cdd")
	senderAddress = crypto.PubkeyToAddress(senderKey.PublicKey)
)

var defaultBackend = backends.NewSimulatedBackend(core.GenesisAlloc{
	senderAddress: {Balance: big.NewInt(1000000000000000000)},
}, 8000000)

// backend.SendTransaction outcome associated with a request id
type testRequestOutcome struct {
	noCommit  bool  // the backend should not automatically mine the transaction
	sendError error // SendTransaction should return with this error
}

// testTxSchedulerBackend wraps a SimulatedBackend and provides a way to determine the result of SendTransaction
type testTxSchedulerBackend struct {
	*backends.SimulatedBackend
	requestOutcomes map[uint64]testRequestOutcome // map of request id to outcome
	lock            sync.Mutex                    // lock for map access and blocking SendTransactionWithID
}

func newTestTxSchedulerBackend(backend *backends.SimulatedBackend) *testTxSchedulerBackend {
	return &testTxSchedulerBackend{
		SimulatedBackend: backend,
		requestOutcomes:  make(map[uint64]testRequestOutcome),
	}
}

func (b *testTxSchedulerBackend) SendTransactionWithID(ctx context.Context, id uint64, tx *types.Transaction) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	outcome, ok := b.requestOutcomes[id]
	if ok {
		if outcome.sendError != nil {
			return outcome.sendError
		}
		err := b.SimulatedBackend.SendTransaction(ctx, tx)
		if err == nil && !outcome.noCommit {
			b.SimulatedBackend.Commit()
		}
		return err
	}
	err := b.SimulatedBackend.SendTransaction(ctx, tx)
	if err == nil {
		b.SimulatedBackend.Commit()
	}
	return err
}

const testHandlerID = "test_TestRequest"

// txSchedulerTester is a helper used for testing TxScheduler implementations
// it saves received notifications to channels so they can easily be checked in tests
// furthermore it can trigger certain errors depending on flags set in the requests
type txSchedulerTester struct {
	lock        sync.Mutex
	txScheduler TxScheduler
	chans       map[uint64]*txSchedulerTesterRequestData // map from request id to channels
	backend     *testTxSchedulerBackend
}

// txSchedulerTesterRequestData is the data txSchedulerTester saves for every request
type txSchedulerTesterRequestData struct {
	ReceiptNotification       chan *TxReceiptNotification
	CancelledNotification     chan *TxCancelledNotification
	PendingNotification       chan *TxPendingNotification
	StatusUnknownNotification chan *TxStatusUnknownNotification
	request                   TxRequest
}

type txSchedulerTesterRequestExtraData struct {
}

func newTxSchedulerTester(backend *testTxSchedulerBackend, txScheduler TxScheduler) (*txSchedulerTester, error) {
	tc := &txSchedulerTester{
		txScheduler: txScheduler,
		backend:     backend,
		chans:       make(map[uint64]*txSchedulerTesterRequestData),
	}
	err := tc.setHandlers(txScheduler)
	if err != nil {
		return nil, err
	}
	return tc, nil
}

// hooks up the TxScheduler handlers to the txSchedulerTester channels
func (tc *txSchedulerTester) setHandlers(txScheduler TxScheduler) error {
	return txScheduler.SetHandlers(testHandlerID, &TxRequestHandlers{
		NotifyReceipt: func(ctx context.Context, id uint64, notification *TxReceiptNotification) error {
			select {
			case tc.getRequest(id).ReceiptNotification <- notification:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		NotifyCancelled: func(ctx context.Context, id uint64, notification *TxCancelledNotification) error {
			select {
			case tc.getRequest(id).CancelledNotification <- notification:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		NotifyPending: func(ctx context.Context, id uint64, notification *TxPendingNotification) error {
			select {
			case tc.getRequest(id).PendingNotification <- notification:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		NotifyStatusUnknown: func(ctx context.Context, id uint64, notification *TxStatusUnknownNotification) error {
			select {
			case tc.getRequest(id).StatusUnknownNotification <- notification:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}

// schedule request with the provided extra data and the transaction outcome
func (tc *txSchedulerTester) schedule(request TxRequest, requestExtraData interface{}, outcome *testRequestOutcome) (uint64, error) {
	// this lock here is crucial as it blocks SendTransaction until the requestOutcomes has been set
	tc.backend.lock.Lock()
	defer tc.backend.lock.Unlock()
	id, err := tc.txScheduler.ScheduleRequest(testHandlerID, request, requestExtraData)
	if err != nil {
		return 0, err
	}
	if outcome != nil {
		tc.backend.requestOutcomes[id] = *outcome
	}

	tc.getRequest(id).request = request
	return id, nil
}

// getRequest gets the txSchedulerTesterRequestData for this id or initializes it if it does not yet exist
func (tc *txSchedulerTester) getRequest(id uint64) *txSchedulerTesterRequestData {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	c, ok := tc.chans[id]
	if !ok {
		tc.chans[id] = &txSchedulerTesterRequestData{
			ReceiptNotification:       make(chan *TxReceiptNotification),
			PendingNotification:       make(chan *TxPendingNotification),
			CancelledNotification:     make(chan *TxCancelledNotification),
			StatusUnknownNotification: make(chan *TxStatusUnknownNotification),
		}
		return tc.chans[id]
	}
	return c
}

// expectStateChangedNotification waits for a StateChangedNotification with the given parameters
func (tc *txSchedulerTester) expectStatusUnknownNotification(ctx context.Context, id uint64, reason string) error {
	var notification *TxStatusUnknownNotification
	request := tc.getRequest(id)
	select {
	case notification = <-request.StatusUnknownNotification:
	case <-ctx.Done():
		return ctx.Err()
	}

	if notification.Reason != reason {
		return fmt.Errorf("reason mismatch. got %s, expected %s", notification.Reason, reason)
	}

	return nil
}

func (tc *txSchedulerTester) expectPendingNotification(ctx context.Context, id uint64) error {
	var notification *TxPendingNotification
	request := tc.getRequest(id)
	select {
	case notification = <-request.PendingNotification:
	case <-ctx.Done():
		return ctx.Err()
	}

	tx := notification.Transaction
	if !bytes.Equal(tx.Data(), request.request.Data) {
		return fmt.Errorf("transaction data mismatch. got %v, expected %v", tx.Data(), request.request.Data)
	}

	if *tx.To() != request.request.To {
		return fmt.Errorf("transaction to mismatch. got %v, expected %v", tx.To(), request.request.To)
	}

	if tx.Value().Cmp(request.request.Value) != 0 {
		return fmt.Errorf("transaction value mismatch. got %v, expected %v", tx.Value(), request.request.Value)
	}

	return nil
}

// expectStateChangedNotification waits for a ReceiptNotification for the given request id and verifies its hash
func (tc *txSchedulerTester) expectReceiptNotification(ctx context.Context, id uint64) error {
	var notification *TxReceiptNotification
	request := tc.getRequest(id)
	select {
	case notification = <-request.ReceiptNotification:
	case <-ctx.Done():
		return ctx.Err()
	}

	tx, pending, err := tc.backend.TransactionByHash(ctx, notification.Receipt.TxHash)
	if err != nil {
		return err
	}
	if pending {
		return errors.New("received a receipt notification for a pending transaction")
	}

	if tx == nil {
		return errors.New("transaction not found")
	}

	if !bytes.Equal(tx.Data(), request.request.Data) {
		return fmt.Errorf("transaction data mismatch. got %v, expected %v", tx.Data(), request.request.Data)
	}

	if *tx.To() != request.request.To {
		return fmt.Errorf("transaction to mismatch. got %v, expected %v", tx.To(), request.request.To)
	}

	if tx.Value().Cmp(request.request.Value) != 0 {
		return fmt.Errorf("transaction value mismatch. got %v, expected %v", tx.Value(), request.request.Value)
	}

	return nil
}

// makeTestRequest creates a simple test request to the 0x0 address
func makeTestRequest() TxRequest {
	return TxRequest{
		To:    common.Address{},
		Value: big.NewInt(0),
		Data:  []byte{},
	}
}

// helper function for queue tests which sets up everything and provides a cleanup function
// if run is true the queue starts processing requests and cleanup function will wait for proper termination
func setupTxQueueTest(run bool) (*TxQueue, *testTxSchedulerBackend, func()) {
	backend := defaultBackend
	backend.Commit()

	testBackend := newTestTxSchedulerBackend(backend)

	store := state.NewInmemoryStore()
	txq := NewTxQueue(store, "test", testBackend, senderKey)
	if run {
		txq.Start()
	}
	return txq, testBackend, func() {
		if run {
			txq.Stop()
		}
		store.Close()
	}
}

// TestTxQueueScheduleRequest tests scheduling a single request when the queue is not running
// Afterwards the queue is started and the correct sequence of notifications is expected
func TestTxQueueScheduleRequest(t *testing.T) {
	txq, backend, clean := setupTxQueueTest(false)
	defer clean()
	tc, err := newTxSchedulerTester(backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	testRequest := &txSchedulerTesterRequestExtraData{}

	id, err := tc.schedule(makeTestRequest(), testRequest, nil)
	if err != nil {
		t.Fatal(err)
	}

	if id != 1 {
		t.Fatal("expected id to be 1")
	}

	var testRequestRetrieved *txSchedulerTesterRequestExtraData
	err = txq.GetExtraData(id, &testRequestRetrieved)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testRequest, testRequestRetrieved) {
		t.Fatalf("got request %v, expected %v", testRequestRetrieved, testRequest)
	}

	txq.Start()
	defer txq.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err = tc.expectPendingNotification(ctx, id); err != nil {
		t.Fatal(err)
	}

	if err = tc.expectReceiptNotification(ctx, id); err != nil {
		t.Fatal(err)
	}
}

// TestTxQueueManyRequests schedules many requests and expects all of them to be successful
func TestTxQueueManyRequests(t *testing.T) {
	txq, backend, clean := setupTxQueueTest(true)
	defer clean()
	tc, err := newTxSchedulerTester(backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	var ids []uint64
	count := 200
	for i := 0; i < count; i++ {
		id, err := tc.schedule(makeTestRequest(), &txSchedulerTesterRequestExtraData{}, nil)
		if err != nil {
			t.Fatal(err)
		}

		ids = append(ids, id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, id := range ids {
		err = tc.expectPendingNotification(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		err = tc.expectReceiptNotification(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestTxQueueActiveTransaction tests that the queue continues to monitor the last pending transaction
func TestTxQueueActiveTransaction(t *testing.T) {
	txq, backend, clean := setupTxQueueTest(false)
	defer clean()

	tc, err := newTxSchedulerTester(backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	txq.Start()

	id, err := tc.schedule(makeTestRequest(), 5, &testRequestOutcome{
		noCommit: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = tc.expectPendingNotification(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	txq.Stop()

	state, err := txq.GetRequestState(id)
	if err != nil {
		t.Fatal(err)
	}
	if state != TxRequestStatePending {
		t.Fatalf("state not pending, was %d", state)
	}

	// start a new queue with the same store and backend
	txq2 := NewTxQueue(txq.store, txq.prefix, txq.backend, txq.privateKey)
	if err != nil {
		t.Fatal(err)
	}
	// reuse the tester so it maintains state about the tx hash and id
	tc.setHandlers(txq2)

	if err != nil {
		t.Fatal(err)
	}

	// the transaction confirmed in the meantime
	backend.Commit()

	txq2.Start()
	defer txq2.Stop()

	err = tc.expectReceiptNotification(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
}

// TestTxQueueErrorDuringSend tests that a request is marked as TxRequestStateStatusUnknown if the send fails
func TestTxQueueErrorDuringSend(t *testing.T) {
	txq, backend, clean := setupTxQueueTest(true)
	defer clean()
	tc, err := newTxSchedulerTester(backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	id, err := tc.schedule(makeTestRequest(), 5, &testRequestOutcome{
		sendError: errors.New("test error"),
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = tc.expectStatusUnknownNotification(ctx, id, "test error")
	if err != nil {
		t.Fatal(err)
	}
}
