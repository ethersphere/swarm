// Copyright 2019 The go-ethereum Authors
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

package testutil

import (
	"testing"

	"github.com/ethereum/go-ethereum/swarm/chunk"
)

// CheckTag checks the first tag in the api struct to be in a certain state
func CheckTag(t *testing.T, tags *chunk.Tags, split, stored, seen, total int) {
	t.Helper()
	i := 0
	// check that the tag was created and incremented accordingly
	tags.Range(func(k, v interface{}) bool {
		vv := v.(*chunk.Tag)

		tSplit := vv.Get(chunk.SPLIT)
		if tSplit != split {
			t.Fatalf("should have had split chunks, got %d want %d", tSplit, split)
		}

		tSeen := vv.Get(chunk.SEEN)
		if tSeen != seen {
			t.Fatalf("should have had seen chunks, got %d want %d", tSeen, seen)
		}

		tStored := vv.Get(chunk.STORED)
		if tStored != stored {
			t.Fatalf("mismatch stored chunks, got %d want %d", tStored, stored)
		}

		tTotal := vv.Total()
		if tTotal != total {
			t.Fatalf("mismatch total chunks, got %d want %d", tTotal, total)
		}
		i++

		return false
	})
	if i == 0 {
		t.Fatal("no tags found")
	}
}
