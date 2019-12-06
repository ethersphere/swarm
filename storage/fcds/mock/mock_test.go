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

package mock_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/ethersphere/swarm/storage/fcds/mock"
	"github.com/ethersphere/swarm/storage/fcds/test"
	"github.com/ethersphere/swarm/storage/mock/mem"
)

// TestFCDS runs a standard series of tests on mock Store implementation.
func TestFCDS(t *testing.T) {
	test.RunAll(t, func(t *testing.T) (fcds.Interface, func()) {
		return mock.NewStore(
			mem.NewGlobalStore().NewNodeStore(
				common.BytesToAddress(make([]byte, 20)),
			),
		), func() {}
	})
}
