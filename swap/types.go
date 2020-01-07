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
	value big.Int
}

var minUint256 = big.NewInt(0)
var maxUint256 = new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)) // 2^256 - 1

// NewUint256 creates a Uint256 struct with an initial underlying value of 0
// no Uint256 should have a nil pointer as its value field
func NewUint256() *Uint256 {
	u := new(Uint256)
	u.value = *minUint256
	return u
}

// Uint64ToUint256 creates a Uint256 struct based on the given uint64 param
// any uint64 is valid as a Uint256
func Uint64ToUint256(base uint64) *Uint256 {
	u := NewUint256()
	u.value = *new(big.Int).SetUint64(base)
	return u
}

// Value returns the underlying private value for a Uint256 struct
func (u *Uint256) Value() *big.Int {
	return &u.value
}

// Set assigns the underlying value of the given Uint256 param to u, and returns the modified receiver struct
// returns an error when the result falls outside of the unsigned 256-bit integer range
func (u *Uint256) Set(value *big.Int) (*Uint256, error) {
	if value.Cmp(maxUint256) == 1 {
		return nil, fmt.Errorf("cannot set Uint256 to %v as it overflows max value of %v", value, maxUint256)
	}
	if value.Cmp(minUint256) == -1 {
		return nil, fmt.Errorf("cannot set Uint256 to %v as it underflows min value of %v", value, minUint256)
	}
	u.value = *value
	return u, nil
}

// Copy sets the underlying value of u to a copy of the given Uint256 param, and returns the modified receiver struct
func (u *Uint256) Copy(v *Uint256) *Uint256 {
	valueCopy := new(big.Int).Set(v.Value())
	u.value = *valueCopy
	return u
}

// Cmp calls the underlying Cmp method for the big.Int stored in a Uint256 struct as its value field
func (u *Uint256) Cmp(v *Uint256) int {
	return u.value.Cmp(v.Value())
}

// Equals returns true if the two Uint256 structs have the same underlying values, false otherwise
func (u *Uint256) Equals(v *Uint256) bool {
	return u.Cmp(v) == 0
}

// Add sets u to augend + addend and returns u as the sum
// returns an error when the result falls outside of the unsigned 256-bit integer range
func (u *Uint256) Add(augend, addend *Uint256) (*Uint256, error) {
	sum := new(big.Int).Add(augend.Value(), addend.Value())
	return u.Set(sum)
}

// Sub sets u to minuend - subtrahend and returns u as the difference
// returns an error when the result falls outside of the unsigned 256-bit integer range
func (u *Uint256) Sub(minuend, subtrahend *Uint256) (*Uint256, error) {
	difference := new(big.Int).Sub(minuend.Value(), subtrahend.Value())
	return u.Set(difference)
}

// Mul sets u to multiplicand * multiplier and returns u as the product
// returns an error when the result falls outside of the unsigned 256-bit integer range
func (u *Uint256) Mul(multiplicand, multiplier *Uint256) (*Uint256, error) {
	product := new(big.Int).Mul(multiplicand.Value(), multiplier.Value())
	return u.Set(product)
}

// String returns the string representation for Uint256 structs
func (u *Uint256) String() string {
	return u.value.String()
}

// MarshalJSON specifies how to marshal a Uint256 struct so that it can be written to disk
func (u Uint256) MarshalJSON() ([]byte, error) {
	return []byte(u.Value().String()), nil
}

// UnmarshalJSON specifies how to unmarshal a Uint256 struct so that it can be reconstructed from disk
func (u *Uint256) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	var value big.Int
	_, ok := value.SetString(string(b), 10)
	if !ok {
		return fmt.Errorf("not a valid integer value: %s", b)
	}
	u.value = value
	return nil
}
