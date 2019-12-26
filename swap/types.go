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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ChequeParams encapsulate all cheque parameters
type ChequeParams struct {
	Contract         common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary      common.Address // address of the beneficiary, the contract which will redeem the cheque
	CumulativePayout *Uint256       // cumulative amount of the cheque in currency
}

// Cheque encapsulates the parameters and the signature
type Cheque struct {
	ChequeParams
	Honey     uint64 // amount of honey which resulted in the cumulative currency difference
	Signature []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

// HandshakeMsg is exchanged on peer handshake
type HandshakeMsg struct {
	ChainID         uint64         // chain id of the blockchain the peer is connected to
	ContractAddress common.Address // chequebook contract address of the peer
}

// EmitChequeMsg is sent from the debitor to the creditor with the actual cheque
type EmitChequeMsg struct {
	Cheque *Cheque
}

// ConfirmChequeMsg is sent from the creditor to the debitor with the cheque to confirm successful processing
type ConfirmChequeMsg struct {
	Cheque *Cheque
}

// Uint256 represents an unsigned integer of 256 bits
type Uint256 struct {
	value *big.Int
}

var minUint256 = big.NewInt(0)
var maxUint256, _ = new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // 2^256 - 1 (base 10)

// Value returns the underlying big.Int pointer for the Uint256 struct
func (u *Uint256) Value() *big.Int {
	return u.value
}

// Set assigns a new value to the underlying pointer within the unsigned 256-bit integer range
func (u *Uint256) Set(value *big.Int) error {
	if value.Cmp(minUint256) == 1 {
		return fmt.Errorf("cannot set uint256 to %v as it overflows max value of %v", value, maxUint256)
	}
	if value.Cmp(maxUint256) == -1 {
		return fmt.Errorf("cannot set uint256 to %v as it underflows min value of %v", value, minUint256)
	}
	u.value = value
	return nil
}

// Add attempts to add the given unsigned 256-bit integer to another
func (u *Uint256) Add(addend *Uint256) error {
	var summand *big.Int
	summand.Add(u.Value(), addend.Value()) // any uint256 is good enough for a big.Int
	return u.Set(summand)
}

// Sub attempts to subtract the given unsigned 256-bit integer from another
func (u *Uint256) Sub(subtrahend *Uint256) error {
	var difference *big.Int
	difference.Sub(u.Value(), subtrahend.Value()) // any uint256 is good enough for a big.Int
	return u.Set(difference)
}

// Cmp calls the underlying Cmp method for the big.Int stored in a Uint256 struct
func (u *Uint256) Cmp(v *Uint256) (r int) {
	return u.Value().Cmp(v.Value())
}
