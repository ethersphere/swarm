package swap

import (
	"crypto/rand"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ethersphere/swarm/state"
)

type Uint256TestCase struct {
	name         string
	baseInteger  *big.Int
	expectsError bool
}

// TestSetUint256 tests the creation of valid and invalid Uint256 structs by calling the Set function
func TestSetUint256(t *testing.T) {
	testCases := []Uint256TestCase{
		{
			name:         "base 0",
			baseInteger:  big.NewInt(0),
			expectsError: false,
		},
		// negative numbers
		{
			name:         "base -1",
			baseInteger:  big.NewInt(-1),
			expectsError: true,
		},
		{
			name:         "base -1 * 2^8",
			baseInteger:  new(big.Int).Mul(new(big.Int).Exp(big.NewInt(2), big.NewInt(8), nil), big.NewInt(-1)),
			expectsError: true,
		},
		{
			name:         "base -1 * 2^64",
			baseInteger:  new(big.Int).Mul(new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil), big.NewInt(-1)),
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
			name:         "base 2^256 - 1",
			baseInteger:  new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)),
			expectsError: false,
		},
		{
			name:         "base 2^256",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1)),
			expectsError: true,
		},
		{
			name:         "base 2^512",
			baseInteger:  new(big.Int).Add(new(big.Int).Exp(big.NewInt(2), big.NewInt(512), nil), big.NewInt(1)),
			expectsError: true,
		},
	}

	testSetUint256(t, testCases)
}

func testSetUint256(t *testing.T, testCases []Uint256TestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := NewUint256().Set(tc.baseInteger)
			if tc.expectsError && err == nil {
				t.Fatalf("expected error when creating new Uint256, but got none")
			}
			if !tc.expectsError {
				if err != nil {
					t.Fatalf("got unexpected error when creating new Uint256: %v", err)
				}
				if result.Value().Cmp(tc.baseInteger) != 0 {
					t.Fatalf("expected value of %v, got %v instead", tc.baseInteger, result.value)
				}
			}
		})
	}
}

// TestCopyUint256 tests the duplication of an existing Uint256 variable
func TestCopyUint256(t *testing.T) {
	r, err := randomUint256()
	if err != nil {
		t.Fatal(err)
	}

	c := NewUint256().Copy(r)

	if !c.Equals(r) {
		t.Fatalf("copy of uint256 %v has an unequal value of %v", r, c)
	}

	if c == r {
		t.Fatalf("copy of uint256 %v shares memory with its base", r)
	}
}

func randomUint256() (*Uint256, error) {
	r, err := rand.Int(rand.Reader, new(big.Int).Sub(maxUint256, minUint256)) // base for random
	if err != nil {
		return nil, err
	}

	randomUint256 := new(big.Int).Add(r, minUint256) // random is within [minUint256, maxUint256]

	return NewUint256().Set(randomUint256)
}

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
		t.Fatalf("retrieved uint256 %v has an unequal balance to the original uint256 %v", u, r)
	}
}
