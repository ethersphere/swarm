// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package multichainresolver

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

// MultichainresolverABI is the input ABI used to generate the binding from.
const MultichainresolverABI = "[{\"inputs\":[{\"name\":\"_rns\",\"type\":\"address\"},{\"name\":\"_publicResolver\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"content\",\"type\":\"bytes32\"}],\"name\":\"ContentChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"chain\",\"type\":\"bytes4\"},{\"indexed\":false,\"name\":\"metadata\",\"type\":\"bytes32\"}],\"name\":\"ChainMetadataChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"chain\",\"type\":\"bytes4\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"string\"}],\"name\":\"ChainAddrChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"node\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AddrChanged\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"addr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"addrValue\",\"type\":\"address\"}],\"name\":\"setAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"}],\"name\":\"content\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"contentValue\",\"type\":\"bytes32\"}],\"name\":\"setContent\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"}],\"name\":\"chainAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"},{\"name\":\"addrValue\",\"type\":\"string\"}],\"name\":\"setChainAddr\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"}],\"name\":\"chainMetadata\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"},{\"name\":\"metadataValue\",\"type\":\"bytes32\"}],\"name\":\"setChainMetadata\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"}],\"name\":\"chainAddrAndMetadata\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"node\",\"type\":\"bytes32\"},{\"name\":\"chain\",\"type\":\"bytes4\"},{\"name\":\"addrValue\",\"type\":\"string\"},{\"name\":\"metadataValue\",\"type\":\"bytes32\"}],\"name\":\"setChainAddrWithMetadata\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Multichainresolver is an auto generated Go binding around an Ethereum contract.
type Multichainresolver struct {
	MultichainresolverCaller     // Read-only binding to the contract
	MultichainresolverTransactor // Write-only binding to the contract
	MultichainresolverFilterer   // Log filterer for contract events
}

// MultichainresolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultichainresolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultichainresolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultichainresolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultichainresolverFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultichainresolverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultichainresolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultichainresolverSession struct {
	Contract     *Multichainresolver // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// MultichainresolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultichainresolverCallerSession struct {
	Contract *MultichainresolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// MultichainresolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultichainresolverTransactorSession struct {
	Contract     *MultichainresolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// MultichainresolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultichainresolverRaw struct {
	Contract *Multichainresolver // Generic contract binding to access the raw methods on
}

// MultichainresolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultichainresolverCallerRaw struct {
	Contract *MultichainresolverCaller // Generic read-only contract binding to access the raw methods on
}

// MultichainresolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultichainresolverTransactorRaw struct {
	Contract *MultichainresolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultichainresolver creates a new instance of Multichainresolver, bound to a specific deployed contract.
func NewMultichainresolver(address common.Address, backend bind.ContractBackend) (*Multichainresolver, error) {
	contract, err := bindMultichainresolver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Multichainresolver{MultichainresolverCaller: MultichainresolverCaller{contract: contract}, MultichainresolverTransactor: MultichainresolverTransactor{contract: contract}, MultichainresolverFilterer: MultichainresolverFilterer{contract: contract}}, nil
}

// NewMultichainresolverCaller creates a new read-only instance of Multichainresolver, bound to a specific deployed contract.
func NewMultichainresolverCaller(address common.Address, caller bind.ContractCaller) (*MultichainresolverCaller, error) {
	contract, err := bindMultichainresolver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultichainresolverCaller{contract: contract}, nil
}

// NewMultichainresolverTransactor creates a new write-only instance of Multichainresolver, bound to a specific deployed contract.
func NewMultichainresolverTransactor(address common.Address, transactor bind.ContractTransactor) (*MultichainresolverTransactor, error) {
	contract, err := bindMultichainresolver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultichainresolverTransactor{contract: contract}, nil
}

// NewMultichainresolverFilterer creates a new log filterer instance of Multichainresolver, bound to a specific deployed contract.
func NewMultichainresolverFilterer(address common.Address, filterer bind.ContractFilterer) (*MultichainresolverFilterer, error) {
	contract, err := bindMultichainresolver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultichainresolverFilterer{contract: contract}, nil
}

