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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

type updateHeader struct {
	UpdateLookup
	multihash bool
	metaHash  []byte

	time               Timestamp
	previousUpdateAddr storage.Address
	nextUpdateTime     uint64
}

const updateLookupLength = 4 + 4 + storage.KeyLength
const updateHeaderLength = updateLookupLength + 1 + storage.KeyLength

// resourceUpdate encapsulates the information sent as part of a resource update
type resourceUpdate struct {
	updateHeader
	data []byte
}

// Update chunk layout
// Header:
// 2 bytes header Length
// 2 bytes data length
// 4 bytes period
// 4 bytes version
// 32 bytes rootAddr reference
// 32 bytes metaHash digest
// 32 bytes previousUpdateAddr reference
// 40 bytes Timestamp
// Data:
// data (datalength bytes)
// Signature:
// signature: 65 bytes (signatureLength constant)
//
// Minimum size is Header + 1 (minimum data length, enforced) + Signature
const minimumUpdateDataLength = updateHeaderLength + 1 + signatureLength
const maxUpdateDataLength = chunkSize - minimumUpdateDataLength

// Signature is an alias for a static byte array with the size of a signature
const signatureLength = 65

type Signature [signatureLength]byte

// SignedResourceUpdate contains signature information about a resource update
type SignedResourceUpdate struct {
	resourceUpdate // actual content that will be put on the chunk, less signature
	signature      *Signature
	updateAddr     storage.Address // resulting chunk address for the update
}

// Verify checks that signatures are valid and that the signer owns the resource to be updated
func (r *SignedResourceUpdate) Verify() (err error) {
	if len(r.data) == 0 {
		return NewError(ErrInvalidValue, "I refuse to waste swarm space for updates with empty values, amigo (data length is 0)")
	}
	if r.signature == nil {
		return NewError(ErrInvalidSignature, "Missing signature field")
	}

	digest := resourceUpdateChunkDigest(r.updateAddr, r.metaHash, r.data)

	// get the address of the signer (which also checks that it's a valid signature)
	ownerAddr, err := getAddressFromDataSig(digest, *r.signature)
	if err != nil {
		return err
	}

	if !bytes.Equal(r.updateAddr, r.GetUpdateAddr()) {
		return NewError(ErrInvalidSignature, "Signature address does not match with ownerAddr")
	}

	// Check if who signed the resource update really owns the resource
	if !verifyResourceOwnership(ownerAddr, r.metaHash, r.rootAddr) {
		return NewError(ErrUnauthorized, fmt.Sprintf("signature is valid but signer does not own the resource: %v", err))
	}

	return nil
}

// Sign executes the signature to validate the resource
func (r *SignedResourceUpdate) Sign(signer Signer) error {

	updateAddr := r.GetUpdateAddr()

	digest := resourceUpdateChunkDigest(updateAddr, r.metaHash, r.data)
	signature, err := signer.Sign(digest)
	if err != nil {
		return err
	}
	ownerAddress, err := getAddressFromDataSig(digest, signature)
	if err != nil {
		return NewError(ErrInvalidSignature, "Error verifying signature")
	}
	if ownerAddress != signer.Address() {
		return NewError(ErrInvalidSignature, "Signer address does not match private key")
	}
	r.signature = &signature
	r.updateAddr = updateAddr
	return nil
}

// create an update chunk.
func (mru *SignedResourceUpdate) newUpdateChunk() (*storage.Chunk, error) {

	if len(mru.rootAddr) != storage.KeyLength || len(mru.metaHash) != storage.KeyLength {
		log.Warn("Call to newUpdateChunk with incorrect rootAddr or metaHash")
		return nil, NewError(ErrInvalidValue, "newUpdateChunk called without rootAddr or metaHash set")
	}

	datalength := len(mru.data)
	if datalength == 0 {
		return nil, NewError(ErrInvalidValue, "cannot update a resource with no data")
	}

	chunk := storage.NewChunk(mru.updateAddr, nil)
	chunk.SData = make([]byte, 2+2+signatureLength+updateHeaderLength+datalength) // initial 4 are uint16 length descriptors for updateHeaderLength and datalength

	// data header length does NOT include the header length prefix bytes themselves
	cursor := 0
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(updateHeaderLength))
	cursor += 2

	// data length
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(datalength))
	cursor += 2

	// header = period + version + multihash flag + rootAddr + metaHash
	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.period)
	cursor += 4

	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.version)
	cursor += 4

	copy(chunk.SData[cursor:], mru.rootAddr[:storage.KeyLength])
	cursor += storage.KeyLength
	copy(chunk.SData[cursor:], mru.metaHash[:storage.KeyLength])
	cursor += storage.KeyLength

	if mru.multihash {
		if isMultihash(mru.data) == 0 {
			return nil, NewError(ErrInvalidValue, "Invalid multihash")
		}
		chunk.SData[cursor] = 0x01
	}
	cursor++

	// add the data
	copy(chunk.SData[cursor:], mru.data)

	// signature is the last item in the chunk data

	if mru.signature == nil {
		return nil, NewError(ErrInvalidSignature, "newUpdateChunk called without a valid signature")
	}

	cursor += datalength
	copy(chunk.SData[cursor:], mru.signature[:])

	chunk.Size = int64(len(chunk.SData))
	return chunk, nil
}

