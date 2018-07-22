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
	"testing"
)

func getTestResourceID() *ResourceID {
	return &ResourceID{
		Topic: NewTopic("world news report, every hour", nil),
		StartTime: Timestamp{
			Time: 1528880400,
		},
		Frequency: 3600,
	}
}

func TestResourceIDSerializerDeserializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestResourceID(), "0x10dd205b00000000100e000000000000776f726c64206e657773207265706f72742c20657665727920686f7572000000")
}

func TestResourceIDSerializerLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestResourceID())
}
