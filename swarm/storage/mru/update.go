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

// Signature is an alias for a static byte array with the size of a signature
const signatureLength = 65

type Signature [signatureLength]byte

// SignedResourceUpdate contains signature information about a resource update
type SignedResourceUpdate struct {
	resourceUpdate // actual content that will be put on the chunk, less signature
	signature      *Signature
	updateAddr     storage.Address // resulting chunk address for the update
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
// Minimum size is Header + 1 (minimum data length, enforced) + Signature = 142 bytes
const minimumUpdateDataLength = 142 // 2 + 2 + 4 + 4 + storage.KeyLength + 32 + storage.KeyLength + timestampLength + 1 + signatureLength
const maxUpdateDataLength = chunkSize - minimumUpdateDataLength

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
		log.Warn("Call to newUpdateChunk with nil rootAddr or metaHash")
		return nil, NewError(ErrInvalidValue, "newUpdateChunk called without rootAddr or metaHash set")
	}
	// a datalength field set to 0 means the content is a multihash
	var datalength int
	if !mru.multihash {
		datalength = len(mru.data)
		if datalength == 0 {
			return nil, NewError(ErrInvalidValue, "cannot update a resource with no data")
		}
	}

	// prepend version, period, metaHash and rootAddr references
	headerlength := 4 + 4 + storage.KeyLength + storage.KeyLength

	actualdatalength := len(mru.data)
	chunk := storage.NewChunk(mru.updateAddr, nil)
	chunk.SData = make([]byte, 2+2+signatureLength+headerlength+actualdatalength) // initial 4 are uint16 length descriptors for headerlength and datalength

	// data header length does NOT include the header length prefix bytes themselves
	cursor := 0
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(headerlength))
	cursor += 2

	// data length
	binary.LittleEndian.PutUint16(chunk.SData[cursor:], uint16(datalength))
	cursor += 2

	// header = period + version + rootAddr + metaHash
	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.period)
	cursor += 4

	binary.LittleEndian.PutUint32(chunk.SData[cursor:], mru.version)
	cursor += 4

	copy(chunk.SData[cursor:], mru.rootAddr[:storage.KeyLength])
	cursor += storage.KeyLength
	copy(chunk.SData[cursor:], mru.metaHash[:storage.KeyLength])
	cursor += storage.KeyLength

	// add the data
	copy(chunk.SData[cursor:], mru.data)

	// signature is the last item in the chunk data

	cursor += actualdatalength
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
	headerlength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2
	datalength := binary.LittleEndian.Uint16(chunkdata[cursor : cursor+2])
	cursor += 2

	if datalength != 0 && int(2+2+headerlength+datalength+signatureLength) != len(chunkdata) {
		return NewError(ErrNothingToReturn, "length specified in header is different than actual chunk size")
	}

	var exclsignlength int
	// we need extra magic if it's a multihash, since we used datalength 0 in header as an indicator of multihash content
	// retrieve the second varint and set this as the data length
	// TODO: merge with isMultihash code
	if datalength == 0 {
		uvarintbuf := bytes.NewBuffer(chunkdata[headerlength+4:])
		i, err := binary.ReadUvarint(uvarintbuf)
		if err != nil {
			errstr := fmt.Sprintf("corrupt multihash, hash id varint could not be read: %v", err)
			log.Warn(errstr)
			return NewError(ErrCorruptData, errstr)

		}
		i, err = binary.ReadUvarint(uvarintbuf)
		if err != nil {
			errstr := fmt.Sprintf("corrupt multihash, hash length field could not be read: %v", err)
			log.Warn(errstr)
			return NewError(ErrCorruptData, errstr)

		}
		exclsignlength = int(headerlength + uint16(i))
	} else {
		exclsignlength = int(headerlength + datalength + 4)
	}

	// the total length excluding signature is headerlength and datalength fields plus the length of the header and the data given in these fields
	exclsignlength = int(headerlength + datalength + 4)
	if exclsignlength > len(chunkdata) || exclsignlength < 14 {
		return NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d longer than actual chunk data length %d", headerlength, exclsignlength, len(chunkdata)))
	} else if exclsignlength < 14 {
		return NewError(ErrNothingToReturn, fmt.Sprintf("Reported headerlength %d + datalength %d is smaller than minimum valid resource chunk length %d", headerlength, datalength, 14))
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

	//Verify that the updateAddr key that identifies this chunk matches its contents:
	if !bytes.Equal(updateAddr, header.GetUpdateAddr()) {
		return NewError(ErrInvalidSignature, "period,version,rootAddr contained in update chunk do not match updateAddr")
	}

	// if multihash content is indicated we check the validity of the multihash
	// \TODO the check above for multihash probably is sufficient also for this case (or can be with a small adjustment) and if so this code should be removed
	var intdatalength int
	var ismultihash bool
	if datalength == 0 {
		var intheaderlength int
		var err error
		intdatalength, intheaderlength, err = multihash.GetMultihashLength(chunkdata[cursor:])
		if err != nil {
			log.Error("multihash parse error", "err", err)
			return err
		}
		intdatalength += intheaderlength
		multihashboundary := cursor + intdatalength
		if len(chunkdata) != multihashboundary && len(chunkdata) < multihashboundary+signatureLength {
			log.Debug("multihash error", "chunkdatalen", len(chunkdata), "multihashboundary", multihashboundary)
			return errors.New("Corrupt multihash data")
		}
		ismultihash = true
	} else {
		intdatalength = int(datalength)
	}
	data = make([]byte, intdatalength)
	copy(data, chunkdata[cursor:cursor+intdatalength])

	// omit signatures if we have no validator
	var signature *Signature
	cursor += intdatalength
	sigdata := chunkdata[cursor : cursor+signatureLength]
	if len(sigdata) > 0 {
		signature = &Signature{}
		copy(signature[:], sigdata)
	}

	header.multihash = ismultihash

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
