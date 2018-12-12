package script

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const defaultRetrieveTimeout = 100 * time.Millisecond

type HandlerParams struct {
	ChunkStore storage.ChunkStore
}

type Handler interface {
	storage.ChunkValidator
	Put(ctx context.Context, chunk *Chunk) error
	Get(ctx context.Context, addr storage.Address) (*Chunk, error)
}

type handler struct {
	HandlerParams
}

func NewHandler(params *HandlerParams) Handler {
	return &handler{
		HandlerParams: *params,
	}

}

func (h *handler) Validate(chunkAddr storage.Address, data []byte) bool {

	var r Chunk
	err := r.UnmarshalBinary(data)
	if err != nil {
		// warn
		return false
	}

	if err := r.Verify(chunkAddr); err != nil {
		log.Debug("Invalid script update chunk", "addr", chunkAddr.Hex(), "err", err.Error())
		return false
	}
	return true
}

func (h *handler) Put(ctx context.Context, chunk *Chunk) error {
	return h.ChunkStore.Put(ctx, storage.NewChunk(chunk.Address(), chunk.Data()))
}

func (h *handler) Get(ctx context.Context, addr storage.Address) (*Chunk, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultRetrieveTimeout)
	defer cancel()

	chunk, err := h.ChunkStore.Get(ctx, addr)
	if err != nil {
		return nil, err
	}

	var r Chunk
	return &r, r.UnmarshalBinary(chunk.Data())
}
