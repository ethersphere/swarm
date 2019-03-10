package storage

import (
	"context"
	"encoding/binary"
	"sync"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/bmt"
)

const (
	altFileHasherMaxLevels = 9 // 22 zetabytes should be enough for anyone
)

type AltFileHasher struct {
	ctx           context.Context // per job context
	branches      int
	segmentSize   int
	chunkSize     int
	batchSegments int
	//hashers       [altFileHasherMaxLevels]bmt.SectionWriter
	//buffers       [altFileHasherMaxLevels][]byte           // holds chunk data on each level (todo; push data to channel on complete). Buffers can hold one batch of data
	levelJobs   [altFileHasherMaxLevels]chan fileHashJob // receives finished writes pending hashing to pass on to output handler
	levelWriteC [altFileHasherMaxLevels]chan []byte
	levelCount  int // number of levels in this job (only determined when Finish() is called
	//finished      bool                                     // finished writing data
	totalBytes     int                             // total data bytes written
	targetCount    [altFileHasherMaxLevels - 1]int // expected segment writes per level
	writeCount     [altFileHasherMaxLevels]int     // number of segment writes received by job buffer per level RENAME
	writeSyncCount int                             // number of external writes to the filehasher RENAME
	//doneC         [altFileHasherMaxLevels]chan struct{}    // used to tell parent that child is done writing on right edge
	resC chan []byte // used to tell hasher that all is done
	//wg    sync.WaitGroup                         // used to tell caller hashing is done (maybe be replced by channel, and doneC only internally)
	//lwg   [altFileHasherMaxLevels]sync.WaitGroup // used to block while the level's hasher is busy
	// TODO replace with rwlock
	lock       sync.Mutex // protect filehasher state vars
	hasherPool sync.Pool
}

func NewAltFileHasher(hasherFunc func() bmt.SectionWriter, segmentSize int, branches int) *AltFileHasher {
	f := &AltFileHasher{
		branches:      branches,
		segmentSize:   segmentSize,
		chunkSize:     branches * segmentSize,
		batchSegments: branches * branches,
		resC:          make(chan []byte),
	}
	for i := 0; i < altFileHasherMaxLevels-1; i++ {
		//f.buffers[i] = make([]byte, f.chunkSize*branches) // 4.6M with 9 levels
		//f.hashers[i] = hasherFunc()
		//f.doneC[i] = make(chan struct{}, 1)

		//	f.levelJobs[i] = make(chan fileHashJob, branches*2-1)
		f.levelWriteC[i] = make(chan []byte)
	}
	f.hasherPool.New = func() interface{} {
		return hasherFunc()
	}
	f.Reset()
	return f
}

// fileHashJob is submitted to level buffer channel when a chunk boundary is crossed on write
type fileHashJob struct {
	index  int                    // index this write belongs to
	c      int                    // write data cursor
	data   []byte                 // data from the write
	hasher chan bmt.SectionWriter // receives the next free hasher to process the data with
	sum    []byte                 // holds the hash result
	last   bool                   // true if this is the last write on the level
}

// enforces sequential parameters for the job descriptions to the level buffer channels
// the hasher is retrieved asynchronously so write can happen even if all hashers are busy
func (f *AltFileHasher) addJob(level int, data []byte, last bool) {
	j := fileHashJob{
		index:  f.getWriteCountSafe(level),
		data:   data,
		last:   last,
		hasher: make(chan bmt.SectionWriter, 1),
	}
	go func(hasher chan<- bmt.SectionWriter) {
		log.Debug("getting hasher", "level", level)
		j.hasher <- f.hasherPool.Get().(*bmt.AsyncHasher)
		log.Debug("got hasher", "level", level)
	}(j.hasher)
	log.Debug("new job", "leve", level, "last", last, "index", j.index)
	f.levelJobs[level] <- j
}

func (f *AltFileHasher) cancel(e error) {
	log.Error("cancel called TODO!")
}

