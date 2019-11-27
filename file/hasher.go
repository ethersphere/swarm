package file

import (
	"sync"

	"github.com/ethersphere/swarm/bmt"
)

// Hasher implements file.SectionWriter
// it is intended to be chainable to accommodate for arbitrary chunk manipulation
// like encryption, erasure coding etc
type Hasher struct {
	writer     *bmt.Hasher
	target     *target
	params     *treeParams
	lastJob    *job
	jobMu      sync.Mutex
	writerPool sync.Pool
	size       int
}

// New creates a new Hasher object
func New(sectionSize int, branches int, dataWriter *bmt.Hasher, refWriterFunc func() bmt.SectionWriter) *Hasher {
	h := &Hasher{
		writer: dataWriter,
		target: newTarget(),
	}
	h.writerPool.New = func() interface{} {
		return refWriterFunc()
	}
	h.params = newTreeParams(sectionSize, branches, h.getWriter)

	return h
}

// Write implements hash.Hash
func (h *Hasher) Write(b []byte) {
	_, err := h.writer.Write(b)
	if err != nil {
		panic(err)
	}
}

// Sum implements hash.Hash
func (h *Hasher) Sum(_ []byte) []byte {
	sectionCount := dataSizeToSectionIndex(h.size, h.params.SectionSize) + 1
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
