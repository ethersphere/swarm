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
var ECDSABin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820fb8bf99a990c7187f714184f4e6aff124679083c701874e81a711141ca23c58f64736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e35363130643161620058"

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
var MathBin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a723158200fefab0c78d05cf589c417d5413989a81b64756aca4748adff95ece235dcefdd64736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e35363130643161620058"

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
var SafeMathBin = "0x607b6023600b82828239805160001a607314601657fe5b30600052607381538281f3fe73000000000000000000000000000000000000000030146080604052600080fdfea265627a7a72315820fd19c6c331a72705e77891c8709997654cc5762235c1d17b0593a9748716427764736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e35363130643161620058"

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
const SimpleSwapABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"swap\",\"type\":\"address\"},{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"chequeHash\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"issuer\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"}],\"name\":\"cashChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"increaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DEFAULT_HARDDEPPOSIT_DECREASE_TIMEOUT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"},{\"name\":\"beneficarySig\",\"type\":\"bytes\"}],\"name\":\"submitCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"submitChequeissuer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"cheques\",\"outputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"paidOut\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"decreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"hardDeposits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"canBeDecreasedAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"prepareDecreaseHardDeposit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"liquidBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"}],\"name\":\"liquidBalanceFor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiary\",\"type\":\"address\"},{\"name\":\"decreaseTimeout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"}],\"name\":\"setCustomHardDepositDecreaseTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalHardDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"beneficiarySig\",\"type\":\"bytes\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"cashCheque\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"_cashChequeInternal\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"serial\",\"type\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"cashTimeout\",\"type\":\"uint256\"},{\"name\":\"issuerSig\",\"type\":\"bytes\"}],\"name\":\"submitChequeBeneficiary\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_issuer\",\"type\":\"address\"},{\"name\":\"defaultHardDepositTimeoutDuration\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiaryPrincipal\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"beneficiaryAgent\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"callee\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"totalPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"requestPayout\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"calleePayout\",\"type\":\"uint256\"}],\"name\":\"ChequeCashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"serial\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"cashTimeout\",\"type\":\"uint256\"}],\"name\":\"ChequeSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"ChequeBounced\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"HardDepositAmountChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseAmount\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreasePrepared\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"beneficiary\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"decreaseTimeout\",\"type\":\"uint256\"}],\"name\":\"HardDepositDecreaseTimeoutChanged\",\"type\":\"event\"}]"

