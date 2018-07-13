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
	"context"
	"io"
)


type chunkEncryption struct {
	spanEncryption encryption.Encryption
	dataEncryption encryption.Encryption
}

type FileHasherStore struct {
	store           ChunkStore
	hashFunc        SwarmHasher
	chunkEncryption *chunkEncryption
	hashSize        int   // content hash size
	refSize         int64 // reference size (content hash + possibly encryption key)
	wg              *sync.WaitGroup
	closed          chan struct{}
}

func newChunkEncryption(chunkSize, refSize int64) *chunkEncryption {
	return &chunkEncryption{
		spanEncryption: encryption.New(0, uint32(chunkSize/refSize), sha3.NewKeccak256),
		dataEncryption: encryption.New(int(chunkSize), 0, sha3.NewKeccak256),
	}
}

// NewFileHasherStore creates a FileHasherStore object, which implements Putter and Getter interfaces.
// With the FileHasherStore you can put and get chunk data (which is just []byte) into a ChunkStore
// and the FileHasherStore will take core of encryption/decryption of data if necessary
func NewFileHasherStore(chunkStore ChunkStore, hashFunc SectionHasherFunc, toEncrypt bool, erasure bool) *FileHasherStore {
	var chunkEncryption *chunkEncryption
	f := func(children []byte) SectionHasher {
		return hashFunc()
	}
	if erasure {
		f = func(children []byte) SectionHasher {

		}
	}
	hashSize := hashFunc().Size()
	refSize := int64(hashSize)
	if toEncrypt {
		refSize += encryption.KeyLength
		chunkEncryption = newChunkEncryption(DefaultChunkSize, refSize)
	}

	return &FileHasherStore{
		store:           chunkStore,
		hashFunc:        hashFunc,
		hashSize:        hashSize,
		refSize:         refSize,
		wg:              &sync.WaitGroup{},
		closed:          make(chan struct{}),
	}
}


// extensions of the base chunk hasher (SectionHasher interface)

// wrapper that completes a batch of child chunks using CRS erasure coding
//
type redundanteChunkHasher struct {
	SectionHasher
	// erasure
}

//
type encryptedChunkHasher struct {
	chunkEncryption *chunkEncryption
	SectionHasher
}

type storeChunkHasher struct {
	SectionHasher
	put func(Address, ChunkData) error
}

// New is the function called by the splitter/filehasher when creating a node
func (fhs *FileHasherStore) NewChunkHasher() SwarmHash {
	return &encryptedChunkHasherStorer{
		hasher: fhs.hashFunc()
		chunkEncryption: chunkEncryption,
		getChunkData: func(int) ChunkData,
	}
}

func (e *encryptedChunkHasherStorer) Write(i int, b []byte) {
	// call encrypt
	e.hasher.Write(i, b)
}

func  (e *encryptedChunkHasherStorer) Sum(b []byte, length int, meta []byte) {
	// length == e.DataSize()

	length = e.complete(length, e.hasher.DataSize(), e.getChunkData, e.hasher.Write)
	// pan and encrypt
	return e.hasher.Sum(b, length, meta)
}


// Put stores the chunkData into the ChunkStore of the FileHasherStore and returns the reference.
// If FileHasherStore has a chunkEncryption object, the data will be encrypted.
// Asynchronous function, the data will not necessarily be stored when it returns.
func (h *FileHasherStore) Put(chunkData ChunkData) (Reference, error) {
	c := chunkData
	size := chunkData.Size()
	var encryptionKey encryption.Key
	if h.chunkEncryption != nil {
		var err error
		c, encryptionKey, err = h.encryptChunkData(chunkData)
		if err != nil {
			return nil, err
		}
	}
	chunk := h.createChunk(c, size)

	h.storeChunk(chunk)

	return Reference(append(chunk.Addr, encryptionKey...)), nil
}

// Get returns data of the chunk with the given reference (retrieved from the ChunkStore of FileHasherStore).
// If the data is encrypted and the reference contains an encryption key, it will be decrypted before
// return.
func (h *FileHasherStore) Get(ref Reference) (ChunkData, error) {
	key, encryptionKey, err := parseReference(ref, h.hashSize)
	if err != nil {
		return nil, err
	}
	toDecrypt := (encryptionKey != nil)

	chunk, err := h.store.Get(key)
	if err != nil {
		return nil, err
	}

	chunkData := chunk.SData
	if toDecrypt {
		var err error
		chunkData, err = h.decryptChunkData(chunkData, encryptionKey)
		if err != nil {
			return nil, err
		}
	}
	return chunkData, nil
}

// Close indicates that no more chunks will be put with the FileHasherStore, so the Wait
// function can return when all the previously put chunks has been stored.
func (h *FileHasherStore) Close() {
	close(h.closed)
}

// Wait returns when
//    1) the Close() function has been called and
//    2) all the chunks which has been Put has been stored
func (h *FileHasherStore) Wait(ctx context.Context) error {
	<-h.closed
	h.wg.Wait()
	return nil
}

func (h *FileHasherStore) createHash(chunkData ChunkData) Address {
	hasher := h.hashFunc()
	hasher.ResetWithLength(chunkData[:8]) // 8 bytes of length
	hasher.Write(chunkData[8:])           // minus 8 []byte length
	return hasher.Sum(nil)
}

func (h *FileHasherStore) createChunk(chunkData ChunkData, chunkSize int64) *Chunk {
	hash := h.createHash(chunkData)
	chunk := NewChunk(hash, nil)
	chunk.SData = chunkData
	chunk.Size = chunkSize

	return chunk
}

func (h *FileHasherStore) encryptChunkData(chunkData ChunkData) (ChunkData, encryption.Key, error) {
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

func (h *FileHasherStore) decryptChunkData(chunkData ChunkData, encryptionKey encryption.Key) (ChunkData, error) {
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
	length := ChunkData(decryptedSpan).Size()
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

func (h *FileHasherStore) RefSize() int64 {
	return h.refSize
}

func (h *FileHasherStore) storeChunk(chunk *Chunk) {
	h.wg.Add(1)
	go func() {
		<-chunk.dbStoredC
		h.wg.Done()
	}()
	h.store.Put(chunk)
}

func parseReference(ref Reference, hashSize int) (Address, encryption.Key, error) {
	encryptedKeyLength := hashSize + encryption.KeyLength
	switch len(ref) {
	case KeyLength:
		return Address(ref), nil, nil
	case encryptedKeyLength:
		encKeyIdx := len(ref) - encryption.KeyLength
		return Address(ref[:encKeyIdx]), encryption.Key(ref[encKeyIdx:]), nil
	default:
		return nil, nil, fmt.Errorf("Invalid reference length, expected %v or %v got %v", hashSize, encryptedKeyLength, len(ref))
	}

}
