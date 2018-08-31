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
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

type SectionHasher interface {
	bmt.SectionWriter
	WriteBuffer(globalCount int64, r io.Reader) (int, error)
}

// FileHasher is instantiated each time a file is swarm hashed
// itself implements the ChunkHasher interface
type FileHasher struct {
	mtx        sync.Mutex               // RW lock to add/read levels push and unshift batches
	pool       sync.Pool                // batch resource pool
	levels     []*level                 // levels of the swarm hash tree
	secsize    int                      // section size
	branches   int                      // branching factor
	hasherFunc func() bmt.SectionWriter // SectionWriter // hasher constructor
	batchSize  int                      // byte length of a batch
	result     chan []byte              // channel to put hash asynchronously
	digestSize int
	dataLength int64
	lnBranches float64
}

//func NewFileHasher(hasherFunc func() SectionWriter, branches int, secSize int) *FileHasher {
func NewFileHasher(hasherFunc func() bmt.SectionWriter, branches int, secSize int) *FileHasher {
	fh := &FileHasher{
		hasherFunc: hasherFunc,
		result:     make(chan []byte),
		branches:   branches,
		secsize:    secSize,
		batchSize:  branches * branches * secSize,
	}
	fh.lnBranches = math.Log(float64(branches))
	fh.pool = sync.Pool{
		New: func() interface{} {
			return fh.newBatch()
		},
	}
	fh.digestSize = secSize //hasherFunc().Size()

	fh.levels = append(fh.levels, &level{
		FileHasher: fh,
		levelIndex: 0,
	})
	return fh
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
	nodes       []*node // nodes of the batches
	parent      *node   // pointer to containing
	batchBuffer []byte  // data buffer for batch (divided between nodes)
	index       int     // offset of the batch
	*level              // pointer to containing level
}

// node represent a chunk and embeds an async interface to the chunk hash used
type node struct {
	lock       sync.Mutex
	hasher     bmt.SectionWriter // async hasher
	pos        int               // index of the node chunk within its batch
	secCnt     int32             // number of sections written
	size       int
	nodeBuffer []byte
	//writeComplete chan struct{}
	*batch // pointer to containing batch
}

// for logging purposes
func (n *node) getBuffer() []byte {
	n.lock.Lock()
	defer n.lock.Unlock()
	b := make([]byte, len(n.nodeBuffer))
	copy(b, n.nodeBuffer)
	return b
}

