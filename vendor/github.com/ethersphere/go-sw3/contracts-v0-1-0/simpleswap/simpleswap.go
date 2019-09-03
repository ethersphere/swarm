// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package simpleswap

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
var ECDSABin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a723058208b264442266842114d00302eed910833246a8ca31d4a342465d27ee38b3c2d0c64736f6c634300050a0032"

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
var MathBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a7230582012a658010ec853ca058fc84f498376cc4e43ab59dbd66c48e1cc4b70ddafb2f464736f6c634300050a0032"

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
var SafeMathBin = "0x60556023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a723058200912251cef3833d2207d3ee96457126d52aa9798065e060e96ec7c2d505645a364736f6c634300050a0032"

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
const SimpleSwapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"availableBalanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\"},{\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"cashChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\"},{\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"},{\"name\":\"callerPayout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"issuer\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DEFAULT_HARDDEPOSIT_DECREASE_TIMEOUT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"paidOut\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"canBeDecreasedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"setCustomHardDepositDecreaseTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalHardDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_issuer\",\"type\":\"address\"},{\"name\":\"defaultHardDepositTimeoutDuration\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"totalPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cumulativePayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"callerPayout\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositAmountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseTimeout\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreaseTimeoutChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdraw\",\"type\":\"event\"}]"

// SimpleSwapBin is the compiled bytecode used for deploying new contracts.
var SimpleSwapBin = "0x6080604052604051611f6e380380611f6e8339818101604052604081101561002657600080fd5b8101908080519060200190929190805190602001909291905050508060008190555081600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060003411156100fe577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b5050611e5f8061010f6000396000f3fe6080604052600436106100dd5760003560e01c806381f03fcb1161007f578063b777035011610059578063b7770350146106a6578063b7ec1a3314610701578063df3243801461072c578063e0bcf13a1461081e576100dd565b806381f03fcb14610576578063946f46a2146105db578063b6343b0d1461062c576100dd565b80631d143848116100bb5780631d1438481461045e5780632e1a7d4d146104b5578063338f3fed146104f05780635eb541601461054b576100dd565b8063065c804f146101545780630d5f2659146101b95780631633fb1d146102ab575b6000341115610152577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b005b34801561016057600080fd5b506101a36004803603602081101561017757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610849565b6040518082815260200191505060405180910390f35b3480156101c557600080fd5b506102a9600480360360608110156101dc57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019064010000000081111561022357600080fd5b82018360208201111561023557600080fd5b8035906020019184600183028401116401000000008311171561025757600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506108ae565b005b3480156102b757600080fd5b5061045c600480360360c08110156102ce57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019064010000000081111561033557600080fd5b82018360208201111561034757600080fd5b8035906020019184600183028401116401000000008311171561036957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929080359060200190929190803590602001906401000000008111156103d657600080fd5b8201836020820111156103e857600080fd5b8035906020019184600183028401116401000000008311171561040a57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506108c1565b005b34801561046a57600080fd5b5061047361096f565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156104c157600080fd5b506104ee600480360360208110156104d857600080fd5b8101908080359060200190929190505050610995565b005b3480156104fc57600080fd5b506105496004803603604081101561051357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610b5b565b005b34801561055757600080fd5b50610560610d80565b6040518082815260200191505060405180910390f35b34801561058257600080fd5b506105c56004803603602081101561059957600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d86565b6040518082815260200191505060405180910390f35b3480156105e757600080fd5b5061062a600480360360208110156105fe57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610d9e565b005b34801561063857600080fd5b5061067b6004803603602081101561064f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610ef1565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b3480156106b257600080fd5b506106ff600480360360408110156106c957600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610f21565b005b34801561070d57600080fd5b50610716611109565b6040518082815260200191505060405180910390f35b34801561073857600080fd5b5061081c6004803603606081101561074f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019064010000000081111561079657600080fd5b8201836020820111156107a857600080fd5b803590602001918460018302840111640100000000831117156107ca57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061113c565b005b34801561082a57600080fd5b50610833611372565b6040518082815260200191505060405180910390f35b60006108a7600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154610899611109565b61137890919063ffffffff16565b9050919050565b6108bc338484600085611400565b505050565b6108d76108d13033878987611925565b84611a06565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff161461095a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180611d956022913960400191505060405180910390fd5b6109678686868585611400565b505050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610a58576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610a60611109565b811115610ab8576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526028815260200180611e036028913960400191505060405180910390fd5b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610b20573d6000803e3d6000fd5b507f5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d816040518082815260200191505060405180910390a150565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c1e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b3073ffffffffffffffffffffffffffffffffffffffff1631610c4b8260035461137890919063ffffffff16565b1115610ca2576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526034815260200180611d616034913960400191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050610cfc82826000015461137890919063ffffffff16565b8160000181905550610d198260035461137890919063ffffffff16565b600381905550600081600301819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60005481565b60016020528060005260406000206000915090505481565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600301544210158015610dfa57506000816003015414155b610e4f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526025815260200180611db76025913960400191505060405180910390fd5b610e6a81600101548260000154611a2290919063ffffffff16565b816000018190555060008160030181905550610e958160010154600354611a2290919063ffffffff16565b6003819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60026020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610fe4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000154821115611084576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526027815260200180611ddc6027913960400191505060405180910390fd5b60008082600201541461109b57816002015461109f565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b60006111376003543073ffffffffffffffffffffffffffffffffffffffff1631611a2290919063ffffffff16565b905090565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146111ff576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b61121361120d308585611aab565b82611a06565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614611296576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526022815260200180611d956022913960400191505060405180910390fd5b81600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f600260008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201546040518082815260200191505060405180910390a2505050565b60035481565b6000808284019050838110156113f6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461152c57611469611463308786611b4b565b82611a06565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461152b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b5b6000611580600160008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205485611a2290919063ffffffff16565b905060006115968261159189610849565b611beb565b905060006115e682600260008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611beb565b90508482101561165e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a2063616e6e6f74207061792063616c6c657200000081525060200191505060405180910390fd5b6000811461171d576116bb81600260008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611a2290919063ffffffff16565b600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000018190555061171681600354611a2290919063ffffffff16565b6003819055505b61176f82600160008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205461137890919063ffffffff16565b600160008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508673ffffffffffffffffffffffffffffffffffffffff166108fc6117df8785611a2290919063ffffffff16565b9081150290604051600060405180830381858888f1935050505015801561180a573d6000803e3d6000fd5b506000851461185b573373ffffffffffffffffffffffffffffffffffffffff166108fc869081150290604051600060405180830381858888f19350505050158015611859573d6000803e3d6000fd5b505b3373ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f858a8a60405180848152602001838152602001828152602001935050505060405180910390a481831461191b577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b5050505050505050565b60008585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b6000611a1a611a1484611c04565b83611c5c565b905092915050565b600082821115611a9a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525060200191505060405180910390fd5b600082840390508091505092915050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000838383604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140182815260200193505050506040516020818303038152906040528051906020012090509392505050565b6000818310611bfa5781611bfc565b825b905092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b60006041825114611c705760009050611d5a565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c1115611cc45760009350505050611d5a565b601b8160ff1614158015611cdc5750601c8160ff1614155b15611ced5760009350505050611d5a565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015611d4a573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fe53696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e636553696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e74a265627a7a72305820b250198c2ed4a8c9ba8db19510d8cb59d6971ebb29dc2d9467f49bb1b0514c0164736f6c634300050a0032"

// DeploySimpleSwap deploys a new Ethereum contract, binding an instance of SimpleSwap to it.
func DeploySimpleSwap(auth *bind.TransactOpts, backend bind.ContractBackend, _issuer common.Address, defaultHardDepositTimeoutDuration *big.Int) (common.Address, *types.Transaction, *SimpleSwap, error) {
	parsed, err := abi.JSON(strings.NewReader(SimpleSwapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SimpleSwapBin), backend, _issuer, defaultHardDepositTimeoutDuration)
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

// DEFAULTHARDDEPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x5eb54160.
//
// Solidity: function DEFAULT_HARDDEPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) DEFAULTHARDDEPOSITDECREASETIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "DEFAULT_HARDDEPOSIT_DECREASE_TIMEOUT")
	return *ret0, err
}

// DEFAULTHARDDEPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x5eb54160.
//
// Solidity: function DEFAULT_HARDDEPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) DEFAULTHARDDEPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _SimpleSwap.Contract.DEFAULTHARDDEPOSITDECREASETIMEOUT(&_SimpleSwap.CallOpts)
}

// DEFAULTHARDDEPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x5eb54160.
//
// Solidity: function DEFAULT_HARDDEPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) DEFAULTHARDDEPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _SimpleSwap.Contract.DEFAULTHARDDEPOSITDECREASETIMEOUT(&_SimpleSwap.CallOpts)
}

// AvailableBalanceFor is a free data retrieval call binding the contract method 0x065c804f.
//
// Solidity: function availableBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) AvailableBalanceFor(opts *bind.CallOpts, beneficiary common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "availableBalanceFor", beneficiary)
	return *ret0, err
}

// AvailableBalanceFor is a free data retrieval call binding the contract method 0x065c804f.
//
// Solidity: function availableBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) AvailableBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.AvailableBalanceFor(&_SimpleSwap.CallOpts, beneficiary)
}

// AvailableBalanceFor is a free data retrieval call binding the contract method 0x065c804f.
//
// Solidity: function availableBalanceFor(address beneficiary) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) AvailableBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.AvailableBalanceFor(&_SimpleSwap.CallOpts, beneficiary)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_SimpleSwap *SimpleSwapCaller) HardDeposits(opts *bind.CallOpts, arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	DecreaseTimeout  *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	ret := new(struct {
		Amount           *big.Int
		DecreaseAmount   *big.Int
		DecreaseTimeout  *big.Int
		CanBeDecreasedAt *big.Int
	})
	out := ret
	err := _SimpleSwap.contract.Call(opts, out, "hardDeposits", arg0)
	return *ret, err
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_SimpleSwap *SimpleSwapSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	DecreaseTimeout  *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _SimpleSwap.Contract.HardDeposits(&_SimpleSwap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_SimpleSwap *SimpleSwapCallerSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	DecreaseTimeout  *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _SimpleSwap.Contract.HardDeposits(&_SimpleSwap.CallOpts, arg0)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_SimpleSwap *SimpleSwapCaller) Issuer(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "issuer")
	return *ret0, err
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_SimpleSwap *SimpleSwapSession) Issuer() (common.Address, error) {
	return _SimpleSwap.Contract.Issuer(&_SimpleSwap.CallOpts)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_SimpleSwap *SimpleSwapCallerSession) Issuer() (common.Address, error) {
	return _SimpleSwap.Contract.Issuer(&_SimpleSwap.CallOpts)
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

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) PaidOut(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "paidOut", arg0)
	return *ret0, err
}

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) PaidOut(arg0 common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.PaidOut(&_SimpleSwap.CallOpts, arg0)
}