// makes sure the hasher is clean before it's returned to the pool
func (f *AltFileHasher) putHasher(h bmt.SectionWriter) {
	h.Reset()
	f.hasherPool.Put(h)
}

// returns true if current write offset of level is on hashing boundary
func (f *AltFileHasher) isChunkBoundary(level int, wc int) bool {
	isboundary := wc%f.branches == 0
	log.Trace("check chunk boundary", "level", level, "wc", wc, "is", isboundary)
	return isboundary
}

func (f *AltFileHasher) getTotalBytesSafe() int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.totalBytes
}

// returns a level's write count
// holds the lock
func (f *AltFileHasher) getWriteCountSafe(level int) int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.writeCount[level]
}

// increments a level's write count
// holds the lock
func (f *AltFileHasher) incWriteCountSafe(level int) int {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.writeCount[level]++
	return f.writeCount[level]
}

func (f *AltFileHasher) isTopLevelSafe(level int) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return level == f.levelCount-1
}

// getPotentialSpan returns the total amount of data that can represented under the given level
// \TODO use a table instead
func (f *AltFileHasher) getPotentialSpan(level int) int {
	span := f.chunkSize
	for i := 0; i < level; i++ {
		span *= f.branches
	}
	return span
}

// makes the filehasher ready for new duty
// implements bmt.SectionWriter
func (f *AltFileHasher) Reset() {
	for i := 0; i < altFileHasherMaxLevels; i++ {
		if i > 0 {
			f.targetCount[i-1] = 0
		}
		f.levelJobs[i] = make(chan fileHashJob, branches*2-1)
		f.writeCount[i] = 0
		f.writeSyncCount = 0
	}
	f.totalBytes = 0
	f.levelCount = 0
	f.ctx = context.TODO()
	f.processJobs()
}

// check whether all writes on all levels have finished
// holds the lock
//func (f *AltFileHasher) isWriteFinishedSafe() bool {
//	f.lock.Lock()
//	defer f.lock.Unlock()
//	return f.finished
//}

// Finish marks the final write of the file
// It returns the root hash of the processed file
func (f *AltFileHasher) Finish(b []byte) []byte {
	f.lock.Lock()

	// if we call finish with additional data
	// include this data in the total length
	if b != nil {
		f.totalBytes += len(b)
	}

	// find our level height and decrease the waitgroup count to used levels only
	f.levelCount = getLevelsFromLength(f.totalBytes, f.segmentSize, f.branches)
	log.Debug("finish set", "levelcount", f.levelCount, "b", len(b))
	for i := altFileHasherMaxLevels; i > f.levelCount; i-- {
		log.Debug("purging unused level chans", "l", i)
		close(f.levelJobs[i-1])
	}

	// calculate the amount of write() calls expected in total
	// start with the amount of data writes (level 0)
	// add number of writes divided by 128 for every additional level
	// we don't use targetCount for level 0, since f.finished annotates that it is reached
	target := (f.totalBytes-1)/f.segmentSize + 1
	log.Debug("setting targetcount", "l", 0, "t", target)
	for i := 1; i < f.levelCount; i++ {
		target = (target-1)/f.branches + 1
		f.targetCount[i] = target
		log.Debug("setting targetcount", "l", i, "t", target)
	}

	f.lock.Unlock()

	// write and return result when we get it back
	if len(b) > 0 {
		f.write(0, f.writeSyncCount, b, false)
		f.writeSyncCount++
	} else {
		//
		//		// if the writecount of write bytecount does not end on a chunk boundary (number of segments in chunk)
		//		// we need to poke the job with a final write message
		//segmentWrites := (f.getTotalBytesSafe()-1)/f.segmentSize + 1
		if f.writeSyncCount%f.branches == 0 {
			log.Trace("write end chunk boundary align", "total", f.totalBytes, "segmentwrites", f.writeSyncCount)
			f.addJob(0, nil, true)
		}
		f.write(0, f.writeSyncCount, nil, true)
		//f.levelWriteC[0] <- nil
	}

	// get the result
	r := <-f.resC

	// free the rest of the level channels
	//	for i := 0; i < f.levelCount; i++ {
	//		log.Debug("purging done chans", "l", i)
	//		close(f.levelJobs[i])
	//	}

	//return the reult
	return r
}

