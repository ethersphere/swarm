package script

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const defaultRetrieveTimeout = 100 * time.Millisecond

//HandlerParams contains the Handler's initialization parameters
type HandlerParams struct {
	ChunkStore storage.ChunkStore
}

// Handler defines a common interface for the script Handler
// It behaves as a chunk store for self-validating chunks
type Handler interface {
	storage.ChunkValidator
	Put(ctx context.Context, chunk *Chunk) error
	Get(ctx context.Context, addr storage.Address) (*Chunk, error)
}

type handler struct {
	HandlerParams
}

// NewHandler builds a new Handler
func NewHandler(params *HandlerParams) Handler {
	return &handler{
		HandlerParams: *params,
	}

}

// Validate implements the storage.ChunkValidator interface
func (h *handler) Validate(chunk storage.Chunk) bool {

	var r Chunk
	err := r.UnmarshalBinary(chunk.Data())
	if err != nil {
		return false
	}

	if err := r.Verify(chunk.Address()); err != nil {
		log.Debug("Invalid script update chunk", "addr", chunk.Address(), "err", err)
		return false
	}
	return true
}

// Put writes a self-validating chunk to the underlying chunkstore. Internally, the
// underlying chunkstore will call Validate() which will verify the chunk is valid.
func (h *handler) Put(ctx context.Context, chunk *Chunk) error {
	return h.ChunkStore.Put(ctx, storage.NewChunk(chunk.Address(), chunk.Data()))
}

// Get will retrieve a chunk by address.
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
