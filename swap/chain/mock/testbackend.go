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
package mock

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core/types"
)

// TestBackend is the backend to use for tests with a simulated backend
type TestBackend struct {
	*backends.SimulatedBackend
}

// SendTransaction adds a commit after a successful send
func (b *TestBackend) SendTransaction(ctx context.Context, tx *types.Transaction) (err error) {
	err = b.SimulatedBackend.SendTransaction(ctx, tx)
	if err == nil {
		b.SimulatedBackend.Commit()
	}
	return err
}

// Close overrides the Close function of the underlying SimulatedBackend so that it does nothing
// This allows the same SimulatedBackend backend to be reused across tests
// This is necessary due to some memory leakage issues with the used version of the SimulatedBackend
func (b *TestBackend) Close() {

}

// NewTestBackend returns a new TestBackend for the given SimulatedBackend
// It also causes an initial commit to make sure that genesis accounts are set up
func NewTestBackend(backend *backends.SimulatedBackend) *TestBackend {
	backend.Commit()
	return &TestBackend{
		SimulatedBackend: backend,
	}
}
