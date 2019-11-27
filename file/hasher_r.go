package file

import (
	"encoding/binary"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
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
}

func NewReferenceFileHasher(hasher *bmt.Hasher, branches int) *ReferenceFileHasher {
	return &ReferenceFileHasher{
		hasher:      hasher,
		branches:    branches,
		segmentSize: hasher.Size(),
		chunkSize:   branches * hasher.Size(),
	}
}

// reads segmentwise from input data and writes
// TODO: Write directly to f.buffer instead of input
// TODO: See if level 0 data can be written directly to hasher without complicating code
func (f *ReferenceFileHasher) Hash(r io.Reader, l int) common.Hash {
	f.totalBytes = l
	// TODO: old implementation of function skewed the level by 1, realign code to new, correct results
	levelCount := getLevelsFromLength(l, f.segmentSize, f.branches) + 1
	log.Trace("level count", "l", levelCount, "b", f.branches, "c", l, "s", f.segmentSize)
	bufLen := f.segmentSize
	for i := 1; i < levelCount; i++ {
		bufLen *= f.branches
	}
	f.cursors = make([]int, levelCount)
	f.buffer = make([]byte, bufLen)
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
		f.write(input, 0, res)
	}
	return common.BytesToHash(f.buffer[f.cursors[levelCount-1] : f.cursors[levelCount-1]+f.segmentSize])
}

// TODO: check if length 0
// performs recursive hashing on complete batches or data end
func (f *ReferenceFileHasher) write(b []byte, level int, end bool) bool {
	log.Debug("write", "l", level, "len", len(b), "b", hexutil.Encode(b), "end", end, "wbc", f.writeByteCount)

	// copy data from buffer to current position of corresponding level in buffer
	copy(f.buffer[f.cursors[level]*f.segmentSize:], b)
	for i, l := range f.cursors {
		log.Trace("cursor", "#", i, "pos", l)
	}

	// if we are at the tree root the result will be in the first segmentSize bytes of the buffer. Return
	if level == len(f.cursors)-1 {
		return true
	}

	// if the offset is the same one level up, then we have a dangling chunk and we merely pass it down the tree
	if end && level > 0 && f.cursors[level] == f.cursors[level+1] {
		res := f.write(b, level+1, end)
		return res
	}

	// we've written to the buffer of this level, so we increment the cursor
	f.cursors[level]++

	// perform recursive writes down the tree if end of output or on batch boundary
	var res bool
	if f.cursors[level]-f.cursors[level+1] == f.branches || end {

		// calculate what the potential span under this chunk will be
		span := f.chunkSize
		for i := 0; i < level; i++ {
			span *= f.branches
		}

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

		// hash the chunk and write it to the current cursor position on the next level
		meta := make([]byte, 8)
		binary.LittleEndian.PutUint64(meta, uint64(dataUnderSpan))
		f.hasher.ResetWithLength(meta)
		writeHashOffset := f.cursors[level+1] * f.segmentSize
		f.hasher.Write(f.buffer[writeHashOffset : writeHashOffset+hashDataSize])
		hashResult := f.hasher.Sum(nil)
		log.Debug("summed", "b", hexutil.Encode(hashResult), "l", f.cursors[level], "l+1", f.cursors[level+1], "spanlength", dataUnderSpan, "span", span, "meta", meta, "from", writeHashOffset, "to", writeHashOffset+hashDataSize, "data", f.buffer[writeHashOffset:writeHashOffset+hashDataSize])
		res = f.write(hashResult, level+1, end)

		// recycle buffer space from the threshold of just written hash
		f.cursors[level] = f.cursors[level+1]
	}
	return res
}
