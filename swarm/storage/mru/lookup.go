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
	"bytes"
	"encoding/binary"
	"hash"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

// LookupParams is used to specify constraints when performing an update lookup
// Limit defines whether or not the lookup should be limited
// If Limit is set to true then Max defines the amount of hops that can be performed
// \TODO this is redundant, just use uint32 with 0 for unlimited hops
type LookupParams struct {
	UpdateLookup
	Limit uint32
}

// RootAddr returns the metadata chunk address
func (r *LookupParams) RootAddr() storage.Address {
	return r.rootAddr
}

func NewLookupParams(rootAddr storage.Address, period, version uint32, limit uint32) *LookupParams {
	return &LookupParams{
		UpdateLookup: UpdateLookup{
			period:   period,
			version:  version,
			rootAddr: rootAddr,
		},
		Limit: limit,
	}
}

// LookupLatest generates lookup parameters that look for the latest version of a resource
func LookupLatest(rootAddr storage.Address) *LookupParams {
	return NewLookupParams(rootAddr, 0, 0, 0)
}

// LookupLatestVersionInPeriod generates lookup parameters that look for the latest version of a resource in a given period
func LookupLatestVersionInPeriod(rootAddr storage.Address, period uint32) *LookupParams {
	return NewLookupParams(rootAddr, period, 0, 0)
}

// LookupVersion generates lookup parameters that look for a specific version of a resource
func LookupVersion(rootAddr storage.Address, period, version uint32) *LookupParams {
	return NewLookupParams(rootAddr, period, version, 0)
}

// UpdateLookup represents the components of a resource update search
type UpdateLookup struct {
	period   uint32
	version  uint32
	rootAddr storage.Address
}

// resourceUpdateChunkAddr calculates the resource update chunk address (formerly known as resourceHash)
func (u *UpdateLookup) GetUpdateAddr() (updateAddr storage.Address) {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(NewResourceHash(u))
	return hasher.Sum(nil)
}

// NewResourceHash will create a deterministic address from the update metadata
// format is: hash(period|version|rootAddr)
func NewResourceHash(u *UpdateLookup) []byte {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, u.period)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, u.version)
	buf.Write(b)
	buf.Write(u.rootAddr[:])
	return buf.Bytes()
}
