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
