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
var ECDSABin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820860ac272f1674c9a0d6cb877e11ae00f0555fef595f0f6cc567b25212f11f3fc64736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e62613932326537360058"

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
var MathBin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a7231582025e9b39ea6fbc3836058c5a199005288ee40422f8524bf8e4b5bb3704fb8017364736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e62613932326537360058"

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
var SafeMathBin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820d9e8e9055bcf685abca53350f2076132904c3ea258ef90aa88c9675d84b8019a64736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e62613932326537360058"

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
const SimpleSwapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"swap\",\"type\":\"address\"},{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"chequeHash\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"issuer\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"}],\"name\":\"cashChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"},{\"name\":\"beneficarySig\",\"type\":\"bytes\"}],\"name\":\"submitCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"submitChequeissuer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"setCustomHardDepositDecreaseTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cheques\",\"outputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"paidOut\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"canBeDecreasedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"liquidBalanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalHardDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"_cashChequeInternal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"submitChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_issuer\",\"type\":\"address\"},{\"name\":\"defaultHardDepositTimeoutDuration\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"callee\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"ChequeSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositAmountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseTimeout\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreaseTimeoutChanged\",\"type\":\"event\"}]"

// SimpleSwapBin is the compiled bytecode used for deploying new contracts.
var SimpleSwapBin = "0x608060405260405161283b38038061283b8339818101604052604081101561002657600080fd5b8101908080519060200190929190805190602001909291905050506000811461004f5780610054565b620151805b60008190555081600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000341115610110577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b505061271a806101216000396000f3fe6080604052600436106101145760003560e01c80636162913b116100a0578063c76a4d3111610064578063c76a4d3114610998578063e0bcf13a146109fd578063e3bb7aec14610a28578063f3c08b1f14610b4e578063f890673b14610bd357610114565b80636162913b146107cd578063946f46a214610847578063b6343b0d14610898578063b777035014610912578063b7ec1a331461096d57610114565b8063338f3fed116100e7578063338f3fed1461031b57806339d9ec4c146103765780634f823a4c146103a157806354fe26141461053e5780635cb189471461064457610114565b8063030aca3e1461018b5780631d1438481461022e5780632329d2a8146102855780632e1a7d4d146102e0575b6000341115610189577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b005b34801561019757600080fd5b50610218600480360360a08110156101ae57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050610cb9565b6040518082815260200191505060405180910390f35b34801561023a57600080fd5b50610243610d6b565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561029157600080fd5b506102de600480360360408110156102a857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610d91565b005b3480156102ec57600080fd5b506103196004803603602081101561030357600080fd5b8101908080359060200190929190505050610da2565b005b34801561032757600080fd5b506103746004803603604081101561033e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610f31565b005b34801561038257600080fd5b5061038b611156565b6040518082815260200191505060405180910390f35b3480156103ad57600080fd5b5061053c600480360360c08110156103c457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190803590602001909291908035906020019064010000000081111561041f57600080fd5b82018360208201111561043157600080fd5b8035906020019184600183028401116401000000008311171561045357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001906401000000008111156104b657600080fd5b8201836020820111156104c857600080fd5b803590602001918460018302840111640100000000831117156104ea57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061115c565b005b34801561054a57600080fd5b50610642600480360360a081101561056157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190803590602001906401000000008111156105bc57600080fd5b8201836020820111156105ce57600080fd5b803590602001918460018302840111640100000000831117156105f057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506112e1565b005b34801561065057600080fd5b506107cb6004803603608081101561066757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156106ae57600080fd5b8201836020820111156106c057600080fd5b803590602001918460018302840111640100000000831117156106e257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192908035906020019064010000000081111561074557600080fd5b82018360208201111561075757600080fd5b8035906020019184600183028401116401000000008311171561077957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050611450565b005b3480156107d957600080fd5b5061081c600480360360208110156107f057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506116ed565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b34801561085357600080fd5b506108966004803603602081101561086a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061171d565b005b3480156108a457600080fd5b506108e7600480360360208110156108bb57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611870565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b34801561091e57600080fd5b5061096b6004803603604081101561093557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506118a0565b005b34801561097957600080fd5b50610982611a88565b6040518082815260200191505060405180910390f35b3480156109a457600080fd5b506109e7600480360360208110156109bb57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611abb565b6040518082815260200191505060405180910390f35b348015610a0957600080fd5b50610a12611b20565b6040518082815260200191505060405180910390f35b348015610a3457600080fd5b50610b4c600480360360c0811015610a4b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190640100000000811115610ab257600080fd5b820183602082011115610ac457600080fd5b80359060200191846001830284011164010000000083111715610ae657600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192908035906020019092919080359060200190929190505050611b26565b005b348015610b5a57600080fd5b50610bd160048036036080811015610b7157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190505050611d42565b005b348015610bdf57600080fd5b50610cb760048036036080811015610bf657600080fd5b8101908080359060200190929190803590602001909291908035906020019092919080359060200190640100000000811115610c3157600080fd5b820183602082011115610c4357600080fd5b80359060200191846001830284011164010000000083111715610c6557600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506120c8565b005b60008584868585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018381526020018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b610d9e3383836000611d42565b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610e65576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610e6d611a88565b811115610ec5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260288152602001806126756028913960400191505060405180910390fd5b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610f2d573d6000803e3d6000fd5b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610ff4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b3073ffffffffffffffffffffffffffffffffffffffff1631611021826003546121b290919063ffffffff16565b1115611078576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260358152602001806126406035913960400191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090506110d28282600001546121b290919063ffffffff16565b81600001819055506110ef826003546121b290919063ffffffff16565b600381905550600081600201819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60005481565b61117261116c3088888888610cb9565b8361223a565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611234576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b61124a6112443088888888610cb9565b8261223a565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff16146112cd576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602281526020018061258c6022913960400191505060405180910390fd5b6112d986868686612256565b505050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146113a4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6113ba6113b43087878787610cb9565b8261223a565b73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff161461143d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602281526020018061258c6022913960400191505060405180910390fd5b61144985858585612256565b5050505050565b6114e8308585604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019350505050604051602081830303815290604052805190602001208361223a565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461154157600080fd5b6115d9308585604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019350505050604051602081830303815290604052805190602001208261223a565b73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff161461161057600080fd5b82600260008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508373ffffffffffffffffffffffffffffffffffffffff167f86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f600260008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201546040518082815260200191505060405180910390a250505050565b60016020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015801561177957506000816003015414155b6117ce576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260258152602001806125ae6025913960400191505060405180910390fd5b6117e98160010154826000015461238d90919063ffffffff16565b816000018190555060008160030181905550611814816001015460035461238d90919063ffffffff16565b6003819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60026020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611963576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000154821115611a03576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260278152602001806125f76027913960400191505060405180910390fd5b600080826002015414611a1a578160020154611a1e565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b6000611ab66003543073ffffffffffffffffffffffffffffffffffffffff163161238d90919063ffffffff16565b905090565b6000611b19600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611b0b611a88565b6121b290919063ffffffff16565b9050919050565b60035481565b81421115611b7f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602281526020018061261e6022913960400191505060405180910390fd5b611c64303386888686604051602001808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018381526020018281526020019650505050505050604051602081830303815290604052805190602001208461223a565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff1614611ce7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602281526020018061258c6022913960400191505060405180910390fd5b611cf386868684611d42565b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611d39573d6000803e3d6000fd5b50505050505050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015611de2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260248152602001806125d36024913960400191505060405180910390fd5b611dfd8160020154826001015461238d90919063ffffffff16565b831115611e55576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602381526020018061269d6023913960400191505060405180910390fd5b6000611ea384600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154612416565b90506000611eba8583611eb4611a88565b01612416565b905060008214611f7b57611f1982600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000015461238d90919063ffffffff16565b600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000181905550611f748260035461238d90919063ffffffff16565b6003819055505b611f928184600201546121b290919063ffffffff16565b83600201819055508573ffffffffffffffffffffffffffffffffffffffff166108fc611fc7868461238d90919063ffffffff16565b9081150290604051600060405180830381858888f19350505050158015611ff2573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff167f5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed138660000154858a8a6040518085815260200184815260200183815260200182815260200194505050505060405180910390a48085146120bf577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b50505050505050565b6120de6120d83033878787610cb9565b8261223a565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16146121a0576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b6121ac33858585612256565b50505050565b600080828401905083811015612230576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b600061224e6122488461242f565b83612487565b905092915050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600001548411612312576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f53696d706c65537761703a20696e76616c69642073657269616c00000000000081525060200191505060405180910390fd5b8381600001819055508281600101819055508142018160030181905550838573ffffffffffffffffffffffffffffffffffffffff167f543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c8585604051808381526020018281526020019250505060405180910390a35050505050565b600082821115612405576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525060200191505060405180910390fd5b600082840390508091505092915050565b60008183106124255781612427565b825b905092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b6000604182511461249b5760009050612585565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c11156124ef5760009350505050612585565b601b8160ff16141580156125075750601c8160ff1614155b156125185760009350505050612585565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa158015612575573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fe53696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a20636865717565206e6f74207965742074696d6564206f757453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a2062656e6566696369617279536967206578706972656453696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e63652053696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e7453696d706c65537761703a206e6f7420656e6f7567682062616c616e6365206f776564a265627a7a72315820223fbdb089d991791d63f936251c28fd112f5d03e761a80be31ccc034df9210b64736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e62613932326537360058"

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

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCaller) DEFAULTHARDDEPPOSITDECREASETIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT")
	return *ret0, err
}

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapSession) DEFAULTHARDDEPPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _SimpleSwap.Contract.DEFAULTHARDDEPPOSITDECREASETIMEOUT(&_SimpleSwap.CallOpts)
}

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_SimpleSwap *SimpleSwapCallerSession) DEFAULTHARDDEPPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _SimpleSwap.Contract.DEFAULTHARDDEPPOSITDECREASETIMEOUT(&_SimpleSwap.CallOpts)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapCaller) ChequeHash(opts *bind.CallOpts, swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _SimpleSwap.contract.Call(opts, out, "chequeHash", swap, beneficiary, serial, amount, cashTimeout)
	return *ret0, err
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	return _SimpleSwap.Contract.ChequeHash(&_SimpleSwap.CallOpts, swap, beneficiary, serial, amount, cashTimeout)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_SimpleSwap *SimpleSwapCallerSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	return _SimpleSwap.Contract.ChequeHash(&_SimpleSwap.CallOpts, swap, beneficiary, serial, amount, cashTimeout)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_SimpleSwap *SimpleSwapCaller) Cheques(opts *bind.CallOpts, arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	ret := new(struct {
		Serial      *big.Int
		Amount      *big.Int
		PaidOut     *big.Int
		CashTimeout *big.Int
	})
	out := ret
	err := _SimpleSwap.contract.Call(opts, out, "cheques", arg0)
	return *ret, err
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_SimpleSwap *SimpleSwapSession) Cheques(arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	return _SimpleSwap.Contract.Cheques(&_SimpleSwap.CallOpts, arg0)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_SimpleSwap *SimpleSwapCallerSession) Cheques(arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	return _SimpleSwap.Contract.Cheques(&_SimpleSwap.CallOpts, arg0)
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

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashChequeInternal(opts *bind.TransactOpts, beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "_cashChequeInternal", beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapSession) CashChequeInternal(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeInternal(&_SimpleSwap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashChequeInternal(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeInternal(&_SimpleSwap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashCheque(opts *bind.TransactOpts, beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "cashCheque", beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapSession) CashCheque(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashCheque(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashCheque(&_SimpleSwap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_SimpleSwap *SimpleSwapTransactor) CashChequeBeneficiary(opts *bind.TransactOpts, beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "cashChequeBeneficiary", beneficiaryAgent, requestPayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_SimpleSwap *SimpleSwapSession) CashChequeBeneficiary(beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeBeneficiary(&_SimpleSwap.TransactOpts, beneficiaryAgent, requestPayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) CashChequeBeneficiary(beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _SimpleSwap.Contract.CashChequeBeneficiary(&_SimpleSwap.TransactOpts, beneficiaryAgent, requestPayout)
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

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SetCustomHardDepositDecreaseTimeout(opts *bind.TransactOpts, beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "setCustomHardDepositDecreaseTimeout", beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SetCustomHardDepositDecreaseTimeout(&_SimpleSwap.TransactOpts, beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SetCustomHardDepositDecreaseTimeout(&_SimpleSwap.TransactOpts, beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitCheque(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitCheque", beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitCheque(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitCheque(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitChequeBeneficiary(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitChequeBeneficiary", serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeBeneficiary(&_SimpleSwap.TransactOpts, serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeBeneficiary(&_SimpleSwap.TransactOpts, serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactor) SubmitChequeissuer(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.contract.Transact(opts, "submitChequeissuer", beneficiary, serial, amount, cashTimeout, beneficiarySig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapSession) SubmitChequeissuer(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeissuer(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, cashTimeout, beneficiarySig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_SimpleSwap *SimpleSwapTransactorSession) SubmitChequeissuer(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _SimpleSwap.Contract.SubmitChequeissuer(&_SimpleSwap.TransactOpts, beneficiary, serial, amount, cashTimeout, beneficiarySig)
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
	BeneficiaryPrincipal common.Address
	BeneficiaryAgent     common.Address
	Callee               common.Address
	Serial               *big.Int
	TotalPayout          *big.Int
	RequestPayout        *big.Int
	CalleePayout         *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterChequeCashed is a free log retrieval operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_SimpleSwap *SimpleSwapFilterer) FilterChequeCashed(opts *bind.FilterOpts, beneficiaryPrincipal []common.Address, beneficiaryAgent []common.Address, callee []common.Address) (*SimpleSwapChequeCashedIterator, error) {

	var beneficiaryPrincipalRule []interface{}
	for _, beneficiaryPrincipalItem := range beneficiaryPrincipal {
		beneficiaryPrincipalRule = append(beneficiaryPrincipalRule, beneficiaryPrincipalItem)
	}
	var beneficiaryAgentRule []interface{}
	for _, beneficiaryAgentItem := range beneficiaryAgent {
		beneficiaryAgentRule = append(beneficiaryAgentRule, beneficiaryAgentItem)
	}
	var calleeRule []interface{}
	for _, calleeItem := range callee {
		calleeRule = append(calleeRule, calleeItem)
	}

	logs, sub, err := _SimpleSwap.contract.FilterLogs(opts, "ChequeCashed", beneficiaryPrincipalRule, beneficiaryAgentRule, calleeRule)
	if err != nil {
		return nil, err
	}
	return &SimpleSwapChequeCashedIterator{contract: _SimpleSwap.contract, event: "ChequeCashed", logs: logs, sub: sub}, nil
}

// WatchChequeCashed is a free log subscription operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_SimpleSwap *SimpleSwapFilterer) WatchChequeCashed(opts *bind.WatchOpts, sink chan<- *SimpleSwapChequeCashed, beneficiaryPrincipal []common.Address, beneficiaryAgent []common.Address, callee []common.Address) (event.Subscription, error) {

	var beneficiaryPrincipalRule []interface{}
	for _, beneficiaryPrincipalItem := range beneficiaryPrincipal {
		beneficiaryPrincipalRule = append(beneficiaryPrincipalRule, beneficiaryPrincipalItem)
	}
	var beneficiaryAgentRule []interface{}
	for _, beneficiaryAgentItem := range beneficiaryAgent {
		beneficiaryAgentRule = append(beneficiaryAgentRule, beneficiaryAgentItem)
	}
	var calleeRule []interface{}
	for _, calleeItem := range callee {
		calleeRule = append(calleeRule, calleeItem)
	}

	logs, sub, err := _SimpleSwap.contract.WatchLogs(opts, "ChequeCashed", beneficiaryPrincipalRule, beneficiaryAgentRule, calleeRule)
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

// ParseChequeCashed is a log parse operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_SimpleSwap *SimpleSwapFilterer) ParseChequeCashed(log types.Log) (*SimpleSwapChequeCashed, error) {
	event := new(SimpleSwapChequeCashed)
	if err := _SimpleSwap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
		return nil, err
	}
	return event, nil
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
	CashTimeout *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChequeSubmitted is a free log retrieval operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
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
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
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

// ParseChequeSubmitted is a log parse operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
func (_SimpleSwap *SimpleSwapFilterer) ParseChequeSubmitted(log types.Log) (*SimpleSwapChequeSubmitted, error) {
	event := new(SimpleSwapChequeSubmitted)
	if err := _SimpleSwap.contract.UnpackLog(event, "ChequeSubmitted", log); err != nil {
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

// SwapABI is the input ABI used to generate the binding from.
const SwapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"swap\",\"type\":\"address\"},{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"chequeHash\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"issuer\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"}],\"name\":\"cashChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"},{\"name\":\"beneficarySig\",\"type\":\"bytes\"}],\"name\":\"submitCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"submitChequeissuer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"setCustomHardDepositDecreaseTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cheques\",\"outputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"paidOut\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"canBeDecreasedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"liquidBalanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalHardDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"_cashChequeInternal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"submitChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"callee\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"ChequeSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositAmountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseTimeout\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreaseTimeoutChanged\",\"type\":\"event\"}]"

// Swap is an auto generated Go binding around an Ethereum contract.
type Swap struct {
	SwapCaller     // Read-only binding to the contract
	SwapTransactor // Write-only binding to the contract
	SwapFilterer   // Log filterer for contract events
}

// SwapCaller is an auto generated read-only Go binding around an Ethereum contract.
type SwapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SwapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SwapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SwapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SwapSession struct {
	Contract     *Swap             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SwapCallerSession struct {
	Contract *SwapCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SwapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SwapTransactorSession struct {
	Contract     *SwapTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SwapRaw is an auto generated low-level Go binding around an Ethereum contract.
type SwapRaw struct {
	Contract *Swap // Generic contract binding to access the raw methods on
}

// SwapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SwapCallerRaw struct {
	Contract *SwapCaller // Generic read-only contract binding to access the raw methods on
}

// SwapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SwapTransactorRaw struct {
	Contract *SwapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSwap creates a new instance of Swap, bound to a specific deployed contract.
func NewSwap(address common.Address, backend bind.ContractBackend) (*Swap, error) {
	contract, err := bindSwap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Swap{SwapCaller: SwapCaller{contract: contract}, SwapTransactor: SwapTransactor{contract: contract}, SwapFilterer: SwapFilterer{contract: contract}}, nil
}

// NewSwapCaller creates a new read-only instance of Swap, bound to a specific deployed contract.
func NewSwapCaller(address common.Address, caller bind.ContractCaller) (*SwapCaller, error) {
	contract, err := bindSwap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SwapCaller{contract: contract}, nil
}

// NewSwapTransactor creates a new write-only instance of Swap, bound to a specific deployed contract.
func NewSwapTransactor(address common.Address, transactor bind.ContractTransactor) (*SwapTransactor, error) {
	contract, err := bindSwap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SwapTransactor{contract: contract}, nil
}

// NewSwapFilterer creates a new log filterer instance of Swap, bound to a specific deployed contract.
func NewSwapFilterer(address common.Address, filterer bind.ContractFilterer) (*SwapFilterer, error) {
	contract, err := bindSwap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SwapFilterer{contract: contract}, nil
}

// bindSwap binds a generic wrapper to an already deployed contract.
func bindSwap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SwapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swap *SwapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Swap.Contract.SwapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swap *SwapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.Contract.SwapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swap *SwapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swap.Contract.SwapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swap *SwapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Swap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swap *SwapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swap *SwapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swap.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_Swap *SwapCaller) DEFAULTHARDDEPPOSITDECREASETIMEOUT(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT")
	return *ret0, err
}

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_Swap *SwapSession) DEFAULTHARDDEPPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _Swap.Contract.DEFAULTHARDDEPPOSITDECREASETIMEOUT(&_Swap.CallOpts)
}

// DEFAULTHARDDEPPOSITDECREASETIMEOUT is a free data retrieval call binding the contract method 0x39d9ec4c.
//
// Solidity: function DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT() constant returns(uint256)
func (_Swap *SwapCallerSession) DEFAULTHARDDEPPOSITDECREASETIMEOUT() (*big.Int, error) {
	return _Swap.Contract.DEFAULTHARDDEPPOSITDECREASETIMEOUT(&_Swap.CallOpts)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_Swap *SwapCaller) ChequeHash(opts *bind.CallOpts, swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "chequeHash", swap, beneficiary, serial, amount, cashTimeout)
	return *ret0, err
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_Swap *SwapSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	return _Swap.Contract.ChequeHash(&_Swap.CallOpts, swap, beneficiary, serial, amount, cashTimeout)
}

// ChequeHash is a free data retrieval call binding the contract method 0x030aca3e.
//
// Solidity: function chequeHash(address swap, address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout) constant returns(bytes32)
func (_Swap *SwapCallerSession) ChequeHash(swap common.Address, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int) ([32]byte, error) {
	return _Swap.Contract.ChequeHash(&_Swap.CallOpts, swap, beneficiary, serial, amount, cashTimeout)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_Swap *SwapCaller) Cheques(opts *bind.CallOpts, arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	ret := new(struct {
		Serial      *big.Int
		Amount      *big.Int
		PaidOut     *big.Int
		CashTimeout *big.Int
	})
	out := ret
	err := _Swap.contract.Call(opts, out, "cheques", arg0)
	return *ret, err
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_Swap *SwapSession) Cheques(arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	return _Swap.Contract.Cheques(&_Swap.CallOpts, arg0)
}

// Cheques is a free data retrieval call binding the contract method 0x6162913b.
//
// Solidity: function cheques(address ) constant returns(uint256 serial, uint256 amount, uint256 paidOut, uint256 cashTimeout)
func (_Swap *SwapCallerSession) Cheques(arg0 common.Address) (struct {
	Serial      *big.Int
	Amount      *big.Int
	PaidOut     *big.Int
	CashTimeout *big.Int
}, error) {
	return _Swap.Contract.Cheques(&_Swap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_Swap *SwapCaller) HardDeposits(opts *bind.CallOpts, arg0 common.Address) (struct {
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
	err := _Swap.contract.Call(opts, out, "hardDeposits", arg0)
	return *ret, err
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_Swap *SwapSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	DecreaseTimeout  *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _Swap.Contract.HardDeposits(&_Swap.CallOpts, arg0)
}

// HardDeposits is a free data retrieval call binding the contract method 0xb6343b0d.
//
// Solidity: function hardDeposits(address ) constant returns(uint256 amount, uint256 decreaseAmount, uint256 decreaseTimeout, uint256 canBeDecreasedAt)
func (_Swap *SwapCallerSession) HardDeposits(arg0 common.Address) (struct {
	Amount           *big.Int
	DecreaseAmount   *big.Int
	DecreaseTimeout  *big.Int
	CanBeDecreasedAt *big.Int
}, error) {
	return _Swap.Contract.HardDeposits(&_Swap.CallOpts, arg0)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_Swap *SwapCaller) Issuer(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "issuer")
	return *ret0, err
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_Swap *SwapSession) Issuer() (common.Address, error) {
	return _Swap.Contract.Issuer(&_Swap.CallOpts)
}

// Issuer is a free data retrieval call binding the contract method 0x1d143848.
//
// Solidity: function issuer() constant returns(address)
func (_Swap *SwapCallerSession) Issuer() (common.Address, error) {
	return _Swap.Contract.Issuer(&_Swap.CallOpts)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_Swap *SwapCaller) LiquidBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "liquidBalance")
	return *ret0, err
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_Swap *SwapSession) LiquidBalance() (*big.Int, error) {
	return _Swap.Contract.LiquidBalance(&_Swap.CallOpts)
}

// LiquidBalance is a free data retrieval call binding the contract method 0xb7ec1a33.
//
// Solidity: function liquidBalance() constant returns(uint256)
func (_Swap *SwapCallerSession) LiquidBalance() (*big.Int, error) {
	return _Swap.Contract.LiquidBalance(&_Swap.CallOpts)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_Swap *SwapCaller) LiquidBalanceFor(opts *bind.CallOpts, beneficiary common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "liquidBalanceFor", beneficiary)
	return *ret0, err
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_Swap *SwapSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _Swap.Contract.LiquidBalanceFor(&_Swap.CallOpts, beneficiary)
}

// LiquidBalanceFor is a free data retrieval call binding the contract method 0xc76a4d31.
//
// Solidity: function liquidBalanceFor(address beneficiary) constant returns(uint256)
func (_Swap *SwapCallerSession) LiquidBalanceFor(beneficiary common.Address) (*big.Int, error) {
	return _Swap.Contract.LiquidBalanceFor(&_Swap.CallOpts, beneficiary)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_Swap *SwapCaller) TotalHardDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Swap.contract.Call(opts, out, "totalHardDeposit")
	return *ret0, err
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_Swap *SwapSession) TotalHardDeposit() (*big.Int, error) {
	return _Swap.Contract.TotalHardDeposit(&_Swap.CallOpts)
}

// TotalHardDeposit is a free data retrieval call binding the contract method 0xe0bcf13a.
//
// Solidity: function totalHardDeposit() constant returns(uint256)
func (_Swap *SwapCallerSession) TotalHardDeposit() (*big.Int, error) {
	return _Swap.Contract.TotalHardDeposit(&_Swap.CallOpts)
}

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_Swap *SwapTransactor) CashChequeInternal(opts *bind.TransactOpts, beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "_cashChequeInternal", beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_Swap *SwapSession) CashChequeInternal(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashChequeInternal(&_Swap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashChequeInternal is a paid mutator transaction binding the contract method 0xf3c08b1f.
//
// Solidity: function _cashChequeInternal(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, uint256 calleePayout) returns()
func (_Swap *SwapTransactorSession) CashChequeInternal(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashChequeInternal(&_Swap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_Swap *SwapTransactor) CashCheque(opts *bind.TransactOpts, beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "cashCheque", beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_Swap *SwapSession) CashCheque(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashCheque(&_Swap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashCheque is a paid mutator transaction binding the contract method 0xe3bb7aec.
//
// Solidity: function cashCheque(address beneficiaryPrincipal, address beneficiaryAgent, uint256 requestPayout, bytes beneficiarySig, uint256 expiry, uint256 calleePayout) returns()
func (_Swap *SwapTransactorSession) CashCheque(beneficiaryPrincipal common.Address, beneficiaryAgent common.Address, requestPayout *big.Int, beneficiarySig []byte, expiry *big.Int, calleePayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashCheque(&_Swap.TransactOpts, beneficiaryPrincipal, beneficiaryAgent, requestPayout, beneficiarySig, expiry, calleePayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_Swap *SwapTransactor) CashChequeBeneficiary(opts *bind.TransactOpts, beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "cashChequeBeneficiary", beneficiaryAgent, requestPayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_Swap *SwapSession) CashChequeBeneficiary(beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashChequeBeneficiary(&_Swap.TransactOpts, beneficiaryAgent, requestPayout)
}

// CashChequeBeneficiary is a paid mutator transaction binding the contract method 0x2329d2a8.
//
// Solidity: function cashChequeBeneficiary(address beneficiaryAgent, uint256 requestPayout) returns()
func (_Swap *SwapTransactorSession) CashChequeBeneficiary(beneficiaryAgent common.Address, requestPayout *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.CashChequeBeneficiary(&_Swap.TransactOpts, beneficiaryAgent, requestPayout)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_Swap *SwapTransactor) DecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "decreaseHardDeposit", beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_Swap *SwapSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _Swap.Contract.DecreaseHardDeposit(&_Swap.TransactOpts, beneficiary)
}

// DecreaseHardDeposit is a paid mutator transaction binding the contract method 0x946f46a2.
//
// Solidity: function decreaseHardDeposit(address beneficiary) returns()
func (_Swap *SwapTransactorSession) DecreaseHardDeposit(beneficiary common.Address) (*types.Transaction, error) {
	return _Swap.Contract.DecreaseHardDeposit(&_Swap.TransactOpts, beneficiary)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_Swap *SwapTransactor) IncreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "increaseHardDeposit", beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_Swap *SwapSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.IncreaseHardDeposit(&_Swap.TransactOpts, beneficiary, amount)
}

// IncreaseHardDeposit is a paid mutator transaction binding the contract method 0x338f3fed.
//
// Solidity: function increaseHardDeposit(address beneficiary, uint256 amount) returns()
func (_Swap *SwapTransactorSession) IncreaseHardDeposit(beneficiary common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.IncreaseHardDeposit(&_Swap.TransactOpts, beneficiary, amount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_Swap *SwapTransactor) PrepareDecreaseHardDeposit(opts *bind.TransactOpts, beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "prepareDecreaseHardDeposit", beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_Swap *SwapSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.PrepareDecreaseHardDeposit(&_Swap.TransactOpts, beneficiary, decreaseAmount)
}

// PrepareDecreaseHardDeposit is a paid mutator transaction binding the contract method 0xb7770350.
//
// Solidity: function prepareDecreaseHardDeposit(address beneficiary, uint256 decreaseAmount) returns()
func (_Swap *SwapTransactorSession) PrepareDecreaseHardDeposit(beneficiary common.Address, decreaseAmount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.PrepareDecreaseHardDeposit(&_Swap.TransactOpts, beneficiary, decreaseAmount)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_Swap *SwapTransactor) SetCustomHardDepositDecreaseTimeout(opts *bind.TransactOpts, beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "setCustomHardDepositDecreaseTimeout", beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_Swap *SwapSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SetCustomHardDepositDecreaseTimeout(&_Swap.TransactOpts, beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SetCustomHardDepositDecreaseTimeout is a paid mutator transaction binding the contract method 0x5cb18947.
//
// Solidity: function setCustomHardDepositDecreaseTimeout(address beneficiary, uint256 decreaseTimeout, bytes issuerSig, bytes beneficiarySig) returns()
func (_Swap *SwapTransactorSession) SetCustomHardDepositDecreaseTimeout(beneficiary common.Address, decreaseTimeout *big.Int, issuerSig []byte, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SetCustomHardDepositDecreaseTimeout(&_Swap.TransactOpts, beneficiary, decreaseTimeout, issuerSig, beneficiarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_Swap *SwapTransactor) SubmitCheque(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "submitCheque", beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_Swap *SwapSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitCheque(&_Swap.TransactOpts, beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitCheque is a paid mutator transaction binding the contract method 0x4f823a4c.
//
// Solidity: function submitCheque(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig, bytes beneficarySig) returns()
func (_Swap *SwapTransactorSession) SubmitCheque(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte, beneficarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitCheque(&_Swap.TransactOpts, beneficiary, serial, amount, cashTimeout, issuerSig, beneficarySig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_Swap *SwapTransactor) SubmitChequeBeneficiary(opts *bind.TransactOpts, serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "submitChequeBeneficiary", serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_Swap *SwapSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitChequeBeneficiary(&_Swap.TransactOpts, serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeBeneficiary is a paid mutator transaction binding the contract method 0xf890673b.
//
// Solidity: function submitChequeBeneficiary(uint256 serial, uint256 amount, uint256 cashTimeout, bytes issuerSig) returns()
func (_Swap *SwapTransactorSession) SubmitChequeBeneficiary(serial *big.Int, amount *big.Int, cashTimeout *big.Int, issuerSig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitChequeBeneficiary(&_Swap.TransactOpts, serial, amount, cashTimeout, issuerSig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_Swap *SwapTransactor) SubmitChequeissuer(opts *bind.TransactOpts, beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "submitChequeissuer", beneficiary, serial, amount, cashTimeout, beneficiarySig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_Swap *SwapSession) SubmitChequeissuer(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitChequeissuer(&_Swap.TransactOpts, beneficiary, serial, amount, cashTimeout, beneficiarySig)
}

// SubmitChequeissuer is a paid mutator transaction binding the contract method 0x54fe2614.
//
// Solidity: function submitChequeissuer(address beneficiary, uint256 serial, uint256 amount, uint256 cashTimeout, bytes beneficiarySig) returns()
func (_Swap *SwapTransactorSession) SubmitChequeissuer(beneficiary common.Address, serial *big.Int, amount *big.Int, cashTimeout *big.Int, beneficiarySig []byte) (*types.Transaction, error) {
	return _Swap.Contract.SubmitChequeissuer(&_Swap.TransactOpts, beneficiary, serial, amount, cashTimeout, beneficiarySig)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_Swap *SwapTransactor) Withdraw(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Swap.contract.Transact(opts, "withdraw", amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_Swap *SwapSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Withdraw(&_Swap.TransactOpts, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 amount) returns()
func (_Swap *SwapTransactorSession) Withdraw(amount *big.Int) (*types.Transaction, error) {
	return _Swap.Contract.Withdraw(&_Swap.TransactOpts, amount)
}

// SwapChequeBouncedIterator is returned from FilterChequeBounced and is used to iterate over the raw logs and unpacked data for ChequeBounced events raised by the Swap contract.
type SwapChequeBouncedIterator struct {
	Event *SwapChequeBounced // Event containing the contract specifics and raw log

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
func (it *SwapChequeBouncedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapChequeBounced)
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
		it.Event = new(SwapChequeBounced)
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
func (it *SwapChequeBouncedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapChequeBouncedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapChequeBounced represents a ChequeBounced event raised by the Swap contract.
type SwapChequeBounced struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterChequeBounced is a free log retrieval operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_Swap *SwapFilterer) FilterChequeBounced(opts *bind.FilterOpts) (*SwapChequeBouncedIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "ChequeBounced")
	if err != nil {
		return nil, err
	}
	return &SwapChequeBouncedIterator{contract: _Swap.contract, event: "ChequeBounced", logs: logs, sub: sub}, nil
}

// WatchChequeBounced is a free log subscription operation binding the contract event 0x3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc4.
//
// Solidity: event ChequeBounced()
func (_Swap *SwapFilterer) WatchChequeBounced(opts *bind.WatchOpts, sink chan<- *SwapChequeBounced) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "ChequeBounced")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapChequeBounced)
				if err := _Swap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
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
func (_Swap *SwapFilterer) ParseChequeBounced(log types.Log) (*SwapChequeBounced, error) {
	event := new(SwapChequeBounced)
	if err := _Swap.contract.UnpackLog(event, "ChequeBounced", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapChequeCashedIterator is returned from FilterChequeCashed and is used to iterate over the raw logs and unpacked data for ChequeCashed events raised by the Swap contract.
type SwapChequeCashedIterator struct {
	Event *SwapChequeCashed // Event containing the contract specifics and raw log

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
func (it *SwapChequeCashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapChequeCashed)
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
		it.Event = new(SwapChequeCashed)
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
func (it *SwapChequeCashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapChequeCashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapChequeCashed represents a ChequeCashed event raised by the Swap contract.
type SwapChequeCashed struct {
	BeneficiaryPrincipal common.Address
	BeneficiaryAgent     common.Address
	Callee               common.Address
	Serial               *big.Int
	TotalPayout          *big.Int
	RequestPayout        *big.Int
	CalleePayout         *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterChequeCashed is a free log retrieval operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_Swap *SwapFilterer) FilterChequeCashed(opts *bind.FilterOpts, beneficiaryPrincipal []common.Address, beneficiaryAgent []common.Address, callee []common.Address) (*SwapChequeCashedIterator, error) {

	var beneficiaryPrincipalRule []interface{}
	for _, beneficiaryPrincipalItem := range beneficiaryPrincipal {
		beneficiaryPrincipalRule = append(beneficiaryPrincipalRule, beneficiaryPrincipalItem)
	}
	var beneficiaryAgentRule []interface{}
	for _, beneficiaryAgentItem := range beneficiaryAgent {
		beneficiaryAgentRule = append(beneficiaryAgentRule, beneficiaryAgentItem)
	}
	var calleeRule []interface{}
	for _, calleeItem := range callee {
		calleeRule = append(calleeRule, calleeItem)
	}

	logs, sub, err := _Swap.contract.FilterLogs(opts, "ChequeCashed", beneficiaryPrincipalRule, beneficiaryAgentRule, calleeRule)
	if err != nil {
		return nil, err
	}
	return &SwapChequeCashedIterator{contract: _Swap.contract, event: "ChequeCashed", logs: logs, sub: sub}, nil
}

// WatchChequeCashed is a free log subscription operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_Swap *SwapFilterer) WatchChequeCashed(opts *bind.WatchOpts, sink chan<- *SwapChequeCashed, beneficiaryPrincipal []common.Address, beneficiaryAgent []common.Address, callee []common.Address) (event.Subscription, error) {

	var beneficiaryPrincipalRule []interface{}
	for _, beneficiaryPrincipalItem := range beneficiaryPrincipal {
		beneficiaryPrincipalRule = append(beneficiaryPrincipalRule, beneficiaryPrincipalItem)
	}
	var beneficiaryAgentRule []interface{}
	for _, beneficiaryAgentItem := range beneficiaryAgent {
		beneficiaryAgentRule = append(beneficiaryAgentRule, beneficiaryAgentItem)
	}
	var calleeRule []interface{}
	for _, calleeItem := range callee {
		calleeRule = append(calleeRule, calleeItem)
	}

	logs, sub, err := _Swap.contract.WatchLogs(opts, "ChequeCashed", beneficiaryPrincipalRule, beneficiaryAgentRule, calleeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapChequeCashed)
				if err := _Swap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
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

// ParseChequeCashed is a log parse operation binding the contract event 0x5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed13.
//
// Solidity: event ChequeCashed(address indexed beneficiaryPrincipal, address indexed beneficiaryAgent, address indexed callee, uint256 serial, uint256 totalPayout, uint256 requestPayout, uint256 calleePayout)
func (_Swap *SwapFilterer) ParseChequeCashed(log types.Log) (*SwapChequeCashed, error) {
	event := new(SwapChequeCashed)
	if err := _Swap.contract.UnpackLog(event, "ChequeCashed", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapChequeSubmittedIterator is returned from FilterChequeSubmitted and is used to iterate over the raw logs and unpacked data for ChequeSubmitted events raised by the Swap contract.
type SwapChequeSubmittedIterator struct {
	Event *SwapChequeSubmitted // Event containing the contract specifics and raw log

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
func (it *SwapChequeSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapChequeSubmitted)
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
		it.Event = new(SwapChequeSubmitted)
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
func (it *SwapChequeSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapChequeSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapChequeSubmitted represents a ChequeSubmitted event raised by the Swap contract.
type SwapChequeSubmitted struct {
	Beneficiary common.Address
	Serial      *big.Int
	Amount      *big.Int
	CashTimeout *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChequeSubmitted is a free log retrieval operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
func (_Swap *SwapFilterer) FilterChequeSubmitted(opts *bind.FilterOpts, beneficiary []common.Address, serial []*big.Int) (*SwapChequeSubmittedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _Swap.contract.FilterLogs(opts, "ChequeSubmitted", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return &SwapChequeSubmittedIterator{contract: _Swap.contract, event: "ChequeSubmitted", logs: logs, sub: sub}, nil
}

// WatchChequeSubmitted is a free log subscription operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
func (_Swap *SwapFilterer) WatchChequeSubmitted(opts *bind.WatchOpts, sink chan<- *SwapChequeSubmitted, beneficiary []common.Address, serial []*big.Int) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}
	var serialRule []interface{}
	for _, serialItem := range serial {
		serialRule = append(serialRule, serialItem)
	}

	logs, sub, err := _Swap.contract.WatchLogs(opts, "ChequeSubmitted", beneficiaryRule, serialRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapChequeSubmitted)
				if err := _Swap.contract.UnpackLog(event, "ChequeSubmitted", log); err != nil {
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

// ParseChequeSubmitted is a log parse operation binding the contract event 0x543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c.
//
// Solidity: event ChequeSubmitted(address indexed beneficiary, uint256 indexed serial, uint256 amount, uint256 cashTimeout)
func (_Swap *SwapFilterer) ParseChequeSubmitted(log types.Log) (*SwapChequeSubmitted, error) {
	event := new(SwapChequeSubmitted)
	if err := _Swap.contract.UnpackLog(event, "ChequeSubmitted", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Swap contract.
type SwapDepositIterator struct {
	Event *SwapDeposit // Event containing the contract specifics and raw log

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
func (it *SwapDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapDeposit)
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
		it.Event = new(SwapDeposit)
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
func (it *SwapDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapDeposit represents a Deposit event raised by the Swap contract.
type SwapDeposit struct {
	Depositor common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address depositor, uint256 amount)
func (_Swap *SwapFilterer) FilterDeposit(opts *bind.FilterOpts) (*SwapDepositIterator, error) {

	logs, sub, err := _Swap.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &SwapDepositIterator{contract: _Swap.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address depositor, uint256 amount)
func (_Swap *SwapFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *SwapDeposit) (event.Subscription, error) {

	logs, sub, err := _Swap.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapDeposit)
				if err := _Swap.contract.UnpackLog(event, "Deposit", log); err != nil {
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
func (_Swap *SwapFilterer) ParseDeposit(log types.Log) (*SwapDeposit, error) {
	event := new(SwapDeposit)
	if err := _Swap.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapHardDepositAmountChangedIterator is returned from FilterHardDepositAmountChanged and is used to iterate over the raw logs and unpacked data for HardDepositAmountChanged events raised by the Swap contract.
type SwapHardDepositAmountChangedIterator struct {
	Event *SwapHardDepositAmountChanged // Event containing the contract specifics and raw log

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
func (it *SwapHardDepositAmountChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapHardDepositAmountChanged)
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
		it.Event = new(SwapHardDepositAmountChanged)
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
func (it *SwapHardDepositAmountChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapHardDepositAmountChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapHardDepositAmountChanged represents a HardDepositAmountChanged event raised by the Swap contract.
type SwapHardDepositAmountChanged struct {
	Beneficiary common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHardDepositAmountChanged is a free log retrieval operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_Swap *SwapFilterer) FilterHardDepositAmountChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*SwapHardDepositAmountChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.FilterLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SwapHardDepositAmountChangedIterator{contract: _Swap.contract, event: "HardDepositAmountChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositAmountChanged is a free log subscription operation binding the contract event 0x2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad.
//
// Solidity: event HardDepositAmountChanged(address indexed beneficiary, uint256 amount)
func (_Swap *SwapFilterer) WatchHardDepositAmountChanged(opts *bind.WatchOpts, sink chan<- *SwapHardDepositAmountChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.WatchLogs(opts, "HardDepositAmountChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapHardDepositAmountChanged)
				if err := _Swap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
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
func (_Swap *SwapFilterer) ParseHardDepositAmountChanged(log types.Log) (*SwapHardDepositAmountChanged, error) {
	event := new(SwapHardDepositAmountChanged)
	if err := _Swap.contract.UnpackLog(event, "HardDepositAmountChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapHardDepositDecreasePreparedIterator is returned from FilterHardDepositDecreasePrepared and is used to iterate over the raw logs and unpacked data for HardDepositDecreasePrepared events raised by the Swap contract.
type SwapHardDepositDecreasePreparedIterator struct {
	Event *SwapHardDepositDecreasePrepared // Event containing the contract specifics and raw log

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
func (it *SwapHardDepositDecreasePreparedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapHardDepositDecreasePrepared)
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
		it.Event = new(SwapHardDepositDecreasePrepared)
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
func (it *SwapHardDepositDecreasePreparedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapHardDepositDecreasePreparedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapHardDepositDecreasePrepared represents a HardDepositDecreasePrepared event raised by the Swap contract.
type SwapHardDepositDecreasePrepared struct {
	Beneficiary    common.Address
	DecreaseAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreasePrepared is a free log retrieval operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_Swap *SwapFilterer) FilterHardDepositDecreasePrepared(opts *bind.FilterOpts, beneficiary []common.Address) (*SwapHardDepositDecreasePreparedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.FilterLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SwapHardDepositDecreasePreparedIterator{contract: _Swap.contract, event: "HardDepositDecreasePrepared", logs: logs, sub: sub}, nil
}

// WatchHardDepositDecreasePrepared is a free log subscription operation binding the contract event 0xc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a.
//
// Solidity: event HardDepositDecreasePrepared(address indexed beneficiary, uint256 decreaseAmount)
func (_Swap *SwapFilterer) WatchHardDepositDecreasePrepared(opts *bind.WatchOpts, sink chan<- *SwapHardDepositDecreasePrepared, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.WatchLogs(opts, "HardDepositDecreasePrepared", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapHardDepositDecreasePrepared)
				if err := _Swap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
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
func (_Swap *SwapFilterer) ParseHardDepositDecreasePrepared(log types.Log) (*SwapHardDepositDecreasePrepared, error) {
	event := new(SwapHardDepositDecreasePrepared)
	if err := _Swap.contract.UnpackLog(event, "HardDepositDecreasePrepared", log); err != nil {
		return nil, err
	}
	return event, nil
}

// SwapHardDepositDecreaseTimeoutChangedIterator is returned from FilterHardDepositDecreaseTimeoutChanged and is used to iterate over the raw logs and unpacked data for HardDepositDecreaseTimeoutChanged events raised by the Swap contract.
type SwapHardDepositDecreaseTimeoutChangedIterator struct {
	Event *SwapHardDepositDecreaseTimeoutChanged // Event containing the contract specifics and raw log

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
func (it *SwapHardDepositDecreaseTimeoutChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SwapHardDepositDecreaseTimeoutChanged)
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
		it.Event = new(SwapHardDepositDecreaseTimeoutChanged)
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
func (it *SwapHardDepositDecreaseTimeoutChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SwapHardDepositDecreaseTimeoutChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SwapHardDepositDecreaseTimeoutChanged represents a HardDepositDecreaseTimeoutChanged event raised by the Swap contract.
type SwapHardDepositDecreaseTimeoutChanged struct {
	Beneficiary     common.Address
	DecreaseTimeout *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterHardDepositDecreaseTimeoutChanged is a free log retrieval operation binding the contract event 0x86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f.
//
// Solidity: event HardDepositDecreaseTimeoutChanged(address indexed beneficiary, uint256 decreaseTimeout)
func (_Swap *SwapFilterer) FilterHardDepositDecreaseTimeoutChanged(opts *bind.FilterOpts, beneficiary []common.Address) (*SwapHardDepositDecreaseTimeoutChangedIterator, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.FilterLogs(opts, "HardDepositDecreaseTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &SwapHardDepositDecreaseTimeoutChangedIterator{contract: _Swap.contract, event: "HardDepositDecreaseTimeoutChanged", logs: logs, sub: sub}, nil
}

// WatchHardDepositDecreaseTimeoutChanged is a free log subscription operation binding the contract event 0x86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f.
//
// Solidity: event HardDepositDecreaseTimeoutChanged(address indexed beneficiary, uint256 decreaseTimeout)
func (_Swap *SwapFilterer) WatchHardDepositDecreaseTimeoutChanged(opts *bind.WatchOpts, sink chan<- *SwapHardDepositDecreaseTimeoutChanged, beneficiary []common.Address) (event.Subscription, error) {

	var beneficiaryRule []interface{}
	for _, beneficiaryItem := range beneficiary {
		beneficiaryRule = append(beneficiaryRule, beneficiaryItem)
	}

	logs, sub, err := _Swap.contract.WatchLogs(opts, "HardDepositDecreaseTimeoutChanged", beneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SwapHardDepositDecreaseTimeoutChanged)
				if err := _Swap.contract.UnpackLog(event, "HardDepositDecreaseTimeoutChanged", log); err != nil {
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
func (_Swap *SwapFilterer) ParseHardDepositDecreaseTimeoutChanged(log types.Log) (*SwapHardDepositDecreaseTimeoutChanged, error) {
	event := new(SwapHardDepositDecreaseTimeoutChanged)
	if err := _Swap.contract.UnpackLog(event, "HardDepositDecreaseTimeoutChanged", log); err != nil {
		return nil, err
	}
	return event, nil
}
