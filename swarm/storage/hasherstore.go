// Copyright 2016 The go-ethereum Authors
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
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/storage/encryption"
)

type chunkEncryption struct {
	spanEncryption encryption.Encryption
	dataEncryption encryption.Encryption
}

type hasherStore struct {
	store           ChunkStore
	hasher          SwarmHash
	chunkEncryption *chunkEncryption
}

func newChunkEncryption() *chunkEncryption {
	return &chunkEncryption{
		spanEncryption: encryption.New(0, math.MaxUint32, sha3.NewKeccak256),
		dataEncryption: encryption.New(4096, 0, sha3.NewKeccak256),
	}
}

func NewHasherStore(chunkStore ChunkStore, hashFunc SwarmHasher, toEncrypt bool) *hasherStore {
	var chunkEncryption *chunkEncryption
	if toEncrypt {
		chunkEncryption = newChunkEncryption()
	}
	return &hasherStore{
		store:           chunkStore,
		hasher:          hashFunc(),
		chunkEncryption: chunkEncryption,
	}
}

func (h *hasherStore) Put(chunkData ChunkData) (Reference, error) {
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

	// need to do this parallelly
	h.store.Put(chunk)

	return Reference(append(chunk.Key, encryptionKey...)), nil

	//TODO: implement wait for storage
	// if chunkC != nil {
	// 	chunkC <- newChunk
	// 	storageWG.Add(1)
	// 	go func() {
	// 		defer storageWG.Done()
	// 		<-newChunk.dbStored
	// 	}()
	//
}

func (h *hasherStore) Get(ref Reference) (ChunkData, error) {
	key, encryptionKey, err := parseReference(ref)
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

func (h *hasherStore) Close() {
	return
}

func (h *hasherStore) createHash(chunkData ChunkData) Key {
	h.hasher.ResetWithLength(chunkData[:8]) // 8 bytes of length
	h.hasher.Write(chunkData[8:])           // minus 8 []byte length
	return h.hasher.Sum(nil)
}

func (h *hasherStore) createChunk(chunkData ChunkData, chunkSize int64) *Chunk {
	hash := h.createHash(chunkData)
	chunk := NewChunk(hash, nil)
	chunk.SData = chunkData
	chunk.Size = chunkSize

	return chunk
}

func (p *hasherStore) encryptChunkData(chunkData ChunkData) (ChunkData, encryption.Key, error) {
	if len(chunkData) < 8 {
		return nil, nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	encryptionKey, err := encryption.GenerateRandomKey()
	if err != nil {
		return nil, nil, err
	}

	c := make(ChunkData, len(chunkData))
	encryptedSpan, err := p.chunkEncryption.spanEncryption.Encrypt(chunkData[:8], encryptionKey)
	if err != nil {
		return nil, nil, err
	}
	encryptedData, err := p.chunkEncryption.dataEncryption.Encrypt(chunkData[8:], encryptionKey)
	if err != nil {
		return nil, nil, err
	}
	copy(c[:8], encryptedSpan)
	copy(c[8:], encryptedData)
	return c, encryptionKey, nil
}

func (p *hasherStore) decryptChunkData(chunkData ChunkData, encryptionKey encryption.Key) (ChunkData, error) {
	if len(chunkData) < 8 {
		return nil, fmt.Errorf("Invalid ChunkData, min length 8 got %v", len(chunkData))
	}

	c := make(ChunkData, len(chunkData))
	decryptedSpan, err := p.chunkEncryption.spanEncryption.Decrypt(chunkData[:8], encryptionKey)
	if err != nil {
		return nil, err
	}
	decryptedData, err := p.chunkEncryption.dataEncryption.Decrypt(chunkData[8:], encryptionKey)
	if err != nil {
		return nil, err
	}
	copy(c[:8], decryptedSpan)
	copy(c[8:], decryptedData)
	return c, nil
}

func parseReference(ref Reference) (Key, encryption.Key, error) {
	encryptedKeyLength := KeyLength + encryption.KeyLength
	switch len(ref) {
	case KeyLength:
		return Key(ref), nil, nil
	case encryptedKeyLength:
		encKeyIdx := len(ref) - encryption.KeyLength
		return Key(ref[:encKeyIdx]), encryption.Key(ref[encKeyIdx:]), nil
	default:
		return nil, nil, fmt.Errorf("Invalid reference length, expected %v or %v got %v", KeyLength, encryptedKeyLength, len(ref))
	}

}
