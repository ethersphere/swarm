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
	"context"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// dispatcher makes sure newly stored chunks make it to the neighbourhood where they
// can be retrieved, i.e. to nodes whose area of responsibility includes the chunks address
// it gathers proof of custody responses validates them and signal the chunk is synced
type dispatcher struct {
	baseAddr       storage.Address             // base address to use in proximity calculation
	sendChunkMsg   func(*chunkMsg) error       // function to send chunk msg
	processReceipt func(storage.Address) error // function to process receipt for a chunk
}

// newDispatcher constructs a new node-wise dispatcher
func newDispatcher(baseAddr storage.Address) *dispatcher {
	return &dispatcher{
		baseAddr: baseAddr,
	}
}

// chunkMsg is the message construct to send chunks to their local neighbourhood
type chunkMsg struct {
	Addr   []byte
	Data   []byte
	Origin []byte
	Nonce  []byte
}

// sendChunk is called on incoming chunks that are to be synced
func (s *dispatcher) sendChunk(ch storage.Chunk) error {
	nonce := newNonce()
	// TODO: proofs for the nonce should be generated and saved
	msg := &chunkMsg{
		Origin: s.baseAddr,
		Addr:   ch.Address()[:],
		Data:   ch.Data(),
		Nonce:  nonce,
	}
	return s.sendChunkMsg(msg)
}

// newNonce creates a random nonce;
// even without POC it is important otherwise resending a chunk is deduplicated by pss
func newNonce() []byte {
	buf := make([]byte, 32)
	t := 0
	for t < len(buf) {
		n, _ := rand.Read(buf[t:])
		t += n
	}
	return buf
}

// receiptMsg is a statement of custody response to a nonce on a chunk
// it is currently a notification only, contains no proof
type receiptMsg struct {
	Addr  []byte
	Nonce []byte
}

// handleReceipt is called by the pss dispatcher on proofTopic msgs
// after processing the receipt, it calls the chunk address to receiptsC
func (s *dispatcher) handleReceipt(msg *receiptMsg, p *p2p.Peer) error {
	return s.processReceipt(msg.Addr)
}

// storer makes sure that chunks sent to them that fall within their area of responsibility
// are stored and synced to their nearest neighbours and issue a receipt as a response
// to the originator
type storer struct {
	chunkStore     storage.ChunkStore // store to put chunks in, and retrieve them
	sendReceiptMsg func(to []byte, r *receiptMsg) error
}

// newStorer constructs a new node-wise storer
func newStorer(chunkStore storage.ChunkStore) *storer {
	s := &storer{
		chunkStore: chunkStore,
	}
	return s
}

// handleChunk is called by the pss dispatcher on chunkTopic msgs
// only if the chunk falls in the nodes area of responsibility
func (s *storer) handleChunk(msg *chunkMsg, p *p2p.Peer) error {
	// TODO: double check if it falls in area of responsibility
	ch := storage.NewChunk(msg.Addr, msg.Data)
	err := s.chunkStore.Put(context.TODO(), ch)
	if err != nil {
		return err
	}
	// TODO: check if originator or relayer is a nearest neighbour then return
	// otherwise send back receipt
	r := &receiptMsg{
		Addr:  msg.Addr,
		Nonce: msg.Nonce,
	}
	return s.sendReceiptMsg(msg.Origin, r)
}
