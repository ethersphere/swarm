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
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	binaryData     []byte          // resulting serialized data
}

// Verify checks that signatures are valid and that the signer owns the resource to be updated
func (r *SignedResourceUpdate) Verify() (err error) {
	if len(r.data) == 0 {
		return NewError(ErrInvalidValue, "I refuse to waste swarm space for updates with empty values, amigo (data length is 0)")
	}
	if r.signature == nil {
		return NewError(ErrInvalidSignature, "Missing signature field")
	}

	digest := r.getDigest(r.updateAddr)

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

	digest := r.getDigest(updateAddr)
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

	if mru.signature == nil {
		return nil, NewError(ErrInvalidSignature, "newUpdateChunk called without a valid signature")
	}

	chunk := storage.NewChunk(mru.updateAddr, nil)
	resourceUpdateLength := mru.resourceUpdate.binaryLength()
	chunk.SData = make([]byte, resourceUpdateLength+signatureLength)

	if err := mru.resourceUpdate.binaryPut(chunk.SData[:resourceUpdateLength]); err != nil {
		return nil, err
	}

	// signature is the last item in the chunk data
	copy(chunk.SData[resourceUpdateLength:], mru.signature[:])

	chunk.Size = int64(len(chunk.SData))
	return chunk, nil
}

// retrieve update metadata from chunk data
func (r *SignedResourceUpdate) parseUpdateChunk(updateAddr storage.Address, chunkdata []byte) error {
	// for update chunk layout see SignedResourceUpdate definition

	if err := r.resourceUpdate.binaryGet(chunkdata); err != nil {
		return err
	}

	var signature *Signature
	cursor := r.resourceUpdate.binaryLength()
	sigdata := chunkdata[cursor : cursor+signatureLength]
	if len(sigdata) > 0 {
		signature = &Signature{}
		copy(signature[:], sigdata)
	}

	if !bytes.Equal(updateAddr, r.updateHeader.GetUpdateAddr()) {
		return NewError(ErrInvalidSignature, "period,version,rootAddr contained in update chunk do not match updateAddr")
	}

	r.signature = signature
	r.updateAddr = updateAddr

	if err := r.Verify(); err != nil {
		return NewError(ErrUnauthorized, fmt.Sprintf("Invalid signature: %v", err))
	}

	return nil

}

// getDigest creates the resource update digest used in signatures (formerly known as keyDataHash)
// the serialized payload is cached in .binaryData
func (r *SignedResourceUpdate) getDigest(updateAddr storage.Address) (result common.Hash) {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	updateData := make([]byte, r.resourceUpdate.binaryLength())
	if err := r.resourceUpdate.binaryPut(updateData); err != nil {
		return result
	}
	hasher.Write(updateAddr) //lookup key
	hasher.Write(updateData) //everything except the signature.

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
