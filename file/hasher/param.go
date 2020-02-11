package hasher

import (
	"context"
	"sync"

	"github.com/ethersphere/swarm/file"
	"github.com/ethersphere/swarm/log"
)

// defines the boundaries of the hashing job and also contains the hash factory functino of the job
// setting Debug means omitting any automatic behavior (for now it means job processing won't auto-start)
type treeParams struct {
	SectionSize int
	Branches    int
	ChunkSize   int
	Spans       []int
	Debug       bool
	hashFunc    file.SectionWriterFunc
	writerPool  sync.Pool
	ctx         context.Context
}

func newTreeParams(hashFunc file.SectionWriterFunc) *treeParams {

	h := hashFunc(context.Background())
	p := &treeParams{
		SectionSize: h.SectionSize(),
		Branches:    h.Branches(),
		ChunkSize:   h.SectionSize() * h.Branches(),
		hashFunc:    hashFunc,
	}
	h.Reset()
	log.Trace("new tree params", "sectionsize", p.SectionSize, "branches", p.Branches, "chunksize", p.ChunkSize)
	p.writerPool.New = func() interface{} {
		hf := p.hashFunc(p.ctx)
		//log.Trace("param new hasher", "h", hf)
		return hf
	}
	p.Spans = generateSpanSizes(p.Branches, 9)
	return p
}

func (p *treeParams) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *treeParams) GetContext() context.Context {
	return p.ctx
}

func (p *treeParams) PutWriter(w file.SectionWriter) {
	w.Reset()
	p.writerPool.Put(w)
}

func (p *treeParams) GetWriter() file.SectionWriter {
	return p.writerPool.Get().(file.SectionWriter)
}
