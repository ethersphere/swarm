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
	"github.com/ethereum/go-ethereum/swarm/chunk"
)

// resourceUpdate encapsulates the information sent as part of a resource update
type resourceUpdate struct {
	updateHeader        // metainformationa about this resource update
	data         []byte // actual data payload
}

// Header: (see updateHeader)
// Data:
// data (datalength bytes)
//
// Minimum size is Header + 1 (minimum data length, enforced)
const minimumUpdateDataLength = updateHeaderLength + 1
const maxUpdateDataLength = chunk.DefaultSize - signatureLength - updateHeaderLength

// binaryPut serializes the resource update information into the given slice
func (r *resourceUpdate) binaryPut(serializedData []byte) error {
	datalength := len(r.data)
	if datalength == 0 {
		return NewError(ErrInvalidValue, "cannot update a resource with no data")
	}

	if datalength > maxUpdateDataLength {
		return NewErrorf(ErrInvalidValue, "data is too big (length=%d). Max length=%d", datalength, maxUpdateDataLength)
	}

	if len(serializedData) != r.binaryLength() {
		return NewErrorf(ErrInvalidValue, "slice passed to putBinary must be of exact size. Expected %d bytes", r.binaryLength())
	}

	var cursor int
	// serialize header (see updateHeader)
	if err := r.updateHeader.binaryPut(serializedData[cursor : cursor+updateHeaderLength]); err != nil {
		return err
	}
	cursor += updateHeaderLength

	// add the data
	copy(serializedData[cursor:], r.data)
	cursor += datalength

	return nil
}

// binaryLength returns the expected number of bytes this structure will take to encode
func (r *resourceUpdate) binaryLength() int {
	return updateHeaderLength + len(r.data)
}

// binaryGet populates this instance from the information contained in the passed byte slice
func (r *resourceUpdate) binaryGet(serializedData []byte) error {
	if len(serializedData) < minimumUpdateDataLength {
		return NewErrorf(ErrNothingToReturn, "chunk less than %d bytes cannot be a resource update chunk", minimumUpdateDataLength)
	}
	dataLength := len(serializedData) - updateHeaderLength
	var cursor int
	// at this point we can be satisfied that we have the correct data length to read
	if err := r.updateHeader.binaryGet(serializedData[cursor : cursor+updateHeaderLength]); err != nil {
		return err
	}
	cursor += updateHeaderLength

	data := serializedData[cursor : cursor+dataLength]
	cursor += dataLength

	// now that all checks have passed, copy data into structure
	r.data = make([]byte, dataLength)
	copy(r.data, data)

	return nil

}
