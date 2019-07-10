// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// ECDSAABI is the input ABI used to generate the binding from.
const ECDSAABI = "[]"

// ECDSABin is the compiled bytecode used for deploying new contracts.
const ECDSABin = `0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72305820aa21db549c95a55155d6ed05acfabca252bd4c39ddb99c4b41875d7f7c5a51a064736f6c634300050a0032`

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

// MathABI is the input ABI used to generate the binding from.
const MathABI = "[]"

// MathBin is the compiled bytecode used for deploying new contracts.
const MathBin = `0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a723058208455b7020ea0f7b9b5e62a6bfe8598823f2747909ba1c36e47e14922368b217664736f6c634300050a0032`

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

// SafeMathABI is the input ABI used to generate the binding from.
const SafeMathABI = "[]"

// SafeMathBin is the compiled bytecode used for deploying new contracts.
const SafeMathBin = `0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72305820196e378d47667cae0547d40767a2a433c2bc7ce96ee0f2908735bb5ca451324d64736f6c634300050a0032`

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

// SimpleSwapABI is the input ABI used to generate the binding from.
const SimpleSwapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"swap\",\"type\":\"address\"},{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"}],\"name\":\"chequeHash\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"},{\"name\":\"ownerSig\",\"type\":\"bytes\"},{\"name\":\"beneficarySig\",\"type\":\"bytes\"}],\"name\":\"submitCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"submitChequeOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cheques\",\"outputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"paidOut\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"},{\"name\":\"diff\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"liquidBalanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"timeout\",\"type\":\"uint256\"},{\"name\":\"ownerSig\",\"type\":\"bytes\"}],\"name\":\"submitChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"timeout\",\"type\":\"uint256\"}],\"name\":\"ChequeSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"bounced\",\"type\":\"uint256\"}],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"}]"

