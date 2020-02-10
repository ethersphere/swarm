package chain

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Backend is the minimum amount of functionality required by the underlying ethereum backend
type Backend interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	TransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
}

// TestBackend is the backend to use for tests with a simulated backend
type TestBackend struct {
	*backends.SimulatedBackend
}

// WaitMined waits until either the transaction with the given hash has been mined or the context is cancelled
func WaitMined(ctx context.Context, b Backend, hash common.Hash) (*types.Receipt, error) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	if ctx == nil {
		ctx = context.Background()
	}

	for {
		receipt, err := b.TransactionReceipt(ctx, hash)
		if receipt != nil {
			return receipt, nil
		}
		if err != nil {
			log.Trace("Receipt retrieval failed", "err", err)
		} else {
			log.Trace("Transaction not yet mined")
		}
		// Wait for the next round.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
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
