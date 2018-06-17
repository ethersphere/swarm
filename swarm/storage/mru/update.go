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
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
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

const updateLookupLength = 4 + 4 + storage.KeyLength
const updateHeaderLength = updateLookupLength + 1 + storage.KeyLength

// resourceUpdate encapsulates the information sent as part of a resource update
type resourceUpdate struct {
	updateHeader
	data []byte
}

// Update chunk layout
// Header:
// 2 bytes header Length
// 2 bytes data length
// 4 bytes period
// 4 bytes version
// 32 bytes rootAddr reference
// 32 bytes metaHash digest
// 32 bytes previousUpdateAddr reference
// 40 bytes Timestamp
// Data:
// data (datalength bytes)
//
// Minimum size is Header + 1 (minimum data length, enforced)
const minimumUpdateDataLength = updateHeaderLength + 1
const maxUpdateDataLength = chunkSize - signatureLength - updateHeaderLength

func (r *resourceUpdate) binaryPut(serializedData []byte) error {
	if len(r.rootAddr) != storage.KeyLength || len(r.metaHash) != storage.KeyLength {
		log.Warn("Call to newUpdateChunk with incorrect rootAddr or metaHash")
		return NewError(ErrInvalidValue, "newUpdateChunk called without rootAddr or metaHash set")
	}

	datalength := len(r.data)
	if datalength == 0 {
		return NewError(ErrInvalidValue, "cannot update a resource with no data")
	}

	if datalength > maxUpdateDataLength {
		return NewErrorf(ErrInvalidValue, "data is too big (length=%d). Max length=%d", datalength, maxUpdateDataLength)
	}

	if len(serializedData) != r.binaryLength() {
		return NewError(ErrInvalidValue, "slice passed to putBinary must be of exact size")
	}
	// data header length does NOT include the header length prefix bytes themselves
	cursor := 0
	binary.LittleEndian.PutUint16(serializedData[cursor:], uint16(updateHeaderLength))
	cursor += 2

	// data length
	binary.LittleEndian.PutUint16(serializedData[cursor:], uint16(datalength))
	cursor += 2

	// header = period + version + multihash flag + rootAddr + metaHash
	binary.LittleEndian.PutUint32(serializedData[cursor:], r.period)
	cursor += 4

	binary.LittleEndian.PutUint32(serializedData[cursor:], r.version)
	cursor += 4

	copy(serializedData[cursor:], r.rootAddr[:storage.KeyLength])
	cursor += storage.KeyLength
	copy(serializedData[cursor:], r.metaHash[:storage.KeyLength])
	cursor += storage.KeyLength

	if r.multihash {
		if isMultihash(r.data) == 0 {
			return NewError(ErrInvalidValue, "Invalid multihash")
		}
		serializedData[cursor] = 0x01
	}
	cursor++

	// add the data
	copy(serializedData[cursor:], r.data)
	cursor += datalength

	return nil
}

func (r *resourceUpdate) binaryLength() int {
	return 2 + 2 + updateHeaderLength + len(r.data) // initial 4 are uint16 length descriptors for updateHeaderLength and datalength
}

func (r *resourceUpdate) binaryGet(serializedData []byte) error {
	if len(serializedData) < minimumUpdateDataLength {
		return NewError(ErrNothingToReturn, fmt.Sprintf("chunk less than %d bytes cannot be a resource update chunk", minimumUpdateDataLength))
	}
	cursor := 0
	declaredHeaderlength := binary.LittleEndian.Uint16(serializedData[cursor : cursor+2])
	if declaredHeaderlength != updateHeaderLength {
		return NewErrorf(ErrCorruptData, "Invalid header length. Expected %d, got %d", updateHeaderLength, declaredHeaderlength)
	}

	cursor += 2
	datalength := int(binary.LittleEndian.Uint16(serializedData[cursor : cursor+2]))
	cursor += 2

	if int(2+2+updateHeaderLength+datalength+signatureLength) != len(serializedData) {
		return NewError(ErrNothingToReturn, "length specified in header is different than actual chunk size")
	}

	// at this point we can be satisfied that the data integrity is ok
	var header updateHeader

	header.period = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4
	header.version = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4

	header.rootAddr = storage.Address(make([]byte, storage.KeyLength))
	header.metaHash = make([]byte, storage.KeyLength)
	copy(header.rootAddr, serializedData[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength
	copy(header.metaHash, serializedData[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength

	if serializedData[cursor] == 0x01 {
		header.multihash = true
	}
	cursor++

	// if multihash content is indicated we check the validity of the multihash
	if header.multihash {
		mhLength, mhHeaderLength, err := multihash.GetMultihashLength(serializedData[cursor:])
		if err != nil {
			log.Error("multihash parse error", "err", err)
			return err
		}
		if datalength != mhLength+mhHeaderLength {
			log.Debug("multihash error", "datalength", datalength, "mhLength", mhLength, "mhHeaderLength", mhHeaderLength)
			return errors.New("Corrupt multihash data")
		}
	}
	r.updateHeader = header
	r.data = make([]byte, datalength)
	copy(r.data, serializedData[cursor:cursor+datalength])
	cursor += datalength

	return nil

}
