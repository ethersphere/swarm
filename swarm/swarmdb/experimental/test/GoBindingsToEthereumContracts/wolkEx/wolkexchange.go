// This file is an automatically generated Go binding. Do not modify as any
// change will likely be lost upon the next re-generation!

package main

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// WolkExchangeABI is the input ABI used to generate the binding from.
const WolkExchangeABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_wolkAmount\",\"type\":\"uint256\"}],\"name\":\"sellWolk\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"end_time\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newExchangeformula\",\"type\":\"address\"}],\"name\":\"setExchangeFormula\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"exchangeFormula\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_buyer\",\"type\":\"address\"}],\"name\":\"purchaseWolk\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isSellPossible\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"refund\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_kycRequirement\",\"type\":\"bool\"}],\"name\":\"updateRequireKYC\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenGenerationMin\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"balance\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isPurchasePossible\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"percentageETHReserve\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_startBlock\",\"type\":\"uint256\"},{\"name\":\"_endTime\",\"type\":\"uint256\"},{\"name\":\"_wolkinc\",\"type\":\"address\"}],\"name\":\"wolkGenesis\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"contributorTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_participant\",\"type\":\"address\"}],\"name\":\"participantStatus\",\"outputs\":[{\"name\":\"status\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reserveBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_participants\",\"type\":\"address[]\"}],\"name\":\"addParticipant\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_isRunning\",\"type\":\"bool\"}],\"name\":\"updateSellPossible\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tokenGenerationMax\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"start_block\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"newOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_tokenAddress\",\"type\":\"address\"},{\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"transferAnyERC20Token\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"allSaleCompleted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"wolkInc\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_isRunning\",\"type\":\"bool\"}],\"name\":\"updatePurchasePossible\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newReserveRatio\",\"type\":\"uint8\"}],\"name\":\"updateReserveRatio\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_participants\",\"type\":\"address[]\"}],\"name\":\"removeParticipant\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_participant\",\"type\":\"address\"}],\"name\":\"tokenGenerationEvent\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_tokenCreated\",\"type\":\"uint256\"}],\"name\":\"WolkCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_tokenDestroyed\",\"type\":\"uint256\"}],\"name\":\"WolkDestroyed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"LogRefund\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_prevOwner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_newOwner\",\"type\":\"address\"}],\"name\":\"OwnerUpdate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"_spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"}]"

