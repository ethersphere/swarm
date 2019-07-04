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
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethersphere/swarm/contracts/swap/contract"
)

type Simple struct {
	Instance *contract.SimpleSwap
}

func (s *Simple) Deploy(auth *bind.TransactOpts, backend bind.ContractBackend, owner common.Address) (addr common.Address, tx *types.Transaction, err error) {
	addr, tx, s.Instance, err = contract.DeploySimpleSwap(auth, backend, owner)
	return addr, tx, err
}

func (s *Simple) ContractDeployedCode() string {
	return contract.ContractDeployedCode
}

func (s *Simple) ContractParams() *Params {
	return &Params{
		ContractCode: contract.SimpleSwapBin,
		ContractAbi:  contract.SimpleSwapABI,
	}
}

func (s *Simple) SubmitChequeBeneficiary(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error) {
	return s.Instance.SubmitChequeBeneficiary(opts, serial, amount, timeout, ownerSig)
}