// SimpleSwapBin is the compiled bytecode used for deploying new contracts.
var SimpleSwapBin = "0x60806040526040516127633803806127638339818101604052604081101561002657600080fd5b8101908080519060200190929190805190602001909291905050508060008190555081600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060003411156100fe577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b50506126548061010f6000396000f3fe6080604052600436106101145760003560e01c8063946f46a2116100a0578063df32438011610064578063df32438014610874578063e0bcf13a14610966578063e3bb7aec14610991578063f3c08b1f14610ab7578063f890673b14610b3c57610114565b8063946f46a2146106be578063b6343b0d1461070f578063b777035014610789578063b7ec1a33146107e4578063c76a4d311461080f57610114565b8063338f3fed116100e7578063338f3fed1461031b57806339d9ec4c146103765780634f823a4c146103a157806354fe26141461053e5780636162913b1461064457610114565b8063030aca3e1461018b5780631d1438481461022e5780632329d2a8146102855780632e1a7d4d146102e0575b6000341115610189577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b005b34801561019757600080fd5b50610218600480360360a08110156101ae57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050610c22565b6040518082815260200191505060405180910390f35b34801561023a57600080fd5b50610243610cd4565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561029157600080fd5b506102de600480360360408110156102a857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610cfa565b005b3480156102ec57600080fd5b506103196004803603602081101561030357600080fd5b8101908080359060200190929190505050610d0b565b005b34801561032757600080fd5b506103746004803603604081101561033e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610e9a565b005b34801561038257600080fd5b5061038b6110bf565b6040518082815260200191505060405180910390f35b3480156103ad57600080fd5b5061053c600480360360c08110156103c457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190803590602001909291908035906020019064010000000081111561041f57600080fd5b82018360208201111561043157600080fd5b8035906020019184600183028401116401000000008311171561045357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001906401000000008111156104b657600080fd5b8201836020820111156104c857600080fd5b803590602001918460018302840111640100000000831117156104ea57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506110c5565b005b34801561054a57600080fd5b50610642600480360360a081101561056157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190803590602001906401000000008111156105bc57600080fd5b8201836020820111156105ce57600080fd5b803590602001918460018302840111640100000000831117156105f057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061124a565b005b34801561065057600080fd5b506106936004803603602081101561066757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506113b9565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b3480156106ca57600080fd5b5061070d600480360360208110156106e157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506113e9565b005b34801561071b57600080fd5b5061075e6004803603602081101561073257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061153c565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b34801561079557600080fd5b506107e2600480360360408110156107ac57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061156c565b005b3480156107f057600080fd5b506107f9611754565b6040518082815260200191505060405180910390f35b34801561081b57600080fd5b5061085e6004803603602081101561083257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611787565b6040518082815260200191505060405180910390f35b34801561088057600080fd5b506109646004803603606081101561089757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001906401000000008111156108de57600080fd5b8201836020820111156108f057600080fd5b8035906020019184600183028401116401000000008311171561091257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506117ec565b005b34801561097257600080fd5b5061097b611a5a565b6040518082815260200191505060405180910390f35b34801561099d57600080fd5b50610ab5600480360360c08110156109b457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190640100000000811115610a1b57600080fd5b820183602082011115610a2d57600080fd5b80359060200191846001830284011164010000000083111715610a4f57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192908035906020019092919080359060200190929190505050611a60565b005b348015610ac357600080fd5b50610b3a60048036036080811015610ada57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190505050611c7c565b005b348015610b4857600080fd5b50610c2060048036036080811015610b5f57600080fd5b8101908080359060200190929190803590602001909291908035906020019092919080359060200190640100000000811115610b9a57600080fd5b820183602082011115610bac57600080fd5b80359060200191846001830284011164010000000083111715610bce57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050612002565b005b60008585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018381526020018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b610d073383836000611c7c565b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610dce576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610dd6611754565b811115610e2e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260288152602001806125af6028913960400191505060405180910390fd5b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610e96573d6000803e3d6000fd5b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610f5d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b3073ffffffffffffffffffffffffffffffffffffffff1631610f8a826003546120ec90919063ffffffff16565b1115610fe1576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252603581526020018061257a6035913960400191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905061103b8282600001546120ec90919063ffffffff16565b8160000181905550611058826003546120ec90919063ffffffff16565b600381905550600081600201819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60005481565b6110db6110d53088888888610c22565b83612174565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461119d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b6111b36111ad3088888888610c22565b82612174565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff1614611236576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806124c66022913960400191505060405180910390fd5b61124286868686612190565b505050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461130d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b61132361131d3087878787610c22565b82612174565b73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff16146113a6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806124c66022913960400191505060405180910390fd5b6113b285858585612190565b5050505050565b60016020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015801561144557506000816003015414155b61149a576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260258152602001806124e86025913960400191505060405180910390fd5b6114b5816001015482600001546122c790919063ffffffff16565b8160000181905550600081600301819055506114e081600101546003546122c790919063ffffffff16565b6003819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60026020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461162f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600001548211156116cf576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260278152602001806125316027913960400191505060405180910390fd5b6000808260020154146116e65781600201546116ea565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b60006117826003543073ffffffffffffffffffffffffffffffffffffffff16316122c790919063ffffffff16565b905090565b60006117e5600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546117d7611754565b6120ec90919063ffffffff16565b9050919050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146118af576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b611947308484604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200182815260200193505050506040516020818303038152906040528051906020012082612174565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff161461197e57600080fd5b81600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f600260008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201546040518082815260200191505060405180910390a2505050565b60035481565b81421115611ab9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806125586022913960400191505060405180910390fd5b611b9e303386888686604051602001808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140183815260200182815260200196505050505050506040516020818303038152906040528051906020012084612174565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff1614611c21576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806124c66022913960400191505060405180910390fd5b611c2d86868684611c7c565b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611c73573d6000803e3d6000fd5b50505050505050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015611d1c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602481526020018061250d6024913960400191505060405180910390fd5b611d37816002015482600101546122c790919063ffffffff16565b831115611d8f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260238152602001806125d76023913960400191505060405180910390fd5b6000611ddd84600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154612350565b90506000611df48583611dee611754565b01612350565b905060008214611eb557611e5382600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546122c790919063ffffffff16565b600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000181905550611eae826003546122c790919063ffffffff16565b6003819055505b611ecc8184600201546120ec90919063ffffffff16565b83600201819055508573ffffffffffffffffffffffffffffffffffffffff166108fc611f0186846122c790919063ffffffff16565b9081150290604051600060405180830381858888f19350505050158015611f2c573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff167f5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed138660000154858a8a6040518085815260200184815260200183815260200182815260200194505050505060405180910390a4808514611ff9577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b50505050505050565b6120186120123033878787610c22565b82612174565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16146120da576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b6120e633858585612190565b50505050565b60008082840190508381101561216a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b600061218861218284612369565b836123c1565b905092915050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000154841161224c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f53696d706c65537761703a20696e76616c69642073657269616c00000000000081525060200191505060405180910390fd5b8381600001819055508281600101819055508142018160030181905550838573ffffffffffffffffffffffffffffffffffffffff167f543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c8585604051808381526020018281526020019250505060405180910390a35050505050565b60008282111561233f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525060200191505060405180910390fd5b600082840390508091505092915050565b600081831061235f5781612361565b825b905092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b600060418251146123d557600090506124bf565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c111561242957600093505050506124bf565b601b8160ff16141580156124415750601c8160ff1614155b1561245257600093505050506124bf565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa1580156124af573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fe53696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a20636865717565206e6f74207965742074696d6564206f757453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a2062656e6566696369617279536967206578706972656453696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e63652053696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e7453696d706c65537761703a206e6f7420656e6f7567682062616c616e6365206f776564a265627a7a7231582019f7cf09441cd293fcc062e67de8f16a471b67cb8832495b645e4d350d01296064736f6c637828302e352e31312d646576656c6f702e323031392e372e31302b636f6d6d69742e35363130643161620058"

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
