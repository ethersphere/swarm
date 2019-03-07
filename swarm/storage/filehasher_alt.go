package storage

import (
	"encoding/binary"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/bmt"
)

const (
	altFileHasherMaxLevels = 9 // 22 zetabytes should be enough for anyone
)

type AltFileHasher struct {
	branches      int
	segmentSize   int
	chunkSize     int
	batchSegments int
	hashers       [altFileHasherMaxLevels]bmt.SectionWriter
	buffers       [altFileHasherMaxLevels][]byte         // holds chunk data on each level (todo; push data to channel on complete). Buffers can hold one batch of data
	levelCount    int                                    // number of levels in this job (only determined when Finish() is called
	finished      bool                                   // finished writing data
	totalBytes    int                                    // total data bytes written
	targetCount   [altFileHasherMaxLevels - 1]int        // expected segment writes per level
	writeCount    [altFileHasherMaxLevels]int            // number of segment writes per level
	doneC         [altFileHasherMaxLevels]chan struct{}  // used to tell parent that child is done writing on right edge
	wg            sync.WaitGroup                         // used to tell caller hashing is done (maybe be replced by channel, and doneC only internally)
	lwg           [altFileHasherMaxLevels]sync.WaitGroup // used when busy hashing
	lock          sync.Mutex                             // protect filehasher state vars
}

func NewAltFileHasher(hasherFunc func() bmt.SectionWriter, segmentSize int, branches int) *AltFileHasher {
	f := &AltFileHasher{
		branches:      branches,
		segmentSize:   segmentSize,
		chunkSize:     branches * segmentSize,
		batchSegments: branches * branches,
	}
	for i := 0; i < altFileHasherMaxLevels-1; i++ {
		f.buffers[i] = make([]byte, f.chunkSize*branches) // 4.6M with 9 levels
		f.hashers[i] = hasherFunc()
		f.doneC[i] = make(chan struct{}, 1)
	}
	f.Reset()
	return f
}

func (f *AltFileHasher) Reset() {
	f.totalBytes = 0
	f.levelCount = 0
	f.wg.Add(altFileHasherMaxLevels)
	for i := 0; i < altFileHasherMaxLevels; i++ {
		if i > 0 {
			f.targetCount[i-1] = 0
		}
		f.writeCount[i] = 0
	}
}

func (f *AltFileHasher) isWriteFinished() bool {
	var finished bool
	f.lock.Lock()
	finished = f.finished
	f.lock.Unlock()
	return finished
}

