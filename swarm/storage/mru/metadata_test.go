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

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func compareByteSliceToExpectedHex(t *testing.T, variableName string, actualValue []byte, expectedHex string) {
	if hexutil.Encode(actualValue) != expectedHex {
		t.Fatalf("Expected %s to be %s, got %s", variableName, expectedHex, hexutil.Encode(actualValue))
	}
}

func TestMarshallingAndUnmarshalling(t *testing.T) {
	ownerAddr := newCharlieSigner().Address()
	metadata := resourceMetadata{
		name:      "world news report, every hour, on the hour",
		startTime: 1528880400,
		frequency: 3600,
		ownerAddr: ownerAddr,
	}

	rootAddr, metaHash, chunkData := metadata.hash() // creates hashes and marshals, in one go

	const expectedRootAddr = "0xa884c9583d9f86e8009bfd5fe7d892790071c2d6cf8acd2c3e16e5f17e9b143e"
	const expectedMetaHash = "0x38e401814e98b251612e40f070fddb756315705fa8f674b8ab00b2b5fa091988"
	const expectedChunkData = "0x00004e000000000010dd205b00000000100e000000000000776f726c64206e657773207265706f72742c20657665727920686f75722c206f6e2074686520686f7572876a8936a7cd0b79ef0735ad0896c1afe278781c"

	compareByteSliceToExpectedHex(t, "rootAddr", rootAddr, expectedRootAddr)
	compareByteSliceToExpectedHex(t, "metaHash", metaHash, expectedMetaHash)
	compareByteSliceToExpectedHex(t, "chunkData", chunkData, expectedChunkData)

	recoveredMetadata := resourceMetadata{}
	recoveredMetadata.unmarshalBinary(chunkData)

	if recoveredMetadata != metadata {
		t.Fatalf("Expected that the recovered metadata equals the marshalled metadata")
	}
}
