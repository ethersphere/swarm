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

package swarmhash

import (
	"crypto/sha256"
	"hash"

	"github.com/ethereum/go-ethereum/common"
)

var (
	data               = []byte("foo")
	dataHash           = common.FromHex("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	dataHashWithLength = common.FromHex("6139f36b7e4e7dadbd1391967339c7673629e4750c02b0545f8dbd6090cbff1e")
)

func init() {
	Add("bar", 16, func() Hash {
		return newTestHasher()
	})
}

func initTest() {
	Init("bar")
	multihashTypeCode = defaultMultihashTypeCode
}

type testHasher struct {
	hash.Hash
}

func newTestHasher() *testHasher {
	return &testHasher{
		Hash: sha256.New(),
	}
}

func (t *testHasher) ResetWithLength(length []byte) {
	t.Reset()
	t.Write(length)
}
