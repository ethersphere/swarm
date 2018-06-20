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
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type updateHeader struct {
	UpdateLookup
	multihash bool
	metaHash  []byte

	time               Timestamp
	previousUpdateAddr storage.Address
	nextUpdateTime     uint64
}

const updateHeaderLength = updateLookupLength + 1 + storage.KeyLength

// 2 bytes header Length
// 2 bytes data length
// 4 bytes period
// 4 bytes version
// 32 bytes rootAddr reference
// 32 bytes metaHash digest
// 32 bytes previousUpdateAddr reference
// 40 bytes Timestamp

func (h *updateHeader) binaryPut(serializedData []byte) error {
	if len(serializedData) != updateHeaderLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to serialize updateHeaderLength. Expected %d, got %d", updateHeaderLength, len(serializedData))
	}
	if len(h.metaHash) != storage.KeyLength {
		log.Warn("Call to updateHeader.binaryPut with incorrect metaHash")
		return NewError(ErrInvalidValue, "updateHeader.binaryPut called without metaHash set")
	}
	if err := h.UpdateLookup.binaryPut(serializedData[:updateLookupLength]); err != nil {
		return err
	}
	cursor := updateLookupLength
	copy(serializedData[cursor:], h.metaHash[:storage.KeyLength])
	cursor += storage.KeyLength

	if h.multihash {
		serializedData[cursor] = 0x01
	}
	cursor++

	return nil
}

func (h *updateHeader) binaryLength() int {
	return updateLookupLength
}

func (h *updateHeader) binaryGet(serializedData []byte) error {
	if len(serializedData) != updateHeaderLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to read updateHeaderLength. Expected %d, got %d", updateHeaderLength, len(serializedData))
	}

	if err := h.UpdateLookup.binaryGet(serializedData[:updateLookupLength]); err != nil {
		return err
	}
	cursor := updateLookupLength
	h.metaHash = make([]byte, storage.KeyLength)
	copy(h.metaHash[:storage.KeyLength], serializedData[cursor:])
	cursor += storage.KeyLength

	if serializedData[cursor] == 0x01 {
		h.multihash = true
	}
	cursor++

	return nil
}
