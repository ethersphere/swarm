package testutillocal

import (
	"context"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/param"
	"golang.org/x/crypto/sha3"
)

var (
	branches = 128
)

func NewBMTHasherFunc(poolSize int) param.SectionWriterFunc {
	if poolSize == 0 {
		poolSize = bmt.PoolSize
	}
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, poolSize)
	refHashFunc := func(_ context.Context) param.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	return refHashFunc
}