// getParentLevel retrieves or creates the next level up from a node/batch/level
// using lock for concurrent access
func (lev *level) getLevel(pl int) (par *level) {
	if pl < len(lev.levels) {
		return lev.levels[pl]
	}
	par = &level{
		levelIndex: pl,
		FileHasher: lev.FileHasher,
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
	log.Warn("getbatch", "b", fmt.Sprintf("%p", pb))
	if pb == nil {
		pb = lev.pool.Get().(*batch)
		pb.index = index
		pb.level = lev
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

func (fh *FileHasher) BlockSize() int {
	return fh.secsize
}

// returns the digest size of the underlying hasher
func (fh *FileHasher) Size() int {
	return fh.digestSize
}

func (fh *FileHasher) WriteSection(idx int64, data []byte) int {
	return 0
}

// newBatch constructs a reuseable batch
func (fh *FileHasher) newBatch() (bt *batch) {
	nodes := make([]*node, fh.branches)
	chunkSize := fh.ChunkSize()
	bt = &batch{
		batchBuffer: make([]byte, fh.branches*chunkSize),
	}
	log.Debug("newbatch", "bufat", fmt.Sprintf("%p", bt.batchBuffer))
	for i := range nodes {
		offset := chunkSize * i
		nodes[i] = &node{
			pos:        i,
			hasher:     fh.hasherFunc(),
			nodeBuffer: bt.batchBuffer[offset : offset+chunkSize],
			batch:      bt,
			//writeComplete: make(chan struct{}),
		}
	}

	log.Debug("newbatch node", "bufat", fmt.Sprintf("%p", nodes[0].batchBuffer), "node frst bufat", fmt.Sprintf("%p", nodes[0].nodeBuffer), "node last bufat", fmt.Sprintf("%p", nodes[len(nodes)-1].nodeBuffer))
	bt.nodes = nodes
	return bt
}

// level depth is index of level ascending from data level towards tree root
func (fh *FileHasher) OffsetToLevelDepth(c int64) int {
	chunkCount := c / int64(fh.ChunkSize())
	level := int(math.Log(float64(chunkCount)) / fh.lnBranches)
	//log.Warn("chunksize", "offset", c, "c", fh.ChunkSize(), "b", fh.branches, "s", fh.secsize, "count", chunkCount, "level", level)
	return level
}

// writes data to offset count position
func (fh *FileHasher) WriteBuffer(globalCount int, r io.Reader) (int, error) {

	// writes are only valid on section thresholds
	if globalCount%fh.BlockSize() > 0 {
		return 0, fmt.Errorf("offset must be multiples of blocksize %d", fh.BlockSize())
	}

	// retrieve the node we are writing to
	batchIndex := globalCount / (fh.branches * fh.ChunkSize())
	batchPos := globalCount % (fh.branches * fh.ChunkSize())
	batchNodeIndex := batchPos / fh.ChunkSize()
	batchNodePos := batchPos % fh.ChunkSize()
	log.Debug("batch", "nodepos", batchNodePos, "node", batchNodeIndex, "global", globalCount, "batchindex", batchIndex, "batchpos", batchPos, "blockSize", fh.BlockSize())
	bt := fh.levels[0].getOrCreateBatch(batchIndex)
	nod := bt.nodes[batchNodeIndex]

	nod.lock.Lock()
	buf := nod.nodeBuffer[batchNodePos : batchNodePos+fh.BlockSize()]
	c, err := r.Read(buf)
	nod.lock.Unlock()
	if err != nil {
		return 0, err
	} else if c < fh.BlockSize() {
		return 0, io.ErrUnexpectedEOF
	}
	nod.hasher.Write(batchNodePos/fh.BlockSize(), buf)
	currentCount := atomic.AddInt32(&nod.secCnt, 1)
	log.Debug("fh writebuf", "c", globalCount, "s", globalCount/fh.BlockSize(), "seccnt", nod.secCnt, "branches", nod.branches, "buflen", len(buf), "node", fmt.Sprintf("%p", nod), "buf", buf[:])
	if currentCount == int32(nod.branches) {
		go nod.done(nod.ChunkSize())
	}
	return fh.BlockSize(), nil
}

// called when the final length of the data is known
func (fh *FileHasher) SetLength(l int64) {
	fh.dataLength = l
	return
}

// dataSpan returns the size of data encoded under the current node
func (n *node) span() uint64 {
	span := uint64(n.ChunkSize())
	for l := 0; l < n.levelIndex; l++ {
		span *= uint64(n.branches)
	}
	return span
}

func (n *node) write(sectionIndex int, section []byte) {
	currentCount := atomic.AddInt32(&n.secCnt, 1)

	log.Debug("writing", "pos", n.pos, "section", sectionIndex, "level", n.levelIndex, "data", section, "buffer", fmt.Sprintf("%p", n.nodeBuffer), "batchbuffer", fmt.Sprintf("%p", n.batchBuffer), "barch", fmt.Sprintf("%p", n.batch), "level", fmt.Sprintf("%p", n.getLevel(n.levelIndex)), "node", fmt.Sprintf("%p", n))
	n.hasher.Write(sectionIndex, section)
	bytePos := sectionIndex * n.BlockSize()
	n.lock.Lock()
	copy(n.nodeBuffer[bytePos:bytePos+n.BlockSize()], section)
	n.lock.Unlock()
	if currentCount == int32(n.branches) {
		go n.done(n.ChunkSize())
	}
}

func (n *node) done(l int) {
	parentBatchIndex := n.index / n.branches
	parentBatch := n.getLevel(n.levelIndex + 1).getOrCreateBatch(parentBatchIndex)
	parentNodeIndex := n.index % n.branches
	parentNode := parentBatch.nodes[parentNodeIndex]
	serializedLength := make([]byte, 8)
	binary.BigEndian.PutUint64(serializedLength, parentNode.span())
	h := n.hasher.Sum(nil, l, serializedLength)
	parentNode.write(n.pos, h)
}

// length is global length
func (n *node) sum(length int64, span int64) {

	if length == 0 {
		n.result <- n.hasher.Sum(nil, 0, nil)
		return
	}
	// span is the total byte size of a complete tree under the current node
	span *= int64(n.branches)

	// dataLength is the actual length of data under the current node
	var dataLength uint64
	dataLength = uint64(length) % uint64(span)

	// meta is the length of actual data in the nodespan
	meta := make([]byte, 8)
	binary.BigEndian.PutUint64(meta, dataLength)

	// bmtLength is the actual length of bytes in the chunk
	// if the node is an intermediate node (level != 0 && len(levels) > 1), bmtLength will be a multiple 32 bytes
	var bmtLength uint64
	if n.levelIndex == 0 {
		bmtLength = dataLength
	} else {
		denom := float64(span / int64(n.branches))
		div := float64(dataLength)
		bmtLength = uint64(div/denom) * uint64(n.secsize)
		log.Debug("bmtlengthcalc", "denom", denom, "div", div, "bmtl", bmtLength)
	}

	// if a new batch would be started
	var parentNode *node
	nextLevel := n.levelIndex + 1
	if nextLevel != len(n.levels) {
		var levelBytePos = length
		for i := 0; i < nextLevel; i++ {
			levelBytePos /= int64(n.branches)
		}
		parentBatchIndex := levelBytePos / int64(n.branches*n.ChunkSize())
		parentNodeIndex := (levelBytePos % int64(n.branches*n.ChunkSize()) / int64(n.ChunkSize()))
		log.Debug("next", "parentbatchindex", parentBatchIndex, "parentnodeindex", parentNodeIndex, "levelbytepos", levelBytePos)
		//if levelBytePos < int64(n.ChunkSize()) {
		if levelBytePos > 0 {
			parentLevel := n.levels[nextLevel]
			parentBatch := parentLevel.getBatch(int(parentBatchIndex))
			log.Debug("parentbatch", "b", fmt.Sprintf("%p", parentBatch), "l", parentLevel)
			if parentBatch != nil {
				parentNode = parentBatch.nodes[parentNodeIndex]
			}
		}
		//parentBatchSpan := span * int64(n.branches)
		//parentNodeIndex := ((length % int64(parentBatchSpan)) - 1) / int64(n.ChunkSize())
		//nodeIndex := (length%span - 1) / int64(n.ChunkSize())
		//nodeBatchIndex := ((length % int64(n.branches)) - 1) / int64(n.branches*n.ChunkSize())
		//		parentBatchIndex := (length - 1) / int64(parentBatchSpan) // + 1
		//
		//		parentLevel := n.levels[n.levelIndex+1]
		//		parentBatch := parentLevel.getBatch(int(parentBatchIndex))
		//		if parentBatch != nil {
		//			parentNode = parentBatch.nodes[nodeIndex]
		//		}
		//		log.Warn("node sum 1", "batchindex", batchIndex, "b", n.branches, "lv", len(n.levels), "nln", n.levelIndex, "nidx", nodeIndex, "parentnode", fmt.Sprintf("%p", parentNode), "parentlevel", parentLevel)

	}

	// are we on the root level?
	if parentNode != nil {
		log.Warn("continue", "hasher", fmt.Sprintf("%p", n.hasher), "parent", fmt.Sprintf("%p", parentNode), "thisnode", fmt.Sprintf("%p", n))
		parentNode.sum(length, span)
		return
	}

	log.Debug("summing", "l", length, "dl", dataLength, "meta", meta, "bmtlength", bmtLength, "hasher", fmt.Sprintf("%p", n.hasher), "thisnode", fmt.Sprintf("%p", n), "l", n.levelIndex, "span", span)
	log.Debug("nodebuffer", "b", n.nodeBuffer)
	hash := n.hasher.Sum(nil, int(bmtLength), meta)
	n.result <- hash
}

func (fh *FileHasher) ChunkSize() int {
	return fh.branches * fh.secsize
}

// Louis note to self: secsize is the same as the size of the reference
// Invoked after we know the actual length of the file
// Will create the last node on the data level of the hash tree matching the length
//func (fh *FileHasher) Sum(b []byte, length int, meta []byte) []byte {
func (fh *FileHasher) Sum(b []byte) []byte {

	// handle edge case where the file is empty
	if fh.dataLength == 0 {
		return fh.hasherFunc().Sum(nil, 0, make([]byte, 8))
	}

	log.Debug("fh sum", "length", fh.dataLength)
	// calculate the index the last batch
	lastBatchIndexInFile := (fh.dataLength - 1) / int64(fh.ChunkSize()*fh.branches)

	// calculate the node index within the last batch
	byteIndexInLastBatch := fh.dataLength - lastBatchIndexInFile*int64(fh.ChunkSize()*fh.branches)
	nodeIndexInLastBatch := (int(byteIndexInLastBatch) - 1) / fh.ChunkSize()

	// get the last node
	lastNode := fh.levels[0].getBatch(int(lastBatchIndexInFile)).nodes[nodeIndexInLastBatch]
	log.Debug("lastnode", "batchindex", lastBatchIndexInFile, "nodeindex", nodeIndexInLastBatch)

	// asynchronously call sum on this node and wait for the final result
	go func() {
		chunkDataLength := int(fh.dataLength) % fh.ChunkSize()
		if chunkDataLength > 0 && fh.dataLength != 0 {
			lastNode.done(chunkDataLength)
		}
		lastNode.sum(fh.dataLength, int64(fh.BlockSize()))
	}()
	return <-fh.result
}

// Reset puts FileHasher in a (re)useable state
func (sh *FileHasher) Reset() {
	sh.mtx.Lock()
	defer sh.mtx.Unlock()
	sh.levels = nil
}
