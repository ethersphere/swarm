package param

import (
	"context"
	"hash"
)

type SectionWriterFunc func(ctx context.Context) SectionWriter

type SectionWriter interface {
	hash.Hash
	SetWriter(hashFunc SectionWriterFunc) SectionWriter
	SeekSection(section int)
	Init(ctx context.Context, errFunc func(error)) // errFunc is used for asynchronous components to signal error and termination
	SetLength(length int)
	SetSpan(length int)
	SectionSize() int // size of the async section unit to use
	Branches() int
}
