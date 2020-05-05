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
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	ctest "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/pss"
	psstest "github.com/ethersphere/swarm/pss/testing"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage"
)

// TestRecoveryHook tests that NewRecoveryHook has been properly invoked
func TestRecoveryHook(t *testing.T) {
	ctx := context.TODO()

	hookWasCalled := false // test variable to check hook func are correctly retrieved

	// setup the hook
	testHook := func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		hookWasCalled = true
		return nil, nil
	}

	// setup recovery hook with testHook
	recoverFunc := NewRecoveryHook(testHook)

	recoverFunc(ctx, chunk.ZeroAddr)

	// verify the hook has been called correctly
	if hookWasCalled != true {
		t.Fatalf("unexpected result for prod Recover func, expected test variable to have a value of %t but is %t instead", true, hookWasCalled)
	}

}

// TestSenderCall verifies that a hook is being called correctly within the netstore
func TestSenderCall(t *testing.T) {
	ctx := context.TODO()
	tags := chunk.NewTags()
	localStore := psstest.NewMockLocalStore(t, tags)

	lstore := chunk.NewValidatorStore(
		localStore,
		storage.NewContentAddressValidator(storage.MakeHashFunc(storage.DefaultHash)),
	)

	baseKey := make([]byte, 32)
	_, err := rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}

	baseAddress := network.NewBzzAddr(baseKey, baseKey)
	// setup netstore
	netStore := storage.NewNetStore(lstore, baseAddress)

	hookWasCalled := false // test variable to check hook func are correctly retrieved

	// setup recovery hook
	testHook := func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		hookWasCalled = true
		return nil, nil
	}

	recoverFunc := NewRecoveryHook(testHook)
	netStore.WithRecoveryCallback(recoverFunc)

	c := ctest.GenerateTestRandomChunk()
	ref := c.Address()

	kad := network.NewKademlia(baseAddress.Over(), network.NewKadParams())
	ret := retrieval.New(kad, netStore, baseAddress, nil)
	netStore.RemoteGet = ret.RequestFromPeers

	netStore.Get(ctx, chunk.ModeGetRequest, storage.NewRequest(ref))

	for {
		// waits until the callback is called or timeout
		select {
		default:
			if hookWasCalled {
				return
			}
		// TODO: change the timeout
		case <-time.After(timeouts.FetcherGlobalTimeout):
			t.Fatalf("no hook was called")
		}
	}

}
