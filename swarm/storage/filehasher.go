// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"encoding/binary"
	"sync"
	"sync/atomic"
)

// SectionHasher is an asynchronous writer interface to a hash
// it allows for concurrent and out-of-order writes of sections of the hash's input buffer
// Sum can be called once the final length is known potentially before all sections are complete
type SectionHasher interface {
	Reset()
	Write(idx int, section []byte)
	Size() int
	BlockSize() int
	ChunkSize() int
	Sum(b []byte, length int, meta []byte) []byte
}

// FileHasher is instantiated each time a file is swarm hashed
// itself implements the ChunkHasher interface
type FileHasher struct {
	mtx        sync.Mutex           // RW lock to add/read levels push and unshift batches
	pool       sync.Pool            // batch resource pool
	levels     []*level             // levels of the swarm hash tree
	secsize    int                  // section size
	branches   int                  // branching factor
	hasherFunc func() SectionHasher // hasher constructor
	result     chan []byte          // channel to put hash asynchronously
	size       int
}

func New(hasherFunc func() SectionHasher, branches int) *FileHasher {
	sh := &FileHasher{
		hasherFunc: hasherFunc,
		result:     make(chan []byte),
	}
	sh.pool = sync.Pool{
		New: func() interface{} {
			return sh.newBatch()
		},
	}
	sh.size = hasherFunc().Size()
	return sh
}

// level captures one level of chunks in the swarm hash tree
// singletons are attached to the lowest level
type level struct {
	levelIndex int // which level of the swarm hash tree
	//batches     []*batch // active batches on the level
	batches     sync.Map
	*FileHasher // pointer to the underlying hasher
}

// batch records chunks subsumed under the same parent intermediate chunk
type batch struct {
	nodes  []*node // nodes of the batches
	index  int     // offset of the node
	parent *node   // pointer to containing
	buffer *bytes.Buffer
	*level // pointer to containing level
}

// node represent a chunk and embeds an async interface to the chunk hash used
type node struct {
	hasher SectionHasher // async hasher
	pos    int           // index of the node chunk within its batch
	secCnt int32         // number of sections written
	*batch               // pointer to containing batch
}

// getParentLevel retrieves or creates the next level up from a node/batch/level
// using lock for concurrent access
func (lev *level) getLevel(pl int) (par *level) {
	if pl < len(lev.levels) {
		return lev.levels[pl]
	}
	par = &level{
		levelIndex: pl,
	}
	lev.levels = append(lev.levels, par)
	return par
}

func (lev *level) getBatch(index int) *batch {
	pbi, ok := lev.batches.Load(index)
	if !ok {
		return nil
	}
	return pbi.(*batch)
}

// retrieve the batch within a level corresponding to the given index
// if it does not currently exist, create it
func (lev *level) getOrCreateBatch(index int) *batch {
	pb := lev.getBatch(index)
	if pb == nil {
		pb = lev.pool.Get().(*batch)
		lev.batches.Store(index, pb)
	}
	return pb
}

// delink unshifts the levels batches
// and releases the popped batch to the batch pools
// must be called after Sum has returned
// section writes or children no longer reference this batch
func (b *batch) delink() {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.batches.Delete(b.index)
	b.pool.Put(b)
}

// returns the digest size of the underlying hasher
func (fh *FileHasher) Size() int {
	return fh.size
}

// newBatch constructs a reuseable batch
func (sh *FileHasher) newBatch() (bt *batch) {
	nodes := make([]*node, sh.branches)
	chunkSize := sh.ChunkSize()
	bt = &batch{
		buffer: make([]byte, sh.branches*chunkSize),
		//buffer: bytes.NewBuffer(make([]byte, 0, sh.branches*sh.ChunkSize())),
	}
	for i := range nodes {
		offset := chunkSize * i
		nodes[i] = &node{
			pos:    i,
			hasher: sh.hasherFunc(),
			buffer: batch[offset : offset+chunkSize],
		}
	}
	batch.nodes = nodes
	return bt
}

func (sh *FileHasher) getNodeSectionBuffer(globalCount int) ([]byte, func()) {
	batchIndex := globalCount / sh.branches * sh.ChunkSize()
	batchPos := globalCount % sh.branches * sh.ChunkSize()
	batchNodeIndex := batchPos / sh.ChunkSize()
	batchNodePos := batchPosIndex % sh.ChunkSize()
	return sh.batches[batchIndex].nodes[batchNodeIndex].getSectionBuffer(batchNodePos)
}

