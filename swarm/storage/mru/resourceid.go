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
	"encoding/binary"
)

// Resource encapsulates the immutable information about a mutable resource :)
type Resource struct {
	StartTime Timestamp `json:"startTime"` // time at which the resource starts to be valid
	Frequency uint64    `json:"frequency"` // expected update frequency for the resource
	Topic     Topic     `json:"topic"`     // resource topic, for the reference of the user, to disambiguate resources with same starttime, frequency or to reference another hash
}

const frequencyLength = 8 // sizeof(uint64)
const nameLengthLength = 1

// ResourceID Layout
// StartTime Timestamp: timestampLength bytes
// frequency: frequencyLength bytes
// TopicLength: topicLength bytes
const ResourceLength = timestampLength + frequencyLength + topicLength

// binaryGet populates the resource metadata from a byte array
func (r *Resource) binaryGet(serializedData []byte) error {
	if len(serializedData) != ResourceLength {
		return NewErrorf(ErrInvalidValue, "Resource to deserialize has an invalid length. Expected it to be exactly %d. Got %d.", ResourceLength, len(serializedData))
	}

	var cursor int
	if err := r.StartTime.binaryGet(serializedData[cursor : cursor+timestampLength]); err != nil {
		return err
	}
	cursor += timestampLength

	r.Frequency = binary.LittleEndian.Uint64(serializedData[cursor : cursor+frequencyLength])
	cursor += frequencyLength

	copy(r.Topic.content[:], serializedData[cursor:cursor+topicLength])
	cursor += topicLength
	return nil
}

// binaryPut encodes the metadata into a byte array
func (r *Resource) binaryPut(serializedData []byte) error {
	if len(serializedData) != ResourceLength {
		return NewErrorf(ErrInvalidValue, "Resource to serialize has an invalid length. Expected it to be exactly %d. Got %d.", ResourceLength, len(serializedData))
	}
	var cursor int
	r.StartTime.binaryPut(serializedData[cursor : cursor+timestampLength])
	cursor += timestampLength

	binary.LittleEndian.PutUint64(serializedData[cursor:cursor+frequencyLength], r.Frequency)
	cursor += frequencyLength

	copy(serializedData[cursor:cursor+topicLength], r.Topic.content[:topicLength])
	cursor += topicLength

	return nil
}

func (r *Resource) binaryLength() int {
	return ResourceLength
}