// SimpleSwapBin is the compiled bytecode used for deploying new contracts.
const SimpleSwapBin = `0x608060405234801561001057600080fd5b50604051611def380380611def8339818101604052602081101561003357600080fd5b810190808051906020019092919050505080600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050611d5a806100956000396000f3fe6080604052600436106100e85760003560e01c8063946f46a21161008a578063c76a4d3111610059578063c76a4d311461074c578063eeec647a146107b1578063f6153ccd14610802578063f890673b1461082d576100e8565b8063946f46a214610602578063b6343b0d14610653578063b7770350146106c6578063b7ec1a3314610721576100e8565b80634f823a4c116100c65780634f823a4c1461028e578063595d8e4b1461042b5780636162913b146105315780638da5cb5b146105ab576100e8565b8063030aca3e146101555780632e1a7d4d146101f8578063338f3fed14610233575b7fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a1005b34801561016157600080fd5b506101e2600480360360a081101561017857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050610913565b6040518082815260200191505060405180910390f35b34801561020457600080fd5b506102316004803603602081101561021b57600080fd5b81019080803590602001909291905050506109c5565b005b34801561023f57600080fd5b5061028c6004803603604081101561025657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610b54565b005b34801561029a57600080fd5b50610429600480360360c08110156102b157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190803590602001909291908035906020019064010000000081111561030c57600080fd5b82018360208201111561031e57600080fd5b8035906020019184600183028401116401000000008311171561034057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001906401000000008111156103a357600080fd5b8201836020820111156103b557600080fd5b803590602001918460018302840111640100000000831117156103d757600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610d79565b005b34801561043757600080fd5b5061052f600480360360a081101561044e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190803590602001906401000000008111156104a957600080fd5b8201836020820111156104bb57600080fd5b803590602001918460018302840111640100000000831117156104dd57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610efe565b005b34801561053d57600080fd5b506105806004803603602081101561055457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061106d565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b3480156105b757600080fd5b506105c061109d565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561060e57600080fd5b506106516004803603602081101561062557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506110c3565b005b34801561065f57600080fd5b506106a26004803603602081101561067657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611280565b60405180848152602001838152602001828152602001935050505060405180910390f35b3480156106d257600080fd5b5061071f600480360360408110156106e957600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506112aa565b005b34801561072d57600080fd5b50610736611493565b6040518082815260200191505060405180910390f35b34801561075857600080fd5b5061079b6004803603602081101561076f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506114c6565b6040518082815260200191505060405180910390f35b3480156107bd57600080fd5b50610800600480360360208110156107d457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061152b565b005b34801561080e57600080fd5b50610817611799565b6040518082815260200191505060405180910390f35b34801561083957600080fd5b506109116004803603608081101561085057600080fd5b810190808035906020019092919080359060200190929190803590602001909291908035906020019064010000000081111561088b57600080fd5b82018360208201111561089d57600080fd5b803590602001918460018302840111640100000000831117156108bf57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061179f565b005b60008584868585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018381526020018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610a88576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f53696d706c65537761703a206e6f74206f776e6572000000000000000000000081525060200191505060405180910390fd5b610a90611493565b811115610ae8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526028815260200180611cfe6028913960400191505060405180910390fd5b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610b50573d6000803e3d6000fd5b5050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c17576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f53696d706c65537761703a206e6f74206f776e6572000000000000000000000081525060200191505060405180910390fd5b3073ffffffffffffffffffffffffffffffffffffffff1631610c448260025461188990919063ffffffff16565b1115610c9b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526035815260200180611cc96035913960400191505060405180910390fd5b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050610cf582826000015461188990919063ffffffff16565b8160000181905550610d128260025461188990919063ffffffff16565b600281905550600081600101819055508273ffffffffffffffffffffffffffffffffffffffff167f316b52a3b151a9fa03623b976971e3d77a6d969d90005343c9cc9f7d89de3a0f82600001546040518082815260200191505060405180910390a2505050565b610d8f610d893088888888610913565b836118a8565b73ffffffffffffffffffffffffffffffffffffffff16600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610e51576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601c8152602001807f53696d706c65537761703a20696e76616c6964206f776e65725369670000000081525060200191505060405180910390fd5b610e67610e613088888888610913565b826118a8565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff1614610eea576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180611c5e6022913960400191505060405180910390fd5b610ef6868686866118c4565b505050505050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610fc1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f53696d706c65537761703a206e6f74206f776e6572000000000000000000000081525060200191505060405180910390fd5b610fd7610fd13087878787610913565b826118a8565b73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff161461105a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180611c5e6022913960400191505060405180910390fd5b611066858585856118c4565b5050505050565b60006020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050600081600101541415611181576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f53696d706c65537761703a206e6f2074696d656f75742073657400000000000081525060200191505060405180910390fd5b80600101544210156111de576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526025815260200180611c806025913960400191505060405180910390fd5b6111f9816002015482600001546119fa90919063ffffffff16565b81600001819055506000816001018190555061122481600201546002546119fa90919063ffffffff16565b6002819055508173ffffffffffffffffffffffffffffffffffffffff167f316b52a3b151a9fa03623b976971e3d77a6d969d90005343c9cc9f7d89de3a0f82600001546040518082815260200191505060405180910390a25050565b60016020528060005260406000206000915090508060000154908060010154908060020154905083565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461136d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260158152602001807f53696d706c65537761703a206e6f74206f776e6572000000000000000000000081525060200191505060405180910390fd5b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600001548210611429576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260208152602001807f53696d706c65537761703a2062616c616e636520696e73756666696369656e7481525060200191505060405180910390fd5b62015180420181600101819055508181600201819055508273ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a836040518082815260200191505060405180910390a2505050565b60006114c16002543073ffffffffffffffffffffffffffffffffffffffff16316119fa90919063ffffffff16565b905090565b6000611524600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611516611493565b61188990919063ffffffff16565b9050919050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600301544210156115ca576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526024815260200180611ca56024913960400191505060405180910390fd5b60006115e7826002015483600101546119fa90919063ffffffff16565b90506000811161165f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f53696d706c65537761703a206e6f2062616c616e6365206f776564000000000081525060200191505060405180910390fd5b60008061166c8584611a1a565b91509150600081146116d85783600001548573ffffffffffffffffffffffffffffffffffffffff167f9a8ffac28ab8409cf23919982224c251b36b3e358ba30c93381469a5b8d9a7558484604051808381526020018281526020019250505060405180910390a361172c565b83600001548573ffffffffffffffffffffffffffffffffffffffff167f67e81448d86cfbbf3135a82e52b8eb4eb9863ec9130b05e836045e45df94d788846040518082815260200191505060405180910390a35b61174382856002015461188990919063ffffffff16565b84600201819055508473ffffffffffffffffffffffffffffffffffffffff166108fc839081150290604051600060405180830381858888f19350505050158015611791573d6000803e3d6000fd5b505050505050565b60025481565b6117b56117af3033878787610913565b826118a8565b73ffffffffffffffffffffffffffffffffffffffff16600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611877576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601c8152602001807f53696d706c65537761703a20696e76616c6964206f776e65725369670000000081525060200191505060405180910390fd5b611883338585856118c4565b50505050565b60008082840190508381101561189e57600080fd5b8091505092915050565b60006118bc6118b684611b0a565b83611b62565b905092915050565b60008060008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000154841161197f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f53696d706c65537761703a20696e76616c69642073657269616c00000000000081525060200191505060405180910390fd5b8381600001819055508281600101819055508142018160030181905550838573ffffffffffffffffffffffffffffffffffffffff167f543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c8585604051808381526020018281526020019250505060405180910390a35050505050565b600082821115611a0957600080fd5b600082840390508091505092915050565b600080611a6983600160008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611c44565b915060008214611ad45781600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000160008282540392505081905550816002600082825403925050819055505b600082840390506000611ae5611493565b9050818110611af657849350611b01565b808401935080820392505b50509250929050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b6000806000806041855114611b7d5760009350505050611c3e565b6020850151925060408501519150606085015160001a9050601b8160ff161015611ba857601b810190505b601b8160ff1614158015611bc05750601c8160ff1614155b15611bd15760009350505050611c3e565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611c2e573d6000803e3d6000fd5b5050506020604051035193505050505b92915050565b6000818310611c535781611c55565b825b90509291505056fe53696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a20636865717565206e6f74207965742074696d6564206f757453696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e63652053696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e74a265627a7a723058203030d65ac8ad110d153afcbf234aaf8245393b549b3a7b6326e221582c48af8f64736f6c634300050a0032`

