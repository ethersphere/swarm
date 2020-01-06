package swap

import "testing"

import "math/big"

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
