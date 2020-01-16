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

	n := 100
	data := getRandomQuantiles(n, 123)
	sort.Sort(data)

	for i := 1; i < n; i++ {
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
				tt.Errorf("expected: %f, got: %f", test.r, test.f.Decimal())
			}
		})
	}
}

func TestClosest(t *testing.T) {
	t.Parallel()

	emptyQ := make(quantiles, 0)
	if emptyQ.Closest(fraction{1, 2}) != nil {
		t.Error("expected nil")
	}

	data := make(quantiles, 5)
	data[0].numerator = 1
	data[0].denominator = 2
	data[1].numerator = 1
	data[1].denominator = 3
	data[2].numerator = 1
	data[2].denominator = 4
	data[3].numerator = 2
	data[3].denominator = 3
	data[4].numerator = 3
	data[4].denominator = 4

	sanityChecks := []fraction{
		{0, 1}, {1, 1}, {1, 2}, {2, 2}, {2, 1},
	}
	for _, check := range sanityChecks {
		if data.Closest(check) == nil {
			t.Error("expected any value, got nil")
		}
	}

	checks := []struct {
		f        fraction
		expected *quantile
	}{
		{fraction{1, 2}, &data[0]},    // exact fraction
		{fraction{4, 8}, &data[0]},    // almost same as above
		{fraction{1, 3}, &data[1]},    // exact fraction
		{fraction{3, 9}, &data[1]},    // almost same as above
		{fraction{1, 4}, &data[2]},    // exact fraction
		{fraction{1, 5}, &data[2]},    // smaller than any quantile
		{fraction{2, 3}, &data[3]},    // exact fraction
		{fraction{3, 4}, &data[4]},    // exact fraction
		{fraction{4, 4}, &data[4]},    // greater than any quantile
		{fraction{2, 1}, &data[4]},    // greater than any quantile
		{fraction{4, 10}, &data[1]},   // 0.4 closest to 1/3 (0.33)
		{fraction{42, 100}, &data[0]}, // 0.42 closest to 1/2 (0.5)
		{fraction{7, 10}, &data[3]},   // 0.7 closest to 2/3 (0.66)
	}

	for _, check := range checks {
		if c := data.Closest(check.f); c != check.expected {
			t.Errorf("invalid quantile: expected fraction: %d/%d, got: %d/%d",
				check.expected.numerator, check.expected.denominator, c.numerator, c.denominator)
		}
	}
}

func getRandomQuantiles(n int, seed int64) quantiles {
	m := uint64(100 * n)
	data := make(quantiles, n)

	rnd := rand.New(rand.NewSource(seed)) // fixed seed for repeatable tests

	for i := 0; i < n; i++ {
		data[i].numerator = rnd.Uint64() % m
		data[i].denominator = rnd.Uint64() % m
	}
	return data
}
