package pot

import (
	"testing"
)

type testAddress []byte

func newTestAddress(a []byte) testAddress {
	if a != nil {
		return testAddress(a)
	}
	return testAddress(make([]byte, 32))
}

func (a testAddress) Address() []byte {
	return []byte(a)
}

// TestDistance tests the correctness of the distance calculation
// as well as the comparison method
func TestDistance(t *testing.T) {
	a := [32]byte{}
	b := [32]byte{}
	a[0] = 0x91
	b[0] = 0x82
	aa := newTestAddress(a[:]) //NewAddressFromBytes(a[:])
	ab := newTestAddress(b[:]) //NewAddressFromBytes(b[:])
	distance, err := Distance(aa, ab)
	if err != nil {
		t.Fatal(err)
	}
	correctDistance := "8593944123082061379093159043613555660984881674403010612303492563087302590464"
	if distance.String() != correctDistance {
		t.Fatalf("Distance calculation mismatch, got %s, expected %s", distance.String(), correctDistance)
	}

	c := [32]byte{}
	c[0] = 0x12
	ac := newTestAddress(c[:])
	_, err = DistanceCmp(aa, ab, ac[:31])
	if err == nil {
		t.Fatal("Expected length mismatch on address to fail")
	}

	cmp, err := DistanceCmp(aa, ab, ac)
	if err != nil {
		t.Fatal(err)
	} else if cmp != -1 {
		t.Fatalf("aaab < aaac, expected -1, got %d", cmp)
	}

	cmp, err = DistanceCmp(aa, ac, ab)
	if err != nil {
		t.Fatal(err)
	} else if cmp != 1 {
		t.Fatalf("aaab > aaac, expected 1, got %d", cmp)
	}

	cmp, err = DistanceCmp(aa, ab, ab)
	if err != nil {
		t.Fatal(err)
	} else if cmp != 0 {
		t.Fatalf("aaab == aaab, expected 0, got %d", cmp)
	}
}