// retrieve update metadata from chunk data
func (r *SignedResourceUpdate) parseUpdateChunk(updateAddr storage.Address, chunkdata []byte) error {
	// for update chunk layout see SignedResourceUpdate definition

	if len(chunkdata) < minimumUpdateDataLength {
		return NewError(ErrNothingToReturn, fmt.Sprintf("chunk less than %d bytes cannot be a resource update chunk", minimumUpdateDataLength))
	}
	cursor := 0
	declaredHeaderlength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	if declaredHeaderlength != updateHeaderLength {
		return NewErrorf(ErrCorruptData, "Invalid header length. Expected %d, got %d", updateHeaderLength, declaredHeaderlength)
	}

	cursor += 2
	datalength := int(binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2]))
	cursor += 2

	if int(2+2+updateHeaderLength+datalength+signatureLength) != len(chunkdata) {
		return NewError(ErrNothingToReturn, "length specified in header is different than actual chunk size")
	}

	// at this point we can be satisfied that the data integrity is ok
	var header updateHeader

	var data []byte
	header.period = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4
	header.version = binary.LittleEndian.Uint32(chunkdata[cursor : cursor+4])
	cursor += 4

	header.rootAddr = storage.Address(make([]byte, storage.KeyLength))
	header.metaHash = make([]byte, storage.KeyLength)
	copy(header.rootAddr, chunkdata[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength
	copy(header.metaHash, chunkdata[cursor:cursor+storage.KeyLength])
	cursor += storage.KeyLength

	if chunkdata[cursor] == 0x01 {
		header.multihash = true
	}
	cursor++

	// done parsing header. Now verify that the updateAddr key that identifies this chunk matches its contents:
	if !bytes.Equal(updateAddr, header.GetUpdateAddr()) {
		return NewError(ErrInvalidSignature, "period,version,rootAddr contained in update chunk do not match updateAddr")
	}

	// if multihash content is indicated we check the validity of the multihash
	if header.multihash {
		mhLength, mhHeaderLength, err := multihash.GetMultihashLength(chunkdata[cursor:])
		if err != nil {
			log.Error("multihash parse error", "err", err)
			return err
		}
		if datalength != mhLength+mhHeaderLength {
			log.Debug("multihash error", "datalength", datalength, "mhLength", mhLength, "mhHeaderLength", mhHeaderLength)
			return errors.New("Corrupt multihash data")
		}
	}
	data = make([]byte, datalength)
	copy(data, chunkdata[cursor:cursor+datalength])

	var signature *Signature
	cursor += datalength
	sigdata := chunkdata[cursor : cursor+signatureLength]
	if len(sigdata) > 0 {
		signature = &Signature{}
		copy(signature[:], sigdata)
	}

	r.signature = signature
	r.updateAddr = updateAddr
	r.resourceUpdate = resourceUpdate{
		updateHeader: header,
		data:         data,
	}

	if err := r.Verify(); err != nil {
		return NewError(ErrUnauthorized, fmt.Sprintf("Invalid signature: %v", err))
	}

	return nil

}

// resourceUpdateChunkDigest creates the resource update digest used in signatures (formerly known as keyDataHash)
func resourceUpdateChunkDigest(updateAddr storage.Address, metaHash []byte, data []byte) common.Hash {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(updateAddr)
	hasher.Write(metaHash)
	hasher.Write(data)
	return common.BytesToHash(hasher.Sum(nil))
}

// getAddressFromDataSig extracts the address of the resource update signer
func getAddressFromDataSig(digest common.Hash, signature Signature) (common.Address, error) {
	pub, err := crypto.SigToPub(digest.Bytes(), signature[:])
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}

func verifyResourceOwnership(ownerAddr common.Address, metaHash []byte, rootAddr storage.Address) bool {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(metaHash)
	hasher.Write(ownerAddr.Bytes())
	rootAddr2 := hasher.Sum(nil)
	return bytes.Equal(rootAddr2, rootAddr)
}
