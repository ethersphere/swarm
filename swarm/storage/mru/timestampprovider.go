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

const timestampLength = 8 + 32

type Timestamp struct {
	Time  uint64
	Proof common.Hash
}

type timestampProvider interface {
	GetCurrentTimestamp() Timestamp
}

type DefaultTimestampProvider struct {
}

func NewDefaultTimestampProvider() *DefaultTimestampProvider {
	return &DefaultTimestampProvider{}
}

func (dtp *DefaultTimestampProvider) GetCurrentTimestamp() Timestamp {
	return Timestamp{
		Time:  uint64(time.Now().Unix()),
		Proof: common.Hash{},
	}
}

func (t *Timestamp) unmarshalBinary(data []byte) error {
	if len(data) != timestampLength {
		return NewError(ErrCorruptData, "timestamp data has the wrong size")
	}
	t.Time = binary.LittleEndian.Uint64(data[:8])
	copy(t.Proof[:], data[8:])
	return nil
}

func (t *Timestamp) marshalBinary() (data []byte) {
	data = make([]byte, timestampLength)
	binary.LittleEndian.PutUint64(data, t.Time)
	copy(data[8:], t.Proof[:])
	return data
}
