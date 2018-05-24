package hash

import (
	"crypto/sha256"
	"hash"

	"github.com/ethereum/go-ethereum/common"
)

var (
	data     = []byte("foo")
	dataHash = common.FromHex("c5aac592460a9ac7845e341090f6f9c81f201b63e5338ee8948a6fe6830c55dc")
)

func init() {
	AddHasher("bar", 16, makeHashFunc)
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

func makeHashFunc(typ string) SwarmHasher {
	return func() SwarmHash {
		return newTestHasher()
	}
}
