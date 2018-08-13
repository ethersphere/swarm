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
	"fmt"
	"hash"
	"strconv"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// Values interface represents a string key-value store
// useful for building query strings
type Values interface {
	Get(key string) string
	Set(key, value string)
}

// LookupBefore constructs an UpdateLookup structure to find updates on or before `time`
// if time == 0, the latest update will be looked up
func LookupBefore(view *View, time uint64) *UpdateLookup {
	return &UpdateLookup{
		View: *view,
		Epoch: lookup.Epoch{
			Time: time,
		},
	}
}

// LookupLatest generates lookup parameters that look for the latest version of a resource
func LookupLatest(view *View) *UpdateLookup {
	return LookupBefore(view, 0)
}

// UpdateLookup represents the components of a resource update search key.
// it is also used to specify constraints when performing an update lookup
type UpdateLookup struct {
	View
	lookup.Epoch
}

// UpdateLookup layout:
// ResourceIDLength bytes
// userAddr common.AddressLength bytes
// 4 bytes period
// 4 bytes version
const updateLookupLength = viewLength + lookup.EpochLength

// UpdateAddr calculates the resource update chunk address corresponding to this lookup key
func (u *UpdateLookup) UpdateAddr() (updateAddr storage.Address) {
	serializedData := make([]byte, updateLookupLength)
	var cursor int
	u.View.binaryPut(serializedData[cursor : cursor+viewLength])
	cursor += viewLength

	eid := u.Epoch.ID()
	copy(serializedData[cursor:cursor+lookup.EpochLength], eid[:])

	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(serializedData)
	return hasher.Sum(nil)
}

// binaryPut serializes this UpdateLookup instance into the provided slice
func (u *UpdateLookup) binaryPut(serializedData []byte) error {
	if len(serializedData) != updateLookupLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to serialize UpdateLookup. Expected %d, got %d", updateLookupLength, len(serializedData))
	}
	var cursor int
	if err := u.View.binaryPut(serializedData[cursor : cursor+viewLength]); err != nil {
		return err
	}
	cursor += viewLength

	epochBytes, err := u.Epoch.MarshalBinary()
	if err != nil {
		return err
	}
	copy(serializedData[cursor:cursor+lookup.EpochLength], epochBytes[:])
	cursor += lookup.EpochLength

	return nil
}

// binaryLength returns the expected size of this structure when serialized
func (u *UpdateLookup) binaryLength() int {
	return updateLookupLength
}

// binaryGet restores the current instance from the information contained in the passed slice
func (u *UpdateLookup) binaryGet(serializedData []byte) error {
	if len(serializedData) != updateLookupLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to read UpdateLookup. Expected %d, got %d", updateLookupLength, len(serializedData))
	}

	var cursor int
	if err := u.View.binaryGet(serializedData[cursor : cursor+viewLength]); err != nil {
		return err
	}
	cursor += viewLength

	if err := u.Epoch.UnmarshalBinary(serializedData[cursor : cursor+lookup.EpochLength]); err != nil {
		return err
	}
	cursor += lookup.EpochLength

	return nil
}

// FromValues deserializes this instance from a string key-value store
// useful to parse query strings
func (u *UpdateLookup) FromValues(values Values, parseView bool) error {
	level, _ := strconv.ParseUint(values.Get("level"), 10, 32)
	u.Epoch.Level = uint8(level)
	u.Epoch.Time, _ = strconv.ParseUint(values.Get("time"), 10, 64)

	if parseView {
		return u.View.FromValues(values)
	}
	return nil
}

// ToValues serializes this structure into the provided string key-value store
// useful to build query strings
func (u *UpdateLookup) ToValues(values Values) {
	values.Set("level", fmt.Sprintf("%d", u.Epoch.Level))
	values.Set("time", fmt.Sprintf("%d", u.Epoch.Time))
	u.View.ToValues(values)
}