// bindMultichainresolver binds a generic wrapper to an already deployed contract.
func bindMultichainresolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultichainresolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Multichainresolver *MultichainresolverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Multichainresolver.Contract.MultichainresolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Multichainresolver *MultichainresolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Multichainresolver.Contract.MultichainresolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Multichainresolver *MultichainresolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Multichainresolver.Contract.MultichainresolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Multichainresolver *MultichainresolverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Multichainresolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Multichainresolver *MultichainresolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Multichainresolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Multichainresolver *MultichainresolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Multichainresolver.Contract.contract.Transact(opts, method, params...)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Multichainresolver *MultichainresolverCaller) Addr(opts *bind.CallOpts, node [32]byte) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Multichainresolver.contract.Call(opts, out, "addr", node)
	return *ret0, err
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Multichainresolver *MultichainresolverSession) Addr(node [32]byte) (common.Address, error) {
	return _Multichainresolver.Contract.Addr(&_Multichainresolver.CallOpts, node)
}

// Addr is a free data retrieval call binding the contract method 0x3b3b57de.
//
// Solidity: function addr(bytes32 node) constant returns(address)
func (_Multichainresolver *MultichainresolverCallerSession) Addr(node [32]byte) (common.Address, error) {
	return _Multichainresolver.Contract.Addr(&_Multichainresolver.CallOpts, node)
}

// ChainAddr is a free data retrieval call binding the contract method 0x8be4b5f6.
//
// Solidity: function chainAddr(bytes32 node, bytes4 chain) constant returns(string)
func (_Multichainresolver *MultichainresolverCaller) ChainAddr(opts *bind.CallOpts, node [32]byte, chain [4]byte) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Multichainresolver.contract.Call(opts, out, "chainAddr", node, chain)
	return *ret0, err
}

// ChainAddr is a free data retrieval call binding the contract method 0x8be4b5f6.
//
// Solidity: function chainAddr(bytes32 node, bytes4 chain) constant returns(string)
func (_Multichainresolver *MultichainresolverSession) ChainAddr(node [32]byte, chain [4]byte) (string, error) {
	return _Multichainresolver.Contract.ChainAddr(&_Multichainresolver.CallOpts, node, chain)
}

// ChainAddr is a free data retrieval call binding the contract method 0x8be4b5f6.
//
// Solidity: function chainAddr(bytes32 node, bytes4 chain) constant returns(string)
func (_Multichainresolver *MultichainresolverCallerSession) ChainAddr(node [32]byte, chain [4]byte) (string, error) {
	return _Multichainresolver.Contract.ChainAddr(&_Multichainresolver.CallOpts, node, chain)
}

// ChainAddrAndMetadata is a free data retrieval call binding the contract method 0x82e3bee6.
//
// Solidity: function chainAddrAndMetadata(bytes32 node, bytes4 chain) constant returns(string, bytes32)
func (_Multichainresolver *MultichainresolverCaller) ChainAddrAndMetadata(opts *bind.CallOpts, node [32]byte, chain [4]byte) (string, [32]byte, error) {
	var (
		ret0 = new(string)
		ret1 = new([32]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Multichainresolver.contract.Call(opts, out, "chainAddrAndMetadata", node, chain)
	return *ret0, *ret1, err
}

// ChainAddrAndMetadata is a free data retrieval call binding the contract method 0x82e3bee6.
//
// Solidity: function chainAddrAndMetadata(bytes32 node, bytes4 chain) constant returns(string, bytes32)
func (_Multichainresolver *MultichainresolverSession) ChainAddrAndMetadata(node [32]byte, chain [4]byte) (string, [32]byte, error) {
	return _Multichainresolver.Contract.ChainAddrAndMetadata(&_Multichainresolver.CallOpts, node, chain)
}

// ChainAddrAndMetadata is a free data retrieval call binding the contract method 0x82e3bee6.
//
// Solidity: function chainAddrAndMetadata(bytes32 node, bytes4 chain) constant returns(string, bytes32)
func (_Multichainresolver *MultichainresolverCallerSession) ChainAddrAndMetadata(node [32]byte, chain [4]byte) (string, [32]byte, error) {
	return _Multichainresolver.Contract.ChainAddrAndMetadata(&_Multichainresolver.CallOpts, node, chain)
}

// ChainMetadata is a free data retrieval call binding the contract method 0xb34e8cd6.
//
// Solidity: function chainMetadata(bytes32 node, bytes4 chain) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverCaller) ChainMetadata(opts *bind.CallOpts, node [32]byte, chain [4]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Multichainresolver.contract.Call(opts, out, "chainMetadata", node, chain)
	return *ret0, err
}

// ChainMetadata is a free data retrieval call binding the contract method 0xb34e8cd6.
//
// Solidity: function chainMetadata(bytes32 node, bytes4 chain) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverSession) ChainMetadata(node [32]byte, chain [4]byte) ([32]byte, error) {
	return _Multichainresolver.Contract.ChainMetadata(&_Multichainresolver.CallOpts, node, chain)
}

