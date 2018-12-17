package lookup_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage/feed/lookup"
	"github.com/ethereum/go-ethereum/swarm/testutil"
)

func TestMarshallers(tx *testing.T) {
	t := testutil.BeginTest(tx, false)
	defer t.FinishTest()

	for i := uint64(1); i < lookup.MaxTime; i *= 3 {
		e := lookup.Epoch{
			Time:  i,
			Level: uint8(i % 20),
		}
		b, err := e.MarshalBinary()
		t.Ok(err)
		t.EqualsKey(fmt.Sprintf("epoch%018d", i), hexutil.Bytes(b))

		var e2 lookup.Epoch
		err = e2.UnmarshalBinary(b)
		t.Ok(err)

		t.Assert(e == e2, "Expected unmarshalled epoch to be equal to marshalled one.")
	}

}

func TestAfter(tx *testing.T) {
	t := testutil.BeginTest(tx, false)
	defer t.FinishTest()

	a := lookup.Epoch{
		Time:  5,
		Level: 3,
	}
	b := lookup.Epoch{
		Time:  6,
		Level: 3,
	}
	c := lookup.Epoch{
		Time:  6,
		Level: 4,
	}

	t.Assert(b.After(a), "Expected 'after' to be true, got false")
	t.Assert(!b.After(b), "Expected 'after' to be false when both epochs are identical, got true")
	t.Assert(b.After(c), "Expected 'after' to be true when both epochs have the same time but the level is lower in the first one, but got false")
}
