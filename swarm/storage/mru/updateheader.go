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

// updateHeader models the non-payload components of a Resource Update
// Extensible placeholder. Right now it contains no additional components.
type updateHeader struct {
	UpdateLookup // UpdateLookup contains the information required to locate this resource (components of the search key used to find it)
}

// updateHeader layout
// updateLookupLength bytes
const updateHeaderLength = updateLookupLength

// binaryPut serializes the resource header information into the given slice
func (h *updateHeader) binaryPut(serializedData []byte) error {
	if len(serializedData) != updateHeaderLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to serialize updateHeaderLength. Expected %d, got %d", updateHeaderLength, len(serializedData))
	}
	if err := h.UpdateLookup.binaryPut(serializedData[:updateLookupLength]); err != nil {
		return err
	}
	return nil
}

// binaryLength returns the expected size of this structure when serialized
func (h *updateHeader) binaryLength() int {
	return updateHeaderLength
}

// binaryGet restores the current updateHeader instance from the information contained in the passed slice
func (h *updateHeader) binaryGet(serializedData []byte) error {
	if len(serializedData) != updateHeaderLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to read updateHeaderLength. Expected %d, got %d", updateHeaderLength, len(serializedData))
	}

	if err := h.UpdateLookup.binaryGet(serializedData[:updateLookupLength]); err != nil {
		return err
	}
	return nil
}