// WolkExchangeBin is the compiled bytecode used for deploying new contracts.
const WolkExchangeBin = `606060405260006006556007805460a860020a60ff021960ff1991821660051716909155600c805460a860020a61ffff0219921660011791909116905560048054600160a060020a033316600160a060020a0319909116179055611c0f806100686000396000f300606060405236156101dc5763ffffffff60e060020a600035041662310e1681146102ed57806306fdde0314610315578063095ea7b31461039f57806316243356146103d557806318160ddd146103e857806323b872dd146103fb5780632659d8ef146104235780632f7a407b14610442578063313ce567146104715780633d8c9b8c14610484578063442d0927146104985780634bb278f3146104ab578063590e1ae3146104be57806362ac6115146104d15780636712e0be146104e957806370a08231146104fc57806379ba50971461051b5780637e1c0c091461052e578063835c638614610541578063847dc0a7146105545780638da5cb5b1461057d578063917d2be214610590578063936b603d146105b557806395d89b41146105c85780639c912a62146105db578063a10954fe146105fa578063a166b4b11461060d578063a9059cbb1461065c578063aad935af1461067e578063b57e6ea114610696578063b87fb3db146106a9578063d4ee1d90146106bc578063dc39d06d146106cf578063dd62ed3e146106f1578063de17910814610716578063e1d3097914610729578063e2542f031461073c578063e469185a14610754578063e814c9411461076d578063f2fde38b146107bc578063fa6b129d146107db575b600034116101e957600080fd5b60075460a860020a900460ff1615156102645730600160a060020a031663fa6b129d343360405160e060020a63ffffffff8516028152600160a060020a0390911660048201526024016000604051808303818588803b151561024a57600080fd5b6125ee5a03f1151561025b57600080fd5b505050506102eb565b600b5442106102e65730600160a060020a0316633d8c9b8c343360006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390911660048201526024016020604051808303818588803b15156102c457600080fd5b6125ee5a03f115156102d557600080fd5b5050505060405180519050506102eb565b600080fd5b005b34156102f857600080fd5b6103036004356107ef565b60405190815260200160405180910390f35b341561032057600080fd5b6103286109a3565b60405160208082528190810183818151815260200191508051906020019080838360005b8381101561036457808201518382015260200161034c565b50505050905090810190601f1680156103915780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34156103aa57600080fd5b6103c1600160a060020a03600435166024356109da565b604051901515815260200160405180910390f35b34156103e057600080fd5b610303610a80565b34156103f357600080fd5b610303610a86565b341561040657600080fd5b6103c1600160a060020a0360043581169060243516604435610a8c565b341561042e57600080fd5b6103c1600160a060020a0360043516610be3565b341561044d57600080fd5b610455610c73565b604051600160a060020a03909116815260200160405180910390f35b341561047c57600080fd5b610303610c87565b610303600160a060020a0360043516610c8c565b34156104a357600080fd5b6103c1610dd9565b34156104b657600080fd5b6102eb610dfc565b34156104c957600080fd5b6102eb610f4f565b34156104dc57600080fd5b6103c160043515156110a0565b34156104f457600080fd5b6103036110d2565b341561050757600080fd5b610303600160a060020a03600435166110e0565b341561052657600080fd5b6102eb6110fb565b341561053957600080fd5b6103036111a2565b341561054c57600080fd5b6103c16111a8565b341561055f57600080fd5b6105676111b8565b60405160ff909116815260200160405180910390f35b341561058857600080fd5b6104556111c1565b341561059b57600080fd5b6103c1600435602435600160a060020a03604435166111d0565b34156105c057600080fd5b61030361124b565b34156105d357600080fd5b610328611251565b34156105e657600080fd5b6103c1600160a060020a0360043516611288565b341561060557600080fd5b6103036112a6565b341561061857600080fd5b6103c160046024813581810190830135806020818102016040519081016040528093929190818152602001838360200280828437509496506112ac95505050505050565b341561066757600080fd5b6103c1600160a060020a036004351660243561132b565b341561068957600080fd5b6103c16004351515611418565b34156106a157600080fd5b6103036114e2565b34156106b457600080fd5b6103036114f1565b34156106c757600080fd5b6104556114f7565b34156106da57600080fd5b6103c1600160a060020a0360043516602435611506565b34156106fc57600080fd5b610303600160a060020a03600435811690602435166115a5565b341561072157600080fd5b6103c16115d0565b341561073457600080fd5b6104556115e0565b341561074757600080fd5b6103c160043515156115f4565b341561075f57600080fd5b6103c160ff600435166116aa565b341561077857600080fd5b6103c1600460248135818101908301358060208181020160405190810160405280939291908181526020018383602002808284375094965061171495505050505050565b34156107c757600080fd5b6102eb600160a060020a036004351661178a565b6102eb600160a060020a03600435166117ec565b600c546000908190760100000000000000000000000000000000000000000000900460ff168015610829575060075460a860020a900460ff165b151561083457600080fd5b600160a060020a0333166000908152602081905260409020548390101561085a57600080fd5b600c546108769084906101009004600160a060020a0316611a26565b9050600160a060020a0330163181901161088f57600080fd5b600160a060020a0333166000908152602081905260409020546108b29084611ac0565b600160a060020a0333166000908152602081905260409020556003546108d89084611ac0565b6003556002546108e89084611ac0565b6002556108ff600160a060020a0330163182611ac0565b600655600160a060020a0333167ff505eb6e610340eed3eea0048f8ec258cda0927f73be2d293288fde9a546f1ab8460405190815260200160405180910390a2600033600160a060020a0316600080516020611bc48339815191528560405190815260200160405180910390a3600160a060020a03331681156108fc0282604051600060405180830381858888f19350505050151561099d57600080fd5b92915050565b60408051908101604052600a81527f574f4c4b20544f4b454e00000000000000000000000000000000000000000000602082015281565b6000811580610a0c5750600160a060020a03338116600090815260016020908152604080832093871683529290522054155b1515610a1757600080fd5b600160a060020a03338116600081815260016020908152604080832094881680845294909152908190208590557f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259085905190815260200160405180910390a350600192915050565b600b5481565b60025490565b600160a060020a0380841660008181526001602090815260408083203390951683529381528382205492825281905291822054839010801590610af65750600160a060020a0380861660009081526001602090815260408083203390941683529290522054839010155b8015610b025750600083115b15610bd657600160a060020a038416600090815260208190526040902054610b2a9084611ad2565b600160a060020a038086166000908152602081905260408082209390935590871681522054610b599084611ac0565b600160a060020a038616600090815260208190526040902055610b7c8184611ac0565b600160a060020a0380871660008181526001602090815260408083203386168452909152908190209390935590861691600080516020611bc48339815191529086905190815260200160405180910390a360019150610bdb565b600091505b509392505050565b60045460009033600160a060020a03908116911614610bfe57fe5b6000610c12670de0b6b3a764000084611a26565b11610c1c57600080fd5b6000610c30670de0b6b3a764000084611af6565b11610c3a57600080fd5b50600c8054600160a060020a0383166101000276ffffffffffffffffffffffffffffffffffffffffffff00199091161790556001919050565b600c546101009004600160a060020a031681565b601281565b600c54600090819060a860020a900460ff168015610cb3575060075460a860020a900460ff165b1515610cbe57600080fd5b60003411610ccb57600080fd5b600c54610ce79034906101009004600160a060020a0316611af6565b905060008111610cf657600080fd5b610d0260035482611ad2565b600355600254610d129082611ad2565b600255600160a060020a038316600090815260208190526040902054610d389082611ad2565b600160a060020a038416600090815260208190526040902055600654610d5e9034611ad2565b600655600160a060020a0383167f7ff6ea1c893a974b2f363e8f8e474a1b52958080d1fffe0d085c286de30035d28260405190815260200160405180910390a282600160a060020a031630600160a060020a0316600080516020611bc48339815191528360405190815260200160405180910390a392915050565b600c54760100000000000000000000000000000000000000000000900460ff1681565b600454600090819033600160a060020a03908116911614610e1957fe5b60075460a860020a900460ff1615610e3057600080fd5b42600b556007546101009004600160a060020a031660009081526020819052604090206a295be96e6406697200000090819055600254909250610e739083611ad2565b6002556007546101009004600160a060020a03167f7ff6ea1c893a974b2f363e8f8e474a1b52958080d1fffe0d085c286de30035d28360405190815260200160405180910390a26007805475ff000000000000000000000000000000000000000000191660a860020a1790819055600354610efe91610ef5919060ff16611b6e565b620186a0611b8f565b6006819055610f1890600160a060020a0330163190611ac0565b9050600160a060020a03331681156108fc0282604051600060405180830381858888f193505050501515610f4b57600080fd5b5050565b600160a060020a03331660009081526008602052604081205481908190118015610f83575060075460a860020a900460ff16155b8015610f905750600b5442115b8015610fa8575060025469021e19e0c9bab240000090105b1515610fb357600080fd5b505033600160a060020a03166000908152602081815260408083208054600890935290832080549184905592909255600254909190610ff29083611ac0565b600255600160a060020a0333167ff505eb6e610340eed3eea0048f8ec258cda0927f73be2d293288fde9a546f1ab8360405190815260200160405180910390a233600160a060020a03167fb6c0eca8138e097d71e2dd31e19a1266487f0553f170b7260ffe68bcbe9ff8a78260405190815260200160405180910390a2600160a060020a03331681156108fc0282604051600060405180830381858888f193505050501515610f4b57600080fd5b60045460009033600160a060020a039081169116146110bb57fe5b50600c805460ff1916911515919091179055600190565b69021e19e0c9bab240000081565b600160a060020a031660009081526020819052604090205490565b60055433600160a060020a0390811691161461111657600080fd5b6004546005547f343765429aea5a34b3ff6a3785a98a5abb2597aca87bfbb58632c173d585373a91600160a060020a039081169116604051600160a060020a039283168152911660208201526040908101905180910390a1600580546004805473ffffffffffffffffffffffffffffffffffffffff19908116600160a060020a03841617909155169055565b60025481565b600c5460a860020a900460ff1681565b60075460ff1681565b600454600160a060020a031681565b60045460009033600160a060020a039081169116146111eb57fe5b60016002541080156111fd5750834311155b151561120857600080fd5b50600a839055600b82905560078054600160a060020a0383166101000274ffffffffffffffffffffffffffffffffffffffff001990911617905560019392505050565b60035481565b60408051908101604052600381527f574c4b0000000000000000000000000000000000000000000000000000000000602082015281565b600160a060020a031660009081526009602052604090205460ff1690565b60065481565b600454600090819033600160a060020a039081169116146112c957fe5b5060005b8251811015611322576001600960008584815181106112e857fe5b90602001906020020151600160a060020a031681526020810191909152604001600020805460ff19169115159190911790556001016112cd565b50600192915050565b600160a060020a0333166000908152602081905260408120548290108015906113545750600082115b1561141057600160a060020a03331660009081526020819052604090205461137c9083611ac0565b600160a060020a0333811660009081526020819052604080822093909355908516815220546113ab9083611ad2565b60008085600160a060020a0316600160a060020a031681526020019081526020016000208190555082600160a060020a031633600160a060020a0316600080516020611bc48339815191528460405190815260200160405180910390a350600161099d565b50600061099d565b60045460009033600160a060020a0390811691161461143357fe5b811561149b57600c5460009061146090670de0b6b3a7640000906101009004600160a060020a0316611a26565b1161146a57600080fd5b600c5460009061149190670de0b6b3a7640000906101009004600160a060020a0316611af6565b1161149b57600080fd5b50600c80548215157601000000000000000000000000000000000000000000000276ff00000000000000000000000000000000000000000000199091161790556001919050565b6a0e79c4e6a3023e8180000081565b600a5481565b600554600160a060020a031681565b60045460009033600160a060020a0390811691161461152157fe5b600454600160a060020a038085169163a9059cbb91168460006040516020015260405160e060020a63ffffffff8516028152600160a060020a0390921660048301526024820152604401602060405180830381600087803b151561158457600080fd5b6102c65a03f1151561159557600080fd5b5050506040518051949350505050565b600160a060020a03918216600090815260016020908152604080832093909416825291909152205490565b60075460a860020a900460ff1681565b6007546101009004600160a060020a031681565b60045460009033600160a060020a0390811691161461160f57fe5b811561167757600c5460009061163c90670de0b6b3a7640000906101009004600160a060020a0316611a26565b1161164657600080fd5b600c5460009061166d90670de0b6b3a7640000906101009004600160a060020a0316611af6565b1161167757600080fd5b50600c805482151560a860020a0275ff000000000000000000000000000000000000000000199091161790556001919050565b60045460009033600160a060020a039081169116146116c557fe5b60075460a860020a900460ff1680156116e1575060018260ff16115b80156116f0575060148260ff16105b15156116fb57600080fd5b506007805460ff831660ff199091161790556001919050565b600454600090819033600160a060020a0390811691161461173157fe5b5060005b82518110156113225760006009600085848151811061175057fe5b90602001906020020151600160a060020a031681526020810191909152604001600020805460ff1916911515919091179055600101611735565b60045433600160a060020a039081169116146117a257fe5b600454600160a060020a03828116911614156117bd57600080fd5b6005805473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a0392909216919091179055565b600160a060020a0381166000908152600960205260408120548190819060ff168061182f5750600160a060020a03331660009081526009602052604090205460ff165b806118505750600160a060020a038416600090815260208190526040812054115b8061185d5750600c5460ff165b8015611873575060075460a860020a900460ff16155b80156118815750600b544211155b801561188d5750600034115b151561189857600080fd5b6002546103e893506118d7906a0e79c4e6a3023e81800000906118d2906901da56a4b0835bf80000906118cd906107d0611b8f565b611ad2565b611b8f565b92506107d08311156118e9576107d092505b6101f48310156118f9576101f492505b600a5443101561190857600080fd5b6119123484611b6e565b915061192060025483611ad2565b90506a0e79c4e6a3023e8180000081111561193a57600080fd5b600281905560035461194c9083611ad2565b600355600160a060020a03808516903016600080516020611bc48339815191528460405190815260200160405180910390a3600160a060020a0384166000908152602081905260409020546119a19083611ad2565b600160a060020a038516600090815260208181526040808320939093556008905220546119ce9034611ad2565b600160a060020a0385166000818152600860205260409081902092909255907f7ff6ea1c893a974b2f363e8f8e474a1b52958080d1fffe0d085c286de30035d29084905190815260200160405180910390a250505050565b6003546006546007546000928392600160a060020a0386169263f7a4c45c92919060ff1688866040516020015260405160e060020a63ffffffff87160281526004810194909452602484019290925260ff1660448301526064820152608401602060405180830381600087803b1515611a9e57600080fd5b6102c65a03f11515611aaf57600080fd5b505050604051805195945050505050565b600082821115611acc57fe5b50900390565b6000828201838110801590611ae75750828110155b1515611aef57fe5b9392505050565b6003546006546007546000928392600160a060020a0386169263949dfa6392919060ff1688866040516020015260405160e060020a63ffffffff87160281526004810194909452602484019290925260ff1660448301526064820152608401602060405180830381600087803b1515611a9e57600080fd5b6000828202831580611ae75750828482811515611b8757fe5b0414611aef57fe5b600080808311611b9b57fe5b8284811515611ba657fe5b0490508284811515611bb457fe5b068184020184141515611aef57fe00ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3efa165627a7a723058209aaada50c1671511e7d4a6f65fcfba9395b602433a033c0d707bfb0dbd61a6500029`