func (n *node) getSectionBuffer(p int) (int, func()) {
	currentCount := atomic.AddInt32(&n.secCnt, 1)
	nodeSectionByteOffset := (batchNodePos / sh.BlockSize()) * sh.BlockSize()
	var doneFunc func()
	if currentCount == int32(n.branches) {
		doneFunc = n.done
	}
	return n.buffer[nodeSectionByteOffset : nodeSectionByteOffset+sh.BlockSize()], batchNodeIndex, doneFunc
}

// dataSpan returns the size of data encoded under the current node, serialized as big endian uint64
func (n *node) dataSpan() []byte {
	//secsize := n.hasher.BlockSize()
	span := uint64(n.hasher.ChunkSize())
	for l := 0; l < n.levelIndex; l++ {
		span *= uint64(n.branches)
	}
	meta := make([]byte, 8)
	binary.BigEndian.PutUint64(meta, span)
	return meta
}

func (n *node) Write(sectionIndex int, section []byte) {
	n.write(sectionIndex, section)
}

func (n *node) write(sectionIndex int, section []byte) {
	currentCount := atomic.AddInt32(&n.secCnt, 1)
	n.hasher.Write(sectionIndex, section)
	if currentCount == int32(n.branches) {
		n.node()
	}
}

func (n *node) done() {
	go func() {
		parentBatchIndex := n.index / n.branches
		parentBatch := n.levels[n.levelIndex+1].getBatch(parentBatchIndex)
		parentNodeIndex := n.index % n.branches
		parentNode := parentBatch.nodes[parentNodeIndex]
		parentNode.write(n.pos, n.hasher.Sum(nil, n.hasher.ChunkSize(), parentNode.dataSpan()))
	}()
}

// length is global length
func (n *node) sum(length int, nodeSpan int) {

	// nodeSpan is the total byte size of a complete tree under the current node
	nodeSpan *= n.branches

	// if a new batch would be started
	batchSpan := nodeSpan * n.branches
	nodeIndex := length % batchSpan
	var parentNode *node
	if nodeIndex == 0 && len(n.levels) > n.levelIndex+1 {
		batchIndex := (length-1)/batchSpan + 1
		parentNode = n.levels[n.levelIndex+1].getBatch(batchIndex).nodes[nodeIndex]
		parentNode.sum(length, nodeSpan)
		return
	}

	// dataLength is the actual length of data under the current node
	dataLength := uint64(length % nodeSpan)

	// meta is the length of actual data in the nodespan
	meta := make([]byte, 8)
	binary.BigEndian.PutUint64(meta, dataLength)

	// bmtLength is the actual length of bytes in the chunk
	// if the node is an intermediate node (level != 0 && len(levels) > 1), bmtLength will be a multiple 32 bytes
	var bmtLength uint64
	if n.levelIndex == 0 {
		bmtLength = dataLength
	} else {
		bmtLength = ((dataLength - 1) / uint64((nodeSpan/n.branches+1)*n.hasher.BlockSize()))
	}

	hash := n.hasher.Sum(nil, int(bmtLength), meta)

	// are we on the root level?
	if parentNode != nil {
		parentNode.sum(length, nodeSpan)
		return
	}

	n.result <- hash
}

func (fh *FileHasher) ChunkSize() int {
	return fh.branches * fh.secsize
}

// Louis note to self: secsize is the same as the size of the reference
// Invoked after we know the actual length of the file
// Will create the last node on the data level of the hash tree matching the length
func (fh *FileHasher) Sum(b []byte, length int, meta []byte) []byte {

	// handle edge case where the file is empty
	if length == 0 {
		return fh.hasherFunc().Sum(nil, 0, make([]byte, 8))
	}

	// calculate the index the last batch
	lastBatchIndexInFile := (length - 1) / fh.ChunkSize() * fh.branches

	// calculate the node index within the last batch
	byteIndexInLastBatch := length - lastBatchIndexInFile*fh.ChunkSize()*fh.branches
	nodeIndexInLastBatch := (byteIndexInLastBatch - 1) / fh.ChunkSize()

	// get the last node
	lastNode := fh.levels[0].getBatch(lastBatchIndexInFile).nodes[nodeIndexInLastBatch]

	// asynchronously call sum on this node and wait for the final result
	go lastNode.sum(length, fh.ChunkSize())
	return <-fh.result
}

// Reset puts FileHasher in a (re)useable state
func (sh *FileHasher) Reset() {
	sh.mtx.Lock()
	defer sh.mtx.Unlock()
	sh.levels = nil
}
