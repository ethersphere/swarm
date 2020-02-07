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
	"crypto/rand"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ethersphere/swarm/state"
)

// TestSet tests the creation of valid and invalid Int256 structs by calling the Set function
func TestInt256Set(t *testing.T) {
	testCases := []BoundedIntTestCase{
		{
			name:         "base 0",
			baseInteger:  big.NewInt(0),
			expectsError: false,
		},
		// negative numbers
		{
			name:         "base -1",
			baseInteger:  big.NewInt(-1),
			expectsError: false,
		},
		{
			name:         "base -1 * 2^8",
			baseInteger:  new(big.Int).Mul(new(big.Int).Exp(big.NewInt(2), big.NewInt(8), nil), big.NewInt(-1)),
			expectsError: false,
		},
		{
			name:         "base -1 * 2^64",
			baseInteger:  new(big.Int).Mul(new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil), big.NewInt(-1)),
			expectsError: false,
		},
		{
			name:         "base -1 * 2^256",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(-1)),
			expectsError: true,
		},
		{
			name:         "base -1 * 2^512",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil), big.NewInt(-1)),
			expectsError: true,
		},
		// positive numbers
		{
			name:         "base 1",
			baseInteger:  big.NewInt(1),
			expectsError: false,
		},
		{
			name:         "base 2^8",
			baseInteger:  new(big.Int).Exp(big.NewInt(2), big.NewInt(8), nil),
			expectsError: false,
		},
		{
			name:         "base 2^128",
			baseInteger:  new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
			expectsError: false,
		},
		{
			name:         "base 2^255 - 1",
			baseInteger:  new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1)),
			expectsError: false,
		},
		{
			name:         "base 2^255",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil), big.NewInt(1)),
			expectsError: true,
		},
		{
			name:         "base 2^512",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil), big.NewInt(1)),
			expectsError: true,
		},
	}

	testInt256Set(t, testCases)
}

func testInt256Set(t *testing.T, testCases []BoundedIntTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewInt256().Set(*tc.baseInteger)
			if tc.expectsError && err == nil {
				t.Fatalf("expected error when creating new Int256, but got none")
			}
			if !tc.expectsError {
				if err != nil {
					t.Fatalf("got unexpected error when creating new Int256: %v", err)
				}
				resultValue := result.Value()
				if (&resultValue).Cmp(tc.baseInteger) != 0 {
					t.Fatalf("expected value of %v, got %v instead", tc.baseInteger, result.value)
				}
			}
		})
	}
}

// TestCopy tests the duplication of an existing Int256 variable
func TestInt256Copy(t *testing.T) {
	r, err := randomInt256()
	if err != nil {
		t.Fatal(err)
	}

	c := NewInt256().Copy(r)

	if !c.Equals(r) {
		t.Fatalf("copy of Int256 %v has an unequal value of %v", r, c)
	}
}

func randomInt256() (*Int256, error) {
	r, err := rand.Int(rand.Reader, new(big.Int).Sub(maxInt256, minInt256)) // base for random
	if err != nil {
		return nil, err
	}

	randomInt256 := new(big.Int).Add(r, minInt256) // random is within [minInt256, maxInt256]

	return NewInt256().Set(*randomInt256)
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

	r, err := randomUint256()
	if err != nil {
		t.Fatal(err)
	}

	k := r.String()

	stateStore.Put(k, r)

	var u *Uint256
	err = stateStore.Get(k, &u)
	if err != nil {
		t.Fatal(err)
	}

	if !u.Equals(r) {
		t.Fatalf("retrieved Int256 %v has an unequal balance to the original Int256 %v", u, r)
	}
}