// DeploySimpleSwap deploys a new Ethereum contract, binding an instance of SimpleSwap to it.
func DeploySimpleSwap(auth *bind.TransactOpts, backend bind.ContractBackend, _owner common.Address) (common.Address, *types.Transaction, *SimpleSwap, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSwapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleSwapBin), backend, _owner)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SimpleSwap{SimpleSwapCaller: SimpleSwapCaller{contract: contract}, SimpleSwapTransactor: SimpleSwapTransactor{contract: contract}, SimpleSwapFilterer: SimpleSwapFilterer{contract: contract}}, nil
}

// SimpleSwap is an auto generated Go binding around an Ethereum contract.
type SimpleSwap struct {
	SimpleSwapCaller     // Read-only binding to the contract
	SimpleSwapTransactor // Write-only binding to the contract
	SimpleSwapFilterer   // Log filterer for contract events
}

// SimpleSwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type SimpleSwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SimpleSwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SimpleSwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SimpleSwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SimpleSwapSession struct {
	Contract     *SimpleSwap       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SimpleSwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SimpleSwapCallerSession struct {
	Contract *SimpleSwapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// SimpleSwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SimpleSwapTransactorSession struct {
	Contract     *SimpleSwapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// SimpleSwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type SimpleSwapRaw struct {
	Contract *SimpleSwap // Generic contract binding to access the raw methods on
}

// SimpleSwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SimpleSwapCallerRaw struct {
	Contract *SimpleSwapCaller // Generic read-only contract binding to access the raw methods on
}

// SimpleSwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SimpleSwapTransactorRaw struct {
	Contract *SimpleSwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSimpleSwap creates a new instance of SimpleSwap, bound to a specific deployed contract.
func NewSimpleSwap(address common.Address, backend bind.ContractBackend) (*SimpleSwap, error) {
	contract, err := bindSimpleSwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SimpleSwap{SimpleSwapCaller: SimpleSwapCaller{contract: contract}, SimpleSwapTransactor: SimpleSwapTransactor{contract: contract}, SimpleSwapFilterer: SimpleSwapFilterer{contract: contract}}, nil
}

// NewSimpleSwapCaller creates a new read-only instance of SimpleSwap, bound to a specific deployed contract.
func NewSimpleSwapCaller(address common.Address, caller bind.ContractCaller) (*SimpleSwapCaller, error) {
	contract, err := bindSimpleSwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapCaller{contract: contract}, nil
}

// NewSimpleSwapTransactor creates a new write-only instance of SimpleSwap, bound to a specific deployed contract.
func NewSimpleSwapTransactor(address common.Address, transactor bind.ContractTransactor) (*SimpleSwapTransactor, error) {
	contract, err := bindSimpleSwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapTransactor{contract: contract}, nil
}

