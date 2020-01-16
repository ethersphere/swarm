package localstore

import (
	"github.com/ethersphere/swarm/shed"
	"math"
	"sort"
)

type fraction struct {
	numerator   uint64
	denominator uint64
}

func (f fraction) Decimal() float64 {
	return float64(f.numerator) / float64(f.denominator)
}

type quantile struct {
	fraction
	Item     shed.Item
	Position uint64
}

type quantiles []quantile

func (q quantiles) Len() int {
	return len(q)
}

func (q quantiles) Less(i, j int) bool {
	return q[i].fraction.Decimal() < q[j].fraction.Decimal()
}

func (q quantiles) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q quantiles) Get(f fraction) (item shed.Item, position uint64, found bool) {
	for _, x := range q {
		if x.fraction == f {
			return x.Item, x.Position, true
		}
	}
	return item, 0, false
}

func (q quantiles) Closest(f fraction) (closest *quantile) {
	for _, x := range q {
		if x.fraction == f {
			return &x
		}
		if closest == nil || math.Abs(x.Decimal()-f.Decimal()) < math.Abs(closest.Decimal()-f.Decimal()) {
			closest = &x
		}
	}
	return closest
}

func (q quantiles) Set(f fraction, item shed.Item, position uint64) {
	for i := range q {
		if q[i].fraction == f {
			q[i].Item = item
			q[i].Position = position
			return
		}
	}
	q = append(q, quantile{
		fraction: f,
		Item:     item,
	})
	sort.Sort(q)
}

func quantilePosition(total, numerator, denominator uint64) uint64 {
	return total / denominator * numerator
}

// based on https://hackmd.io/t-OQFK3mTsGfrpLCqDrdlw#Synced-chunks
// TODO: review and document exact quantiles for chunks
func chunkQuantileFraction(po, responsibilityRadius int) fraction {
	if po < responsibilityRadius {
		// More Distant Chunks
		n := uint64(responsibilityRadius - po)
		return fraction{numerator: n, denominator: n + 1}
	}
	// Most Proximate Chunks
	return fraction{numerator: 1, denominator: 3}
}
