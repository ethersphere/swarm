// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rskresolver

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// RskresolverABI is the input ABI used to generate the binding from.
const RskresolverABI = "[{\"inputs\":[{\"name\":\"rnsAddr\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"kind\",\"type\":\"bytes32\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"interfaceID\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"addr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"addrValue\",\"type\":\"address\"}],\"name\":\"setAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"content\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"setContent\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Rskresolver is an auto generated Go binding around an Ethereum contract.
type Rskresolver struct {
	RskresolverCaller     // Read-only binding to the contract
	RskresolverTransactor // Write-only binding to the contract
	RskresolverFilterer   // Log filterer for contract events
}

// RskresolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type RskresolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RskresolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RskresolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RskresolverFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RskresolverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RskresolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RskresolverSession struct {
	Contract     *Rskresolver      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RskresolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RskresolverCallerSession struct {
	Contract *RskresolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// RskresolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RskresolverTransactorSession struct {
	Contract     *RskresolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// RskresolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type RskresolverRaw struct {
	Contract *Rskresolver // Generic contract binding to access the raw methods on
}

// RskresolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RskresolverCallerRaw struct {
	Contract *RskresolverCaller // Generic read-only contract binding to access the raw methods on
}

// RskresolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RskresolverTransactorRaw struct {
	Contract *RskresolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRskresolver creates a new instance of Rskresolver, bound to a specific deployed contract.
func NewRskresolver(address common.Address, backend bind.ContractBackend) (*Rskresolver, error) {
	contract, err := bindRskresolver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Rskresolver{RskresolverCaller: RskresolverCaller{contract: contract}, RskresolverTransactor: RskresolverTransactor{contract: contract}, RskresolverFilterer: RskresolverFilterer{contract: contract}}, nil
}

// NewRskresolverCaller creates a new read-only instance of Rskresolver, bound to a specific deployed contract.
func NewRskresolverCaller(address common.Address, caller bind.ContractCaller) (*RskresolverCaller, error) {
	contract, err := bindRskresolver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RskresolverCaller{contract: contract}, nil
}

// NewRskresolverTransactor creates a new write-only instance of Rskresolver, bound to a specific deployed contract.
func NewRskresolverTransactor(address common.Address, transactor bind.ContractTransactor) (*RskresolverTransactor, error) {
	contract, err := bindRskresolver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RskresolverTransactor{contract: contract}, nil
}

// NewRskresolverFilterer creates a new log filterer instance of Rskresolver, bound to a specific deployed contract.
func NewRskresolverFilterer(address common.Address, filterer bind.ContractFilterer) (*RskresolverFilterer, error) {
	contract, err := bindRskresolver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RskresolverFilterer{contract: contract}, nil
}

// bindRskresolver binds a generic wrapper to an already deployed contract.
func bindRskresolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RskresolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rskresolver *RskresolverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Rskresolver.Contract.RskresolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rskresolver *RskresolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rskresolver.Contract.RskresolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rskresolver *RskresolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rskresolver.Contract.RskresolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rskresolver *RskresolverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Rskresolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rskresolver *RskresolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rskresolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rskresolver *RskresolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rskresolver.Contract.contract.Transact(opts, method, params...)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Rskresolver *RskresolverCaller) Addr(opts *bind.CallOpts, node [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Rskresolver.contract.Call(opts, out, "addr", node)
	return *ret0, err
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Rskresolver *RskresolverSession) Addr(node [32]byte) (common.Address, error) {
	return _Rskresolver.Contract.Addr(&_Rskresolver.CallOpts, node)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Rskresolver *RskresolverCallerSession) Addr(node [32]byte) (common.Address, error) {
	return _Rskresolver.Contract.Addr(&_Rskresolver.CallOpts, node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Rskresolver *RskresolverCaller) Content(opts *bind.CallOpts, node [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Rskresolver.contract.Call(opts, out, "content", node)
	return *ret0, err
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Rskresolver *RskresolverSession) Content(node [32]byte) ([32]byte, error) {
	return _Rskresolver.Contract.Content(&_Rskresolver.CallOpts, node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Rskresolver *RskresolverCallerSession) Content(node [32]byte) ([32]byte, error) {
	return _Rskresolver.Contract.Content(&_Rskresolver.CallOpts, node)
}

// Has is a free data retrieval call binding the contract method 0x41b9dc2b.
//
// Solidity: function has(bytes32 node, bytes32 kind) constant returns(bool)
func (_Rskresolver *RskresolverCaller) Has(opts *bind.CallOpts, node [32]byte, kind [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Rskresolver.contract.Call(opts, out, "has", node, kind)
	return *ret0, err
}

// Has is a free data retrieval call binding the contract method 0x41b9dc2b.
//
// Solidity: function has(bytes32 node, bytes32 kind) constant returns(bool)
func (_Rskresolver *RskresolverSession) Has(node [32]byte, kind [32]byte) (bool, error) {
	return _Rskresolver.Contract.Has(&_Rskresolver.CallOpts, node, kind)
}

// Has is a free data retrieval call binding the contract method 0x41b9dc2b.
//
// Solidity: function has(bytes32 node, bytes32 kind) constant returns(bool)
func (_Rskresolver *RskresolverCallerSession) Has(node [32]byte, kind [32]byte) (bool, error) {
	return _Rskresolver.Contract.Has(&_Rskresolver.CallOpts, node, kind)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) constant returns(bool)
func (_Rskresolver *RskresolverCaller) SupportsInterface(opts *bind.CallOpts, interfaceID [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Rskresolver.contract.Call(opts, out, "supportsInterface", interfaceID)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) constant returns(bool)
func (_Rskresolver *RskresolverSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _Rskresolver.Contract.SupportsInterface(&_Rskresolver.CallOpts, interfaceID)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceID) constant returns(bool)
func (_Rskresolver *RskresolverCallerSession) SupportsInterface(interfaceID [4]byte) (bool, error) {
	return _Rskresolver.Contract.SupportsInterface(&_Rskresolver.CallOpts, interfaceID)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Rskresolver *RskresolverTransactor) SetAddr(opts *bind.TransactOpts, node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Rskresolver.contract.Transact(opts, "setAddr", node, addrValue)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Rskresolver *RskresolverSession) SetAddr(node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Rskresolver.Contract.SetAddr(&_Rskresolver.TransactOpts, node, addrValue)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Rskresolver *RskresolverTransactorSession) SetAddr(node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Rskresolver.Contract.SetAddr(&_Rskresolver.TransactOpts, node, addrValue)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 hash) returns()
func (_Rskresolver *RskresolverTransactor) SetContent(opts *bind.TransactOpts, node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Rskresolver.contract.Transact(opts, "setContent", node, hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 hash) returns()
func (_Rskresolver *RskresolverSession) SetContent(node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Rskresolver.Contract.SetContent(&_Rskresolver.TransactOpts, node, hash)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 hash) returns()
func (_Rskresolver *RskresolverTransactorSession) SetContent(node [32]byte, hash [32]byte) (*types.Transaction, error) {
	return _Rskresolver.Contract.SetContent(&_Rskresolver.TransactOpts, node, hash)
}
