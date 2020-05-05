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
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	ctest "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/timeouts"
	"github.com/ethersphere/swarm/pss"
	psstest "github.com/ethersphere/swarm/pss/testing"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
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
	testHandler := newTestHandler(t)
	recoverFunc := NewRecoveryHook(testHook, testHandler)

	// TODO: replace publisher byte string with equivalent zero value
	recoverFunc(ctx, chunk.ZeroAddr, "0226f213613e843a413ad35b40f193910d26eb35f00154afcde9ded57479a6224a")

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
	testHandler := newTestHandler(t)

	recoverFunc := NewRecoveryHook(testHook, testHandler)
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

func newTestHandler(t *testing.T) *feed.Handler {
	datadir, err := ioutil.TempDir("", "fh")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(datadir, "prod")
	fhParams := &feed.HandlerParams{}
	fh := feed.NewHandler(fhParams)

	db, err := localstore.New(path, make([]byte, 32), nil)
	if err != nil {
		t.Fatal(err)
	}

	localStore := chunk.NewValidatorStore(db, storage.NewContentAddressValidator(storage.MakeHashFunc(storage.SHA3Hash)), fh)

	netStore := storage.NewNetStore(localStore, network.NewBzzAddr(make([]byte, 32), nil))
	netStore.RemoteGet = func(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
		return nil, func() {}, errors.New("not found")
	}
	fh.SetStore(netStore)

	return fh
}
