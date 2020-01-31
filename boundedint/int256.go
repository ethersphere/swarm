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

package boundedint

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

// Int256 represents an signed integer of 256 bits
type Int256 struct {
	value big.Int
}

var minInt256 = new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)) // -(2^255)
var maxInt256 = new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1))  // 2^255 - 1

// NewInt256 creates a Int256 struct with a minimum initial underlying value
func NewInt256() *Int256 {
	u := new(Int256)
	u.value = *new(big.Int).Set(minInt256)
	return u
}

// Int64ToInt256 creates a Int256 struct based on the given int64 param
// any int64 is valid as a Int256
func Int64ToInt256(base int64) *Int256 {
	u := NewInt256()
	u.value = *new(big.Int).SetInt64(base)
	return u
}

// Value returns the underlying private value for a Int256 struct
func (u *Int256) Value() big.Int {
	return u.value
}

// Set assigns the underlying value of the given Int256 param to u, and returns the modified receiver struct
// returns an error when the result falls outside of the unsigned 256-bit integer range
func (u *Int256) Set(value big.Int) (*Int256, error) {
	if value.Cmp(maxInt256) == 1 {
		return nil, fmt.Errorf("cannot set Int256 to %v as it overflows max value of %v", value, maxInt256)
	}
	if value.Cmp(minInt256) == -1 {
		return nil, fmt.Errorf("cannot set Int256 to %v as it underflows min value of %v", value, minInt256)
	}
	u.value = *new(big.Int).Set(&value)
	return u, nil
}

// Copy sets the underlying value of u to a copy of the given Int256 param, and returns the modified receiver struct
func (u *Int256) Copy(v *Int256) *Int256 {
	u.value = *new(big.Int).Set(&v.value)
	return u
}

// Cmp calls the underlying Cmp method for the big.Int stored in a Int256 struct as its value field
func (u *Int256) Cmp(v *Int256) int {
	return u.value.Cmp(&v.value)
}

// Equals returns true if the two Int256 structs have the same underlying values, false otherwise
func (u *Int256) Equals(v *Int256) bool {
	return u.Cmp(v) == 0
}

// Add sets u to augend + addend and returns u as the sum
// returns an error when the result falls outside of the signed 256-bit integer range
func (u *Int256) Add(augend, addend *Int256) (*Int256, error) {
	sum := new(big.Int).Add(&augend.value, &addend.value)
	return u.Set(*sum)
}

// Sub sets u to minuend - subtrahend and returns u as the difference
// returns an error when the result falls outside of the signed 256-bit integer range
func (u *Int256) Sub(minuend, subtrahend *Int256) (*Int256, error) {
	difference := new(big.Int).Sub(&minuend.value, &subtrahend.value)
	return u.Set(*difference)
}

// Mul sets u to multiplicand * multiplier and returns u as the product
// returns an error when the result falls outside of the signed 256-bit integer range
func (u *Int256) Mul(multiplicand, multiplier *Int256) (*Int256, error) {
	product := new(big.Int).Mul(&multiplicand.value, &multiplier.value)
	return u.Set(*product)
}

// String returns the string representation for Int256 structs
func (u *Int256) String() string {
	return u.value.String()
}

// MarshalJSON implements the json.Marshaler interface
// it specifies how to marshal a Int256 struct so that it can be written to disk
func (u Int256) MarshalJSON() ([]byte, error) {
	return []byte(u.value.String()), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
// it specifies how to unmarshal a Int256 struct so that it can be reconstructed from disk
func (u *Int256) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	var value big.Int
	_, ok := value.SetString(string(b), 10)
	if !ok {
		return fmt.Errorf("not a valid integer value: %s", b)
	}
	_, err := u.Set(value)
	return err
}

// EncodeRLP implements the rlp.Encoder interface
// it makes sure the value field is encoded even though it is private
func (u *Int256) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &u.value)
}

// DecodeRLP implements the rlp.Decoder interface
// it makes sure the value field is decoded even though it is private
func (u *Int256) DecodeRLP(s *rlp.Stream) error {
	if err := s.Decode(&u.value); err != nil {
		return nil
	}
	_, err := u.Set(u.value)
	return err
}
