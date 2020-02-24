package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/state"
	mock "github.com/ethersphere/swarm/swap/chain/mock"
)

var (
	senderKey, _  = crypto.HexToECDSA("634fb5a872396d9693e5c9f9d7233cfa93f395c093371017ff44aa9ae6564cdd")
	senderAddress = crypto.PubkeyToAddress(senderKey.PublicKey)
)

var defaultBackend = backends.NewSimulatedBackend(core.GenesisAlloc{
	senderAddress: {Balance: big.NewInt(1000000000000000000)},
}, 8000000)

func newTestBackend() *mock.TestBackend {
	return mock.NewTestBackend(defaultBackend)
}

var testRequestTypeID = TxRequestTypeID{
	Handler:     "test",
	RequestType: "TestRequest",
}

var dummyTypeID = TxRequestTypeID{
	Handler:     "dummy",
	RequestType: "DummyRequest",
}

// txSchedulerTester is a helper used for testing TxScheduler implementations
// it saves received notifications to channels so they can easily be checked in tests
// furthermore it can trigger certain errors depending on flags set in the requests
type txSchedulerTester struct {
	lock        sync.Mutex
	txScheduler TxScheduler
	chans       map[uint64]*txSchedulerTesterRequestData
	backend     Backend
}

// txSchedulerTesterRequestData is the data txSchedulerTester saves for every request
type txSchedulerTesterRequestData struct {
	ReceiptNotification      chan *TxReceiptNotification
	StateChangedNotification chan *TxStateChangedNotification
	hash                     common.Hash
}

type txSchedulerTesterRequest struct {
	NoCommit        bool // the transaction should not be automatically mined
	ErrorDuringSend bool // send should return with an error
}

