package store

import (
	"context"

	"github.com/ethersphere/swarm/bmt"
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
	span       int
	errFunc    func(error)
}

// New creates a new FileStore with the supplied chunk.Store
func New(chunkStore chunk.Store, writerFunc param.SectionWriterFunc) *FileStore {
	f := &FileStore{
		chunkStore: chunkStore,
	}
	f.w = writerFunc(f.ctx)
	return f
}

func (f *FileStore) SetWriter(hashFunc param.SectionWriterFunc) param.SectionWriter {
	f.w = hashFunc(f.ctx)
	return f
}

// Init implements param.SectionWriter
func (f *FileStore) Init(ctx context.Context, errFunc func(error)) {
	f.ctx = ctx
	f.errFunc = errFunc
}

// Reset implements param.SectionWriter
func (f *FileStore) Reset() {
	f.span = 0
	f.data = [][]byte{}
	f.w.Reset()
}

func (f *FileStore) SeekSection(index int) {
	f.w.SeekSection(index)
}

// Write implements param.SectionWriter
// it asynchronously writes to the underlying writer while caching the data slice
func (f *FileStore) Write(b []byte) (int, error) {
	f.data = append(f.data, b)
	return f.w.Write(b)
}

// Sum implements param.SectionWriter
// calls underlying writer's Sum and sends the result with data as a chunk to chunk.Store
func (f *FileStore) Sum(b []byte) []byte {
	ref := f.w.Sum(b)
	go func(ref []byte) {
		b = bmt.LengthToSpan(f.span)
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

func (f *FileStore) SetSpan(length int) {
	f.span = length
	f.w.SetSpan(length)
}

func (f *FileStore) SetLength(length int) {
	f.w.SetLength(length)
}

// SectionSize implements param.SectionWriter
func (f *FileStore) BlockSize() int {
	return f.w.BlockSize()
}

// SectionSize implements param.SectionWriter
func (f *FileStore) SectionSize() int {
	return f.w.SectionSize()
}

// DigestSize implements param.SectionWriter
func (f *FileStore) Size() int {
	return f.w.Size()
}

// Branches implements param.SectionWriter
func (f *FileStore) Branches() int {
	return f.w.Branches()
}
