package localstore

import (
	"math"
	"sort"

	"github.com/ethersphere/swarm/shed"
)

type Fraction struct {
	Numerator   uint64
	Denominator uint64
}

func (f Fraction) Decimal() float64 {
	return float64(f.Numerator) / float64(f.Denominator)
}

type quantile struct {
	Fraction
	Item     shed.Item
	Position uint64
}

type quantiles []quantile

func (q quantiles) Len() int {
	return len(q)
}

func (q quantiles) Less(i, j int) bool {
	// TODO(tzdybal) - is it reasonable to use common denominator instead of Decimal() function?
	return q[i].Fraction.Decimal() < q[j].Fraction.Decimal()
}

func (q quantiles) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q quantiles) Get(f Fraction) (item shed.Item, position uint64, found bool) {
	for _, x := range q {
		if x.Fraction == f {
			return x.Item, x.Position, true
		}
	}
	return item, 0, false
}

func (q quantiles) Closest(f Fraction) (closest *quantile) {
	for i, x := range q {
		if x.Fraction == f {
			return &q[i]
		}
		if closest == nil || math.Abs(x.Decimal()-f.Decimal()) < math.Abs(closest.Decimal()-f.Decimal()) {
			closest = &q[i]
		}
	}
	return closest
}

func (q quantiles) Set(f Fraction, item shed.Item, position uint64) quantiles {
	for i := range q {
		if q[i].Fraction == f {
			q[i].Item = item
			q[i].Position = position
			return q
		}
	}
	newQ := append(q, quantile{
		Fraction: f,
		Item:     item,
	})
	sort.Sort(newQ)
	return newQ
}

func quantilePosition(total, numerator, denominator uint64) uint64 {
	return total / denominator * numerator
}

