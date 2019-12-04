package store

import (
	"context"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/param"
)

type FileStore struct {
	chunkStore chunk.Store
	w          param.SectionWriter
	ctx        context.Context
	data       [][]byte
	errFunc    func(error)
}

func New(chunkStore chunk.Store) *FileStore {
	return &FileStore{
		chunkStore: chunkStore,
	}
}

func (f *FileStore) Init(ctx context.Context, errFunc func(error)) {
	f.ctx = ctx
	f.errFunc = errFunc
}

func (f *FileStore) Link(writerFunc func() param.SectionWriter) {
	f.w = writerFunc()
}

func (f *FileStore) Reset(ctx context.Context) {
	f.ctx = ctx
}

func (f *FileStore) Write(index int, b []byte) {
	f.data = append(f.data, b)
}

func (f *FileStore) Sum(b []byte, length int, span []byte) []byte {
	ref := f.w.Sum(b, length, span)
	go func(ref []byte) {
		var b []byte
		for _, data := range f.data {
			b = append(b, data...)
		}
		ch := chunk.NewChunk(ref, b)
		_, err := f.chunkStore.Put(f.ctx, chunk.ModePutUpload, ch)
		if err != nil {
			f.errFunc(err)
		}
	}(ref)
	return ref
}

func (f *FileStore) SectionSize() int {
	return chunk.DefaultSize
}

func (f *FileStore) DigestSize() int {
	return f.w.DigestSize()
}
