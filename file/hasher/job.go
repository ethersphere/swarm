package hasher

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
)

// necessary metadata across asynchronous input
type jobUnit struct {
	index int
	data  []byte
	count int
}

// encapsulates one single intermediate chunk to be hashed
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
	writer param.SectionWriter // underlying data processor
	doneC  chan struct{}       // pointer to target doneC channel, set to nil in process() when closed

	mu sync.Mutex
}

func newJob(params *treeParams, tgt *target, jobIndex *jobIndex, lvl int, dataSection int) *job {
	jb := &job{
		params:      params,
		index:       jobIndex,
		level:       lvl,
		dataSection: dataSection,
		writeC:      make(chan jobUnit),
		target:      tgt,
		doneC:       nil,
	}
	if jb.index == nil {
		jb.index = newJobIndex(9)
	}
	targetLevel := tgt.Level()
	if targetLevel == 0 {
		log.Trace("target not set", "level", lvl)
		jb.doneC = tgt.doneC

	} else {
		targetCount := tgt.Count()
		jb.endCount = int32(jb.targetCountToEndCount(targetCount))
	}
	log.Trace("target count", "level", lvl, "count", tgt.Count())

	jb.index.Add(jb)
	return jb
}

func (jb *job) start() {
	jb.writer = jb.params.GetWriter()
	go jb.process()
}

// implements Stringer interface
func (jb *job) String() string {
	return fmt.Sprintf("job: l:%d,s:%d", jb.level, jb.dataSection)
}

// atomically increments the write counter of the job
func (jb *job) inc() int {
	return int(atomic.AddInt32(&jb.cursorSection, 1))
}

// atomically returns the write counter of the job
func (jb *job) count() int {
	return int(atomic.LoadInt32(&jb.cursorSection))
}

// size returns the byte size of the span the job represents
// if job is last index in a level and writes have been finalized, it will return the target size
// otherwise, regardless of job index, it will return the size according to the current write count
// TODO: returning expected size in one case and actual size in another can lead to confusion
func (jb *job) size() int {
	jb.mu.Lock()
	count := int(jb.cursorSection) //jb.count()
	endCount := int(jb.endCount)   //int(atomic.LoadInt32(&jb.endCount))
	jb.mu.Unlock()
	if endCount%jb.params.Branches == 0 {
		return count * jb.params.SectionSize * jb.params.Spans[jb.level]
	}
	log.Trace("size", "sections", jb.target.sections, "size", jb.target.Size(), "endcount", endCount, "level", jb.level)
	return jb.target.Size() % (jb.params.Spans[jb.level] * jb.params.SectionSize * jb.params.Branches)
}

