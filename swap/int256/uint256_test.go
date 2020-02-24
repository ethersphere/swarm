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
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ethersphere/swarm/state"
)

// the following test cases cover a range of values to be used to create uint256 variables
// these variables are expected to be created successfully when using integer values
// contained in the closed interval between 0 and 2^256
var uint256TestCases = []testCase{
	{
		name:         "case 0",
		value:        big.NewInt(0),
		expectsError: false,
	},
	// negative numbers
	{
		name:         "case -1",
		value:        big.NewInt(-1),
		expectsError: true,
	},
	{
		name:         "case -1 * 2^8",
		value:        new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(8), nil)),
		expectsError: true,
	},
	{
		name:         "case -1 * 2^64",
		value:        new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)),
		expectsError: true,
	},
	// positive numbers
	{
		name:         "case 1",
		value:        big.NewInt(1),
		expectsError: false,
	},
	{
		name:         "case 2^8",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(8), nil),
		expectsError: false,
	},
	{
		name:         "case 2^128",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		expectsError: false,
	},
	{
		name:         "case 2^256 - 1",
		value:        new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)),
		expectsError: false,
	},
	{
		name:         "case 2^256",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil),
		expectsError: true,
	},
	{
		name:         "case 2^512",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil),
		expectsError: true,
	},
}

// TestSet tests the creation of valid and invalid Uint256 structs by calling the Set function
func TestUint256Set(t *testing.T) {
	for _, tc := range uint256TestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewUint256().Set(*tc.value)
			if tc.expectsError && err == nil {
				t.Fatalf("expected error when creating new Uint256, but got none")
			}
			if !tc.expectsError {
				if err != nil {
					t.Fatalf("got unexpected error when creating new Uint256: %v", err)
				}
				resultValue := result.Value()
				if (&resultValue).Cmp(tc.value) != 0 {
					t.Fatalf("expected value of %v, got %v instead", tc.value, result.value)
				}
			}
		})
	}
}

// TestCopy tests the du:heartplication of an existing Uint256 variable
func TestUint256Copy(t *testing.T) {
	// pick test value
	i := new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil) // 2^128
	v, err := NewUint256().Set(*i)
	if err != nil {
		t.Fatalf("got unexpected error when creating new Uint256: %v", err)
	}

	// copy picked value
	c := NewUint256().Copy(v)

	if !c.Equals(v) {
		t.Fatalf("copy of Uint256 %v has an unequal value of %v", v, c)
	}

	_, err = v.Add(v, Uint256From(1))
	if err != nil {
		t.Fatalf("got unexpected error when increasing test case %v: %v", v, err)
	}

	// value of copy should not have changed
	if c.Equals(v) {
		t.Fatalf("copy of Uint256 %v had an unexpected change of value to %v", v, c)
	}
}

// TestStore indirectly tests the marshaling and unmarshaling of a random Uint256 variable
func TestUint256Store(t *testing.T) {
	testDir, err := ioutil.TempDir("", "uint256_test_store")
	if err != nil {
		t.Fatal(err)
	}

	stateStore, err := state.NewDBStore(testDir)
	defer stateStore.Close()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range uint256TestCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.expectsError {
				r, err := NewUint256().Set(*tc.value)
				if err != nil {
					t.Fatalf("got unexpected error when creating new Uint256: %v", err)
				}

				k := r.String()

				stateStore.Put(k, r)

				var u *Uint256
				err = stateStore.Get(k, &u)
				if err != nil {
					t.Fatal(err)
				}

				if !u.Equals(r) {
					t.Fatalf("retrieved Uint256 %v has an unequal balance to the original Uint256 %v", u, r)
				}
			}
		})
	}
}
