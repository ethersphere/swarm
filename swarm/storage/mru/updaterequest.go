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
	"encoding/json"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// Signature is an alias for a static byte array with the size of a signature
const signatureLength = 65

type Signature [signatureLength]byte

// updateRequestJSON represents a JSON-serialized UpdateRequest
type updateRequestJSON struct {
	Name      string `json:"name"`
	Frequency uint64 `json:"frequency"`
	StartTime uint64 `json:"startTime,omitempty"`
	OwnerAddr string `json:"ownerAddr"`
	RootAddr  string `json:"rootAddr,omitempty"`
	MetaHash  string `json:"metaHash,omitempty"`
	Version   uint32 `json:"version"`
	Period    uint32 `json:"period"`
	Data      string `json:"data,omitempty"`
	Multihash bool   `json:"multiHash"`
	Signature string `json:"signature,omitempty"`
}

// SignedResourceUpdate contains signature information about a resource update
type SignedResourceUpdate struct {
	resourceData // actual content that will be put on the chunk, less signature
	signature    *Signature
	updateAddr   storage.Address // resulting chunk address for the update
}

// Update chunk layout
// Header:
// 2 bytes header Length
// 2 bytes data length
// 4 bytes period
// 4 bytes version
// 32 bytes rootAddr reference
// 32 bytes metaHash digest
// Data:
// data (datalength bytes)
// Signature:
// signature: 65 bytes (signatureLength constant)
//
// Minimum size is Header + 1 (minimum data length, enforced) + Signature = 142 bytes

const minimumUpdateDataLength = 142
const maxUpdateDataLength = chunkSize - minimumUpdateDataLength

// UpdateRequest represents an update and/or resource create message
type UpdateRequest struct {
	SignedResourceUpdate
	resourceMetadata
}

// NewUpdateRequest returns a ready to sign UpdateRequest message to create a new resource
func NewUpdateRequest(name string, frequency, startTime uint64, ownerAddr common.Address, data []byte, multihash bool) (*UpdateRequest, error) {
	if !isSafeName(name) {
		return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s' when creating NewMruRequest", name))
	}

	updateRequest := &UpdateRequest{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceData: resourceData{
				version:   1,
				period:    1,
				data:      data,
				multihash: multihash,
			},
		},
		resourceMetadata: resourceMetadata{
			name:      name,
			frequency: frequency,
			startTime: startTime,
			ownerAddr: ownerAddr,
		},
	}

	updateRequest.rootAddr, updateRequest.metaHash, _ = updateRequest.resourceMetadata.hash()

	return updateRequest, nil
}

// Frequency Returns the resource expected update frequency
func (r *UpdateRequest) Frequency() uint64 {
	return r.frequency
}

// Name returns the resource human-readable name
func (r *UpdateRequest) Name() string {
	return r.name
}

// Multihash returns true if the resource data should be interpreted as a multihash
func (r *UpdateRequest) Multihash() bool {
	return r.multihash
}

// Version returns the resource version to publish
func (r *UpdateRequest) Version() uint32 {
	return r.version
}

// Period returns in which period the resource will be published
func (r *UpdateRequest) Period() uint32 {
	return r.period
}

// StartTime returns the time that the resource was/will be created at
func (r *UpdateRequest) StartTime() uint64 {
	return r.startTime
}

// OwnerAddr returns the resource owner's address
func (r *UpdateRequest) OwnerAddr() common.Address {
	return r.ownerAddr
}

// RootAddr returns the metadata chunk address
func (r *UpdateRequest) RootAddr() storage.Address {
	return r.rootAddr
}

