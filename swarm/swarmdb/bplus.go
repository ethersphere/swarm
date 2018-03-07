// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// This file is part of a SWARMDB fork of the go-ethereum library and has been adapted from
//  https://github.com/cznic/b/blob/master/btree.go
// which has the following LICENSE terms:
// --------------------------------------------------------------------
// Copyright (c) 2014 The b Authors. All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//   * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//   * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//   * Neither the names of the authors nor the names of the
//contributors may be used to endorse or promote products derived from
//this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
// --------------------------------------------------------------------

// General background:
//  B+ trees are an ordered data structure that are widely used in filesystems and databases.
//  B+ trees have "X" intermediate nodes (which point to X or D nodes) and "D" data nodes (which contain key-value pairs)
//  The D nodes at the bottom are in a doubly linked list, ordered by a primary key.

// SWARMDB uses B+ trees and Hash trees to store a dynamically changing tree of "keys", either of primary keys or secondarykey+primary key combinations
// The core of this B+ tree abstraction is to support a memory-based cache for SWARMDB chunks, where the SWARMDB Manager coordinates indexes of a table.
// The SWARMDB Manager relies on the indexes to support
// (1) "buffered" interactions, where a StartBuffer/FlushBuffer combination is required
// (2) "unbuffered" interactions, where Put/Get/Insert/Delete is executed immediately
// The X + D nodes of B+trees are mapped to and from 4K SWARMDB chunks in swarmPut/swarmGet calls,
// and the top level node of the B+ tree (and HashDB) is kept in a table root (itself updated with ENS)

// Loads up the top level node of the B+ tree only
// X nodes are SWARMDB chunks with 32 key-hashid combinations (64 bytes)
//   0: key - hashid
// ...
//  31: key - hashid
// where each hashid points to another X or D node
// at the bottom of each X node is the "parent" type and the "child type"

// D nodes are SWARMDB chunks with 32 key-hashid combinations (64 bytes)
//   0: key - hashid
// ...
//  31: key - hashid
// The hashid actually point to K nodes where raw records are stored (see kademliadb.go)
// At the bottom of each D  are prev/next chunk pointers

// The SWARMDB chunks are
//  - written via swarmdb.StoreDBChunk in swarmPut
//  - read via swarmdb.RetrieveDBChunk in swarmGet
// using the user object passed in the constructor, which has public/private keys needed for encryption/decryption
package swarmdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	//sdbc "swarmdbcommon"
	"io"
	"math"
	"strings"
	"sync"
)

// Tree is a B+tree.
type Tree struct {
	c          int
	cmp        Cmp
	cmpPrimary Cmp
	first      *d
	last       *d
	r          interface{}
	ver        int64

	swarmdb           *SwarmDB
	buffered          bool
	columnType        sdbc.ColumnType
	columnTypePrimary sdbc.ColumnType
	secondary         bool
	encrypted         int
	hashid            []byte
}

const (
	kx             = 3
	kd             = 3
	KEYS_PER_CHUNK = 32 // TODO - OPTIMIZE THIS based on X + D node Chunk structur
	KV_SIZE        = 64
	K_SIZE         = 32 // TODO - WHAT allows for bigger keys?
	V_SIZE         = 32
	HASH_SIZE      = 32
)

type (
	// Cmp compares a and b. Return value is:
	Cmp func(a, b []byte /*K*/) int
	//
	//	< 0 if a <  b
	//	  0 if a == b
	//	> 0 if a >  b
	//

	d struct { // data page
		c int
		d [2*kd + 1]de
		n *d
		p *d

		// used in open, insert, delete
		hashid    []byte
		dirty     bool
		notloaded bool

		// used for linked list traversal
		prevhashid []byte
		nexthashid []byte
	}

	de struct { // d element
		k []byte // interface{} /*K*/
		v []byte // interface{} /*V*/
	}

	// Enumerator captures the state of enumerating a tree. It is returned
	// from the Seek* methods. The enumerator is aware of any mutations
	// made to the tree in the process of enumerating it and automatically
	// resumes the enumeration at the proper key, if possible.
	//
	// However, once an Enumerator returns io.EOF to signal "no more
	// items", it does no more attempt to "resync" on tree mutation(s).  In
	// other words, io.EOF from an Enumerator is "sticky" (idempotent).
	Enumerator struct {
		err error
		hit bool
		i   int
		k   []byte /*K*/
		q   *d
		t   *Tree
		ver int64
	}

	xe struct { // x element
		ch interface{}
		k  []byte // interface{} /*K*/
	}

	x struct {
		c int
		x [2*kx + 2]xe

		// used in open, insert, delete
		hashid    []byte
		dirty     bool
		notloaded bool
	}
)

func init() {
	if kd < 1 {
		panic(fmt.Errorf("kd %d: out of range", kd))
	}

	if kx < 2 {
		panic(fmt.Errorf("kx %d: out of range", kx))
	}
}

var (
	btDPool = sync.Pool{New: func() interface{} { return &d{} }}
	btEPool = btEpool{sync.Pool{New: func() interface{} { return &Enumerator{} }}}
	btTPool = btTpool{sync.Pool{New: func() interface{} { return &Tree{} }}}
	btXPool = sync.Pool{New: func() interface{} { return &x{} }}
)

type btTpool struct{ sync.Pool }

func (p *btTpool) get(cmp Cmp, cmpPrimary Cmp) *Tree {
	x := p.Get().(*Tree)
	x.cmp = cmp
	x.cmpPrimary = cmpPrimary
	return x
}

type btEpool struct{ sync.Pool }

func (p *btEpool) get(err error, hit bool, i int, k []byte /*K*/, q *d, t *Tree, ver int64) *Enumerator {
	x := p.Get().(*Enumerator)
	x.err, x.hit, x.i, x.k, x.q, x.t, x.ver = err, hit, i, k, q, t, ver
	return x
}

var ( // R/O zero values
	zd  d
	zde de
	ze  Enumerator
	zk  []byte // interface{} /*K*/
	zt  Tree
	zx  x
	zxe xe
)

