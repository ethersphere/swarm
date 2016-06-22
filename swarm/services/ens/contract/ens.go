// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package contract

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ResolverABI is the input ABI used to generate the binding from.
const ResolverABI = `[{"constant":false,"inputs":[{"name":"name","type":"bytes32[]"},{"name":"rtype","type":"bytes16"},{"name":"ttl","type":"uint32"},{"name":"len","type":"uint16"},{"name":"data","type":"bytes32"}],"name":"setPrivateRR","outputs":[],"type":"function"},{"constant":true,"inputs":[{"name":"id","type":"bytes32"}],"name":"getExtended","outputs":[{"name":"data","type":"bytes"}],"type":"function"},{"constant":false,"inputs":[{"name":"name","type":"bytes32[]"}],"name":"deletePrivateRR","outputs":[],"type":"function"},{"constant":true,"inputs":[{"name":"nodeId","type":"bytes12"},{"name":"qtype","type":"bytes32"},{"name":"index","type":"uint16"}],"name":"resolve","outputs":[{"name":"rcode","type":"uint16"},{"name":"rtype","type":"bytes16"},{"name":"ttl","type":"uint32"},{"name":"len","type":"uint16"},{"name":"data","type":"bytes32"}],"type":"function"},{"constant":false,"inputs":[{"name":"name","type":"string"}],"name":"deleteRR","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"mayUpdate","outputs":[{"name":"","type":"bool"}],"type":"function"},{"constant":true,"inputs":[{"name":"nodeId","type":"bytes12"},{"name":"label","type":"bytes32"}],"name":"findResolver","outputs":[{"name":"rcode","type":"uint16"},{"name":"ttl","type":"uint32"},{"name":"rnode","type":"bytes12"},{"name":"raddress","type":"address"}],"type":"function"},{"constant":false,"inputs":[{"name":"name","type":"string"},{"name":"rtype","type":"bytes16"},{"name":"ttl","type":"uint32"},{"name":"len","type":"uint16"},{"name":"data","type":"bytes32"}],"name":"setRR","outputs":[],"type":"function"}]`

// ResolverBin is the compiled bytecode used for deploying new contracts.
const ResolverBin = `0x`

