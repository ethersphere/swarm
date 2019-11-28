package file

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
)

// keeps an index of all the existing jobs for a file hashing operation
// sorted by level
//
// it also keeps all the "top hashes", ie hashes on first data section index of every level
// these are needed in case of balanced tree results, since the hashing result would be
// lost otherwise, due to the job not having any intermediate storage of any data
type jobIndex struct {
	maxLevels int
	jobs      []sync.Map
	topHashes [][]byte
	mu        sync.Mutex
}

func newJobIndex(maxLevels int) *jobIndex {
	ji := &jobIndex{
		maxLevels: maxLevels,
	}
	for i := 0; i < maxLevels; i++ {
		ji.jobs = append(ji.jobs, sync.Map{})
	}
	return ji
}

// implements Stringer interface
func (ji *jobIndex) String() string {
	return fmt.Sprintf("%p", ji)
}

// Add adds a job to the index at the level
// and data section index specified in the job
func (ji *jobIndex) Add(jb *job) {
	log.Trace("adding job", "job", jb)
	ji.jobs[jb.level].Store(jb.dataSection, jb)
}

// Get retrieves a job from the job index
// based on the level of the job and its data section index
// if a job for the level and section index does not exist this method returns nil
func (ji *jobIndex) Get(lvl int, section int) *job {
	jb, ok := ji.jobs[lvl].Load(section)
	if !ok {
		return nil
	}
	return jb.(*job)
}

// Delete removes a job from the job index
// leaving it to be garbage collected when
// the reference in the main code is relinquished
func (ji *jobIndex) Delete(jb *job) {
	ji.jobs[jb.level].Delete(jb.dataSection)
}

// AddTopHash should be called by a job when a hash is written to the first index of a level
// since the job doesn't store any data written to it (just passing it through to the underlying writer)
// this is needed for the edge case of balanced trees
func (ji *jobIndex) AddTopHash(ref []byte) {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	ji.topHashes = append(ji.topHashes, ref)
	log.Trace("added top hash", "length", len(ji.topHashes), "index", ji)
}

// GetJobHash gets the current top hash for a particular level set by AddTopHash
func (ji *jobIndex) GetTopHash(lvl int) []byte {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	return ji.topHashes[lvl-1]
}

func (ji *jobIndex) GetTopHashLevel() int {
	ji.mu.Lock()
	defer ji.mu.Unlock()
	return len(ji.topHashes)
}

// passed to a job to determine at which data lengths and levels a job should terminate
type target struct {
	size     int32         // bytes written
	sections int32         // sections written
	level    int32         // target level calculated from bytes written against branching factor and sector size
	resultC  chan []byte   // channel to receive root hash
	doneC    chan struct{} // when this channel is closed all jobs will calculate their end write count
	mu       sync.Mutex
}

func newTarget() *target {
	return &target{
		resultC: make(chan []byte),
		doneC:   make(chan struct{}),
	}
}

// Set is called when the final length of the data to be written is known
// TODO: method can be simplified to calculate sections and level internally
func (t *target) Set(size int, sections int, level int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.size = int32(size)
	t.sections = int32(sections)
	t.level = int32(level)
	//	atomic.StoreInt32(&t.size, int32(size))
	//	atomic.StoreInt32(&t.sections, int32(sections))
	//	atomic.StoreInt32(&t.level, int32(level))
	log.Trace("target set", "size", size, "sections", sections, "level", level)
	close(t.doneC)
}

// Count returns the total section count for the target
// it should only be called after Set()
func (t *target) Count() int {
	return int(atomic.LoadInt32(&t.sections)) + 1
}

// Done returns the channel in which the root hash will be sent
func (t *target) Done() <-chan []byte {
	return t.resultC
}

type jobUnit struct {
	index int
	data  []byte
}

// encapsulates one single chunk to be hashed
type job struct {
	target *target
	params *treeParams
	index  *jobIndex

	level            int    // level in tree
	dataSection      int    // data section index
	cursorSection    int32  // next write position in job
	endCount         int32  // number of writes to be written to this job (0 means write to capacity)
	lastSectionSize  int    // data size on the last data section write
	firstSectionData []byte // store first section of data written to solve the dangling chunk edge case

	writeC chan jobUnit
	writer bmt.SectionWriter // underlying data processor

	mu sync.Mutex
}

func newJob(params *treeParams, tgt *target, jobIndex *jobIndex, lvl int, dataSection int) *job {
	jb := &job{
		params:      params,
		index:       jobIndex,
		level:       lvl,
		dataSection: dataSection,
		writer:      params.hashFunc(),
		writeC:      make(chan jobUnit),
		target:      tgt,
	}
	if jb.index == nil {
		jb.index = newJobIndex(9)
	}

	jb.index.Add(jb)
	if !params.Debug {
		go jb.process()
	}
	return jb
}

