// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/swarm/network"
)

var (
	errUnknown = errors.New("unknown error")
)

type mockRetrieve struct {
	requests map[string]int
}

func NewMockRetrieve() *mockRetrieve {
	return &mockRetrieve{requests: make(map[string]int)}
}

func (m *mockRetrieve) retrieve(ctx context.Context, r *Request) (quitc chan struct{}, err error) {

	hkey := hex.EncodeToString(r.Address())
	m.requests[hkey] += 1

	// on second call return error
	if m.requests[hkey] == 2 {
		return nil, errUnknown
	}

	// on third call return data
	if m.requests[hkey] == 3 {
		go func() {
			time.Sleep(100 * time.Millisecond)
			r.SetData(r.Address(), []byte{0, 1})
		}()

		return nil, nil
	}

	return nil, nil
}

func TestNetstoreFailedRequest(t *testing.T) {
	searchTimeout = 300 * time.Millisecond

	// setup
	addr := network.RandomAddr() // tested peers peer address

	// temp datadir
	datadir, err := ioutil.TempDir("", "netstore")
	if err != nil {
		t.Fatal(err)
	}
	params := NewDefaultLocalStoreParams()
	params.Init(datadir)
	params.BaseKey = addr.Over()
	localStore, err := NewTestLocalStoreForAddr(params)
	if err != nil {
		t.Fatal(err)
	}

	r := NewMockRetrieve()
	netStore, err := NewNetStore(localStore, r.retrieve)
	if err != nil {
		t.Fatal(err)
	}

	key := Address{}
	ctx := context.Background()
	// first call is done by the retry on ErrChunkNotFound, no need to do it here
	// _, err = netStore.Get(key)
	// if err == nil || err != ErrChunkNotFound {
	// 	t.Fatalf("expected to get ErrChunkNotFound, but got: %s", err)
	// }

	// second call
	_, err = Get(ctx, netStore, key)
	if got := r.requests[hex.EncodeToString(key)]; got != 2 {
		t.Fatalf("expected to have called retrieve two times, but got: %v", got)
	}
	if err != errUnknown {
		t.Fatalf("expected to get an unknown error, but got: %s", err)
	}

	// third call
	chunk, err := Get(ctx, netStore, key)
	if got := r.requests[hex.EncodeToString(key)]; got != 3 {
		t.Fatalf("expected to have called retrieve three times, but got: %v", got)
	}
	if err != nil || chunk == nil {
		t.Fatalf("expected to get a chunk but got: %v, %s", chunk, err)
	}
	if len(chunk.Data()) != 3 {
		t.Fatalf("expected to get a chunk with size 3, but got: %v", chunk.Data())
	}
}
