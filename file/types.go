package file

import (
	"context"
	"hash"
)

type SectionWriterFunc func(ctx context.Context) SectionWriter

type SectionWriter interface {
	hash.Hash                                           // Write,Sum,Reset,Size,BlockSize
	SetWriter(hashFunc SectionWriterFunc) SectionWriter // chain another SectionWriter the current instance
	SetSpan(length int)                                 // set data span of chunk
	SectionSize() int                                   // section size of this SectionWriter
	Branches() int                                      // branch factor of this SectionWriter
}
