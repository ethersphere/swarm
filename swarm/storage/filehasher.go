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
	levelIndex  int // which level of the swarm hash tree
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
	hasher     bmt.SectionWriter // async hasher
	pos        int               // index of the node chunk within its batch
	secCnt     int32             // number of sections written
	size       int
	nodeBuffer []byte
	*batch     // pointer to containing batch
	lock       sync.Mutex
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
	log.Warn("creating level", "l", pl)
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
	for _, n := range b.nodes {
		n.hasher.Reset()
	}
	for i, _ := range b.batchBuffer {
		b.batchBuffer[i] = byte(0x0)
	}
	b.pool.Put(b)
}

// TODO: rename as blocksize in bmt is hardcoded 2*segmentsize (is that correct?) to avoid ambiguity
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
	for i := range nodes {
		offset := chunkSize * i
		nodes[i] = &node{
			pos:        i,
			hasher:     fh.hasherFunc(),
			nodeBuffer: bt.batchBuffer[offset : offset+chunkSize],
			batch:      bt,
		}
	}

	bt.nodes = nodes
	return bt
}

// writes data to offset count position
func (fh *FileHasher) WriteBuffer(globalCount int, buf []byte) (int, error) {

	// writes are only valid on section thresholds
	if globalCount%fh.BlockSize() > 0 {
		return 0, fmt.Errorf("offset must be multiples of blocksize %d", fh.BlockSize())
	}

	// retrieve the node we are writing to
	batchIndex := globalCount / (fh.branches * fh.ChunkSize())
	batchPos := globalCount % (fh.branches * fh.ChunkSize())
	batchNodeIndex := batchPos / fh.ChunkSize()
	batchNodePos := batchPos % fh.ChunkSize()
	bt := fh.levels[0].getOrCreateBatch(batchIndex)
	nod := bt.nodes[batchNodeIndex]

	nod.hasher.Write(batchNodePos/fh.BlockSize(), buf)
	currentCount := atomic.AddInt32(&nod.secCnt, 1)
	log.Trace("fh writebuf", "c", globalCount, "s", globalCount/fh.BlockSize(), "seccnt", nod.secCnt, "branches", nod.branches, "buflen", len(buf), "node", fmt.Sprintf("%p", nod), "batch", fmt.Sprintf("%p", nod.batch), "buf", buf[:])
	if currentCount == int32(nod.branches) {
		nod.done(nod.ChunkSize(), nod.ChunkSize(), nod.getOrCreateParent())
	}
	return fh.BlockSize(), nil
}

// called when the final length of the data is known
func (fh *FileHasher) SetLength(l int64) {
	fh.dataLength = l
	return
}

// dataSpan returns the size of data encoded under the current node
func (n *node) span(l uint64) uint64 {
	span := uint64(n.ChunkSize())
	var lev int
	for lev = 0; lev < n.levelIndex; lev++ {
		span *= uint64(n.branches)
	}
	if l < span && lev == 0 {
		return l
	}
	return span
}

func (n *node) write(sectionIndex int, section []byte) {
	currentCount := atomic.AddInt32(&n.secCnt, 1)

	log.Debug("write intermediate", "pos", n.pos, "section", sectionIndex, "level", n.levelIndex, "data", section, "buffer", fmt.Sprintf("%p", n.nodeBuffer), "batchbuffer", fmt.Sprintf("%p", n.batchBuffer), "batch", fmt.Sprintf("%p", n.batch), "node", fmt.Sprintf("%p", n))
	n.hasher.Write(sectionIndex, section)
	bytePos := sectionIndex * n.BlockSize()
	copy(n.nodeBuffer[bytePos:bytePos+n.BlockSize()], section)
	if currentCount == int32(n.branches) {
		if n.levelIndex == 0 {
			go n.done(n.ChunkSize(), n.ChunkSize(), n.getOrCreateParent())
		} else {
			span := n.ChunkSize()
			for i := 0; i < n.levelIndex; i++ {
				span *= n.branches
			}
			go n.done(n.ChunkSize(), span, n.getOrCreateParent())
		}
	}
}

func (n *node) getOrCreateParent() *node {
	parentBatchIndex := n.index / n.branches
	parentBatch := n.getLevel(n.levelIndex + 1).getOrCreateBatch(parentBatchIndex)
	parentNodeIndex := n.index % n.branches
	return parentBatch.nodes[parentNodeIndex]
}

func (n *node) done(nodeLength int, spanLength int, parentNode *node) {
	serializedLength := make([]byte, 8)
	binary.LittleEndian.PutUint64(serializedLength, uint64(spanLength))
	log.Debug("node done", "n", fmt.Sprintf("%p", n), "serl", serializedLength, "parent", fmt.Sprintf("%p", parentNode), "l", nodeLength, "pos", n.pos)
	h := n.hasher.Sum(nil, nodeLength, serializedLength)
	parentNode.write(n.pos, h)
	if n.pos == n.branches-1 {
		log.Debug("delink", "n", fmt.Sprintf("%p", n), "b", fmt.Sprintf("%p", n.batch))
		//n.batch.delink()
	}
}

