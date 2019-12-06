package hasher

import (
	"context"
	"sync"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
)

// defines the boundaries of the hashing job and also contains the hash factory functino of the job
// setting Debug means omitting any automatic behavior (for now it means job processing won't auto-start)
type treeParams struct {
	SectionSize int
	Branches    int
	ChunkSize   int
	Spans       []int
	Debug       bool
	hashFunc    func() param.SectionWriter
	writerPool  sync.Pool
	ctx         context.Context
}

func newTreeParams(hashFunc func() param.SectionWriter) *treeParams {

	h := hashFunc()
	p := &treeParams{
		SectionSize: h.SectionSize(),
		Branches:    h.Branches(),
		ChunkSize:   h.SectionSize() * h.Branches(),
		hashFunc:    hashFunc,
		ctx:         context.Background(),
	}
	h.Reset(p.ctx)
	log.Trace("new tree params", "sectionsize", p.SectionSize, "branches", p.Branches, "chunksize", p.ChunkSize)
	p.writerPool.New = func() interface{} {
		return p.hashFunc()
	}

	span := 1
	for i := 0; i < 9; i++ {
		p.Spans = append(p.Spans, span)
		span *= p.Branches
	}
	return p
}

func (p *treeParams) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *treeParams) GetContext() context.Context {
	return p.ctx
}

func (p *treeParams) PutWriter(w param.SectionWriter) {
	w.Reset(p.ctx)
	p.writerPool.Put(w)

}

func (p *treeParams) GetWriter() param.SectionWriter {
	return p.writerPool.Get().(param.SectionWriter)
}
