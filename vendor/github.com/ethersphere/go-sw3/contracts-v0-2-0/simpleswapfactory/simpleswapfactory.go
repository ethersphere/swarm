// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package simpleswapfactory

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

// ContextABI is the input ABI used to generate the binding from.
const ContextABI = "[{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// Context is an auto generated Go binding around an Ethereum contract.
type Context struct {
	ContextCaller     // Read-only binding to the contract
	ContextTransactor // Write-only binding to the contract
	ContextFilterer   // Log filterer for contract events
}

// ContextCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContextCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContextTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContextFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContextSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContextSession struct {
	Contract     *Context          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContextCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContextCallerSession struct {
	Contract *ContextCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ContextTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContextTransactorSession struct {
	Contract     *ContextTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ContextRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContextRaw struct {
	Contract *Context // Generic contract binding to access the raw methods on
}

// ContextCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContextCallerRaw struct {
	Contract *ContextCaller // Generic read-only contract binding to access the raw methods on
}

// ContextTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContextTransactorRaw struct {
	Contract *ContextTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContext creates a new instance of Context, bound to a specific deployed contract.
func NewContext(address common.Address, backend bind.ContractBackend) (*Context, error) {
	contract, err := bindContext(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Context{ContextCaller: ContextCaller{contract: contract}, ContextTransactor: ContextTransactor{contract: contract}, ContextFilterer: ContextFilterer{contract: contract}}, nil
}

// NewContextCaller creates a new read-only instance of Context, bound to a specific deployed contract.
func NewContextCaller(address common.Address, caller bind.ContractCaller) (*ContextCaller, error) {
	contract, err := bindContext(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContextCaller{contract: contract}, nil
}

// NewContextTransactor creates a new write-only instance of Context, bound to a specific deployed contract.
func NewContextTransactor(address common.Address, transactor bind.ContractTransactor) (*ContextTransactor, error) {
	contract, err := bindContext(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContextTransactor{contract: contract}, nil
}

// NewContextFilterer creates a new log filterer instance of Context, bound to a specific deployed contract.
func NewContextFilterer(address common.Address, filterer bind.ContractFilterer) (*ContextFilterer, error) {
	contract, err := bindContext(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContextFilterer{contract: contract}, nil
}

// bindContext binds a generic wrapper to an already deployed contract.
func bindContext(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContextABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Context.Contract.ContextCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.ContextTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Context *ContextCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Context.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Context *ContextTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Context.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Context *ContextTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Context.Contract.contract.Transact(opts, method, params...)
}

// ECDSAABI is the input ABI used to generate the binding from.
const ECDSAABI = "[]"

// ECDSABin is the compiled bytecode used for deploying new contracts.
var ECDSABin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820c087da536b18c4187b7cacb424c2e4c58e707414a9642df27f1db97cccc3b7f464736f6c634300050d0032"

// DeployECDSA deploys a new Ethereum contract, binding an instance of ECDSA to it.
func DeployECDSA(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ECDSA, error) {
	parsed, err := abi.JSON(strings.NewReader(ECDSAABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ECDSABin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ECDSA{ECDSACaller: ECDSACaller{contract: contract}, ECDSATransactor: ECDSATransactor{contract: contract}, ECDSAFilterer: ECDSAFilterer{contract: contract}}, nil
}

// ECDSA is an auto generated Go binding around an Ethereum contract.
type ECDSA struct {
	ECDSACaller     // Read-only binding to the contract
	ECDSATransactor // Write-only binding to the contract
	ECDSAFilterer   // Log filterer for contract events
}

// ECDSACaller is an auto generated read-only Go binding around an Ethereum contract.
type ECDSACaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSATransactor is an auto generated write-only Go binding around an Ethereum contract.
type ECDSATransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ECDSAFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSASession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ECDSASession struct {
	Contract     *ECDSA            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECDSACallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ECDSACallerSession struct {
	Contract *ECDSACaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ECDSATransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ECDSATransactorSession struct {
	Contract     *ECDSATransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECDSARaw is an auto generated low-level Go binding around an Ethereum contract.
type ECDSARaw struct {
	Contract *ECDSA // Generic contract binding to access the raw methods on
}

// ECDSACallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ECDSACallerRaw struct {
	Contract *ECDSACaller // Generic read-only contract binding to access the raw methods on
}

// ECDSATransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ECDSATransactorRaw struct {
	Contract *ECDSATransactor // Generic write-only contract binding to access the raw methods on
}

// NewECDSA creates a new instance of ECDSA, bound to a specific deployed contract.
func NewECDSA(address common.Address, backend bind.ContractBackend) (*ECDSA, error) {
	contract, err := bindECDSA(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ECDSA{ECDSACaller: ECDSACaller{contract: contract}, ECDSATransactor: ECDSATransactor{contract: contract}, ECDSAFilterer: ECDSAFilterer{contract: contract}}, nil
}

// NewECDSACaller creates a new read-only instance of ECDSA, bound to a specific deployed contract.
func NewECDSACaller(address common.Address, caller bind.ContractCaller) (*ECDSACaller, error) {
	contract, err := bindECDSA(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSACaller{contract: contract}, nil
}

// NewECDSATransactor creates a new write-only instance of ECDSA, bound to a specific deployed contract.
func NewECDSATransactor(address common.Address, transactor bind.ContractTransactor) (*ECDSATransactor, error) {
	contract, err := bindECDSA(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSATransactor{contract: contract}, nil
}

// NewECDSAFilterer creates a new log filterer instance of ECDSA, bound to a specific deployed contract.
func NewECDSAFilterer(address common.Address, filterer bind.ContractFilterer) (*ECDSAFilterer, error) {
	contract, err := bindECDSA(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ECDSAFilterer{contract: contract}, nil
}

// bindECDSA binds a generic wrapper to an already deployed contract.
func bindECDSA(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ECDSAABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSA *ECDSARaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ECDSA.Contract.ECDSACaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSA *ECDSARaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSA.Contract.ECDSATransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSA *ECDSARaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSA.Contract.ECDSATransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSA *ECDSACallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ECDSA.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSA *ECDSATransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSA.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSA *ECDSATransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSA.Contract.contract.Transact(opts, method, params...)
}

// ERC20ABI is the input ABI used to generate the binding from.
const ERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ERC20Bin is the compiled bytecode used for deploying new contracts.
var ERC20Bin = "0x6080604052610e3a806100136000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c806370a082311161005b57806370a08231146101fd578063a457c2d714610255578063a9059cbb146102bb578063dd62ed3e1461032157610088565b8063095ea7b31461008d57806318160ddd146100f357806323b872dd146101115780633950935114610197575b600080fd5b6100d9600480360360408110156100a357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610399565b604051808215151515815260200191505060405180910390f35b6100fb6103b7565b6040518082815260200191505060405180910390f35b61017d6004803603606081101561012757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506103c1565b604051808215151515815260200191505060405180910390f35b6101e3600480360360408110156101ad57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061049a565b604051808215151515815260200191505060405180910390f35b61023f6004803603602081101561021357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061054d565b6040518082815260200191505060405180910390f35b6102a16004803603604081101561026b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610595565b604051808215151515815260200191505060405180910390f35b610307600480360360408110156102d157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610662565b604051808215151515815260200191505060405180910390f35b6103836004803603604081101561033757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610680565b6040518082815260200191505060405180910390f35b60006103ad6103a6610707565b848461070f565b6001905092915050565b6000600254905090565b60006103ce848484610906565b61048f846103da610707565b61048a85604051806060016040528060288152602001610d7060289139600160008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000610440610707565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610bbc9092919063ffffffff16565b61070f565b600190509392505050565b60006105436104a7610707565b8461053e85600160006104b8610707565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610c7c90919063ffffffff16565b61070f565b6001905092915050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60006106586105a2610707565b8461065385604051806060016040528060258152602001610de160259139600160006105cc610707565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610bbc9092919063ffffffff16565b61070f565b6001905092915050565b600061067661066f610707565b8484610906565b6001905092915050565b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905092915050565b600033905090565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415610795576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526024815260200180610dbd6024913960400191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16141561081b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180610d286022913960400191505060405180910390fd5b80600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925836040518082815260200191505060405180910390a3505050565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16141561098c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526025815260200180610d986025913960400191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610a12576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526023815260200180610d056023913960400191505060405180910390fd5b610a7d81604051806060016040528060268152602001610d4a602691396000808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610bbc9092919063ffffffff16565b6000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610b10816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610c7c90919063ffffffff16565b6000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a3505050565b6000838311158290610c69576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b83811015610c2e578082015181840152602081019050610c13565b50505050905090810190601f168015610c5b5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b5060008385039050809150509392505050565b600080828401905083811015610cfa576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b809150509291505056fe45524332303a207472616e7366657220746f20746865207a65726f206164647265737345524332303a20617070726f766520746f20746865207a65726f206164647265737345524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e636545524332303a207472616e7366657220616d6f756e74206578636565647320616c6c6f77616e636545524332303a207472616e736665722066726f6d20746865207a65726f206164647265737345524332303a20617070726f76652066726f6d20746865207a65726f206164647265737345524332303a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa265627a7a7231582053b4b44d2fad2123e25198aa5ed2b9e7b2ed14bfa733eb6d7d64e1cf403ba57964736f6c634300050d0032"

// DeployERC20 deploys a new Ethereum contract, binding an instance of ERC20 to it.
func DeployERC20(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ERC20, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ERC20Bin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20{ERC20Caller: ERC20Caller{contract: contract}, ERC20Transactor: ERC20Transactor{contract: contract}, ERC20Filterer: ERC20Filterer{contract: contract}}, nil
}

// ERC20 is an auto generated Go binding around an Ethereum contract.
type ERC20 struct {
	ERC20Caller     // Read-only binding to the contract
	ERC20Transactor // Write-only binding to the contract
	ERC20Filterer   // Log filterer for contract events
}

// ERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20Session struct {
	Contract     *ERC20            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20CallerSession struct {
	Contract *ERC20Caller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20TransactorSession struct {
	Contract     *ERC20Transactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20Raw struct {
	Contract *ERC20 // Generic contract binding to access the raw methods on
}

// ERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20CallerRaw struct {
	Contract *ERC20Caller // Generic read-only contract binding to access the raw methods on
}

// ERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20TransactorRaw struct {
	Contract *ERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20 creates a new instance of ERC20, bound to a specific deployed contract.
func NewERC20(address common.Address, backend bind.ContractBackend) (*ERC20, error) {
	contract, err := bindERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20{ERC20Caller: ERC20Caller{contract: contract}, ERC20Transactor: ERC20Transactor{contract: contract}, ERC20Filterer: ERC20Filterer{contract: contract}}, nil
}

// NewERC20Caller creates a new read-only instance of ERC20, bound to a specific deployed contract.
func NewERC20Caller(address common.Address, caller bind.ContractCaller) (*ERC20Caller, error) {
	contract, err := bindERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20Caller{contract: contract}, nil
}

// NewERC20Transactor creates a new write-only instance of ERC20, bound to a specific deployed contract.
func NewERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*ERC20Transactor, error) {
	contract, err := bindERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20Transactor{contract: contract}, nil
}

// NewERC20Filterer creates a new log filterer instance of ERC20, bound to a specific deployed contract.
func NewERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*ERC20Filterer, error) {
	contract, err := bindERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20Filterer{contract: contract}, nil
}

// bindERC20 binds a generic wrapper to an already deployed contract.
func bindERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20 *ERC20Raw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20.Contract.ERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20 *ERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20.Contract.ERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20 *ERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20.Contract.ERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20 *ERC20CallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20 *ERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20 *ERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20 *ERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "allowance", owner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20 *ERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20.Contract.Allowance(&_ERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20 *ERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20.Contract.Allowance(&_ERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20 *ERC20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "balanceOf", account)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20 *ERC20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20.Contract.BalanceOf(&_ERC20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20 *ERC20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20.Contract.BalanceOf(&_ERC20.CallOpts, account)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20Session) TotalSupply() (*big.Int, error) {
	return _ERC20.Contract.TotalSupply(&_ERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20 *ERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20.Contract.TotalSupply(&_ERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20 *ERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20 *ERC20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Approve(&_ERC20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20 *ERC20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Approve(&_ERC20.TransactOpts, spender, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20 *ERC20Transactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20 *ERC20Session) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.DecreaseAllowance(&_ERC20.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20 *ERC20TransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.DecreaseAllowance(&_ERC20.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20 *ERC20Transactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20 *ERC20Session) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.IncreaseAllowance(&_ERC20.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20 *ERC20TransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.IncreaseAllowance(&_ERC20.TransactOpts, spender, addedValue)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20Transactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20Session) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Transfer(&_ERC20.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20TransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.Transfer(&_ERC20.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20Transactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20Session) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.TransferFrom(&_ERC20.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20 *ERC20TransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20.Contract.TransferFrom(&_ERC20.TransactOpts, sender, recipient, amount)
}

// ERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20 contract.
type ERC20ApprovalIterator struct {
	Event *ERC20Approval // Event containing the contract specifics and raw log

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
func (it *ERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20Approval)
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
		it.Event = new(ERC20Approval)
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
func (it *ERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20Approval represents a Approval event raised by the ERC20 contract.
type ERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20 *ERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20ApprovalIterator{contract: _ERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20 *ERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20Approval)
				if err := _ERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20 *ERC20Filterer) ParseApproval(log types.Log) (*ERC20Approval, error) {
	event := new(ERC20Approval)
	if err := _ERC20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20 contract.
type ERC20TransferIterator struct {
	Event *ERC20Transfer // Event containing the contract specifics and raw log

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
func (it *ERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20Transfer)
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
		it.Event = new(ERC20Transfer)
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
func (it *ERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20Transfer represents a Transfer event raised by the ERC20 contract.
type ERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20 *ERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20TransferIterator{contract: _ERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20 *ERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20Transfer)
				if err := _ERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20 *ERC20Filterer) ParseTransfer(log types.Log) (*ERC20Transfer, error) {
	event := new(ERC20Transfer)
	if err := _ERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20MintableABI is the input ABI used to generate the binding from.
const ERC20MintableABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"MinterAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"MinterRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"addMinter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isMinter\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceMinter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ERC20MintableBin is the compiled bytecode used for deploying new contracts.
var ERC20MintableBin = "0x608060405262000024620000186200002a60201b60201c565b6200003260201b60201c565b62000257565b600033905090565b6200004d8160036200009360201b620012a81790919060201c565b8073ffffffffffffffffffffffffffffffffffffffff167f6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f660405160405180910390a250565b620000a582826200017760201b60201c565b1562000119576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601f8152602001807f526f6c65733a206163636f756e7420616c72656164792068617320726f6c650081525060200191505060405180910390fd5b60018260000160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055505050565b60008073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16141562000200576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180620018506022913960400191505060405180910390fd5b8260000160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16905092915050565b6115e980620002676000396000f3fe608060405234801561001057600080fd5b50600436106100b45760003560e01c8063983b2d5611610071578063983b2d56146102e7578063986502751461032b578063a457c2d714610335578063a9059cbb1461039b578063aa271e1a14610401578063dd62ed3e1461045d576100b4565b8063095ea7b3146100b957806318160ddd1461011f57806323b872dd1461013d57806339509351146101c357806340c10f191461022957806370a082311461028f575b600080fd5b610105600480360360408110156100cf57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506104d5565b604051808215151515815260200191505060405180910390f35b6101276104f3565b6040518082815260200191505060405180910390f35b6101a96004803603606081101561015357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506104fd565b604051808215151515815260200191505060405180910390f35b61020f600480360360408110156101d957600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506105d6565b604051808215151515815260200191505060405180910390f35b6102756004803603604081101561023f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610689565b604051808215151515815260200191505060405180910390f35b6102d1600480360360208110156102a557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610704565b6040518082815260200191505060405180910390f35b610329600480360360208110156102fd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061074c565b005b6103336107bd565b005b6103816004803603604081101561034b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506107cf565b604051808215151515815260200191505060405180910390f35b6103e7600480360360408110156103b157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061089c565b604051808215151515815260200191505060405180910390f35b6104436004803603602081101561041757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506108ba565b604051808215151515815260200191505060405180910390f35b6104bf6004803603604081101561047357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506108d7565b6040518082815260200191505060405180910390f35b60006104e96104e261095e565b8484610966565b6001905092915050565b6000600254905090565b600061050a848484610b5d565b6105cb8461051661095e565b6105c6856040518060600160405280602881526020016114fd60289139600160008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600061057c61095e565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610e139092919063ffffffff16565b610966565b600190509392505050565b600061067f6105e361095e565b8461067a85600160006105f461095e565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610ed390919063ffffffff16565b610966565b6001905092915050565b600061069b61069661095e565b6108ba565b6106f0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260308152602001806114ac6030913960400191505060405180910390fd5b6106fa8383610f5b565b6001905092915050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b61075c61075761095e565b6108ba565b6107b1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260308152602001806114ac6030913960400191505060405180910390fd5b6107ba81611116565b50565b6107cd6107c861095e565b611170565b565b60006108926107dc61095e565b8461088d85604051806060016040528060258152602001611590602591396001600061080661095e565b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610e139092919063ffffffff16565b610966565b6001905092915050565b60006108b06108a961095e565b8484610b5d565b6001905092915050565b60006108d08260036111ca90919063ffffffff16565b9050919050565b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905092915050565b600033905090565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614156109ec576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602481526020018061156c6024913960400191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610a72576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806114646022913960400191505060405180910390fd5b80600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925836040518082815260200191505060405180910390a3505050565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161415610be3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260258152602001806115476025913960400191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610c69576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260238152602001806114416023913960400191505060405180910390fd5b610cd481604051806060016040528060268152602001611486602691396000808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610e139092919063ffffffff16565b6000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610d67816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610ed390919063ffffffff16565b6000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a3505050565b6000838311158290610ec0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b83811015610e85578082015181840152602081019050610e6a565b50505050905090810190601f168015610eb25780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b5060008385039050809150509392505050565b600080828401905083811015610f51576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415610ffe576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601f8152602001807f45524332303a206d696e7420746f20746865207a65726f20616464726573730081525060200191505060405180910390fd5b61101381600254610ed390919063ffffffff16565b60028190555061106a816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054610ed390919063ffffffff16565b6000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a35050565b61112a8160036112a890919063ffffffff16565b8073ffffffffffffffffffffffffffffffffffffffff167f6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f660405160405180910390a250565b61118481600361138390919063ffffffff16565b8073ffffffffffffffffffffffffffffffffffffffff167fe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb6669260405160405180910390a250565b60008073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff161415611251576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806115256022913960400191505060405180910390fd5b8260000160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16905092915050565b6112b282826111ca565b15611325576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601f8152602001807f526f6c65733a206163636f756e7420616c72656164792068617320726f6c650081525060200191505060405180910390fd5b60018260000160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055505050565b61138d82826111ca565b6113e2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260218152602001806114dc6021913960400191505060405180910390fd5b60008260000160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550505056fe45524332303a207472616e7366657220746f20746865207a65726f206164647265737345524332303a20617070726f766520746f20746865207a65726f206164647265737345524332303a207472616e7366657220616d6f756e7420657863656564732062616c616e63654d696e746572526f6c653a2063616c6c657220646f6573206e6f74206861766520746865204d696e74657220726f6c65526f6c65733a206163636f756e7420646f6573206e6f74206861766520726f6c6545524332303a207472616e7366657220616d6f756e74206578636565647320616c6c6f77616e6365526f6c65733a206163636f756e7420697320746865207a65726f206164647265737345524332303a207472616e736665722066726f6d20746865207a65726f206164647265737345524332303a20617070726f76652066726f6d20746865207a65726f206164647265737345524332303a2064656372656173656420616c6c6f77616e63652062656c6f77207a65726fa265627a7a72315820d814ae3015d19c38aed909871af148dee0d88a6be96dc04f5c4a57e543a3f83364736f6c634300050d0032526f6c65733a206163636f756e7420697320746865207a65726f2061646472657373"

// DeployERC20Mintable deploys a new Ethereum contract, binding an instance of ERC20Mintable to it.
func DeployERC20Mintable(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ERC20Mintable, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20MintableABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ERC20MintableBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20Mintable{ERC20MintableCaller: ERC20MintableCaller{contract: contract}, ERC20MintableTransactor: ERC20MintableTransactor{contract: contract}, ERC20MintableFilterer: ERC20MintableFilterer{contract: contract}}, nil
}

// ERC20Mintable is an auto generated Go binding around an Ethereum contract.
type ERC20Mintable struct {
	ERC20MintableCaller     // Read-only binding to the contract
	ERC20MintableTransactor // Write-only binding to the contract
	ERC20MintableFilterer   // Log filterer for contract events
}

// ERC20MintableCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20MintableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MintableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20MintableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MintableFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20MintableFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20MintableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20MintableSession struct {
	Contract     *ERC20Mintable    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20MintableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20MintableCallerSession struct {
	Contract *ERC20MintableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ERC20MintableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20MintableTransactorSession struct {
	Contract     *ERC20MintableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ERC20MintableRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20MintableRaw struct {
	Contract *ERC20Mintable // Generic contract binding to access the raw methods on
}

// ERC20MintableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20MintableCallerRaw struct {
	Contract *ERC20MintableCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20MintableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20MintableTransactorRaw struct {
	Contract *ERC20MintableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20Mintable creates a new instance of ERC20Mintable, bound to a specific deployed contract.
func NewERC20Mintable(address common.Address, backend bind.ContractBackend) (*ERC20Mintable, error) {
	contract, err := bindERC20Mintable(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20Mintable{ERC20MintableCaller: ERC20MintableCaller{contract: contract}, ERC20MintableTransactor: ERC20MintableTransactor{contract: contract}, ERC20MintableFilterer: ERC20MintableFilterer{contract: contract}}, nil
}

// NewERC20MintableCaller creates a new read-only instance of ERC20Mintable, bound to a specific deployed contract.
func NewERC20MintableCaller(address common.Address, caller bind.ContractCaller) (*ERC20MintableCaller, error) {
	contract, err := bindERC20Mintable(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableCaller{contract: contract}, nil
}

// NewERC20MintableTransactor creates a new write-only instance of ERC20Mintable, bound to a specific deployed contract.
func NewERC20MintableTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20MintableTransactor, error) {
	contract, err := bindERC20Mintable(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableTransactor{contract: contract}, nil
}

// NewERC20MintableFilterer creates a new log filterer instance of ERC20Mintable, bound to a specific deployed contract.
func NewERC20MintableFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20MintableFilterer, error) {
	contract, err := bindERC20Mintable(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableFilterer{contract: contract}, nil
}

// bindERC20Mintable binds a generic wrapper to an already deployed contract.
func bindERC20Mintable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20MintableABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Mintable *ERC20MintableRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20Mintable.Contract.ERC20MintableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Mintable *ERC20MintableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.ERC20MintableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Mintable *ERC20MintableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.ERC20MintableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20Mintable *ERC20MintableCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20Mintable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20Mintable *ERC20MintableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20Mintable *ERC20MintableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20Mintable.contract.Call(opts, out, "allowance", owner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20Mintable.Contract.Allowance(&_ERC20Mintable.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ERC20Mintable.Contract.Allowance(&_ERC20Mintable.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20Mintable.contract.Call(opts, out, "balanceOf", account)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20Mintable.Contract.BalanceOf(&_ERC20Mintable.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ERC20Mintable.Contract.BalanceOf(&_ERC20Mintable.CallOpts, account)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_ERC20Mintable *ERC20MintableCaller) IsMinter(opts *bind.CallOpts, account common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ERC20Mintable.contract.Call(opts, out, "isMinter", account)
	return *ret0, err
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_ERC20Mintable *ERC20MintableSession) IsMinter(account common.Address) (bool, error) {
	return _ERC20Mintable.Contract.IsMinter(&_ERC20Mintable.CallOpts, account)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_ERC20Mintable *ERC20MintableCallerSession) IsMinter(account common.Address) (bool, error) {
	return _ERC20Mintable.Contract.IsMinter(&_ERC20Mintable.CallOpts, account)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20Mintable.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20Mintable *ERC20MintableSession) TotalSupply() (*big.Int, error) {
	return _ERC20Mintable.Contract.TotalSupply(&_ERC20Mintable.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_ERC20Mintable *ERC20MintableCallerSession) TotalSupply() (*big.Int, error) {
	return _ERC20Mintable.Contract.TotalSupply(&_ERC20Mintable.CallOpts)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_ERC20Mintable *ERC20MintableTransactor) AddMinter(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "addMinter", account)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_ERC20Mintable *ERC20MintableSession) AddMinter(account common.Address) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.AddMinter(&_ERC20Mintable.TransactOpts, account)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_ERC20Mintable *ERC20MintableTransactorSession) AddMinter(account common.Address) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.AddMinter(&_ERC20Mintable.TransactOpts, account)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Approve(&_ERC20Mintable.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Approve(&_ERC20Mintable.TransactOpts, spender, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.DecreaseAllowance(&_ERC20Mintable.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.DecreaseAllowance(&_ERC20Mintable.TransactOpts, spender, subtractedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.IncreaseAllowance(&_ERC20Mintable.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.IncreaseAllowance(&_ERC20Mintable.TransactOpts, spender, addedValue)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) Mint(opts *bind.TransactOpts, account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "mint", account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Mint(&_ERC20Mintable.TransactOpts, account, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address account, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) Mint(account common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Mint(&_ERC20Mintable.TransactOpts, account, amount)
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_ERC20Mintable *ERC20MintableTransactor) RenounceMinter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "renounceMinter")
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_ERC20Mintable *ERC20MintableSession) RenounceMinter() (*types.Transaction, error) {
	return _ERC20Mintable.Contract.RenounceMinter(&_ERC20Mintable.TransactOpts)
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_ERC20Mintable *ERC20MintableTransactorSession) RenounceMinter() (*types.Transaction, error) {
	return _ERC20Mintable.Contract.RenounceMinter(&_ERC20Mintable.TransactOpts)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Transfer(&_ERC20Mintable.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.Transfer(&_ERC20Mintable.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.TransferFrom(&_ERC20Mintable.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_ERC20Mintable *ERC20MintableTransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20Mintable.Contract.TransferFrom(&_ERC20Mintable.TransactOpts, sender, recipient, amount)
}

// ERC20MintableApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ERC20Mintable contract.
type ERC20MintableApprovalIterator struct {
	Event *ERC20MintableApproval // Event containing the contract specifics and raw log

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
func (it *ERC20MintableApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MintableApproval)
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
		it.Event = new(ERC20MintableApproval)
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
func (it *ERC20MintableApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MintableApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MintableApproval represents a Approval event raised by the ERC20Mintable contract.
type ERC20MintableApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ERC20MintableApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20Mintable.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableApprovalIterator{contract: _ERC20Mintable.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ERC20MintableApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ERC20Mintable.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MintableApproval)
				if err := _ERC20Mintable.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) ParseApproval(log types.Log) (*ERC20MintableApproval, error) {
	event := new(ERC20MintableApproval)
	if err := _ERC20Mintable.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20MintableMinterAddedIterator is returned from FilterMinterAdded and is used to iterate over the raw logs and unpacked data for MinterAdded events raised by the ERC20Mintable contract.
type ERC20MintableMinterAddedIterator struct {
	Event *ERC20MintableMinterAdded // Event containing the contract specifics and raw log

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
func (it *ERC20MintableMinterAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MintableMinterAdded)
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
		it.Event = new(ERC20MintableMinterAdded)
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
func (it *ERC20MintableMinterAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MintableMinterAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MintableMinterAdded represents a MinterAdded event raised by the ERC20Mintable contract.
type ERC20MintableMinterAdded struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMinterAdded is a free log retrieval operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) FilterMinterAdded(opts *bind.FilterOpts, account []common.Address) (*ERC20MintableMinterAddedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ERC20Mintable.contract.FilterLogs(opts, "MinterAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableMinterAddedIterator{contract: _ERC20Mintable.contract, event: "MinterAdded", logs: logs, sub: sub}, nil
}

// WatchMinterAdded is a free log subscription operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) WatchMinterAdded(opts *bind.WatchOpts, sink chan<- *ERC20MintableMinterAdded, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ERC20Mintable.contract.WatchLogs(opts, "MinterAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MintableMinterAdded)
				if err := _ERC20Mintable.contract.UnpackLog(event, "MinterAdded", log); err != nil {
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

// ParseMinterAdded is a log parse operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) ParseMinterAdded(log types.Log) (*ERC20MintableMinterAdded, error) {
	event := new(ERC20MintableMinterAdded)
	if err := _ERC20Mintable.contract.UnpackLog(event, "MinterAdded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20MintableMinterRemovedIterator is returned from FilterMinterRemoved and is used to iterate over the raw logs and unpacked data for MinterRemoved events raised by the ERC20Mintable contract.
type ERC20MintableMinterRemovedIterator struct {
	Event *ERC20MintableMinterRemoved // Event containing the contract specifics and raw log

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
func (it *ERC20MintableMinterRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MintableMinterRemoved)
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
		it.Event = new(ERC20MintableMinterRemoved)
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
func (it *ERC20MintableMinterRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MintableMinterRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MintableMinterRemoved represents a MinterRemoved event raised by the ERC20Mintable contract.
type ERC20MintableMinterRemoved struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMinterRemoved is a free log retrieval operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) FilterMinterRemoved(opts *bind.FilterOpts, account []common.Address) (*ERC20MintableMinterRemovedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ERC20Mintable.contract.FilterLogs(opts, "MinterRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableMinterRemovedIterator{contract: _ERC20Mintable.contract, event: "MinterRemoved", logs: logs, sub: sub}, nil
}

// WatchMinterRemoved is a free log subscription operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) WatchMinterRemoved(opts *bind.WatchOpts, sink chan<- *ERC20MintableMinterRemoved, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ERC20Mintable.contract.WatchLogs(opts, "MinterRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MintableMinterRemoved)
				if err := _ERC20Mintable.contract.UnpackLog(event, "MinterRemoved", log); err != nil {
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

// ParseMinterRemoved is a log parse operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_ERC20Mintable *ERC20MintableFilterer) ParseMinterRemoved(log types.Log) (*ERC20MintableMinterRemoved, error) {
	event := new(ERC20MintableMinterRemoved)
	if err := _ERC20Mintable.contract.UnpackLog(event, "MinterRemoved", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20MintableTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ERC20Mintable contract.
type ERC20MintableTransferIterator struct {
	Event *ERC20MintableTransfer // Event containing the contract specifics and raw log

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
func (it *ERC20MintableTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20MintableTransfer)
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
		it.Event = new(ERC20MintableTransfer)
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
func (it *ERC20MintableTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20MintableTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20MintableTransfer represents a Transfer event raised by the ERC20Mintable contract.
type ERC20MintableTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ERC20MintableTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20Mintable.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ERC20MintableTransferIterator{contract: _ERC20Mintable.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ERC20MintableTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ERC20Mintable.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20MintableTransfer)
				if err := _ERC20Mintable.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ERC20Mintable *ERC20MintableFilterer) ParseTransfer(log types.Log) (*ERC20MintableTransfer, error) {
	event := new(ERC20MintableTransfer)
	if err := _ERC20Mintable.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapABI is the input ABI used to generate the binding from.
const ERC20SimpleSwapABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_issuer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_defaultHardDepositTimeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"callerPayout\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositAmountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"}],\"name\":\"HardDepositTimeoutChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"beneficiarySig\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"callerPayout\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"cashChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"defaultHardDepositTimeout\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"decreaseAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"canBeDecreasedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"issuer\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"liquidBalanceFor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"paidOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"beneficiary\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"hardDepositTimeout\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"setCustomHardDepositTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalHardDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalPaidOut\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ERC20SimpleSwapBin is the compiled bytecode used for deploying new contracts.
var ERC20SimpleSwapBin = "0x608060405234801561001057600080fd5b506040516122eb3803806122eb8339818101604052606081101561003357600080fd5b8101908080519060200190929190805190602001909291908051906020019092919050505082600660006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555081600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550806000819055505050506121f8806100f36000396000f3fe608060405234801561001057600080fd5b506004361061010b5760003560e01c8063946f46a2116100a2578063b7ec1a3311610071578063b7ec1a3314610612578063c76a4d3114610630578063d4c9a8e814610688578063e0bcf13a1461076d578063fc0c546a1461078b5761010b565b8063946f46a2146104f5578063b6343b0d14610539578063b69ef8a8146105a6578063b7770350146105c45761010b565b80631d143848116100de5780631d143848146103d75780632e1a7d4d14610421578063338f3fed1461044f57806381f03fcb1461049d5761010b565b80630d5f26591461011057806312101021146101f55780631357e1dc146102135780631633fb1d14610231575b600080fd5b6101f36004803603606081101561012657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019064010000000081111561016d57600080fd5b82018360208201111561017f57600080fd5b803590602001918460018302840111640100000000831117156101a157600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506107d5565b005b6101fd6107e8565b6040518082815260200191505060405180910390f35b61021b6107ee565b6040518082815260200191505060405180910390f35b6103d5600480360360c081101561024757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156102ae57600080fd5b8201836020820111156102c057600080fd5b803590602001918460018302840111640100000000831117156102e257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001909291908035906020019064010000000081111561034f57600080fd5b82018360208201111561036157600080fd5b8035906020019184600183028401116401000000008311171561038357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506107f4565b005b6103df6108a2565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61044d6004803603602081101561043757600080fd5b81019080803590602001909291905050506108c8565b005b61049b6004803603604081101561046557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610b49565b005b6104df600480360360208110156104b357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d5e565b6040518082815260200191505060405180910390f35b6105376004803603602081101561050b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d76565b005b61057b6004803603602081101561054f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610ec9565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b6105ae610ef9565b6040518082815260200191505060405180910390f35b610610600480360360408110156105da57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610fda565b005b61061a6111c2565b6040518082815260200191505060405180910390f35b6106726004803603602081101561064657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506111e5565b6040518082815260200191505060405180910390f35b61076b6004803603606081101561069e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156106e557600080fd5b8201836020820111156106f757600080fd5b8035906020019184600183028401116401000000008311171561071957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061124a565b005b61077561143e565b6040518082815260200191505060405180910390f35b610793611444565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6107e333848460008561146a565b505050565b60005481565b60035481565b61080a6108043033878987611b8e565b84611c6f565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161461088d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806121076022913960400191505060405180910390fd5b61089a868686858561146a565b505050505050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461098b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6109936111c2565b8111156109eb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602881526020018061219c6028913960400191505060405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16836040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b158015610ab657600080fd5b505af1158015610aca573d6000803e3d6000fd5b505050506040513d6020811015610ae057600080fd5b8101908080519060200190929190505050610b46576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b50565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c0c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610c14610ef9565b610c2982600554611c8b90919063ffffffff16565b1115610c80576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260348152602001806120d36034913960400191505060405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050610cda828260000154611c8b90919063ffffffff16565b8160000181905550610cf782600554611c8b90919063ffffffff16565b600581905550600081600301819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60026020528060005260406000206000915090505481565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600301544210158015610dd257506000816003015414155b610e27576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260258152602001806121296025913960400191505060405180910390fd5b610e4281600101548260000154611d1390919063ffffffff16565b816000018190555060008160030181905550610e6d8160010154600554611d1390919063ffffffff16565b6005819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60046020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b6000600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b158015610f9a57600080fd5b505afa158015610fae573d6000803e3d6000fd5b505050506040513d6020811015610fc457600080fd5b8101908080519060200190929190505050905090565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461109d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050806000015482111561113d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260278152602001806121756027913960400191505060405180910390fd5b600080826002015414611154578160020154611158565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b60006111e06005546111d2610ef9565b611d1390919063ffffffff16565b905090565b6000611243600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546112356111c2565b611c8b90919063ffffffff16565b9050919050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461130d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b61132161131b308585611d5d565b82611c6f565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16146113a4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806121076022913960400191505060405180910390fd5b81600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf0836040518082815260200191505060405180910390a2505050565b60055481565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611596576114d36114cd308786611dfd565b82611c6f565b73ffffffffffffffffffffffffffffffffffffffff16600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611595576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b5b60006115ea600260008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205485611d1390919063ffffffff16565b90506000611600826115fb896111e5565b611e9d565b9050600061165082600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611e9d565b9050848210156116c8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a2063616e6e6f74207061792063616c6c657200000081525060200191505060405180910390fd5b600081146117875761172581600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611d1390919063ffffffff16565b600460008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000018190555061178081600554611d1390919063ffffffff16565b6005819055505b6117d982600260008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054611c8b90919063ffffffff16565b600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555061183182600354611c8b90919063ffffffff16565b600381905550600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb886118898886611d1390919063ffffffff16565b6040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b1580156118f257600080fd5b505af1158015611906573d6000803e3d6000fd5b505050506040513d602081101561191c57600080fd5b8101908080519060200190929190505050611982576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b60008514611ac457600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb33876040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b158015611a3357600080fd5b505af1158015611a47573d6000803e3d6000fd5b505050506040513d6020811015611a5d57600080fd5b8101908080519060200190929190505050611ac3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b5b3373ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f858a8a60405180848152602001838152602001828152602001935050505060405180910390a4818314611b84577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b5050505050505050565b60008585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b6000611c83611c7d84611eb6565b83611f0e565b905092915050565b600080828401905083811015611d09576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b6000611d5583836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250612012565b905092915050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000818310611eac5781611eae565b825b905092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b60006041825114611f22576000905061200c565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c1115611f76576000935050505061200c565b601b8160ff1614158015611f8e5750601c8160ff1614155b15611f9f576000935050505061200c565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611ffc573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b60008383111582906120bf576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b83811015612084578082015181840152602081019050612069565b50505050905090810190601f1680156120b15780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b506000838503905080915050939250505056fe53696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e636553696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a2053696d706c65537761703a207472616e73666572206661696c656453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e74a265627a7a72315820a2829ec1c4ee1a4aa11d90f04bfe854a020b9b07943b7ae56603fb1d301604be64736f6c634300050d0032"

// DeployERC20SimpleSwap deploys a new Ethereum contract, binding an instance of ERC20SimpleSwap to it.
func DeployERC20SimpleSwap(auth *bind.TransactOpts, backend bind.ContractBackend, _issuer common.Address, _token common.Address, _defaultHardDepositTimeout *big.Int) (common.Address, *types.Transaction, *ERC20SimpleSwap, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20SimpleSwapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ERC20SimpleSwapBin), backend, _issuer, _token, _defaultHardDepositTimeout)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ERC20SimpleSwap{ERC20SimpleSwapCaller: ERC20SimpleSwapCaller{contract: contract}, ERC20SimpleSwapTransactor: ERC20SimpleSwapTransactor{contract: contract}, ERC20SimpleSwapFilterer: ERC20SimpleSwapFilterer{contract: contract}}, nil
}

// ERC20SimpleSwap is an auto generated Go binding around an Ethereum contract.
type ERC20SimpleSwap struct {
	ERC20SimpleSwapCaller     // Read-only binding to the contract
	ERC20SimpleSwapTransactor // Write-only binding to the contract
	ERC20SimpleSwapFilterer   // Log filterer for contract events
}

// ERC20SimpleSwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type ERC20SimpleSwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SimpleSwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ERC20SimpleSwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SimpleSwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ERC20SimpleSwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ERC20SimpleSwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ERC20SimpleSwapSession struct {
	Contract     *ERC20SimpleSwap  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ERC20SimpleSwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ERC20SimpleSwapCallerSession struct {
	Contract *ERC20SimpleSwapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ERC20SimpleSwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ERC20SimpleSwapTransactorSession struct {
	Contract     *ERC20SimpleSwapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ERC20SimpleSwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type ERC20SimpleSwapRaw struct {
	Contract *ERC20SimpleSwap // Generic contract binding to access the raw methods on
}

// ERC20SimpleSwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ERC20SimpleSwapCallerRaw struct {
	Contract *ERC20SimpleSwapCaller // Generic read-only contract binding to access the raw methods on
}

// ERC20SimpleSwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ERC20SimpleSwapTransactorRaw struct {
	Contract *ERC20SimpleSwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewERC20SimpleSwap creates a new instance of ERC20SimpleSwap, bound to a specific deployed contract.
func NewERC20SimpleSwap(address common.Address, backend bind.ContractBackend) (*ERC20SimpleSwap, error) {
	contract, err := bindERC20SimpleSwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwap{ERC20SimpleSwapCaller: ERC20SimpleSwapCaller{contract: contract}, ERC20SimpleSwapTransactor: ERC20SimpleSwapTransactor{contract: contract}, ERC20SimpleSwapFilterer: ERC20SimpleSwapFilterer{contract: contract}}, nil
}

// NewERC20SimpleSwapCaller creates a new read-only instance of ERC20SimpleSwap, bound to a specific deployed contract.
func NewERC20SimpleSwapCaller(address common.Address, caller bind.ContractCaller) (*ERC20SimpleSwapCaller, error) {
	contract, err := bindERC20SimpleSwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapCaller{contract: contract}, nil
}

// NewERC20SimpleSwapTransactor creates a new write-only instance of ERC20SimpleSwap, bound to a specific deployed contract.
func NewERC20SimpleSwapTransactor(address common.Address, transactor bind.ContractTransactor) (*ERC20SimpleSwapTransactor, error) {
	contract, err := bindERC20SimpleSwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapTransactor{contract: contract}, nil
}

// NewERC20SimpleSwapFilterer creates a new log filterer instance of ERC20SimpleSwap, bound to a specific deployed contract.
func NewERC20SimpleSwapFilterer(address common.Address, filterer bind.ContractFilterer) (*ERC20SimpleSwapFilterer, error) {
	contract, err := bindERC20SimpleSwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapFilterer{contract: contract}, nil
}

// bindERC20SimpleSwap binds a generic wrapper to an already deployed contract.
func bindERC20SimpleSwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ERC20SimpleSwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20SimpleSwap *ERC20SimpleSwapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20SimpleSwap.Contract.ERC20SimpleSwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20SimpleSwap *ERC20SimpleSwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.ERC20SimpleSwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20SimpleSwap *ERC20SimpleSwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.ERC20SimpleSwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ERC20SimpleSwap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.contract.Transact(opts, method, params...)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) Balance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "balance")
	return *ret0, err
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) Balance() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.Balance(&_ERC20SimpleSwap.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) Balance() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.Balance(&_ERC20SimpleSwap.CallOpts)
}

// DefaultHardDepositTimeout is a free data retrieval call binding the contract method 0x12101021.
//
// Solidity: function defaultHardDepositTimeout() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) DefaultHardDepositTimeout(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "defaultHardDepositTimeout")
	return *ret0, err
}

// DefaultHardDepositTimeout is a free data retrieval call binding the contract method 0x12101021.
//
// Solidity: function defaultHardDepositTimeout() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) DefaultHardDepositTimeout() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.DefaultHardDepositTimeout(&_ERC20SimpleSwap.CallOpts)
}

// DefaultHardDepositTimeout is a free data retrieval call binding the contract method 0x12101021.
//
// Solidity: function defaultHardDepositTimeout() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) DefaultHardDepositTimeout() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.DefaultHardDepositTimeout(&_ERC20SimpleSwap.CallOpts)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 timeout, uint256 canBeDecreasedAt)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) HardDeposits(opts *bind.CallOpts, arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	Timeout          *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	ret := new(struct {
		Amount           *big.Int
		DecreaseAmount   *big.Int
		Timeout          *big.Int
		CanBeDecreasedAt *big.Int
	})
	out := ret
	err := _ERC20SimpleSwap.contract.Call(opts, out, "hardDeposits", arg0)
	return *ret, err
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 timeout, uint256 canBeDecreasedAt)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	Timeout          *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _ERC20SimpleSwap.Contract.HardDeposits(&_ERC20SimpleSwap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 timeout, uint256 canBeDecreasedAt)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	Timeout          *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _ERC20SimpleSwap.Contract.HardDeposits(&_ERC20SimpleSwap.CallOpts, arg0)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) Issuer(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "issuer")
	return *ret0, err
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) Issuer() (common.Address, error) {
	return _ERC20SimpleSwap.Contract.Issuer(&_ERC20SimpleSwap.CallOpts)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) Issuer() (common.Address, error) {
	return _ERC20SimpleSwap.Contract.Issuer(&_ERC20SimpleSwap.CallOpts)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) LiquidBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "liquidBalance")
	return *ret0, err
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) LiquidBalance() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.LiquidBalance(&_ERC20SimpleSwap.CallOpts)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) LiquidBalance() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.LiquidBalance(&_ERC20SimpleSwap.CallOpts)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) LiquidBalanceFor(opts *bind.CallOpts, beneficiary common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "liquidBalanceFor", beneficiary)
	return *ret0, err
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.LiquidBalanceFor(&_ERC20SimpleSwap.CallOpts, beneficiary)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.LiquidBalanceFor(&_ERC20SimpleSwap.CallOpts, beneficiary)
}

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) PaidOut(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "paidOut", arg0)
	return *ret0, err
}

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) PaidOut(arg0 common.Address) (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.PaidOut(&_ERC20SimpleSwap.CallOpts, arg0)
}

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) PaidOut(arg0 common.Address) (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.PaidOut(&_ERC20SimpleSwap.CallOpts, arg0)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "token")
	return *ret0, err
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) Token() (common.Address, error) {
	return _ERC20SimpleSwap.Contract.Token(&_ERC20SimpleSwap.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() constant returns(address)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) Token() (common.Address, error) {
	return _ERC20SimpleSwap.Contract.Token(&_ERC20SimpleSwap.CallOpts)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) TotalHardDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "totalHardDeposit")
	return *ret0, err
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) TotalHardDeposit() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.TotalHardDeposit(&_ERC20SimpleSwap.CallOpts)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) TotalHardDeposit() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.TotalHardDeposit(&_ERC20SimpleSwap.CallOpts)
}

// TotalPaidOut is a free data retrieval call binding the contract method 0x1357e1dc.
//
// Solidity: function totalPaidOut() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCaller) TotalPaidOut(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ERC20SimpleSwap.contract.Call(opts, out, "totalPaidOut")
	return *ret0, err
}

// TotalPaidOut is a free data retrieval call binding the contract method 0x1357e1dc.
//
// Solidity: function totalPaidOut() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) TotalPaidOut() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.TotalPaidOut(&_ERC20SimpleSwap.CallOpts)
}

// TotalPaidOut is a free data retrieval call binding the contract method 0x1357e1dc.
//
// Solidity: function totalPaidOut() constant returns(uint256)
func (_ERC20SimpleSwap *ERC20SimpleSwapCallerSession) TotalPaidOut() (*big.Int, error) {
	return _ERC20SimpleSwap.Contract.TotalPaidOut(&_ERC20SimpleSwap.CallOpts)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) CashCheque(opts *bind.TransactOpts, beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "cashCheque", beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) CashCheque(beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.CashCheque(&_ERC20SimpleSwap.TransactOpts, beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) CashCheque(beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.CashCheque(&_ERC20SimpleSwap.TransactOpts, beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) CashChequeBeneficiary(opts *bind.TransactOpts, recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "cashChequeBeneficiary", recipient, cumulativePayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) CashChequeBeneficiary(recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.CashChequeBeneficiary(&_ERC20SimpleSwap.TransactOpts, recipient, cumulativePayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) CashChequeBeneficiary(recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.CashChequeBeneficiary(&_ERC20SimpleSwap.TransactOpts, recipient, cumulativePayout, issuerSig)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) DecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "decreaseHardDeposit", beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.DecreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.DecreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) IncreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "increaseHardDeposit", beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.IncreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.IncreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary, amount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) PrepareDecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "prepareDecreaseHardDeposit", beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_ERC20SimpleSwap.TransactOpts, beneficiary, decreaseAmount)
}

// SetCustomHardDepositTimeout is a paid mutator transaction binding the contract method 0xd4c9a8e8.
//
// Solidity: function setCustomHardDepositTimeout(address beneficiary, uint256 hardDepositTimeout, bytes beneficiarySig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) SetCustomHardDepositTimeout(opts *bind.TransactOpts, beneficiary common.Address, hardDepositTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "setCustomHardDepositTimeout", beneficiary, hardDepositTimeout, beneficiarySig)
}

// SetCustomHardDepositTimeout is a paid mutator transaction binding the contract method 0xd4c9a8e8.
//
// Solidity: function setCustomHardDepositTimeout(address beneficiary, uint256 hardDepositTimeout, bytes beneficiarySig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) SetCustomHardDepositTimeout(beneficiary common.Address, hardDepositTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.SetCustomHardDepositTimeout(&_ERC20SimpleSwap.TransactOpts, beneficiary, hardDepositTimeout, beneficiarySig)
}

// SetCustomHardDepositTimeout is a paid mutator transaction binding the contract method 0xd4c9a8e8.
//
// Solidity: function setCustomHardDepositTimeout(address beneficiary, uint256 hardDepositTimeout, bytes beneficiarySig) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) SetCustomHardDepositTimeout(beneficiary common.Address, hardDepositTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.SetCustomHardDepositTimeout(&_ERC20SimpleSwap.TransactOpts, beneficiary, hardDepositTimeout, beneficiarySig)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactor) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.contract.Transact(opts, "withdraw", amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.Withdraw(&_ERC20SimpleSwap.TransactOpts, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_ERC20SimpleSwap *ERC20SimpleSwapTransactorSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _ERC20SimpleSwap.Contract.Withdraw(&_ERC20SimpleSwap.TransactOpts, amount)
}

// ERC20SimpleSwapChequeBouncedIterator is returned from FilterChequeBounced and is used to iterate over the raw logs and unpacked data for ChequeBounced events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapChequeBouncedIterator struct {
	Event *ERC20SimpleSwapChequeBounced // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapChequeBouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapChequeBounced)
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
		it.Event = new(ERC20SimpleSwapChequeBounced)
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
func (it *ERC20SimpleSwapChequeBouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapChequeBouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapChequeBounced represents a ChequeBounced event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapChequeBounced struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterChequeBounced is a free log retrieval operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterChequeBounced(opts *bind.FilterOpts) (*ERC20SimpleSwapChequeBouncedIterator, error) {

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "ChequeBounced")
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapChequeBouncedIterator{contract: _ERC20SimpleSwap.contract, event: "ChequeBounced", logs: logs, sub: sub}, nil
}

// WatchChequeBounced is a free log subscription operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchChequeBounced(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapChequeBounced) (event.Subscription, error) {

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "ChequeBounced")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapChequeBounced)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
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

// ParseChequeBounced is a log parse operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseChequeBounced(log types.Log) (*ERC20SimpleSwapChequeBounced, error) {
	event := new(ERC20SimpleSwapChequeBounced)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapChequeCashedIterator is returned from FilterChequeCashed and is used to iterate over the raw logs and unpacked data for ChequeCashed events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapChequeCashedIterator struct {
	Event *ERC20SimpleSwapChequeCashed // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapChequeCashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapChequeCashed)
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
		it.Event = new(ERC20SimpleSwapChequeCashed)
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
func (it *ERC20SimpleSwapChequeCashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapChequeCashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapChequeCashed represents a ChequeCashed event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapChequeCashed struct {
	Beneficiary      common.Address
	Recipient        common.Address
	Caller           common.Address
	TotalPayout      *big.Int
	CumulativePayout *big.Int
	CallerPayout     *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterChequeCashed is a free log retrieval operation binding the contract event 0x950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f.
//
// Solidity: event ChequeCashed(address indexed beneficiary, address indexed recipient, address indexed caller, uint256 totalPayout, uint256 cumulativePayout, uint256 callerPayout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterChequeCashed(opts *bind.FilterOpts, beneficiary []common.Address, recipient []common.Address, caller []common.Address) (*ERC20SimpleSwapChequeCashedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "ChequeCashed", beneficiaryRule, recipientRule, callerRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapChequeCashedIterator{contract: _ERC20SimpleSwap.contract, event: "ChequeCashed", logs: logs, sub: sub}, nil
}

// WatchChequeCashed is a free log subscription operation binding the contract event 0x950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f.
//
// Solidity: event ChequeCashed(address indexed beneficiary, address indexed recipient, address indexed caller, uint256 totalPayout, uint256 cumulativePayout, uint256 callerPayout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchChequeCashed(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapChequeCashed, beneficiary []common.Address, recipient []common.Address, caller []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "ChequeCashed", beneficiaryRule, recipientRule, callerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapChequeCashed)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
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

// ParseChequeCashed is a log parse operation binding the contract event 0x950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f.
//
// Solidity: event ChequeCashed(address indexed beneficiary, address indexed recipient, address indexed caller, uint256 totalPayout, uint256 cumulativePayout, uint256 callerPayout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseChequeCashed(log types.Log) (*ERC20SimpleSwapChequeCashed, error) {
	event := new(ERC20SimpleSwapChequeCashed)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapHardDepositAmountChangedIterator is returned from FilterHardDepositAmountChanged and is used to iterate over the raw logs and unpacked data for HardDepositAmountChanged events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositAmountChangedIterator struct {
	Event *ERC20SimpleSwapHardDepositAmountChanged // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapHardDepositAmountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapHardDepositAmountChanged)
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
		it.Event = new(ERC20SimpleSwapHardDepositAmountChanged)
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
func (it *ERC20SimpleSwapHardDepositAmountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapHardDepositAmountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapHardDepositAmountChanged represents a HardDepositAmountChanged event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositAmountChanged struct {
	Beneficiary common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositAmountChanged is a free log retrieval operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterHardDepositAmountChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*ERC20SimpleSwapHardDepositAmountChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapHardDepositAmountChangedIterator{contract: _ERC20SimpleSwap.contract, event: "HardDepositAmountChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositAmountChanged is a free log subscription operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchHardDepositAmountChanged(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapHardDepositAmountChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapHardDepositAmountChanged)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
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

// ParseHardDepositAmountChanged is a log parse operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseHardDepositAmountChanged(log types.Log) (*ERC20SimpleSwapHardDepositAmountChanged, error) {
	event := new(ERC20SimpleSwapHardDepositAmountChanged)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapHardDepositDecreasePreparedIterator is returned from FilterHardDepositDecreasePrepared and is used to iterate over the raw logs and unpacked data for HardDepositDecreasePrepared events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositDecreasePreparedIterator struct {
	Event *ERC20SimpleSwapHardDepositDecreasePrepared // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapHardDepositDecreasePreparedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapHardDepositDecreasePrepared)
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
		it.Event = new(ERC20SimpleSwapHardDepositDecreasePrepared)
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
func (it *ERC20SimpleSwapHardDepositDecreasePreparedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapHardDepositDecreasePreparedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapHardDepositDecreasePrepared represents a HardDepositDecreasePrepared event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositDecreasePrepared struct {
	Beneficiary    common.Address
	DecreaseAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreasePrepared is a free log retrieval operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterHardDepositDecreasePrepared(opts *bind.FilterOpts, beneficiary []common.Address) (*ERC20SimpleSwapHardDepositDecreasePreparedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapHardDepositDecreasePreparedIterator{contract: _ERC20SimpleSwap.contract, event: "HardDepositDecreasePrepared", logs: logs, sub: sub}, nil
}

// WatchHardDepositDecreasePrepared is a free log subscription operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchHardDepositDecreasePrepared(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapHardDepositDecreasePrepared, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapHardDepositDecreasePrepared)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
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

// ParseHardDepositDecreasePrepared is a log parse operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseHardDepositDecreasePrepared(log types.Log) (*ERC20SimpleSwapHardDepositDecreasePrepared, error) {
	event := new(ERC20SimpleSwapHardDepositDecreasePrepared)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapHardDepositTimeoutChangedIterator is returned from FilterHardDepositTimeoutChanged and is used to iterate over the raw logs and unpacked data for HardDepositTimeoutChanged events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositTimeoutChangedIterator struct {
	Event *ERC20SimpleSwapHardDepositTimeoutChanged // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapHardDepositTimeoutChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapHardDepositTimeoutChanged)
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
		it.Event = new(ERC20SimpleSwapHardDepositTimeoutChanged)
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
func (it *ERC20SimpleSwapHardDepositTimeoutChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapHardDepositTimeoutChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapHardDepositTimeoutChanged represents a HardDepositTimeoutChanged event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapHardDepositTimeoutChanged struct {
	Beneficiary common.Address
	Timeout     *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositTimeoutChanged is a free log retrieval operation binding the contract event 0x7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf0.
//
// Solidity: event HardDepositTimeoutChanged(address indexed beneficiary, uint256 timeout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterHardDepositTimeoutChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*ERC20SimpleSwapHardDepositTimeoutChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "HardDepositTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapHardDepositTimeoutChangedIterator{contract: _ERC20SimpleSwap.contract, event: "HardDepositTimeoutChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositTimeoutChanged is a free log subscription operation binding the contract event 0x7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf0.
//
// Solidity: event HardDepositTimeoutChanged(address indexed beneficiary, uint256 timeout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchHardDepositTimeoutChanged(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapHardDepositTimeoutChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "HardDepositTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapHardDepositTimeoutChanged)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositTimeoutChanged", log); err != nil {
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

// ParseHardDepositTimeoutChanged is a log parse operation binding the contract event 0x7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf0.
//
// Solidity: event HardDepositTimeoutChanged(address indexed beneficiary, uint256 timeout)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseHardDepositTimeoutChanged(log types.Log) (*ERC20SimpleSwapHardDepositTimeoutChanged, error) {
	event := new(ERC20SimpleSwapHardDepositTimeoutChanged)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "HardDepositTimeoutChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ERC20SimpleSwapWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapWithdrawIterator struct {
	Event *ERC20SimpleSwapWithdraw // Event containing the contract specifics and raw log

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
func (it *ERC20SimpleSwapWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ERC20SimpleSwapWithdraw)
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
		it.Event = new(ERC20SimpleSwapWithdraw)
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
func (it *ERC20SimpleSwapWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ERC20SimpleSwapWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ERC20SimpleSwapWithdraw represents a Withdraw event raised by the ERC20SimpleSwap contract.
type ERC20SimpleSwapWithdraw struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d.
//
// Solidity: event Withdraw(uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) FilterWithdraw(opts *bind.FilterOpts) (*ERC20SimpleSwapWithdrawIterator, error) {

	logs, sub, err := _ERC20SimpleSwap.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &ERC20SimpleSwapWithdrawIterator{contract: _ERC20SimpleSwap.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d.
//
// Solidity: event Withdraw(uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *ERC20SimpleSwapWithdraw) (event.Subscription, error) {

	logs, sub, err := _ERC20SimpleSwap.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ERC20SimpleSwapWithdraw)
				if err := _ERC20SimpleSwap.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d.
//
// Solidity: event Withdraw(uint256 amount)
func (_ERC20SimpleSwap *ERC20SimpleSwapFilterer) ParseWithdraw(log types.Log) (*ERC20SimpleSwapWithdraw, error) {
	event := new(ERC20SimpleSwapWithdraw)
	if err := _ERC20SimpleSwap.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IERC20ABI is the input ABI used to generate the binding from.
const IERC20ABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// IERC20 is an auto generated Go binding around an Ethereum contract.
type IERC20 struct {
	IERC20Caller     // Read-only binding to the contract
	IERC20Transactor // Write-only binding to the contract
	IERC20Filterer   // Log filterer for contract events
}

// IERC20Caller is an auto generated read-only Go binding around an Ethereum contract.
type IERC20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type IERC20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IERC20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IERC20Session struct {
	Contract     *IERC20           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IERC20CallerSession struct {
	Contract *IERC20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// IERC20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IERC20TransactorSession struct {
	Contract     *IERC20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20Raw is an auto generated low-level Go binding around an Ethereum contract.
type IERC20Raw struct {
	Contract *IERC20 // Generic contract binding to access the raw methods on
}

// IERC20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IERC20CallerRaw struct {
	Contract *IERC20Caller // Generic read-only contract binding to access the raw methods on
}

// IERC20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IERC20TransactorRaw struct {
	Contract *IERC20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewIERC20 creates a new instance of IERC20, bound to a specific deployed contract.
func NewIERC20(address common.Address, backend bind.ContractBackend) (*IERC20, error) {
	contract, err := bindIERC20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IERC20{IERC20Caller: IERC20Caller{contract: contract}, IERC20Transactor: IERC20Transactor{contract: contract}, IERC20Filterer: IERC20Filterer{contract: contract}}, nil
}

// NewIERC20Caller creates a new read-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Caller(address common.Address, caller bind.ContractCaller) (*IERC20Caller, error) {
	contract, err := bindIERC20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Caller{contract: contract}, nil
}

// NewIERC20Transactor creates a new write-only instance of IERC20, bound to a specific deployed contract.
func NewIERC20Transactor(address common.Address, transactor bind.ContractTransactor) (*IERC20Transactor, error) {
	contract, err := bindIERC20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20Transactor{contract: contract}, nil
}

// NewIERC20Filterer creates a new log filterer instance of IERC20, bound to a specific deployed contract.
func NewIERC20Filterer(address common.Address, filterer bind.ContractFilterer) (*IERC20Filterer, error) {
	contract, err := bindIERC20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IERC20Filterer{contract: contract}, nil
}

// bindIERC20 binds a generic wrapper to an already deployed contract.
func bindIERC20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IERC20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20Raw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.IERC20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.IERC20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20 *IERC20CallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _IERC20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20 *IERC20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20 *IERC20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_IERC20 *IERC20Caller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "allowance", owner, spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_IERC20 *IERC20Session) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) constant returns(uint256)
func (_IERC20 *IERC20CallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _IERC20.Contract.Allowance(&_IERC20.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_IERC20 *IERC20Caller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "balanceOf", account)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_IERC20 *IERC20Session) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) constant returns(uint256)
func (_IERC20 *IERC20CallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20.Contract.BalanceOf(&_IERC20.CallOpts, account)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20Caller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _IERC20.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20Session) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_IERC20 *IERC20CallerSession) TotalSupply() (*big.Int, error) {
	return _IERC20.Contract.TotalSupply(&_IERC20.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Approve(&_IERC20.TransactOpts, spender, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) Transfer(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transfer", recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, recipient, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) Transfer(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.Transfer(&_IERC20.TransactOpts, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Transactor) TransferFrom(opts *bind.TransactOpts, sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.contract.Transact(opts, "transferFrom", sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20Session) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, sender, recipient, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address sender, address recipient, uint256 amount) returns(bool)
func (_IERC20 *IERC20TransactorSession) TransferFrom(sender common.Address, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _IERC20.Contract.TransferFrom(&_IERC20.TransactOpts, sender, recipient, amount)
}

// IERC20ApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the IERC20 contract.
type IERC20ApprovalIterator struct {
	Event *IERC20Approval // Event containing the contract specifics and raw log

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
func (it *IERC20ApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Approval)
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
		it.Event = new(IERC20Approval)
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
func (it *IERC20ApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20ApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Approval represents a Approval event raised by the IERC20 contract.
type IERC20Approval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*IERC20ApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &IERC20ApprovalIterator{contract: _IERC20.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *IERC20Approval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Approval)
				if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_IERC20 *IERC20Filterer) ParseApproval(log types.Log) (*IERC20Approval, error) {
	event := new(IERC20Approval)
	if err := _IERC20.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IERC20TransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the IERC20 contract.
type IERC20TransferIterator struct {
	Event *IERC20Transfer // Event containing the contract specifics and raw log

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
func (it *IERC20TransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IERC20Transfer)
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
		it.Event = new(IERC20Transfer)
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
func (it *IERC20TransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IERC20TransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IERC20Transfer represents a Transfer event raised by the IERC20 contract.
type IERC20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*IERC20TransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &IERC20TransferIterator{contract: _IERC20.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *IERC20Transfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _IERC20.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IERC20Transfer)
				if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_IERC20 *IERC20Filterer) ParseTransfer(log types.Log) (*IERC20Transfer, error) {
	event := new(IERC20Transfer)
	if err := _IERC20.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	return event, nil
}

// MathABI is the input ABI used to generate the binding from.
const MathABI = "[]"

// MathBin is the compiled bytecode used for deploying new contracts.
var MathBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a7231582000a7fe8f11c52b79f03af2d1ee3eb200d3fd399e2d019af8439d5bba62f64c0964736f6c634300050d0032"

// DeployMath deploys a new Ethereum contract, binding an instance of Math to it.
func DeployMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Math, error) {
	parsed, err := abi.JSON(strings.NewReader(MathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Math{MathCaller: MathCaller{contract: contract}, MathTransactor: MathTransactor{contract: contract}, MathFilterer: MathFilterer{contract: contract}}, nil
}

// Math is an auto generated Go binding around an Ethereum contract.
type Math struct {
	MathCaller     // Read-only binding to the contract
	MathTransactor // Write-only binding to the contract
	MathFilterer   // Log filterer for contract events
}

// MathCaller is an auto generated read-only Go binding around an Ethereum contract.
type MathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MathSession struct {
	Contract     *Math             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MathCallerSession struct {
	Contract *MathCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// MathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MathTransactorSession struct {
	Contract     *MathTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MathRaw is an auto generated low-level Go binding around an Ethereum contract.
type MathRaw struct {
	Contract *Math // Generic contract binding to access the raw methods on
}

// MathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MathCallerRaw struct {
	Contract *MathCaller // Generic read-only contract binding to access the raw methods on
}

// MathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MathTransactorRaw struct {
	Contract *MathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMath creates a new instance of Math, bound to a specific deployed contract.
func NewMath(address common.Address, backend bind.ContractBackend) (*Math, error) {
	contract, err := bindMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Math{MathCaller: MathCaller{contract: contract}, MathTransactor: MathTransactor{contract: contract}, MathFilterer: MathFilterer{contract: contract}}, nil
}

// NewMathCaller creates a new read-only instance of Math, bound to a specific deployed contract.
func NewMathCaller(address common.Address, caller bind.ContractCaller) (*MathCaller, error) {
	contract, err := bindMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MathCaller{contract: contract}, nil
}

// NewMathTransactor creates a new write-only instance of Math, bound to a specific deployed contract.
func NewMathTransactor(address common.Address, transactor bind.ContractTransactor) (*MathTransactor, error) {
	contract, err := bindMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MathTransactor{contract: contract}, nil
}

// NewMathFilterer creates a new log filterer instance of Math, bound to a specific deployed contract.
func NewMathFilterer(address common.Address, filterer bind.ContractFilterer) (*MathFilterer, error) {
	contract, err := bindMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MathFilterer{contract: contract}, nil
}

// bindMath binds a generic wrapper to an already deployed contract.
func bindMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Math *MathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Math.Contract.MathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Math *MathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Math.Contract.MathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Math *MathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Math.Contract.MathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Math *MathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Math.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Math *MathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Math.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Math *MathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Math.Contract.contract.Transact(opts, method, params...)
}

// MinterRoleABI is the input ABI used to generate the binding from.
const MinterRoleABI = "[{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"MinterAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"MinterRemoved\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"addMinter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"isMinter\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceMinter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// MinterRole is an auto generated Go binding around an Ethereum contract.
type MinterRole struct {
	MinterRoleCaller     // Read-only binding to the contract
	MinterRoleTransactor // Write-only binding to the contract
	MinterRoleFilterer   // Log filterer for contract events
}

// MinterRoleCaller is an auto generated read-only Go binding around an Ethereum contract.
type MinterRoleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterRoleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MinterRoleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterRoleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MinterRoleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MinterRoleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MinterRoleSession struct {
	Contract     *MinterRole       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MinterRoleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MinterRoleCallerSession struct {
	Contract *MinterRoleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MinterRoleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MinterRoleTransactorSession struct {
	Contract     *MinterRoleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MinterRoleRaw is an auto generated low-level Go binding around an Ethereum contract.
type MinterRoleRaw struct {
	Contract *MinterRole // Generic contract binding to access the raw methods on
}

// MinterRoleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MinterRoleCallerRaw struct {
	Contract *MinterRoleCaller // Generic read-only contract binding to access the raw methods on
}

// MinterRoleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MinterRoleTransactorRaw struct {
	Contract *MinterRoleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMinterRole creates a new instance of MinterRole, bound to a specific deployed contract.
func NewMinterRole(address common.Address, backend bind.ContractBackend) (*MinterRole, error) {
	contract, err := bindMinterRole(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MinterRole{MinterRoleCaller: MinterRoleCaller{contract: contract}, MinterRoleTransactor: MinterRoleTransactor{contract: contract}, MinterRoleFilterer: MinterRoleFilterer{contract: contract}}, nil
}

// NewMinterRoleCaller creates a new read-only instance of MinterRole, bound to a specific deployed contract.
func NewMinterRoleCaller(address common.Address, caller bind.ContractCaller) (*MinterRoleCaller, error) {
	contract, err := bindMinterRole(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MinterRoleCaller{contract: contract}, nil
}

// NewMinterRoleTransactor creates a new write-only instance of MinterRole, bound to a specific deployed contract.
func NewMinterRoleTransactor(address common.Address, transactor bind.ContractTransactor) (*MinterRoleTransactor, error) {
	contract, err := bindMinterRole(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MinterRoleTransactor{contract: contract}, nil
}

// NewMinterRoleFilterer creates a new log filterer instance of MinterRole, bound to a specific deployed contract.
func NewMinterRoleFilterer(address common.Address, filterer bind.ContractFilterer) (*MinterRoleFilterer, error) {
	contract, err := bindMinterRole(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MinterRoleFilterer{contract: contract}, nil
}

// bindMinterRole binds a generic wrapper to an already deployed contract.
func bindMinterRole(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MinterRoleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MinterRole *MinterRoleRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MinterRole.Contract.MinterRoleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MinterRole *MinterRoleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinterRole.Contract.MinterRoleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MinterRole *MinterRoleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MinterRole.Contract.MinterRoleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MinterRole *MinterRoleCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MinterRole.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MinterRole *MinterRoleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinterRole.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MinterRole *MinterRoleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MinterRole.Contract.contract.Transact(opts, method, params...)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_MinterRole *MinterRoleCaller) IsMinter(opts *bind.CallOpts, account common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MinterRole.contract.Call(opts, out, "isMinter", account)
	return *ret0, err
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_MinterRole *MinterRoleSession) IsMinter(account common.Address) (bool, error) {
	return _MinterRole.Contract.IsMinter(&_MinterRole.CallOpts, account)
}

// IsMinter is a free data retrieval call binding the contract method 0xaa271e1a.
//
// Solidity: function isMinter(address account) constant returns(bool)
func (_MinterRole *MinterRoleCallerSession) IsMinter(account common.Address) (bool, error) {
	return _MinterRole.Contract.IsMinter(&_MinterRole.CallOpts, account)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_MinterRole *MinterRoleTransactor) AddMinter(opts *bind.TransactOpts, account common.Address) (*types.Transaction, error) {
	return _MinterRole.contract.Transact(opts, "addMinter", account)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_MinterRole *MinterRoleSession) AddMinter(account common.Address) (*types.Transaction, error) {
	return _MinterRole.Contract.AddMinter(&_MinterRole.TransactOpts, account)
}

// AddMinter is a paid mutator transaction binding the contract method 0x983b2d56.
//
// Solidity: function addMinter(address account) returns()
func (_MinterRole *MinterRoleTransactorSession) AddMinter(account common.Address) (*types.Transaction, error) {
	return _MinterRole.Contract.AddMinter(&_MinterRole.TransactOpts, account)
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_MinterRole *MinterRoleTransactor) RenounceMinter(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MinterRole.contract.Transact(opts, "renounceMinter")
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_MinterRole *MinterRoleSession) RenounceMinter() (*types.Transaction, error) {
	return _MinterRole.Contract.RenounceMinter(&_MinterRole.TransactOpts)
}

// RenounceMinter is a paid mutator transaction binding the contract method 0x98650275.
//
// Solidity: function renounceMinter() returns()
func (_MinterRole *MinterRoleTransactorSession) RenounceMinter() (*types.Transaction, error) {
	return _MinterRole.Contract.RenounceMinter(&_MinterRole.TransactOpts)
}

// MinterRoleMinterAddedIterator is returned from FilterMinterAdded and is used to iterate over the raw logs and unpacked data for MinterAdded events raised by the MinterRole contract.
type MinterRoleMinterAddedIterator struct {
	Event *MinterRoleMinterAdded // Event containing the contract specifics and raw log

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
func (it *MinterRoleMinterAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MinterRoleMinterAdded)
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
		it.Event = new(MinterRoleMinterAdded)
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
func (it *MinterRoleMinterAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MinterRoleMinterAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MinterRoleMinterAdded represents a MinterAdded event raised by the MinterRole contract.
type MinterRoleMinterAdded struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMinterAdded is a free log retrieval operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_MinterRole *MinterRoleFilterer) FilterMinterAdded(opts *bind.FilterOpts, account []common.Address) (*MinterRoleMinterAddedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _MinterRole.contract.FilterLogs(opts, "MinterAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return &MinterRoleMinterAddedIterator{contract: _MinterRole.contract, event: "MinterAdded", logs: logs, sub: sub}, nil
}

// WatchMinterAdded is a free log subscription operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_MinterRole *MinterRoleFilterer) WatchMinterAdded(opts *bind.WatchOpts, sink chan<- *MinterRoleMinterAdded, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _MinterRole.contract.WatchLogs(opts, "MinterAdded", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MinterRoleMinterAdded)
				if err := _MinterRole.contract.UnpackLog(event, "MinterAdded", log); err != nil {
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

// ParseMinterAdded is a log parse operation binding the contract event 0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6.
//
// Solidity: event MinterAdded(address indexed account)
func (_MinterRole *MinterRoleFilterer) ParseMinterAdded(log types.Log) (*MinterRoleMinterAdded, error) {
	event := new(MinterRoleMinterAdded)
	if err := _MinterRole.contract.UnpackLog(event, "MinterAdded", log); err != nil {
		return nil, err
	}
	return event, nil
}

// MinterRoleMinterRemovedIterator is returned from FilterMinterRemoved and is used to iterate over the raw logs and unpacked data for MinterRemoved events raised by the MinterRole contract.
type MinterRoleMinterRemovedIterator struct {
	Event *MinterRoleMinterRemoved // Event containing the contract specifics and raw log

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
func (it *MinterRoleMinterRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MinterRoleMinterRemoved)
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
		it.Event = new(MinterRoleMinterRemoved)
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
func (it *MinterRoleMinterRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MinterRoleMinterRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MinterRoleMinterRemoved represents a MinterRemoved event raised by the MinterRole contract.
type MinterRoleMinterRemoved struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMinterRemoved is a free log retrieval operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_MinterRole *MinterRoleFilterer) FilterMinterRemoved(opts *bind.FilterOpts, account []common.Address) (*MinterRoleMinterRemovedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _MinterRole.contract.FilterLogs(opts, "MinterRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return &MinterRoleMinterRemovedIterator{contract: _MinterRole.contract, event: "MinterRemoved", logs: logs, sub: sub}, nil
}

// WatchMinterRemoved is a free log subscription operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_MinterRole *MinterRoleFilterer) WatchMinterRemoved(opts *bind.WatchOpts, sink chan<- *MinterRoleMinterRemoved, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _MinterRole.contract.WatchLogs(opts, "MinterRemoved", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MinterRoleMinterRemoved)
				if err := _MinterRole.contract.UnpackLog(event, "MinterRemoved", log); err != nil {
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

// ParseMinterRemoved is a log parse operation binding the contract event 0xe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692.
//
// Solidity: event MinterRemoved(address indexed account)
func (_MinterRole *MinterRoleFilterer) ParseMinterRemoved(log types.Log) (*MinterRoleMinterRemoved, error) {
	event := new(MinterRoleMinterRemoved)
	if err := _MinterRole.contract.UnpackLog(event, "MinterRemoved", log); err != nil {
		return nil, err
	}
	return event, nil
}

// RolesABI is the input ABI used to generate the binding from.
const RolesABI = "[]"

// RolesBin is the compiled bytecode used for deploying new contracts.
var RolesBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a723158209a68d01e8b2b237a89fef9b2aa16f6200a28b6f0cb4b41af99c6e125cf59ef1e64736f6c634300050d0032"

// DeployRoles deploys a new Ethereum contract, binding an instance of Roles to it.
func DeployRoles(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Roles, error) {
	parsed, err := abi.JSON(strings.NewReader(RolesABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(RolesBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Roles{RolesCaller: RolesCaller{contract: contract}, RolesTransactor: RolesTransactor{contract: contract}, RolesFilterer: RolesFilterer{contract: contract}}, nil
}

// Roles is an auto generated Go binding around an Ethereum contract.
type Roles struct {
	RolesCaller     // Read-only binding to the contract
	RolesTransactor // Write-only binding to the contract
	RolesFilterer   // Log filterer for contract events
}

// RolesCaller is an auto generated read-only Go binding around an Ethereum contract.
type RolesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RolesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RolesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RolesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RolesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RolesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RolesSession struct {
	Contract     *Roles            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RolesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RolesCallerSession struct {
	Contract *RolesCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RolesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RolesTransactorSession struct {
	Contract     *RolesTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RolesRaw is an auto generated low-level Go binding around an Ethereum contract.
type RolesRaw struct {
	Contract *Roles // Generic contract binding to access the raw methods on
}

// RolesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RolesCallerRaw struct {
	Contract *RolesCaller // Generic read-only contract binding to access the raw methods on
}

// RolesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RolesTransactorRaw struct {
	Contract *RolesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRoles creates a new instance of Roles, bound to a specific deployed contract.
func NewRoles(address common.Address, backend bind.ContractBackend) (*Roles, error) {
	contract, err := bindRoles(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Roles{RolesCaller: RolesCaller{contract: contract}, RolesTransactor: RolesTransactor{contract: contract}, RolesFilterer: RolesFilterer{contract: contract}}, nil
}

// NewRolesCaller creates a new read-only instance of Roles, bound to a specific deployed contract.
func NewRolesCaller(address common.Address, caller bind.ContractCaller) (*RolesCaller, error) {
	contract, err := bindRoles(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RolesCaller{contract: contract}, nil
}

// NewRolesTransactor creates a new write-only instance of Roles, bound to a specific deployed contract.
func NewRolesTransactor(address common.Address, transactor bind.ContractTransactor) (*RolesTransactor, error) {
	contract, err := bindRoles(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RolesTransactor{contract: contract}, nil
}

// NewRolesFilterer creates a new log filterer instance of Roles, bound to a specific deployed contract.
func NewRolesFilterer(address common.Address, filterer bind.ContractFilterer) (*RolesFilterer, error) {
	contract, err := bindRoles(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RolesFilterer{contract: contract}, nil
}

// bindRoles binds a generic wrapper to an already deployed contract.
func bindRoles(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RolesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Roles *RolesRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Roles.Contract.RolesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Roles *RolesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Roles.Contract.RolesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Roles *RolesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Roles.Contract.RolesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Roles *RolesCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Roles.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Roles *RolesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Roles.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Roles *RolesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Roles.Contract.contract.Transact(opts, method, params...)
}

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBin is the compiled bytecode used for deploying new contracts.
var SafeMathBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820044a15379c034d2be95cff236f2765160889caff1b82a6d65660cc33f8505af664736f6c634300050d0032"

// DeploySafeMath deploys a new Ethereum contract, binding an instance of SafeMath to it.
func DeploySafeMath(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SafeMath, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SafeMathBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// SafeMath is an auto generated Go binding around an Ethereum contract.
type SafeMath struct {
	SafeMathCaller     // Read-only binding to the contract
	SafeMathTransactor // Write-only binding to the contract
	SafeMathFilterer   // Log filterer for contract events
}

// SafeMathCaller is an auto generated read-only Go binding around an Ethereum contract.
type SafeMathCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SafeMathTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SafeMathFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SafeMathSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SafeMathSession struct {
	Contract     *SafeMath         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SafeMathCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SafeMathCallerSession struct {
	Contract *SafeMathCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// SafeMathTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SafeMathTransactorSession struct {
	Contract     *SafeMathTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// SafeMathRaw is an auto generated low-level Go binding around an Ethereum contract.
type SafeMathRaw struct {
	Contract *SafeMath // Generic contract binding to access the raw methods on
}

// SafeMathCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SafeMathCallerRaw struct {
	Contract *SafeMathCaller // Generic read-only contract binding to access the raw methods on
}

// SafeMathTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SafeMathTransactorRaw struct {
	Contract *SafeMathTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSafeMath creates a new instance of SafeMath, bound to a specific deployed contract.
func NewSafeMath(address common.Address, backend bind.ContractBackend) (*SafeMath, error) {
	contract, err := bindSafeMath(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SafeMath{SafeMathCaller: SafeMathCaller{contract: contract}, SafeMathTransactor: SafeMathTransactor{contract: contract}, SafeMathFilterer: SafeMathFilterer{contract: contract}}, nil
}

// NewSafeMathCaller creates a new read-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathCaller(address common.Address, caller bind.ContractCaller) (*SafeMathCaller, error) {
	contract, err := bindSafeMath(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathCaller{contract: contract}, nil
}

// NewSafeMathTransactor creates a new write-only instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathTransactor(address common.Address, transactor bind.ContractTransactor) (*SafeMathTransactor, error) {
	contract, err := bindSafeMath(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SafeMathTransactor{contract: contract}, nil
}

// NewSafeMathFilterer creates a new log filterer instance of SafeMath, bound to a specific deployed contract.
func NewSafeMathFilterer(address common.Address, filterer bind.ContractFilterer) (*SafeMathFilterer, error) {
	contract, err := bindSafeMath(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SafeMathFilterer{contract: contract}, nil
}

// bindSafeMath binds a generic wrapper to an already deployed contract.
func bindSafeMath(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SafeMathABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.SafeMathCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.SafeMathTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SafeMath *SafeMathCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SafeMath.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SafeMath *SafeMathTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SafeMath *SafeMathTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SafeMath.Contract.contract.Transact(opts, method, params...)
}

// SimpleSwapFactoryABI is the input ABI used to generate the binding from.
const SimpleSwapFactoryABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ERC20Address\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddress\",\"type\":\"address\"}],\"name\":\"SimpleSwapDeployed\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"ERC20Address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"issuer\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"defaultHardDepositTimeoutDuration\",\"type\":\"uint256\"}],\"name\":\"deploySimpleSwap\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"deployedContracts\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SimpleSwapFactoryBin is the compiled bytecode used for deploying new contracts.
var SimpleSwapFactoryBin = "0x608060405234801561001057600080fd5b506040516127093803806127098339818101604052602081101561003357600080fd5b810190808051906020019092919050505080600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050612674806100956000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c8063576d727114610046578063a6021ace146100d4578063c70242ad1461011e575b600080fd5b6100926004803603604081101561005c57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061017a565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100dc610301565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6101606004803603602081101561013457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610327565b604051808215151515815260200191505060405180910390f35b60008083600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16846040516101ae90610347565b808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019350505050604051809103906000f08015801561023a573d6000803e3d6000fd5b50905060016000808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055507fc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c00481604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a18091505092915050565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60006020528060005260406000206000915054906101000a900460ff1681565b6122eb806103558339019056fe608060405234801561001057600080fd5b506040516122eb3803806122eb8339818101604052606081101561003357600080fd5b8101908080519060200190929190805190602001909291908051906020019092919050505082600660006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555081600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550806000819055505050506121f8806100f36000396000f3fe608060405234801561001057600080fd5b506004361061010b5760003560e01c8063946f46a2116100a2578063b7ec1a3311610071578063b7ec1a3314610612578063c76a4d3114610630578063d4c9a8e814610688578063e0bcf13a1461076d578063fc0c546a1461078b5761010b565b8063946f46a2146104f5578063b6343b0d14610539578063b69ef8a8146105a6578063b7770350146105c45761010b565b80631d143848116100de5780631d143848146103d75780632e1a7d4d14610421578063338f3fed1461044f57806381f03fcb1461049d5761010b565b80630d5f26591461011057806312101021146101f55780631357e1dc146102135780631633fb1d14610231575b600080fd5b6101f36004803603606081101561012657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019064010000000081111561016d57600080fd5b82018360208201111561017f57600080fd5b803590602001918460018302840111640100000000831117156101a157600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506107d5565b005b6101fd6107e8565b6040518082815260200191505060405180910390f35b61021b6107ee565b6040518082815260200191505060405180910390f35b6103d5600480360360c081101561024757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156102ae57600080fd5b8201836020820111156102c057600080fd5b803590602001918460018302840111640100000000831117156102e257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001909291908035906020019064010000000081111561034f57600080fd5b82018360208201111561036157600080fd5b8035906020019184600183028401116401000000008311171561038357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506107f4565b005b6103df6108a2565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61044d6004803603602081101561043757600080fd5b81019080803590602001909291905050506108c8565b005b61049b6004803603604081101561046557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610b49565b005b6104df600480360360208110156104b357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d5e565b6040518082815260200191505060405180910390f35b6105376004803603602081101561050b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d76565b005b61057b6004803603602081101561054f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610ec9565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b6105ae610ef9565b6040518082815260200191505060405180910390f35b610610600480360360408110156105da57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610fda565b005b61061a6111c2565b6040518082815260200191505060405180910390f35b6106726004803603602081101561064657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506111e5565b6040518082815260200191505060405180910390f35b61076b6004803603606081101561069e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156106e557600080fd5b8201836020820111156106f757600080fd5b8035906020019184600183028401116401000000008311171561071957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061124a565b005b61077561143e565b6040518082815260200191505060405180910390f35b610793611444565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6107e333848460008561146a565b505050565b60005481565b60035481565b61080a6108043033878987611b8e565b84611c6f565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161461088d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806121076022913960400191505060405180910390fd5b61089a868686858561146a565b505050505050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461098b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6109936111c2565b8111156109eb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602881526020018061219c6028913960400191505060405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16836040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b158015610ab657600080fd5b505af1158015610aca573d6000803e3d6000fd5b505050506040513d6020811015610ae057600080fd5b8101908080519060200190929190505050610b46576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b50565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c0c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610c14610ef9565b610c2982600554611c8b90919063ffffffff16565b1115610c80576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260348152602001806120d36034913960400191505060405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050610cda828260000154611c8b90919063ffffffff16565b8160000181905550610cf782600554611c8b90919063ffffffff16565b600581905550600081600301819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60026020528060005260406000206000915090505481565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600301544210158015610dd257506000816003015414155b610e27576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260258152602001806121296025913960400191505060405180910390fd5b610e4281600101548260000154611d1390919063ffffffff16565b816000018190555060008160030181905550610e6d8160010154600554611d1390919063ffffffff16565b6005819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60046020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b6000600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b158015610f9a57600080fd5b505afa158015610fae573d6000803e3d6000fd5b505050506040513d6020811015610fc457600080fd5b8101908080519060200190929190505050905090565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461109d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050806000015482111561113d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260278152602001806121756027913960400191505060405180910390fd5b600080826002015414611154578160020154611158565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b60006111e06005546111d2610ef9565b611d1390919063ffffffff16565b905090565b6000611243600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546112356111c2565b611c8b90919063ffffffff16565b9050919050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461130d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b61132161131b308585611d5d565b82611c6f565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16146113a4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806121076022913960400191505060405180910390fd5b81600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf0836040518082815260200191505060405180910390a2505050565b60055481565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611596576114d36114cd308786611dfd565b82611c6f565b73ffffffffffffffffffffffffffffffffffffffff16600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611595576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b5b60006115ea600260008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205485611d1390919063ffffffff16565b90506000611600826115fb896111e5565b611e9d565b9050600061165082600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611e9d565b9050848210156116c8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a2063616e6e6f74207061792063616c6c657200000081525060200191505060405180910390fd5b600081146117875761172581600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611d1390919063ffffffff16565b600460008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000018190555061178081600554611d1390919063ffffffff16565b6005819055505b6117d982600260008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054611c8b90919063ffffffff16565b600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555061183182600354611c8b90919063ffffffff16565b600381905550600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb886118898886611d1390919063ffffffff16565b6040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b1580156118f257600080fd5b505af1158015611906573d6000803e3d6000fd5b505050506040513d602081101561191c57600080fd5b8101908080519060200190929190505050611982576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b60008514611ac457600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb33876040518363ffffffff1660e01b8152600401808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200192505050602060405180830381600087803b158015611a3357600080fd5b505af1158015611a47573d6000803e3d6000fd5b505050506040513d6020811015611a5d57600080fd5b8101908080519060200190929190505050611ac3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602781526020018061214e6027913960400191505060405180910390fd5b5b3373ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f858a8a60405180848152602001838152602001828152602001935050505060405180910390a4818314611b84577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b5050505050505050565b60008585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b6000611c83611c7d84611eb6565b83611f0e565b905092915050565b600080828401905083811015611d09576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b6000611d5583836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250612012565b905092915050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000818310611eac5781611eae565b825b905092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b60006041825114611f22576000905061200c565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c1115611f76576000935050505061200c565b601b8160ff1614158015611f8e5750601c8160ff1614155b15611f9f576000935050505061200c565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611ffc573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b60008383111582906120bf576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b83811015612084578082015181840152602081019050612069565b50505050905090810190601f1680156120b15780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b506000838503905080915050939250505056fe53696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e636553696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a2053696d706c65537761703a207472616e73666572206661696c656453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e74a265627a7a72315820a2829ec1c4ee1a4aa11d90f04bfe854a020b9b07943b7ae56603fb1d301604be64736f6c634300050d0032a265627a7a72315820bfdcb3c4bb558adb9e8c3072b91d763f9e5b62aad72502d95e326660aaa3cbb364736f6c634300050d0032"

// DeploySimpleSwapFactory deploys a new Ethereum contract, binding an instance of SimpleSwapFactory to it.
func DeploySimpleSwapFactory(auth *bind.TransactOpts, backend bind.ContractBackend, _ERC20Address common.Address) (common.Address, *types.Transaction, *SimpleSwapFactory, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSwapFactoryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleSwapFactoryBin), backend, _ERC20Address)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleSwapFactory{SimpleSwapFactoryCaller: SimpleSwapFactoryCaller{contract: contract}, SimpleSwapFactoryTransactor: SimpleSwapFactoryTransactor{contract: contract}, SimpleSwapFactoryFilterer: SimpleSwapFactoryFilterer{contract: contract}}, nil
}

// SimpleSwapFactory is an auto generated Go binding around an Ethereum contract.
type SimpleSwapFactory struct {
	SimpleSwapFactoryCaller     // Read-only binding to the contract
	SimpleSwapFactoryTransactor // Write-only binding to the contract
	SimpleSwapFactoryFilterer   // Log filterer for contract events
}

// SimpleSwapFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleSwapFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleSwapFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleSwapFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleSwapFactorySession struct {
	Contract     *SimpleSwapFactory // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SimpleSwapFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleSwapFactoryCallerSession struct {
	Contract *SimpleSwapFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// SimpleSwapFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleSwapFactoryTransactorSession struct {
	Contract     *SimpleSwapFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// SimpleSwapFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleSwapFactoryRaw struct {
	Contract *SimpleSwapFactory // Generic contract binding to access the raw methods on
}

// SimpleSwapFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleSwapFactoryCallerRaw struct {
	Contract *SimpleSwapFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleSwapFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleSwapFactoryTransactorRaw struct {
	Contract *SimpleSwapFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleSwapFactory creates a new instance of SimpleSwapFactory, bound to a specific deployed contract.
func NewSimpleSwapFactory(address common.Address, backend bind.ContractBackend) (*SimpleSwapFactory, error) {
	contract, err := bindSimpleSwapFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFactory{SimpleSwapFactoryCaller: SimpleSwapFactoryCaller{contract: contract}, SimpleSwapFactoryTransactor: SimpleSwapFactoryTransactor{contract: contract}, SimpleSwapFactoryFilterer: SimpleSwapFactoryFilterer{contract: contract}}, nil
}

// NewSimpleSwapFactoryCaller creates a new read-only instance of SimpleSwapFactory, bound to a specific deployed contract.
func NewSimpleSwapFactoryCaller(address common.Address, caller bind.ContractCaller) (*SimpleSwapFactoryCaller, error) {
	contract, err := bindSimpleSwapFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFactoryCaller{contract: contract}, nil
}

// NewSimpleSwapFactoryTransactor creates a new write-only instance of SimpleSwapFactory, bound to a specific deployed contract.
func NewSimpleSwapFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleSwapFactoryTransactor, error) {
	contract, err := bindSimpleSwapFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFactoryTransactor{contract: contract}, nil
}

// NewSimpleSwapFactoryFilterer creates a new log filterer instance of SimpleSwapFactory, bound to a specific deployed contract.
func NewSimpleSwapFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleSwapFactoryFilterer, error) {
	contract, err := bindSimpleSwapFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFactoryFilterer{contract: contract}, nil
}

// bindSimpleSwapFactory binds a generic wrapper to an already deployed contract.
func bindSimpleSwapFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSwapFactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSwapFactory *SimpleSwapFactoryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSwapFactory.Contract.SimpleSwapFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSwapFactory *SimpleSwapFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.SimpleSwapFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSwapFactory *SimpleSwapFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.SimpleSwapFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSwapFactory *SimpleSwapFactoryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSwapFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSwapFactory *SimpleSwapFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSwapFactory *SimpleSwapFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.contract.Transact(opts, method, params...)
}

// ERC20Address is a free data retrieval call binding the contract method 0xa6021ace.
//
// Solidity: function ERC20Address() constant returns(address)
func (_SimpleSwapFactory *SimpleSwapFactoryCaller) ERC20Address(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleSwapFactory.contract.Call(opts, out, "ERC20Address")
	return *ret0, err
}

// ERC20Address is a free data retrieval call binding the contract method 0xa6021ace.
//
// Solidity: function ERC20Address() constant returns(address)
func (_SimpleSwapFactory *SimpleSwapFactorySession) ERC20Address() (common.Address, error) {
	return _SimpleSwapFactory.Contract.ERC20Address(&_SimpleSwapFactory.CallOpts)
}

// ERC20Address is a free data retrieval call binding the contract method 0xa6021ace.
//
// Solidity: function ERC20Address() constant returns(address)
func (_SimpleSwapFactory *SimpleSwapFactoryCallerSession) ERC20Address() (common.Address, error) {
	return _SimpleSwapFactory.Contract.ERC20Address(&_SimpleSwapFactory.CallOpts)
}

// DeployedContracts is a free data retrieval call binding the contract method 0xc70242ad.
//
// Solidity: function deployedContracts(address ) constant returns(bool)
func (_SimpleSwapFactory *SimpleSwapFactoryCaller) DeployedContracts(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _SimpleSwapFactory.contract.Call(opts, out, "deployedContracts", arg0)
	return *ret0, err
}

// DeployedContracts is a free data retrieval call binding the contract method 0xc70242ad.
//
// Solidity: function deployedContracts(address ) constant returns(bool)
func (_SimpleSwapFactory *SimpleSwapFactorySession) DeployedContracts(arg0 common.Address) (bool, error) {
	return _SimpleSwapFactory.Contract.DeployedContracts(&_SimpleSwapFactory.CallOpts, arg0)
}

// DeployedContracts is a free data retrieval call binding the contract method 0xc70242ad.
//
// Solidity: function deployedContracts(address ) constant returns(bool)
func (_SimpleSwapFactory *SimpleSwapFactoryCallerSession) DeployedContracts(arg0 common.Address) (bool, error) {
	return _SimpleSwapFactory.Contract.DeployedContracts(&_SimpleSwapFactory.CallOpts, arg0)
}

// DeploySimpleSwap is a paid mutator transaction binding the contract method 0x576d7271.
//
// Solidity: function deploySimpleSwap(address issuer, uint256 defaultHardDepositTimeoutDuration) returns(address)
func (_SimpleSwapFactory *SimpleSwapFactoryTransactor) DeploySimpleSwap(opts *bind.TransactOpts, issuer common.Address, defaultHardDepositTimeoutDuration *big.Int) (*types.Transaction, error) {
	return _SimpleSwapFactory.contract.Transact(opts, "deploySimpleSwap", issuer, defaultHardDepositTimeoutDuration)
}

// DeploySimpleSwap is a paid mutator transaction binding the contract method 0x576d7271.
//
// Solidity: function deploySimpleSwap(address issuer, uint256 defaultHardDepositTimeoutDuration) returns(address)
func (_SimpleSwapFactory *SimpleSwapFactorySession) DeploySimpleSwap(issuer common.Address, defaultHardDepositTimeoutDuration *big.Int) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.DeploySimpleSwap(&_SimpleSwapFactory.TransactOpts, issuer, defaultHardDepositTimeoutDuration)
}

// DeploySimpleSwap is a paid mutator transaction binding the contract method 0x576d7271.
//
// Solidity: function deploySimpleSwap(address issuer, uint256 defaultHardDepositTimeoutDuration) returns(address)
func (_SimpleSwapFactory *SimpleSwapFactoryTransactorSession) DeploySimpleSwap(issuer common.Address, defaultHardDepositTimeoutDuration *big.Int) (*types.Transaction, error) {
	return _SimpleSwapFactory.Contract.DeploySimpleSwap(&_SimpleSwapFactory.TransactOpts, issuer, defaultHardDepositTimeoutDuration)
}

// SimpleSwapFactorySimpleSwapDeployedIterator is returned from FilterSimpleSwapDeployed and is used to iterate over the raw logs and unpacked data for SimpleSwapDeployed events raised by the SimpleSwapFactory contract.
type SimpleSwapFactorySimpleSwapDeployedIterator struct {
	Event *SimpleSwapFactorySimpleSwapDeployed // Event containing the contract specifics and raw log

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
func (it *SimpleSwapFactorySimpleSwapDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapFactorySimpleSwapDeployed)
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
		it.Event = new(SimpleSwapFactorySimpleSwapDeployed)
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
func (it *SimpleSwapFactorySimpleSwapDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapFactorySimpleSwapDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapFactorySimpleSwapDeployed represents a SimpleSwapDeployed event raised by the SimpleSwapFactory contract.
type SimpleSwapFactorySimpleSwapDeployed struct {
	ContractAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSimpleSwapDeployed is a free log retrieval operation binding the contract event 0xc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c004.
//
// Solidity: event SimpleSwapDeployed(address contractAddress)
func (_SimpleSwapFactory *SimpleSwapFactoryFilterer) FilterSimpleSwapDeployed(opts *bind.FilterOpts) (*SimpleSwapFactorySimpleSwapDeployedIterator, error) {

	logs, sub, err := _SimpleSwapFactory.contract.FilterLogs(opts, "SimpleSwapDeployed")
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFactorySimpleSwapDeployedIterator{contract: _SimpleSwapFactory.contract, event: "SimpleSwapDeployed", logs: logs, sub: sub}, nil
}

// WatchSimpleSwapDeployed is a free log subscription operation binding the contract event 0xc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c004.
//
// Solidity: event SimpleSwapDeployed(address contractAddress)
func (_SimpleSwapFactory *SimpleSwapFactoryFilterer) WatchSimpleSwapDeployed(opts *bind.WatchOpts, sink chan<- *SimpleSwapFactorySimpleSwapDeployed) (event.Subscription, error) {

	logs, sub, err := _SimpleSwapFactory.contract.WatchLogs(opts, "SimpleSwapDeployed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapFactorySimpleSwapDeployed)
				if err := _SimpleSwapFactory.contract.UnpackLog(event, "SimpleSwapDeployed", log); err != nil {
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

// ParseSimpleSwapDeployed is a log parse operation binding the contract event 0xc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c004.
//
// Solidity: event SimpleSwapDeployed(address contractAddress)
func (_SimpleSwapFactory *SimpleSwapFactoryFilterer) ParseSimpleSwapDeployed(log types.Log) (*SimpleSwapFactorySimpleSwapDeployed, error) {
	event := new(SimpleSwapFactorySimpleSwapDeployed)
	if err := _SimpleSwapFactory.contract.UnpackLog(event, "SimpleSwapDeployed", log); err != nil {
		return nil, err
	}
	return event, nil
}