// Write writes data provided from the buffer to the hasher
// \TODO currently not safe to write intermediate data of length not multiple of 32 bytes
func (f *AltFileHasher) Write(b []byte) {
	f.lock.Lock()
	f.totalBytes += len(b)
	f.lock.Unlock()
	for i := 0; i < len(b); i += 32 {
		f.write(0, f.writeSyncCount, b, false)
	}
	f.writeSyncCount++
}

// write signals the level channel handler that a new write has taken place
// it creates a new write job when write count hits chunk boundaries
// TODO pass writecount offset through function to avoid segmentwrite calculation
func (f *AltFileHasher) write(level int, offset int, b []byte, last bool) {
	log.Trace("write chunk boundary align", "offset", offset, "total", f.getTotalBytesSafe(), "level", level, "last", last, "datalength", len(b))
	if f.isChunkBoundary(level, offset) {
		f.addJob(level, b, last)
	}
	log.Debug("write levelwritec", "level", level, "last", last, "wc", offset)
	if len(b) > 0 {
		f.levelWriteC[level] <- b
	}
	if last {
		f.levelWriteC[level] <- nil
	}
}

// itarts one loop for every level that accepts hashing job
// propagates sequential writes up the levels
func (f *AltFileHasher) processJobs() {
	for i := 0; i < altFileHasherMaxLevels; i++ {
		go func(i int) {
			for {
				select {
				case j, ok := <-f.levelJobs[i]:
					if !ok {
						log.Trace("job channel closed", "i", i)
						return
					}
					if f.isTopLevelSafe(i) {
						dataPtr := <-f.levelWriteC[i]
						log.Debug("this is top level so all done", "i", i, "root", hexutil.Encode(dataPtr))
						close(f.levelJobs[i])
						f.resC <- dataPtr
						return
					}
					log.Debug("have job write", "level", i, "j", j)
					h := <-j.hasher
					var finished bool
					for !finished {
						var writeCount int
						var dataPtr []byte
						select {
						case dataPtr = <-f.levelWriteC[i]:
							writeCount = f.getWriteCountSafe(i)
							if len(dataPtr) == 0 {
								j.last = true
							}
							log.Trace("job write chan", "level", i, "data", dataPtr, "wc", writeCount, "last", j.last)
							if !j.last {
								netOffset := (writeCount % f.batchSegments)
								h.Write(netOffset%f.branches, dataPtr)
							}
							if len(dataPtr) > 0 {
								writeCount = f.incWriteCountSafe(i)
							}
						case <-f.ctx.Done():
							return
						}
						if (writeCount != 0 && f.isChunkBoundary(i, writeCount)) || j.last {
							log.Trace("chunk boundary|last", "last", j.last, "wc", writeCount, "level", i)
							f.doHash(h, i, &j)
							finished = true
						}

					}
				case <-f.ctx.Done():
					log.Debug("job exiting", "level", i, "err", f.ctx.Err())
					close(f.levelJobs[i])
					return
				}
			}
		}(i)
	}
}

