// Copyright 2020 The Swarm Authors
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

package int256

import (
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/rlp"
)

// Uint256 represents an unsigned integer of 256 bits
type Uint256 struct {
	value *big.Int
}

var minUint256 = big.NewInt(0)
var maxUint256 = new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)) // 2^256 - 1

// NewUint256 creates a Uint256 struct with an initial underlying value of the given param
// returns an error when the value cannot be correctly set
func NewUint256(value *big.Int) (*Uint256, error) {
	u := new(Uint256)
	return u.set(value)
}

// Uint256From creates a Uint256 struct based on the given uint64 param
// any uint64 is valid as a Uint256
func Uint256From(base uint64) *Uint256 {
	u := new(Uint256)
	u.value = new(big.Int).SetUint64(base)
	return u
}

// Copy creates and returns a new Int256 instance, with its underlying value set matching the receiver
func (u *Uint256) Copy() *Uint256 {
	v := new(Uint256)
	v.value = new(big.Int).Set(u.value)
	return v
}

// Value returns the underlying private value for a Uint256 struct
func (u *Uint256) Value() *big.Int {
	return new(big.Int).Set(u.value)
}

// set assigns the underlying value of the given Uint256 param to u, and returns the modified receiver struct
// returns an error when the value cannot be correctly set
func (u *Uint256) set(value *big.Int) (*Uint256, error) {
	if err := checkUint256Bounds(value); err != nil {
		return nil, err
	}
	if u.value == nil {
		u.value = new(big.Int)
	}
	u.value.Set(value)
	return u, nil
}

// checkUint256NBounds returns an error when the given value falls outside of the unsigned 256-bit integer range or is nil
// returns nil otherwise
func checkUint256Bounds(value *big.Int) error {
	if value == nil {
		return fmt.Errorf("cannot set Uint256 to a nil value")
	}
	if value.Cmp(maxUint256) == 1 {
		return fmt.Errorf("cannot set Uint256 to %v as it overflows max value of %v", value, maxUint256)
	}
	if value.Cmp(minUint256) == -1 {
		return fmt.Errorf("cannot set Uint256 to %v as it underflows min value of %v", value, minUint256)
	}
	return nil
}

// Add sets u to augend + addend and returns u as the sum
// returns an error when the value cannot be correctly set
func (u *Uint256) Add(augend, addend *Uint256) (*Uint256, error) {
	sum := new(big.Int).Add(augend.value, addend.value)
	return u.set(sum)
}

// Sub sets u to minuend - subtrahend and returns u as the difference
// returns an error when the value cannot be correctly set
func (u *Uint256) Sub(minuend, subtrahend *Uint256) (*Uint256, error) {
	difference := new(big.Int).Sub(minuend.value, subtrahend.value)
	return u.set(difference)
}

// Mul sets u to multiplicand * multiplier and returns u as the product
// returns an error when the value cannot be correctly set
func (u *Uint256) Mul(multiplicand, multiplier *Uint256) (*Uint256, error) {
	product := new(big.Int).Mul(multiplicand.value, multiplier.value)
	return u.set(product)
}

// Cmp calls the underlying Cmp method for the big.Int stored in a Uint256 struct as its value field
func (u *Uint256) Cmp(v *Uint256) int {
	return u.value.Cmp(v.value)
}

// Equals returns true if the two Uint256 structs have the same underlying values, false otherwise
func (u *Uint256) Equals(v *Uint256) bool {
	return u.Cmp(v) == 0
}

// String returns the string representation for Uint256 structs
func (u *Uint256) String() string {
	return u.value.String()
}

// MarshalJSON implements the json.Marshaler interface
// it specifies how to marshal a Uint256 struct so that it can be written to disk
func (u *Uint256) MarshalJSON() ([]byte, error) {
	// number is wrapped in quotes to prevent json number overflowing
	return []byte(strconv.Quote(u.value.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// it specifies how to unmarshal a Uint256 struct so that it can be reconstructed from disk
func (u *Uint256) UnmarshalJSON(b []byte) error {
	var value big.Int
	// value string must be unquoted due to marshaling
	strValue, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	_, ok := (&value).SetString(strValue, 10)
	if !ok {
		return fmt.Errorf("not a valid integer value: %s", b)
	}

	if err := checkUint256Bounds(&value); err != nil {
		return err
	}
	_, err = u.set(&value)
	return err
}

// EncodeRLP implements the rlp.Encoder interface
// it makes sure the value field is encoded even though it is private
func (u *Uint256) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &u.value)
}

// DecodeRLP implements the rlp.Decoder interface
// it makes sure the value field is decoded even though it is private
func (u *Uint256) DecodeRLP(s *rlp.Stream) error {
	if err := s.Decode(&u.value); err != nil {
		return err
	}
	if err := checkUint256Bounds(u.value); err != nil {
		return err
	}
	return nil
}