// length is global length
func (n *node) sum(length int64, potentialSpan int64) {

	if length == 0 {
		n.result <- n.hasher.Sum(nil, 0, nil)
		return
	}
	// span is the total byte size of a complete tree under the current node
	potentialSpan *= int64(n.branches)

	// dataLength is the actual length of data under the current node
	// bmtLength is the actual length of bytes in the chunk to be summed
	// if the node is an intermediate node (level != 0 && len(levels) > 1), bmtLength will be a multiple 32 bytes
	var dataLength uint64
	dataLength = uint64(length) % uint64(potentialSpan)

	// meta is the length of actual data in the nodespan serialized little-endian
	meta := make([]byte, 8)
	if dataLength == 0 {
		binary.LittleEndian.PutUint64(meta, uint64(length))
	} else {
		binary.LittleEndian.PutUint64(meta, dataLength)
	}

	// we already checked on top if length is 0. If it is 0 here, it's on span threshold and a full chunk write
	// otherwise we do not have a full chunk write, and need to make the underlying hash sum
	if dataLength == 0 {
		// get the parent node if it exists
		parentNode := n.getParent(length)
		parentNode.sum(length, potentialSpan)
		return
	}

	var bmtLength int
	if n.levelIndex == 0 {
		bmtLength = int(dataLength)
	} else {
		log.Debug("calc bmtl", "dl", dataLength, "span", potentialSpan)
		bmtLength = int(((dataLength-1)/uint64((potentialSpan/int64(n.branches))) + 1) * uint64(n.BlockSize()))
	}

	log.Debug("bmtl", "l", bmtLength, "dl", dataLength, "n", fmt.Sprintf("%p", n), "pos", n.pos, "seccnt", n.secCnt)

	if n.secCnt > 1 {
		log.Debug("seccnt > 1", "nbuf", n.nodeBuffer, "dl", dataLength, "n", fmt.Sprintf("%p", n), "l", n.levelIndex)
		n.done(int(bmtLength), int(dataLength), n.getOrCreateParent())
		parentNode := n.getParent(length)
		parentNode.sum(length, potentialSpan)
		return
	}

	if n.index == 0 {
		if n.pos == 0 {
			// if it's on data level, we have to make the hash
			// otherwise it's already hashed
			if n.levelIndex == 0 {
				n.result <- n.hasher.Sum(nil, bmtLength, meta)
				return
			}
			log.Debug("result direct no hash", "n", fmt.Sprintf("%p", n), "l", n.levelIndex)
			n.result <- n.nodeBuffer[:n.BlockSize()]
			return
			// TODO: instead of this situation we should find the correct parent directly and write the hash to it
		} else if n.levelIndex > 0 {
			parentNode := n.getParent(length)
			parentNode.write(n.pos, n.nodeBuffer)
			parentNode.sum(length, potentialSpan)
			return
		}

	}

	var levelCount int
	prevIdx := n.index
	for i := prevIdx; i > 0; i /= n.branches {
		prevIdx = i
		levelCount++
	}

	// get the top node. This will always have free capacity
	topRoot := n.levels[len(n.levels)-1].getBatch(0).nodes[0]
	danglingTop := n.levelIndex + levelCount
	log.Debug("levelcount", "l", levelCount, "previdx", prevIdx, "n", fmt.Sprintf("%p", n), "nindex", n.index)
	var nodeToWrite *node
	// if there is a tree unconnected to the root, append to this and write result to root
	if danglingTop == len(n.levels) {
		nodeToWrite := n.levels[danglingTop].getBatch(0).nodes[prevIdx%n.branches]
		log.Debug("have dangling", "n", nodeToWrite)
		nodeToWrite.write(int(nodeToWrite.secCnt), n.hasher.Sum(nil, n.BlockSize(), meta))

	} else {
		nodeToWrite = n
	}

	log.Debug("nodetowrite", "n", fmt.Sprintf("%p", nodeToWrite), "sec", nodeToWrite.secCnt, "meta", meta)
	topRoot.write(int(topRoot.secCnt), nodeToWrite.hasher.Sum(nil, int(nodeToWrite.secCnt)*n.BlockSize(), meta))
	binary.LittleEndian.PutUint64(meta, uint64(length))
	log.Debug("top", "n", topRoot.nodeBuffer)
	n.result <- topRoot.hasher.Sum(nil, int(topRoot.secCnt)*n.BlockSize(), meta)
}

func (fh *FileHasher) ChunkSize() int {
	return fh.branches * fh.secsize
}

func (n *node) getParent(length int64) *node {
	nextLevel := n.levelIndex + 1
	if len(n.levels) > nextLevel {
		var levelBytePos = length
		for i := 0; i < nextLevel; i++ {
			levelBytePos /= int64(n.branches)
		}
		parentBatchIndex := (levelBytePos - 1) / int64(n.branches*n.ChunkSize())
		parentNodeIndex := (levelBytePos % int64(n.branches*n.ChunkSize()) / int64(n.ChunkSize()))
		parentLevel := n.levels[nextLevel]
		parentBatch := parentLevel.getBatch(int(parentBatchIndex))
		log.Debug("parentbatch", "b", fmt.Sprintf("%p", parentBatch), "level", nextLevel, "nodeindex", parentNodeIndex)
		if parentBatch != nil {
			return parentBatch.nodes[parentNodeIndex]
		}
	}
	return nil
}

// Invoked after we know the actual length of the file
// Will create the last node on the data level of the hash tree matching the length
func (fh *FileHasher) Sum(b []byte) []byte {

	// handle edge case where the file is empty
	if fh.dataLength == 0 {
		return fh.hasherFunc().Sum(nil, 0, make([]byte, 8))
	}

	// calculate the index the last batch
	lastBatchIndexInFile := (fh.dataLength - 1) / int64(fh.ChunkSize()*fh.branches)

	// calculate the node index within the last batch
	byteIndexInLastBatch := fh.dataLength - lastBatchIndexInFile*int64(fh.ChunkSize()*fh.branches)
	nodeIndexInLastBatch := (int(byteIndexInLastBatch) - 1) / fh.ChunkSize()

	// get the last node on the data level
	lastNode := fh.levels[0].getBatch(int(lastBatchIndexInFile)).nodes[nodeIndexInLastBatch]
	// asynchronously call sum on this node and wait for the final result
	go func() {
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
