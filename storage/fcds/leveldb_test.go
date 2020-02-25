package fcds

import (
	"testing"
)

func TestNextShard(t *testing.T) {
	ms, err := NewMetaStore("", true)
	if err != nil {
		t.Fatal(err)
	}
	ms.free[3]++
	ms.free[3]++
	ms.free[3]++
	v, b := ms.NextShard()
	if !b {
		t.Fatal("expected free slots on shard")
	}
	if v != 3 {
		t.Fatal(v)
	}

	ms.free[2]++
	ms.free[2]++
	ms.free[2]++
	v, b = ms.NextShard()
	if !b {
		t.Fatal("expected free slots on shard")
	}
	if v != 2 {
		t.Fatal(v)
	}
	ms.free[3]++
	v, b = ms.NextShard()
	if !b {
		t.Fatal("expected free slots on shard")
	}
	if v != 3 {
		t.Fatal(v)
	}
}
