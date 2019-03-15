package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

const (
	defaultPadSize     = 18
	defaultSegmentSize = 32
)

const (
	writerModePool   = iota // use sync.Pool for managing hasher allocation
	writerModeGC            // only allocate new hashers, rely on GC to reap them
	writerModeManual        // handle a pre-allocated hasher pool with buffered channels
)

var (
	hashPool            sync.Pool
	mockPadding         = [defaultPadSize * defaultSegmentSize]byte{}
	FileHasherAlgorithm = DefaultHash
)

func init() {
	for i := 0; i < len(mockPadding); i++ {
		mockPadding[i] = 0x01
	}
	hashPool.New = func() interface{} {

		pool := bmt.NewTreePool(sha3.NewKeccak256, 128, bmt.PoolSize)
		h := bmt.New(pool)
		return h.NewAsyncWriter(false)
	}
}

func getHasher() bmt.SectionWriter {
	return hashPool.Get().(bmt.SectionWriter)
}

func putHasher(h bmt.SectionWriter) {
	h.Reset()
	hashPool.Put(h)
}

// defines the chained writer interface
type SectionHasherTwo interface {
	bmt.SectionWriter
	BatchSize() uint64 // sections to write before sum should be called
	PadSize() uint64   // additional sections that will be written on sum
}

// used for benchmarks against pyramid hasher which uses sync hasher
type treeHasherWrapper struct {
	*bmt.Hasher
}

func newTreeHasherWrapper() *treeHasherWrapper {
	pool := bmt.NewTreePool(sha3.NewKeccak256, 128, bmt.PoolSize)
	h := bmt.New(pool)
	return &treeHasherWrapper{
		Hasher: h,
	}
}

// implements SectionHasherTwo
func (h *treeHasherWrapper) Write(index int, b []byte) {
	h.Hasher.Write(b)
}

// implements SectionHasherTwo
func (h *treeHasherWrapper) Sum(b []byte, length int, span []byte) []byte {
	return h.Hasher.Sum(b)
}

// implements SectionHasherTwo
func (h *treeHasherWrapper) BatchSize() uint64 {
	return 128
}

// implements SectionHasherTwo
func (h *treeHasherWrapper) PadSize() uint64 {
	return 0
}

// implements SectionHasherTwo
func (h *treeHasherWrapper) SectionSize() int {
	return 32
}

// FileChunker is a chainable FileHasher writer that creates chunks on write and sum
// TODO not implemented
type FileChunker struct {
	branches uint64
}

func NewFileChunker() *FileChunker {
	return &FileChunker{
		branches: 128,
	}

}

// implements SectionHasherTwo
func (f *FileChunker) Write(index int, b []byte) {
	log.Trace("got write", "b", len(b))
}

// implements SectionHasherTwo
func (f *FileChunker) Sum(b []byte, length int, span []byte) []byte {
	log.Warn("got sum", "b", hexutil.Encode(b), "span", span)
	return b[:f.SectionSize()]
}

// implements SectionHasherTwo
func (f *FileChunker) BatchSize() uint64 {
	return branches
}

// implements SectionHasherTwo
func (f *FileChunker) PadSize() uint64 {
	return 0
}

// implements SectionHasherTwo
func (f *FileChunker) SectionSize() int {
	return 32
}

// implements SectionHasherTwo
func (f *FileChunker) Reset() {
	return
}

// FilePadder is a chainable FileHasher writer that pads the data written to it on sum
// illustrates possible erasure coding interface
type FilePadder struct {
	hasher bmt.SectionWriter
	writer SectionHasherTwo
	buffer []byte
	limit  int // write count limit (in segments)
	writeC chan int
}

func NewFilePadder(writer SectionHasherTwo) *FilePadder {
	if writer == nil {
		panic("writer can't be nil")
	}
	p := &FilePadder{
		writer: writer,
		limit:  110,
	}

	p.writeC = make(chan int, writer.BatchSize())
	p.Reset()
	return p
}

// implements SectionHasherTwo
func (p *FilePadder) BatchSize() uint64 {
	return p.writer.BatchSize() - p.PadSize()
}

// implements SectionHasherTwo
func (p *FilePadder) PadSize() uint64 {
	return 18
}

