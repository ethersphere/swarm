package file

import (
	"context"
	"hash"
)

// SectionWriterFunc defines the required function signature to be used to create a SectionWriter
type SectionWriterFunc func(ctx context.Context) SectionWriter

// SectionWriter is a chainable interface for processing of chunk data
//
// Implementations should pass data to underlying writer before performing their own sum calculations
type SectionWriter interface {
	hash.Hash                                                 // Write,Sum,Reset,Size,BlockSize
	SetWriter(hashFunc SectionWriterFunc) SectionWriter       // chain another SectionWriter the current instance
	SetSpan(length int)                                       // set data span of chunk
	SectionSize() int                                         // section size of this SectionWriter
	Branches() int                                            // branch factor of this SectionWriter
	SumIndexed(prepended_data []byte, span_length int) []byte // Blocking call returning the sum of the data from underlying writer, setting the final data length to span_length
	WriteIndexed(int, []byte)                                 // Write to a particular data section, enabling asynchronous writing to any position of any level
}
