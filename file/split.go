package file

import (
	"sync"

	"github.com/ethersphere/swarm/bmt"
)

// it is intended to be chainable to accommodate for arbitrary chunk manipulation
// like encryption, erasure coding etc
type Hasher struct {
	target *target
	params *treeParams
	index  *jobIndex

	writeC     chan []byte
	doneC      chan struct{}
	job        *job // current level 1 job being written to
	writerPool sync.Pool
	hasherPool sync.Pool
	size       int
	count      int
}

// New creates a new Hasher object using the given sectionSize and branch factor
// hasherFunc is used to create *bmt.Hashers to hash the incoming data
// writerFunc is used as the underlying bmt.SectionWriter for the asynchronous hasher jobs. It may be pipelined to other components with the same interface
func New(sectionSize int, branches int, hasherFunc func() *bmt.Hasher, writerFunc func() bmt.SectionWriter) *Hasher {
	h := &Hasher{
		target: newTarget(),
		index:  newJobIndex(9),
		writeC: make(chan []byte, branches),
	}
	h.writerPool.New = func() interface{} {
		return writerFunc()
	}
	h.hasherPool.New = func() interface{} {
		return hasherFunc()
	}
	h.params = newTreeParams(sectionSize, branches, h.getWriter)
	h.job = newJob(h.params, h.target, h.index, 1, 0)

	return h
}

// TODO: enforce buffered writes and limits
// TODO: attempt omit modulo calc on every pass
func (h *Hasher) Write(b []byte) {
	if h.count%branches == 0 && h.count > 0 {
		h.job = h.job.Next()
	}
	go func(i int, jb *job) {
		hasher := h.getHasher(len(b))
		_, err := hasher.Write(b)
		if err != nil {
			panic(err)
		}
		jb.write(i%h.params.Branches, hasher.Sum(nil))
		h.putHasher(hasher)
	}(h.count, h.job)
	h.size += len(b)
	h.count++
}

func (h *Hasher) Sum(_ []byte) []byte {
	sectionCount := dataSizeToSectionIndex(h.size, h.params.SectionSize)
	targetLevel := getLevelsFromLength(h.size, h.params.SectionSize, h.params.Branches)
	h.target.Set(h.size, sectionCount, targetLevel)
	return <-h.target.Done()
}

// proxy for sync.Pool
func (h *Hasher) putHasher(w *bmt.Hasher) {
	h.hasherPool.Put(w)
}

// proxy for sync.Pool
func (h *Hasher) getHasher(l int) *bmt.Hasher {
	span := lengthToSpan(l)
	hasher := h.hasherPool.Get().(*bmt.Hasher)
	hasher.ResetWithLength(span)
	return hasher
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
