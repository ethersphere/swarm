// Copyright 2019 The Swarm Authors
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

package syncer

import (
	"context"
	"fmt"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

// StreamProvider interface provides a lightweight abstraction that allows an easily-pluggable
// stream provider as part of the Stream! protocol specification.
// Since Stream! thoroughly defines the concepts of a stream, intervals, clients and servers, the
// interface therefore needs only a pluggable provider.
// The domain interpretable notions which are at the discretion of the implementing
// provider therefore are - sourcing data (get, put, subscribe for constant new data, and need data
// which is to decide whether to retrieve data or not), retrieving cursors from the data store, the
// implementation of which streams to maintain with a certain peer and providing functionality
// to expose, parse and encode values related to the string represntation of the stream
type StreamProvider interface {

	// NeedData informs the caller whether a certain chunk needs to be fetched from another peer or not.
	// Typically this will involve checking whether a certain chunk exists locally.
	// In case a chunk does not exist locally - a `wait` function returns upon chunk delivery
	NeedData(ctx context.Context, ctx chunk.Address) (bool, wait func(context.Context) error)

	// Get a particular chunk identified by addr from the local storage
	Get(ctx context.Context, addr chunk.Address) ([]byte, error)

	// Put a certain chunk into the local storage
	Put(ctx context.Context, addr chunk.Address, data []byte) error

	// Subscribe to a data stream from an arbitrary data source
	Subscribe(key interface{}, from, to uint64) (<-chan chunk.Descriptor, func())

	// Cursor returns the last known Cursor for a given Stream Key
	Cursor(interface{}) (uint64, error)

	// RunUpdateStreams is a provider specific implementation on how to maintain running streams with
	// an arbitrary Peer. This method should always be run in a separate goroutine
	RunUpdateStreams(p *Peer)

	// StreamName returns the Name of the Stream (see ID)
	StreamName() string

	// ParseStream from a standard pipe-separated string and return the Stream Key
	ParseKey(string) interface{}

	// EncodeStream from a Stream Key to a Stream pipe-separated string representation
	EncodeKey(interface{}) string

	IntervalKey(ID) string

	Boundedness() bool
}

type StreamInfoReq struct {
	Streams []ID
}

type StreamInfoRes struct {
	Streams []StreamDescriptor
}

type StreamDescriptor struct {
	Stream  ID
	Cursor  uint64
	Bounded bool
}

type GetRange struct {
	Ruid      uint
	Stream    ID
	From      uint64
	To        uint64 `rlp:nil`
	BatchSize uint
	Roundtrip bool
}

type OfferedHashes struct {
	Ruid      uint
	LastIndex uint
	Hashes    []byte
}

type WantedHashes struct {
	Ruid      uint
	BitVector []byte
}

type ChunkDelivery struct {
	Ruid      uint
	LastIndex uint
	Chunks    []DeliveredChunk
}

type DeliveredChunk struct {
	Addr storage.Address //chunk address
	Data []byte          //chunk data
}

type BatchDone struct {
	Ruid uint
	Last uint
}

type StreamState struct {
	Stream  ID
	Code    uint16
	Message string
}

// Stream defines a unique stream identifier in a textual representation
type ID struct {
	// Name is used for the Stream provider identification
	Name string
	// Key is the name of specific data stream within the stream provider. The semantics of this value
	// is at the discretion of the stream provider implementation
	Key string
}

func NewID(name string, key string) ID {
	return Stream{
		Name: name,
		Key:  key,
	}
}

// String return a stream id based on all Stream fields.
func (s ID) String() string {
	return fmt.Sprintf("%s|%s", s.Name, s.Key)
}