// write writes the provided data directly to the underlying hasher
// and performs recursive hashing on complete batches or data end
// b is the data to write
// offset is the level's segment we are writing to
// level is the tree level we are writing to
// currentTotal is the current total of data bytes written so far
// TODO: ensure local copies of all thread unsafe vars
//func (f *AltFileHasher) write(b []byte, offset int, level int, currentTotal int) {
//
//	// copy state vars so we don't have to keep lock across the call
//	wc := f.getWriteCountSafe(level)
//	f.lock.Lock()
//	targetCount := f.targetCount[level]
//	f.lock.Unlock()
//
//	// only for log, delete on prod
//	if b == nil {
//		log.Debug("write", "level", level, "offset", offset, "length", "nil", "wc", wc, "total", currentTotal)
//	} else {
//		l := 32
//		if len(b) < l {
//			l = len(b)
//		}
//		log.Debug("write", "level", level, "offset", offset, "length", len(b), "wc", wc, "data", b[:l], "total", currentTotal)
//	}
//
//	// if top level then b is the root hash which means we are finished
//	// write it to the topmost buffer and release the waitgroup blocking  and then return
//	// \TODO should never be called when we refactor to separate hasher level buffer handler
//	if f.isTopLevelSafe(level) {
//		copy(f.buffers[level], b)
//		f.wg.Done()
//		log.Debug("top done", "level", level)
//		return
//	}
//
//	// only write if we have data
//	// b will never be nil except data level where it can be nil if no additional data is written upon the call to Finish()
//	// (else) if b is nil, and if the data is on a chunk boundary, the data will already have been hashed, which means we're done with that level
//	if len(b) > 0 {
//
//		// get the segment within the batch we are in
//		netOffset := (offset % f.batchSegments)
//
//		// write to the current level's hasher
//		f.hashers[level].Write(netOffset%f.branches, b)
//
//		// copy the data into the buffer
//		copy(f.buffers[level][netOffset*f.segmentSize:], b)
//
//		// increment the write count
//		wc = f.incWriteCountSafe(level)
//
//	} else if wc%f.branches == 0 {
//		f.wg.Done()
//		f.doneC[level] <- struct{}{}
//		return
//	}
//
//	// execute the hasher if:
//	// - we are on a chunk edge
//	// - we are on the data level and writes are set to finished
//	// - we are above data level, writes are finished, and expected level write count is reached
//	executeHasher := false
//	if wc%f.branches == 0 {
//		log.Debug("executehasher", "reason", "edge", "level", level, "offset", offset)
//		executeHasher = true
//	} else if f.finished && level == 0 {
//		log.Debug("executehasher", "reason", "data done", "level", level, "offset", offset)
//		executeHasher = true
//	} else if f.finished && targetCount > 0 && targetCount == wc {
//		<-f.doneC[level-1]
//		log.Debug("executehasher", "reason", "target done", "level", level, "offset", offset, "wc", wc)
//		executeHasher = true
//	}
//
//	// if this was a nil data finish instruction and we are on boundary, we may be still hashing asynchronously. Wait for it to finish
//	// if we are on boundary, no need to hash further
//	if f.finished && len(b) == 0 && level == 0 {
//		f.lwg[0].Wait()
//		log.Debug("finished and 0", "wc", wc)
//	}
//
//	if executeHasher {
//		f.doHash()
//	}
//}

// synchronous method that hashes the data contained in the job
// modifies fileHashJob in place
func (f *AltFileHasher) doHash(h bmt.SectionWriter, level int, j *fileHashJob) {

	// check for the dangling chunk
	offset := f.getWriteCountSafe(level)
	if level > 0 && j.last {
		writeCountBelow := f.getWriteCountSafe(level - 1)
		f.lock.Lock()
		log.Debug("danglecheck", "offset", offset, "f.batchSegments", f.batchSegments, "wc", writeCountBelow)
		childWrites := writeCountBelow % f.batchSegments
		if offset%f.branches == 0 && childWrites <= f.branches {
			log.Debug("dangle done", "level", level, "writeCount", j.c)
			f.lock.Unlock()
			f.write(level+1, offset, j.data, true)
			return
		}
		f.lock.Unlock()
	}

	// calculate what the potential span under this chunk will be
	span := f.getPotentialSpan(level)

	// calculate the actual data under this span
	// if data is fully written, the current chunk may be shorter than the span
	var dataUnderSpan int
	if j.last {
		dataUnderSpan = (f.getTotalBytesSafe()-1)%span + 1
	} else {
		dataUnderSpan = span
	}

	// calculate the length of the actual data in this chunk (the data to be hashed)
	var hashDataSize int
	if level == 0 {
		hashDataSize = dataUnderSpan
	} else {
		hashDataSize = ((dataUnderSpan-1)/(span/f.branches) + 1) * f.segmentSize
	}

	meta := make([]byte, 8)
	binary.LittleEndian.PutUint64(meta, uint64(dataUnderSpan))
	log.Debug("hash", "level", level, "size", hashDataSize, "meta", meta, "wc", j.c, "hasher", h, "gettotalbytes", f.getTotalBytesSafe(), "last", j.last, "span", span)

	j.sum = h.Sum(nil, hashDataSize, meta)

	// also write to output
	go func() {
		log.Trace("TODO write out to chunk", "sum", hexutil.Encode(j.sum), "data", hexutil.Encode(j.data))
	}()
	f.putHasher(h)

	// write to next level hasher

	// TODO here we are copying data bytes, can we get away with referencing underlying buffer?
	//go func(j *fileHashJob) {
	log.Trace("next level write", "level", level+1, "digest", hexutil.Encode(j.sum))

	parentOffset := (offset - 1) / f.branches
	f.write(level+1, parentOffset, j.sum, j.last)

	// close this job channel if this is the last write
	if j.last {
		log.Trace("dohash last close chan", "level", level)
		close(f.levelJobs[level])
	}
	//}(j)
}