// DeployWolkExchange deploys a new Ethereum contract, binding an instance of WolkExchange to it.
func DeployWolkExchange(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *WolkExchange, error) {
	parsed, err := abi.JSON(strings.NewReader(WolkExchangeABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WolkExchangeBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &WolkExchange{WolkExchangeCaller: WolkExchangeCaller{contract: contract}, WolkExchangeTransactor: WolkExchangeTransactor{contract: contract}}, nil
}

// WolkExchange is an auto generated Go binding around an Ethereum contract.
type WolkExchange struct {
	WolkExchangeCaller     // Read-only binding to the contract
	WolkExchangeTransactor // Write-only binding to the contract
}

// WolkExchangeCaller is an auto generated read-only Go binding around an Ethereum contract.
type WolkExchangeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WolkExchangeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WolkExchangeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WolkExchangeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WolkExchangeSession struct {
	Contract     *WolkExchange     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WolkExchangeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WolkExchangeCallerSession struct {
	Contract *WolkExchangeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// WolkExchangeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WolkExchangeTransactorSession struct {
	Contract     *WolkExchangeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// WolkExchangeRaw is an auto generated low-level Go binding around an Ethereum contract.
type WolkExchangeRaw struct {
	Contract *WolkExchange // Generic contract binding to access the raw methods on
}

// WolkExchangeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WolkExchangeCallerRaw struct {
	Contract *WolkExchangeCaller // Generic read-only contract binding to access the raw methods on
}

// WolkExchangeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WolkExchangeTransactorRaw struct {
	Contract *WolkExchangeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWolkExchange creates a new instance of WolkExchange, bound to a specific deployed contract.
func NewWolkExchange(address common.Address, backend bind.ContractBackend) (*WolkExchange, error) {
	contract, err := bindWolkExchange(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WolkExchange{WolkExchangeCaller: WolkExchangeCaller{contract: contract}, WolkExchangeTransactor: WolkExchangeTransactor{contract: contract}}, nil
}

// NewWolkExchangeCaller creates a new read-only instance of WolkExchange, bound to a specific deployed contract.
func NewWolkExchangeCaller(address common.Address, caller bind.ContractCaller) (*WolkExchangeCaller, error) {
	contract, err := bindWolkExchange(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &WolkExchangeCaller{contract: contract}, nil
}

// NewWolkExchangeTransactor creates a new write-only instance of WolkExchange, bound to a specific deployed contract.
func NewWolkExchangeTransactor(address common.Address, transactor bind.ContractTransactor) (*WolkExchangeTransactor, error) {
	contract, err := bindWolkExchange(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &WolkExchangeTransactor{contract: contract}, nil
}

// bindWolkExchange binds a generic wrapper to an already deployed contract.
func bindWolkExchange(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WolkExchangeABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WolkExchange *WolkExchangeRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _WolkExchange.Contract.WolkExchangeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WolkExchange *WolkExchangeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WolkExchange.Contract.WolkExchangeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WolkExchange *WolkExchangeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WolkExchange.Contract.WolkExchangeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WolkExchange *WolkExchangeCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _WolkExchange.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WolkExchange *WolkExchangeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WolkExchange.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WolkExchange *WolkExchangeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WolkExchange.Contract.contract.Transact(opts, method, params...)
}

// AllSaleCompleted is a free data retrieval call binding the contract method 0xde179108.
//
// Solidity: function allSaleCompleted() constant returns(bool)
func (_WolkExchange *WolkExchangeCaller) AllSaleCompleted(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "allSaleCompleted")
	return *ret0, err
}

// AllSaleCompleted is a free data retrieval call binding the contract method 0xde179108.
//
// Solidity: function allSaleCompleted() constant returns(bool)
func (_WolkExchange *WolkExchangeSession) AllSaleCompleted() (bool, error) {
	return _WolkExchange.Contract.AllSaleCompleted(&_WolkExchange.CallOpts)
}

// AllSaleCompleted is a free data retrieval call binding the contract method 0xde179108.
//
// Solidity: function allSaleCompleted() constant returns(bool)
func (_WolkExchange *WolkExchangeCallerSession) AllSaleCompleted() (bool, error) {
	return _WolkExchange.Contract.AllSaleCompleted(&_WolkExchange.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_WolkExchange *WolkExchangeCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "allowance", _owner, _spender)
	return *ret0, err
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_WolkExchange *WolkExchangeSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _WolkExchange.Contract.Allowance(&_WolkExchange.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(_owner address, _spender address) constant returns(remaining uint256)
func (_WolkExchange *WolkExchangeCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _WolkExchange.Contract.Allowance(&_WolkExchange.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_WolkExchange *WolkExchangeCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "balanceOf", _owner)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_WolkExchange *WolkExchangeSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _WolkExchange.Contract.BalanceOf(&_WolkExchange.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_owner address) constant returns(balance uint256)
func (_WolkExchange *WolkExchangeCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _WolkExchange.Contract.BalanceOf(&_WolkExchange.CallOpts, _owner)
}

// ContributorTokens is a free data retrieval call binding the contract method 0x936b603d.
//
// Solidity: function contributorTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) ContributorTokens(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "contributorTokens")
	return *ret0, err
}

// ContributorTokens is a free data retrieval call binding the contract method 0x936b603d.
//
// Solidity: function contributorTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) ContributorTokens() (*big.Int, error) {
	return _WolkExchange.Contract.ContributorTokens(&_WolkExchange.CallOpts)
}

// ContributorTokens is a free data retrieval call binding the contract method 0x936b603d.
//
// Solidity: function contributorTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) ContributorTokens() (*big.Int, error) {
	return _WolkExchange.Contract.ContributorTokens(&_WolkExchange.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) Decimals() (*big.Int, error) {
	return _WolkExchange.Contract.Decimals(&_WolkExchange.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) Decimals() (*big.Int, error) {
	return _WolkExchange.Contract.Decimals(&_WolkExchange.CallOpts)
}

// End_time is a free data retrieval call binding the contract method 0x16243356.
//
// Solidity: function end_time() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) End_time(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "end_time")
	return *ret0, err
}

// End_time is a free data retrieval call binding the contract method 0x16243356.
//
// Solidity: function end_time() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) End_time() (*big.Int, error) {
	return _WolkExchange.Contract.End_time(&_WolkExchange.CallOpts)
}

// End_time is a free data retrieval call binding the contract method 0x16243356.
//
// Solidity: function end_time() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) End_time() (*big.Int, error) {
	return _WolkExchange.Contract.End_time(&_WolkExchange.CallOpts)
}

// ExchangeFormula is a free data retrieval call binding the contract method 0x2f7a407b.
//
// Solidity: function exchangeFormula() constant returns(address)
func (_WolkExchange *WolkExchangeCaller) ExchangeFormula(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "exchangeFormula")
	return *ret0, err
}

// ExchangeFormula is a free data retrieval call binding the contract method 0x2f7a407b.
//
// Solidity: function exchangeFormula() constant returns(address)
func (_WolkExchange *WolkExchangeSession) ExchangeFormula() (common.Address, error) {
	return _WolkExchange.Contract.ExchangeFormula(&_WolkExchange.CallOpts)
}

// ExchangeFormula is a free data retrieval call binding the contract method 0x2f7a407b.
//
// Solidity: function exchangeFormula() constant returns(address)
func (_WolkExchange *WolkExchangeCallerSession) ExchangeFormula() (common.Address, error) {
	return _WolkExchange.Contract.ExchangeFormula(&_WolkExchange.CallOpts)
}

// IsPurchasePossible is a free data retrieval call binding the contract method 0x835c6386.
//
// Solidity: function isPurchasePossible() constant returns(bool)
func (_WolkExchange *WolkExchangeCaller) IsPurchasePossible(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "isPurchasePossible")
	return *ret0, err
}

// IsPurchasePossible is a free data retrieval call binding the contract method 0x835c6386.
//
// Solidity: function isPurchasePossible() constant returns(bool)
func (_WolkExchange *WolkExchangeSession) IsPurchasePossible() (bool, error) {
	return _WolkExchange.Contract.IsPurchasePossible(&_WolkExchange.CallOpts)
}

// IsPurchasePossible is a free data retrieval call binding the contract method 0x835c6386.
//
// Solidity: function isPurchasePossible() constant returns(bool)
func (_WolkExchange *WolkExchangeCallerSession) IsPurchasePossible() (bool, error) {
	return _WolkExchange.Contract.IsPurchasePossible(&_WolkExchange.CallOpts)
}

// IsSellPossible is a free data retrieval call binding the contract method 0x442d0927.
//
// Solidity: function isSellPossible() constant returns(bool)
func (_WolkExchange *WolkExchangeCaller) IsSellPossible(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "isSellPossible")
	return *ret0, err
}

// IsSellPossible is a free data retrieval call binding the contract method 0x442d0927.
//
// Solidity: function isSellPossible() constant returns(bool)
func (_WolkExchange *WolkExchangeSession) IsSellPossible() (bool, error) {
	return _WolkExchange.Contract.IsSellPossible(&_WolkExchange.CallOpts)
}

// IsSellPossible is a free data retrieval call binding the contract method 0x442d0927.
//
// Solidity: function isSellPossible() constant returns(bool)
func (_WolkExchange *WolkExchangeCallerSession) IsSellPossible() (bool, error) {
	return _WolkExchange.Contract.IsSellPossible(&_WolkExchange.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_WolkExchange *WolkExchangeCaller) Name(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "name")
	return *ret0, err
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_WolkExchange *WolkExchangeSession) Name() (string, error) {
	return _WolkExchange.Contract.Name(&_WolkExchange.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() constant returns(string)
func (_WolkExchange *WolkExchangeCallerSession) Name() (string, error) {
	return _WolkExchange.Contract.Name(&_WolkExchange.CallOpts)
}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() constant returns(address)
func (_WolkExchange *WolkExchangeCaller) NewOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "newOwner")
	return *ret0, err
}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() constant returns(address)
func (_WolkExchange *WolkExchangeSession) NewOwner() (common.Address, error) {
	return _WolkExchange.Contract.NewOwner(&_WolkExchange.CallOpts)
}

// NewOwner is a free data retrieval call binding the contract method 0xd4ee1d90.
//
// Solidity: function newOwner() constant returns(address)
func (_WolkExchange *WolkExchangeCallerSession) NewOwner() (common.Address, error) {
	return _WolkExchange.Contract.NewOwner(&_WolkExchange.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_WolkExchange *WolkExchangeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_WolkExchange *WolkExchangeSession) Owner() (common.Address, error) {
	return _WolkExchange.Contract.Owner(&_WolkExchange.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_WolkExchange *WolkExchangeCallerSession) Owner() (common.Address, error) {
	return _WolkExchange.Contract.Owner(&_WolkExchange.CallOpts)
}

// ParticipantStatus is a free data retrieval call binding the contract method 0x9c912a62.
//
// Solidity: function participantStatus(_participant address) constant returns(status bool)
func (_WolkExchange *WolkExchangeCaller) ParticipantStatus(opts *bind.CallOpts, _participant common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "participantStatus", _participant)
	return *ret0, err
}

// ParticipantStatus is a free data retrieval call binding the contract method 0x9c912a62.
//
// Solidity: function participantStatus(_participant address) constant returns(status bool)
func (_WolkExchange *WolkExchangeSession) ParticipantStatus(_participant common.Address) (bool, error) {
	return _WolkExchange.Contract.ParticipantStatus(&_WolkExchange.CallOpts, _participant)
}

// ParticipantStatus is a free data retrieval call binding the contract method 0x9c912a62.
//
// Solidity: function participantStatus(_participant address) constant returns(status bool)
func (_WolkExchange *WolkExchangeCallerSession) ParticipantStatus(_participant common.Address) (bool, error) {
	return _WolkExchange.Contract.ParticipantStatus(&_WolkExchange.CallOpts, _participant)
}

// PercentageETHReserve is a free data retrieval call binding the contract method 0x847dc0a7.
//
// Solidity: function percentageETHReserve() constant returns(uint8)
func (_WolkExchange *WolkExchangeCaller) PercentageETHReserve(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "percentageETHReserve")
	return *ret0, err
}

// PercentageETHReserve is a free data retrieval call binding the contract method 0x847dc0a7.
//
// Solidity: function percentageETHReserve() constant returns(uint8)
func (_WolkExchange *WolkExchangeSession) PercentageETHReserve() (uint8, error) {
	return _WolkExchange.Contract.PercentageETHReserve(&_WolkExchange.CallOpts)
}

// PercentageETHReserve is a free data retrieval call binding the contract method 0x847dc0a7.
//
// Solidity: function percentageETHReserve() constant returns(uint8)
func (_WolkExchange *WolkExchangeCallerSession) PercentageETHReserve() (uint8, error) {
	return _WolkExchange.Contract.PercentageETHReserve(&_WolkExchange.CallOpts)
}

// ReserveBalance is a free data retrieval call binding the contract method 0xa10954fe.
//
// Solidity: function reserveBalance() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) ReserveBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "reserveBalance")
	return *ret0, err
}

