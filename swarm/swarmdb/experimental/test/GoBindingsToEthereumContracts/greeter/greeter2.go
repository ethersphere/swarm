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

// GreeterABI is the input ABI used to generate the binding from.
const GreeterABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"kill\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"greet\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_greeting\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// GreeterBin is the compiled bytecode used for deploying new contracts.
const GreeterBin = `6060604052604051610ca2380380610ca2833981016040528080519060200190919080518201919060200180519060200190919080518201919060200150505b83600360005060003373ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050819055508260006000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100c257805160ff19168380011785556100f3565b828001600101855582156100f3579182015b828111156100f25782518260005055916020019190600101906100d4565b5b50905061011e9190610100565b8082111561011a5760008181506000905550600101610100565b5090565b50508060016000509080519060200190828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061016f57805160ff19168380011785556101a0565b828001600101855582156101a0579182015b8281111561019f578251826000505591602001919060010190610181565b5b5090506101cb91906101ad565b808211156101c757600081815060009055506001016101ad565b5090565b505081600260006101000a81548160ff021916908302179055505b50505050610aaa806101f86000396000f360606040523615610095576000357c01000000000000000000000000000000000000000000000000000000009004806306fdde03146100a257806323b872dd1461011d578063313ce5671461015b57806370a082311461018157806395d89b41146101ad578063a9059cbb14610228578063cae9ca5114610249578063dc3080f2146102c5578063dd62ed3e146102fa57610095565b6100a05b610002565b565b005b6100af600480505061032f565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f16801561010f5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b61014560048080359060200190919080359060200190919080359060200190919050506107f3565b6040518082815260200191505060405180910390f35b6101686004805050610471565b604051808260ff16815260200191505060405180910390f35b6101976004808035906020019091905050610484565b6040518082815260200191505060405180910390f35b6101ba60048050506103d0565b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f16801561021a5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b61024760048080359060200190919080359060200190919050506104f5565b005b6102af6004808035906020019091908035906020019091908035906020019082018035906020019191908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050909091905050610680565b6040518082815260200191505060405180910390f35b6102e460048080359060200190919080359060200190919050506104ca565b6040518082815260200191505060405180910390f35b610319600480803590602001909190803590602001909190505061049f565b6040518082815260200191505060405180910390f35b60006000508054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156103c85780601f1061039d576101008083540402835291602001916103c8565b820191906000526020600020905b8154815290600101906020018083116103ab57829003601f168201915b505050505081565b60016000508054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156104695780601f1061043e57610100808354040283529160200191610469565b820191906000526020600020905b81548152906001019060200180831161044c57829003601f168201915b505050505081565b600260009054906101000a900460ff1681565b60036000506020528060005260406000206000915090505481565b6004600050602052816000526040600020600050602052806000526040600020600091509150505481565b6005600050602052816000526040600020600050602052806000526040600020600091509150505481565b80600360005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054101561053157610002565b600360005060008373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505481600360005060008573ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505401101561059e57610002565b80600360005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555080600360005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055508173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a35b5050565b6000600083600460005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060008773ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600050819055508490508073ffffffffffffffffffffffffffffffffffffffff16638f4ffcb133863087604051857c0100000000000000000000000000000000000000000000000000000000028152600401808573ffffffffffffffffffffffffffffffffffffffff1681526020018481526020018373ffffffffffffffffffffffffffffffffffffffff168152602001806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600f02600301f150905090810190601f1680156107c85780820380516001836020036101000a031916815260200191505b50955050505050506000604051808303816000876161da5a03f115610002575050505b509392505050565b600081600360005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005054101561083157610002565b600360005060008473ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505482600360005060008673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505401101561089e57610002565b600460005060008573ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060003373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505482600560005060008773ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060003373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000505401111561095f57610002565b81600360005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054039250508190555081600360005060008573ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282825054019250508190555081600560005060008673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060005060003373ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828282505401925050819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a35b939250505056`

