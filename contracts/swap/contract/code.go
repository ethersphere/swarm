// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package swap package wraps the 'swap' Ethereum smart contract.
// It is an abstraction layer to hide implementation details about the different
// Swap contract iterations (SimpleSwap, Swap, etc.)
package contract

// ContractDeployedCode is used to detect suicides. This constant needs to be
// updated when the contract code is changed.
// **CURRENT: SIMPLESWAP**
//simpleSwapBin
const ContractDeployedCode = "0x60806040526004361061012a5760003560e01c8063946f46a2116100ab578063d3000b8b1161006f578063d3000b8b14610992578063df32438014610a35578063e0bcf13a14610b27578063e3bb7aec14610b52578063f3c08b1f14610c78578063f890673b14610cfd5761012a565b8063946f46a2146107dc578063b6343b0d1461082d578063b7770350146108a7578063b7ec1a3314610902578063c76a4d311461092d5761012a565b806339d9ec4c116100f257806339d9ec4c1461038c5780634f823a4c146103b757806354fe2614146105545780636162913b1461065a5780636c16f684146106d45761012a565b8063030aca3e146101a15780631d143848146102445780632329d2a81461029b5780632e1a7d4d146102f6578063338f3fed14610331575b600034111561019f577fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c3334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a15b005b3480156101ad57600080fd5b5061022e600480360360a08110156101c457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050610de3565b6040518082815260200191505060405180910390f35b34801561025057600080fd5b50610259610e95565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156102a757600080fd5b506102f4600480360360408110156102be57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610ebb565b005b34801561030257600080fd5b5061032f6004803603602081101561031957600080fd5b8101908080359060200190929190505050610ecc565b005b34801561033d57600080fd5b5061038a6004803603604081101561035457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061105b565b005b34801561039857600080fd5b506103a1611280565b6040518082815260200191505060405180910390f35b3480156103c357600080fd5b50610552600480360360c08110156103da57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190803590602001909291908035906020019064010000000081111561043557600080fd5b82018360208201111561044757600080fd5b8035906020019184600183028401116401000000008311171561046957600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803590602001906401000000008111156104cc57600080fd5b8201836020820111156104de57600080fd5b8035906020019184600183028401116401000000008311171561050057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050611286565b005b34801561056057600080fd5b50610658600480360360a081101561057757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190803590602001906401000000008111156105d257600080fd5b8201836020820111156105e457600080fd5b8035906020019184600183028401116401000000008311171561060657600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061140b565b005b34801561066657600080fd5b506106a96004803603602081101561067d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061157a565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b3480156106e057600080fd5b50610761600480360360a08110156106f757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190803590602001909291905050506115aa565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156107a1578082015181840152602081019050610786565b50505050905090810190601f1680156107ce5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156107e857600080fd5b5061082b600480360360208110156107ff57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611655565b005b34801561083957600080fd5b5061087c6004803603602081101561085057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506117a8565b6040518085815260200184815260200183815260200182815260200194505050505060405180910390f35b3480156108b357600080fd5b50610900600480360360408110156108ca57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506117d8565b005b34801561090e57600080fd5b506109176119c0565b6040518082815260200191505060405180910390f35b34801561093957600080fd5b5061097c6004803603602081101561095057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506119f3565b6040518082815260200191505060405180910390f35b34801561099e57600080fd5b50610a1f600480360360a08110156109b557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035906020019092919080359060200190929190505050611a58565b6040518082815260200191505060405180910390f35b348015610a4157600080fd5b50610b2560048036036060811015610a5857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190640100000000811115610a9f57600080fd5b820183602082011115610ab157600080fd5b80359060200191846001830284011164010000000083111715610ad357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050611b12565b005b348015610b3357600080fd5b50610b3c611d80565b6040518082815260200191505060405180910390f35b348015610b5e57600080fd5b50610c76600480360360c0811015610b7557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190640100000000811115610bdc57600080fd5b820183602082011115610bee57600080fd5b80359060200191846001830284011164010000000083111715610c1057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192908035906020019092919080359060200190929190505050611d86565b005b348015610c8457600080fd5b50610cfb60048036036080811015610c9b57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080359060200190929190505050611fa2565b005b348015610d0957600080fd5b50610de160048036036080811015610d2057600080fd5b8101908080359060200190929190803590602001909291908035906020019092919080359060200190640100000000811115610d5b57600080fd5b820183602082011115610d6d57600080fd5b80359060200191846001830284011164010000000083111715610d8f57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050612328565b005b60008585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018381526020018281526020019550505050505060405160208183030381529060405280519060200120905095945050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b610ec83383836000611fa2565b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610f8f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b610f976119c0565b811115610fef576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260288152602001806128d56028913960400191505060405180910390fd5b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611057573d6000803e3d6000fd5b5050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461111e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b3073ffffffffffffffffffffffffffffffffffffffff163161114b8260035461241290919063ffffffff16565b11156111a2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260358152602001806128a06035913960400191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090506111fc82826000015461241290919063ffffffff16565b81600001819055506112198260035461241290919063ffffffff16565b600381905550600081600201819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a2505050565b60005481565b61129c6112963088888888610de3565b8361249a565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461135e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b61137461136e3088888888610de3565b8261249a565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff16146113f7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806127ec6022913960400191505060405180910390fd5b611403868686866124b6565b505050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146114ce576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6114e46114de3087878787610de3565b8261249a565b73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff1614611567576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806127ec6022913960400191505060405180910390fd5b611573858585856124b6565b5050505050565b60016020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b60608585858585604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b815260140184815260200183815260200182815260200195505050505050604051602081830303815290604052905095945050505050565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050806003015442101580156116b157506000816003015414155b611706576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602581526020018061280e6025913960400191505060405180910390fd5b611721816001015482600001546125ed90919063ffffffff16565b81600001819055506000816003018190555061174c81600101546003546125ed90919063ffffffff16565b6003819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040518082815260200191505060405180910390a25050565b60026020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461189b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b6000600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000209050806000015482111561193b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260278152602001806128576027913960400191505060405180910390fd5b600080826002015414611952578160020154611956565b6000545b905080420182600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a846040518082815260200191505060405180910390a250505050565b60006119ee6003543073ffffffffffffffffffffffffffffffffffffffff16316125ed90919063ffffffff16565b905090565b6000611a51600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154611a436119c0565b61241290919063ffffffff16565b9050919050565b6000611b078686868686604051602001808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018481526020018381526020018281526020019550505050505060405160208183030381529060405280519060200120612676565b905095945050505050565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611bd5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260168152602001807f53696d706c65537761703a206e6f74206973737565720000000000000000000081525060200191505060405180910390fd5b611c6d308484604051602001808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019350505050604051602081830303815290604052805190602001208261249a565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614611ca457600080fd5b81600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f86b5d1492f68620b7cc58d71bd1380193d46a46d90553b73e919e0c6f319fe1f600260008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201546040518082815260200191505060405180910390a2505050565b60035481565b81421115611ddf576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602281526020018061287e6022913960400191505060405180910390fd5b611ec4303386888686604051602001808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018581526020018473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018381526020018281526020019650505050505050604051602081830303815290604052805190602001208461249a565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff1614611f47576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260228152602001806127ec6022913960400191505060405180910390fd5b611f5386868684611fa2565b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611f99573d6000803e3d6000fd5b50505050505050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015612042576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260248152602001806128336024913960400191505060405180910390fd5b61205d816002015482600101546125ed90919063ffffffff16565b8311156120b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260238152602001806128fd6023913960400191505060405180910390fd5b600061210384600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546126ce565b9050600061211a85836121146119c0565b016126ce565b9050600082146121db5761217982600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001546125ed90919063ffffffff16565b600260008973ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600001819055506121d4826003546125ed90919063ffffffff16565b6003819055505b6121f281846002015461241290919063ffffffff16565b83600201819055508573ffffffffffffffffffffffffffffffffffffffff166108fc61222786846125ed90919063ffffffff16565b9081150290604051600060405180830381858888f19350505050158015612252573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff168873ffffffffffffffffffffffffffffffffffffffff167f5920b90d620e15c47f9e2f42adac6a717078eb0403d85477ad9be9493458ed138660000154858a8a6040518085815260200184815260200183815260200182815260200194505050505060405180910390a480851461231f577f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b50505050505050565b61233e6123383033878787610de3565b8261249a565b73ffffffffffffffffffffffffffffffffffffffff16600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614612400576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601d8152602001807f53696d706c65537761703a20696e76616c69642069737375657253696700000081525060200191505060405180910390fd5b61240c338585856124b6565b50505050565b600080828401905083811015612490576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601b8152602001807f536166654d6174683a206164646974696f6e206f766572666c6f77000000000081525060200191505060405180910390fd5b8091505092915050565b60006124ae6124a884612676565b836126e7565b905092915050565b6000600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905080600001548411612572576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601a8152602001807f53696d706c65537761703a20696e76616c69642073657269616c00000000000081525060200191505060405180910390fd5b8381600001819055508281600101819055508142018160030181905550838573ffffffffffffffffffffffffffffffffffffffff167f543b37a2abe69e287f27911f3802739c2f6271e8eb02ae6303a3cd9443bac03c8585604051808381526020018281526020019250505060405180910390a35050505050565b600082821115612665576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252601e8152602001807f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525060200191505060405180910390fd5b600082840390508091505092915050565b60008160405160200180807f19457468657265756d205369676e6564204d6573736167653a0a333200000000815250601c01828152602001915050604051602081830303815290604052805190602001209050919050565b60008183106126dd57816126df565b825b905092915050565b600060418251146126fb57600090506127e5565b60008060006020850151925060408501519150606085015160001a90507f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08260001c111561274f57600093505050506127e5565b601b8160ff16141580156127675750601c8160ff1614155b1561277857600093505050506127e5565b60018682858560405160008152602001604052604051808581526020018460ff1660ff1681526020018381526020018281526020019450505050506020604051602081039080840390855afa1580156127d5573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fe53696d706c65537761703a20696e76616c69642062656e656669636961727953696753696d706c65537761703a206465706f736974206e6f74207965742074696d6564206f757453696d706c65537761703a20636865717565206e6f74207965742074696d6564206f757453696d706c65537761703a2068617264206465706f736974206e6f742073756666696369656e7453696d706c65537761703a2062656e6566696369617279536967206578706972656453696d706c65537761703a2068617264206465706f7369742063616e6e6f74206265206d6f7265207468616e2062616c616e63652053696d706c65537761703a206c697175696442616c616e6365206e6f742073756666696369656e7453696d706c65537761703a206e6f7420656e6f7567682062616c616e6365206f776564a265627a7a723058205ce630547310e8e185ec7f6c489d9f9e096545f5c05aa5980535b0e31cd1a2e064736f6c634300050a0032"