// ReserveBalance is a free data retrieval call binding the contract method 0xa10954fe.
//
// Solidity: function reserveBalance() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) ReserveBalance() (*big.Int, error) {
	return _WolkExchange.Contract.ReserveBalance(&_WolkExchange.CallOpts)
}

// ReserveBalance is a free data retrieval call binding the contract method 0xa10954fe.
//
// Solidity: function reserveBalance() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) ReserveBalance() (*big.Int, error) {
	return _WolkExchange.Contract.ReserveBalance(&_WolkExchange.CallOpts)
}

// Start_block is a free data retrieval call binding the contract method 0xb87fb3db.
//
// Solidity: function start_block() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) Start_block(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "start_block")
	return *ret0, err
}

// Start_block is a free data retrieval call binding the contract method 0xb87fb3db.
//
// Solidity: function start_block() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) Start_block() (*big.Int, error) {
	return _WolkExchange.Contract.Start_block(&_WolkExchange.CallOpts)
}

// Start_block is a free data retrieval call binding the contract method 0xb87fb3db.
//
// Solidity: function start_block() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) Start_block() (*big.Int, error) {
	return _WolkExchange.Contract.Start_block(&_WolkExchange.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_WolkExchange *WolkExchangeCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "symbol")
	return *ret0, err
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_WolkExchange *WolkExchangeSession) Symbol() (string, error) {
	return _WolkExchange.Contract.Symbol(&_WolkExchange.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() constant returns(string)
func (_WolkExchange *WolkExchangeCallerSession) Symbol() (string, error) {
	return _WolkExchange.Contract.Symbol(&_WolkExchange.CallOpts)
}

// TokenGenerationMax is a free data retrieval call binding the contract method 0xb57e6ea1.
//
// Solidity: function tokenGenerationMax() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) TokenGenerationMax(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "tokenGenerationMax")
	return *ret0, err
}

// TokenGenerationMax is a free data retrieval call binding the contract method 0xb57e6ea1.
//
// Solidity: function tokenGenerationMax() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) TokenGenerationMax() (*big.Int, error) {
	return _WolkExchange.Contract.TokenGenerationMax(&_WolkExchange.CallOpts)
}

