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

package syncer

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
	}{
		{
			po: 0, prevDepth: -1, newDepth: 0,
			subBins: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 1, prevDepth: -1, newDepth: 0,
			subBins: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 2, prevDepth: -1, newDepth: 0,
			subBins: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 0, prevDepth: -1, newDepth: 1,
			subBins: []int{0},
		},
		{
			po: 1, prevDepth: -1, newDepth: 1,
			subBins: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 2, prevDepth: -1, newDepth: 2,
			subBins: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 3, prevDepth: -1, newDepth: 2,
			subBins: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 1, prevDepth: -1, newDepth: 2,
			subBins: []int{1},
		},
		{
			po: 0, prevDepth: 0, newDepth: 0, // 0-16 -> 0-16
		},
		{
			po: 1, prevDepth: 0, newDepth: 0, // 0-16 -> 0-16
		},
		{
			po: 0, prevDepth: 0, newDepth: 1, // 0-16 -> 0
			quitBins: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 0, prevDepth: 0, newDepth: 2, // 0-16 -> 0
			quitBins: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 1, prevDepth: 0, newDepth: 1, // 0-16 -> 1-16
			quitBins: []int{0},
		},
		{
			po: 1, prevDepth: 1, newDepth: 0, // 1-16 -> 0-16
			subBins: []int{0},
		},
		{
			po: 4, prevDepth: 0, newDepth: 1, // 0-16 -> 1-16
			quitBins: []int{0},
		},
		{
			po: 4, prevDepth: 0, newDepth: 4, // 0-16 -> 4-16
			quitBins: []int{0, 1, 2, 3},
		},
		{
			po: 4, prevDepth: 0, newDepth: 5, // 0-16 -> 4
			quitBins: []int{0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 4, prevDepth: 5, newDepth: 0, // 4 -> 0-16
			subBins: []int{0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			po: 4, prevDepth: 5, newDepth: 6, // 4 -> 4
		},
	} {
		subBins, quitBins := syncSubscriptionsDiff(tc.po, tc.prevDepth, tc.newDepth, max)
		if fmt.Sprint(subBins) != fmt.Sprint(tc.subBins) {
			t.Errorf("po: %v, prevDepth: %v, newDepth: %v: got subBins %v, want %v", tc.po, tc.prevDepth, tc.newDepth, subBins, tc.subBins)
		}
		if fmt.Sprint(quitBins) != fmt.Sprint(tc.quitBins) {
			t.Errorf("po: %v, prevDepth: %v, newDepth: %v: got quitBins %v, want %v", tc.po, tc.prevDepth, tc.newDepth, quitBins, tc.quitBins)
		}
	}
}
