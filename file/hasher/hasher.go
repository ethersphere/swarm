package hasher

import (
	"context"
	"sync"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/param"
)

// BMTSyncSectionWriter is a wrapper for bmt.Hasher to implement the param.SectionWriter interface
type BMTSyncSectionWriter struct {
	hasher *bmt.Hasher
	data   []byte
}

// NewBMTSyncSectionWriter creates a new BMTSyncSectionWriter
func NewBMTSyncSectionWriter(hasher *bmt.Hasher) param.SectionWriter {
	return &BMTSyncSectionWriter{
		hasher: hasher,
	}
}

// Init implements param.SectionWriter
func (b *BMTSyncSectionWriter) Init(_ context.Context, errFunc func(error)) {
}

// Link implements param.SectionWriter
func (b *BMTSyncSectionWriter) Link(_ func() param.SectionWriter) {
}

// Sum implements param.SectionWriter
func (b *BMTSyncSectionWriter) Sum(extra []byte, _ int, span []byte) []byte {
	b.hasher.ResetWithLength(span)
	b.hasher.Write(b.data)
	return b.hasher.Sum(extra)
}

// Reset implements param.SectionWriter
func (b *BMTSyncSectionWriter) Reset(_ context.Context) {
	b.hasher.Reset()
}

// Write implements param.SectionWriter
func (b *BMTSyncSectionWriter) Write(_ int, data []byte) {
	b.data = data
}

// SectionSize implements param.SectionWriter
func (b *BMTSyncSectionWriter) SectionSize() int {
	return b.hasher.ChunkSize()
}

// DigestSize implements param.SectionWriter
func (b *BMTSyncSectionWriter) DigestSize() int {
	return b.hasher.Size()
}

// Branches implements param.SectionWriter
func (b *BMTSyncSectionWriter) Branches() int {
	return b.hasher.Count()
}

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
// TODO: sectionSize and branches should be inferred from underlying writer, not shared across job and hasher
func New(sectionSize int, branches int, hasherFunc func() param.SectionWriter) *Hasher {
	h := &Hasher{
		target: newTarget(),
		index:  newJobIndex(9),
	}
	h.params = newTreeParams(sectionSize, branches, h.getWriter)
	h.writerPool.New = func() interface{} {
		return h.params.hashFunc()
	}
	h.hasherPool.New = func() interface{} {
		return hasherFunc()
	}
	h.job = newJob(h.params, h.target, h.index, 1, 0)
	return h
}

// Init implements param.SectionWriter
func (h *Hasher) Init(ctx context.Context, errFunc func(error)) {
	h.params.SetContext(ctx)
}

// Link implements param.SectionWriter
func (h *Hasher) Link(writerFunc func() param.SectionWriter) {
	h.params.hashFunc = writerFunc
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
		hasher := h.getHasher(len(b))
		hasher.Write(0, b)
		l := len(b)
		span := bmt.LengthToSpan(l)
		jb.write(i%h.params.Branches, hasher.Sum(nil, l, span))
		h.putHasher(hasher)
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

// proxy for sync.Pool
func (h *Hasher) putHasher(w param.SectionWriter) {
	h.hasherPool.Put(w)
}

// proxy for sync.Pool
func (h *Hasher) getHasher(l int) param.SectionWriter {
	//span := bmt.LengthToSpan(l)
	hasher := h.hasherPool.Get().(param.SectionWriter)
	hasher.Reset(h.params.ctx) //WithLength(span)
	return hasher
}

// proxy for sync.Pool
func (h *Hasher) putWriter(w param.SectionWriter) {
	w.Reset(h.params.ctx)
	h.writerPool.Put(w)
}

// proxy for sync.Pool
func (h *Hasher) getWriter() param.SectionWriter {
	return h.writerPool.Get().(param.SectionWriter)
}
