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
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// updateRequestJSON represents a JSON-serialized UpdateRequest
type updateRequestJSON struct {
	Name      string `json:"name,omitempty"`
	Frequency uint64 `json:"frequency,omitempty"`
	StartTime uint64 `json:"startTime,omitempty"`
	OwnerAddr string `json:"ownerAddr,omitempty"`
	RootAddr  string `json:"rootAddr,omitempty"`
	MetaHash  string `json:"metaHash,omitempty"`
	Version   uint32 `json:"version"`
	Period    uint32 `json:"period"`
	Data      string `json:"data,omitempty"`
	Multihash bool   `json:"multiHash"`
	Signature string `json:"signature,omitempty"`
}

// Request represents an update and/or resource create message
type Request struct {
	SignedResourceUpdate
	resourceMetadata
}

// NewCreateRequest returns a ready to sign Request message to create a new resource
func NewCreateRequest(name string, frequency, startTime uint64, ownerAddr common.Address, data []byte, multihash bool) (*Request, error) {
	if !isSafeName(name) {
		return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s' when creating a new UpdateRequest", name))
	}

	// get the current time
	if startTime == 0 {
		startTime = uint64(time.Now().Unix())
	}

	request := &Request{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				updateHeader: updateHeader{
					UpdateLookup: UpdateLookup{
						version: 1,
						period:  1,
					},
					multihash: multihash,
				},
				data: data,
			},
		},
		resourceMetadata: resourceMetadata{
			name:      name,
			frequency: frequency,
			startTime: startTime,
			ownerAddr: ownerAddr,
		},
	}

	request.rootAddr, request.metaHash, _ = request.resourceMetadata.hash()

	return request, nil
}

// Frequency Returns the resource expected update frequency
func (r *Request) Frequency() uint64 {
	return r.frequency
}

// Name returns the resource human-readable name
func (r *Request) Name() string {
	return r.name
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
func (r *Request) StartTime() uint64 {
	return r.startTime
}

// OwnerAddr returns the resource owner's address
func (r *Request) OwnerAddr() common.Address {
	return r.ownerAddr
}

// Sign executes the signature to validate the resource and sets the owner address field
func (r *Request) Sign(signer Signer) error {
	if err := r.SignedResourceUpdate.Sign(signer); err != nil {
		return err
	}
	r.ownerAddr = signer.Address()
	return nil
}

// SetData stores the payload data the resource will be updated with
func (r *Request) SetData(data []byte) {
	r.signature = nil
	r.data = data
	r.frequency = 0 //mark as update
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
		resourceMetadata: resourceMetadata{
			name:      j.Name,
			frequency: j.Frequency,
			startTime: j.StartTime,
		},
	}

	if j.OwnerAddr != "" {
		ownerAddrBytes, err := hexutil.Decode(j.OwnerAddr)
		if err != nil || len(ownerAddrBytes) != common.AddressLength {
			return nil, NewError(ErrInvalidValue, "Cannot decode ownerAddr")
		}
		copy(r.ownerAddr[:], ownerAddrBytes)
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
		// for new resource creation, rootAddr and metaHash are optional because
		// we can derive them from the content itself.
		// however, if the user sent them, we check them for consistency.

		// make sure name only contains ascii values
		if !isSafeName(j.Name) {
			return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s'", j.Name))
		}
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
		r.updateAddr = r.GetUpdateAddr()
		copy(r.signature[:], sigBytes)
	}
	return r, nil
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
	if updateRequest.frequency == 0 {
		ownerAddrString = ""
	} else {
		ownerAddrString = hexutil.Encode(updateRequest.ownerAddr[:])
	}

	requestJSON := &updateRequestJSON{
		Name:      updateRequest.name,
		Frequency: updateRequest.frequency,
		StartTime: updateRequest.startTime,
		Version:   updateRequest.version,
		Period:    updateRequest.period,
		OwnerAddr: ownerAddrString,
		Data:      dataHashString,
		Multihash: updateRequest.multihash,
		Signature: signatureString,
		RootAddr:  rootAddrString,
		MetaHash:  metaHashString,
	}

	return json.Marshal(requestJSON)
}
