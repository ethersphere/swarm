// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package main

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// SimplestensABI is the input ABI used to generate the binding from.
const SimplestensABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"content\",\"outputs\":[{\"name\":\"ret\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"setContent\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"ContentChanged\",\"type\":\"event\"}]"

// SimplestensBin is the compiled bytecode used for deploying new contracts.
const SimplestensBin = `6060604052341561000f57600080fd5b6101148061001e6000396000f30060606040526004361060485763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416632dff69418114604d578063c3d014d6146072575b600080fd5b3415605757600080fd5b6060600435608a565b60405190815260200160405180910390f35b3415607c57600080fd5b6088600435602435609c565b005b60009081526020819052604090205490565b6000828152602081905260409081902082905582907f0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc9083905190815260200160405180910390a250505600a165627a7a723058200d58a9cdff1508f1bba2a044957ac49166d7e01363236a1578b52188a133ca060029`

// DeploySimplestens deploys a new Ethereum contract, binding an instance of Simplestens to it.
func DeploySimplestens(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Simplestens, error) {
	parsed, err := abi.JSON(strings.NewReader(SimplestensABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimplestensBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Simplestens{SimplestensCaller: SimplestensCaller{contract: contract}, SimplestensTransactor: SimplestensTransactor{contract: contract}}, nil
}

// Simplestens is an auto generated Go binding around an Ethereum contract.
type Simplestens struct {
	SimplestensCaller     // Read-only binding to the contract
	SimplestensTransactor // Write-only binding to the contract
}

// SimplestensCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimplestensCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimplestensTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimplestensTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimplestensSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimplestensSession struct {
	Contract     *Simplestens      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimplestensCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimplestensCallerSession struct {
	Contract *SimplestensCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// SimplestensTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimplestensTransactorSession struct {
	Contract     *SimplestensTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SimplestensRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimplestensRaw struct {
	Contract *Simplestens // Generic contract binding to access the raw methods on
}

// SimplestensCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimplestensCallerRaw struct {
	Contract *SimplestensCaller // Generic read-only contract binding to access the raw methods on
}

// SimplestensTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimplestensTransactorRaw struct {
	Contract *SimplestensTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimplestens creates a new instance of Simplestens, bound to a specific deployed contract.
func NewSimplestens(address common.Address, backend bind.ContractBackend) (*Simplestens, error) {
	contract, err := bindSimplestens(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Simplestens{SimplestensCaller: SimplestensCaller{contract: contract}, SimplestensTransactor: SimplestensTransactor{contract: contract}}, nil
}

// NewSimplestensCaller creates a new read-only instance of Simplestens, bound to a specific deployed contract.
func NewSimplestensCaller(address common.Address, caller bind.ContractCaller) (*SimplestensCaller, error) {
	contract, err := bindSimplestens(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &SimplestensCaller{contract: contract}, nil
}

// NewSimplestensTransactor creates a new write-only instance of Simplestens, bound to a specific deployed contract.
func NewSimplestensTransactor(address common.Address, transactor bind.ContractTransactor) (*SimplestensTransactor, error) {
	contract, err := bindSimplestens(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &SimplestensTransactor{contract: contract}, nil
}

// bindSimplestens binds a generic wrapper to an already deployed contract.
func bindSimplestens(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimplestensABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simplestens *SimplestensRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Simplestens.Contract.SimplestensCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simplestens *SimplestensRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simplestens.Contract.SimplestensTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simplestens *SimplestensRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simplestens.Contract.SimplestensTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Simplestens *SimplestensCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Simplestens.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Simplestens *SimplestensTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Simplestens.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Simplestens *SimplestensTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Simplestens.Contract.contract.Transact(opts, method, params...)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(node bytes32) constant returns(ret bytes32)
func (_Simplestens *SimplestensCaller) Content(opts *bind.CallOpts, node [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Simplestens.contract.Call(opts, out, "content", node)
	return *ret0, err
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(node bytes32) constant returns(ret bytes32)
func (_Simplestens *SimplestensSession) Content(node [32]byte) ([32]byte, error) {
	return _Simplestens.Contract.Content(&_Simplestens.CallOpts, node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(node bytes32) constant returns(ret bytes32)
func (_Simplestens *SimplestensCallerSession) Content(node [32]byte) ([32]byte, error) {
	return _Simplestens.Contract.Content(&_Simplestens.CallOpts, node)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(node bytes32, hash bytes32) returns()
func (_Simplestens *SimplestensTransactor) SetContent(opts *bind.TransactOpts, node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Simplestens.contract.Transact(opts, "setContent", node, hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(node bytes32, hash bytes32) returns()
func (_Simplestens *SimplestensSession) SetContent(node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Simplestens.Contract.SetContent(&_Simplestens.TransactOpts, node, hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(node bytes32, hash bytes32) returns()
func (_Simplestens *SimplestensTransactorSession) SetContent(node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Simplestens.Contract.SetContent(&_Simplestens.TransactOpts, node, hash)
}
