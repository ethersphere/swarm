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

package hasherstore

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/swarm/constants"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/encryption"
	"golang.org/x/crypto/sha3"
)

type HasherStore struct {
	store     storage.ChunkStore
	toEncrypt bool
	hashFunc  storage.SwarmHasher
	hashSize  int           // content hash size
	refSize   int64         // reference size (content hash + possibly encryption key)
	errC      chan error    // global error channel
	doneC     chan struct{} // closed by Close() call to indicate that count is the final number of chunks
	quitC     chan struct{} // closed to quit unterminated routines
	// nrChunks is used with atomic functions
	// it is required to be at the end of the struct to ensure 64bit alignment for arm architecture
	// see: https://golang.org/pkg/sync/atomic/#pkg-note-BUG
	nrChunks uint64 // number of chunks to store
}

// NewHasherStore creates a HasherStore object, which implements Putter and Getter interfaces.
// With the HasherStore you can put and get chunk data (which is just []byte) into a ChunkStore
// and the HasherStore will take core of encryption/decryption of data if necessary
func NewHasherStore(store storage.ChunkStore, hashFunc storage.SwarmHasher, toEncrypt bool) *HasherStore {
	hashSize := hashFunc().Size()
	refSize := int64(hashSize)
	if toEncrypt {
		refSize += encryption.KeyLength
	}

	h := &HasherStore{
		store:     store,
		toEncrypt: toEncrypt,
		hashFunc:  hashFunc,
		hashSize:  hashSize,
		refSize:   refSize,
		errC:      make(chan error),
		doneC:     make(chan struct{}),
		quitC:     make(chan struct{}),
	}

	return h
}

// Put stores the chunkData into the ChunkStore of the HasherStore and returns the reference.
// If HasherStore has a chunkEncryption object, the data will be encrypted.
// Asynchronous function, the data will not necessarily be stored when it returns.
func (h *HasherStore) Put(ctx context.Context, chunkData storage.ChunkData) (storage.Reference, error) {
	c := chunkData
	var encryptionKey encryption.Key
	if h.toEncrypt {
		var err error
		c, encryptionKey, err = h.encryptChunkData(chunkData)
		if err != nil {
			return nil, err
		}
	}
	chunk := h.createChunk(c)
	h.storeChunk(ctx, chunk)

	return storage.Reference(append(chunk.Address(), encryptionKey...)), nil
}