// implements SectionHasherTwo
func (p *FilePadder) Size() int {
	return p.hasher.SectionSize()
}

// implements SectionHasherTwo
// ignores index
// TODO bmt.SectionWriter.Write interface should return write count
func (p *FilePadder) Write(index int, b []byte) {
	//log.Debug("padder write", "index", index, "l", len(b), "c", atomic.AddUint64(&p.debugSize, uint64(len(b))))
	log.Debug("padder write", "index", index, "l", len(b))
	if index > p.limit {
		panic(fmt.Sprintf("write index beyond limit; %d > %d", index, p.limit))
	}
	p.hasher.Write(index, b)
	p.writeBuffer(index, b)
	p.writeC <- len(b)
}

func (p *FilePadder) writeBuffer(index int, b []byte) {
	bytesIndex := index * p.SectionSize()
	copy(p.buffer[bytesIndex:], b[:p.SectionSize()])
}

// implements SectionHasherTwo
// ignores span
func (p *FilePadder) Sum(b []byte, length int, span []byte) []byte {
	var writeCount int
	select {
	case c, ok := <-p.writeC:
		if !ok {
			break
		}
		writeCount += c
		if writeCount == length {
			break
		}
	}

	// at this point we are not concurrent anymore
	// TODO optimize
	padding := p.pad(nil)
	for i := 0; i < len(padding); i += p.hasher.SectionSize() {
		log.Debug("padder pad", "i", i, "limit", p.limit)
		p.hasher.Write(p.limit, padding[i:])
		p.writeBuffer(p.limit, padding[i:])
		p.limit++
	}
	s := p.hasher.Sum(b, length+len(padding), span)
	//p.writer.Sum(append(s, p.buffer...), length, span)
	chunk := NewChunk(Address(s), p.buffer)
	log.Warn("have chunk", "chunk", chunk, "chunkdata", chunk.sdata)
	putHasher(p.hasher)
	return s
}

// implements SectionHasherTwo
func (p *FilePadder) Reset() {
	p.hasher = getHasher()
	p.buffer = make([]byte, (p.PadSize()+p.BatchSize())*uint64(p.SectionSize()))
}

// implements SectionHasherTwo
// panics if called after sum and before reset
func (p *FilePadder) SectionSize() int {
	return p.hasher.SectionSize()
}

// performs data padding on the supplied data
// returns padding
func (p *FilePadder) pad(b []byte) []byte {
	return mockPadding[:]
}

// utility structure for controlling asynchronous tree hashing of the file
type hasherJob struct {
	parent        *hasherJob
	dataOffset    uint64 // global write count this job represents
	levelOffset   uint64 // offset on this level
	count         uint64 // amount of writes on this job
	edge          int    // > 0 on last write, incremented by 1 every level traversed on right edge
	debugHash     []byte
	debugLifetime uint32
	writer        SectionHasherTwo
}

// reuse hasherjob with new offsets
// not thread-safe
func (h *hasherJob) reset(w SectionHasherTwo, dataOffset uint64, levelOffset uint64, edge int) {
	h.debugLifetime++
	h.count = 0
	h.dataOffset = dataOffset
	h.levelOffset = levelOffset
	h.writer = w
}

func (h *hasherJob) inc() uint64 {
	return atomic.AddUint64(&h.count, 1)
}

// FileMuxer manages the build tree of the data
type FileMuxer struct {
	branches        int               // cached branch count
	sectionSize     int               // cached segment size of writer
	writerBatchSize uint64            // cached chunk size of chained writer
	parentBatchSize uint64            // cached number of writes before change parent
	writerPadSize   uint64            // cached padding size of the chained writer
	topJob          *hasherJob        // keeps pointer to the current topmost job
	lastJob         *hasherJob        // keeps pointer to the current data write job
	lastWrite       uint64            // keeps the last data write count
	targetCount     uint64            // set when sum is called, is total length of data
	targetLevel     int               // set when sum is called, is tree level of root chunk
	balancedTable   map[uint64]uint64 // maps write counts to bytecounts for
	debugJobChange  uint32            // debug counter for job reset calls
	debugJobCreate  uint32            // debug counter for new job allocations

	getWriter  func() SectionHasherTwo // mode-dependent function to assign hasher
	putWriter  func(SectionHasherTwo)  // mode-dependent function to release hasher
	writerFunc func() SectionHasherTwo // hasher function used by manual and GC modes

	writerQueue       chan struct{}         // throttles allocation of hashers
	writerPool        sync.Pool             // chained writers providing hashing in Pool mode
	writerManualQueue chan SectionHasherTwo // chained writers providing hashing in Manual mode
}

