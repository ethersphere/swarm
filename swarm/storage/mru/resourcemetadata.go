package mru

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type resourceMetadata struct {
	startTime uint64
	frequency uint64
	name      string
	ownerAddr common.Address
}

func (r *resourceMetadata) UnmarshalBinary(chunkData []byte) error {
	metadataChunkLength := binary.LittleEndian.Uint16(chunkData[2:6])
	data := chunkData[8:]

	r.startTime = binary.LittleEndian.Uint64(data[:8])
	r.frequency = binary.LittleEndian.Uint64(data[8:16])
	r.name = string(data[16 : 16+metadataChunkLength-metadataChunkOffsetSize])
	copy(r.ownerAddr[:], data[16+metadataChunkLength-metadataChunkOffsetSize:])

	return nil
}

func (r *resourceMetadata) MarshalBinary() []byte {
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

func (r *resourceMetadata) hash() (rootAddr, metaHash []byte, chunkData []byte) {

	chunkData = r.MarshalBinary()
	metaHash, rootAddr = metadataHash(chunkData)
	return metaHash, rootAddr, chunkData

}

func metadataHash(chunkData []byte) (rootAddr, metaHash []byte) {
	hasher := hashPool.Get().(storage.SwarmHash)
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
