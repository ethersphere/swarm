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

type StreamInfoReq struct {
	Streams []uint
}

type StreamInfoRes struct {
	Streams []StreamDescriptor
}

type StreamDescriptor struct {
	Name    string
	Cursor  uint64
	Bounded bool
}

type GetRange struct {
	Ruid      uint
	Stream    string
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
	Chunks    [][]byte
}

type BatchDone struct {
	Ruid uint
	Last uint
}

type StreamState struct {
	Stream  string
	Code    uint16
	Message string
}
