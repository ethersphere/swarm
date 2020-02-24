package testutillocal

import (
	"context"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/file"
	"golang.org/x/crypto/sha3"
)

var (
	branches = 128
)

func NewBMTHasherFunc(poolSize int) file.SectionWriterFunc {
	if poolSize == 0 {
		poolSize = bmt.PoolSize
	}
	poolAsync := bmt.NewTreePool(sha3.NewLegacyKeccak256, branches, poolSize)
	refHashFunc := func(_ context.Context) file.SectionWriter {
		return bmt.New(poolAsync).NewAsyncWriter(false)
	}
	return refHashFunc
}
