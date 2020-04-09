package localstore

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/ethersphere/swarm/shed"
)

func TestSorting(t *testing.T) {
	t.Parallel()

	n := 100
	data := getRandomQuantiles(n, 123)
	sort.Sort(data)

	for i := 1; i < n; i++ {
		// compare without trusting any methods - use common denominator
		x, y := data[i], data[i-1]
		if x.Numerator*y.Denominator < y.Numerator*x.Denominator {
			t.Error("quantiles not ordered correctly")
		}
	}
}

func TestFraction(t *testing.T) {
	t.Parallel()

	tolerance := float64(0.0000001)

	tests := []struct {
		f Fraction
		r float64
	}{
		{Fraction{1, 2}, 0.5},
		{Fraction{2, 4}, 0.5},
		{Fraction{3, 4}, 0.75},
		{Fraction{4, 5}, 0.8},
		{Fraction{1, 5}, 0.2},
		{Fraction{99, 100}, 0.99},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Fraction %d/%d", test.f.Numerator, test.f.Denominator), func(tt *testing.T) {
			if math.Abs(test.f.Decimal()-test.r) > tolerance {
				tt.Errorf("expected: %f, got: %f", test.r, test.f.Decimal())
			}
		})
	}
}

func TestClosest(t *testing.T) {
	t.Parallel()

	emptyQ := make(quantiles, 0)
	if emptyQ.Closest(Fraction{1, 2}) != nil {
		t.Error("expected nil")
	}

	data := make(quantiles, 5)
	data[0].Numerator = 1
	data[0].Denominator = 2
	data[1].Numerator = 1
	data[1].Denominator = 3
	data[2].Numerator = 1
	data[2].Denominator = 4
	data[3].Numerator = 2
	data[3].Denominator = 3
	data[4].Numerator = 3
	data[4].Denominator = 4

	sanityChecks := []Fraction{
		{0, 1}, {1, 1}, {1, 2}, {2, 2}, {2, 1},
	}
	for _, check := range sanityChecks {
		if data.Closest(check) == nil {
			t.Error("expected any value, got nil")
		}
	}

	checks := []struct {
		f        Fraction
		expected *quantile
	}{
		{Fraction{1, 2}, &data[0]},    // exact Fraction
		{Fraction{4, 8}, &data[0]},    // almost same as above
		{Fraction{1, 3}, &data[1]},    // exact Fraction
		{Fraction{3, 9}, &data[1]},    // almost same as above
		{Fraction{1, 4}, &data[2]},    // exact Fraction
		{Fraction{1, 5}, &data[2]},    // smaller than any quantile
		{Fraction{2, 3}, &data[3]},    // exact Fraction
		{Fraction{3, 4}, &data[4]},    // exact Fraction
		{Fraction{4, 4}, &data[4]},    // greater than any quantile
		{Fraction{2, 1}, &data[4]},    // greater than any quantile
		{Fraction{4, 10}, &data[1]},   // 0.4 closest to 1/3 (0.33)
		{Fraction{42, 100}, &data[0]}, // 0.42 closest to 1/2 (0.5)
		{Fraction{7, 10}, &data[3]},   // 0.7 closest to 2/3 (0.66)
	}

	for _, check := range checks {
		if c := data.Closest(check.f); c != check.expected {
			t.Errorf("invalid quantile: expected Fraction: %d/%d, got: %d/%d",
				check.expected.Numerator, check.expected.Denominator, c.Numerator, c.Denominator)
		}
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	var q quantiles

	if len(q) != 0 {
		t.Errorf("Prerequisite is false")
	}

	q = q.Set(Fraction{1, 1}, shed.Item{}, 1)

	if len(q) == 0 {
		t.Errorf("quantiles.Set doesn't work")
	}
}

func getRandomQuantiles(n int, seed int64) quantiles {
	m := uint64(100 * n)
	data := make(quantiles, n)

	rnd := rand.New(rand.NewSource(seed)) // fixed seed for repeatable tests

	for i := 0; i < n; i++ {
		data[i].Numerator = rnd.Uint64() % m
		data[i].Denominator = rnd.Uint64() % m
	}
	return data
}