// ChainMetadata is a free data retrieval call binding the contract method 0xb34e8cd6.
//
// Solidity: function chainMetadata(bytes32 node, bytes4 chain) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverCallerSession) ChainMetadata(node [32]byte, chain [4]byte) ([32]byte, error) {
	return _Multichainresolver.Contract.ChainMetadata(&_Multichainresolver.CallOpts, node, chain)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverCaller) Content(opts *bind.CallOpts, node [32]byte) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Multichainresolver.contract.Call(opts, out, "content", node)
	return *ret0, err
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverSession) Content(node [32]byte) ([32]byte, error) {
	return _Multichainresolver.Contract.Content(&_Multichainresolver.CallOpts, node)
}

// Content is a free data retrieval call binding the contract method 0x2dff6941.
//
// Solidity: function content(bytes32 node) constant returns(bytes32)
func (_Multichainresolver *MultichainresolverCallerSession) Content(node [32]byte) ([32]byte, error) {
	return _Multichainresolver.Contract.Content(&_Multichainresolver.CallOpts, node)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_Multichainresolver *MultichainresolverCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Multichainresolver.contract.Call(opts, out, "supportsInterface", interfaceId)
	return *ret0, err
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_Multichainresolver *MultichainresolverSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Multichainresolver.Contract.SupportsInterface(&_Multichainresolver.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) constant returns(bool)
func (_Multichainresolver *MultichainresolverCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Multichainresolver.Contract.SupportsInterface(&_Multichainresolver.CallOpts, interfaceId)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Multichainresolver *MultichainresolverTransactor) SetAddr(opts *bind.TransactOpts, node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Multichainresolver.contract.Transact(opts, "setAddr", node, addrValue)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Multichainresolver *MultichainresolverSession) SetAddr(node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetAddr(&_Multichainresolver.TransactOpts, node, addrValue)
}

// SetAddr is a paid mutator transaction binding the contract method 0xd5fa2b00.
//
// Solidity: function setAddr(bytes32 node, address addrValue) returns()
func (_Multichainresolver *MultichainresolverTransactorSession) SetAddr(node [32]byte, addrValue common.Address) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetAddr(&_Multichainresolver.TransactOpts, node, addrValue)
}

// SetChainAddr is a paid mutator transaction binding the contract method 0xd278b400.
//
// Solidity: function setChainAddr(bytes32 node, bytes4 chain, string addrValue) returns()
func (_Multichainresolver *MultichainresolverTransactor) SetChainAddr(opts *bind.TransactOpts, node [32]byte, chain [4]byte, addrValue string) (*types.Transaction, error) {
	return _Multichainresolver.contract.Transact(opts, "setChainAddr", node, chain, addrValue)
}

// SetChainAddr is a paid mutator transaction binding the contract method 0xd278b400.
//
// Solidity: function setChainAddr(bytes32 node, bytes4 chain, string addrValue) returns()
func (_Multichainresolver *MultichainresolverSession) SetChainAddr(node [32]byte, chain [4]byte, addrValue string) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainAddr(&_Multichainresolver.TransactOpts, node, chain, addrValue)
}

// SetChainAddr is a paid mutator transaction binding the contract method 0xd278b400.
//
// Solidity: function setChainAddr(bytes32 node, bytes4 chain, string addrValue) returns()
func (_Multichainresolver *MultichainresolverTransactorSession) SetChainAddr(node [32]byte, chain [4]byte, addrValue string) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainAddr(&_Multichainresolver.TransactOpts, node, chain, addrValue)
}