// DeployGreeter deploys a new Ethereum contract, binding an instance of Greeter to it.
func DeployGreeter(auth *bind.TransactOpts, backend bind.ContractBackend, _greeting string) (common.Address, *types.Transaction, *Greeter, error) {
	parsed, err := abi.JSON(strings.NewReader(GreeterABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(GreeterBin), backend, _greeting)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Greeter{GreeterCaller: GreeterCaller{contract: contract}, GreeterTransactor: GreeterTransactor{contract: contract}}, nil
}

// Greeter is an auto generated Go binding around an Ethereum contract.
type Greeter struct {
	GreeterCaller     // Read-only binding to the contract
	GreeterTransactor // Write-only binding to the contract
}

// GreeterCaller is an auto generated read-only Go binding around an Ethereum contract.
type GreeterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreeterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GreeterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GreeterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GreeterSession struct {
	Contract     *Greeter          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GreeterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GreeterCallerSession struct {
	Contract *GreeterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GreeterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GreeterTransactorSession struct {
	Contract     *GreeterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GreeterRaw is an auto generated low-level Go binding around an Ethereum contract.
type GreeterRaw struct {
	Contract *Greeter // Generic contract binding to access the raw methods on
}

// GreeterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GreeterCallerRaw struct {
	Contract *GreeterCaller // Generic read-only contract binding to access the raw methods on
}

// GreeterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GreeterTransactorRaw struct {
	Contract *GreeterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGreeter creates a new instance of Greeter, bound to a specific deployed contract.
func NewGreeter(address common.Address, backend bind.ContractBackend) (*Greeter, error) {
	contract, err := bindGreeter(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Greeter{GreeterCaller: GreeterCaller{contract: contract}, GreeterTransactor: GreeterTransactor{contract: contract}}, nil
}

// NewGreeterCaller creates a new read-only instance of Greeter, bound to a specific deployed contract.
func NewGreeterCaller(address common.Address, caller bind.ContractCaller) (*GreeterCaller, error) {
	contract, err := bindGreeter(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &GreeterCaller{contract: contract}, nil
}

// NewGreeterTransactor creates a new write-only instance of Greeter, bound to a specific deployed contract.
func NewGreeterTransactor(address common.Address, transactor bind.ContractTransactor) (*GreeterTransactor, error) {
	contract, err := bindGreeter(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &GreeterTransactor{contract: contract}, nil
}

// bindGreeter binds a generic wrapper to an already deployed contract.
func bindGreeter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GreeterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Greeter *GreeterRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Greeter.Contract.GreeterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Greeter *GreeterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Greeter.Contract.GreeterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Greeter *GreeterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Greeter.Contract.GreeterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Greeter *GreeterCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Greeter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Greeter *GreeterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Greeter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Greeter *GreeterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Greeter.Contract.contract.Transact(opts, method, params...)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Greeter *GreeterCaller) Greet(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Greeter.contract.Call(opts, out, "greet")
	return *ret0, err
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Greeter *GreeterSession) Greet() (string, error) {
	return _Greeter.Contract.Greet(&_Greeter.CallOpts)
}

// Greet is a free data retrieval call binding the contract method 0xcfae3217.
//
// Solidity: function greet() constant returns(string)
func (_Greeter *GreeterCallerSession) Greet() (string, error) {
	return _Greeter.Contract.Greet(&_Greeter.CallOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Greeter *GreeterTransactor) Kill(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Greeter.contract.Transact(opts, "kill")
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Greeter *GreeterSession) Kill() (*types.Transaction, error) {
	return _Greeter.Contract.Kill(&_Greeter.TransactOpts)
}

// Kill is a paid mutator transaction binding the contract method 0x41c0e1b5.
//
// Solidity: function kill() returns()
func (_Greeter *GreeterTransactorSession) Kill() (*types.Transaction, error) {
	return _Greeter.Contract.Kill(&_Greeter.TransactOpts)
}
