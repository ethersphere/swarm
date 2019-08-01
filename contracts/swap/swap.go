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

// Package swap wraps the 'swap' Ethereum smart contract.
// It is an abstraction layer to hide implementation details about the different
// Swap contract iterations (Simple Swap, Soft Swap, etc.)
package swap

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethersphere/swarm/contracts/swap/contract"
)

var (
	// ErrNotASwapContract is given when an address is verified not to have a SWAP contract based on its bytecode
	ErrNotASwapContract = errors.New("not a swap contract")
	// ErrTransactionReverted is given when the transaction that submits or cashes a cheque is reverted
	ErrTransactionReverted = errors.New("Transaction reverted")
)

// Backend wraps all methods required for contract deployment.
type Backend interface {
	bind.ContractBackend
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	//TODO: needed? BalanceAt(ctx context.Context, address common.Address, blockNum *big.Int) (*big.Int, error)
}

// Deploy deploys an instance of the underlying contract and returns its `Contract` abstraction
func Deploy(auth *bind.TransactOpts, backend bind.ContractBackend, owner common.Address, harddepositTimeout time.Duration) (common.Address, Contract, *types.Transaction, error) {
	addr, tx, s, err := contract.DeploySimpleSwap(auth, backend, owner, big.NewInt(int64(harddepositTimeout)))
	c := simpleContract{instance: s}
	return addr, c, tx, err
}

// InstanceAt creates a new instance of a contract at a specific address.
// It assumes that there is an existing contract instance at the given address, or an error is returned
// This function is needed to communicate with remote Swap contracts (e.g. sending a cheque)
func InstanceAt(address common.Address, backend bind.ContractBackend) (Contract, error) {
	simple, err := contract.NewSimpleSwap(address, backend)
	c := simpleContract{instance: simple}
	return c, err
}

// Contract interface defines the methods exported from the underlying go-bindings for the smart contract
type Contract interface {
	// Submit a cheque to the beneficiary
	SubmitChequeBeneficiary(opts *bind.TransactOpts, backend Backend, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Receipt, error)
	// Cash the cheque by the beneficiary
	CashChequeBeneficiary(auth *bind.TransactOpts, backend Backend, beneficiary common.Address, requestPayout *big.Int) (*types.Receipt, error)
	// Return contract info (e.g. deployed address)
	ContractParams() *Params
	// Return the contract owner from the blockchain
	Issuer(opts *bind.CallOpts) (common.Address, error)
	// Return the last cheque
	Cheques(opts *bind.CallOpts, addr common.Address) (*ChequeResult, error)
}

// ChequeResult is needed because the underlying `Cheques` method returns an untyped struct
type ChequeResult struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}

// Params encapsulates some contract parameters (currently mostly informational)
type Params struct {
	ContractCode, ContractAbi string
}

// ValidateCode checks that the on-chain code at address matches the expected swap
// contract code.
// TODO: have this as a package level function and pass the SimpleSwapBin as argument
func ValidateCode(ctx context.Context, b bind.ContractBackend, address common.Address) error {
	codeReadFromAddress, err := b.CodeAt(ctx, address, nil)
	if err != nil {
		return err
	}
	referenceCode := common.FromHex(contract.ContractDeployedCode)
	if !bytes.Equal(codeReadFromAddress, referenceCode) {
		return ErrNotASwapContract
	}
	return nil
}

// WaitFunc is the default function to wait for transactions
// We can overwrite this in tests so that we don't need to wait for mining
var WaitFunc = waitForTx

// waitForTx waits for transaction to be mined and returns the receipt
func waitForTx(auth *bind.TransactOpts, backend Backend, tx *types.Transaction) (*types.Receipt, error) {
	// it blocks here until tx is mined
	receipt, err := bind.WaitMined(auth.Context, backend, tx)
	if err != nil {
		return nil, err
	}
	// indicate whether the transaction did not revert
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, ErrTransactionReverted
	}
	return receipt, nil
}

type simpleContract struct {
	instance *contract.SimpleSwap
}

// ContractParams returns contract information
func (s simpleContract) ContractParams() *Params {
	return &Params{
		ContractCode: contract.SimpleSwapBin,
		ContractAbi:  contract.SimpleSwapABI,
	}
}

// Cheques returns the last cheque from the smart contract
func (s simpleContract) Cheques(opts *bind.CallOpts, addr common.Address) (*ChequeResult, error) {
	r, err := s.instance.Cheques(opts, addr)
	if err != nil {
		return nil, err
	}
	result := &ChequeResult{
		Serial:      r.Serial,
		Amount:      r.Amount,
		PaidOut:     r.PaidOut,
		CashTimeout: r.CashTimeout,
	}
	return result, nil
}

// Issuer returns the contract owner from the blockchain
func (s simpleContract) Issuer(opts *bind.CallOpts) (common.Address, error) {
	return s.instance.Issuer(opts)
}

// SubmitChequeBeneficiary prepares to send a call to submitChequeBeneficiary and blocks until the transaction is mined.
func (s simpleContract) SubmitChequeBeneficiary(auth *bind.TransactOpts, backend Backend, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Receipt, error) {
	tx, err := s.instance.SubmitChequeBeneficiary(auth, serial, amount, timeout, ownerSig)
	if err != nil {
		return nil, err
	}
	return WaitFunc(auth, backend, tx)
}

// CashChequeBeneficiary cashes the cheque.
func (s simpleContract) CashChequeBeneficiary(auth *bind.TransactOpts, backend Backend, beneficiary common.Address, requestPayout *big.Int) (*types.Receipt, error) {
	tx, err := s.instance.CashChequeBeneficiary(auth, beneficiary, requestPayout)
	if err != nil {
		return nil, err
	}
	return WaitFunc(auth, backend, tx)
}
