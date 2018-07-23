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
	"crypto"
	"hash"

	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/storage/encryption"
)

// async hasher interface
type SectionHasherFunc func() SectionHasher

// the chunk hasher to link to each node in the filehasher
type ChunkHasherFunc func(data []byte, children [][]byte) SectionHasher

// ChunkStorer storage interface for the chunk hasher to use
type ChunkStorer interface {
	Store(addr []byte, meta []byte, data []byte)
}

type ChunkHasherParams struct {
	Encrypt      bool        // whether to encrypt chunk content
	Erasure      bool        // whether to inflace batches to support per-level CRS erasure codes
	Storer       ChunkStorer // the storer to call Store on to store a hashes chunk
	BaseHash     crypto.Hash // the base hash used by
	PoolCap      int         // BMT tree pool capacity
	SegmentCount int         // BMT base segment count on data level
	pool         *bmt.TreePool
}

func (p *ChunkHasherParams) getPool() *bmt.TreePool {
	if p.pool == nil {
		p.pool = bmt.NewTreePool(p.BaseHash.New, p.SegmentCount, p.PoolCap)
	}
	return p.pool
}

func (p *ChunkHasherParams) getHasherFunc() func() hash.Hash {
	return p.BaseHash.New
}

func (p *ChunkHasherParams) getHasher() SectionHasher {
	return bmt.New(p.getPool()).NewAsyncWriter(true)
}

// NewFileHasherStore creates a FileHasherStore object, which implements Putter and Getter interfaces.
// With the FileHasherStore you can put and get chunk data (which is just []byte) into a ChunkStore
// and the FileHasherStore will take core of encryption/decryption of data if necessary
func NewChunkHasherFunc(params ChunkHasherParams) ChunkHasherFunc {
	h := bmt.New(params.pool)
	f := func(data []byte, children [][]byte) SectionHasher {
		return params.getHasher()
	}
	// Storer is wraps the innermost SectionHasher if needed
	if params.Storer != nil {
		f = func(data []byte, children [][]byte) SectionHasher {
			return NewChunkHasherWithStorage(f(data, children), params.Storer, data)
		}
	}
	// encryption applied if needed (halves the capacity of batches)
	if params.Encrypt {
		f = func(data []byte, children [][]byte) SectionHasher {
			return NewChunkHasherWithEncryption(f(data, children), params.getHasherFunc())
		}
	}
	// Erasure coding as the outermost layer
	if params.Erasure {
		f = func(data []byte, children [][]byte) SectionHasher {
			// erasure := // takes children as arg
			return NewChunkHasherWithReundancy(f(data, children), nil)
		}
	}
	return f
}

// extensions of the base chunk hasher (SectionHasher interface)

// ChunkHasherWithRedundancy extends the ChunkHasher
// it completes a batch of child chunks using CRS erasure coding
// called when Sum is called on the chunk
type ChunkHasherWithRedundancy struct {
	SectionHasher
	erasure Erasure
}

// placeholder for per-batch redundancy coding
type Erasure interface{}

func NewChunkHasherWithReundancy(h SectionHasher, erasure Erasure) *ChunkHasherWithRedundancy {
	return &ChunkHasherWithRedundancy{
		SectionHasher: h,
		erasure:       erasure,
	}

}

// ChunkHasherWithStorage extends the ChunkHasher
// after Sum is called it stores the nodebuffer by calling put
type ChunkHasherWithStorage struct {
	SectionHasher
	ChunkStorer
	data []byte
}

// NewChunkHasherWithStorage returns a new ChunkHasherWithStorage
func NewChunkHasherWithStorage(h SectionHasher, cs ChunkStorer, data []byte) *ChunkHasherWithStorage {
	return &ChunkHasherWithStorage{
		SectionHasher: h,
		ChunkStorer:   cs,
		data:          data,
	}
}

// Sum calls the embedded hasher Sum
func (ch *ChunkHasherWithStorage) Sum(b []byte, length int, meta []byte) []byte {
	addr := ch.SectionHasher.Sum(b, length, meta)
	// this call will spawn a go routine
	ch.Store(addr, meta, ch.data)
	return addr
}

// ChunkHasherWithEncryption extends the ChunkHasher
// it encrypts the sections before writing them to the Hasher
// as well as encrypts the meta before calling Sum on the Hasher
// if segment count is 124 ChunkHasherWithEncryption reads
type ChunkHasherWithEncryption struct {
	SectionHasher
	encryptionKey []byte
	meta          encryption.Encryption
	data          encryption.Encryption
}

func NewChunkHasherWithEncryption(h SectionHasher, hasherFunc Hasher) *ChunkHasherWithEncryption {
	encryptionKey, err := encryption.GenerateRandomKey()
	if err != nil {
		panic(err.Error())
	}
	return &ChunkHasherWithEncryption{
		SectionHasher: h,
		encryptionKey: encryptionKey,
		// :FIXME:
		meta: encryption.New(encryptionKey, 0, uint32(h.ChunkSize()/2*h.Size()), hasherFunc),
		data: encryption.New(encryptionKey, int(h.ChunkSize()), 0, hasherFunc),
	}
}

// ChunkSize overrides ChunkSize N of the embedded ChunkHasher
// returns N/2 since sections will be doubled
func (ch *ChunkHasherWithEncryption) ChunkSize() int {
	return ch.SectionHasher.ChunkSize() / 2
}

// BlockSize returns the hash size of the underlying ChunkHasher
func (ch *ChunkHasherWithEncryption) BlockSize() int {
	return ch.SectionHasher.Size()
}

// Write uses the ChunkEncryption to encrypt the section
func (ch *ChunkHasherWithEncryption) Write(i int, plaintext []byte) {
	// :FIXME:
	encryptedAddr, err := ch.data.EncryptSegment(2*i, plaintext)
	if err != nil {
		panic(err.Error())
	}
	encryptedKey, err := ch.data.EncryptSegment(2*i+1, ch.encryptionKey)
	if err != nil {
		panic(err.Error())
	}
	ch.SectionHasher.Write(i*2, append(encryptedAddr, encryptedKey...))
}

// Sum encrypts meta
func (ch *ChunkHasherWithEncryption) Sum(b []byte, length int, meta []byte) []byte {
	ciphertext, err := ch.meta.Encrypt(meta, ch.encryptionKey)
	// pad and encrypt
	ch.SectionHasher.Sum(b, ch.SectionHasher.ChunkSize(), ciphertext)
	return append(ciphertext, ch.encryptionKey...)
}
