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

package mru

import (
	"encoding/binary"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Encodes a point in time as a Unix epoch and provides
// space for proof the timestamp was not created prior to its time
type Timestamp struct {
	Time  uint64      // Unix epoch timestamp, in seconds
	Proof common.Hash // space to hold proof of the timestamp, e.g., a block hash
}

// 8 bytes Time
// 32 bytes hash length
const timestampLength = 8 + 32

// timestampProvider interface describes a source of timestamp information
type timestampProvider interface {
	Now() Timestamp // returns the current timestamp information
}

// binaryGet populates the timestamp structure from the given byte slice
func (t *Timestamp) binaryGet(data []byte) error {
	if len(data) != timestampLength {
		return NewError(ErrCorruptData, "timestamp data has the wrong size")
	}
	t.Time = binary.LittleEndian.Uint64(data[:8])
	copy(t.Proof[:], data[8:])
	return nil
}

// binaryPut Serializes a Timestamp to a byte slice
func (t *Timestamp) binaryPut(data []byte) error {
	if len(data) != timestampLength {
		return NewError(ErrCorruptData, "timestamp data has the wrong size")
	}
	binary.LittleEndian.PutUint64(data, t.Time)
	copy(data[8:], t.Proof[:])
	return nil
}

type DefaultTimestampProvider struct {
}

// NewDefaultTimestampProvider creates a system clock based timestamp provider
func NewDefaultTimestampProvider() *DefaultTimestampProvider {
	return &DefaultTimestampProvider{}
}

// Now returns the current time according to this provider
func (dtp *DefaultTimestampProvider) Now() Timestamp {
	return Timestamp{
		Time:  uint64(time.Now().Unix()),
		Proof: common.Hash{},
	}
}
