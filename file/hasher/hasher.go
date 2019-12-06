package hasher

import (
	"context"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
)

// Hasher is a bmt.SectionWriter that executes the file hashing algorithm on arbitary data
type Hasher struct {
	target  *target
	params  *treeParams
	index   *jobIndex
	errFunc func(error)
	ctx     context.Context

	job   *job // current level 1 job being written to
	size  int
	count int
}

// New creates a new Hasher object using the given sectionSize and branch factor
// hasherFunc is used to create *bmt.Hashers to hash the incoming data
// writerFunc is used as the underlying bmt.SectionWriter for the asynchronous hasher jobs. It may be pipelined to other components with the same interface
// TODO: sectionSize and branches should be inferred from underlying writer, not shared across job and hasher
func New(hashFunc param.SectionWriterFunc) *Hasher {
	h := &Hasher{
		target: newTarget(),
		index:  newJobIndex(9),
		params: newTreeParams(hashFunc),
	}
	h.job = newJob(h.params, h.target, h.index, 1, 0)
	return h
}

func (h *Hasher) Connect(hashFunc param.SectionWriterFunc) param.SectionWriter {
	h.params = newTreeParams(hashFunc)
	return h
}

// Init implements param.SectionWriter
func (h *Hasher) Init(ctx context.Context, errFunc func(error)) {
	h.errFunc = errFunc
	h.params.SetContext(ctx)
	h.job.start()
}

// Write implements param.SectionWriter
// It as a non-blocking call that hashes a data chunk and passes the resulting reference to the hash job representing
// the intermediate chunk holding the data references
// TODO: enforce buffered writes and limits
// TODO: attempt omit modulo calc on every pass
// TODO: preallocate full size span slice
func (h *Hasher) Write(index int, b []byte) {
	if h.count%h.params.Branches == 0 && h.count > 0 {
		h.job = h.job.Next()
	}
	go func(i int, jb *job) {
		hasher := h.params.GetWriter()
		hasher.Write(-1, b)
		l := len(b)
		log.Trace("data write", "count", i, "size", l)
		span := bmt.LengthToSpan(l)
		jb.write(i%h.params.Branches, hasher.Sum(nil, l, span))
		h.params.PutWriter(hasher)
	}(h.count, h.job)
	h.size += len(b)
	h.count++
}

// Sum implements param.SectionWriter
// It is a blocking call that calculates the target level and section index of the received data
// and alerts hasher jobs the end of write is reached
// It returns the root hash
func (h *Hasher) Sum(_ []byte, length int, _ []byte) []byte {
	sectionCount := dataSizeToSectionIndex(h.size, h.params.SectionSize)
	targetLevel := getLevelsFromLength(h.size, h.params.SectionSize, h.params.Branches)
	h.target.Set(h.size, sectionCount, targetLevel)
	return <-h.target.Done()
}

// Reset implements param.SectionWriter
func (h *Hasher) Reset(ctx context.Context) {
	h.params.ctx = ctx
}

// SectionSize implements param.SectionWriter
func (h *Hasher) SectionSize() int {
	return h.params.ChunkSize
}

// DigestSize implements param.SectionWriter
func (h *Hasher) DigestSize() int {
	return h.params.SectionSize
}

// DigestSize implements param.SectionWriter
func (h *Hasher) Branches() int {
	return h.params.Branches
}