// add data to job
// does no checking for data length or index validity
// TODO: rename index param not to confuse with index object
func (jb *job) write(index int, data []byte) {

	jb.inc()

	// if a write is received at the first datasection of a level we need to store this hash
	// in case of a balanced tree and we need to send it to resultC later
	// at the time of hasing of a balanced tree we have no way of knowing for sure whether
	// that is the end of the job or not
	if jb.dataSection == 0 && index == 0 {
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

	log.Trace("starting job process", "level", jb.level, "sec", jb.dataSection)

	var processCount int
	defer jb.destroy()

	// is set when data write is finished, AND
	// the final data section falls within the span of this job
	// if not, loop will only exit on Branches writes
OUTER:
	for {
		select {

		// enter here if new data is written to the job
		// TODO: Error if calculated write count exceed chunk
		case entry := <-jb.writeC:

			// split the contents to fit the underlying SectionWriter
			entrySections := len(entry.data) / jb.writer.SectionSize()
			jb.mu.Lock()
			endCount := int(jb.endCount)
			oldProcessCount := processCount
			processCount += entrySections
			jb.mu.Unlock()
			if entry.index == 0 {
				jb.firstSectionData = entry.data
			}
			log.Trace("job entry", "datasection", jb.dataSection, "num sections", entrySections, "level", jb.level, "processCount", oldProcessCount, "endcount", endCount, "index", entry.index, "data", hexutil.Encode(entry.data))

			// TODO: this write is superfluous when the received data is the root hash
			var offset int
			for i := 0; i < entrySections; i++ {
				idx := entry.index + i
				data := entry.data[offset : offset+jb.writer.SectionSize()]
				log.Trace("job write", "datasection", jb.dataSection, "level", jb.level, "processCount", oldProcessCount+i, "endcount", endCount, "index", entry.index+i, "data", hexutil.Encode(data))
				jb.writer.SeekSection(idx)
				jb.writer.Write(data)
				offset += jb.writer.SectionSize()
			}

			// since newcount is incremented above it can only equal endcount if this has been set in the case below,
			// which means data write has been completed
			// otherwise if we reached the chunk limit we also continue to hashing
			if processCount == endCount {
				log.Trace("quitting writec - endcount", "c", processCount, "level", jb.level)
				break OUTER
			}
			if processCount == jb.writer.Branches() {
				log.Trace("quitting writec - branches")
				break OUTER
			}

		// enter here if data writes have been completed
		// TODO: this case currently executes for all cycles after data write is complete for which writes to this job do not happen. perhaps it can be improved
		case <-jb.doneC:
			jb.mu.Lock()
			jb.doneC = nil
			log.Trace("doneloop", "level", jb.level, "processCount", processCount, "endcount", jb.endCount)
			//count := jb.count()

			// if the target count falls within the span of this job
			// set the endcount so we know we have to do extra calculations for
			// determining span in case of unbalanced tree
			targetCount := jb.target.Count()
			jb.endCount = int32(jb.targetCountToEndCount(targetCount))
			log.Trace("doneloop done", "level", jb.level, "targetcount", jb.target.Count(), "endcount", jb.endCount)

			// if we have reached the end count for this chunk, we proceed to hashing
			// this case is important when write to the level happen after this goroutine
			// registers that data writes have been completed
			if processCount > 0 && processCount == int(jb.endCount) {
				log.Trace("quitting donec", "level", jb.level, "processcount", processCount)
				jb.mu.Unlock()
				break OUTER
			}
			jb.mu.Unlock()
		}
	}

	jb.sum()
}

func (jb *job) sum() {

	targetLevel := jb.target.Level()
	if targetLevel == jb.level {
		jb.target.resultC <- jb.index.GetTopHash(jb.level)
		return
	}

	// get the size of the span and execute the hash digest of the content
	size := jb.size()
	//span := bmt.LengthToSpan(size)
	refSize := jb.count() * jb.params.SectionSize
	jb.writer.SetLength(refSize)
	jb.writer.SetSpan(size)
	log.Trace("job sum", "count", jb.count(), "refsize", refSize, "size", size, "datasection", jb.dataSection, "level", jb.level, "targetlevel", targetLevel, "endcount", jb.endCount)
	ref := jb.writer.Sum(nil)

	// endCount > 0 means this is the last chunk on the level
	// the hash from the level below the target level will be the result
	belowRootLevel := targetLevel - 1
	if jb.endCount > 0 && jb.level == belowRootLevel {
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
	if jb.endCount == 1 {
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
	return endIndex, ok
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
	jbp := newJob(jb.params, jb.target, jb.index, jb.level+1, newDataSection)
	jbp.start()
	return jbp
}

// Next creates the job for the next data section span on the same level as the receiver job
// this is only meant to be called once for each job, consecutive calls will overwrite index with new empty job
func (jb *job) Next() *job {
	jbn := newJob(jb.params, jb.target, jb.index, jb.level, jb.dataSection+jb.params.Spans[jb.level+1])
	jbn.start()
	return jbn
}

// cleans up the job; reset hasher and remove pointer to job from index
func (jb *job) destroy() {
	if jb.writer != nil {
		jb.params.PutWriter(jb.writer)
	}
	jb.index.Delete(jb)
}
