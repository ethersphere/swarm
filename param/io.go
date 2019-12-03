package param

import (
	"context"
)

// SectionWriter is an asynchronous segment/section writer interface
type SectionWriter interface {
	Link(writerFunc func() SectionWriter)
	Reset(ctx context.Context)                    // standard init to be called before reuse
	Write(index int, data []byte)                 // write into section of index
	Sum(b []byte, length int, span []byte) []byte // returns the hash of the buffer
	SectionSize() int                             // size of the async section unit to use
	DigestSize() int
}
