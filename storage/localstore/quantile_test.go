package localstore

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

func TestSorting(t *testing.T) {
	t.Parallel()
	n := uint64(10)
	m := 10 * n
	data := make(quantiles, n)

	rnd := rand.New(rand.NewSource(123)) // fixed seed for repeatable tests

	for i := uint64(0); i < n; i++ {
		data[i].numerator = rnd.Uint64() % m
		data[i].denominator = rnd.Uint64() % m
	}

	sort.Sort(data)

	for i := uint64(1); i < n; i++ {
		// compare without trusting any methods - use common denominator
		x, y := data[i], data[i-1]
		if x.numerator*y.denominator < y.numerator*x.denominator {
			t.Error("quantiles not ordered correctly")
		}
	}
}

func TestFraction(t *testing.T) {
	t.Parallel()

	tolerance := float64(0.0000001)

	tests := []struct {
		f fraction
		r float64
	}{
		{fraction{1, 2}, 0.5},
		{fraction{2, 4}, 0.5},
		{fraction{3, 4}, 0.75},
		{fraction{4, 5}, 0.8},
		{fraction{1, 5}, 0.2},
		{fraction{99, 100}, 0.99},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Fraction %d/%d", test.f.numerator, test.f.denominator), func(tt *testing.T) {
			if math.Abs(test.f.Decimal()-test.r) > tolerance {
				tt.Errorf("expected: %f, received: %f", test.r, test.f.Decimal())
			}
		})
	}
}