// NewSimpleSwapFilterer creates a new log filterer instance of SimpleSwap, bound to a specific deployed contract.
func NewSimpleSwapFilterer(address common.Address, filterer bind.ContractFilterer) (*SimpleSwapFilterer, error) {
	contract, err := bindSimpleSwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapFilterer{contract: contract}, nil
}

// bindSimpleSwap binds a generic wrapper to an already deployed contract.
func bindSimpleSwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSwap *SimpleSwapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSwap.Contract.SimpleSwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSwap *SimpleSwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SimpleSwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSwap *SimpleSwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SimpleSwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SimpleSwap *SimpleSwapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _SimpleSwap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SimpleSwap *SimpleSwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SimpleSwap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SimpleSwap *SimpleSwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SimpleSwap.Contract.contract.Transact(opts, method, params...)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 timeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapCaller) ChequeHash(opts *bind.CallOpts, swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "chequeHash", swap, beneficiary, serial, amount, timeout)
	return *ret0, err
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 timeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int) ([32]byte, error) {
	return _SimpleSwap.Contract.ChequeHash(&_SimpleSwap.CallOpts, swap, beneficiary, serial, amount, timeout)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 timeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapCallerSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int) ([32]byte, error) {
	return _SimpleSwap.Contract.ChequeHash(&_SimpleSwap.CallOpts, swap, beneficiary, serial, amount, timeout)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 timeout)
func (_SimpleSwap *SimpleSwapCaller) Cheques(opts *bind.CallOpts, arg0 common.Address) (struct {
	Serial  *big.Int
	Amount  *big.Int
	PaidOut *big.Int
	Timeout *big.Int
}, error) {
	ret := new(struct {
		Serial  *big.Int
		Amount  *big.Int
		PaidOut *big.Int
		Timeout *big.Int
	})
	out := ret
	err := _SimpleSwap.contract.Call(opts, out, "cheques", arg0)
	return *ret, err
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 timeout)
func (_SimpleSwap *SimpleSwapSession) Cheques(arg0 common.Address) (struct {
	Serial  *big.Int
	Amount  *big.Int
	PaidOut *big.Int
	Timeout *big.Int
}, error) {
	return _SimpleSwap.Contract.Cheques(&_SimpleSwap.CallOpts, arg0)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 timeout)
func (_SimpleSwap *SimpleSwapCallerSession) Cheques(arg0 common.Address) (struct {
	Serial  *big.Int
	Amount  *big.Int
	PaidOut *big.Int
	Timeout *big.Int
}, error) {
	return _SimpleSwap.Contract.Cheques(&_SimpleSwap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 timeout, uint256 diff)
func (_SimpleSwap *SimpleSwapCaller) HardDeposits(opts *bind.CallOpts, arg0 common.Address) (struct {
	Amount  *big.Int
	Timeout *big.Int
	Diff    *big.Int
}, error) {
	ret := new(struct {
		Amount  *big.Int
		Timeout *big.Int
		Diff    *big.Int
	})
	out := ret
	err := _SimpleSwap.contract.Call(opts, out, "hardDeposits", arg0)
	return *ret, err
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 timeout, uint256 diff)
func (_SimpleSwap *SimpleSwapSession) HardDeposits(arg0 common.Address) (struct {
	Amount  *big.Int
	Timeout *big.Int
	Diff    *big.Int
}, error) {
	return _SimpleSwap.Contract.HardDeposits(&_SimpleSwap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 timeout, uint256 diff)
func (_SimpleSwap *SimpleSwapCallerSession) HardDeposits(arg0 common.Address) (struct {
	Amount  *big.Int
	Timeout *big.Int
	Diff    *big.Int
}, error) {
	return _SimpleSwap.Contract.HardDeposits(&_SimpleSwap.CallOpts, arg0)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) LiquidBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "liquidBalance")
	return *ret0, err
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) LiquidBalance() (*big.Int, error) {
	return _SimpleSwap.Contract.LiquidBalance(&_SimpleSwap.CallOpts)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) LiquidBalance() (*big.Int, error) {
	return _SimpleSwap.Contract.LiquidBalance(&_SimpleSwap.CallOpts)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) LiquidBalanceFor(opts *bind.CallOpts, beneficiary common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "liquidBalanceFor", beneficiary)
	return *ret0, err
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.LiquidBalanceFor(&_SimpleSwap.CallOpts, beneficiary)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.LiquidBalanceFor(&_SimpleSwap.CallOpts, beneficiary)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleSwap *SimpleSwapCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleSwap *SimpleSwapSession) Owner() (common.Address, error) {
	return _SimpleSwap.Contract.Owner(&_SimpleSwap.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_SimpleSwap *SimpleSwapCallerSession) Owner() (common.Address, error) {
	return _SimpleSwap.Contract.Owner(&_SimpleSwap.CallOpts)
}

// TotalDeposit is a free data retrieval call binding the contract method 0xf6153ccd.
//
// Solidity: function totalDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) TotalDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "totalDeposit")
	return *ret0, err
}

