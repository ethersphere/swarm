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
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// updateRequestJSON represents a JSON-serialized UpdateRequest
type updateRequestJSON struct {
	Name           string `json:"name,omitempty"`
	Frequency      uint64 `json:"frequency,omitempty"`
	StartTime      uint64 `json:"startTime,omitempty"`
	StartTimeProof string `json:"startTimeProof,omitempty"`
	OwnerAddr      string `json:"ownerAddr,omitempty"`
	RootAddr       string `json:"rootAddr,omitempty"`
	MetaHash       string `json:"metaHash,omitempty"`
	Version        uint32 `json:"version"`
	Period         uint32 `json:"period"`
	Data           string `json:"data,omitempty"`
	Multihash      bool   `json:"multiHash"`
	Signature      string `json:"signature,omitempty"`
}

// Request represents an update and/or resource create message
type Request struct {
	SignedResourceUpdate
	metadata ResourceMetadata
}

var zeroAddr = common.Address{}

// NewCreateRequest returns a ready to sign Request message to create a new resource
func NewCreateRequest(metadata *ResourceMetadata) (*Request, error) {

	// get the current time
	if metadata.StartTime.Time == 0 {
		metadata.StartTime.Time = uint64(time.Now().Unix())
	}

	if metadata.OwnerAddr == zeroAddr {
		return nil, NewError(ErrInvalidValue, "OwnerAddr is not set")
	}

	request := &Request{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				updateHeader: updateHeader{
					UpdateLookup: UpdateLookup{
						version: 1,
						period:  1,
					},
				},
			},
		},
		metadata: *metadata,
	}

	var err error
	request.rootAddr, request.metaHash, _, err = request.metadata.hashAndSerialize()
	if err != nil {
		return nil, err
	}
	return request, nil
}

// Frequency returns the resource's expected update frequency
func (r *Request) Frequency() uint64 {
	return r.metadata.Frequency
}

// Name returns the resource human-readable name
func (r *Request) Name() string {
	return r.metadata.Name
}

// Multihash returns true if the resource data should be interpreted as a multihash
func (r *Request) Multihash() bool {
	return r.multihash
}

// Period returns in which period the resource will be published
func (r *Request) Period() uint32 {
	return r.period
}

// Version returns the resource version to publish
func (r *Request) Version() uint32 {
	return r.version
}

// RootAddr returns the metadata chunk address
func (r *Request) RootAddr() storage.Address {
	return r.rootAddr
}

// StartTime returns the time that the resource was/will be created at
func (r *Request) StartTime() Timestamp {
	return r.metadata.StartTime
}

// OwnerAddr returns the resource owner's address
func (r *Request) OwnerAddr() common.Address {
	return r.metadata.OwnerAddr
}

// Sign executes the signature to validate the resource and sets the owner address field
func (r *Request) Sign(signer Signer) error {
	if r.metadata.OwnerAddr != zeroAddr && r.metadata.OwnerAddr != signer.Address() {
		return NewError(ErrInvalidSignature, "Signer does not match current owner of the resource")
	}

	if err := r.SignedResourceUpdate.Sign(signer); err != nil {
		return err
	}
	r.metadata.OwnerAddr = signer.Address()
	return nil
}

// SetData stores the payload data the resource will be updated with
func (r *Request) SetData(data []byte, multihash bool) {
	r.data = data
	r.multihash = multihash
	r.dirty()
}

func (r *Request) dirty() {
	r.signature = nil
	if r.period != 1 || r.version != 1 {
		r.metadata.Frequency = 0 // mark as update if this is not the first request
	}
}

func (r *Request) IsNew() bool {
	return r.metadata.Frequency > 0 && (r.period <= 1 || r.version <= 1)
}

func (r *Request) IsUpdate() bool {
	return r.signature != nil
}

