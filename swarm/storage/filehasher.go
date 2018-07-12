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
	SectionSize() int
	Sum(b []byte, length int, meta []byte) []byte
}

// FileHasher is instantiated each time a file is swarm hashed
// itself implements the ChunkHasher interface
type FileHasher struct {
	mtx         sync.Mutex           // RW lock to add/read levels push and unshift batches
	pool        sync.Pool            // batch resource pool
	levels      []*level             // levels of the swarm hash tree
	secsize     int                  // section size
	chunks      int                  // number of chunks read
	offset      int                  // byte offset (cursor) within chunk
	read        int                  // length of input data read
	length      int                  // known length of input data
	branches    int                  // branching factor
	hasherFunc  func() SectionHasher // hasher constructor
	result      chan []byte          // channel to put hash asynchronously
	lastSection []byte               // last section to record
	lastSecPos  int                  // pos of section within last section
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
	return sh
}

// level captures one level of chunks in the swarm hash tree
// singletons are attached to the lowest level
type level struct {
	lev         int      // which level of the swarm hash tree
	batches     []*batch // active batches on the level
	*FileHasher          // pointer to the underlying hasher
}

// batch records chunks subsumed under the same parent intermediate chunk
type batch struct {
	nodes  []*node // nodes of the batches
	index  int     // offset of the node
	parent *node   // pointer to containing
	*level         // pointer to containing level
}

// node represent a chunk and embeds an async interface to the chunk hash used
type node struct {
	hasher    SectionHasher // async hasher
	pos       int           // index of the node chunk within its batch
	secCnt    int32         // number of sections written
	maxSecCnt int32         // maximum number of sections written
	*batch                  // pointer to containing batch
}

// getParentLevel retrieves or creates the next level up from a node/batch/level
// using lock for concurrent access
func (lev *level) getLevel(pl int) (par *level) {
	if pl < len(lev.levels) {
		return lev.levels[pl]
	}
	par = &level{
		lev: pl,
	}
	lev.levels = append(lev.levels, par)
	return par
}

// getParent retrieves the parent node for the batch, creating a new batch if needed
// allownil set to true will return a nil if parent
func (b *batch) getParent(allowNil bool) (n *node) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.parent != nil || allowNil {
		return b.parent
	}
	b.parent = b.getParentNode()
	return b.parent
}

// getBatch looks up the parent batch on the next level up
// caller must hold the lock
func (lev *level) getBatch(index int) (pb *batch) {
	// parent batch is memoised and typically expect 1 or 2 batches
	// so this simple way of getting the appropriate batch is ok
	for _, pb = range lev.batches {
		if pb.index == index {
			return pb
		}
	}
	return nil
}

// getParentNode retrieves the parent node based on the batch indexes
// if a new level or batch is required it creates them
// caller must hold the lock
func (b *batch) getParentNode() *node {
	pos := b.index % b.branches
	pi := 0
	if b.index > 0 {
		pi = (b.index - 1) / b.branches
	}
	b.mtx.Lock()
	defer b.mtx.Unlock()
	pl := b.getLevel(b.lev + 1)
	pb := pl.getBatch(pi)
	if pb != nil {
		return pb.nodes[pos]
	}
	pb = b.pool.Get().(*batch)
	pb.level = pl
	pb.index = b.index / b.branches

	pl.batches = append(pl.batches, pb)
	return pb.nodes[pos]
}

// delink unshifts the levels batches
// and releases the popped batch to the batch pools
// must be called after Sum has returned
// section writes or children no longer reference this batch
func (b *batch) delink() {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	first := b.batches[0]
	if first.index != b.index {
		panic("non-initial batch finished first")
	}
	b.pool.Put(first)
	b.batches = b.batches[1:]
}

// newBatch constructs a reuseable batch
func (sh *FileHasher) newBatch() *batch {
	nodes := make([]*node, sh.branches)
	for i, _ := range nodes {
		nodes[i] = &node{
			pos:    i,
			hasher: sh.hasherFunc(),
		}
	}
	return &batch{
		nodes: nodes,
	}
}

// dataSpan returns the
func (n *node) dataSpan() int64 {
	secsize := n.hasher.SectionSize()
	span := int64(4096 / secsize)
	for l := 0; l < n.lev; l++ {
		span *= int64(n.branches)
	}
	return span
}

// SimpleSplitter implements the hash.Hash interface for synchronous read from data
// as data is written to it, it chops the input stream to section size buffers
// and calls the section write on the SectionHasher