//func (f *AltFileHasher) doHash_() {
//	// if we are still hashing the data for this level, wait until we are done
//	f.lwg[level].Wait()
//
//	// check for the dangling chunk
//	if level > 0 && f.finished {
//		cwc := f.getWriteCountSafe(level - 1)
//
//		f.lock.Lock()
//		log.Debug("danglecheck", "offset", offset, "f.batchSegments", f.batchSegments, "cwc", cwc)
//		childWrites := cwc % f.batchSegments
//		if offset%f.branches == 0 && childWrites <= f.branches {
//			log.Debug("dangle done", "level", level, "wc", wc)
//			parentOffset := (wc - 1) / f.branches
//			f.lock.Unlock()
//			f.wg.Done()
//			f.doneC[level] <- struct{}{}
//			f.write(b, parentOffset, level+1, currentTotal)
//			return
//		}
//		f.lock.Unlock()
//	}
//
//	f.lwg[level].Add(1)
//
//	// calculate what the potential span under this chunk will be
//	span := f.getPotentialSpan(level)
//
//	// calculate the actual data under this span
//	// if data is fully written, the current chunk may be shorter than the span
//	var dataUnderSpan int
//	if f.isWriteFinishedSafe() {
//		dataUnderSpan = (currentTotal-1)%span + 1
//	} else {
//		dataUnderSpan = span
//	}
//
//	// calculate the length of the actual data in this chunk (the data to be hashed)
//	var hashDataSize int
//	if level == 0 {
//		hashDataSize = dataUnderSpan
//	} else {
//		hashDataSize = ((dataUnderSpan-1)/(span/f.branches) + 1) * f.segmentSize
//	}
//
//	meta := make([]byte, 8)
//	binary.LittleEndian.PutUint64(meta, uint64(dataUnderSpan))
//	log.Debug("hash", "level", level, "size", hashDataSize, "meta", meta, "wc", wc)
//	hashResult := f.hashers[level].Sum(nil, hashDataSize, meta)
//	f.hashers[level].Reset()
//
//	// hash the chunk and write it to the current cursor position on the next level
//	go func(level int, wc int, finished bool, currentTotal int, targetCount int) {
//		// if the hasher on the level above is still working, wait for it
//		f.lwg[level+1].Wait()
//		log.Debug("gofunc hash up", "level", level, "wc", wc)
//		parentOffset := (wc - 1) / f.branches
//		if (level == 0 && finished) || targetCount == wc {
//			log.Debug("done", "level", level)
//			f.wg.Done()
//			f.doneC[level] <- struct{}{}
//		}
//		f.write(hashResult, parentOffset, level+1, currentTotal)
//		f.lwg[level].Done()
//	}(level, wc, f.finished, currentTotal, targetCount)
//}
