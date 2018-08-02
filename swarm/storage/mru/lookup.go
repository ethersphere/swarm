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

// RootAddr returns the metadata chunk address
func (lp *LookupParams) ViewID() *ResourceViewID {
	return &lp.viewID
}

func (lp *LookupParams) FromURL(url *url.URL, parseView bool) error {
	query := url.Query()
	version, _ := strconv.ParseUint(query.Get("version"), 10, 32)
	period, _ := strconv.ParseUint(query.Get("period"), 10, 32)
	limit, _ := strconv.ParseUint(query.Get("limit"), 10, 32)

	if period == 0 && version != 0 {
		return NewError(ErrInvalidValue, "cannot have version !=0 if period is 0")
	}
	lp.version = uint32(version)
	lp.period = uint32(period)
	lp.Limit = uint32(limit)
	if parseView {
		return lp.viewID.FromURL(url)
	}
	return nil
}

func (lp *LookupParams) ToURL(url *url.URL) {
	query := url.Query()
	if lp.period != 0 {
		query.Set("period", fmt.Sprintf("%d", lp.period))
	}
	if lp.version != 0 {
		query.Set("version", fmt.Sprintf("%d", lp.version))
	}
	if lp.Limit != 0 {
		query.Set("limit", fmt.Sprintf("%d", lp.version))
	}
	url.RawQuery = query.Encode()
	lp.viewID.ToURL(url)
}

func NewLookupParams(viewID *ResourceViewID, period, version uint32, limit uint32) *LookupParams {
	return &LookupParams{
		UpdateLookup: UpdateLookup{
			period:  period,
			version: version,
			viewID:  *viewID,
		},
		Limit: limit,
	}
}

// LookupLatest generates lookup parameters that look for the latest version of a resource
func LookupLatest(viewID *ResourceViewID) *LookupParams {
	return NewLookupParams(viewID, 0, 0, 0)
}

// LookupLatestVersionInPeriod generates lookup parameters that look for the latest version of a resource in a given period
func LookupLatestVersionInPeriod(viewID *ResourceViewID, period uint32) *LookupParams {
	return NewLookupParams(viewID, period, 0, 0)
}

// LookupVersion generates lookup parameters that look for a specific version of a resource
func LookupVersion(viewID *ResourceViewID, period, version uint32) *LookupParams {
	return NewLookupParams(viewID, period, version, 0)
}

// UpdateLookup represents the components of a resource update search key
type UpdateLookup struct {
	viewID  ResourceViewID
	period  uint32
	version uint32
}

// UpdateLookup layout:
// ResourceIDLength bytes
// ownerAddr common.AddressLength bytes
// 4 bytes period
// 4 bytes version
const updateLookupLength = resourceViewIDLength + 4 + 4

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
	if err := u.viewID.binaryPut(serializedData[cursor : cursor+resourceViewIDLength]); err != nil {
		return err
	}
	cursor += resourceViewIDLength

	binary.LittleEndian.PutUint32(serializedData[cursor:cursor+4], u.period)
	cursor += 4

	binary.LittleEndian.PutUint32(serializedData[cursor:cursor+4], u.version)
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
	if err := u.viewID.binaryGet(serializedData[cursor : cursor+resourceViewIDLength]); err != nil {
		return err
	}
	cursor += resourceViewIDLength

	u.period = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4

	u.version = binary.LittleEndian.Uint32(serializedData[cursor : cursor+4])
	cursor += 4

	return nil
}
