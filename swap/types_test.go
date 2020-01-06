package swap

import (
	"math/big"
	"testing"
)

type Uint256TestCase struct {
	name         string
	baseInteger  *big.Int
	expectsError bool
}

func TestNewUint256(t *testing.T) {
	testCases := []Uint256TestCase{
		{
			name:         "base 0",
			baseInteger:  big.NewInt(0),
			expectsError: false,
		},
		{
			name:         "base -1",
			baseInteger:  big.NewInt(-1),
			expectsError: true,
		},
		{
			name:         "base -256",
			baseInteger:  big.NewInt(-256),
			expectsError: true,
		},
		{
			name:         "base -2,147,483,648",
			baseInteger:  big.NewInt(-2147483648),
			expectsError: true,
		},
	}

	testNewUint256(t, testCases)
}

func testNewUint256(t *testing.T, testCases []Uint256TestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUint256().Set(tc.baseInteger)
			if tc.expectsError && err == nil {
				t.Fatalf("expected error when creating new Uint256, but got none")
			}
			if !tc.expectsError && err != nil {
				t.Fatalf("got unexpected error when creating new Uint256: %v", err)
			}
		})
	}
}
