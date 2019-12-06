package param

import (
	"context"
)

type SectionWriterFunc func(ctx context.Context) SectionWriter

// SectionWriter is an asynchronous segment/section writer interface
type SectionWriter interface {
	Connect(hashFunc SectionWriterFunc) SectionWriter
	Init(ctx context.Context, errFunc func(error)) // errFunc is used for asynchronous components to signal error and termination
	Reset(ctx context.Context)                     // standard init to be called before reuse
	Write(index int, data []byte)                  // write into section of index
	Sum(b []byte, length int, span []byte) []byte  // returns the hash of the buffer
	SectionSize() int                              // size of the async section unit to use
	DigestSize() int
	Branches() int
}
