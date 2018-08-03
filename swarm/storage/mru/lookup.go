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

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// LookupParams is used to specify constraints when performing an update lookup
// Limit defines whether or not the lookup should be limited
// If Limit is set to true then Max defines the amount of hops that can be performed
type LookupParams struct {
	UpdateLookup
	Limit uint32
}

type Values interface {
	Get(key string) string
	Set(key, value string)
}

func (lp *LookupParams) FromValues(values Values, parseView bool) error {
	limit, _ := strconv.ParseUint(values.Get("limit"), 10, 32)

	lp.Limit = uint32(limit)
	return lp.UpdateLookup.FromValues(values, parseView)
}

func (lp *LookupParams) ToValues(values url.Values) {
	if lp.Limit != 0 {
		values.Set("limit", fmt.Sprintf("%d", lp.Version))
	}
	lp.UpdateLookup.ToValues(values)
}

func NewLookupParams(view *View, period, version uint32, limit uint32) *LookupParams {
	return &LookupParams{
		UpdateLookup: UpdateLookup{
			Period:  period,
			Version: version,
			View:    *view,
		},
		Limit: limit,
	}
}

// LookupLatest generates lookup parameters that look for the latest version of a resource
func LookupLatest(view *View) *LookupParams {
	return NewLookupParams(view, 0, 0, 0)
}

// LookupLatestVersionInPeriod generates lookup parameters that look for the latest version of a resource in a given period
func LookupLatestVersionInPeriod(view *View, period uint32) *LookupParams {
	return NewLookupParams(view, period, 0, 0)
}

// LookupVersion generates lookup parameters that look for a specific version of a resource
func LookupVersion(view *View, period, version uint32) *LookupParams {
	return NewLookupParams(view, period, version, 0)
}

// UpdateLookup represents the components of a resource update search key
type UpdateLookup struct {
	View
	Period  uint32
	Version uint32
}

// UpdateLookup layout:
// ResourceIDLength bytes
// userAddr common.AddressLength bytes
// 4 bytes period
// 4 bytes version
const updateLookupLength = viewLength + 4 + 4

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

	binary.LittleEndian.PutUint32(serializedData[cursor:cursor+4], u.Period)
	cursor += 4

	binary.LittleEndian.PutUint32(serializedData[cursor:cursor+4], u.Version)
	cursor += 4

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

	u.Period = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4

	u.Version = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4

	return nil
}

func (u *UpdateLookup) FromValues(values Values, parseView bool) error {
	version, _ := strconv.ParseUint(values.Get("version"), 10, 32)
	period, _ := strconv.ParseUint(values.Get("period"), 10, 32)

	if period == 0 && version != 0 {
		return NewError(ErrInvalidValue, "cannot have version !=0 if period is 0")
	}
	u.Version = uint32(version)
	u.Period = uint32(period)
	if parseView {
		return u.View.FromValues(values)
	}
	return nil
}

func (u *UpdateLookup) ToValues(values Values) {
	if u.Period != 0 {
		values.Set("period", fmt.Sprintf("%d", u.Period))
	}
	if u.Version != 0 {
		values.Set("version", fmt.Sprintf("%d", u.Version))
	}
	u.View.ToValues(values)
}
