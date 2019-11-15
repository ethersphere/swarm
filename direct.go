package swarm

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

type directAccessAPI struct {
	chunkStore storage.ChunkStore
}

func NewDirectAccessAPI(chunkStore storage.ChunkStore) *directAccessAPI {
	return &directAccessAPI{
		chunkStore: chunkStore,
	}
}

func (d *directAccessAPI) GetByReference(addr storage.Address) (hexutil.Bytes, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chunk, err := d.chunkStore.Get(ctx, chunk.ModeGetRequest, addr)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(chunk.Data()), nil
}
