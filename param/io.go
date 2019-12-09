package param

import (
	"context"
	"hash"
	"io"
)

type SectionWriterFunc func(ctx context.Context) SectionWriter

type SectionWriter interface {
	hash.Hash
	io.Seeker
	SetWriter(hashFunc SectionWriterFunc) SectionWriter
	Init(ctx context.Context, errFunc func(error)) // errFunc is used for asynchronous components to signal error and termination
	SetLength(length int)
	SectionSize() int // size of the async section unit to use
	DigestSize() int
	Branches() int
}
