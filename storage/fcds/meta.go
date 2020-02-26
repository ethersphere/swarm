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

package fcds

import (
	"encoding/binary"
	"fmt"

	"github.com/ethersphere/swarm/chunk"
)

// MetaStore defines methods to store and manage
// chunk meta information in Store FCDS implementation.
type MetaStore interface {
	Get(addr chunk.Address) (*Meta, error)
	Set(addr chunk.Address, shard uint8, reclaimed bool, m *Meta) error
	Remove(addr chunk.Address, shard uint8) error
	Count() (int, error)
	Iterate(func(chunk.Address, *Meta) (stop bool, err error)) error
	FreeOffset(shard uint8) (offset int64, err error)
	NextShard() (shard uint8, hasFree bool)
	Close() error
}

// Meta stores chunk data size and its offset in a file.
type Meta struct {
	Shard  uint8
	Size   uint16
	Offset int64
}

// MarshalBinary returns binary encoded value of meta chunk information.
func (m *Meta) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 12)
	binary.BigEndian.PutUint64(data[:8], uint64(m.Offset))
	binary.BigEndian.PutUint16(data[8:10], m.Size)
	binary.BigEndian.PutUint16(data[10:12], uint16(m.Shard))
	return data, nil
}

// UnmarshalBinary sets meta chunk information from encoded data.
func (m *Meta) UnmarshalBinary(data []byte) error {
	m.Offset = int64(binary.BigEndian.Uint64(data[:8]))
	m.Size = binary.BigEndian.Uint16(data[8:10])
	m.Shard = uint8(binary.BigEndian.Uint16(data[10:12]))
	return nil
}

func (m *Meta) String() (s string) {
	if m == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{Size: %v, Offset %v}", m.Size, m.Offset)
}
