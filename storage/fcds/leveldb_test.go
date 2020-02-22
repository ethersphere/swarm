package fcds

import (
	"fmt"
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
	v := ms.NextShard()
	fmt.Println(v)
	if v != 3 {
		t.Fatal(v)
	}

	ms.free[2]++
	ms.free[2]++
	ms.free[2]++
	v = ms.NextShard()
	fmt.Println(v)
	if v != 2 {
		t.Fatal(v)
	}
	ms.free[3]++
	v = ms.NextShard()
	fmt.Println(v)
	if v != 3 {
		t.Fatal(v)
	}
}
