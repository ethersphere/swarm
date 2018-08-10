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
	"fmt"
	"hash"
	"net/url"
	"strconv"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// LookupParams is used to specify constraints when performing an update lookup
// Limit defines whether or not the lookup should be limited
// If Limit is set to true then Max defines the amount of hops that can be performed
type LookupParams struct {
	UpdateLookup        // last known epoch/level. 0 means no guessing
	Time         uint64 // Find updates with a timestamp <= Time. 0 means now (find latest)
}

// Values interface represents a string key-value store
// useful for building query strings
type Values interface {
	Get(key string) string
	Set(key, value string)
}

// FromValues deserializes this instance from a string key-value store
// useful to parse query strings
func (lp *LookupParams) FromValues(values Values, parseView bool) error {
	lp.Time, _ = strconv.ParseUint(values.Get("time"), 10, 64)
	return lp.UpdateLookup.FromValues(values, parseView)
}

// ToValues serializes this structure into the provided string key-value store
// useful to build query strings
func (lp *LookupParams) ToValues(values url.Values) {
	values.Set("time", fmt.Sprintf("%d", lp.Time))
	lp.UpdateLookup.ToValues(values)
}

// NewLookupParams constructs a LookupParams structure with the provided lookup parameters
func NewLookupParams(view *View, time uint64) *LookupParams {
	return &LookupParams{
		UpdateLookup: UpdateLookup{
			View: *view,
		},
		Time: time,
	}
}

// LookupLatest generates lookup parameters that look for the latest version of a resource
func LookupLatest(view *View) *LookupParams {
	return NewLookupParams(view, 0)
}

// UpdateLookup represents the components of a resource update search key.Version
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
	u.binaryPut(serializedData)
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

	serializedData[cursor] = u.Epoch.Level
	cursor++

	binary.LittleEndian.PutUint64(serializedData[cursor:cursor+8], u.Epoch.BaseTime)
	cursor += 8

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

	u.Level = serializedData[cursor]
	cursor++

	u.BaseTime = binary.LittleEndian.Uint64(serializedData[cursor : cursor+8])
	cursor += 8

	return nil
}

// FromValues deserializes this instance from a string key-value store
// useful to parse query strings
func (u *UpdateLookup) FromValues(values Values, parseView bool) error {
	level, _ := strconv.ParseUint(values.Get("level"), 10, 32)
	u.Level = uint8(level)
	u.BaseTime, _ = strconv.ParseUint(values.Get("basetime"), 10, 64)

	if parseView {
		return u.View.FromValues(values)
	}
	return nil
}

// ToValues serializes this structure into the provided string key-value store
// useful to build query strings
func (u *UpdateLookup) ToValues(values Values) {
	values.Set("level", fmt.Sprintf("%d", u.Level))
	values.Set("basetime", fmt.Sprintf("%d", u.Epoch))
	u.View.ToValues(values)
}