func clr(q interface{}) {
	switch x := q.(type) {
	case *x:
		for i := 0; i <= x.c; i++ { // Ch0 Sep0 ... Chn-1 Sepn-1 Chn
			clr(x.x[i].ch)
		}
		*x = zx
		btXPool.Put(x)
	case *d:
		*x = zd
		btDPool.Put(x)
	}
}

// -------------------------------------------------------------------------- x

func newX(ch0 interface{}) *x {
	r := btXPool.Get().(*x)
	r.x[0].ch = ch0
	return r
}

func (q *x) extract(i int) {
	q.c--
	if i < q.c {
		copy(q.x[i:], q.x[i+1:q.c+1])
		q.x[q.c].ch = q.x[q.c+1].ch
		q.x[q.c].k = zk  // GC
		q.x[q.c+1] = zxe // GC
	}
}

func (q *x) insert(i int, k []byte /*K*/, ch interface{}) *x {
	c := q.c
	if i < c {
		q.x[c+1].ch = q.x[c].ch
		copy(q.x[i+2:], q.x[i+1:c])
		q.x[i+1].k = q.x[i].k
	}
	c++
	q.c = c
	q.x[i].k = k
	q.x[i+1].ch = ch
	q.dirty = true
	return q
}

func (q *x) siblings(i int) (l, r *d) {
	if i >= 0 {
		if i > 0 {
			l = q.x[i-1].ch.(*d)
		}
		if i < q.c {
			r = q.x[i+1].ch.(*d)
		}
	}
	return
}

// -------------------------------------------------------------------------- d

func (l *d) mvL(r *d, c int) {
	copy(l.d[l.c:], r.d[:c])
	copy(r.d[:], r.d[c:r.c])
	l.c += c
	r.c -= c
}

func (l *d) mvR(r *d, c int) {
	copy(r.d[c:], r.d[:r.c])
	copy(r.d[:c], l.d[l.c-c:])
	r.c += c
	l.c -= c
}

// ----------------------------------------------------------------------- Tree

// BPlusTree returns a newly created, empty Tree. The compare function is used for key collation.
// The SWARMDB Manager instantiates a B+ tree with the top level SWARMDB hashid along with a specific user u
// that holds the public/private keys for writing chunks
// To manage ordering a "cmp" is instantiated with a suitable columnType (blob/float/string/integer/...)
//TODO: No error ever returned -- are we checking everything correctly?
func NewBPlusTreeDB(u *SWARMDBUser, swarmdb *SwarmDB, hashid []byte, columnType sdbc.ColumnType, secondary bool, columnTypePrimary sdbc.ColumnType, encrypted int) (t *Tree, err error) {
	// set up the comparator
	cmpPrimary := cmpBytes
	if secondary == true {
		switch columnTypePrimary {
		case sdbc.CT_FLOAT:
			cmpPrimary = cmpFloat
		case sdbc.CT_STRING:
			cmpPrimary = cmpString
		case sdbc.CT_INTEGER:
			cmpPrimary = cmpInt64
		}
	}

	switch columnType {
	case sdbc.CT_BLOB:
		t = btTPool.get(cmpBytes, cmpPrimary)
	case sdbc.CT_FLOAT:
		t = btTPool.get(cmpFloat, cmpPrimary)
	case sdbc.CT_STRING:
		t = btTPool.get(cmpString, cmpPrimary)
	case sdbc.CT_INTEGER:
		t = btTPool.get(cmpInt64, cmpPrimary)
	}
	t.columnType = columnType
	t.columnTypePrimary = columnType

	// root level
	t.hashid = hashid

	// used to SWARMGET
	t.swarmdb = swarmdb
	t.secondary = secondary
	t.encrypted = encrypted

	// get the top level node (only)
	t.swarmGet(u)
	return t, nil
}

// SWARMDB manager can buffer their operations with StartBuffer/FlushBuffer
func (t *Tree) StartBuffer(u *SWARMDBUser) (ok bool, err error) {
	t.buffered = true
	return true, nil
}

// when FlushBuffer is called (from the SWARMDB Manager), all the updated nodes in memory are "flushed out" to SWARM
func (t *Tree) FlushBuffer(u *SWARMDBUser) (ok bool, err error) {
	if t.buffered {
		t.buffered = false
		new_hashid, changed, errPut := t.swarmPut(u)
		if errPut != nil {
			return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[bplus:FlushBuffer] swarmPut - %s", errPut.Error()), ErrorCode: 471, ErrorMessage: "Failure encountered attempting to Flush nodes"}
		}
		if changed {
			t.hashid = new_hashid
			return true, nil
		}
	}
	return true, nil
}

// helper function
func (t *Tree) check_flush(u *SWARMDBUser) (ok bool, err error) {
	if t.buffered {
		return false, nil
	}

	new_hashid, changed, errPut := t.swarmPut(u)
	if errPut != nil {
		return false, sdbc.GenerateSWARMDBError(errPut, fmt.Sprintf("[bplus:check_flush] swarmPut - %s", errPut.Error()))
	}

	if changed {
		t.hashid = new_hashid
		return true, nil
	}
	return false, nil
}

// Close performs Clear and recycles t to a pool for possible later reuse. No
// references to t should exist or such references must not be used afterwards.
func (t *Tree) Close(u *SWARMDBUser) (ok bool, err error) {
	if t.buffered {
		ok, err = t.FlushBuffer(u)
		if err != nil {
			return false, sdbc.GenerateSWARMDBError(err, `[bplus:Close] FlushBuffer `+err.Error())
		}
	}
	t.Clear()
	*t = zt
	btTPool.Put(t)
	return true, nil
}

// --
func get_chunk_nodetype(buf []byte) (nodetype string) {
	return string(buf[CHUNK_SIZE-65 : CHUNK_SIZE-64])
}

func set_chunk_nodetype(buf []byte, nodetype string) {
	copy(buf[CHUNK_SIZE-65:], []byte(nodetype))
}

