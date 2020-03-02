// Copyright 2020 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package fcds

import (
	"sort"
	"testing"
)

func TestShardSlotSort(t *testing.T) {

	for _, tc := range []struct {
		freeSlots   []int // how many free slots in which shard (slice index denotes shard id, value denotes number of free slots.
		expectOrder []int // the order of bins expected to show up (slice value denotes shard id).
	}{
		{
			freeSlots:   []int{0, 0, 0, 0},
			expectOrder: []int{0, 1, 2, 3},
		},
		{
			freeSlots:   []int{0, 1, 0, 0},
			expectOrder: []int{1, 0, 2, 3},
		},
		{
			freeSlots:   []int{0, 0, 2, 0},
			expectOrder: []int{2, 0, 1, 3},
		},
		{
			freeSlots:   []int{0, 0, 0, 1},
			expectOrder: []int{3, 0, 1, 2},
		},
		{
			freeSlots:   []int{1, 1, 0, 0},
			expectOrder: []int{0, 1, 2, 3},
		},
		{
			freeSlots:   []int{1, 0, 0, 1},
			expectOrder: []int{0, 3, 1, 2},
		},
		{
			freeSlots:   []int{1, 2, 0, 0},
			expectOrder: []int{1, 0, 2, 3},
		},
		{
			freeSlots:   []int{0, 3, 2, 1},
			expectOrder: []int{1, 2, 3, 0},
		},
	} {
		s := make([]ShardSlot, len(tc.freeSlots))

		for i, v := range tc.freeSlots {
			s[i] = ShardSlot{Shard: uint8(i), Slots: int64(v)}
		}
		sort.Sort(bySlots(s))

		for i, v := range s {
			if v.Shard != uint8(tc.expectOrder[i]) {
				t.Fatalf("expected shard index %d to be %d but got %d", i, tc.expectOrder[i], v.Shard)
			}
		}

	}
}
