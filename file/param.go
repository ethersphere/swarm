package file

import (
	"sync"

	"github.com/ethersphere/swarm/bmt"
)

// defines the boundaries of the hashing job and also contains the hash factory functino of the job
// setting Debug means omitting any automatic behavior (for now it means job processing won't auto-start)
type treeParams struct {
	SectionSize int
	Branches    int
	Spans       []int
	Debug       bool
	hashFunc    func() bmt.SectionWriter
	writerPool  sync.Pool
}

func newTreeParams(section int, branches int, hashFunc func() bmt.SectionWriter) *treeParams {

	p := &treeParams{
		SectionSize: section,
		Branches:    branches,
		hashFunc:    hashFunc,
	}
	p.writerPool.New = func() interface{} {
		return hashFunc()
	}

	span := 1
	for i := 0; i < 9; i++ {
		p.Spans = append(p.Spans, span)
		span *= p.Branches
	}
	return p
}
