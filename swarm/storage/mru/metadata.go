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

package mru

import (
	"encoding/binary"
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// resourceMetadata encapsulates the immutable information about a mutable resource :)
// once serialized into a chunk, the resource can be retrieved by knowing its content-addressed rootAddr
type resourceMetadata struct {
	startTime uint64         // time at which the resource starts to be valid
	frequency uint64         // expected update frequency for the resource
	name      string         // name of the resource, for the reference of the user
	ownerAddr common.Address // public address of the resource owner
}

// unmarshalBinary populates the resource metadata from a byte array
func (r *resourceMetadata) unmarshalBinary(chunkData []byte) error {
	metadataChunkLength := binary.LittleEndian.Uint16(chunkData[2:6])
	data := chunkData[8:]

	r.startTime = binary.LittleEndian.Uint64(data[:8])
	r.frequency = binary.LittleEndian.Uint64(data[8:16])
	r.name = string(data[16 : 16+metadataChunkLength-metadataChunkOffsetSize])
	copy(r.ownerAddr[:], data[16+metadataChunkLength-metadataChunkOffsetSize:])

	return nil
}

// marshalBinary encodes the metadata into a byte array
func (r *resourceMetadata) marshalBinary() []byte {
	metadataChunkLength := metadataChunkOffsetSize + len(r.name)
	chunkData := make([]byte, metadataChunkLength+8)
	binary.LittleEndian.PutUint16(chunkData[2:6], uint16(metadataChunkLength))

	data := chunkData[8:]

	// root block has first two bytes both set to 0, which distinguishes from update bytes
	binary.LittleEndian.PutUint64(data[:8], r.startTime)
	binary.LittleEndian.PutUint64(data[8:16], r.frequency)
	copy(data[16:16+len(r.name)], []byte(r.name))
	copy(data[16+len(r.name):], r.ownerAddr[:])

	return chunkData
}

// hash returns the root chunk addr and metadata hash that help identify and ascertain ownership of this resource
func (r *resourceMetadata) hash() (rootAddr, metaHash []byte, chunkData []byte) {

	chunkData = r.marshalBinary()
	rootAddr, metaHash = metadataHash(chunkData)
	return rootAddr, metaHash, chunkData

}

// creates a metadata chunk out of a resourceMetadata structure
//TODO: refactor to a method of the resourceMetadata structure
func (metadata *resourceMetadata) newChunk() (chunk *storage.Chunk, metaHash []byte) {
	// the metadata chunk contains a timestamp of when the resource starts to be valid
	// and also how frequently it is expected to be updated
	// from this we know at what time we should look for updates, and how often
	// it also contains the name of the resource, so we know what resource we are working with

	// the key (rootAddr) of the metadata chunk is content-addressed
	// if it wasn't we couldn't replace it later
	// resolving this relationship is left up to external agents (for example ENS)
	rootAddr, metaHash, chunkData := metadata.hash()

	// make the chunk and send it to swarm
	chunk = storage.NewChunk(rootAddr, nil)
	chunk.SData = chunkData
	chunk.Size = int64(len(chunkData))

	return chunk, metaHash
}

// metadataHash returns te root address and metadata hash that help identify and ascertain ownership of this resource
func metadataHash(chunkData []byte) (rootAddr, metaHash []byte) {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(chunkData[:len(chunkData)-common.AddressLength])
	metaHash = hasher.Sum(nil)
	hasher.Reset()
	hasher.Write(metaHash)
	hasher.Write(chunkData[len(chunkData)-common.AddressLength:])
	rootAddr = hasher.Sum(nil)
	return
}
