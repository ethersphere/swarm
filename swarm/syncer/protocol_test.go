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

package syncer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// TestDispatcherSendChunk tests if the dispatcher.sendChunk is called
// a chunk message appears on the chunkC channel
func TestDispatcherSendChunk(t *testing.T) {
	baseAddr := network.RandomAddr().OAddr
	s := newDispatcher(baseAddr)
	timeout := time.NewTimer(100 * time.Millisecond)
	var chmsg *chunkMsg
	called := make(chan *chunkMsg)
	s.sendChunkMsg = func(msg *chunkMsg) error {
		called <- msg
		return nil
	}

	chunk := storage.GenerateRandomChunk(100)
	addr := chunk.Address()
	chunkData := chunk.Data()
	go s.sendChunk(chunk)
	select {
	case chmsg = <-called:
	case <-timeout.C:
		t.Fatal("timeout waiting for chunk message on channel")
	}
	if !bytes.Equal(chmsg.Addr, addr) {
		t.Fatalf("expected chunk message address %v, got %v", addr, chmsg.Addr)
	}
	if !bytes.Equal(chmsg.Origin, baseAddr) {
		t.Fatalf("expected Origin address %v, got %v", baseAddr, chmsg.Origin)
	}
	if !bytes.Equal(chmsg.Data, chunkData) {
		t.Fatalf("expected chunk message data %v, got %v", chunkData, chmsg.Data)
	}
	if len(chmsg.Nonce) != 32 {
		t.Fatalf("expected nonce to be 32 bytes long, got %v", len(chmsg.Nonce))
	}
}

// TestDispatcherHandleReceipt tests that if handleReceipt is called with a receipt message
// then processReceipt is called with the address
func TestDispatcherHandleProof(t *testing.T) {
	baseAddr := network.RandomAddr().OAddr
	s := newDispatcher(baseAddr)
	timeout := time.NewTimer(100 * time.Millisecond)
	called := make(chan storage.Address)
	s.processReceipt = func(a storage.Address) error {
		called <- a
		return nil
	}

	chunk := storage.GenerateRandomChunk(100)
	addr := chunk.Address()
	nonce := newNonce()
	msg := &receiptMsg{addr, nonce}
	peer := p2p.NewPeer(enode.ID{}, "", nil)
	var next []byte
	go s.handleReceipt(msg, peer)
	select {
	case next = <-called:
	case <-timeout.C:
		t.Fatal("timeout waiting for receipt address on channel")
	}
	if !bytes.Equal(next, addr) {
		t.Fatalf("expected receipt address %v, got %v", addr, next)
	}
}

// TestStorerHandleChunk that if storer.handleChunk is called then the
// chunk gets stored and receipt is created that sendReceiptMsg is called with
func TestStorerHandleChunk(t *testing.T) {
	// set up storer
	origin := network.RandomAddr().OAddr
	chunkStore := storage.NewMapChunkStore()
	s := newStorer(chunkStore)
	timeout := time.NewTimer(100 * time.Millisecond)
	called := make(chan *receiptMsg)
	var destination []byte
	s.sendReceiptMsg = func(to []byte, msg *receiptMsg) error {
		called <- msg
		destination = to
		return nil
	}
	// create a chunk message and call handleChunk on it
	chunk := storage.GenerateRandomChunk(100)
	addr := chunk.Address()
	data := chunk.Data()
	peer := p2p.NewPeer(enode.ID{}, "", nil)
	nonce := newNonce()
	chmsg := &chunkMsg{
		Origin: origin,
		Addr:   addr,
		Data:   data,
		Nonce:  nonce,
	}
	go s.handleChunk(chmsg, peer)

	var r *receiptMsg
	select {
	case r = <-called:
	case <-timeout.C:
		t.Fatal("timeout waiting for chunk message on channel")
	}
	if _, err := chunkStore.Get(context.TODO(), addr); err != nil {
		t.Fatalf("expected chunk with address %v to be stored in chunkStore", addr)
	}
	if !bytes.Equal(destination, origin) {
		t.Fatalf("expected destination to equal origin %v, got %v", origin, destination)
	}
	if !bytes.Equal(r.Addr, addr) {
		t.Fatalf("expected receipt msg address to be chunk address %v, got %v", addr, r.Addr)
	}
}
