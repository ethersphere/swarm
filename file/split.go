package file

import (
	"io"

	"github.com/ethersphere/swarm/bmt"
)

type Splitter struct {
	r io.Reader
	w bmt.SectionWriter
}

func NewSplitter(r io.Reader, w bmt.SectionWriter) *Splitter {
	s := &Splitter{
		r: r,
		w: w,
	}
	return s
}

// TODO: enforce buffer capacity and auto-grow
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
