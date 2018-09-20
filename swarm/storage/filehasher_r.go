package storage

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/bmt"
)

type ReferenceFileHasher struct {
	hasher         *bmt.Hasher
	branches       int
	segmentSize    int
	buffer         []byte
	cursors        []int
	chunkSize      int
	totalBytes     int
	writeByteCount int
	writeCount     int
	swap           []byte
}

func NewReferenceFileHasher(hasher *bmt.Hasher, branches int) *ReferenceFileHasher {
	return &ReferenceFileHasher{
		hasher:      hasher,
		branches:    branches,
		segmentSize: hasher.Size(),
		chunkSize:   branches * hasher.Size(),
	}
}

func getLevelsFromLength(l int, segmentSize int, branches int) int {
	if l == 0 {
		return 0
	} else if l <= segmentSize*branches {
		return 2
	}
	c := (l - 1) / (segmentSize)

	return int(math.Log(float64(c))/math.Log(float64(branches)) + 2)
}

func (f *ReferenceFileHasher) Hash(r io.Reader, l int) common.Hash {
	f.totalBytes = l
	levelCount := getLevelsFromLength(l, f.segmentSize, f.branches)
	log.Debug("level count", "l", levelCount, "b", f.branches, "c", l, "s", f.segmentSize)
	bufLen := f.segmentSize
	for i := 1; i < levelCount; i++ {
		bufLen *= f.branches
	}
	f.cursors = make([]int, levelCount)
	f.buffer = make([]byte, bufLen)
	f.swap = make([]byte, f.segmentSize)
	var res bool
	for !res {
		input := make([]byte, f.segmentSize)
		c, err := r.Read(input)
		log.Trace("read", "c", c, "wbc", f.writeByteCount)
		if err != nil {
			if err == io.EOF {
				log.Debug("haveeof")
				res = true
			} else {
				panic(err)
			}
		} else if c < f.segmentSize {
			input = input[:c]
		}
		f.writeByteCount += c
		if f.writeByteCount == f.totalBytes {
			res = true
		}
		res = f.write(input, 0, res)
	}
	return common.BytesToHash(f.buffer[f.cursors[levelCount-1] : f.cursors[levelCount-1]+f.segmentSize])
}

// TODO: check length 0
func (f *ReferenceFileHasher) write(b []byte, level int, end bool) bool {
	log.Trace("write", "l", level, "len", len(b), "b", b, "end", end, "wbc", f.writeByteCount)

	// copy data from buffer to current position of corresponding level in buffer
	copy(f.buffer[f.cursors[level]*f.segmentSize:], b)
	for i, l := range f.cursors {
		log.Debug("cursor", "#", i, "pos", l)
	}

	// if we are at the tree root the result will be in the first segmentSize bytes of the buffer. Return
	if level == len(f.cursors)-1 {
		return true
	}

	if end && level > 0 && f.cursors[level] == f.cursors[level+1] {
		res := f.write(b, level+1, end)
		return res
	}
	// increment the position of this level in buffer
	f.cursors[level]++

	// perform recursive writes down the tree if end of output or on batch boundary
	var res bool
	if f.cursors[level]-f.cursors[level+1] == f.branches || end {
		if f.cursors[level] == f.cursors[level+1] && f.cursors[level] > 0 {
			log.Debug("short return in write")
			return true
		}

		// calculate what the potential span under this chunk will be
		span := f.chunkSize
		for i := 0; i < level; i++ {
			span *= f.branches
		}

		// if we have a dangling chunk, simply pass it up
		// calculate the data in this chunk (the data to be hashed)
		var dataUnderSpan int
		if end {
			dataUnderSpan = (f.totalBytes-1)%span + 1
		} else {
			dataUnderSpan = span
		}

		// calculate the actual data under this span
		var hashDataSize int
		if level == 0 {
			hashDataSize = dataUnderSpan
		} else {
			hashDataSize = ((dataUnderSpan-1)/(span/f.branches) + 1) * f.segmentSize
		}

		meta := make([]byte, 8)
		binary.LittleEndian.PutUint64(meta, uint64(dataUnderSpan))
		f.hasher.ResetWithLength(meta)
		writeHashOffset := f.cursors[level+1] * f.segmentSize
		f.hasher.Write(f.buffer[writeHashOffset : writeHashOffset+hashDataSize])
		copy(f.swap, f.hasher.Sum(nil))
		log.Debug("summed", "b", f.swap, "l", f.cursors[level], "l+1", f.cursors[level+1], "spanlength", dataUnderSpan, "span", span, "meta", meta, "from", writeHashOffset, "to", writeHashOffset+hashDataSize, "data", f.buffer[writeHashOffset:writeHashOffset+hashDataSize])
		res = f.write(f.swap, level+1, end)
		f.cursors[level] = f.cursors[level+1]
	}
	return res
}