// --
func get_chunk_childtype(buf []byte) (nodetype string) {
	return string(buf[CHUNK_SIZE-66 : CHUNK_SIZE-65])
}

func set_chunk_childtype(buf []byte, nodetype string) {
	copy(buf[CHUNK_SIZE-66:], []byte(nodetype))
}

func (t *Tree) GetRootHash() (hashid []byte) {
	return t.hashid
}

func (t *Tree) swarmGet(u *SWARMDBUser) (success bool, err error) {
	if t.r != nil {
		switch z := t.r.(type) {
		case (*x):
			return z.swarmGet(u, t.swarmdb)
		case (*d):
			return z.swarmGet(u, t.swarmdb)
		}
	}

	// Core interface with DBChunkstore is here
	// do a read from local file system, filling in: (a) hashid and (b) items
	buf, err := t.swarmdb.RetrieveDBChunk(u, t.hashid)
	if err != nil {
		return false, sdbc.GenerateSWARMDBError(err, `[bplus:swarmGet] RetrieveDBChunk `+err.Error())
	}

	// two bytes
	nodetype := get_chunk_nodetype(buf)
	childtype := get_chunk_childtype(buf)
	if nodetype == "X" {
		// create X node
		t.r = btXPool.Get().(*x)
		switch z := t.r.(type) {
		case (*x):
			z.hashid = t.hashid
			for i := 0; i < KEYS_PER_CHUNK; i++ {
				k := buf[i*KV_SIZE : i*KV_SIZE+K_SIZE]
				hashid := buf[i*KV_SIZE+K_SIZE : i*KV_SIZE+KV_SIZE]
				if valid_hashid(hashid) && i < 2*kx+2 { //
					z.c++
					if childtype == "X" {
						x := btXPool.Get().(*x)
						z.x[i].ch = x
						z.x[i].k = k
						x.notloaded = true
						x.hashid = hashid
					} else if childtype == "D" {
						x := btDPool.Get().(*d)
						z.x[i].ch = x
						z.x[i].k = k
						x.notloaded = true
						x.hashid = hashid
					}
				}
			}
		}
	} else {
		// create D node
		t.r = btDPool.Get().(*d)
		switch z := t.r.(type) {
		case (*d):
			z.hashid = t.hashid
			for i := 0; i < KEYS_PER_CHUNK; i++ {
				k := buf[i*KV_SIZE : i*KV_SIZE+K_SIZE]
				hashid := buf[i*KV_SIZE+K_SIZE : i*KV_SIZE+KV_SIZE]
				if valid_hashid(hashid) && i < 2*kd {
					z.c++
					x := btDPool.Get().(*d)
					z.d[i].k = k
					z.d[i].v = hashid
					x.notloaded = true
					x.hashid = hashid
				}
			}
		}
	}

	switch z := t.r.(type) {
	case (*x):
		z.c--
	}
	return true, nil
}

// a hashid is valid if its not 0
func valid_hashid(hashid []byte) (valid bool) {
	valid = false
	for i := 0; i < len(hashid); i++ {
		if hashid[i] != 0 {
			return true
		}
	}
	return valid
}

func (q *x) swarmGet(u *SWARMDBUser, swarmdb DBChunkstorage) (success bool, err error) {
	if q.notloaded {
	} else {
		return false, nil
	}

	// do a read from local file system, filling in: (a) hashid and (b) items
	buf, err := swarmdb.RetrieveDBChunk(u, q.hashid)
	if err != nil {
		return false, sdbc.GenerateSWARMDBError(err, `[bplus:swarmGet] RetrieveDBChunk `+err.Error())
	}

	childtype := get_chunk_childtype(buf)
	for i := 0; i < KEYS_PER_CHUNK; i++ {
		if i < 2*kx+2 {
			k := buf[i*KV_SIZE : i*KV_SIZE+K_SIZE]
			hashid := buf[i*KV_SIZE+K_SIZE : i*KV_SIZE+KV_SIZE]
			if valid_hashid(hashid) {
				if childtype == "X" {
					x := btXPool.Get().(*x)
					q.x[i].ch = x
					q.x[i].k = k
					x.notloaded = true
					x.hashid = hashid
				} else if childtype == "D" {
					q.c++
					x := btDPool.Get().(*d)
					q.x[i].ch = x
					q.x[i].k = []byte(k)
					x.notloaded = true
					x.hashid = hashid
				}
			}
		}
	}
	q.notloaded = false
	return true, nil
}

func (q *d) swarmGet(u *SWARMDBUser, swarmdb DBChunkstorage) (success bool, err error) {
	if q.notloaded {
	} else {
		return false, nil
	}

	// do a read from local file system, filling in: (a) hashid and (b) items
	buf, err := swarmdb.RetrieveDBChunk(u, q.hashid)
	if err != nil {
		return false, sdbc.GenerateSWARMDBError(err, `[bplus:swarmGet] RetrieveDBChunk `+err.Error())
	}
	for i := 0; i < KEYS_PER_CHUNK; i++ {
		k := buf[i*KV_SIZE : i*KV_SIZE+K_SIZE]
		hashid := buf[i*KV_SIZE+K_SIZE : i*KV_SIZE+KV_SIZE]
		//fmt.Printf(" LOAD-C|%d (%d)|%x|%v\n", i, 2*kd+1, hashid, valid_hashid(hashid))
		if valid_hashid(hashid) && (i < 2*kd+1) {
			q.c++
			q.d[i].k = k
			q.d[i].v = hashid
			q.notloaded = false
		}
	}
	q.prevhashid = buf[CHUNK_SIZE-HASH_SIZE*2 : CHUNK_SIZE-HASH_SIZE]
	q.nexthashid = buf[CHUNK_SIZE-HASH_SIZE : CHUNK_SIZE]
	q.notloaded = false
	return true, nil
}