// Sign executes the signature to validate the resource and sets the owner address field
func (r *UpdateRequest) Sign(signer Signer) error {
	if err := r.SignedResourceUpdate.Sign(signer); err != nil {
		return err
	}
	r.ownerAddr = signer.Address()
	return nil
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

	if !bytes.Equal(r.updateAddr, resourceUpdateChunkAddr(r.period, r.version, r.rootAddr)) {
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

	updateAddr := resourceUpdateChunkAddr(r.period, r.version, r.rootAddr)

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

// SetData stores the payload data the resource will be updated with
func (r *UpdateRequest) SetData(data []byte) {
	r.signature = nil
	r.data = data
	r.frequency = 0 //mark as update
}

// decode takes an update request JSON and returns an UpdateRequest
func (j *updateRequestJSON) decode() (*UpdateRequest, error) {

	// make sure name only contains ascii values
	if !isSafeName(j.Name) {
		return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s'", j.Name))
	}

	r := &UpdateRequest{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceData: resourceData{
				version:   j.Version,
				period:    j.Period,
				multihash: j.Multihash,
			},
		},
		resourceMetadata: resourceMetadata{
			name:      j.Name,
			frequency: j.Frequency,
			startTime: j.StartTime,
		},
	}

	ownerAddrBytes, err := hexutil.Decode(j.OwnerAddr)
	if err != nil || len(ownerAddrBytes) != common.AddressLength {
		return nil, NewError(ErrInvalidValue, "Cannot decode ownerAddr")
	}
	copy(r.ownerAddr[:], ownerAddrBytes)

	if j.Data != "" {
		r.data, err = hexutil.Decode(j.Data)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode data")
		}
	}

	var declaredRootAddr storage.Address
	var declaredMetaHash []byte

	if j.RootAddr != "" {
		declaredRootAddr, err = hexutil.Decode(j.RootAddr)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode rootAddr")
		}
	}

	if j.MetaHash != "" {
		declaredMetaHash, err = hexutil.Decode(j.MetaHash)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode metaHash")
		}
	}

	if r.frequency > 0 { // we use frequency > 0 to know it is a new resource creation
		// for new resource creation, rootAddr and metaHash are optional.
		// however, if the user sent them, we check them.
		r.rootAddr, r.metaHash, _ = r.resourceMetadata.hash()
		if j.RootAddr != "" && !bytes.Equal(declaredRootAddr, r.rootAddr) {
			return nil, NewError(ErrInvalidValue, "rootAddr does not match resource metadata")
		}
		if j.MetaHash != "" && !bytes.Equal(declaredMetaHash, r.metaHash) {
			return nil, NewError(ErrInvalidValue, "metaHash does not match resource metadata")
		}

	} else {
		//Update message
		r.rootAddr = declaredRootAddr
		r.metaHash = declaredMetaHash
	}

	if j.Signature != "" {
		sigBytes, err := hexutil.Decode(j.Signature)
		if err != nil || len(sigBytes) != signatureLength {
			return nil, NewError(ErrInvalidSignature, "Cannot decode signature")
		}
		r.signature = new(Signature)
		r.updateAddr = resourceUpdateChunkAddr(r.period, r.version, r.rootAddr)
		copy(r.signature[:], sigBytes)
	}
	return r, nil
}

// DecodeUpdateRequest takes a JSON structure stored in a byte array and returns a live UpdateRequest object
func DecodeUpdateRequest(rawData []byte) (*UpdateRequest, error) {
	var requestJSON updateRequestJSON
	if err := json.Unmarshal(rawData, &requestJSON); err != nil {
		return nil, err
	}
	return requestJSON.decode()
}

// EncodeUpdateRequest takes an update request and encodes it as a JSON structure into a byte array
func EncodeUpdateRequest(mruRequest *UpdateRequest) (rawData []byte, err error) {
	var signatureString, dataHashString, rootAddrString, metaHashString string
	if mruRequest.signature != nil {
		signatureString = hexutil.Encode(mruRequest.signature[:])
	}
	if mruRequest.data != nil {
		dataHashString = hexutil.Encode(mruRequest.data)
	}
	if mruRequest.rootAddr != nil {
		rootAddrString = hexutil.Encode(mruRequest.rootAddr)
	}
	if mruRequest.metaHash != nil {
		metaHashString = hexutil.Encode(mruRequest.metaHash)
	}

	requestJSON := &updateRequestJSON{
		Name:      mruRequest.name,
		Frequency: mruRequest.frequency,
		StartTime: mruRequest.startTime,
		Version:   mruRequest.version,
		Period:    mruRequest.period,
		OwnerAddr: hexutil.Encode(mruRequest.ownerAddr[:]),
		Data:      dataHashString,
		Multihash: mruRequest.multihash,
		Signature: signatureString,
		RootAddr:  rootAddrString,
		MetaHash:  metaHashString,
	}

	return json.Marshal(requestJSON)
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

// resourceUpdateChunkAddr calculates the resource update chunk address (formerly known as resourceHash)
func resourceUpdateChunkAddr(period uint32, version uint32, rootAddr storage.Address) storage.Address {
	hasher := hashPool.Get().(storage.SwarmHash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(NewResourceHash(period, version, rootAddr))
	return hasher.Sum(nil)
}

// NewResourceHash will create a deterministic address from the update metadata
// format is: hash(period|version|rootAddr)
func NewResourceHash(period uint32, version uint32, rootAddr storage.Address) []byte {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, period)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, version)
	buf.Write(b)
	buf.Write(rootAddr[:])
	return buf.Bytes()
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