// DeployResolver deploys a new Ethereum contract, binding an instance of Resolver to it.
func DeployResolver(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Resolver, error) {
	parsed, err := abi.JSON(strings.NewReader(ResolverABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ResolverBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Resolver{ResolverCaller: ResolverCaller{contract: contract}, ResolverTransactor: ResolverTransactor{contract: contract}}, nil
}

// Resolver is an auto generated Go binding around an Ethereum contract.
type Resolver struct {
	ResolverCaller     // Read-only binding to the contract
	ResolverTransactor // Write-only binding to the contract
}

// ResolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type ResolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ResolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ResolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ResolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ResolverSession struct {
	Contract     *Resolver         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ResolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ResolverCallerSession struct {
	Contract *ResolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ResolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ResolverTransactorSession struct {
	Contract     *ResolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ResolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type ResolverRaw struct {
	Contract *Resolver // Generic contract binding to access the raw methods on
}

// ResolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ResolverCallerRaw struct {
	Contract *ResolverCaller // Generic read-only contract binding to access the raw methods on
}

// ResolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ResolverTransactorRaw struct {
	Contract *ResolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewResolver creates a new instance of Resolver, bound to a specific deployed contract.
func NewResolver(address common.Address, backend bind.ContractBackend) (*Resolver, error) {
	contract, err := bindResolver(address, backend.(bind.ContractCaller), backend.(bind.ContractTransactor))
	if err != nil {
		return nil, err
	}
	return &Resolver{ResolverCaller: ResolverCaller{contract: contract}, ResolverTransactor: ResolverTransactor{contract: contract}}, nil
}

// NewResolverCaller creates a new read-only instance of Resolver, bound to a specific deployed contract.
func NewResolverCaller(address common.Address, caller bind.ContractCaller) (*ResolverCaller, error) {
	contract, err := bindResolver(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &ResolverCaller{contract: contract}, nil
}

// NewResolverTransactor creates a new write-only instance of Resolver, bound to a specific deployed contract.
func NewResolverTransactor(address common.Address, transactor bind.ContractTransactor) (*ResolverTransactor, error) {
	contract, err := bindResolver(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &ResolverTransactor{contract: contract}, nil
}

// bindResolver binds a generic wrapper to an already deployed contract.
func bindResolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ResolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Resolver *ResolverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Resolver.Contract.ResolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Resolver *ResolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Resolver.Contract.ResolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Resolver *ResolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Resolver.Contract.ResolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Resolver *ResolverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Resolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Resolver *ResolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Resolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Resolver *ResolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Resolver.Contract.contract.Transact(opts, method, params...)
}

// FindResolver is a free data retrieval call binding the contract method 0xedc0277c.
//
// Solidity: function findResolver(nodeId bytes12, label bytes32) constant returns(rcode uint16, ttl uint32, rnode bytes12, raddress address)
func (_Resolver *ResolverCaller) FindResolver(opts *bind.CallOpts, nodeId [12]byte, label [32]byte) (struct {
	Rcode    uint16
	Ttl      uint32
	Rnode    [12]byte
	Raddress common.Address
}, error) {
	ret := new(struct {
		Rcode    uint16
		Ttl      uint32
		Rnode    [12]byte
		Raddress common.Address
	})
	out := ret
	err := _Resolver.contract.Call(opts, out, "findResolver", nodeId, label)
	return *ret, err
}

// FindResolver is a free data retrieval call binding the contract method 0xedc0277c.
//
// Solidity: function findResolver(nodeId bytes12, label bytes32) constant returns(rcode uint16, ttl uint32, rnode bytes12, raddress address)
func (_Resolver *ResolverSession) FindResolver(nodeId [12]byte, label [32]byte) (struct {
	Rcode    uint16
	Ttl      uint32
	Rnode    [12]byte
	Raddress common.Address
}, error) {
	return _Resolver.Contract.FindResolver(&_Resolver.CallOpts, nodeId, label)
}

// FindResolver is a free data retrieval call binding the contract method 0xedc0277c.
//
// Solidity: function findResolver(nodeId bytes12, label bytes32) constant returns(rcode uint16, ttl uint32, rnode bytes12, raddress address)
func (_Resolver *ResolverCallerSession) FindResolver(nodeId [12]byte, label [32]byte) (struct {
	Rcode    uint16
	Ttl      uint32
	Rnode    [12]byte
	Raddress common.Address
}, error) {
	return _Resolver.Contract.FindResolver(&_Resolver.CallOpts, nodeId, label)
}

// GetExtended is a free data retrieval call binding the contract method 0x8021061c.
//
// Solidity: function getExtended(id bytes32) constant returns(data bytes)
func (_Resolver *ResolverCaller) GetExtended(opts *bind.CallOpts, id [32]byte) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _Resolver.contract.Call(opts, out, "getExtended", id)
	return *ret0, err
}

// GetExtended is a free data retrieval call binding the contract method 0x8021061c.
//
// Solidity: function getExtended(id bytes32) constant returns(data bytes)
func (_Resolver *ResolverSession) GetExtended(id [32]byte) ([]byte, error) {
	return _Resolver.Contract.GetExtended(&_Resolver.CallOpts, id)
}

// GetExtended is a free data retrieval call binding the contract method 0x8021061c.
//
// Solidity: function getExtended(id bytes32) constant returns(data bytes)
func (_Resolver *ResolverCallerSession) GetExtended(id [32]byte) ([]byte, error) {
	return _Resolver.Contract.GetExtended(&_Resolver.CallOpts, id)
}

// MayUpdate is a free data retrieval call binding the contract method 0xe8c70084.
//
// Solidity: function mayUpdate() constant returns(bool)
func (_Resolver *ResolverCaller) MayUpdate(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Resolver.contract.Call(opts, out, "mayUpdate")
	return *ret0, err
}

// MayUpdate is a free data retrieval call binding the contract method 0xe8c70084.
//
// Solidity: function mayUpdate() constant returns(bool)
func (_Resolver *ResolverSession) MayUpdate() (bool, error) {
	return _Resolver.Contract.MayUpdate(&_Resolver.CallOpts)
}

// MayUpdate is a free data retrieval call binding the contract method 0xe8c70084.
//
// Solidity: function mayUpdate() constant returns(bool)
func (_Resolver *ResolverCallerSession) MayUpdate() (bool, error) {
	return _Resolver.Contract.MayUpdate(&_Resolver.CallOpts)
}

// Resolve is a free data retrieval call binding the contract method 0xa16fdafa.
//
// Solidity: function resolve(nodeId bytes12, qtype bytes32, index uint16) constant returns(rcode uint16, rtype bytes16, ttl uint32, len uint16, data bytes32)
func (_Resolver *ResolverCaller) Resolve(opts *bind.CallOpts, nodeId [12]byte, qtype [32]byte, index uint16) (struct {
	Rcode uint16
	Rtype [16]byte
	Ttl   uint32
	Len   uint16
	Data  [32]byte
}, error) {
	ret := new(struct {
		Rcode uint16
		Rtype [16]byte
		Ttl   uint32
		Len   uint16
		Data  [32]byte
	})
	out := ret
	err := _Resolver.contract.Call(opts, out, "resolve", nodeId, qtype, index)
	return *ret, err
}

// Resolve is a free data retrieval call binding the contract method 0xa16fdafa.
//
// Solidity: function resolve(nodeId bytes12, qtype bytes32, index uint16) constant returns(rcode uint16, rtype bytes16, ttl uint32, len uint16, data bytes32)
func (_Resolver *ResolverSession) Resolve(nodeId [12]byte, qtype [32]byte, index uint16) (struct {
	Rcode uint16
	Rtype [16]byte
	Ttl   uint32
	Len   uint16
	Data  [32]byte
}, error) {
	return _Resolver.Contract.Resolve(&_Resolver.CallOpts, nodeId, qtype, index)
}

// Resolve is a free data retrieval call binding the contract method 0xa16fdafa.
//
// Solidity: function resolve(nodeId bytes12, qtype bytes32, index uint16) constant returns(rcode uint16, rtype bytes16, ttl uint32, len uint16, data bytes32)
func (_Resolver *ResolverCallerSession) Resolve(nodeId [12]byte, qtype [32]byte, index uint16) (struct {
	Rcode uint16
	Rtype [16]byte
	Ttl   uint32
	Len   uint16
	Data  [32]byte
}, error) {
	return _Resolver.Contract.Resolve(&_Resolver.CallOpts, nodeId, qtype, index)
}

// DeletePrivateRR is a paid mutator transaction binding the contract method 0x89c0d9ef.
//
// Solidity: function deletePrivateRR(name bytes32[]) returns()
func (_Resolver *ResolverTransactor) DeletePrivateRR(opts *bind.TransactOpts, name [][32]byte) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "deletePrivateRR", name)
}

// DeletePrivateRR is a paid mutator transaction binding the contract method 0x89c0d9ef.
//
// Solidity: function deletePrivateRR(name bytes32[]) returns()
func (_Resolver *ResolverSession) DeletePrivateRR(name [][32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.DeletePrivateRR(&_Resolver.TransactOpts, name)
}

// DeletePrivateRR is a paid mutator transaction binding the contract method 0x89c0d9ef.
//
// Solidity: function deletePrivateRR(name bytes32[]) returns()
func (_Resolver *ResolverTransactorSession) DeletePrivateRR(name [][32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.DeletePrivateRR(&_Resolver.TransactOpts, name)
}

// DeleteRR is a paid mutator transaction binding the contract method 0xaceafcc4.
//
// Solidity: function deleteRR(name string) returns()
func (_Resolver *ResolverTransactor) DeleteRR(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "deleteRR", name)
}

// DeleteRR is a paid mutator transaction binding the contract method 0xaceafcc4.
//
// Solidity: function deleteRR(name string) returns()
func (_Resolver *ResolverSession) DeleteRR(name string) (*types.Transaction, error) {
	return _Resolver.Contract.DeleteRR(&_Resolver.TransactOpts, name)
}

// DeleteRR is a paid mutator transaction binding the contract method 0xaceafcc4.
//
// Solidity: function deleteRR(name string) returns()
func (_Resolver *ResolverTransactorSession) DeleteRR(name string) (*types.Transaction, error) {
	return _Resolver.Contract.DeleteRR(&_Resolver.TransactOpts, name)
}

// SetPrivateRR is a paid mutator transaction binding the contract method 0x60ae74ae.
//
// Solidity: function setPrivateRR(name bytes32[], rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverTransactor) SetPrivateRR(opts *bind.TransactOpts, name [][32]byte, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "setPrivateRR", name, rtype, ttl, len, data)
}

// SetPrivateRR is a paid mutator transaction binding the contract method 0x60ae74ae.
//
// Solidity: function setPrivateRR(name bytes32[], rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverSession) SetPrivateRR(name [][32]byte, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.SetPrivateRR(&_Resolver.TransactOpts, name, rtype, ttl, len, data)
}