func (t *Tree) swarmPut(u *SWARMDBUser) (new_hashid []byte, changed bool, err error) {
	q := t.r
	if q == nil {
		return
	}

	switch x := q.(type) {
	case *x: // intermediate node -- descend on the next pass
		// fmt.Printf("ROOT XNode %x [dirty=%v|notloaded=%v]\n", x.hashid, x.dirty, x.notloaded)
		var errPut error
		new_hashid, changed, errPut = x.swarmPut(u, t.swarmdb, t.columnType, t.encrypted)
		if errPut != nil {
			return new_hashid, changed, err
		}
		if changed {
			t.hashid = x.hashid
		}
	case *d: // data node -- EXACT match
		// fmt.Printf("ROOT DNode %x [dirty=%v|notloaded=%v]\n", x.hashid, x.dirty, x.notloaded)
		new_hashid, changed, err = x.swarmPut(u, t.swarmdb, t.columnType, t.encrypted)
		if changed {
			t.hashid = x.hashid
		}
	}

	return new_hashid, changed, nil
}

func (q *x) swarmPut(u *SWARMDBUser, swarmdb DBChunkstorage, columnType sdbc.ColumnType, encrypted int) (new_hashid []byte, changed bool, err error) {
	// recurse through children
	// fmt.Printf("put XNode [c=%d] %x [dirty=%v|notloaded=%v]\n", q.c, q.hashid, q.dirty, q.notloaded)
	for i := 0; i <= q.c; i++ {
		switch z := q.x[i].ch.(type) {
		case *x:
			if z.dirty {
				_, _, err = z.swarmPut(u, swarmdb, columnType, encrypted)
				if err != nil {
					return new_hashid, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:swarmPut] swarmPut - %s", err.Error()))
				}
			}
		case *d:
			if z.dirty {
				_, _, err = z.swarmPut(u, swarmdb, columnType, encrypted)
				if err != nil {
					return new_hashid, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:swarmPut] swarmPut - %s", err.Error()))
				}

			}
		}
	}

	// compute the data here
	sdata := make([]byte, CHUNK_SIZE)
	childtype := "X"
	for i := 0; i <= q.c; i++ {
		switch z := q.x[i].ch.(type) {
		case *x:
			copy(sdata[i*KV_SIZE:], q.x[i].k)        // max 32 bytes
			copy(sdata[i*KV_SIZE+K_SIZE:], z.hashid) // max 32 bytes
		case *d:
			copy(sdata[i*KV_SIZE:], q.x[i].k)        // max 32 bytes
			copy(sdata[i*KV_SIZE+K_SIZE:], z.hashid) // max 32 bytes
			childtype = "D"
		}
	}

	set_chunk_nodetype(sdata, "X")
	set_chunk_childtype(sdata, childtype)

	// Core interface with DBChunkstore is here
	new_hashid, err = swarmdb.StoreDBChunk(u, sdata, encrypted)
	if err != nil {
		return q.hashid, false, sdbc.GenerateSWARMDBError(err, `[bplus:swarmPut] StoreDBChunk `+err.Error())
	}
	q.hashid = new_hashid
	return new_hashid, true, nil
}

func (q *d) swarmPut(u *SWARMDBUser, swarmdb DBChunkstorage, columnType sdbc.ColumnType, encrypted int) (new_hashid []byte, changed bool, err error) {
	// fmt.Printf("put DNode [c=%d] [dirty=%v|notloaded=%v, prev=%x, next=%x]\n", q.c, q.dirty, q.notloaded, q.prevhashid, q.nexthashid)
	if q.n != nil {
		if q.n.dirty {
			_, _, err = q.n.swarmPut(u, swarmdb, columnType, encrypted)
			if err != nil {
				return new_hashid, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:swarmPut] swarmPut - %s", err.Error()))
			}
		}
		q.nexthashid = q.n.hashid
		// fmt.Printf(" -- NEXT: %x [%v]\n", q.nexthashid, q.n.dirty)
	}
	q.dirty = false

	if q.p != nil {
		if q.p.dirty {
			_, _, err = q.p.swarmPut(u, swarmdb, columnType, encrypted)
			if err != nil {
				return new_hashid, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:swarmPut] swarmPut - %s", err.Error()))
			}
		}
		q.prevhashid = q.p.hashid
		// fmt.Printf(" -- PREV: %x [%v]\n", q.prevhashid, q.p.dirty)
	}
	// fmt.Printf("N: %x P: %x\n", q.n, q.p) //  q.prevhashid, q.nexthashid

	sdata := make([]byte, CHUNK_SIZE)
	for i := 0; i < q.c; i++ {
		// fmt.Printf("STORE-C|%d|%s|%x\n", i, KeyToString(columnType, q.d[i].k), q.d[i].v)
		copy(sdata[i*KV_SIZE:], q.d[i].k)        // max 32 bytes
		copy(sdata[i*KV_SIZE+K_SIZE:], q.d[i].v) // max 32 bytes
	}

	if q.p != nil {
		copy(sdata[CHUNK_SIZE-HASH_SIZE*2:], q.prevhashid) // 32 bytes
	}
	if q.n != nil {
		copy(sdata[CHUNK_SIZE-HASH_SIZE*2+HASH_SIZE:], q.nexthashid) // 32 bytes
	}

	set_chunk_nodetype(sdata, "D")
	set_chunk_childtype(sdata, "C")

	// Core interface with DBChunkstore is here
	new_hashid, err = swarmdb.StoreDBChunk(u, sdata, encrypted)
	if err != nil {
		return q.hashid, false, sdbc.GenerateSWARMDBError(err, `[bplus:swarmPut] StoreDBChunk `+err.Error())
	}
	q.hashid = new_hashid

	return new_hashid, true, nil
}

// Clear removes all K/V pairs from the tree.
func (t *Tree) Clear() {
	if t.r == nil {
		return
	}

	clr(t.r)
	t.c, t.first, t.last, t.r = 0, nil, nil, nil
	t.ver++
}