// implements Stringer interface
func (jb *job) String() string {
	return fmt.Sprintf("job: l:%d,s:%d,c:%d", jb.level, jb.dataSection, jb.count())
}

// atomically increments the write counter of the job
func (jb *job) inc() int {
	return int(atomic.AddInt32(&jb.cursorSection, 1))
	//jb.cursorSection++
	//return int(jb.cursorSection)
}

// atomically returns the write counter of the job
func (jb *job) count() int {
	return int(atomic.LoadInt32(&jb.cursorSection))
	//return int(jb.cursorSection)
}

// size returns the byte size of the span the job represents
// if job is last index in a level and writes have been finalized, it will return the target size
// otherwise, regardless of job index, it will return the size according to the current write count
// TODO: returning expected size in one case and actual size in another can lead to confusion
// TODO: two atomic ops, may change value inbetween
func (jb *job) size() int {
	count := jb.count()
	endCount := int(atomic.LoadInt32(&jb.endCount))
	if endCount == 0 {
		return count * jb.params.SectionSize * jb.params.Spans[jb.level]
	}
	log.Trace("size", "sections", jb.target.sections, "endcount", endCount, "level", jb.level)
	return int(jb.target.size) % (jb.params.Spans[jb.level] * jb.params.SectionSize * jb.params.Branches)
}

// add data to job
// does no checking for data length or index validity
func (jb *job) write(index int, data []byte) {

	// if a write is received at the first datasection of a level we need to store this hash
	// in case of a balanced tree and we need to send it to resultC later
	// at the time of hasing of a balanced tree we have no way of knowing for sure whether
	// that is the end of the job or not
	if jb.dataSection == 0 {
		topHashLevel := jb.index.GetTopHashLevel()
		if topHashLevel < jb.level {
			log.Trace("have tophash", "level", jb.level, "ref", hexutil.Encode(data))
			jb.index.AddTopHash(data)
		}
	}
	jb.writeC <- jobUnit{
		index: index,
		data:  data,
	}
}

// runs in loop until:
// - sectionSize number of job writes have occurred (one full chunk)
// - data write is finalized and targetcount for this chunk was already reached
// - data write is finalized and targetcount is reached on a subsequent job write
func (jb *job) process() {

	doneC := jb.target.doneC
	defer jb.destroy()

	// is set when data write is finished, AND
	// the final data section falls within the span of this job
	// if not, loop will only exit on Branches writes
	endCount := 0
OUTER:
	for {
		select {

		// enter here if new data is written to the job
		case entry := <-jb.writeC:
			jb.mu.Lock()
			if entry.index == 0 {
				jb.firstSectionData = entry.data
			}
			newCount := jb.inc()
			log.Trace("job write", "datasection", jb.dataSection, "level", jb.level, "count", newCount, "endcount", endCount, "index", entry.index, "data", hexutil.Encode(entry.data))
			// this write is superfluous when the received data is the root hash
			jb.writer.Write(entry.index, entry.data)

			// since newcount is incremented above it can only equal endcount if this has been set in the case below,
			// which means data write has been completed
			// otherwise if we reached the chunk limit we also continue to hashing
			if newCount == endCount {
				log.Trace("quitting writec - endcount")
				jb.mu.Unlock()
				break OUTER
			}
			if newCount == jb.params.Branches {
				log.Trace("quitting writec - branches")
				jb.mu.Unlock()
				break OUTER
			}
			jb.mu.Unlock()

		// enter here if data writes have been completed
		// TODO: this case currently executes for all cycles after data write is complete for which writes to this job do not happen. perhaps it can be improved
		case <-doneC:

			jb.mu.Lock()

			// we can never have count 0 and have a completed job
			// this is the easiest check we can make
			log.Trace("doneloop", "level", jb.level, "count", jb.count(), "endcount", endCount)
			count := jb.count()
			if count == 0 {
				jb.mu.Unlock()
				continue
			}
			doneC = nil

			// if the target count falls within the span of this job
			// set the endcount so we know we have to do extra calculations for
			// determining span in case of unbalanced tree
			targetCount := jb.target.Count()
			endCount = jb.targetCountToEndCount(targetCount)
			jb.endCount = int32(endCount)
			//atomic.StoreInt32(&jb.endCount, int32(endCount))
			log.Trace("doneloop done", "level", jb.level, "targetcount", jb.target.Count(), "endcount", endCount)

			// if we have reached the end count for this chunk, we proceed to hashing
			// this case is important when write to the level happen after this goroutine
			// registers that data writes have been completed
			if count == int(endCount) {
				log.Trace("quitting donec", "level", jb.level, "count", jb.count())
				jb.mu.Unlock()
				break OUTER
			}
			jb.mu.Unlock()
		}
	}

	targetLevel := atomic.LoadInt32(&jb.target.level)
	if int(targetLevel) == jb.level {
		jb.target.resultC <- jb.index.GetTopHash(jb.level)
		return
	}

	// get the size of the span and execute the hash digest of the content
	size := jb.size()
	span := lengthToSpan(size)
	refSize := jb.count() * jb.params.SectionSize
	log.Trace("job sum", "count", jb.count(), "refsize", refSize, "size", size, "datasection", jb.dataSection, "span", span, "level", jb.level, "targetlevel", targetLevel, "endcount", endCount)
	ref := jb.writer.Sum(nil, refSize, span)

	// endCount > 0 means this is the last chunk on the level
	// the hash from the level below the target level will be the result
	belowRootLevel := int(targetLevel) - 1
	if endCount > 0 && jb.level == belowRootLevel {
		jb.target.resultC <- ref
		return
	}

	// retrieve the parent and the corresponding section in it to write to
	parent := jb.parent()
	log.Trace("have parent", "level", jb.level, "jb p", fmt.Sprintf("%p", jb), "jbp p", fmt.Sprintf("%p", parent))
	nextLevel := jb.level + 1
	parentSection := dataSectionToLevelSection(jb.params, nextLevel, jb.dataSection)

	// in the event that we have a balanced tree and a chunk with single reference below the target level
	// we move the single reference up to the penultimate level
	if endCount == 1 {
		ref = jb.firstSectionData
		for parent.level < belowRootLevel {
			log.Trace("parent write skip", "level", parent.level)
			oldParent := parent
			parent = parent.parent()
			oldParent.destroy()
			nextLevel += 1
			parentSection = dataSectionToLevelSection(jb.params, nextLevel, jb.dataSection)
		}
	}
	parent.write(parentSection, ref)

}

