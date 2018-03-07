// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package main

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TriggerABI is the input ABI used to generate the binding from.
const TriggerABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_trigger\",\"type\":\"uint256\"}],\"name\":\"trigger\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"type\":\"constructor\"},{\"payable\":false,\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_trigger\",\"type\":\"uint256\"}],\"name\":\"TriggerEvt\",\"type\":\"event\"}]"

// TriggerBin is the compiled bytecode used for deploying new contracts.
const TriggerBin = `0x6060604052341561000f57600080fd5b5b60008054600160a060020a03191633600160a060020a03161790555b5b6101538061003c6000396000f300606060405236156100495763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663893d20e88114610061578063ed684cc61461009d575b341561005457600080fd5b61005f5b600080fd5b565b005b341561006c57600080fd5b6100746100b5565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200160405180910390f35b34156100a857600080fd5b61005f6004356100d2565b005b60005473ffffffffffffffffffffffffffffffffffffffff165b90565b7f7453df022b3c775a1d8aad3cd61495415e1799d0e8fb0462baf8ef58e6797a4b338260405173ffffffffffffffffffffffffffffffffffffffff909216825260208201526040908101905180910390a15b505600a165627a7a72305820772d358084f9458a185cf12f7befe153592c2e6fa2a7c405d1c62457e44c54f80029`

// DeployTrigger deploys a new Ethereum contract, binding an instance of Trigger to it.
func DeployTrigger(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Trigger, error) {
	parsed, err := abi.JSON(strings.NewReader(TriggerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TriggerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Trigger{TriggerCaller: TriggerCaller{contract: contract}, TriggerTransactor: TriggerTransactor{contract: contract}}, nil
}

// Trigger is an auto generated Go binding around an Ethereum contract.
type Trigger struct {
	TriggerCaller     // Read-only binding to the contract
	TriggerTransactor // Write-only binding to the contract
}

// TriggerCaller is an auto generated read-only Go binding around an Ethereum contract.
type TriggerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TriggerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TriggerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TriggerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TriggerSession struct {
	Contract     *Trigger          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TriggerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TriggerCallerSession struct {
	Contract *TriggerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// TriggerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TriggerTransactorSession struct {
	Contract     *TriggerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// TriggerRaw is an auto generated low-level Go binding around an Ethereum contract.
type TriggerRaw struct {
	Contract *Trigger // Generic contract binding to access the raw methods on
}

// TriggerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TriggerCallerRaw struct {
	Contract *TriggerCaller // Generic read-only contract binding to access the raw methods on
}

// TriggerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TriggerTransactorRaw struct {
	Contract *TriggerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTrigger creates a new instance of Trigger, bound to a specific deployed contract.
func NewTrigger(address common.Address, backend bind.ContractBackend) (*Trigger, error) {
	contract, err := bindTrigger(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Trigger{TriggerCaller: TriggerCaller{contract: contract}, TriggerTransactor: TriggerTransactor{contract: contract}}, nil
}

// NewTriggerCaller creates a new read-only instance of Trigger, bound to a specific deployed contract.
func NewTriggerCaller(address common.Address, caller bind.ContractCaller) (*TriggerCaller, error) {
	contract, err := bindTrigger(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &TriggerCaller{contract: contract}, nil
}

// NewTriggerTransactor creates a new write-only instance of Trigger, bound to a specific deployed contract.
func NewTriggerTransactor(address common.Address, transactor bind.ContractTransactor) (*TriggerTransactor, error) {
	contract, err := bindTrigger(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &TriggerTransactor{contract: contract}, nil
}

// bindTrigger binds a generic wrapper to an already deployed contract.
func bindTrigger(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TriggerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Trigger *TriggerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Trigger.Contract.TriggerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Trigger *TriggerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Trigger.Contract.TriggerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Trigger *TriggerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Trigger.Contract.TriggerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Trigger *TriggerCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Trigger.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Trigger *TriggerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Trigger.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Trigger *TriggerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Trigger.Contract.contract.Transact(opts, method, params...)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Trigger *TriggerCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Trigger.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Trigger *TriggerSession) GetOwner() (common.Address, error) {
	return _Trigger.Contract.GetOwner(&_Trigger.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Trigger *TriggerCallerSession) GetOwner() (common.Address, error) {
	return _Trigger.Contract.GetOwner(&_Trigger.CallOpts)
}

// Trigger is a paid mutator transaction binding the contract method 0xed684cc6.
//
// Solidity: function trigger(_trigger uint256) returns()
func (_Trigger *TriggerTransactor) Trigger(opts *bind.TransactOpts, _trigger *big.Int) (*types.Transaction, error) {
	return _Trigger.contract.Transact(opts, "trigger", _trigger)
}

// Trigger is a paid mutator transaction binding the contract method 0xed684cc6.
//
// Solidity: function trigger(_trigger uint256) returns()
func (_Trigger *TriggerSession) Trigger(_trigger *big.Int) (*types.Transaction, error) {
	return _Trigger.Contract.Trigger(&_Trigger.TransactOpts, _trigger)
}

// Trigger is a paid mutator transaction binding the contract method 0xed684cc6.
//
// Solidity: function trigger(_trigger uint256) returns()
func (_Trigger *TriggerTransactorSession) Trigger(_trigger *big.Int) (*types.Transaction, error) {
	return _Trigger.Contract.Trigger(&_Trigger.TransactOpts, _trigger)
}
