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
	"fmt"

	"github.com/ethersphere/swarm/storage"
)

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

// Stream defines a unique stream identifier.
type ID struct {
	// Name is used for Client and Server functions identification.
	Name string
	// Key is the name of specific stream data.
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
