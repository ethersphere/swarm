package file

import (
	"sync"

	"github.com/ethersphere/swarm/bmt"
)

// Hasher implements file.SectionWriter
// it is intended to be chainable to accommodate for arbitrary chunk manipulation
// like encryption, erasure coding etc
type Hasher struct {
	writer *bmt.Hasher
	target *target
	params *treeParams
	index  *jobIndex

	writeC     chan []byte
	doneC      chan struct{}
	job        *job // current level 1 job being written to
	writerPool sync.Pool
	size       int
	count      int
}

// New creates a new Hasher object
func New(sectionSize int, branches int, dataWriter *bmt.Hasher, refWriterFunc func() bmt.SectionWriter) *Hasher {
	h := &Hasher{
		writer: dataWriter,
		target: newTarget(),
		index:  newJobIndex(9),
		writeC: make(chan []byte, branches),
	}
	h.writerPool.New = func() interface{} {
		return refWriterFunc()
	}
	h.params = newTreeParams(sectionSize, branches, h.getWriter)
	h.job = newJob(h.params, h.target, h.index, 1, 0)

	return h
}

// Write implements hash.Hash
// TODO: enforce buffered writes and limits
func (h *Hasher) Write(b []byte) {
	if h.count > 0 && h.count%branches == 0 {
		jb := h.job
		h.job = h.job.Next()
		jb.destroy()
	}
	span := lengthToSpan(len(b))
	h.writer.ResetWithLength(span)
	_, err := h.writer.Write(b)
	if err != nil {
		panic(err)
	}
	h.size += len(b)
	h.job.write(h.count%h.params.Branches, h.writer.Sum(nil))
	h.count++

}

// Sum implements hash.Hash
func (h *Hasher) Sum(_ []byte) []byte {
	sectionCount := dataSizeToSectionIndex(h.size, h.params.SectionSize)
	targetLevel := getLevelsFromLength(h.size, h.params.SectionSize, h.params.Branches)
	h.target.Set(h.size, sectionCount, targetLevel)
	var ref []byte
	select {
	case ref = <-h.target.Done():
	}
	return ref
}

// proxy for sync.Pool
func (h *Hasher) putWriter(w bmt.SectionWriter) {
	w.Reset()
	h.writerPool.Put(w)
}

// proxy for sync.Pool
func (h *Hasher) getWriter() bmt.SectionWriter {
	return h.writerPool.Get().(bmt.SectionWriter)
}
