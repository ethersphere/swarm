package lookup_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

func TestMarshallers(t *testing.T) {

	for i := uint64(1); i < lookup.MaxTime; i *= 3 {
		e := lookup.Epoch{
			Time:  i,
			Level: uint8(i % 20),
		}
		b, err := e.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		var e2 lookup.Epoch
		if err := e2.UnmarshalBinary(b); err != nil {
			t.Fatal(err)
		}
		if e != e2 {
			t.Fatal("Expected unmarshalled epoch to be equal to marshalled onet.Fatal(err)")
		}
	}

}

func TestAfter(t *testing.T) {
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

	after := b.After(a)
	if after != true {
		t.Fatal("Expected 'after' to be true, got false")
	}
	after = b.After(b)
	if after != false {
		t.Fatal("Expected 'after' to be false when both epochs are identical, got true")
	}
	after = b.After(c)
	if after != true {
		t.Fatal("Expected 'after' to be true when both epochs have the same time but the level is lower in the first one, but got false")
	}

}
