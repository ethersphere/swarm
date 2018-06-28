// Copyright 2018 The go-ethereum Authors
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
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/storage/encryption"
)

type chunkEncryption struct {
	spanEncryption encryption.Encryption
	dataEncryption encryption.Encryption
}

type hasherStore struct {
	store           ChunkStore
	hashFunc        SwarmHasher
	chunkEncryption *chunkEncryption
	hashSize        int           // content hash size
	refSize         int64         // reference size (content hash + possibly encryption key)
	count           uint64        // number of chunks to store
	errC            chan error    // global error channel
	doneC           chan struct{} // closed by Close() call to indicate that count is the final number of chunks
	quitC           chan struct{} // closed to quit unterminated routines
}

func newChunkEncryption(chunkSize, refSize int64) *chunkEncryption {
	return &chunkEncryption{
		spanEncryption: encryption.New(0, uint32(chunkSize/refSize), sha3.NewKeccak256),
		dataEncryption: encryption.New(int(chunkSize), 0, sha3.NewKeccak256),
	}
}

// NewHasherStore creates a hasherStore object, which implements Putter and Getter interfaces.
// With the HasherStore you can put and get chunk data (which is just []byte) into a DPA
// and the hasherStore will take core of encryption/decryption of data if necessary
func NewHasherStore(store ChunkStore, hashFunc SwarmHasher, toEncrypt bool) *hasherStore {
	var chunkEncryption *chunkEncryption

	hashSize := hashFunc().Size()
	refSize := int64(hashSize)
	if toEncrypt {
		refSize += encryption.KeyLength
		chunkEncryption = newChunkEncryption(DefaultChunkSize, refSize)
	}

	h := &hasherStore{
		store:           store,
		hashFunc:        hashFunc,
		chunkEncryption: chunkEncryption,
		hashSize:        hashSize,
		refSize:         refSize,
		errC:            make(chan error),
		doneC:           make(chan struct{}),
		quitC:           make(chan struct{}),
	}

	return h
}

// Put stores the chunkData into the ChunkStore of the hasherStore and returns the reference.
// If hasherStore has a chunkEncryption object, the data will be encrypted.
// Asynchronous function, the data will not necessarily be stored when it returns.
func (h *hasherStore) Put(ctx context.Context, chunkData ChunkData) (Reference, error) {
	c := chunkData
	var encryptionKey encryption.Key
	if h.chunkEncryption != nil {
		var err error
		c, encryptionKey, err = h.encryptChunkData(chunkData)
		if err != nil {
			return nil, err
		}
	}
	chunk := h.createChunk(c)
	h.storeChunk(ctx, chunk)

	return Reference(append(chunk.Address(), encryptionKey...)), nil
}

// Get returns data of the chunk with the given reference (retrieved from the ChunkStore of hasherStore).
// If the data is encrypted and the reference contains an encryption key, it will be decrypted before
// return.
func (h *hasherStore) Get(ctx context.Context, ref Reference) (ChunkData, error) {
	addr, encryptionKey, err := parseReference(ref, h.hashSize)
	if err != nil {
		return nil, err
	}

	chunk, err := h.store.Get(ctx, addr)
	if err != nil {
		return nil, err
	}

	chunkData := ChunkData(chunk.Data())
	toDecrypt := (encryptionKey != nil)
	if toDecrypt {
		var err error
		chunkData, err = h.decryptChunkData(chunkData, encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return chunkData, nil
}

// Close indicates that no more chunks will be put with the hasherStore, so the Wait
// function can return when all the previously put chunks has been stored.
func (h *hasherStore) Close() {
	close(h.doneC)
}

// Wait returns when
//    1) the Close() function has been called and
//    2) all the chunks which has been Put has been stored
func (h *hasherStore) Wait(ctx context.Context) error {
	defer close(h.quitC)
	var n uint64
	var done bool
	doneC := h.doneC
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-doneC:
			done = true
			doneC = nil
		case err := <-h.errC:
			if err != nil {
				return err
			}
			n++
		}
		if done {
			if n >= h.count {
				return nil
			}
		}
	}
}

func (h *hasherStore) createHash(chunkData ChunkData) Address {
	hasher := h.hashFunc()
	hasher.ResetWithLength(chunkData[:8]) // 8 bytes of length
	hasher.Write(chunkData[8:])           // minus 8 []byte length
	return hasher.Sum(nil)
}

func (h *hasherStore) createChunk(chunkData ChunkData) *chunk {
	hash := h.createHash(chunkData)
	chunk := NewChunk(hash, chunkData)
	return chunk
}

func (h *hasherStore) encryptChunkData(chunkData ChunkData) (ChunkData, encryption.Key, error) {
	if len(chunkData) < 8 {
		return nil, nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	encryptionKey, err := encryption.GenerateRandomKey()
	if err != nil {
		return nil, nil, err
	}

	encryptedSpan, err := h.chunkEncryption.spanEncryption.Encrypt(chunkData[:8], encryptionKey)
	if err != nil {
		return nil, nil, err
	}
	encryptedData, err := h.chunkEncryption.dataEncryption.Encrypt(chunkData[8:], encryptionKey)
	if err != nil {
		return nil, nil, err
	}
	c := make(ChunkData, len(encryptedSpan)+len(encryptedData))
	copy(c[:8], encryptedSpan)
	copy(c[8:], encryptedData)
	return c, encryptionKey, nil
}

func (h *hasherStore) decryptChunkData(chunkData ChunkData, encryptionKey encryption.Key) (ChunkData, error) {
	if len(chunkData) < 8 {
		return nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	decryptedSpan, err := h.chunkEncryption.spanEncryption.Decrypt(chunkData[:8], encryptionKey)
	if err != nil {
		return nil, err
	}

	decryptedData, err := h.chunkEncryption.dataEncryption.Decrypt(chunkData[8:], encryptionKey)
	if err != nil {
		return nil, err
	}

	// removing extra bytes which were just added for padding
	length := int64(ChunkData(decryptedSpan).Size())
	for length > DefaultChunkSize {
		length = length + (DefaultChunkSize - 1)
		length = length / DefaultChunkSize
		length *= h.refSize
	}

	c := make(ChunkData, length+8)
	copy(c[:8], decryptedSpan)
	copy(c[8:], decryptedData[:length])

	return c[:length+8], nil
}

func (h *hasherStore) RefSize() int64 {
	return h.refSize
}

func (h *hasherStore) storeChunk(ctx context.Context, chunk *chunk) {
	atomic.AddUint64(&h.count, 1)
	go func() {
		select {
		case h.errC <- h.store.Put(ctx, chunk):
		case <-h.quitC:
		}
	}()
}

func parseReference(ref Reference, hashSize int) (Address, encryption.Key, error) {
	encryptedRefLength := hashSize + encryption.KeyLength
	switch len(ref) {
	case AddressLength:
		return Address(ref), nil, nil
	case encryptedRefLength:
		encKeyIdx := len(ref) - encryption.KeyLength
		return Address(ref[:encKeyIdx]), encryption.Key(ref[encKeyIdx:]), nil
	default:
		return nil, nil, fmt.Errorf("Invalid reference length, expected %v or %v got %v", hashSize, encryptedRefLength, len(ref))
	}

}