// TotalDeposit is a free data retrieval call binding the contract method 0xf6153ccd.
//
// Solidity: function totalDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) TotalDeposit() (*big.Int, error) {
	return _SimpleSwap.Contract.TotalDeposit(&_SimpleSwap.CallOpts)
}

// TotalDeposit is a free data retrieval call binding the contract method 0xf6153ccd.
//
// Solidity: function totalDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) TotalDeposit() (*big.Int, error) {
	return _SimpleSwap.Contract.TotalDeposit(&_SimpleSwap.CallOpts)
}

// CashCheque is a paid mutator transaction binding the contract method 0xeeec647a.
//
// Solidity: function cashCheque(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashCheque(opts *bind.TransactOpts, beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "cashCheque", beneficiary)
}

// CashCheque is a paid mutator transaction binding the contract method 0xeeec647a.
//
// Solidity: function cashCheque(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapSession) CashCheque(beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiary)
}

// CashCheque is a paid mutator transaction binding the contract method 0xeeec647a.
//
// Solidity: function cashCheque(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashCheque(beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapTransactor) DecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "decreaseHardDeposit", beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.Contract.DecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _SimpleSwap.Contract.DecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_SimpleSwap *SimpleSwapTransactor) IncreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "increaseHardDeposit", beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_SimpleSwap *SimpleSwapSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.IncreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.IncreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, amount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 diff) returns()
func (_SimpleSwap *SimpleSwapTransactor) PrepareDecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, diff *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "prepareDecreaseHardDeposit", beneficiary, diff)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 diff) returns()
func (_SimpleSwap *SimpleSwapSession) PrepareDecreaseHardDeposit(beneficiary common.Address, diff *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, diff)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 diff) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) PrepareDecreaseHardDeposit(beneficiary common.Address, diff *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, diff)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitCheque(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitCheque", beneficiary, serial, amount, timeout, ownerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitCheque(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, timeout, ownerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitCheque(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, timeout, ownerSig, beneficarySig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitChequeBeneficiary(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitChequeBeneficiary", serial, amount, timeout, ownerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeBeneficiary(&_SimpleSwap.TransactOpts, serial, amount, timeout, ownerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 timeout, bytes ownerSig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, timeout *big.Int, ownerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeBeneficiary(&_SimpleSwap.TransactOpts, serial, amount, timeout, ownerSig)
}

// SubmitChequeOwner is a paid mutator transaction binding the contract method 0x595d8e4b.
//
// Solidity: function submitChequeOwner(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitChequeOwner(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitChequeOwner", beneficiary, serial, amount, timeout, beneficiarySig)
}

// SubmitChequeOwner is a paid mutator transaction binding the contract method 0x595d8e4b.
//
// Solidity: function submitChequeOwner(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitChequeOwner(beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeOwner(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, timeout, beneficiarySig)
}

// SubmitChequeOwner is a paid mutator transaction binding the contract method 0x595d8e4b.
//
// Solidity: function submitChequeOwner(address beneficiary, uint256 serial, uint256 amount, uint256 timeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitChequeOwner(beneficiary common.Address, serial *big.Int, amount *big.Int, timeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeOwner(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, timeout, beneficiarySig)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_SimpleSwap *SimpleSwapTransactor) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "withdraw", amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_SimpleSwap *SimpleSwapSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.Withdraw(&_SimpleSwap.TransactOpts, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.Withdraw(&_SimpleSwap.TransactOpts, amount)
}

