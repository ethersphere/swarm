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

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

func getTestUpdateLookup() *UpdateLookup {
	return &UpdateLookup{
		View:  *getTestResourceView(),
		Epoch: lookup.GetFirstEpoch(1000),
	}
}

func getTestLookupParams() *LookupParams {
	ul := getTestUpdateLookup()
	return &LookupParams{
		TimeLimit: 5000,
		View:      ul.View,
		Hint:      ul.Epoch,
	}
}

func TestUpdateLookupUpdateAddr(t *testing.T) {
	ul := getTestUpdateLookup()
	updateAddr := ul.UpdateAddr()
	compareByteSliceToExpectedHex(t, "updateAddr", updateAddr, "0x8b24583ec293e085f4c78aaee66d1bc5abfb8b4233304d14a349afa57af2a783")
}

func TestUpdateLookupSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestUpdateLookup(), "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781ce803000000000019")
}

func TestUpdateLookupLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestUpdateLookup())
}

func TestLookupParamsValues(t *testing.T) {
	var expected = KV{"hint.level": "25", "hint.time": "1000", "time": "5000", "topic": "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000", "user": "0x876A8936A7Cd0b79Ef0735AD0896c1AFe278781c"}

	lp := getTestLookupParams()
	testValueSerializer(t, lp, expected)

}
