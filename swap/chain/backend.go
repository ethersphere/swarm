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

package chain

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
			// some clients treat an unconfirmed transaction as an error, other simply return null
			log.Trace("receipt retrieval failed", "err", err)
		}
		if receipt != nil {
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