func (t *Tree) cat(p *x, q, r *d, pi int) {
	t.ver++
	q.mvL(r, r.c)
	if r.n != nil {
		r.n.p = q
	} else {
		t.last = q
	}
	q.n = r.n
	r.dirty = true
	q.dirty = true
	*r = zd
	btDPool.Put(r)
	if p.c > 1 {
		p.extract(pi)
		p.x[pi].ch = q
		return
	}

	switch x := t.r.(type) {
	case *x:
		*x = zx
		btXPool.Put(x)
	case *d:
		*x = zd
		btDPool.Put(x)
	}
	t.r = q
}

func (t *Tree) catX(p, q, r *x, pi int) {
	t.ver++
	q.x[q.c].k = p.x[pi].k
	copy(q.x[q.c+1:], r.x[:r.c])
	q.c += r.c + 1
	q.x[q.c].ch = r.x[r.c].ch
	*r = zx
	btXPool.Put(r)
	if p.c > 1 {
		p.c--
		pc := p.c
		if pi < pc {
			p.x[pi].k = p.x[pi+1].k
			copy(p.x[pi+1:], p.x[pi+2:pc+1])
			p.x[pc].ch = p.x[pc+1].ch
			p.x[pc].k = zk     // GC
			p.x[pc+1].ch = nil // GC
		}
		return
	}

	switch x := t.r.(type) {
	case *x:
		*x = zx
		btXPool.Put(x)
	case *d:
		*x = zd
		btDPool.Put(x)
	}
	t.r = q
}

// Delete removes the k's KV pair, if it exists, in which case Delete returns true.
// TODO: the Delete => underflow situation  needs full testing
func (t *Tree) Delete(u *SWARMDBUser, k []byte /*K*/) (ok bool, err error) {
	pi := -1
	var p *x
	q := t.r
	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Delete] checkload - %s", err.Error()))

		}
		var i int
		i, ok = t.find(q, k)
		if ok {

			switch x := q.(type) {
			case *x:
				if x.c < kx && q != t.r {
					x, i = t.underflowX(p, x, pi, i)
				}
				pi = i + 1
				p = x
				q = x.x[pi].ch
				x.dirty = true // optimization: this should really be if something is *actually* deleted
				continue
			case *d:
				t.extract(x, i)
				if x.c >= kd {
					return true, nil
				}

				if q != t.r {
					t.underflow(p, x, pi)
				} else if t.c == 0 {

					t.Clear()
				}
				x.dirty = true // we found the key and  actually deleted it!
				_, err = t.check_flush(u)
				if err != nil {
					return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Delete] check_flush - %s", err.Error()))
				}
				return true, nil
			}
		}

		switch x := q.(type) {
		case *x:
			if x.c < kx && q != t.r {
				x, i = t.underflowX(p, x, pi, i)
			}
			pi = i
			p = x
			q = x.x[i].ch
			x.dirty = true // optimization: this should really be if something is *actually* deleted
		case *d:
			return false, nil // we got to the bottom and key was not found
		}
	}
}

func (t *Tree) extract(q *d, i int) { // (r interface{} /*V*/) {
	t.ver++
	//r = q.d[i].v // prepared for Extract
	q.c--
	if i < q.c {
		copy(q.d[i:], q.d[i+1:q.c+1])
	}
	q.d[q.c] = zde // GC
	t.c--
	return
}

// find - does a binary search in an X or D node to find the index [ok is true if found]
func (t *Tree) find(q interface{}, k []byte /*K*/) (i int, ok bool) {
	var mk []byte /*K*/
	l := 0
	switch x := q.(type) {
	case *x:
		h := x.c - 1
		for l <= h {
			m := (l + h) >> 1
			mk = x.x[m].k
			switch cmp := t.cmp(k, mk); {
			case cmp > 0:
				l = m + 1
			case cmp == 0:
				return m, true
			default:
				h = m - 1
			}
		}
	case *d:
		h := x.c - 1
		for l <= h {
			m := (l + h) >> 1
			mk = x.d[m].k
			//fmt.Printf(" d node (c=%d): m=%d t.cmp( k=(%s),  mk=(%s)) = %d \n", x.c, m,  KeyToString(t.columnType, k), KeyToString(t.columnType, mk), t.cmp(k, mk))
			switch cmp := t.cmp(k, mk); {
			case cmp > 0:
				l = m + 1
			case cmp == 0:
				return m, true
			default:
				h = m - 1
			}
		}
	}
	return l, false
}

// This is a helper function called by Get/.. to support lazy loading -- if the node you are processing is notloaded, then load it!
func checkload(u *SWARMDBUser, swarmdb DBChunkstorage, q interface{}) (err error) {
	switch x := q.(type) {
	case *x: // intermediate node -- descend on the next pass
		if x.notloaded {
			_, err = x.swarmGet(u, swarmdb)
			if err != nil {
				return &sdbc.SWARMDBError{Message: fmt.Sprintf("[bplus:checkload] swarmGet - %s", err.Error()), ErrorCode: 473, ErrorMessage: "Failure encountered checking load"}
			}
		}
	case *d: // data node -- EXACT match
		if x.notloaded {
			x.swarmGet(u, swarmdb)
			if err != nil {
				return &sdbc.SWARMDBError{Message: fmt.Sprintf("[bplus:checkload] swarmGet - %s", err.Error()), ErrorCode: 473, ErrorMessage: "Failure encountered checking load"}
			}
		}
	}
	return nil
}