// determine whether the given data section count falls within the span of the current job
func (jb *job) targetWithinJob(targetSection int) (int, bool) {
	var endIndex int
	var ok bool

	// span one level above equals the data size of 128 units of one section on this level
	// using the span table saves one multiplication
	//dataBoundary := dataSectionToLevelBoundary(jb.params, jb.level, jb.dataSection)
	dataBoundary := dataSectionToLevelBoundary(jb.params, jb.level, jb.dataSection)
	upperLimit := dataBoundary + jb.params.Spans[jb.level+1]

	// the data section is the data section index where the span of this job starts
	if targetSection >= dataBoundary && targetSection < upperLimit {

		// data section index must be divided by corresponding section size on the job's level
		// then wrap on branch period to find the correct section within this job
		endIndex = (targetSection / jb.params.Spans[jb.level]) % jb.params.Branches

		ok = true
	}
	log.Trace("within", "level", jb.level, "datasection", jb.dataSection, "boundary", dataBoundary, "upper", upperLimit, "target", targetSection, "endindex", endIndex, "ok", ok)
	return int(endIndex), ok
}

// if last data index falls within the span, return the appropriate end count for the level
// otherwise return 0 (which means job write until limit)
func (jb *job) targetCountToEndCount(targetCount int) int {
	endIndex, ok := jb.targetWithinJob(targetCount - 1)
	if !ok {
		return 0
	}
	return endIndex + 1
}

// returns the parent job of the receiver job
// a new parent job is created if none exists for the slot
func (jb *job) parent() *job {
	jb.index.mu.Lock()
	defer jb.index.mu.Unlock()
	newLevel := jb.level + 1
	// Truncate to even quotient which is the actual logarithmic boundary of the data section under the span
	newDataSection := dataSectionToLevelBoundary(jb.params, jb.level+1, jb.dataSection)
	parent := jb.index.Get(newLevel, newDataSection)
	if parent != nil {
		return parent
	}
	return newJob(jb.params, jb.target, jb.index, jb.level+1, newDataSection)
}

// Next creates the job for the next data section span on the same level as the receiver job
// this is only meant to be called once for each job, consequtive calls will overwrite index with new empty job
func (jb *job) Next() *job {
	return newJob(jb.params, jb.target, jb.index, jb.level, jb.dataSection+jb.params.Spans[jb.level+1])
}

// cleans up the job; reset hasher and remove pointer to job from index
func (jb *job) destroy() {
	jb.writer.Reset()
	jb.index.Delete(jb)
}