// decode takes an update request JSON and returns an UpdateRequest
func (j *updateRequestJSON) decode() (*Request, error) {

	r := &Request{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				updateHeader: updateHeader{
					UpdateLookup: UpdateLookup{
						version: j.Version,
						period:  j.Period,
					},
					multihash: j.Multihash,
				},
			},
		},
		metadata: ResourceMetadata{
			Name:      j.Name,
			Frequency: j.Frequency,
			StartTime: Timestamp{
				Time: j.StartTime,
			},
		},
	}

	if err := decodeHexArray(r.metadata.StartTime.Proof[:], j.StartTimeProof, common.HashLength, "startTimeProof"); err != nil {
		return nil, err
	}

	if err := decodeHexArray(r.metadata.OwnerAddr[:], j.OwnerAddr, common.AddressLength, "ownerAddr"); err != nil {
		return nil, err
	}

	var err error
	if j.Data != "" {
		r.data, err = hexutil.Decode(j.Data)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode data")
		}
	}

	var declaredRootAddr storage.Address
	var declaredMetaHash []byte

	declaredRootAddr, err = decodeHexSlice(j.RootAddr, storage.KeyLength, "rootAddr")
	if err != nil {
		return nil, err
	}

	declaredMetaHash, err = decodeHexSlice(j.MetaHash, 32, "metaHash")
	if err != nil {
		return nil, err
	}

	if r.IsNew() {
		// for new resource creation, rootAddr and metaHash are optional because
		// we can derive them from the content itself.
		// however, if the user sent them, we check them for consistency.

		r.rootAddr, r.metaHash, _, err = r.metadata.hashAndSerialize()
		if err != nil {
			return nil, err
		}
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
		r.updateAddr = r.Addr()
		copy(r.signature[:], sigBytes)
	}
	return r, nil
}

func decodeHexArray(dst []byte, src string, expectedLength int, name string) error {
	bytes, err := decodeHexSlice(src, expectedLength, name)
	if err != nil {
		return err
	}
	if bytes != nil {
		copy(dst, bytes)
	}
	return nil
}

func decodeHexSlice(src string, expectedLength int, name string) (bytes []byte, err error) {
	if src != "" {
		bytes, err = hexutil.Decode(src)
		if err != nil || len(bytes) != expectedLength {
			return nil, NewErrorf(ErrInvalidValue, "Cannot decode %s", name)
		}
	}
	return bytes, nil
}

// DecodeUpdateRequest takes a JSON structure stored in a byte array and returns a live UpdateRequest object
func DecodeUpdateRequest(rawData []byte) (*Request, error) {
	var requestJSON updateRequestJSON
	if err := json.Unmarshal(rawData, &requestJSON); err != nil {
		return nil, err
	}
	return requestJSON.decode()
}

// EncodeUpdateRequest takes an update request and encodes it as a JSON structure into a byte array
func EncodeUpdateRequest(updateRequest *Request) (rawData []byte, err error) {
	var signatureString, dataHashString, rootAddrString, metaHashString string
	if updateRequest.signature != nil {
		signatureString = hexutil.Encode(updateRequest.signature[:])
	}
	if updateRequest.data != nil {
		dataHashString = hexutil.Encode(updateRequest.data)
	}
	if updateRequest.rootAddr != nil {
		rootAddrString = hexutil.Encode(updateRequest.rootAddr)
	}
	if updateRequest.metaHash != nil {
		metaHashString = hexutil.Encode(updateRequest.metaHash)
	}
	var ownerAddrString string
	if updateRequest.metadata.Frequency == 0 {
		ownerAddrString = ""
	} else {
		ownerAddrString = hexutil.Encode(updateRequest.metadata.OwnerAddr[:])
	}
	var startTimeProofString string
	if updateRequest.metadata.StartTime.Time == 0 {
		startTimeProofString = ""
	} else {
		startTimeProofString = hexutil.Encode(updateRequest.metadata.StartTime.Proof[:])
	}

	requestJSON := &updateRequestJSON{
		Name:           updateRequest.metadata.Name,
		Frequency:      updateRequest.metadata.Frequency,
		StartTime:      updateRequest.metadata.StartTime.Time,
		StartTimeProof: startTimeProofString,
		Version:        updateRequest.version,
		Period:         updateRequest.period,
		OwnerAddr:      ownerAddrString,
		Data:           dataHashString,
		Multihash:      updateRequest.multihash,
		Signature:      signatureString,
		RootAddr:       rootAddrString,
		MetaHash:       metaHashString,
	}

	return json.Marshal(requestJSON)
}