func (f *AltFileHasher) Finish(b []byte) []byte {
	f.lock.Lock()

	// if we call finish with additional data
	// include this data in the total length
	if b != nil {
		f.totalBytes += len(b)
	}
	f.finished = true

	// find our level height and decrease the waitgroup count to used levels only
	f.levelCount = getLevelsFromLength(f.totalBytes, f.segmentSize, f.branches)
	log.Debug("finish set", "levelcount", f.levelCount, "b", len(b))
	for i := altFileHasherMaxLevels; i > f.levelCount; i-- {
		log.Debug("purging unused level wg", "l", i)
		f.wg.Done()
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
	f.write(b, f.writeCount[0], 0, f.totalBytes)
	f.wg.Wait()
	return f.buffers[f.levelCount-1][:f.segmentSize]
}

// Write writes data provided from the buffer to the hasher
// \TODO currently not safe to write intermediate data of length not multiple of 32 bytes
func (f *AltFileHasher) Write(b []byte) {
	f.totalBytes += len(b)
	for i := 0; i < len(b); i += 32 {
		f.write(b[i:], f.writeCount[0], 0, f.totalBytes)
	}
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

// write writes the provided data directly to the underlying hasher
// and performs recursive hashing on complete batches or data end
// b is the data to write
// offset is the level's segment we are writing to
// level is the tree level we are writing to
// currentTotal is the current total of data bytes written so far
// TODO: ensure local copies of all thread unsafe vars
//func (f *AltFileHasher) write(b []byte, offset int, level int) {
func (f *AltFileHasher) write(b []byte, offset int, level int, currentTotal int) {

	// copy state vars so we don't have to keep lock across the call
	f.lock.Lock()
	wc := f.writeCount[level]
	//currentTotal := f.totalBytes
	targetCount := f.targetCount[level]
	f.lock.Unlock()

	// only for log, delete on prod
	if b == nil {
		log.Debug("write", "level", level, "offset", offset, "length", "nil", "wc", wc, "total", currentTotal)
	} else {
		l := 32
		if len(b) < l {
			l = len(b)
		}
		log.Debug("write", "level", level, "offset", offset, "length", len(b), "wc", wc, "data", b[:l], "total", currentTotal)
	}

	// if top level then b is the root hash which means we are finished
	// write it to the topmost buffer and release the waitgroup blocking  and then return
	f.lock.Lock()
	if level == f.levelCount-1 {
		copy(f.buffers[level], b)
		f.lock.Unlock()
		f.wg.Done()
		log.Debug("top done", "level", level)
		return
	}
	f.lock.Unlock()

	// only write if we have data
	// b will never be nil except data level where it can be nil if no additional data is written upon the call to Finish()
	// (else) if b is nil, and if the data is on a chunk boundary, the data will already have been hashed, which means we're done with that level
	if len(b) > 0 {

		// get the segment within the batch we are in
		netOffset := (offset % f.batchSegments)

		// write to the current level's hasher
		f.hashers[level].Write(netOffset%f.branches, b)

		// copy the data into the buffer
		// TODO do we need this on the data level? should this be pipe write to something else?
		copy(f.buffers[level][netOffset*f.segmentSize:], b)

		// increment the write count
		f.lock.Lock()
		f.writeCount[level]++
		wc = f.writeCount[level]
		f.lock.Unlock()

	} else if wc%f.branches == 0 {
		f.wg.Done()
		f.doneC[level] <- struct{}{}
		return
	}

	// execute the hasher if:
	// - we are on a chunk edge
	// - we are on the data level and writes are set to finished
	// - we are above data level, writes are finished, and expected level write count is reached
	executeHasher := false
	if wc%f.branches == 0 {
		log.Debug("executehasher", "reason", "edge", "level", level, "offset", offset)
		executeHasher = true
	} else if f.finished && level == 0 {
		log.Debug("executehasher", "reason", "data done", "level", level, "offset", offset)
		executeHasher = true
	} else if f.finished && targetCount > 0 && targetCount == wc {
		<-f.doneC[level-1]
		log.Debug("executehasher", "reason", "target done", "level", level, "offset", offset, "wc", wc)
		executeHasher = true
	}

	// if this was a nil data finish instruction and we are on boundary, we may be still hashing asynchronously. Wait for it to finish
	// if we are on boundary, no need to hash further
	if f.finished && len(b) == 0 && level == 0 {
		f.lwg[0].Wait()
		log.Debug("finished and 0", "wc", wc)
	}

	if executeHasher {
		f.lwg[level].Wait()
		// check for the dangling chunk
		if level > 0 && f.finished {
			f.lock.Lock()
			cwc := f.writeCount[level-1]

			log.Debug("danglecheck", "offset", offset, "f.batchSegments", f.batchSegments, "cwc", cwc)
			// TODO: verify why do we need the latter part again?
			childWrites := cwc % f.batchSegments
			//if offset%f.batchSegments == 0 && childWrites < f.branches {
			//if offset%f.branches == 0 && childWrites < f.branches && childWrites > 0 {
			if offset%f.branches == 0 && childWrites <= f.branches {
				//		f.lwg[level+1].Wait()
				log.Debug("dangle done", "level", level, "wc", wc)
				parentOffset := (wc - 1) / f.branches
				f.lock.Unlock()
				f.wg.Done()
				f.doneC[level] <- struct{}{}
				//f.write(b, parentOffset, level+1)
				f.write(b, parentOffset, level+1, currentTotal)
				return
			}
			f.lock.Unlock()
		}

		f.lwg[level].Add(1)

		// calculate what the potential span under this chunk will be
		span := f.getPotentialSpan(level)

		// calculate the actual data under this span
		// if data is fully written, the current chunk may be shorter than the span
		var dataUnderSpan int
		if f.isWriteFinished() {
			dataUnderSpan = (currentTotal-1)%span + 1
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

		// hash the chunk and write it to the current cursor position on the next level
		meta := make([]byte, 8)
		binary.LittleEndian.PutUint64(meta, uint64(dataUnderSpan))
		log.Debug("hash", "level", level, "size", hashDataSize, "meta", meta, "wc", wc)
		hashResult := f.hashers[level].Sum(nil, hashDataSize, meta)
		f.hashers[level].Reset()
		go func(level int, wc int, finished bool, total int, targetCount int) {
			// if the hasher on the level above is still working, wait for it
			f.lwg[level+1].Wait()
			log.Debug("gofunc hash up", "level", level, "wc", wc)
			parentOffset := (wc - 1) / f.branches
			if (level == 0 && finished) || targetCount == wc {
				log.Debug("done", "level", level)
				f.wg.Done()
				f.doneC[level] <- struct{}{}
			}
			//f.write(hashResult, parentOffset, level+1)
			f.write(hashResult, parentOffset, level+1, total)
			f.lwg[level].Done()
		}(level, wc, f.finished, currentTotal, targetCount)
	}
}