// Get returns the value associated with k and true if it exists. Otherwise Get
// returns (zero-value, false).
func (t *Tree) Get(u *SWARMDBUser, key []byte /*K*/) (v []byte /*V*/, ok bool, err error) {
	log.Debug("[bplus:Get]")
	q := t.r
	//	if q == nil {
	//		return
	//	}

	k := make([]byte, K_SIZE)
	copy(k, key)

	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return v, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Get] checkload - %s", err.Error()))
		}

		var i int

		// binary search on the node => i
		if i, ok = t.find(q, k); ok {
			// found it
			switch x := q.(type) {
			case *x: // intermediate node -- descend on the next pass
				q = x.x[i+1].ch
				continue
			case *d: // data node -- EXACT match
				// fmt.Printf("*************** EUREKA %v\n", x.d[i].v)
				return x.d[i].v, true, nil
			}
		}
		// descend down the tree using the binary search
		switch x := q.(type) {
		case *x:
			//fmt.Printf(" X not FOUND (%d) i:%d k:[%s]\n", i, t.columnType, KeyToString(t.columnType, k))
			q = x.x[i].ch
		default:
			//fmt.Printf(" D not FOUND (%d) i:%d k:[%s]\n", i, t.columnType, KeyToString(t.columnType, k))
			return zk, false, nil
			//TODO: Does this just mean that it's "empty"?  If so, then I think we should just say "true" (or not?)
		}
	}
}

// This actually inserts
func (t *Tree) insert(q *d, i int, k []byte /*K*/, v []byte /*V*/) *d {
	t.ver++
	c := q.c
	if i < c {
		copy(q.d[i+1:], q.d[i:c])
	}
	c++
	q.c = c
	q.d[i].k = k
	q.d[i].v = v
	t.c++
	q.dirty = true
	return q
}

func print_spaces(nspaces int) {
	for i := 0; i < nspaces; i++ {
		fmt.Printf("  ")
	}
}

// NOTE: this only prints the portion of the tree that is actually LOADED
func (t *Tree) Print(u *SWARMDBUser) {
	q := t.r
	if q == nil {
		return
	}

	switch x := q.(type) {
	case *x: // intermediate node -- descend on the next pass
		fmt.Printf("ROOT Node (X) [%x] [dirty=%v|notloaded=%v]\n", x.hashid, x.dirty, x.notloaded)
		x.print(t.columnType, 0)
	case *d: // data node -- EXACT match
		fmt.Printf("ROOT Node (D) [%x] [dirty=%v|notloaded=%v]\n", x.hashid, x.dirty, x.notloaded)
		x.print(t.columnType, 0)
	}
	return
}

func (q *x) print(columnType sdbc.ColumnType, level int) {
	print_spaces(level)
	fmt.Printf("XNode (%x) [c=%d] (LEVEL %d) [dirty=%v|notloaded=%v]\n", q.hashid, q.c, level, q.dirty, q.notloaded)
	if q.notloaded == false {
		for i := 0; i <= q.c; i++ {
			print_spaces(level + 1)
			fmt.Printf("Child %d|%v\n", i, level+1) // , KeyToString(columnType, q.x[i].k))
			switch z := q.x[i].ch.(type) {
			case *x:
				z.print(columnType, level+1)
			case *d:
				z.print(columnType, level+1)
			}
		}
	}

	return
}

func (q *d) print(columnType sdbc.ColumnType, level int) {
	print_spaces(level)
	fmt.Printf("DNode %x [c=%d] (LEVEL %d) [dirty=%v|notloaded=%v|prev=%x|next=%x]\n", q.hashid, q.c, level, q.dirty, q.notloaded, q.prevhashid, q.nexthashid)
	for i := 0; i < q.c; i++ {
		print_spaces(level + 1)
		fmt.Printf("DATA %d (L%d)|%s|%v\n", i, level+1, KeyToString(columnType, q.d[i].k), ValueToString(q.d[i].v))
	}
	return
}

func (t *Tree) overflow(p *x, q *d, pi, i int, k []byte /*K*/, v []byte /*V*/) {
	t.ver++
	l, r := p.siblings(pi)

	if l != nil && l.c < 2*kd && i != 0 {
		l.dirty = true
		l.mvL(q, 1)
		t.insert(q, i-1, k, v)
		p.x[pi-1].k = q.d[0].k
		return
	}

	if r != nil && r.c < 2*kd {
		r.dirty = true
		if i < 2*kd {
			q.mvR(r, 1)
			t.insert(q, i, k, v)
			p.x[pi].k = r.d[0].k
			return
		}

		t.insert(r, 0, k, v)
		p.x[pi].k = k
		return
	}

	t.split(p, q, pi, i, k, v)
}

// Seek returns an Enumerator positioned on an item such that k >= item's key.
// ok reports if k == item.key The Enumerator's position is possibly after the
// last item in the tree.
func (t *Tree) Seek(u *SWARMDBUser, key []byte /*K*/) (e OrderedDatabaseCursor, ok bool, err error) {
	k := make([]byte, K_SIZE)
	copy(k, key)

	q := t.r
	if q == nil {
		e = btEPool.get(nil, false, 0, k, nil, t, t.ver)
		return
	}

	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return e, false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Seek] checkload - %s", err.Error()))
		}

		var i int
		if i, ok = t.find(q, k); ok {
			switch x := q.(type) {
			case *x:
				q = x.x[i+1].ch
				continue
			case *d: // err, hit, i, k, q, t, ver
				return btEPool.get(nil, ok, i, k, x, t, t.ver), true, nil
			}
		}

		switch x := q.(type) {
		case *x:
			q = x.x[i].ch
		case *d:
			return btEPool.get(nil, ok, i, k, x, t, t.ver), false, nil
		}
	}
	return e, false, nil
}

func (t *Tree) SeekFirst(u *SWARMDBUser) (e OrderedDatabaseCursor, err error) {
	k := make([]byte, K_SIZE)
	q := t.r
	if q == nil {
		e = btEPool.get(nil, false, 0, k, nil, t, t.ver)
		return
	}

	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return e, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:SeekFirst] checkload - %s", err.Error()))
		}

		var i int

		i = 0
		switch x := q.(type) {
		case *x:
			q = x.x[i].ch
		case *d:
			return btEPool.get(nil, true, i, k, x, t, t.ver), nil
		}
	}
	return e, nil
}

