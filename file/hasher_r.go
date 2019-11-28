package file

import (
	"io"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
)

// ReferenceFileHasher is a non-performant source of truth implementation for the file hashing algorithm used in Swarm
// the aim of its design is that is should be easy to understand
// TODO: bmt.Hasher should instead be passed as hash.Hash and ResetWithLength() should be abolished
type ReferenceFileHasher struct {
	params         *treeParams
	hasher         *bmt.Hasher // synchronous hasher
	chunkSize      int         // cached chunk size, equals branches * sectionSize
	buffer         []byte      // keeps intermediate chunks during hashing
	cursors        []int       // write cursors in sectionSize units for each tree level
	totalBytes     int         // total data bytes to be written
	totalLevel     int         // total number of levels in tree. (level 0 is the data level)
	writeByteCount int         // amount of bytes currently written
	writeCount     int         // amount of sections currently written
}

// NewReferenceFileHasher creates a new file hasher with the supplied branch factor
// the section count will be the Size() of the hasher
func NewReferenceFileHasher(hasher *bmt.Hasher, branches int) *ReferenceFileHasher {
	f := &ReferenceFileHasher{
		params:    newTreeParams(hasher.Size(), branches, nil),
		hasher:    hasher,
		chunkSize: branches * hasher.Size(),
	}
	return f
}

// Hash executes l reads of up to sectionSize bytes from r
// and performs the filehashing algorithm on the data
// it returns the root hash
func (f *ReferenceFileHasher) Hash(r io.Reader, l int) []byte {

	f.totalBytes = l
	f.totalLevel = getLevelsFromLength(l, f.params.SectionSize, f.params.Branches) + 1
	log.Trace("Starting reference file hasher", "levels", f.totalLevel, "length", f.totalBytes, "b", f.params.Branches, "s", f.params.SectionSize)

	// prepare a buffer for intermediate the chunks
	bufLen := f.params.SectionSize
	for i := 1; i < f.totalLevel; i++ {
		bufLen *= f.params.Branches
	}
	f.buffer = make([]byte, bufLen)
	f.cursors = make([]int, f.totalLevel)

	var res bool
	for !res {

		// read a data section into input copy buffer
		input := make([]byte, f.params.SectionSize)
		c, err := r.Read(input)
		log.Trace("read", "bytes", c, "total read", f.writeByteCount)
		if err != nil {
			if err == io.EOF {
				panic("EOF")
			} else {
				panic(err)
			}
		}

		// read only up to the announced length, since we dimensioned buffer and level count accordingly
		readSize := f.params.SectionSize
		remainingBytes := f.totalBytes - f.writeByteCount
		if remainingBytes <= f.params.SectionSize {
			readSize = remainingBytes
			input = input[:remainingBytes]
			res = true
		}
		f.writeByteCount += readSize
		f.write(input, 0, res)
	}
	if f.cursors[f.totalLevel-1] != 0 {
		panic("totallevel cursor misaligned")
	}
	return f.buffer[0:f.params.SectionSize]
}

// performs recursive hashing on complete batches or data end
func (f *ReferenceFileHasher) write(b []byte, level int, end bool) bool {

	log.Trace("write", "level", level, "bytes", len(b), "total written", f.writeByteCount, "end", end, "data", hexutil.Encode(b))

	// copy data from input copy buffer to current position of corresponding level in intermediate chunk buffer
	copy(f.buffer[f.cursors[level]*f.params.SectionSize:], b)
	for i, l := range f.cursors {
		log.Trace("cursor", "level", i, "position", l)
	}

	// if we are at the tree root the result will be in the first sectionSize bytes of the buffer.
	// the true bool return will bubble up to the data write frame in the call stack and terminate the loop
	//if level == len(f.cursors)-1 {
	if level == f.totalLevel-1 {
		return true
	}

	// if we are at the end of the write, AND
	// if the offset of a chunk reference is the same one level up, THEN
	// we have a "dangling chunk" and we merely pass it to the next level
	if end && level > 0 && f.cursors[level] == f.cursors[level+1] {
		res := f.write(b, level+1, end)
		return res
	}

	// we've written to the buffer a particular level
	// so we increment the cursor of that level
	f.cursors[level]++

	// hash the intermediate chunk buffer data for this level if:
	// - the difference of cursors between this level and the one above equals the branch factor (equals one full chunk of data)
	// - end is set
	// the resulting digest will be written to the corresponding section of the level above
	var res bool
	if f.cursors[level]-f.cursors[level+1] == f.params.Branches || end {

		// calculate the actual data under this span
		// if we're at end, the span is given by the period of the potential span
		// if not, it will be the full span (since we then must have full chunk writes in the levels below)
		var dataUnderSpan int
		span := f.params.Spans[level] * chunkSize
		if end {
			dataUnderSpan = (f.totalBytes-1)%span + 1
		} else {
			dataUnderSpan = span
		}

		// calculate the data in this chunk (the data to be hashed)
		// on level 0 it is merely the actual spanned data
		// on levels above data level, we get number of sections the data equals, and divide by the level span
		var hashDataSize int
		if level == 0 {
			hashDataSize = dataUnderSpan
		} else {
			dataSectionCount := dataSizeToSectionCount(dataUnderSpan, f.params.SectionSize)
			// TODO: this is the same as dataSectionToLevelSection, but without wrap to 0 on end boundary. Inspect whether the function should be amended, and necessary changes made to Hasher
			levelSectionCount := (dataSectionCount-1)/f.params.Spans[level] + 1
			hashDataSize = levelSectionCount * f.params.SectionSize
		}

		// prepare the hasher,
		// write data since previous hash operation from the current level cursor position
		// and sum
		spanBytes := lengthToSpan(dataUnderSpan)
		f.hasher.ResetWithLength(spanBytes)
		hasherWriteOffset := f.cursors[level+1] * f.params.SectionSize
		f.hasher.Write(f.buffer[hasherWriteOffset : hasherWriteOffset+hashDataSize])
		hashResult := f.hasher.Sum(nil)
		log.Debug("summed", "level", level, "cursor", f.cursors[level], "parent cursor", f.cursors[level+1], "span", spanBytes, "digest", hexutil.Encode(hashResult))

		// write the digest to the current cursor position of the next level
		// note the f.write() call will move the next level's cursor according to the write and possible hash operation
		res = f.write(hashResult, level+1, end)

		// recycle buffer space from the threshold of just written hash
		f.cursors[level] = f.cursors[level+1]
	}
	return res
}
