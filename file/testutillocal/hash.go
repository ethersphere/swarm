package testutillocal

import (
	"context"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/file"
	asyncbmt "github.com/ethersphere/swarm/file/bmt"
	"golang.org/x/crypto/sha3"
)

var (
	branches = 128
)

// NewBMTHasherFunc is a test helper that creates a new asynchronous hasher with a specified poolsize
func NewBMTHasherFunc(poolSize int) file.SectionWriterFunc {
	if poolSize == 0 {
		poolSize = bmt.PoolSize
	}
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, poolSize)
	refHashFunc := func(ctx context.Context) file.SectionWriter {
		bmtHasher := bmt.New(poolAsync)
		return asyncbmt.NewAsyncHasher(ctx, bmtHasher, false, nil)
	}
	return refHashFunc
}