func newTxSchedulerTester(backend Backend, txScheduler TxScheduler) (*txSchedulerTester, error) {
	t := &txSchedulerTester{
		txScheduler: txScheduler,
		backend:     backend,
		chans:       make(map[uint64]*txSchedulerTesterRequestData),
	}
	err := txScheduler.SetHandlers(testRequestTypeID, &TxRequestHandlers{
		Send:               t.SendTransactionRequest,
		NotifyReceipt:      t.NotifyReceipt,
		NotifyStateChanged: t.NotifyStateChanged,
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// getRequest gets the txSchedulerTesterRequestData for this id or initializes it if it does not yet exist
func (tc *txSchedulerTester) getRequest(id uint64) *txSchedulerTesterRequestData {
	tc.lock.Lock()
	defer tc.lock.Unlock()
	c, ok := tc.chans[id]
	if !ok {
		tc.chans[id] = &txSchedulerTesterRequestData{
			ReceiptNotification:      make(chan *TxReceiptNotification),
			StateChangedNotification: make(chan *TxStateChangedNotification),
		}
		return tc.chans[id]
	}
	return c
}

// expectStateChangedNotification waits for a StateChangedNotification with the given parameters
func (tc *txSchedulerTester) expectStateChangedNotification(ctx context.Context, id uint64, oldState TxRequestState, newState TxRequestState) error {
	var notification *TxStateChangedNotification
	request := tc.getRequest(id)
	select {
	case notification = <-request.StateChangedNotification:
	case <-ctx.Done():
		return ctx.Err()
	}

	if notification.OldState != oldState {
		return fmt.Errorf("wrong old state. got %v, expected %v", notification.OldState, oldState)
	}

	if notification.NewState != newState {
		return fmt.Errorf("wrong new state. got %v, expected %v", notification.NewState, newState)
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

	receipt, err := tc.backend.TransactionReceipt(ctx, request.hash)
	if err != nil {
		return err
	}
	if receipt == nil {
		return errors.New("no receipt found for transaction")
	}

	if notification.Receipt.TxHash != request.hash {
		return fmt.Errorf("wrong old state. got %v, expected %v", notification.Receipt.TxHash, request.hash)
	}

	return nil
}

func (tc *txSchedulerTester) NotifyReceipt(ctx context.Context, id uint64, notification *TxReceiptNotification) error {
	select {
	case tc.getRequest(id).ReceiptNotification <- notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (tc *txSchedulerTester) NotifyStateChanged(ctx context.Context, id uint64, notification *TxStateChangedNotification) error {
	select {
	case tc.getRequest(id).StateChangedNotification <- notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// simple low level transaction logic
func (tc *txSchedulerTester) SendTransactionRequest(id uint64, backend Backend, opts *bind.TransactOpts) (hash common.Hash, err error) {
	var nonce uint64
	if opts.Nonce == nil {
		nonce, err = backend.PendingNonceAt(opts.Context, opts.From)
		if err != nil {
			return common.Hash{}, err
		}
	} else {
		nonce = opts.Nonce.Uint64()
	}

	signed, err := opts.Signer(types.HomesteadSigner{}, opts.From, types.NewTransaction(nonce, common.Address{}, big.NewInt(0), 100000, big.NewInt(int64(10000000)), []byte{}))
	if err != nil {
		return common.Hash{}, err
	}

	var request *txSchedulerTesterRequest
	err = tc.txScheduler.GetRequest(id, &request)
	if err != nil {
		return common.Hash{}, err
	}

	if request.ErrorDuringSend {
		return common.Hash{}, errors.New("simulated error during send")
	}

	if request.NoCommit {
		err = backend.(*mock.TestBackend).SendTransactionNoCommit(opts.Context, signed)
	} else {
		err = backend.SendTransaction(opts.Context, signed)
	}
	if err != nil {
		return common.Hash{}, err
	}

	tc.getRequest(id).hash = signed.Hash()
	return signed.Hash(), nil
}

// helper function for queue tests which sets up everything and provides a cleanup function
// if run is true the queue starts processing requests and cleanup function will wait for proper termination
func setupTxQueueTest(run bool) (*TxQueue, func()) {
	backend := newTestBackend()
	store := state.NewInmemoryStore()
	txq := NewTxQueue(store, "test", backend, senderKey)
	if run {
		txq.Start()
	}
	return txq, func() {
		if run {
			txq.Stop()
		}
		store.Close()
		backend.Close()
	}
}

// TestTxQueueScheduleRequest tests scheduling a single request when the queue is not running
// Afterwards the queue is started and the correct sequence of notifications is expected
func TestTxQueueScheduleRequest(t *testing.T) {
	txq, clean := setupTxQueueTest(false)
	defer clean()
	tc, err := newTxSchedulerTester(txq.backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	testRequest := &txSchedulerTesterRequest{}

	id, err := txq.ScheduleRequest(testRequestTypeID, testRequest)
	if err != nil {
		t.Fatal(err)
	}

	if id != 1 {
		t.Fatal("expected id to be 1")
	}

	var testRequestRetrieved *txSchedulerTesterRequest
	err = txq.GetRequest(id, &testRequestRetrieved)
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

	if err = tc.expectStateChangedNotification(ctx, id, TxRequestStateQueued, TxRequestStatePending); err != nil {
		t.Fatal(err)
	}

	if err = tc.expectStateChangedNotification(ctx, id, TxRequestStatePending, TxRequestStateConfirmed); err != nil {
		t.Fatal(err)
	}

	if err = tc.expectReceiptNotification(ctx, id); err != nil {
		t.Fatal(err)
	}
}

// TestTxQueueManyRequests schedules many requests and expects all of them to be successful
func TestTxQueueManyRequests(t *testing.T) {
	txq, clean := setupTxQueueTest(true)
	defer clean()
	tc, err := newTxSchedulerTester(txq.backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	var ids []uint64
	count := 200
	for i := 0; i < count; i++ {
		id, err := txq.ScheduleRequest(testRequestTypeID, &txSchedulerTesterRequest{})
		if err != nil {
			t.Fatal(err)
		}

		ids = append(ids, id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, id := range ids {
		err := tc.expectStateChangedNotification(ctx, id, TxRequestStateQueued, TxRequestStatePending)
		if err != nil {
			t.Fatal(err)
		}
		err = tc.expectStateChangedNotification(ctx, id, TxRequestStatePending, TxRequestStateConfirmed)
		if err != nil {
			t.Fatal(err)
		}
		err = tc.expectReceiptNotification(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestTxQueueNoHandler schedules a request with no send handler
func TestTxQueueNoHandler(t *testing.T) {
	txq, clean := setupTxQueueTest(true)
	defer clean()
	tc, err := newTxSchedulerTester(txq.backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	txq.SetHandlers(dummyTypeID, &TxRequestHandlers{
		NotifyStateChanged: tc.NotifyStateChanged,
	})

	id, err := txq.ScheduleRequest(dummyTypeID, &txSchedulerTesterRequest{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = tc.expectStateChangedNotification(ctx, id, TxRequestStateQueued, TxRequestStateCancelled)
	if err != nil {
		t.Fatal(err)
	}
}

// TestTxQueueActiveTransaction tests that the queue continues to monitor the last pending transaction
func TestTxQueueActiveTransaction(t *testing.T) {
	txq, clean := setupTxQueueTest(false)
	defer clean()

	tc, err := newTxSchedulerTester(txq.backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	txq.Start()

	id, err := txq.ScheduleRequest(testRequestTypeID, &txSchedulerTesterRequest{
		NoCommit: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = tc.expectStateChangedNotification(ctx, id, TxRequestStateQueued, TxRequestStatePending)
	if err != nil {
		t.Fatal(err)
	}

	txq.Stop()

	// start a new queue with the same store and backend
	txq2 := NewTxQueue(txq.store, txq.prefix, txq.backend, txq.privateKey)
	if err != nil {
		t.Fatal(err)
	}
	// reuse the tester so it maintains state about the tx hash and id
	err = txq2.SetHandlers(testRequestTypeID, &TxRequestHandlers{
		Send:               tc.SendTransactionRequest,
		NotifyReceipt:      tc.NotifyReceipt,
		NotifyStateChanged: tc.NotifyStateChanged,
	})
	if err != nil {
		t.Fatal(err)
	}

	// the transaction confirmed in the meantime
	txq2.backend.(*mock.TestBackend).Commit()

	txq2.Start()
	defer txq2.Stop()

	err = tc.expectStateChangedNotification(ctx, id, TxRequestStatePending, TxRequestStateConfirmed)
	if err != nil {
		t.Fatal(err)
	}

	err = tc.expectReceiptNotification(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
}

// TestTxQueueErrorDuringSend tests that a request is marked as TxRequestStateStatusUnknown if the send fails
func TestTxQueueErrorDuringSend(t *testing.T) {
	txq, clean := setupTxQueueTest(true)
	defer clean()
	tc, err := newTxSchedulerTester(txq.backend, txq)
	if err != nil {
		t.Fatal(err)
	}

	id, err := txq.ScheduleRequest(testRequestTypeID, &txSchedulerTesterRequest{
		ErrorDuringSend: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = tc.expectStateChangedNotification(ctx, id, TxRequestStateQueued, TxRequestStateStatusUnknown)
	if err != nil {
		t.Fatal(err)
	}
}