// TokenGenerationMax is a free data retrieval call binding the contract method 0xb57e6ea1.
//
// Solidity: function tokenGenerationMax() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) TokenGenerationMax() (*big.Int, error) {
	return _WolkExchange.Contract.TokenGenerationMax(&_WolkExchange.CallOpts)
}

// TokenGenerationMin is a free data retrieval call binding the contract method 0x6712e0be.
//
// Solidity: function tokenGenerationMin() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) TokenGenerationMin(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "tokenGenerationMin")
	return *ret0, err
}

// TokenGenerationMin is a free data retrieval call binding the contract method 0x6712e0be.
//
// Solidity: function tokenGenerationMin() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) TokenGenerationMin() (*big.Int, error) {
	return _WolkExchange.Contract.TokenGenerationMin(&_WolkExchange.CallOpts)
}

// TokenGenerationMin is a free data retrieval call binding the contract method 0x6712e0be.
//
// Solidity: function tokenGenerationMin() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) TokenGenerationMin() (*big.Int, error) {
	return _WolkExchange.Contract.TokenGenerationMin(&_WolkExchange.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) TotalSupply() (*big.Int, error) {
	return _WolkExchange.Contract.TotalSupply(&_WolkExchange.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) TotalSupply() (*big.Int, error) {
	return _WolkExchange.Contract.TotalSupply(&_WolkExchange.CallOpts)
}

// TotalTokens is a free data retrieval call binding the contract method 0x7e1c0c09.
//
// Solidity: function totalTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeCaller) TotalTokens(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "totalTokens")
	return *ret0, err
}

