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

// SendTransactionNoCommit provides access to the underlying SendTransaction function without the auto commit
func (b *TestBackend) SendTransactionNoCommit(ctx context.Context, tx *types.Transaction) (err error) {
	return b.SimulatedBackend.SendTransaction(ctx, tx)
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
