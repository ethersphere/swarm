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
func CheckTag(t *testing.T, tags *chunk.Tags, state chunk.State, exp, expTotal int) {
	t.Helper()
	i := 0
	tags.Range(func(k, v interface{}) bool {
		i++
		tag := v.(*chunk.Tag)
		count, total, err := tag.Status(state)
		if err != nil {
			t.Fatal(err)
		}

		if count != exp {
			t.Fatalf("expected count to be %d, got %d", exp, count)
		}

		if total != expTotal {
			t.Fatalf("expected total to be %d, got %d", expTotal, total)
		}
		return false
	})

	if i == 0 {
		t.Fatal("did not find any tags")
	}
}
