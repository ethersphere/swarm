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
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/network"
)

type mockRetrieve struct {
	requests map[string]int
}

func NewMockRetrieve() *mockRetrieve {
	return &mockRetrieve{requests: make(map[string]int)}
}

func (m *mockRetrieve) retrieve(chunk *Chunk) error {
	m.requests[hex.EncodeToString(chunk.Key)] += 1

	return nil
}

func TestNetstoreFailedRequest(t *testing.T) {
	// setup
	addr := network.RandomAddr() // tested peers peer address

	// temp datadir
	datadir, err := ioutil.TempDir("", "netstore")
	if err != nil {
		t.Fatal(err)
	}

	localStore, err := NewTestLocalStoreForAddr(datadir, addr.Over())
	if err != nil {
		t.Fatal(err)
	}

	r := NewMockRetrieve()
	netStore := NewNetStore(localStore, r.retrieve)

	key := Key{}
	_, err = netStore.Get(key)
	if err == nil || err != ErrChunkNotFound {
		t.Fatalf("expected to get ErrChunkNotFound, but got: %s", err)
	}

	_, err = netStore.Get(key)
	if got := r.requests[hex.EncodeToString(key)]; got != 2 {
		t.Fatalf("expected to have called retrieve two times, but got: %v", got)
	}
	if err != nil {
		t.Fatalf("expected to get the chunk on second request, but got: %s", err)
	}
}
