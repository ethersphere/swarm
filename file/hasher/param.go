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
	hashFunc    param.SectionWriterFunc
	writerPool  sync.Pool
	ctx         context.Context
}

func newTreeParams(hashFunc param.SectionWriterFunc) *treeParams {

	h := hashFunc(context.Background())
	p := &treeParams{
		SectionSize: h.SectionSize(),
		Branches:    h.Branches(),
		ChunkSize:   h.SectionSize() * h.Branches(),
		hashFunc:    hashFunc,
	}
	h.Reset(context.Background())
	log.Trace("new tree params", "sectionsize", p.SectionSize, "branches", p.Branches, "chunksize", p.ChunkSize)
	p.writerPool.New = func() interface{} {
		hf := p.hashFunc(p.ctx)
		log.Trace("param new hasher", "h", hf)
		return hf
	}
	p.Spans = generateSpanSizes(p.branches, 9)
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