// TotalTokens is a free data retrieval call binding the contract method 0x7e1c0c09.
//
// Solidity: function totalTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeSession) TotalTokens() (*big.Int, error) {
	return _WolkExchange.Contract.TotalTokens(&_WolkExchange.CallOpts)
}

// TotalTokens is a free data retrieval call binding the contract method 0x7e1c0c09.
//
// Solidity: function totalTokens() constant returns(uint256)
func (_WolkExchange *WolkExchangeCallerSession) TotalTokens() (*big.Int, error) {
	return _WolkExchange.Contract.TotalTokens(&_WolkExchange.CallOpts)
}

// WolkInc is a free data retrieval call binding the contract method 0xe1d30979.
//
// Solidity: function wolkInc() constant returns(address)
func (_WolkExchange *WolkExchangeCaller) WolkInc(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WolkExchange.contract.Call(opts, out, "wolkInc")
	return *ret0, err
}

// WolkInc is a free data retrieval call binding the contract method 0xe1d30979.
//
// Solidity: function wolkInc() constant returns(address)
func (_WolkExchange *WolkExchangeSession) WolkInc() (common.Address, error) {
	return _WolkExchange.Contract.WolkInc(&_WolkExchange.CallOpts)
}

// WolkInc is a free data retrieval call binding the contract method 0xe1d30979.
//
// Solidity: function wolkInc() constant returns(address)
func (_WolkExchange *WolkExchangeCallerSession) WolkInc() (common.Address, error) {
	return _WolkExchange.Contract.WolkInc(&_WolkExchange.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_WolkExchange *WolkExchangeTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_WolkExchange *WolkExchangeSession) AcceptOwnership() (*types.Transaction, error) {
	return _WolkExchange.Contract.AcceptOwnership(&_WolkExchange.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_WolkExchange *WolkExchangeTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _WolkExchange.Contract.AcceptOwnership(&_WolkExchange.TransactOpts)
}

// AddParticipant is a paid mutator transaction binding the contract method 0xa166b4b1.
//
// Solidity: function addParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) AddParticipant(opts *bind.TransactOpts, _participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "addParticipant", _participants)
}

// AddParticipant is a paid mutator transaction binding the contract method 0xa166b4b1.
//
// Solidity: function addParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeSession) AddParticipant(_participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.AddParticipant(&_WolkExchange.TransactOpts, _participants)
}

// AddParticipant is a paid mutator transaction binding the contract method 0xa166b4b1.
//
// Solidity: function addParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) AddParticipant(_participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.AddParticipant(&_WolkExchange.TransactOpts, _participants)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.Approve(&_WolkExchange.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(_spender address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.Approve(&_WolkExchange.TransactOpts, _spender, _value)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_WolkExchange *WolkExchangeTransactor) Finalize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "finalize")
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_WolkExchange *WolkExchangeSession) Finalize() (*types.Transaction, error) {
	return _WolkExchange.Contract.Finalize(&_WolkExchange.TransactOpts)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_WolkExchange *WolkExchangeTransactorSession) Finalize() (*types.Transaction, error) {
	return _WolkExchange.Contract.Finalize(&_WolkExchange.TransactOpts)
}