// SimpleSwapChequeBouncedIterator is returned from FilterChequeBounced and is used to iterate over the raw logs and unpacked data for ChequeBounced events raised by the SimpleSwap contract.
type SimpleSwapChequeBouncedIterator struct {
	Event *SimpleSwapChequeBounced // Event containing the contract specifics and raw log

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
func (it *SimpleSwapChequeBouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapChequeBounced)
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
		it.Event = new(SimpleSwapChequeBounced)
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
func (it *SimpleSwapChequeBouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapChequeBouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapChequeBounced represents a ChequeBounced event raised by the SimpleSwap contract.
type SimpleSwapChequeBounced struct {
	Beneficiary common.Address
	Serial      *big.Int
	Paid        *big.Int
	Bounced     *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChequeBounced is a free log retrieval operation binding the contract event 0x9a8ffac28ab8409cf23919982224c251b36b3e358ba30c93381469a5b8d9a755.
//
// Solidity: event ChequeBounced(address indexed beneficiary, uint256 indexed serial, uint256 paid, uint256 bounced)
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeBounced(opts *bind.FilterOpts, beneficiary []common.Address, serial []*big.Int) (*SimpleSwapChequeBouncedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeBounced", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeBouncedIterator{contract: _SimpleSwap.contract, event: "ChequeBounced", logs: logs, sub: sub}, nil
}

// WatchChequeBounced is a free log subscription operation binding the contract event 0x9a8ffac28ab8409cf23919982224c251b36b3e358ba30c93381469a5b8d9a755.
//
// Solidity: event ChequeBounced(address indexed beneficiary, uint256 indexed serial, uint256 paid, uint256 bounced)
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeBounced(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeBounced, beneficiary []common.Address, serial []*big.Int) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeBounced", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapChequeBounced)
				if err := _SimpleSwap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
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

// SimpleSwapChequeCashedIterator is returned from FilterChequeCashed and is used to iterate over the raw logs and unpacked data for ChequeCashed events raised by the SimpleSwap contract.
type SimpleSwapChequeCashedIterator struct {
	Event *SimpleSwapChequeCashed // Event containing the contract specifics and raw log

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
func (it *SimpleSwapChequeCashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapChequeCashed)
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
		it.Event = new(SimpleSwapChequeCashed)
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
func (it *SimpleSwapChequeCashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapChequeCashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapChequeCashed represents a ChequeCashed event raised by the SimpleSwap contract.
type SimpleSwapChequeCashed struct {
	Beneficiary common.Address
	Serial      *big.Int
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChequeCashed is a free log retrieval operation binding the contract event 0x67e81448d86cfbbf3135a82e52b8eb4eb9863ec9130b05e836045e45df94d788.
//
// Solidity: event ChequeCashed(address indexed beneficiary, uint256 indexed serial, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeCashed(opts *bind.FilterOpts, beneficiary []common.Address, serial []*big.Int) (*SimpleSwapChequeCashedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeCashed", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeCashedIterator{contract: _SimpleSwap.contract, event: "ChequeCashed", logs: logs, sub: sub}, nil
}

// WatchChequeCashed is a free log subscription operation binding the contract event 0x67e81448d86cfbbf3135a82e52b8eb4eb9863ec9130b05e836045e45df94d788.
//
// Solidity: event ChequeCashed(address indexed beneficiary, uint256 indexed serial, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeCashed(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeCashed, beneficiary []common.Address, serial []*big.Int) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeCashed", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapChequeCashed)
				if err := _SimpleSwap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
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

// SimpleSwapChequeSubmittedIterator is returned from FilterChequeSubmitted and is used to iterate over the raw logs and unpacked data for ChequeSubmitted events raised by the SimpleSwap contract.
type SimpleSwapChequeSubmittedIterator struct {
	Event *SimpleSwapChequeSubmitted // Event containing the contract specifics and raw log

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
func (it *SimpleSwapChequeSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapChequeSubmitted)
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
		it.Event = new(SimpleSwapChequeSubmitted)
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
func (it *SimpleSwapChequeSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapChequeSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapChequeSubmitted represents a ChequeSubmitted event raised by the SimpleSwap contract.
type SimpleSwapChequeSubmitted struct {
	Beneficiary common.Address
	Serial      *big.Int
	Amount      *big.Int
	Timeout     *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChequeSubmitted is a free log retrieval operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 timeout)
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeSubmitted(opts *bind.FilterOpts, beneficiary []common.Address, serial []*big.Int) (*SimpleSwapChequeSubmittedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeSubmitted", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeSubmittedIterator{contract: _SimpleSwap.contract, event: "ChequeSubmitted", logs: logs, sub: sub}, nil
}

// WatchChequeSubmitted is a free log subscription operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 timeout)
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeSubmitted(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeSubmitted, beneficiary []common.Address, serial []*big.Int) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeSubmitted", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapChequeSubmitted)
				if err := _SimpleSwap.contract.UnpackLog(event, "ChequeSubmitted", log); err != nil {
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

// SimpleSwapDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the SimpleSwap contract.
type SimpleSwapDepositIterator struct {
	Event *SimpleSwapDeposit // Event containing the contract specifics and raw log

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
func (it *SimpleSwapDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapDeposit)
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
		it.Event = new(SimpleSwapDeposit)
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
func (it *SimpleSwapDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapDeposit represents a Deposit event raised by the SimpleSwap contract.
type SimpleSwapDeposit struct {
	Depositor common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address depositor, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) FilterDeposit(opts *bind.FilterOpts) (*SimpleSwapDepositIterator, error) {

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &SimpleSwapDepositIterator{contract: _SimpleSwap.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address depositor, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *SimpleSwapDeposit) (event.Subscription, error) {

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapDeposit)
				if err := _SimpleSwap.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// SimpleSwapHardDepositChangedIterator is returned from FilterHardDepositChanged and is used to iterate over the raw logs and unpacked data for HardDepositChanged events raised by the SimpleSwap contract.
type SimpleSwapHardDepositChangedIterator struct {
	Event *SimpleSwapHardDepositChanged // Event containing the contract specifics and raw log

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
func (it *SimpleSwapHardDepositChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapHardDepositChanged)
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
		it.Event = new(SimpleSwapHardDepositChanged)
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
func (it *SimpleSwapHardDepositChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapHardDepositChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapHardDepositChanged represents a HardDepositChanged event raised by the SimpleSwap contract.
type SimpleSwapHardDepositChanged struct {
	Beneficiary common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositChanged is a free log retrieval operation binding the contract event 0x316b52a3b151a9fa03623b976971e3d77a6d969d90005343c9cc9f7d89de3a0f.
//
// Solidity: event HardDepositChanged(address indexed beneficiary, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) FilterHardDepositChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*SimpleSwapHardDepositChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "HardDepositChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapHardDepositChangedIterator{contract: _SimpleSwap.contract, event: "HardDepositChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositChanged is a free log subscription operation binding the contract event 0x316b52a3b151a9fa03623b976971e3d77a6d969d90005343c9cc9f7d89de3a0f.
//
// Solidity: event HardDepositChanged(address indexed beneficiary, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) WatchHardDepositChanged(opts *bind.WatchOpts, sink chan<- *SimpleSwapHardDepositChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "HardDepositChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapHardDepositChanged)
				if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositChanged", log); err != nil {
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

// SimpleSwapHardDepositDecreasePreparedIterator is returned from FilterHardDepositDecreasePrepared and is used to iterate over the raw logs and unpacked data for HardDepositDecreasePrepared events raised by the SimpleSwap contract.
type SimpleSwapHardDepositDecreasePreparedIterator struct {
	Event *SimpleSwapHardDepositDecreasePrepared // Event containing the contract specifics and raw log

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
func (it *SimpleSwapHardDepositDecreasePreparedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapHardDepositDecreasePrepared)
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
		it.Event = new(SimpleSwapHardDepositDecreasePrepared)
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
func (it *SimpleSwapHardDepositDecreasePreparedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapHardDepositDecreasePreparedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapHardDepositDecreasePrepared represents a HardDepositDecreasePrepared event raised by the SimpleSwap contract.
type SimpleSwapHardDepositDecreasePrepared struct {
	Beneficiary common.Address
	Diff        *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreasePrepared is a free log retrieval operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 diff)
func (_SimpleSwap *SimpleSwapFilterer) FilterHardDepositDecreasePrepared(opts *bind.FilterOpts, beneficiary []common.Address) (*SimpleSwapHardDepositDecreasePreparedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapHardDepositDecreasePreparedIterator{contract: _SimpleSwap.contract, event: "HardDepositDecreasePrepared", logs: logs, sub: sub}, nil
}

// WatchHardDepositDecreasePrepared is a free log subscription operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 diff)
func (_SimpleSwap *SimpleSwapFilterer) WatchHardDepositDecreasePrepared(opts *bind.WatchOpts, sink chan<- *SimpleSwapHardDepositDecreasePrepared, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapHardDepositDecreasePrepared)
				if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
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
