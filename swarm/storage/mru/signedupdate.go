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
	"hash"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// SignedResourceUpdate represents a resource update with all the necessary information to prove ownership of the resource
type SignedResourceUpdate struct {
	resourceUpdate // actual content that will be put on the chunk, less signature
	signature      *Signature
	updateAddr     storage.Address // resulting chunk address for the update (not serialized, for internal use)
	binaryData     []byte          // resulting serialized data (not serialized, for efficiency/internal use)
}

// SignedResourceUpdate layout
// resourceUpdate bytes
// SignatureLength bytes
const minimumSignedUpdateLength = minimumUpdateDataLength + signatureLength

// Verify checks that signatures are valid and that the signer owns the resource to be updated
func (r *SignedResourceUpdate) Verify() (err error) {
	if len(r.data) == 0 {
		return NewError(ErrInvalidValue, "Update does not contain data")
	}
	if r.signature == nil {
		return NewError(ErrInvalidSignature, "Missing signature field")
	}

	digest, err := r.GetDigest()
	if err != nil {
		return err
	}

	// get the address of the signer (which also checks that it's a valid signature)
	r.view.User, err = getUserAddr(digest, *r.signature)
	if err != nil {
		return err
	}

	// check that the lookup information contained in the chunk matches the updateAddr (chunk search key)
	// that was used to retrieve this chunk
	// if this validation fails, someone forged a chunk.
	if !bytes.Equal(r.updateAddr, r.UpdateAddr()) {
		return NewError(ErrInvalidSignature, "Signature address does not match with update user address")
	}

	return nil
}

// Sign executes the signature to validate the resource
func (r *SignedResourceUpdate) Sign(signer Signer) error {
	r.view.User = signer.Address()
	r.binaryData = nil           //invalidate serialized data
	digest, err := r.GetDigest() // computes digest and serializes into .binaryData
	if err != nil {
		return err
	}

	signature, err := signer.Sign(digest)
	if err != nil {
		return err
	}

	// Although the Signer interface returns the public address of the signer,
	// recover it from the signature to see if they match
	userAddr, err := getUserAddr(digest, signature)
	if err != nil {
		return NewError(ErrInvalidSignature, "Error verifying signature")
	}

	if userAddr != signer.Address() { // sanity check to make sure the Signer is declaring the same address used to sign!
		return NewError(ErrInvalidSignature, "Signer address does not match update user address")
	}

	r.signature = &signature
	r.updateAddr = r.UpdateAddr()
	return nil
}

// create an update chunk.
func (r *SignedResourceUpdate) toChunk() (*storage.Chunk, error) {

	// Check that the update is signed and serialized
	// For efficiency, data is serialized during signature and cached in
	// the binaryData field when computing the signature digest in .getDigest()
	if r.signature == nil || r.binaryData == nil {
		return nil, NewError(ErrInvalidSignature, "toChunk called without a valid signature or payload data. Call .Sign() first.")
	}

	chunk := storage.NewChunk(r.updateAddr, nil)
	resourceUpdateLength := r.resourceUpdate.binaryLength()
	chunk.SData = r.binaryData

	// signature is the last item in the chunk data
	copy(chunk.SData[resourceUpdateLength:], r.signature[:])

	chunk.Size = int64(len(chunk.SData))
	return chunk, nil
}

// fromChunk populates this structure from chunk data. It does not verify the signature is valid.
func (r *SignedResourceUpdate) fromChunk(updateAddr storage.Address, chunkdata []byte) error {
	// for update chunk layout see SignedResourceUpdate definition

	//deserialize the resource update portion
	if err := r.resourceUpdate.binaryGet(chunkdata[:len(chunkdata)-signatureLength]); err != nil {
		return err
	}

	// Extract the signature
	var signature *Signature
	cursor := r.resourceUpdate.binaryLength()
	sigdata := chunkdata[cursor : cursor+signatureLength]
	if len(sigdata) > 0 {
		signature = &Signature{}
		copy(signature[:], sigdata)
	}

	r.signature = signature
	r.updateAddr = updateAddr
	r.binaryData = chunkdata

	return nil

}

// GetDigest creates the resource update digest used in signatures
// the serialized payload is cached in .binaryData
func (r *SignedResourceUpdate) GetDigest() (result common.Hash, err error) {
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	dataLength := r.resourceUpdate.binaryLength()
	if r.binaryData == nil {
		r.binaryData = make([]byte, dataLength+signatureLength)
		if err := r.resourceUpdate.binaryPut(r.binaryData[:dataLength]); err != nil {
			return result, err
		}
	}
	hasher.Write(r.binaryData[:dataLength]) //everything except the signature.

	return common.BytesToHash(hasher.Sum(nil)), nil
}

func (r *SignedResourceUpdate) FromValues(values Values, data []byte, parseView bool) error {
	signatureBytes, err := hexutil.Decode(values.Get("signature"))
	if err != nil {
		r.signature = nil
	} else {
		if len(signatureBytes) != signatureLength {
			return NewError(ErrInvalidSignature, "Incorrect signature length")
		}
		r.signature = new(Signature)
		copy(r.signature[:], signatureBytes)
	}
	err = r.resourceUpdate.FromValues(values, data, parseView)
	if err != nil {
		return err
	}
	r.updateAddr = r.UpdateAddr()
	return err
}

func (r *SignedResourceUpdate) ToValues(values Values) []byte {
	values.Set("signature", hexutil.Encode(r.signature[:]))
	return r.resourceUpdate.ToValues(values)
}

// getUserAddr extracts the address of the resource update signer
func getUserAddr(digest common.Hash, signature Signature) (common.Address, error) {
	pub, err := crypto.SigToPub(digest.Bytes(), signature[:])
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pub), nil
}
