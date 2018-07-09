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
	metadata := ResourceMetadata{
		Name: "world news report, every hour, on the hour",
		StartTime: Timestamp{
			Time: 1528880400,
		},
		Frequency: 3600,
		Owner:     ownerAddr,
	}

	rootAddr, metaHash, chunkData, err := metadata.serializeAndHash() // creates hashes and marshals, in one go
	if err != nil {
		t.Fatal(err)
	}
	const expectedRootAddr = "0xfb0ed7efa696bdb0b54cd75554cc3117ffc891454317df7dd6fefad978e2f2fb"
	const expectedMetaHash = "0xf74a10ce8f26ffc8bfaa07c3031a34b2c61f517955e7deb1592daccf96c69cf0"
	const expectedChunkData = "0x00004f0010dd205b00000000100e0000000000002a776f726c64206e657773207265706f72742c20657665727920686f75722c206f6e2074686520686f7572876a8936a7cd0b79ef0735ad0896c1afe278781c"

	compareByteSliceToExpectedHex(t, "rootAddr", rootAddr, expectedRootAddr)
	compareByteSliceToExpectedHex(t, "metaHash", metaHash, expectedMetaHash)
	compareByteSliceToExpectedHex(t, "chunkData", chunkData, expectedChunkData)

	recoveredMetadata := ResourceMetadata{}
	recoveredMetadata.binaryGet(chunkData)

	if recoveredMetadata != metadata {
		t.Fatalf("Expected that the recovered metadata equals the marshalled metadata")
	}
}