// PurchaseWolk is a paid mutator transaction binding the contract method 0x3d8c9b8c.
//
// Solidity: function purchaseWolk(_buyer address) returns(uint256)
func (_WolkExchange *WolkExchangeTransactor) PurchaseWolk(opts *bind.TransactOpts, _buyer common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "purchaseWolk", _buyer)
}

// PurchaseWolk is a paid mutator transaction binding the contract method 0x3d8c9b8c.
//
// Solidity: function purchaseWolk(_buyer address) returns(uint256)
func (_WolkExchange *WolkExchangeSession) PurchaseWolk(_buyer common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.PurchaseWolk(&_WolkExchange.TransactOpts, _buyer)
}

// PurchaseWolk is a paid mutator transaction binding the contract method 0x3d8c9b8c.
//
// Solidity: function purchaseWolk(_buyer address) returns(uint256)
func (_WolkExchange *WolkExchangeTransactorSession) PurchaseWolk(_buyer common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.PurchaseWolk(&_WolkExchange.TransactOpts, _buyer)
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_WolkExchange *WolkExchangeTransactor) Refund(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "refund")
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_WolkExchange *WolkExchangeSession) Refund() (*types.Transaction, error) {
	return _WolkExchange.Contract.Refund(&_WolkExchange.TransactOpts)
}

// Refund is a paid mutator transaction binding the contract method 0x590e1ae3.
//
// Solidity: function refund() returns()
func (_WolkExchange *WolkExchangeTransactorSession) Refund() (*types.Transaction, error) {
	return _WolkExchange.Contract.Refund(&_WolkExchange.TransactOpts)
}

// RemoveParticipant is a paid mutator transaction binding the contract method 0xe814c941.
//
// Solidity: function removeParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) RemoveParticipant(opts *bind.TransactOpts, _participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "removeParticipant", _participants)
}

// RemoveParticipant is a paid mutator transaction binding the contract method 0xe814c941.
//
// Solidity: function removeParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeSession) RemoveParticipant(_participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.RemoveParticipant(&_WolkExchange.TransactOpts, _participants)
}

// RemoveParticipant is a paid mutator transaction binding the contract method 0xe814c941.
//
// Solidity: function removeParticipant(_participants address[]) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) RemoveParticipant(_participants []common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.RemoveParticipant(&_WolkExchange.TransactOpts, _participants)
}

// SellWolk is a paid mutator transaction binding the contract method 0x00310e16.
//
// Solidity: function sellWolk(_wolkAmount uint256) returns(uint256)
func (_WolkExchange *WolkExchangeTransactor) SellWolk(opts *bind.TransactOpts, _wolkAmount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "sellWolk", _wolkAmount)
}

// SellWolk is a paid mutator transaction binding the contract method 0x00310e16.
//
// Solidity: function sellWolk(_wolkAmount uint256) returns(uint256)
func (_WolkExchange *WolkExchangeSession) SellWolk(_wolkAmount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.SellWolk(&_WolkExchange.TransactOpts, _wolkAmount)
}

// SellWolk is a paid mutator transaction binding the contract method 0x00310e16.
//
// Solidity: function sellWolk(_wolkAmount uint256) returns(uint256)
func (_WolkExchange *WolkExchangeTransactorSession) SellWolk(_wolkAmount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.SellWolk(&_WolkExchange.TransactOpts, _wolkAmount)
}

// SetExchangeFormula is a paid mutator transaction binding the contract method 0x2659d8ef.
//
// Solidity: function setExchangeFormula(_newExchangeformula address) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) SetExchangeFormula(opts *bind.TransactOpts, _newExchangeformula common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "setExchangeFormula", _newExchangeformula)
}

// SetExchangeFormula is a paid mutator transaction binding the contract method 0x2659d8ef.
//
// Solidity: function setExchangeFormula(_newExchangeformula address) returns(success bool)
func (_WolkExchange *WolkExchangeSession) SetExchangeFormula(_newExchangeformula common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.SetExchangeFormula(&_WolkExchange.TransactOpts, _newExchangeformula)
}

// SetExchangeFormula is a paid mutator transaction binding the contract method 0x2659d8ef.
//
// Solidity: function setExchangeFormula(_newExchangeformula address) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) SetExchangeFormula(_newExchangeformula common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.SetExchangeFormula(&_WolkExchange.TransactOpts, _newExchangeformula)
}

// TokenGenerationEvent is a paid mutator transaction binding the contract method 0xfa6b129d.
//
// Solidity: function tokenGenerationEvent(_participant address) returns()
func (_WolkExchange *WolkExchangeTransactor) TokenGenerationEvent(opts *bind.TransactOpts, _participant common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "tokenGenerationEvent", _participant)
}

// TokenGenerationEvent is a paid mutator transaction binding the contract method 0xfa6b129d.
//
// Solidity: function tokenGenerationEvent(_participant address) returns()
func (_WolkExchange *WolkExchangeSession) TokenGenerationEvent(_participant common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.TokenGenerationEvent(&_WolkExchange.TransactOpts, _participant)
}