// SetChainAddrWithMetadata is a paid mutator transaction binding the contract method 0xe335bee4.
//
// Solidity: function setChainAddrWithMetadata(bytes32 node, bytes4 chain, string addrValue, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverTransactor) SetChainAddrWithMetadata(opts *bind.TransactOpts, node [32]byte, chain [4]byte, addrValue string, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.contract.Transact(opts, "setChainAddrWithMetadata", node, chain, addrValue, metadataValue)
}

// SetChainAddrWithMetadata is a paid mutator transaction binding the contract method 0xe335bee4.
//
// Solidity: function setChainAddrWithMetadata(bytes32 node, bytes4 chain, string addrValue, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverSession) SetChainAddrWithMetadata(node [32]byte, chain [4]byte, addrValue string, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainAddrWithMetadata(&_Multichainresolver.TransactOpts, node, chain, addrValue, metadataValue)
}

// SetChainAddrWithMetadata is a paid mutator transaction binding the contract method 0xe335bee4.
//
// Solidity: function setChainAddrWithMetadata(bytes32 node, bytes4 chain, string addrValue, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverTransactorSession) SetChainAddrWithMetadata(node [32]byte, chain [4]byte, addrValue string, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainAddrWithMetadata(&_Multichainresolver.TransactOpts, node, chain, addrValue, metadataValue)
}

// SetChainMetadata is a paid mutator transaction binding the contract method 0x245d4d9a.
//
// Solidity: function setChainMetadata(bytes32 node, bytes4 chain, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverTransactor) SetChainMetadata(opts *bind.TransactOpts, node [32]byte, chain [4]byte, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.contract.Transact(opts, "setChainMetadata", node, chain, metadataValue)
}

// SetChainMetadata is a paid mutator transaction binding the contract method 0x245d4d9a.
//
// Solidity: function setChainMetadata(bytes32 node, bytes4 chain, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverSession) SetChainMetadata(node [32]byte, chain [4]byte, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainMetadata(&_Multichainresolver.TransactOpts, node, chain, metadataValue)
}

// SetChainMetadata is a paid mutator transaction binding the contract method 0x245d4d9a.
//
// Solidity: function setChainMetadata(bytes32 node, bytes4 chain, bytes32 metadataValue) returns()
func (_Multichainresolver *MultichainresolverTransactorSession) SetChainMetadata(node [32]byte, chain [4]byte, metadataValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetChainMetadata(&_Multichainresolver.TransactOpts, node, chain, metadataValue)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 contentValue) returns()
func (_Multichainresolver *MultichainresolverTransactor) SetContent(opts *bind.TransactOpts, node [32]byte, contentValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.contract.Transact(opts, "setContent", node, contentValue)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 contentValue) returns()
func (_Multichainresolver *MultichainresolverSession) SetContent(node [32]byte, contentValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetContent(&_Multichainresolver.TransactOpts, node, contentValue)
}

// SetContent is a paid mutator transaction binding the contract method 0xc3d014d6.
//
// Solidity: function setContent(bytes32 node, bytes32 contentValue) returns()
func (_Multichainresolver *MultichainresolverTransactorSession) SetContent(node [32]byte, contentValue [32]byte) (*types.Transaction, error) {
	return _Multichainresolver.Contract.SetContent(&_Multichainresolver.TransactOpts, node, contentValue)
}