// PaidOut is a free data retrieval call binding the contract method 0x81f03fcb.
//
// Solidity: function paidOut(address ) constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) PaidOut(arg0 common.Address) (*big.Int, error) {
	return _SimpleSwap.Contract.PaidOut(&_SimpleSwap.CallOpts, arg0)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) TotalHardDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "totalHardDeposit")
	return *ret0, err
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) TotalHardDeposit() (*big.Int, error) {
	return _SimpleSwap.Contract.TotalHardDeposit(&_SimpleSwap.CallOpts)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) TotalHardDeposit() (*big.Int, error) {
	return _SimpleSwap.Contract.TotalHardDeposit(&_SimpleSwap.CallOpts)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashCheque(opts *bind.TransactOpts, beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "cashCheque", beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapSession) CashCheque(beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashCheque is a paid mutator transaction binding the contract method 0x1633fb1d.
//
// Solidity: function cashCheque(address beneficiary, address recipient, uint256 cumulativePayout, bytes beneficiarySig, uint256 callerPayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashCheque(beneficiary common.Address, recipient common.Address, cumulativePayout *big.Int, beneficiarySig []byte, callerPayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiary, recipient, cumulativePayout, beneficiarySig, callerPayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashChequeBeneficiary(opts *bind.TransactOpts, recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "cashChequeBeneficiary", recipient, cumulativePayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapSession) CashChequeBeneficiary(recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeBeneficiary(&_SimpleSwap.TransactOpts, recipient, cumulativePayout, issuerSig)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x0d5f2659.
//
// Solidity: function cashChequeBeneficiary(address recipient, uint256 cumulativePayout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashChequeBeneficiary(recipient common.Address, cumulativePayout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeBeneficiary(&_SimpleSwap.TransactOpts, recipient, cumulativePayout, issuerSig)
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
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_SimpleSwap *SimpleSwapTransactor) PrepareDecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "prepareDecreaseHardDeposit", beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_SimpleSwap *SimpleSwapSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.PrepareDecreaseHardDeposit(&_SimpleSwap.TransactOpts, beneficiary, decreaseAmount)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0xdf324380.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SetCustomHardDepositDecreaseTimeout(opts *bind.TransactOpts, beneficiary common.Address, decreaseTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "setCustomHardDepositDecreaseTimeout", beneficiary, decreaseTimeout, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0xdf324380.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SetCustomHardDepositDecreaseTimeout(&_SimpleSwap.TransactOpts, beneficiary, decreaseTimeout, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0xdf324380.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SetCustomHardDepositDecreaseTimeout(&_SimpleSwap.TransactOpts, beneficiary, decreaseTimeout, beneficiarySig)
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
	Raw types.Log // Blockchain specific contextual infos
}

// FilterChequeBounced is a free log retrieval operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeBounced(opts *bind.FilterOpts) (*SimpleSwapChequeBouncedIterator, error) {

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeBounced")
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeBouncedIterator{contract: _SimpleSwap.contract, event: "ChequeBounced", logs: logs, sub: sub}, nil
}

// WatchChequeBounced is a free log subscription operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeBounced(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeBounced) (event.Subscription, error) {

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeBounced")
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

// ParseChequeBounced is a log parse operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_SimpleSwap *SimpleSwapFilterer) ParseChequeBounced(log types.Log) (*SimpleSwapChequeBounced, error) {
	event := new(SimpleSwapChequeBounced)
	if err := _SimpleSwap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
		return nil, err
	}
	return event, nil
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
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeCashed(opts *bind.FilterOpts, beneficiary []common.Address, recipient []common.Address, caller []common.Address) (*SimpleSwapChequeCashedIterator, error) {

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

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeCashed", beneficiaryRule, recipientRule, callerRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeCashedIterator{contract: _SimpleSwap.contract, event: "ChequeCashed", logs: logs, sub: sub}, nil
}

