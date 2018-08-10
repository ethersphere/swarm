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

// Resource encapsulates the immutable information about a mutable resource :)
type Resource struct {
	Topic Topic `json:"topic"` // resource topic, for the reference of the user, to disambiguate resources with same starttime, frequency or to reference another hash
}

// ResourceID Layout
// TopicLength: topicLength bytes

// ResourceLength returns the byte length of the Resource structure
const ResourceLength = TopicLength

// binaryGet populates the resource metadata from a byte array
func (r *Resource) binaryGet(serializedData []byte) error {
	if len(serializedData) != ResourceLength {
		return NewErrorf(ErrInvalidValue, "Resource to deserialize has an invalid length. Expected it to be exactly %d. Got %d.", ResourceLength, len(serializedData))
	}

	var cursor int
	copy(r.Topic.content[:], serializedData[cursor:cursor+TopicLength])
	cursor += TopicLength
	return nil
}

// binaryPut encodes the metadata into a byte array
func (r *Resource) binaryPut(serializedData []byte) error {
	if len(serializedData) != ResourceLength {
		return NewErrorf(ErrInvalidValue, "Resource to serialize has an invalid length. Expected it to be exactly %d. Got %d.", ResourceLength, len(serializedData))
	}
	var cursor int
	copy(serializedData[cursor:cursor+TopicLength], r.Topic.content[:TopicLength])
	cursor += TopicLength

	return nil
}

func (r *Resource) binaryLength() int {
	return ResourceLength
}