// Get returns data of the chunk with the given reference (retrieved from the ChunkStore of HasherStore).
// If the data is encrypted and the reference contains an encryption key, it will be decrypted before
// return.
func (h *HasherStore) Get(ctx context.Context, ref storage.Reference) (storage.ChunkData, error) {
	addr, encryptionKey, err := parseReference(ref, h.hashSize)
	if err != nil {
		return nil, err
	}

	chunk, err := h.store.Get(ctx, addr)
	if err != nil {
		return nil, err
	}

	chunkData := storage.ChunkData(chunk.Data())
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

// Close indicates that no more chunks will be put with the HasherStore, so the Wait
// function can return when all the previously put chunks has been stored.
func (h *HasherStore) Close() {
	close(h.doneC)
}

// Wait returns when
//    1) the Close() function has been called and
//    2) all the chunks which has been Put has been stored
func (h *HasherStore) Wait(ctx context.Context) error {
	defer close(h.quitC)
	var nrStoredChunks uint64 // number of stored chunks
	var done bool
	doneC := h.doneC
	for {
		select {
		// if context is done earlier, just return with the error
		case <-ctx.Done():
			return ctx.Err()
		// doneC is closed if all chunks have been submitted, from then we just wait until all of them are also stored
		case <-doneC:
			done = true
			doneC = nil
		// a chunk has been stored, if err is nil, then successfully, so increase the stored chunk counter
		case err := <-h.errC:
			if err != nil {
				return err
			}
			nrStoredChunks++
		}
		// if all the chunks have been submitted and all of them are stored, then we can return
		if done {
			if nrStoredChunks >= atomic.LoadUint64(&h.nrChunks) {
				return nil
			}
		}
	}
}

func (h *HasherStore) createHash(chunkData storage.ChunkData) storage.Address {
	hasher := h.hashFunc()
	hasher.ResetWithLength(chunkData[:8]) // 8 bytes of length
	hasher.Write(chunkData[8:])           // minus 8 []byte length
	return hasher.Sum(nil)
}

func (h *HasherStore) createChunk(chunkData storage.ChunkData) *storage.Chunk {
	hash := h.createHash(chunkData)
	chunk := storage.NewChunk(hash, chunkData)
	return &chunk
}

func (h *HasherStore) encryptChunkData(chunkData storage.ChunkData) (storage.ChunkData, encryption.Key, error) {
	if len(chunkData) < 8 {
		return nil, nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	key, encryptedSpan, encryptedData, err := h.encrypt(chunkData)
	if err != nil {
		return nil, nil, err
	}
	c := make(storage.ChunkData, len(encryptedSpan)+len(encryptedData))
	copy(c[:8], encryptedSpan)
	copy(c[8:], encryptedData)
	return c, key, nil
}

func (h *HasherStore) decryptChunkData(chunkData storage.ChunkData, encryptionKey encryption.Key) (storage.ChunkData, error) {
	if len(chunkData) < 8 {
		return nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	decryptedSpan, decryptedData, err := h.decrypt(chunkData, encryptionKey)
	if err != nil {
		return nil, err
	}

	// removing extra bytes which were just added for padding
	length := storage.ChunkData(decryptedSpan).Size()
	for length > constants.DefaultChunkSize {
		length = length + (constants.DefaultChunkSize - 1)
		length = length / constants.DefaultChunkSize
		length *= uint64(h.refSize)
	}

	c := make(storage.ChunkData, length+8)
	copy(c[:8], decryptedSpan)
	copy(c[8:], decryptedData[:length])

	return c, nil
}

func (h *HasherStore) RefSize() int64 {
	return h.refSize
}

func (h *HasherStore) encrypt(chunkData storage.ChunkData) (encryption.Key, []byte, []byte, error) {
	key := encryption.GenerateRandomKey(encryption.KeyLength)
	encryptedSpan, err := h.newSpanEncryption(key).Encrypt(chunkData[:8])
	if err != nil {
		return nil, nil, nil, err
	}
	encryptedData, err := h.newDataEncryption(key).Encrypt(chunkData[8:])
	if err != nil {
		return nil, nil, nil, err
	}
	return key, encryptedSpan, encryptedData, nil
}

func (h *HasherStore) decrypt(chunkData storage.ChunkData, key encryption.Key) ([]byte, []byte, error) {
	encryptedSpan, err := h.newSpanEncryption(key).Encrypt(chunkData[:8])
	if err != nil {
		return nil, nil, err
	}
	encryptedData, err := h.newDataEncryption(key).Encrypt(chunkData[8:])
	if err != nil {
		return nil, nil, err
	}
	return encryptedSpan, encryptedData, nil
}

func (h *HasherStore) newSpanEncryption(key encryption.Key) encryption.Encryption {
	return encryption.New(key, 0, uint32(constants.DefaultChunkSize/h.refSize), sha3.NewLegacyKeccak256)
}

func (h *HasherStore) newDataEncryption(key encryption.Key) encryption.Encryption {
	return encryption.New(key, int(constants.DefaultChunkSize), 0, sha3.NewLegacyKeccak256)
}

func (h *HasherStore) storeChunk(ctx context.Context, chunk *storage.Chunk) {
	atomic.AddUint64(&h.nrChunks, 1)
	go func() {
		select {
		case h.errC <- h.store.Put(ctx, *chunk):
		case <-h.quitC:
		}
	}()
}

func parseReference(ref storage.Reference, hashSize int) (storage.Address, encryption.Key, error) {
	encryptedRefLength := hashSize + encryption.KeyLength
	switch len(ref) {
	case constants.AddressLength:
		return storage.Address(ref), nil, nil
	case encryptedRefLength:
		encKeyIdx := len(ref) - encryption.KeyLength
		return storage.Address(ref[:encKeyIdx]), encryption.Key(ref[encKeyIdx:]), nil
	default:
		return nil, nil, fmt.Errorf("Invalid reference length, expected %v or %v got %v", hashSize, encryptedRefLength, len(ref))
	}
}