func NewFileMuxer(writerFunc func() SectionHasherTwo, mode int) (*FileMuxer, error) {

	if writerFunc == nil {
		return nil, errors.New("writer cannot be nil")
	}

	// create new instance and cache frequenctly used values
	branches := writer.BatchSize() + writer.PadSize()
	writer := writerFunc()
	f := &FileMuxer{
		branches:        int(branches),
		sectionSize:     writer.SectionSize(),
		writerBatchSize: writer.BatchSize(),
		parentBatchSize: writer.BatchSize() * branches,
		writerPadSize:   writer.PadSize(),
		//writerQueue:     make(chan struct{}, 1000),
		balancedTable: make(map[uint64]uint64),
		writerFunc:    writerFunc,
	}

	// see writerMode*
	switch mode {
	case writerModeManual:
		f.writerManualQueue = make(chan SectionHasherTwo, 1000)

		for i := 0; i < 1000; i++ {
			f.writerManualQueue <- writerFunc()
		}
		f.getWriter = f.getWriterManual
		f.putWriter = f.putWriterManual
	case writerModeGC:

		f.getWriter = f.getWriterGC
		f.putWriter = f.putWriterGC

	case writerModePool:
		f.writerPool.New = func() interface{} {
			return writerFunc()
		}
		f.getWriter = f.getWriterPool
		f.putWriter = f.putWriterPool
	}

	// create lookup table for data write counts that result in balanced trees
	lastBoundary := uint64(1)
	f.balancedTable[lastBoundary] = uint64(f.sectionSize)
	for i := 1; i < 9; i++ {
		lastBoundary *= uint64(f.branches)
		f.balancedTable[lastBoundary] = lastBoundary * uint64(f.sectionSize)
	}

	// create the hasherJob object for the data level.
	f.lastJob = &hasherJob{
		writer: f.getWriter(),
	}
	f.topJob = f.lastJob

	return f, nil
}

// implements SectionHasherTwo
func (m *FileMuxer) BatchSize() uint64 {
	return m.writerBatchSize + m.writerPadSize
}

// implements SectionHasherTwo
func (m *FileMuxer) PadSize() uint64 {
	return 0
}

// implements SectionHasherTwo
func (m *FileMuxer) SectionSize() int {
	return m.sectionSize
}

// implements SectionHasherTwo
func (m *FileMuxer) Write(index int, b []byte) {
	//log.Trace("data write", "offset", index, "jobcount", m.lastJob.count, "batchsize", m.writerBatchSize)

	m.write(m.lastJob, index%m.branches, b, true)
	m.lastWrite++
}

// implements SectionHasherTwo
// TODO is noop
func (m *FileMuxer) Sum(b []byte, length int, span []byte) []byte {
	log.Warn("filemux sum called, not implemented", "b", b, "l", length, "span", span)
	return nil
}

// implements SectionHasherTwo
// TODO is noop
func (m *FileMuxer) Reset() {
	log.Warn("filemux reset called, not implemented")
}

