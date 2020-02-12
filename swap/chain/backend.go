package chain

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// ErrTransactionReverted is given when the transaction that cashes a cheque is reverted
	ErrTransactionReverted = errors.New("Transaction reverted")
)

// Backend is the minimum amount of functionality required by the underlying ethereum backend
type Backend interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	TransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
}

// WaitMined waits until either the transaction with the given hash has been mined or the context is cancelled
// this is an adapted version of go-ethereums bind.WaitMined
func WaitMined(ctx context.Context, b Backend, hash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := b.TransactionReceipt(ctx, hash)
		if err != nil {
			log.Error("receipt retrieval failed", "err", err)
		}
		if receipt != nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return nil, ErrTransactionReverted
			}
			return receipt, nil
		}

		log.Trace("transaction not yet mined", "tx", hash)
		// Wait for the next round.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
}
