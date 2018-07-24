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
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// updateRequestJSON represents a JSON-serialized UpdateRequest
type updateRequestJSON struct {
	ViewID    *ResourceViewID `json:"viewId"`
	Version   uint32          `json:"version,omitempty"`
	Period    uint32          `json:"period,omitempty"`
	Data      string          `json:"data,omitempty"`
	Signature string          `json:"signature,omitempty"`
}

// Request represents an update and/or resource create message
type Request struct {
	SignedResourceUpdate
	isNew bool
}

var zeroAddr = common.Address{}

// NewCreateUpdateRequest returns a ready to sign request to create and initialize a resource with data
func NewCreateUpdateRequest(metadata *ResourceID) (*Request, error) {

	request, err := NewCreateRequest(metadata)
	if err != nil {
		return nil, err
	}

	// get the current time
	now := TimestampProvider.Now().Time

	request.version = 1
	request.period, err = getNextPeriod(metadata.StartTime.Time, now, metadata.Frequency)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// NewCreateRequest returns a request to create a new resource
func NewCreateRequest(metadata *ResourceID) (request *Request, err error) {
	if metadata.StartTime.Time == 0 { // get the current time
		metadata.StartTime = TimestampProvider.Now()
	}

	request = new(Request)
	request.viewID.resourceID = *metadata
	request.isNew = true
	return request, nil
}

// Frequency returns the resource's expected update frequency
func (r *Request) Frequency() uint64 {
	return r.viewID.resourceID.Frequency
}

// Topic returns the resource's topic
func (r *Request) Topic() Topic {
	return r.viewID.resourceID.Topic
}

// Period returns in which period the resource will be published
func (r *Request) Period() uint32 {
	return r.period
}

// Version returns the resource version to publish
func (r *Request) Version() uint32 {
	return r.version
}

// ViewID returns the resource's view ID
func (r *Request) ViewID() *ResourceViewID {
	return &r.viewID
}

// StartTime returns the time that the resource was/will be created at
func (r *Request) StartTime() Timestamp {
	return r.viewID.resourceID.StartTime
}

// Owner returns the resource owner's address
func (r *Request) Owner() common.Address {
	return r.viewID.ownerAddr
}

// Sign executes the signature to validate the resource and sets the owner address field
func (r *Request) Sign(signer Signer) error {
	/*	if r.viewID.ownerAddr != zeroAddr && r.viewID.ownerAddr != signer.Address() {
		return NewError(ErrInvalidSignature, "Signer does not match current owner of the resource")
	}*/

	if err := r.SignedResourceUpdate.Sign(signer); err != nil {
		return err
	}
	return nil
}

// SetData stores the payload data the resource will be updated with
func (r *Request) SetData(data []byte) {
	r.data = data
	r.signature = nil
	/*	if !r.isNew {
			r.viewID.resourceID.Frequency = 0 // mark as update
		}
	*/
}

func (r *Request) IsNew() bool {
	return r.viewID.resourceID.Frequency > 0 && (r.period <= 1 || r.version <= 1)
}

func (r *Request) IsUpdate() bool {
	return r.signature != nil
}

// fromJSON takes an update request JSON and populates an UpdateRequest
func (r *Request) fromJSON(j *updateRequestJSON) error {

	r.version = j.Version
	r.period = j.Period
	r.viewID = *j.ViewID

	var err error
	if j.Data != "" {
		r.data, err = hexutil.Decode(j.Data)
		if err != nil {
			return NewError(ErrInvalidValue, "Cannot decode data")
		}
	}

	if j.Signature != "" {
		sigBytes, err := hexutil.Decode(j.Signature)
		if err != nil || len(sigBytes) != signatureLength {
			return NewError(ErrInvalidSignature, "Cannot decode signature")
		}
		r.signature = new(Signature)
		r.updateAddr = r.UpdateAddr()
		copy(r.signature[:], sigBytes)
	}
	return nil
}

func decodeHexArray(dst []byte, src, name string) error {
	bytes, err := decodeHexSlice(src, len(dst), name)
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

// UnmarshalJSON takes a JSON structure stored in a byte array and populates the Request object
// Implements json.Unmarshaler interface
func (r *Request) UnmarshalJSON(rawData []byte) error {
	var requestJSON updateRequestJSON
	if err := json.Unmarshal(rawData, &requestJSON); err != nil {
		return err
	}
	return r.fromJSON(&requestJSON)
}

// MarshalJSON takes an update request and encodes it as a JSON structure into a byte array
// Implements json.Marshaler interface
func (r *Request) MarshalJSON() (rawData []byte, err error) {
	var signatureString, dataString string
	if r.signature != nil {
		signatureString = hexutil.Encode(r.signature[:])
	}
	if r.data != nil {
		dataString = hexutil.Encode(r.data)
	}

	requestJSON := &updateRequestJSON{
		ViewID:    &r.viewID,
		Version:   r.version,
		Period:    r.period,
		Data:      dataString,
		Signature: signatureString,
	}

	return json.Marshal(requestJSON)
}
