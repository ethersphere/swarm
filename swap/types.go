// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package swap

import (
	"github.com/ethereum/go-ethereum/common"
)

// ChequeParams encapsulate all cheque parameters
type ChequeParams struct {
	Contract         common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary      common.Address // address of the beneficiary, the contract which will redeem the cheque
	CumulativePayout uint64         // cumulative amount of the cheque in currency
}

// Cheque encapsulates the parameters and the signature
type Cheque struct {
	ChequeParams
	Honey     uint64 // amount of honey which resulted in the cumulative currency difference
	Signature []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

// HandshakeMsg is exchanged on peer handshake
type HandshakeMsg struct {
	ContractAddress common.Address
}

// EmitChequeMsg is sent from the debitor to the creditor with the actual check
type EmitChequeMsg struct {
	Cheque *Cheque
}
