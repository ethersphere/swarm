package store

import (
	"context"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
)

// FileStore implements param.SectionWriter
// It intercepts data between source and hasher
// and compiles the data with the received hash on sum
// to a chunk to be passed to underlying chunk.Store.Put
type FileStore struct {
	chunkStore chunk.Store
	w          param.SectionWriter
	ctx        context.Context
	data       [][]byte
	errFunc    func(error)
}

// New creates a new FileStore with the supplied chunk.Store
func New(chunkStore chunk.Store) *FileStore {
	return &FileStore{
		chunkStore: chunkStore,
	}
}

// Init implements param.SectionWriter
func (f *FileStore) Init(ctx context.Context, errFunc func(error)) {
	f.ctx = ctx
	f.errFunc = errFunc
}

// Link implements param.SectionWriter
func (f *FileStore) Link(writerFunc func() param.SectionWriter) {
	f.w = writerFunc()
}

// Reset implements param.SectionWriter
func (f *FileStore) Reset(ctx context.Context) {
	f.ctx = ctx
}

// Write implements param.SectionWriter
// it asynchronously writes to the underlying writer while caching the data slice
func (f *FileStore) Write(index int, b []byte) {
	f.w.Write(index, b)
	f.data = append(f.data, b)
}

// Sum implements param.SectionWriter
// calls underlying writer's Sum and sends the result with data as a chunk to chunk.Store
func (f *FileStore) Sum(b []byte, length int, span []byte) []byte {
	ref := f.w.Sum(b, length, span)
	go func(ref []byte) {
		b = span
		for _, data := range f.data {
			b = append(b, data...)
		}
		ch := chunk.NewChunk(ref, b)
		_, err := f.chunkStore.Put(f.ctx, chunk.ModePutUpload, ch)
		log.Trace("filestore put chunk", "ch", ch)
		if err != nil {
			f.errFunc(err)
		}
	}(ref)
	return ref
}

// SectionSize implements param.SectionWriter
func (f *FileStore) SectionSize() int {
	return chunk.DefaultSize
}

// DigestSize implements param.SectionWriter
func (f *FileStore) DigestSize() int {
	return f.w.DigestSize()
}

// Branches implements param.SectionWriter
func (f *FileStore) Branches() int {
	return f.w.Branches()
}