// handles recursive writing across tree levels
// b byte is not thread safe
// index is internal within a job (batchsize / sectionsize)
func (m *FileMuxer) write(h *hasherJob, index int, b []byte, groundlevel bool) {

	// if we are crossing a batch write size, we spawn a new job
	// and point the data writer's job pointer lastJob to it
	newcount := h.inc()
	if newcount > m.writerBatchSize {
	}

	// write the data to the chain and sum it if:
	// * the write is on a threshold, or
	// * if we're done writing
	//go func(h *hasherJob, newcount uint64, index int, b []byte) {
	lifetime := atomic.LoadUint32(&h.debugLifetime)
	log.Trace("job write", "job", fmt.Sprintf("%p", h), "w", fmt.Sprintf("%p", h.writer), "count", newcount, "index", index, "lifetime", lifetime, "data", hexutil.Encode(b))
	// write to the chained writer
	h.writer.Write(index, b)

	// check threshold or done
	if newcount == m.writerBatchSize || h.edge > 0 {

		//go func(index int, w SectionHasherTwo, p *hasherJob) {
		go m.sum(b, index, newcount, h.dataOffset, h.levelOffset, h, h.writer, h.parent)

		newLevelOffset := h.dataOffset + newcount - 1
		var sameParent bool
		if newLevelOffset%m.parentBatchSize > 0 {
			sameParent = true
		}
		newDataOffset := h.dataOffset
		if groundlevel {
			newDataOffset += newcount - 1
		}

		// TODO edge
		h.reset(m.getWriter(), newDataOffset, newLevelOffset, 0)

		// groundlevel is synchronous, so we don't have to worry about race here
		atomic.AddUint32(&m.debugJobChange, 1)
		log.Debug("changing jobs", "dataoffset", h.dataOffset, "leveloffset", h.levelOffset, "sameparent", sameParent, "groundlevel", groundlevel)

	}
}

// handles recursive feedback writes of chained sum call
// as the hasherJob from the context calling this is asynchronously reset
// the relevant values to use for calculation must be copied
// if parent doesn't exist (new level) a new one is created
// releases the hasher used by the hasherJob at time of calling this method
func (m *FileMuxer) sum(b []byte, index int, count uint64, dataOffset uint64, levelOffset uint64, job *hasherJob, w SectionHasherTwo, p *hasherJob) {

	thisJobLength := (count * uint64(m.sectionSize)) + uint64(len(b)%m.sectionSize)

	// span is the total size under the chunk
	// BUG dataoffset needs modulo levelindex
	spanBytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(spanBytes, uint64(dataOffset+thisJobLength))

	log.Debug("jobwrite sum", "w", fmt.Sprintf("%p", w), "l", thisJobLength, "span", spanBytes)
	// sum the data using the chained writer

	s := w.Sum(
		nil,
		int(thisJobLength),
		spanBytes,
	)

	// reset the chained writer
	m.putWriter(w)

	// we only create a parent object on a job on the first write
	// this way, if it is nil and we are working the right edge, we know when to skip
	if p == nil {
		job.parent = m.newJob(dataOffset, levelOffset)
		atomic.AddUint32(&m.debugJobCreate, 1)
		log.Debug("set parent", "child", fmt.Sprintf("%p", job), "parent", fmt.Sprintf("%p", job.parent))
	}
	// write to the parent job
	// the section index to write to is divided by the branches
	m.write(job.parent, (index-1)/m.branches, s, false)

	log.Debug("hash result", "s", hexutil.Encode(s), "length", thisJobLength)

}

// creates a new hasherJob
func (m *FileMuxer) newJob(dataOffset uint64, levelOffset uint64) *hasherJob {
	return &hasherJob{
		dataOffset:  dataOffset,
		levelOffset: (levelOffset-1)/uint64(m.branches) + 1,
		writer:      m.getWriter(),
	}
}

// see writerMode consts
func (m *FileMuxer) getWriterGC() SectionHasherTwo {
	return m.writerFunc()
}

// see writerMode consts
func (m *FileMuxer) putWriterGC(w SectionHasherTwo) {
	// noop
}

// see writerMode consts
func (m *FileMuxer) getWriterPool() SectionHasherTwo {
	//m.writerQueue <- struct{}{}
	return m.writerPool.Get().(SectionHasherTwo)
}

// see writerMode consts
func (m *FileMuxer) putWriterPool(writer SectionHasherTwo) {
	writer.Reset()
	m.writerPool.Put(writer)
	//<-m.writerQueue
}

// see writerMode consts
func (m *FileMuxer) getWriterManual() SectionHasherTwo {
	return <-m.writerManualQueue
}

// see writerMode consts
func (m *FileMuxer) putWriterManual(writer SectionHasherTwo) {
	writer.Reset()
	m.writerManualQueue <- writer
}

// calculates if the given data write length results in a balanced tree
func (m *FileMuxer) isBalancedBoundary(count uint64) bool {
	_, ok := m.balancedTable[count]
	return ok
}
