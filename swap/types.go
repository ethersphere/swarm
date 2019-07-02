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

type Cheque struct {
	Contract    common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary common.Address
	Serial      uint64 // cumulative amount of all funds sent
	Amount      uint64 // cumulative amount of all funds sent
	Timeout     uint64
	Sig         []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

type ChequeRequestMsg struct {
}
type EmitChequeMsg struct {
	Cheque Cheque
}
type ErrorMsg struct{}
type ConfirmMsg struct{}
