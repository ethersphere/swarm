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

type testCase struct {
	name         string
	value        *big.Int
	expectsError bool
}

// the following test cases cover a range of values to be used to create int256 variables
// these variables are expected to be created successfully when using integer values
// contained in the closed interval between -2^255 and 2^255 - 1
var int256TestCases = []testCase{
	{
		name:         "case 0",
		value:        big.NewInt(0),
		expectsError: false,
	},
	// negative numbers
	{
		name:         "case -1",
		value:        big.NewInt(-1),
		expectsError: false,
	},
	{
		name:         "case -1 * 2^8",
		value:        big.NewInt(-256),
		expectsError: false,
	},
	{
		name:         "case -1 * 2^64",
		value:        new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)),
		expectsError: false,
	},
	{
		name:         "case -1 * 2^255",
		value:        new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)),
		expectsError: false,
	},
	{
		name:         "case -1 * 2^255 - 1",
		value:        new(big.Int).Sub(new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)), big.NewInt(1)),
		expectsError: true,
	},
	{
		name:         "case -1 * 2^512",
		value:        new(big.Int).Mul(big.NewInt(-1), new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil)),
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
		value:        big.NewInt(256),
		expectsError: false,
	},
	{
		name:         "case 2^128",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		expectsError: false,
	},
	{
		name:         "case 2^255 - 1",
		value:        new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1)),
		expectsError: false,
	},
	{
		name:         "case 2^255",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil),
		expectsError: true,
	},
	{
		name:         "case 2^512",
		value:        new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil),
		expectsError: true,
	},
}

// TestSet tests the creation of valid and invalid Int256 structs by calling the Set function
func TestInt256Set(t *testing.T) {
	for _, tc := range int256TestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewInt256(tc.value)
			if tc.expectsError && err == nil {
				t.Fatalf("expected error when creating new Int256, but got none")
			}
			if !tc.expectsError {
				if err != nil {
					t.Fatalf("got unexpected error when creating new Int256: %v", err)
				}
				resultValue := result.Value()
				if resultValue.Cmp(tc.value) != 0 {
					t.Fatalf("expected value of %v, got %v instead", tc.value, result.value)
				}
			}
		})
	}
}

// TestCopy tests the duplication of an existing Int256 variable
func TestInt256Copy(t *testing.T) {
	// pick test value
	i := new(big.Int).Exp(big.NewInt(-2), big.NewInt(128), nil) // -2^128
	v, err := NewInt256(i)
	if err != nil {
		t.Fatalf("got unexpected error when creating new Int256: %v", err)
	}

	// copy picked value
	c := v.Copy()

	if !c.Equals(v) {
		t.Fatalf("copy of Int256 %v has an unequal value of %v", v, c)
	}

	_, err = v.Add(v, Int256From(1))
	if err != nil {
		t.Fatalf("got unexpected error when increasing test case %v: %v", v, err)
	}

	// value of copy should not have changed
	if c.Equals(v) {
		t.Fatalf("copy of Int256 %v had an unexpected change of value to %v", v, c)
	}
}

// TestIn256ValueInjection verifies that the underlying value for Int256 structs
// is not able to be manipulated through the getter
func TestIn256ValueInjection(t *testing.T) {
	i := new(big.Int).SetInt64(-12)
	v, err := NewInt256(i)
	if err != nil {
		t.Fatalf("got unexpected error when creating new Int256: %v", err)
	}

	vv := v.Value()
	vvc := v.Value()

	vv.SetInt64(-15)

	if vvc.Cmp(vv) == 0 {
		t.Fatalf("values are equal though one should change: %v %v", vv, vvc)
	}
}

// TestStore indirectly tests the marshaling and unmarshaling of a random Int256 variable
func TestInt256Store(t *testing.T) {
	testDir, err := ioutil.TempDir("", "int256_test_store")
	if err != nil {
		t.Fatal(err)
	}

	stateStore, err := state.NewDBStore(testDir)
	defer stateStore.Close()
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range int256TestCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.expectsError {
				r, err := NewInt256(tc.value)
				if err != nil {
					t.Fatalf("got unexpected error when creating new Int256: %v", err)
				}

				k := r.String()

				stateStore.Put(k, r)

				var u *Int256
				err = stateStore.Get(k, &u)
				if err != nil {
					t.Fatal(err)
				}

				if !u.Equals(r) {
					t.Fatalf("retrieved Int256 %v has an unequal balance to the original Uint256 %v", u, r)
				}
			}
		})
	}
}