// TokenGenerationEvent is a paid mutator transaction binding the contract method 0xfa6b129d.
//
// Solidity: function tokenGenerationEvent(_participant address) returns()
func (_WolkExchange *WolkExchangeTransactorSession) TokenGenerationEvent(_participant common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.TokenGenerationEvent(&_WolkExchange.TransactOpts, _participant)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.Transfer(&_WolkExchange.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(_to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.Transfer(&_WolkExchange.TransactOpts, _to, _value)
}

// TransferAnyERC20Token is a paid mutator transaction binding the contract method 0xdc39d06d.
//
// Solidity: function transferAnyERC20Token(_tokenAddress address, _amount uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) TransferAnyERC20Token(opts *bind.TransactOpts, _tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "transferAnyERC20Token", _tokenAddress, _amount)
}

// TransferAnyERC20Token is a paid mutator transaction binding the contract method 0xdc39d06d.
//
// Solidity: function transferAnyERC20Token(_tokenAddress address, _amount uint256) returns(success bool)
func (_WolkExchange *WolkExchangeSession) TransferAnyERC20Token(_tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferAnyERC20Token(&_WolkExchange.TransactOpts, _tokenAddress, _amount)
}

// TransferAnyERC20Token is a paid mutator transaction binding the contract method 0xdc39d06d.
//
// Solidity: function transferAnyERC20Token(_tokenAddress address, _amount uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) TransferAnyERC20Token(_tokenAddress common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferAnyERC20Token(&_WolkExchange.TransactOpts, _tokenAddress, _amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferFrom(&_WolkExchange.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(_from address, _to address, _value uint256) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferFrom(&_WolkExchange.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_WolkExchange *WolkExchangeTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_WolkExchange *WolkExchangeSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferOwnership(&_WolkExchange.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(_newOwner address) returns()
func (_WolkExchange *WolkExchangeTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.TransferOwnership(&_WolkExchange.TransactOpts, _newOwner)
}

// UpdatePurchasePossible is a paid mutator transaction binding the contract method 0xe2542f03.
//
// Solidity: function updatePurchasePossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) UpdatePurchasePossible(opts *bind.TransactOpts, _isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "updatePurchasePossible", _isRunning)
}

// UpdatePurchasePossible is a paid mutator transaction binding the contract method 0xe2542f03.
//
// Solidity: function updatePurchasePossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeSession) UpdatePurchasePossible(_isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdatePurchasePossible(&_WolkExchange.TransactOpts, _isRunning)
}

// UpdatePurchasePossible is a paid mutator transaction binding the contract method 0xe2542f03.
//
// Solidity: function updatePurchasePossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) UpdatePurchasePossible(_isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdatePurchasePossible(&_WolkExchange.TransactOpts, _isRunning)
}

// UpdateRequireKYC is a paid mutator transaction binding the contract method 0x62ac6115.
//
// Solidity: function updateRequireKYC(_kycRequirement bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) UpdateRequireKYC(opts *bind.TransactOpts, _kycRequirement bool) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "updateRequireKYC", _kycRequirement)
}

// UpdateRequireKYC is a paid mutator transaction binding the contract method 0x62ac6115.
//
// Solidity: function updateRequireKYC(_kycRequirement bool) returns(success bool)
func (_WolkExchange *WolkExchangeSession) UpdateRequireKYC(_kycRequirement bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateRequireKYC(&_WolkExchange.TransactOpts, _kycRequirement)
}

// UpdateRequireKYC is a paid mutator transaction binding the contract method 0x62ac6115.
//
// Solidity: function updateRequireKYC(_kycRequirement bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) UpdateRequireKYC(_kycRequirement bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateRequireKYC(&_WolkExchange.TransactOpts, _kycRequirement)
}

// UpdateReserveRatio is a paid mutator transaction binding the contract method 0xe469185a.
//
// Solidity: function updateReserveRatio(_newReserveRatio uint8) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) UpdateReserveRatio(opts *bind.TransactOpts, _newReserveRatio uint8) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "updateReserveRatio", _newReserveRatio)
}

// UpdateReserveRatio is a paid mutator transaction binding the contract method 0xe469185a.
//
// Solidity: function updateReserveRatio(_newReserveRatio uint8) returns(success bool)
func (_WolkExchange *WolkExchangeSession) UpdateReserveRatio(_newReserveRatio uint8) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateReserveRatio(&_WolkExchange.TransactOpts, _newReserveRatio)
}

// UpdateReserveRatio is a paid mutator transaction binding the contract method 0xe469185a.
//
// Solidity: function updateReserveRatio(_newReserveRatio uint8) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) UpdateReserveRatio(_newReserveRatio uint8) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateReserveRatio(&_WolkExchange.TransactOpts, _newReserveRatio)
}

// UpdateSellPossible is a paid mutator transaction binding the contract method 0xaad935af.
//
// Solidity: function updateSellPossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) UpdateSellPossible(opts *bind.TransactOpts, _isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "updateSellPossible", _isRunning)
}

// UpdateSellPossible is a paid mutator transaction binding the contract method 0xaad935af.
//
// Solidity: function updateSellPossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeSession) UpdateSellPossible(_isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateSellPossible(&_WolkExchange.TransactOpts, _isRunning)
}

// UpdateSellPossible is a paid mutator transaction binding the contract method 0xaad935af.
//
// Solidity: function updateSellPossible(_isRunning bool) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) UpdateSellPossible(_isRunning bool) (*types.Transaction, error) {
	return _WolkExchange.Contract.UpdateSellPossible(&_WolkExchange.TransactOpts, _isRunning)
}

// WolkGenesis is a paid mutator transaction binding the contract method 0x917d2be2.
//
// Solidity: function wolkGenesis(_startBlock uint256, _endTime uint256, _wolkinc address) returns(success bool)
func (_WolkExchange *WolkExchangeTransactor) WolkGenesis(opts *bind.TransactOpts, _startBlock *big.Int, _endTime *big.Int, _wolkinc common.Address) (*types.Transaction, error) {
	return _WolkExchange.contract.Transact(opts, "wolkGenesis", _startBlock, _endTime, _wolkinc)
}

// WolkGenesis is a paid mutator transaction binding the contract method 0x917d2be2.
//
// Solidity: function wolkGenesis(_startBlock uint256, _endTime uint256, _wolkinc address) returns(success bool)
func (_WolkExchange *WolkExchangeSession) WolkGenesis(_startBlock *big.Int, _endTime *big.Int, _wolkinc common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.WolkGenesis(&_WolkExchange.TransactOpts, _startBlock, _endTime, _wolkinc)
}

// WolkGenesis is a paid mutator transaction binding the contract method 0x917d2be2.
//
// Solidity: function wolkGenesis(_startBlock uint256, _endTime uint256, _wolkinc address) returns(success bool)
func (_WolkExchange *WolkExchangeTransactorSession) WolkGenesis(_startBlock *big.Int, _endTime *big.Int, _wolkinc common.Address) (*types.Transaction, error) {
	return _WolkExchange.Contract.WolkGenesis(&_WolkExchange.TransactOpts, _startBlock, _endTime, _wolkinc)
}
