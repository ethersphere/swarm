// Copyright 2020 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bmt

import (
	"context"
	"errors"
	"sync"

	"github.com/ethersphere/swarm/bmt"
)

// NewAsyncWriter extends Hasher with an interface for concurrent segment.GetSection() writes
// TODO: Instead of explicitly setting double size of segment should be dynamic and chunked internally. If not, we have to keep different bmt hashers generation functions for different purposes in the same instance, or cope with added complexity of bmt hasher generation functions having to receive parameters
func NewAsyncHasher(ctx context.Context, h *bmt.Hasher, double bool, errFunc func(error)) *AsyncHasher {
	secsize := h.SectionSize()
	if double {
		secsize *= 2
	}
	seccount := h.Branches()
	if double {
		seccount /= 2
	}
	return &AsyncHasher{
		Hasher:   h,
		double:   double,
		secsize:  secsize,
		seccount: seccount,
		ctx:      ctx,
		errFunc:  errFunc,
	}
}

// AsyncHasher extends BMT Hasher with an asynchronous segment.GetSection() writer interface
// AsyncHasher cannot be used as with a hash.Hash interface: It must be used with the
// right indexes and length and the right number of sections
// It is unsafe and does not check indexes and section data lengths
//
// behaviour is undefined if
// * non-final sections are shorter or longer than secsize
// * if final section does not match length
// * write a section with index that is higher than length/secsize
// * set length in Sum call when length/secsize < maxsec
//
// * if Sum() is not called on a Hasher that is fully written
//   a process will block, can be terminated with Reset
// * it will not leak processes if not all sections are written but it blocks
//   and keeps the resource which can be released calling Reset()
type AsyncHasher struct {
	*bmt.Hasher            // extends the Hasher
	mtx         sync.Mutex // to lock the cursor access
	double      bool       // whether to use double segments (call Hasher.writeSection)
	secsize     int        // size of base section (size of hash or double)
	seccount    int        // base section count
	write       func(i int, section []byte, final bool)
	errFunc     func(error)
	ctx         context.Context
	all         bool // if all written in one go, temporary workaround
}

func (sw *AsyncHasher) raiseError(err string) {
	if sw.errFunc != nil {
		sw.errFunc(errors.New(err))
	}
}

// Reset implements file.SectionWriter
func (sw *AsyncHasher) Reset() {
	sw.all = false
	sw.Hasher.Reset()
}

// SectionSize implements file.SectionWriter
func (sw *AsyncHasher) SectionSize() int {
	return sw.secsize
}

// Branches implements file.SectionWriter
func (sw *AsyncHasher) Branches() int {
	return sw.seccount
}

// WriteSection writes the i-th section of the BMT base
// this function can and is meant to be called concurrently
// it sets max segment threadsafely
func (sw *AsyncHasher) WriteIndexed(i int, section []byte) {
	sw.mtx.Lock()
	defer sw.mtx.Unlock()
	t := sw.GetTree()
	// cursor keeps track of the rightmost.GetSection() written so far
	// if index is lower than cursor then just write non-final section as is
	if i < sw.Hasher.GetCursor() {
		// if index is not the rightmost, safe to write section
		go sw.WriteSection(i, section, sw.double, false)
		return
	}
	// if there is a previous rightmost.GetSection() safe to write section
	if t.GetOffset() > 0 {
		if i == sw.Hasher.GetCursor() {
			// i==cursor implies cursor was set by Hash call so we can write section as final one
			// since it can be shorter, first we copy it to the padded buffer
			//t.GetSection() = make([]byte, sw.secsize)
			//copy(t.GetSection(), section)
			// TODO: Consider whether the section here needs to be copied, maybe we can enforce not change the original slice
			copySection := make([]byte, sw.secsize)
			copy(copySection, section)
			t.SetSection(copySection)
			go sw.Hasher.WriteSection(i, t.GetSection(), sw.double, true)
			return
		}
		// the rightmost section just changed, so we write the previous one as non-final
		go sw.WriteSection(sw.Hasher.GetCursor(), t.GetSection(), sw.double, false)
	}
	// set i as the index of the righmost.GetSection() written so far
	// set t.GetOffset() to cursor*secsize+1
	sw.Hasher.SetCursor(i)
	t.SetOffset(i*sw.secsize + 1)
	copySection := make([]byte, sw.secsize)
	copy(copySection, section)
	t.SetSection(copySection)
}

// Sum can be called any time once the length and the span is known
// potentially even before all segments have been written
// in such cases Sum will block until all segments are present and
// the hash for the length can be calculated.
//
// b: digest is appended to b
// length: known length of the input (unsafe; undefined if out of range)
// meta: metadata to hash together with BMT root for the final digest
//   e.g., span for protection against existential forgery
func (sw *AsyncHasher) SumIndexed(b []byte, length int) (s []byte) {
	sw.mtx.Lock()
	t := sw.GetTree()
	if length == 0 {
		sw.ReleaseTree()
		sw.mtx.Unlock()
		s = sw.Hasher.GetZeroHash()
		return
	} else {
		// for non-zero input the rightmost.GetSection() is written to the tree asynchronously
		// if the actual last.GetSection() has been written (sw.Hasher.GetCursor() == length/t.secsize)
		maxsec := (length - 1) / sw.secsize
		if t.GetOffset() > 0 {
			go sw.Hasher.WriteSection(sw.Hasher.GetCursor(), t.GetSection(), sw.double, maxsec == sw.Hasher.GetCursor())
		}
		// sesw.Hasher.GetCursor() to maxsec so final section is written when it arrives
		sw.Hasher.SetCursor(maxsec)
		t.SetOffset(length)
		// TODO: must this t.result channel be within lock?
		result := t.GetResult()
		sw.mtx.Unlock()
		// wait for the result or reset
		s = <-result
	}
	// relesase the tree back to the pool
	meta := t.GetSpan()
	sw.ReleaseTree()
	// hash together meta and BMT root hash using the pools
	hsh := sw.Hasher.GetHasher()
	hsh.Reset()
	hsh.Write(meta)
	hsh.Write(s)
	return hsh.Sum(b)
}
