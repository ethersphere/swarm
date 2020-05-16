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

package prod

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	ctest "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/pss"
	psstest "github.com/ethersphere/swarm/pss/testing"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
)

// TestRecoveryHook tests that a recovery hook can be created and called
func TestRecoveryHook(t *testing.T) {
	// test variables needed to be correctly set for any recovery hook to reach the sender func
	chunkAddr := ctest.GenerateTestRandomChunk().Address()
	ctx := context.WithValue(context.Background(), "publisher", "0226f213613e843a413ad35b40f193910d26eb35f00154afcde9ded57479a6224a")
	handler := newTestRecoveryFeedsHandler(t)
	fallbackPublisher := ""

	// setup the sender
	hookWasCalled := false // test variable to check if hook is called
	testSender := func(ctx context.Context, targets trojan.Targets, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		hookWasCalled = true
		return nil, nil
	}

	// create recovery hook and call it
	recoveryHook := NewRecoveryHook(testSender, handler, fallbackPublisher)
	if err := recoveryHook(ctx, chunkAddr); err != nil {
		t.Fatal(err)
	}

	if hookWasCalled != true {
		t.Fatalf("recovery hook was not called")
	}
}

// RecoveryHookTestCase is a struct used as test cases for the TestRecoveryHookCalls func
type RecoveryHookTestCase struct {
	name           string
	ctx            context.Context
	feedsHandler   feed.GenericHandler
	expectsFailure bool
}

// TestRecoveryHookCalls verifies that recovery hooks are being called as expected when net store attempts to get a chunk
func TestRecoveryHookCalls(t *testing.T) {
	// generate test chunk and store
	netStore := newTestNetStore(t)
	c := ctest.GenerateTestRandomChunk()
	ref := c.Address()

	// test cases variables
	dummyContext := context.Background() // has no publisher
	publisherContext := context.WithValue(context.Background(), "publisher", "0226f213613e843a413ad35b40f193910d26eb35f00154afcde9ded57479a6224a")
	dummyHandler := feed.NewDummyHandler() // returns empty content for feed
	feedsHandler := newTestRecoveryFeedsHandler(t)

	for _, tc := range []RecoveryHookTestCase{
		{
			name:           "no publisher, no feed content",
			ctx:            dummyContext,
			feedsHandler:   dummyHandler,
			expectsFailure: true,
		},
		{
			name:           "publisher set, no feed content",
			ctx:            publisherContext,
			feedsHandler:   dummyHandler,
			expectsFailure: true,
		},
		{
			name:           "feed content set, no publisher",
			ctx:            dummyContext,
			feedsHandler:   feedsHandler,
			expectsFailure: true,
		},
		{
			name:           "publisher and feed content set",
			ctx:            publisherContext,
			feedsHandler:   feedsHandler,
			expectsFailure: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hookWasCalled := make(chan bool, 1) // channel to check if hook is called

			// setup recovery hook
			testHook := func(ctx context.Context, targets trojan.Targets, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
				hookWasCalled <- true
				return nil, nil
			}
			recoverFunc := NewRecoveryHook(testHook, tc.feedsHandler, "")

			// set hook in net store
			netStore.WithRecoveryCallback(recoverFunc)

			// fetch test chunk
			netStore.Get(tc.ctx, chunk.ModeGetRequest, storage.NewRequest(ref))

			// checks whether the callback is invoked or the test case times out
			select {
			case <-hookWasCalled:
				if !tc.expectsFailure {
					return
				} else {
					t.Fatal("recovery hook was unexpectedly called")
				}
			case <-time.After(100 * time.Millisecond):
				if tc.expectsFailure {
					return
				} else {
					t.Fatal("recovery hook was not called when expected")
				}
			}
		})
	}
}

// newTestNetStore creates a test store with a set RemoteGet func
func newTestNetStore(t *testing.T) *storage.NetStore {
	// generate address
	baseKey := make([]byte, 32)
	_, err := rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}
	baseAddress := network.NewBzzAddr(baseKey, baseKey)

	// generate net store
	tags := chunk.NewTags()
	localStore := psstest.NewMockLocalStore(t, tags)
	lstore := chunk.NewValidatorStore(
		localStore,
		storage.NewContentAddressValidator(storage.MakeHashFunc(storage.DefaultHash)),
	)
	netStore := storage.NewNetStore(lstore, baseAddress)

	// generate retrieval
	kad := network.NewKademlia(baseAddress.Over(), network.NewKadParams())
	ret := retrieval.New(kad, netStore, baseAddress, nil)

	// set retrieval on netstore and return
	netStore.RemoteGet = ret.RequestFromPeers
	return netStore
}

// newTestRecoveryFeedsHandler returns a DummyHandler with binary content which can be correctly unmarshalled
func newTestRecoveryFeedsHandler(t *testing.T) *feed.DummyHandler {
	h := feed.NewDummyHandler()

	// test targets
	t1 := trojan.Target([]byte{57, 120})
	t2 := trojan.Target([]byte{209, 156})
	t3 := trojan.Target([]byte{156, 38})
	targets := trojan.Targets([]trojan.Target{t1, t2, t3})

	// marshal into bytes and set as mock feed content
	b, err := json.Marshal(targets)
	if err != nil {
		t.Fatal(err)
	}
	h.SetContent(b)

	return h
}
