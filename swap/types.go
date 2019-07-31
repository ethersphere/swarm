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

package swap

import (
	"github.com/ethereum/go-ethereum/common"
)

// TODO: add handshake protocol where we exchange last cheque (this is useful if one node disconnects)
// FIXME: Check the contract bytecode of the counterparty

// ChequeParams encapsulate all cheque parameters
type ChequeParams struct {
	Contract    common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary common.Address // address of the beneficiary, the contract which will redeem the cheque
	Serial      uint64         // monotonically increasing serial number
	Amount      uint64         // cumulative amount of the cheque in currency
	Honey       uint64         // amount of honey which resulted in the cumulative currency difference
	Timeout     uint64         // timeout for cashing in
}

// Cheque encapsulates the parameters and the signature
// TODO: There should be a request cheque struct that only gives the Serial
type Cheque struct {
	ChequeParams
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

// ErrorMsg is sent in case of an error TODO: specify error conditions and when this needs to be sent
type ErrorMsg struct{}