func (t *Tree) SeekLast(u *SWARMDBUser) (e OrderedDatabaseCursor, err error) {
	k := make([]byte, K_SIZE)
	q := t.r
	if q == nil {
		e = btEPool.get(nil, false, 0, k, nil, t, t.ver)
		return
	}

	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return e, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:SeekLast] checkload - %s", err.Error()))
		}

		switch x := q.(type) {
		case *x:
			q = x.x[x.c].ch
		case *d:
			i := x.c - 1
			return btEPool.get(nil, true, i, k, x, t, t.ver), nil
		}
	}
	return e, nil
}

// Put(k,v) -- actually puts the key
// TODO: add checks for byte length input on key/value
func (t *Tree) Put(u *SWARMDBUser, key []byte /*K*/, v []byte /*V*/) (okresult bool, err error) {
	// fmt.Printf(" -- B+ Tree Put: %s => %s\n", KeyToString(t.columnType, key), ValueToString(v))
	k := make([]byte, K_SIZE)
	copy(k, key)

	pi := -1
	var p *x
	q := t.r
	if q == nil {
		// returns a "d" element which is a linked list (c = int, d array of data elements, n (next), p (prev)
		z := t.insert(btDPool.Get().(*d), 0, k, v)
		t.r, t.first, t.last = z, z, z
		return
	}

	// go down each level, from the "x" intermediate nodes to the "d" data nodes
	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Put] checkload - %s", err.Error()))
		}

		i, ok := t.find(q, k)
		if ok {
			// the key is found
			switch x := q.(type) {
			case *x:
				// for the intermediate level
				i++
				if x.c > 2*kx {
					x, i = t.splitX(p, x, pi, i)
				}
				pi = i
				p = x
				q = x.x[i].ch
				continue
			case *d:
				x.d[i].v = v
				x.dirty = true // we updated the value but did not insert anything
				_, err = t.check_flush(u)
				if err != nil {
					return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Put] check_flush - %s", err.Error()))
				}
				return true, nil
			}
		}

		switch x := q.(type) {
		case *x:
			if x.c > 2*kx {
				x, i = t.splitX(p, x, pi, i)
			}
			pi = i
			p = x
			q = x.x[i].ch
			x.dirty = true // we updated the value at the intermediate node
		case *d:
			switch {
			case x.c < 2*kd: // insert
				t.insert(x, i, k, v)
			default:
				t.overflow(p, x, pi, i, k, v)
			}
			x.dirty = true // we inserted the value at the intermediate node or leaf node
			_, err = t.check_flush(u)
			if err != nil {
				return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Put] check_flush - %s", err.Error()))
			}
			return true, nil
		}
	}
}

func (t *Tree) Insert(u *SWARMDBUser, k []byte /*K*/, v []byte /*V*/) (okres bool, err error) {
	pi := -1
	var p *x
	q := t.r
	if q == nil {
		// returns a "d" element which is a linked list (c = int, d array of data elements, n (next), p (prev)
		z := t.insert(btDPool.Get().(*d), 0, k, v)
		t.r, t.first, t.last = z, z, z
		return true, nil
	}

	// go down each level, from the "x" intermediate nodes to the "d" data nodes
	for {
		err = checkload(u, t.swarmdb, q)
		if err != nil {
			return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Insert] checkload - %s", err.Error()))
		}

		i, ok := t.find(q, k)
		if ok {
			var dkerr *sdbc.DuplicateKeyError
			return false, dkerr
		}

		switch x := q.(type) {
		case *x:
			if x.c > 2*kx {
				x, i = t.splitX(p, x, pi, i)
			}
			pi = i
			p = x
			q = x.x[i].ch
			x.dirty = true // we updated the value at the intermediate node
		case *d:
			switch {
			case x.c < 2*kd: // insert
				t.insert(x, i, k, v)
			default:
				t.overflow(p, x, pi, i, k, v)
			}
			x.dirty = true // we inserted the value at the intermediate node or leaf node
			_, err = t.check_flush(u)
			if err != nil {
				return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:Insert] check_flush - %s", err.Error()))
			}

			return true, nil
		}
	}
}

func (t *Tree) split(p *x, q *d, pi, i int, k []byte /*K*/, v []byte /*V*/) {
	t.ver++
	r := btDPool.Get().(*d)
	if q.n != nil {
		// insert new node into linked list
		r.n = q.n // new node "next" points to old node "next"
		r.n.p = r // new node "prev" points to old node
	} else {
		// its the last node of the linked list!
		t.last = r
	}
	q.n = r // old node "next" points to new node
	r.p = q // new node "prev" points to prev node
	r.dirty = true
	q.dirty = true

	copy(r.d[:], q.d[kd:2*kd])
	for i := range q.d[kd:] {
		q.d[kd+i] = zde
	}
	q.c = kd
	r.c = kd
	var done bool
	if i > kd {
		done = true
		t.insert(r, i-kd, k, v)
	}
	if pi >= 0 {
		p.insert(pi, r.d[0].k, r)
	} else {
		t.r = newX(q).insert(0, r.d[0].k, r)
	}
	if done {
		return
	}

	t.insert(q, i, k, v)
}

func (t *Tree) splitX(p *x, q *x, pi int, i int) (*x, int) {
	t.ver++
	r := btXPool.Get().(*x)
	copy(r.x[:], q.x[kx+1:])
	q.c = kx
	r.c = kx
	r.dirty = true
	if pi >= 0 {
		p.insert(pi, q.x[kx].k, r)
	} else {
		t.r = newX(q).insert(0, q.x[kx].k, r)
	}

	q.x[kx].k = zk
	for i := range q.x[kx+1:] {
		q.x[kx+i+1] = zxe
	}
	if i > kx {
		q = r
		i -= kx + 1
	}

	return q, i
}

func (t *Tree) underflow(p *x, q *d, pi int) {
	t.ver++
	l, r := p.siblings(pi)

	if l != nil && l.c+q.c >= 2*kd {
		l.mvR(q, 1)
		p.x[pi-1].k = q.d[0].k
		return
	}

	if r != nil && q.c+r.c >= 2*kd {
		q.mvL(r, 1)
		p.x[pi].k = r.d[0].k
		r.d[r.c] = zde // GC
		return
	}

	if l != nil {
		t.cat(p, l, q, pi-1)
		return
	}

	t.cat(p, q, r, pi)
}