// SetPrivateRR is a paid mutator transaction binding the contract method 0x60ae74ae.
//
// Solidity: function setPrivateRR(name bytes32[], rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverTransactorSession) SetPrivateRR(name [][32]byte, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.SetPrivateRR(&_Resolver.TransactOpts, name, rtype, ttl, len, data)
}

// SetRR is a paid mutator transaction binding the contract method 0xef32ac57.
//
// Solidity: function setRR(name string, rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverTransactor) SetRR(opts *bind.TransactOpts, name string, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "setRR", name, rtype, ttl, len, data)
}

// SetRR is a paid mutator transaction binding the contract method 0xef32ac57.
//
// Solidity: function setRR(name string, rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverSession) SetRR(name string, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.SetRR(&_Resolver.TransactOpts, name, rtype, ttl, len, data)
}

// SetRR is a paid mutator transaction binding the contract method 0xef32ac57.
//
// Solidity: function setRR(name string, rtype bytes16, ttl uint32, len uint16, data bytes32) returns()
func (_Resolver *ResolverTransactorSession) SetRR(name string, rtype [16]byte, ttl uint32, len uint16, data [32]byte) (*types.Transaction, error) {
	return _Resolver.Contract.SetRR(&_Resolver.TransactOpts, name, rtype, ttl, len, data)
}
