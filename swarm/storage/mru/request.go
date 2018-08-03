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

//TODO AFTER PR REVIEW: Merge this file with signedupdate.go

// updateRequestJSON represents a JSON-serialized UpdateRequest
type updateRequestJSON struct {
	View      *View  `json:"view"`
	Version   uint32 `json:"version,omitempty"`
	Period    uint32 `json:"period,omitempty"`
	Data      string `json:"data,omitempty"`
	Signature string `json:"signature,omitempty"`
}

var zeroAddr = common.Address{}

// NewCreateUpdateRequest returns a ready to sign request to create and initialize a resource with data
func NewCreateUpdateRequest(metadata *Resource) (*Request, error) {

	request, err := NewCreateRequest(metadata, zeroAddr)
	if err != nil {
		return nil, err
	}

	// get the current time
	now := TimestampProvider.Now().Time

	request.Version = 1
	request.Period, err = getNextPeriod(metadata.StartTime.Time, now, metadata.Frequency)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// NewCreateRequest returns a request to create a new resource
func NewCreateRequest(metadata *Resource, userAddr common.Address) (request *Request, err error) {
	if metadata.StartTime.Time == 0 { // get the current time
		metadata.StartTime = TimestampProvider.Now()
	}

	request = new(Request)
	request.View.Resource = *metadata
	request.View.User = userAddr
	return request, nil
}

// SetData stores the payload data the resource will be updated with
func (r *Request) SetData(data []byte) {
	r.data = data
	r.Signature = nil
}

// IsUpdate returns true if this request models a signed update or otherwise it is a signature request
func (r *Request) IsUpdate() bool {
	return r.Signature != nil
}

// fromJSON takes an update request JSON and populates an UpdateRequest
func (r *Request) fromJSON(j *updateRequestJSON) error {

	r.Version = j.Version
	r.Period = j.Period
	r.View = *j.View

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
		r.Signature = new(Signature)
		r.updateAddr = r.UpdateAddr()
		copy(r.Signature[:], sigBytes)
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
	if r.Signature != nil {
		signatureString = hexutil.Encode(r.Signature[:])
	}
	if r.data != nil {
		dataString = hexutil.Encode(r.data)
	}

	requestJSON := &updateRequestJSON{
		View:      &r.View,
		Version:   r.Version,
		Period:    r.Period,
		Data:      dataString,
		Signature: signatureString,
	}

	return json.Marshal(requestJSON)
}
