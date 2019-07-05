// Copyright 2019 The Swarm Authors
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

package newstream

import (
	"fmt"
	"testing"

	"github.com/ethersphere/swarm/network"
)

// TestSyncSubscriptionsDiff validates the output of syncSubscriptionsDiff
// function for various arguments.
func TestSyncSubscriptionsDiff(t *testing.T) {
	max := network.NewKadParams().MaxProxDisplay
	for _, tc := range []struct {
		po, prevDepth, newDepth int
		subBins, quitBins       []int
		syncWithinDepth         bool
	}{
		// tests for old syncBins logic that establish streams on all bins (not push-sync adjusted)
		{
			po: 0, prevDepth: -1, newDepth: 0,
			subBins:         []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: -1, newDepth: 0,
			subBins:         []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 2, prevDepth: -1, newDepth: 0,
			subBins:         []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 0, prevDepth: -1, newDepth: 1,
			subBins:         []int{0},
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: -1, newDepth: 1,
			subBins:         []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 2, prevDepth: -1, newDepth: 2,
			subBins:         []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 3, prevDepth: -1, newDepth: 2,
			subBins:         []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: -1, newDepth: 2,
			subBins:         []int{1},
			syncWithinDepth: false,
		},
		{
			po: 0, prevDepth: 0, newDepth: 0, // 0-16 -> 0-16
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: 0, newDepth: 0, // 0-16 -> 0-16
			syncWithinDepth: false,
		},
		{
			po: 0, prevDepth: 0, newDepth: 1, // 0-16 -> 0
			quitBins:        []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 0, prevDepth: 0, newDepth: 2, // 0-16 -> 0
			quitBins:        []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: 0, newDepth: 1, // 0-16 -> 1-16
			quitBins:        []int{0},
			syncWithinDepth: false,
		},
		{
			po: 1, prevDepth: 1, newDepth: 0, // 1-16 -> 0-16
			subBins:         []int{0},
			syncWithinDepth: false,
		},
		{
			po: 4, prevDepth: 0, newDepth: 1, // 0-16 -> 1-16
			quitBins:        []int{0},
			syncWithinDepth: false,
		},
		{
			po: 4, prevDepth: 0, newDepth: 4, // 0-16 -> 4-16
			quitBins:        []int{0, 1, 2, 3},
			syncWithinDepth: false,
		},
		{
			po: 4, prevDepth: 0, newDepth: 5, // 0-16 -> 4
			quitBins:        []int{0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 4, prevDepth: 5, newDepth: 0, // 4 -> 0-16
			subBins:         []int{0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: false,
		},
		{
			po: 4, prevDepth: 5, newDepth: 6, // 4 -> 4
			syncWithinDepth: false,
		},

		// tests for syncBins logic to establish streams only within depth
		{
			po: 0, prevDepth: 5, newDepth: 6,
			syncWithinDepth: true,
		},
		{
			po: 1, prevDepth: 5, newDepth: 6,
			syncWithinDepth: true,
		},
		{
			po: 7, prevDepth: 5, newDepth: 6, // 5-16 -> 6-16
			quitBins:        []int{5},
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: 5, newDepth: 6, // 5-16 -> 6-16
			quitBins:        []int{5},
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: 0, newDepth: 6, // 0-16 -> 6-16
			quitBins:        []int{0, 1, 2, 3, 4, 5},
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: -1, newDepth: 0, // [] -> 0-16
			subBins:         []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: -1, newDepth: 7, // [] -> 7-16
			subBins:         []int{7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: -1, newDepth: 10, // [] -> []
			syncWithinDepth: true,
		},
		{
			po: 9, prevDepth: 8, newDepth: 10, // 8-16 -> []
			quitBins:        []int{8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: true,
		},
		{
			po: 1, prevDepth: 0, newDepth: 0, // [] -> []
			syncWithinDepth: true,
		},
		{
			po: 1, prevDepth: 0, newDepth: 8, // 0-16 -> []
			quitBins:        []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			syncWithinDepth: true,
		},
	} {
		subBins, quitBins := syncSubscriptionsDiff(tc.po, tc.prevDepth, tc.newDepth, max, tc.syncWithinDepth)
		if fmt.Sprint(subBins) != fmt.Sprint(tc.subBins) {
			t.Errorf("po: %v, prevDepth: %v, newDepth: %v, syncWithinDepth: %t: got subBins %v, want %v", tc.po, tc.prevDepth, tc.newDepth, tc.syncWithinDepth, subBins, tc.subBins)
		}
		if fmt.Sprint(quitBins) != fmt.Sprint(tc.quitBins) {
			t.Errorf("po: %v, prevDepth: %v, newDepth: %v, syncWithinDepth: %t: got quitBins %v, want %v", tc.po, tc.prevDepth, tc.newDepth, tc.syncWithinDepth, quitBins, tc.quitBins)
		}
	}
}
