package hasher

import (
	"sync"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/param"
)

// Hasher is a bmt.SectionWriter that executes the file hashing algorithm on arbitary data
type Hasher struct {
	target *target
	params *treeParams
	index  *jobIndex

	job        *job // current level 1 job being written to
	writerPool sync.Pool
	hasherPool sync.Pool
	size       int
	count      int
}

// New creates a new Hasher object using the given sectionSize and branch factor
// hasherFunc is used to create *bmt.Hashers to hash the incoming data
// writerFunc is used as the underlying bmt.SectionWriter for the asynchronous hasher jobs. It may be pipelined to other components with the same interface
func New(sectionSize int, branches int, hasherFunc func() *bmt.Hasher, writerFunc func() param.SectionWriter) *Hasher {
	h := &Hasher{
		target: newTarget(),
		index:  newJobIndex(9),
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

// Write implements bmt.SectionWriter
// It as a non-blocking call that hashes a data chunk and passes the resulting reference to the hash job representing
// the intermediate chunk holding the data references
// TODO: enforce buffered writes and limits
// TODO: attempt omit modulo calc on every pass
func (h *Hasher) Write(index int, b []byte) {
	if h.count%h.params.Branches == 0 && h.count > 0 {
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

// Sum implements bmt.SectionWriter
// It is a blocking call that calculates the target level and section index of the received data
// and alerts hasher jobs the end of write is reached
// It returns the root hash
func (h *Hasher) Sum(_ []byte, length int, _ []byte) []byte {
	sectionCount := dataSizeToSectionIndex(h.size, h.params.SectionSize)
	targetLevel := getLevelsFromLength(h.size, h.params.SectionSize, h.params.Branches)
	h.target.Set(h.size, sectionCount, targetLevel)
	return <-h.target.Done()
}

func (h *Hasher) Reset() {
}

func (h *Hasher) SectionSize() int {
	return h.params.ChunkSize
}

func (h *Hasher) Branches() int {
	return h.params.Branches
}

func (h *Hasher) ChunkSize() int {
	return h.params.ChunkSize
}

func (h *Hasher) DigestSize() int {
	return h.params.SectionSize
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
func (h *Hasher) putWriter(w param.SectionWriter) {
	w.Reset()
	h.writerPool.Put(w)
}

// proxy for sync.Pool
func (h *Hasher) getWriter() param.SectionWriter {
	return h.writerPool.Get().(param.SectionWriter)
}