// WatchChequeCashed is a free log subscription operation binding the contract event 0x950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f.
//
// Solidity: event ChequeCashed(address indexed beneficiary, address indexed recipient, address indexed caller, uint256 totalPayout, uint256 cumulativePayout, uint256 callerPayout)
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeCashed(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeCashed, beneficiary []common.Address, recipient []common.Address, caller []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeCashed", beneficiaryRule, recipientRule, callerRule)
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

// ParseChequeCashed is a log parse operation binding the contract event 0x950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f.
//
// Solidity: event ChequeCashed(address indexed beneficiary, address indexed recipient, address indexed caller, uint256 totalPayout, uint256 cumulativePayout, uint256 callerPayout)
func (_SimpleSwap *SimpleSwapFilterer) ParseChequeCashed(log types.Log) (*SimpleSwapChequeCashed, error) {
	event := new(SimpleSwapChequeCashed)
	if err := _SimpleSwap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
		return nil, err
	}
	return event, nil
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

// ParseDeposit is a log parse operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address depositor, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) ParseDeposit(log types.Log) (*SimpleSwapDeposit, error) {
	event := new(SimpleSwapDeposit)
	if err := _SimpleSwap.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSwapHardDepositAmountChangedIterator is returned from FilterHardDepositAmountChanged and is used to iterate over the raw logs and unpacked data for HardDepositAmountChanged events raised by the SimpleSwap contract.
type SimpleSwapHardDepositAmountChangedIterator struct {
	Event *SimpleSwapHardDepositAmountChanged // Event containing the contract specifics and raw log

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
func (it *SimpleSwapHardDepositAmountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapHardDepositAmountChanged)
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
		it.Event = new(SimpleSwapHardDepositAmountChanged)
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
func (it *SimpleSwapHardDepositAmountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapHardDepositAmountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapHardDepositAmountChanged represents a HardDepositAmountChanged event raised by the SimpleSwap contract.
type SimpleSwapHardDepositAmountChanged struct {
	Beneficiary common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositAmountChanged is a free log retrieval operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) FilterHardDepositAmountChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*SimpleSwapHardDepositAmountChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapHardDepositAmountChangedIterator{contract: _SimpleSwap.contract, event: "HardDepositAmountChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositAmountChanged is a free log subscription operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) WatchHardDepositAmountChanged(opts *bind.WatchOpts, sink chan<- *SimpleSwapHardDepositAmountChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapHardDepositAmountChanged)
				if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
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
func (_SimpleSwap *SimpleSwapFilterer) ParseHardDepositAmountChanged(log types.Log) (*SimpleSwapHardDepositAmountChanged, error) {
	event := new(SimpleSwapHardDepositAmountChanged)
	if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
		return nil, err
	}
	return event, nil
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
	Beneficiary    common.Address
	DecreaseAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreasePrepared is a free log retrieval operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
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
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
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

// ParseHardDepositDecreasePrepared is a log parse operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_SimpleSwap *SimpleSwapFilterer) ParseHardDepositDecreasePrepared(log types.Log) (*SimpleSwapHardDepositDecreasePrepared, error) {
	event := new(SimpleSwapHardDepositDecreasePrepared)
	if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSwapHardDepositDecreaseTimeoutChangedIterator is returned from FilterHardDepositDecreaseTimeoutChanged and is used to iterate over the raw logs and unpacked data for HardDepositDecreaseTimeoutChanged events raised by the SimpleSwap contract.
type SimpleSwapHardDepositDecreaseTimeoutChangedIterator struct {
	Event *SimpleSwapHardDepositDecreaseTimeoutChanged // Event containing the contract specifics and raw log

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
func (it *SimpleSwapHardDepositDecreaseTimeoutChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapHardDepositDecreaseTimeoutChanged)
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
		it.Event = new(SimpleSwapHardDepositDecreaseTimeoutChanged)
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
func (it *SimpleSwapHardDepositDecreaseTimeoutChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapHardDepositDecreaseTimeoutChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapHardDepositDecreaseTimeoutChanged represents a HardDepositDecreaseTimeoutChanged event raised by the SimpleSwap contract.
type SimpleSwapHardDepositDecreaseTimeoutChanged struct {
	Beneficiary     common.Address
	DecreaseTimeout *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreaseTimeoutChanged is a free log retrieval operation binding the contract event 0x86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f.
//
// Solidity: event HardDepositDecreaseTimeoutChanged(address indexed beneficiary, uint256 decreaseTimeout)
func (_SimpleSwap *SimpleSwapFilterer) FilterHardDepositDecreaseTimeoutChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*SimpleSwapHardDepositDecreaseTimeoutChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "HardDepositDecreaseTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapHardDepositDecreaseTimeoutChangedIterator{contract: _SimpleSwap.contract, event: "HardDepositDecreaseTimeoutChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositDecreaseTimeoutChanged is a free log subscription operation binding the contract event 0x86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f.
//
// Solidity: event HardDepositDecreaseTimeoutChanged(address indexed beneficiary, uint256 decreaseTimeout)
func (_SimpleSwap *SimpleSwapFilterer) WatchHardDepositDecreaseTimeoutChanged(opts *bind.WatchOpts, sink chan<- *SimpleSwapHardDepositDecreaseTimeoutChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "HardDepositDecreaseTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapHardDepositDecreaseTimeoutChanged)
				if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositDecreaseTimeoutChanged", log); err != nil {
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

// ParseHardDepositDecreaseTimeoutChanged is a log parse operation binding the contract event 0x86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f.
//
// Solidity: event HardDepositDecreaseTimeoutChanged(address indexed beneficiary, uint256 decreaseTimeout)
func (_SimpleSwap *SimpleSwapFilterer) ParseHardDepositDecreaseTimeoutChanged(log types.Log) (*SimpleSwapHardDepositDecreaseTimeoutChanged, error) {
	event := new(SimpleSwapHardDepositDecreaseTimeoutChanged)
	if err := _SimpleSwap.contract.UnpackLog(event, "HardDepositDecreaseTimeoutChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SimpleSwapWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the SimpleSwap contract.
type SimpleSwapWithdrawIterator struct {
	Event *SimpleSwapWithdraw // Event containing the contract specifics and raw log

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
func (it *SimpleSwapWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SimpleSwapWithdraw)
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
		it.Event = new(SimpleSwapWithdraw)
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
func (it *SimpleSwapWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SimpleSwapWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SimpleSwapWithdraw represents a Withdraw event raised by the SimpleSwap contract.
type SimpleSwapWithdraw struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d.
//
// Solidity: event Withdraw(uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) FilterWithdraw(opts *bind.FilterOpts) (*SimpleSwapWithdrawIterator, error) {

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return &SimpleSwapWithdrawIterator{contract: _SimpleSwap.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x5b6b431d4476a211bb7d41c20d1aab9ae2321deee0d20be3d9fc9b1093fa6e3d.
//
// Solidity: event Withdraw(uint256 amount)
func (_SimpleSwap *SimpleSwapFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *SimpleSwapWithdraw) (event.Subscription, error) {

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "Withdraw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SimpleSwapWithdraw)
				if err := _SimpleSwap.contract.UnpackLog(event, "Withdraw", log); err != nil {
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
func (_SimpleSwap *SimpleSwapFilterer) ParseWithdraw(log types.Log) (*SimpleSwapWithdraw, error) {
	event := new(SimpleSwapWithdraw)
	if err := _SimpleSwap.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	return event, nil
}
