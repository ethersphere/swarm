package file

import (
	"io"

	"github.com/ethersphere/swarm/bmt"
)

// TODO: grow buffer on demand to reduce allocs
// Splitter returns the result of a data stream from a bmt.SectionWriter
type Splitter struct {
	r io.Reader
	w bmt.SectionWriter
}

// NewSplitter creates a new Splitter object
func NewSplitter(r io.Reader, w bmt.SectionWriter) *Splitter {
	s := &Splitter{
		r: r,
		w: w,
	}
	return s
}

// Split is a blocking call that consumes and passes data from its reader to its SectionWriter
// according to the SectionWriter's SectionSize
// On EOF from the reader it calls Sum on the bmt.SectionWriter and returns the result
func (s *Splitter) Split() ([]byte, error) {
	wc := 0
	l := 0
	for {
		d := make([]byte, s.w.SectionSize())
		c, err := s.r.Read(d)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		s.w.Write(wc, d)
		wc++
		l += c
	}
	return s.w.Sum(nil, l, nil), nil
}