// MultichainresolverAddrChangedIterator is returned from FilterAddrChanged and is used to iterate over the raw logs and unpacked data for AddrChanged events raised by the Multichainresolver contract.
type MultichainresolverAddrChangedIterator struct {
	Event *MultichainresolverAddrChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MultichainresolverAddrChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultichainresolverAddrChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MultichainresolverAddrChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MultichainresolverAddrChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultichainresolverAddrChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultichainresolverAddrChanged represents a AddrChanged event raised by the Multichainresolver contract.
type MultichainresolverAddrChanged struct {
	Node [32]byte
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAddrChanged is a free log retrieval operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address addr)
func (_Multichainresolver *MultichainresolverFilterer) FilterAddrChanged(opts *bind.FilterOpts, node [][32]byte) (*MultichainresolverAddrChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Multichainresolver.contract.FilterLogs(opts, "AddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &MultichainresolverAddrChangedIterator{contract: _Multichainresolver.contract, event: "AddrChanged", logs: logs, sub: sub}, nil
}

// WatchAddrChanged is a free log subscription operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address addr)
func (_Multichainresolver *MultichainresolverFilterer) WatchAddrChanged(opts *bind.WatchOpts, sink chan<- *MultichainresolverAddrChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Multichainresolver.contract.WatchLogs(opts, "AddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultichainresolverAddrChanged)
				if err := _Multichainresolver.contract.UnpackLog(event, "AddrChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddrChanged is a log parse operation binding the contract event 0x52d7d861f09ab3d26239d492e8968629f95e9e318cf0b73bfddc441522a15fd2.
//
// Solidity: event AddrChanged(bytes32 indexed node, address addr)
func (_Multichainresolver *MultichainresolverFilterer) ParseAddrChanged(log types.Log) (*MultichainresolverAddrChanged, error) {
	event := new(MultichainresolverAddrChanged)
	if err := _Multichainresolver.contract.UnpackLog(event, "AddrChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// MultichainresolverChainAddrChangedIterator is returned from FilterChainAddrChanged and is used to iterate over the raw logs and unpacked data for ChainAddrChanged events raised by the Multichainresolver contract.
type MultichainresolverChainAddrChangedIterator struct {
	Event *MultichainresolverChainAddrChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MultichainresolverChainAddrChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultichainresolverChainAddrChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MultichainresolverChainAddrChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MultichainresolverChainAddrChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultichainresolverChainAddrChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultichainresolverChainAddrChanged represents a ChainAddrChanged event raised by the Multichainresolver contract.
type MultichainresolverChainAddrChanged struct {
	Node  [32]byte
	Chain [4]byte
	Addr  string
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterChainAddrChanged is a free log retrieval operation binding the contract event 0x6a3e28813f2e2e5bcd0436779f8c5cb179ceadf0379291a818b9078e772b178d.
//
// Solidity: event ChainAddrChanged(bytes32 indexed node, bytes4 chain, string addr)
func (_Multichainresolver *MultichainresolverFilterer) FilterChainAddrChanged(opts *bind.FilterOpts, node [][32]byte) (*MultichainresolverChainAddrChangedIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Multichainresolver.contract.FilterLogs(opts, "ChainAddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return &MultichainresolverChainAddrChangedIterator{contract: _Multichainresolver.contract, event: "ChainAddrChanged", logs: logs, sub: sub}, nil
}

// WatchChainAddrChanged is a free log subscription operation binding the contract event 0x6a3e28813f2e2e5bcd0436779f8c5cb179ceadf0379291a818b9078e772b178d.
//
// Solidity: event ChainAddrChanged(bytes32 indexed node, bytes4 chain, string addr)
func (_Multichainresolver *MultichainresolverFilterer) WatchChainAddrChanged(opts *bind.WatchOpts, sink chan<- *MultichainresolverChainAddrChanged, node [][32]byte) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _Multichainresolver.contract.WatchLogs(opts, "ChainAddrChanged", nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultichainresolverChainAddrChanged)
				if err := _Multichainresolver.contract.UnpackLog(event, "ChainAddrChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChainAddrChanged is a log parse operation binding the contract event 0x6a3e28813f2e2e5bcd0436779f8c5cb179ceadf0379291a818b9078e772b178d.
//
// Solidity: event ChainAddrChanged(bytes32 indexed node, bytes4 chain, string addr)
func (_Multichainresolver *MultichainresolverFilterer) ParseChainAddrChanged(log types.Log) (*MultichainresolverChainAddrChanged, error) {
	event := new(MultichainresolverChainAddrChanged)
	if err := _Multichainresolver.contract.UnpackLog(event, "ChainAddrChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// MultichainresolverChainMetadataChangedIterator is returned from FilterChainMetadataChanged and is used to iterate over the raw logs and unpacked data for ChainMetadataChanged events raised by the Multichainresolver contract.
type MultichainresolverChainMetadataChangedIterator struct {
	Event *MultichainresolverChainMetadataChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MultichainresolverChainMetadataChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultichainresolverChainMetadataChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MultichainresolverChainMetadataChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MultichainresolverChainMetadataChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultichainresolverChainMetadataChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultichainresolverChainMetadataChanged represents a ChainMetadataChanged event raised by the Multichainresolver contract.
type MultichainresolverChainMetadataChanged struct {
	Node     [32]byte
	Chain    [4]byte
	Metadata [32]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterChainMetadataChanged is a free log retrieval operation binding the contract event 0x92c52f77ad49286096555eb922ca7a09249e8dd525cf58cd162fb1165686fad4.
//
// Solidity: event ChainMetadataChanged(bytes32 node, bytes4 chain, bytes32 metadata)
func (_Multichainresolver *MultichainresolverFilterer) FilterChainMetadataChanged(opts *bind.FilterOpts) (*MultichainresolverChainMetadataChangedIterator, error) {

	logs, sub, err := _Multichainresolver.contract.FilterLogs(opts, "ChainMetadataChanged")
	if err != nil {
		return nil, err
	}
	return &MultichainresolverChainMetadataChangedIterator{contract: _Multichainresolver.contract, event: "ChainMetadataChanged", logs: logs, sub: sub}, nil
}

// WatchChainMetadataChanged is a free log subscription operation binding the contract event 0x92c52f77ad49286096555eb922ca7a09249e8dd525cf58cd162fb1165686fad4.
//
// Solidity: event ChainMetadataChanged(bytes32 node, bytes4 chain, bytes32 metadata)
func (_Multichainresolver *MultichainresolverFilterer) WatchChainMetadataChanged(opts *bind.WatchOpts, sink chan<- *MultichainresolverChainMetadataChanged) (event.Subscription, error) {

	logs, sub, err := _Multichainresolver.contract.WatchLogs(opts, "ChainMetadataChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultichainresolverChainMetadataChanged)
				if err := _Multichainresolver.contract.UnpackLog(event, "ChainMetadataChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChainMetadataChanged is a log parse operation binding the contract event 0x92c52f77ad49286096555eb922ca7a09249e8dd525cf58cd162fb1165686fad4.
//
// Solidity: event ChainMetadataChanged(bytes32 node, bytes4 chain, bytes32 metadata)
func (_Multichainresolver *MultichainresolverFilterer) ParseChainMetadataChanged(log types.Log) (*MultichainresolverChainMetadataChanged, error) {
	event := new(MultichainresolverChainMetadataChanged)
	if err := _Multichainresolver.contract.UnpackLog(event, "ChainMetadataChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// MultichainresolverContentChangedIterator is returned from FilterContentChanged and is used to iterate over the raw logs and unpacked data for ContentChanged events raised by the Multichainresolver contract.
type MultichainresolverContentChangedIterator struct {
	Event *MultichainresolverContentChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MultichainresolverContentChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MultichainresolverContentChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MultichainresolverContentChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MultichainresolverContentChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MultichainresolverContentChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MultichainresolverContentChanged represents a ContentChanged event raised by the Multichainresolver contract.
type MultichainresolverContentChanged struct {
	Node    [32]byte
	Content [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterContentChanged is a free log retrieval operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 node, bytes32 content)
func (_Multichainresolver *MultichainresolverFilterer) FilterContentChanged(opts *bind.FilterOpts) (*MultichainresolverContentChangedIterator, error) {

	logs, sub, err := _Multichainresolver.contract.FilterLogs(opts, "ContentChanged")
	if err != nil {
		return nil, err
	}
	return &MultichainresolverContentChangedIterator{contract: _Multichainresolver.contract, event: "ContentChanged", logs: logs, sub: sub}, nil
}

// WatchContentChanged is a free log subscription operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 node, bytes32 content)
func (_Multichainresolver *MultichainresolverFilterer) WatchContentChanged(opts *bind.WatchOpts, sink chan<- *MultichainresolverContentChanged) (event.Subscription, error) {

	logs, sub, err := _Multichainresolver.contract.WatchLogs(opts, "ContentChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MultichainresolverContentChanged)
				if err := _Multichainresolver.contract.UnpackLog(event, "ContentChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseContentChanged is a log parse operation binding the contract event 0x0424b6fe0d9c3bdbece0e7879dc241bb0c22e900be8b6c168b4ee08bd9bf83bc.
//
// Solidity: event ContentChanged(bytes32 node, bytes32 content)
func (_Multichainresolver *MultichainresolverFilterer) ParseContentChanged(log types.Log) (*MultichainresolverContentChanged, error) {
	event := new(MultichainresolverContentChanged)
	if err := _Multichainresolver.contract.UnpackLog(event, "ContentChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}