// Reset puts FileHasher in a (re)useable state
func (sh *FileHasher) Reset() {
	sh.mtx.Lock()
	defer sh.mtx.Unlock()
	sh.levels = nil
}

// //
// func (sh *FileHasher) Write(buf []byte) {
// 	chunkSize := sh.secsize * sh.branches
// 	start := sh.offset / sh.secsize
// 	pos := sh.sections % sh.branches
// 	n := sh.getLevel(0).getBatch(sh.chunks).nodes[pos]
// 	read := chunkSize - sh.offset
// 	copy(n.chunk[sh.offset:], buf)
// 	var canBeFinal, isFinal bool
// 	// assuming input never exceeds set length
// 	if len(buf) <= read {
// 		read = len(buf)
// 		canBeFinal = true
// 		sh.mtx.Lock()
// 		sizeKnown := sh.length > 0
// 		if sizeKnown {
// 			isFinal = sh.chunks*chunkSize-sh.length <= chunkSize
// 		} else {
// 			canBeFinal = false
// 			sh.mtx.Unlock()
// 		}
// 	}
// 	end := start + (sh.offset%sh.secsize+read)/sh.secsize - 1
// 	// if current chunk reaches the end
// 	// write the final section
// 	if canBeFinal {
// 		end--
// 		lastSecSize := (sh.offset + read) % sh.secsize
// 		lastSecOffset := end * sh.secsize
// 		sh.lastSection = n.chunk[lastSecOffset : lastSecOffset+lastSecSize]
// 		sh.lastSecPos = end
// 		// lock should be kept until lastSection and
// 		sh.mtx.Unlock()
// 		if isFinal {
// 			n.write(end, sh.lastSection, true)
// 		}
// 	}
// 	f := func() {
// 		for i := start; i < end; i++ {
// 			n.write(i, n.chunk[i*sh.secsize:(i+1)*sh.secsize], false)
// 		}
// 	}
//
// 	sh.offset = (sh.offset + read) % sh.secsize * sh.branches
// 	rest := buf[read:]
// 	if len(rest) == 0 {
// 		go f()
// 		return
// 	}
// 	sh.Write(rest)
// }

// Sum
func (sh *FileHasher) Sum(b []byte, length int, meta []byte) []byte {
	chunkSize := sh.secsize * sh.branches
	sh.mtx.Lock()
	if sh.read >= sh.length {
		n := sh.getNode(sh.lastSecPos)
		n.write(sh.lastSecPos, sh.lastSection, true)
	}
	sh.mtx.Unlock()
	return <-sh.result
}

// write writes the section to the node at section idx
// the final parameter indicates that the section is final
// i.e., the read input buffer has been consumed
func (n *node) write(idx int, section []byte, final bool) {
	// write the section to the hasher
	n.hasher.Write(idx, section)
	var inferred bool
	var maxSecCnt int32
	if final {
		// set number of chunks based on last index and save it
		maxSecCnt = int32(idx + 1)
		atomic.StoreInt32(&n.maxSecCnt, maxSecCnt)
	} else {
		// load max number of sections (known from a previous call to final or hash)
		maxSecCnt = atomic.LoadInt32(&n.maxSecCnt)
		if maxSecCnt == 0 {
			inferred = true
			maxSecCnt = int32(n.branches)
		}
	}

	// another section is written, increment secCnt
	secCnt := atomic.AddInt32(&n.secCnt, 1)

	// if all branches been written do sum
	// since secCnt is > 0 by now, the condition  is not satisfied iff
	// * maxSecCnt is set and reached or
	// * secCnt is n.branches
	if secCnt%maxSecCnt > 0 {
		return
	}
	// final flag either because
	// * argument explicit about it OR
	// * was set earlier by a call to final
	go func() {
		defer n.batch.delink()
		final = final || !inferred
		corr := n.hasher.SectionSize() - len(section)
		length := int(maxSecCnt)*n.hasher.SectionSize() - corr
		// can subtract corr directly from span assuming that shorter sections can only occur on level 0
		span := n.dataSpan()*int64(maxSecCnt) - int64(corr)
		meta := make([]byte, 8)
		binary.BigEndian.PutUint64(meta, uint64(span))
		// blocking call to Sum (releases resource, so node hasher is reusable)
		hash := n.hasher.Sum(nil, length, meta)
		// before return, delink the batch
		defer n.delink()
		// if the final section is batch 0 / pos 0 then it is
		allowNil := final && n.index == 0 && n.pos == 0
		pn := n.getParent(allowNil)
		if pn == nil {
			n.result <- hash
			return
		}
		pn.write(n.pos, hash, final)
	}()
}