func (t *Tree) underflowX(p *x, q *x, pi int, i int) (*x, int) {
	t.ver++
	var l, r *x

	if pi >= 0 {
		if pi > 0 {
			l = p.x[pi-1].ch.(*x)
		}
		if pi < p.c {
			r = p.x[pi+1].ch.(*x)
		}
	}

	if l != nil && l.c > kx {
		q.x[q.c+1].ch = q.x[q.c].ch
		copy(q.x[1:], q.x[:q.c])
		q.x[0].ch = l.x[l.c].ch
		q.x[0].k = p.x[pi-1].k
		q.c++
		i++
		l.c--
		p.x[pi-1].k = l.x[l.c].k
		return q, i
	}

	if r != nil && r.c > kx {
		q.x[q.c].k = p.x[pi].k
		q.c++
		q.x[q.c].ch = r.x[0].ch
		p.x[pi].k = r.x[0].k
		copy(r.x[:], r.x[1:r.c])
		r.c--
		rc := r.c
		r.x[rc].ch = r.x[rc+1].ch
		r.x[rc].k = zk
		r.x[rc+1].ch = nil
		return q, i
	}

	if l != nil {
		i += l.c + 1
		t.catX(p, l, q, pi-1)
		q = l
		return q, i
	}

	t.catX(p, q, r, pi)
	return q, i
}

// ----------------------------------------------------------------- Enumerator

// Close recycles e to a pool for possible later reuse. No references to e
// should exist or such references must not be used afterwards.
func (e *Enumerator) Close() {
	*e = ze
	btEPool.Put(e)
}

// Next returns the currently enumerated item, if it exists and moves to the
// next item in the key collation order. If there is no item to return, err ==
// io.EOF is returned.
func (e *Enumerator) Next(u *SWARMDBUser) (k []byte /*K*/, v []byte /*V*/, err error) {
	if err = e.err; err != nil {
		return
	}

	if e.q == nil {
		e.err, err = io.EOF, io.EOF
		return
	}

	if e.i >= e.q.c {
		if err = e.next(u); err != nil {
			// TODO: err handling
			return
		}
	}

	i := e.q.d[e.i]
	k, v = i.k, i.v
	e.k, e.hit = k, true
	e.next(u)
	return
}

func (e *Enumerator) next(u *SWARMDBUser) (err error) {
	if e.q == nil {
		e.err = io.EOF
		return io.EOF
	}

	switch {
	case e.i < e.q.c-1:
		e.i++
	default:
		if valid_hashid(e.q.nexthashid) && e.q.n == nil {
			r := btDPool.Get().(*d)
			r.p = e.q
			r.hashid = e.q.nexthashid
			r.notloaded = true
			_, err = r.swarmGet(u, e.t.swarmdb)
			if err != nil {
				return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:next] swarmGet - %s", err.Error()))
			}
			e.q = r
			e.i = 0
		} else {
			if e.q, e.i = e.q.n, 0; e.q == nil {
				e.err = io.EOF
			}
		}
	}

	return e.err
}

// Prev returns the currently enumerated item, if it exists and moves to the
// previous item in the key collation order. If there is no item to return, err
// == io.EOF is returned.
func (e *Enumerator) Prev(u *SWARMDBUser) (k []byte /*K*/, v []byte /*V*/, err error) {
	if err = e.err; err != nil {
		return
	}
	if e.q == nil {
		e.err, err = io.EOF, io.EOF
		return
	}

	if !e.hit {
		// move to previous because Seek overshoots if there's no hit
		if err = e.prev(u); err != nil {
			return
		}
	}

	if e.i >= e.q.c {
		if err = e.prev(u); err != nil {
			return
		}
	}

	i := e.q.d[e.i]
	k, v = i.k, i.v
	e.k, e.hit = k, true
	e.prev(u)
	return
}

func (e *Enumerator) prev(u *SWARMDBUser) (err error) {
	if e.q == nil {
		e.err = io.EOF
		return io.EOF
	}

	switch {
	case e.i > 0:
		e.i--
	default:
		if valid_hashid(e.q.prevhashid) && e.q.p == nil {
			r := btDPool.Get().(*d)
			r.hashid = e.q.prevhashid
			r.notloaded = true
			_, err = r.swarmGet(u, e.t.swarmdb)
			if err != nil {
				return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[bplus:prev] swarmGet - %s", err.Error()))
			}
			e.q = r
			if r.c >= 0 {
				e.i = r.c - 1
			} else {
				e.i = 0
			}
		} else {
			if e.q = e.q.p; e.q == nil {
				e.err = io.EOF
				break
			}
			e.i = e.q.c
		}
	}
	return e.err
}

// ----- COMPARATORS -- depending on the tree column Type, one of these will be used
func cmpBytes(a, b []byte) int {
	// Compare returns an integer comparing two byte slices lexicographically.
	// The result will be 0 if a==b, -1 if a < b, and +1 if a > b. A nil argument is equivalent to an empty slice.
	return bytes.Compare(a, b)
}

func cmpString(a, b []byte) int {
	as := string(a)
	bs := string(b)
	return strings.Compare(as, bs)
}

func cmpFloat(a, b []byte) int {
	abits := binary.BigEndian.Uint64(a)
	af := math.Float64frombits(abits)

	bbits := binary.BigEndian.Uint64(b)
	bf := math.Float64frombits(bbits)

	if af < bf {
		return -1
	} else if af > bf {
		return +1
	} else {
		return 0
	}
}

// ints are 64 bit / 8 byte
func cmpInt64(a, b []byte) int {
	ai := binary.BigEndian.Uint64(a)
	bi := binary.BigEndian.Uint64(b)
	if ai < bi {
		return -1
	} else if ai > bi {
		return +1
	} else {
		return 0
	}
}
