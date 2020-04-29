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
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

// TestRecoveryHook tests that a timeout in netstore
// invokes correctly recovery hook
func TestRecoveryHook(t *testing.T) {
	// setup recovery hook
	// verify that hook is correctly invoked
	ctx := context.TODO()

	handlerWasCalled := false // test variable to check handler funcs are correctly retrieved

	// register first handler
	testHandler := func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		handlerWasCalled = true
		return nil, nil
	}

	recoverFunc := NewRecoveryHook(testHandler)
	// call recoverFunc
	recoverFunc(ctx, chunk.ZeroAddr)

	// verify the hook has been called
	if handlerWasCalled != true {
		t.Fatalf("unexpected result for prod Recover func, expected test variable to have a value of %v but is %v instead", true, handlerWasCalled)
	}

}

func newMockLocalStore(t *testing.T, tags *chunk.Tags) *localstore.DB {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{Tags: tags})
	if err != nil {
		t.Fatal(err)
	}

	return localStore
}

// TestSenderCall verifies that pss send is being called correctly
func TestSenderCall(t *testing.T) {
	ctx := context.TODO()
	tags := chunk.NewTags()
	localStore := newMockLocalStore(t, tags)

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

	handlerWasCalled := false // test variable to check handler funcs are correctly retrieved

	// setup recovery hook
	testHandler := func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		handlerWasCalled = true
		return nil, nil
	}

	//func(ctx context.Context, chunkAddress chunk.Address)
	recoverFunc := NewRecoveryHook(testHandler)
	netStore.WithRecoveryCallback(recoverFunc)

	c := GenerateTestRandomChunk()
	ref := c.Address()

	kad := network.NewKademlia(baseAddress.Over(), network.NewKadParams())
	ret := retrieval.New(kad, netStore, baseAddress, nil)
	netStore.RemoteGet = ret.RequestFromPeers

	netStore.Get(ctx, chunk.ModeGetRequest, storage.NewRequest(ref))

	for {
		// waits until the callback is called or timeout
		select {
		default:
			if handlerWasCalled {
				return
			}
		// TODO: change the timeout
		case <-time.After(timeouts.FetcherGlobalTimeout):
			t.Fatalf("no handler was called")
		}
	}

}

func GenerateTestRandomChunk() chunk.Chunk {
	data := make([]byte, chunk.DefaultSize)
	rand.Read(data)
	key := make([]byte, 32)
	rand.Read(key)
	return chunk.NewChunk(key, data)
}
