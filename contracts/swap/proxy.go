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

// Package swap package wraps the 'swap' Ethereum smart contract.
// It is an abstraction layer to hide implementation details about the different
// Swap contract iterations (SimpleSwap, Swap, etc.)
package swap

import (
	"bytes"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Backend wraps all methods required for chequebook operation.
type Backend interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	//BalanceAt(ctx context.Context, address common.Address, blockNum *big.Int) (*big.Int, error)
}

type Proxy struct {
	Wrapper Wrapper
}

func NewProxy() *Proxy {
	return &Proxy{
		Wrapper: &Simple{},
	}
}

// Proxy wraps all methods required for swap operation.
type Wrapper interface {
	Deploy(auth *bind.TransactOpts, backend bind.ContractBackend, owner common.Address) (common.Address, *types.Transaction, error)
	SubmitChequeBeneficiary(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error)
	ContractDeployedCode() string // TODO: needed?
	ContractParams() *Params
}

type Params struct {
	ContractCode, ContractAbi string
}

// ValidateCode checks that the on-chain code at address matches the expected swap
// contract code. This is used to detect suicided contracts.
func (a *Proxy) ValidateCode(ctx context.Context, b bind.ContractBackend, address common.Address) (bool, error) {
	code, err := b.CodeAt(ctx, address, nil)
	if err != nil {
		return false, err
	}
	//TODO: which is ContractDeployedCode and how to set it?
	return bytes.Equal(code, common.FromHex(a.Wrapper.ContractDeployedCode())), nil
}

func (a *Proxy) Deploy(auth *bind.TransactOpts, backend bind.ContractBackend, owner common.Address) (common.Address, *types.Transaction, error) {
	return a.Wrapper.Deploy(auth, backend, owner)
}

func (a *Proxy) ContractParams() *Params {
	return a.Wrapper.ContractParams()
}

func (a *Proxy) SubmitCheque(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error) {
	return a.Wrapper.SubmitChequeBeneficiary(opts, serial, amount, timeout, ownerSig)
}
