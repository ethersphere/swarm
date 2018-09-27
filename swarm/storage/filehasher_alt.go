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
	branches    int
	segmentSize int
	hashers     [altFileHasherMaxLevels]bmt.SectionWriter
	buffers     [altFileHasherMaxLevels - 1][]byte
	levelCount  int
	chunkSize   int
	finished    bool
	totalBytes  int
	targetCount [altFileHasherMaxLevels - 1]int
	writeCount  [altFileHasherMaxLevels]int
	doneC       [altFileHasherMaxLevels]chan struct{}
	wg          sync.WaitGroup                         // used when level done
	lwg         [altFileHasherMaxLevels]sync.WaitGroup // used when busy hashing
	lock        sync.Mutex
}

func NewAltFileHasher(hasherFunc func() bmt.SectionWriter, segmentSize int, branches int) *AltFileHasher {
	f := &AltFileHasher{
		branches:    branches,
		segmentSize: segmentSize,
		chunkSize:   branches * segmentSize,
	}
	for i := 0; i < altFileHasherMaxLevels-1; i++ {
		f.buffers[i] = make([]byte, f.chunkSize)
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
	if b != nil {
		f.totalBytes += len(b)
	}
	f.finished = true

	// find our level height and release the unused levels
	f.levelCount = getLevelsFromLength(f.totalBytes, f.segmentSize, f.branches)

	log.Debug("finish set", "levelcount", f.levelCount)
	for i := altFileHasherMaxLevels; i > f.levelCount; i-- {
		log.Debug("purging unused level wg", "l", i)
		f.wg.Done()
	}

	// calculate the amount of writes expected on each level
	target := (f.totalBytes-1)/f.segmentSize + 1
	for i := 1; i < f.levelCount; i++ {
		target = (target-1)/f.branches + 1
		f.targetCount[i] = target
		log.Debug("setting targetcount", "l", i, "t", target)
	}

	// write and return result when we get it back
	f.write(b, f.writeCount[0], 0)
	f.wg.Wait()
	return f.buffers[f.levelCount-1][:f.segmentSize]
}

func (f *AltFileHasher) Write(b []byte) {
	f.totalBytes += len(b)
	f.write(b, f.writeCount[0], 0)
}

func (f *AltFileHasher) getPotentialSpan(level int) int {
	span := f.chunkSize
	for i := 0; i < level; i++ {
		span *= f.branches
	}
	return span
}

// TODO: ensure local copies of all thread unsafe vars
// performs recursive hashing on complete batches or data end
func (f *AltFileHasher) write(b []byte, offset int, level int) {

	if b == nil {
		log.Debug("write", "level", level, "offset", offset, "length", "nil")
	} else {
		l := 32
		if len(b) < l {
			l = len(b)
		}
		log.Debug("write", "level", level, "offset", offset, "length", len(b), "data", b[:l])
	}

	// top level then return
	if level == f.levelCount-1 {
		copy(f.buffers[level], b)
		f.lock.Lock()
		f.wg.Done()
		f.lock.Unlock()
		log.Debug("top done", "level", level)
		return
	}

	// thread safe writecount
	// b will never be nil except bottom level, which will have already been hashed if on chunk boundary
	f.lock.Lock()
	wc := f.writeCount[level]
	f.lock.Unlock()

	// only write if we have data
	// data might be nil when upon write finish
	if b != nil {
		f.hashers[level].Write(offset%f.branches, b)
		f.lock.Lock()
		f.writeCount[level]++
		wc = f.writeCount[level]
		f.lock.Unlock()
	} else if wc%f.branches == 0 {
		f.lock.Lock()
		f.wg.Done()
		f.lock.Unlock()
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
	} else if f.finished && f.targetCount[level] > 0 && f.targetCount[level] == wc {
		<-f.doneC[level-1]
		log.Debug("executehasher", "reason", "target done", "level", level, "offset", offset)
		executeHasher = true
	}

	if executeHasher {

		// check for the dangling chunk
		if level > 0 && f.finished {
			f.lock.Lock()
			cwc := f.writeCount[level-1]
			f.lock.Unlock()
			if offset%f.branches == 0 && cwc%(f.branches*f.branches) < f.branches {
				log.Debug("dangle", "level", level)
				parentOffset := (wc - 1) / f.branches
				f.write(b, parentOffset, level+1)
				f.lock.Lock()
				f.wg.Done()
				f.lock.Unlock()
				f.doneC[level] <- struct{}{}
				return
			}
		}

		f.lock.Lock()
		f.lwg[level].Add(1)
		f.lock.Unlock()

		// calculate what the potential span under this chunk will be
		span := f.getPotentialSpan(level)

		// calculate the actual data under this span
		// if data is fully written, the current chunk may be shorter than the span
		var dataUnderSpan int
		if f.isWriteFinished() {
			dataUnderSpan = (f.totalBytes-1)%span + 1
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
		go func(level int, wc int, finished bool) {
			// if the hasher on the level about is still working, wait for it
			f.lwg[level+1].Wait()

			parentOffset := (wc - 1) / f.branches
			if (level == 0 && finished) || f.targetCount[level] == wc {
				log.Debug("done", "level", level)
				f.lock.Lock()
				f.wg.Done()
				f.lock.Unlock()
				f.doneC[level] <- struct{}{}
			}
			f.write(hashResult, parentOffset, level+1)
			f.lock.Lock()
			f.lwg[level].Done()
			f.lock.Unlock()
		}(level, wc, f.finished)
	}
}

func (f *AltFileHasher) wgDoneFunc(level int, prune bool) func() {
	log.Warn("done", "level", level, "prune", prune)
	return func() {
		f.lock.Lock()
		f.wg.Done()
		f.lock.Unlock()
	}
}
