package hash

import (
	"bytes"
	"testing"
)

var (
	expected = make([]byte, len(dataHash)+2)
)

func init() {
	expected[0] = 0x1b
	expected[1] = 0x20
	copy(expected[2:], dataHash)

}

func TestNewMultihash(t *testing.T) {
	AddHasher("BMT", 1, makeHashFunc)
	mh, err := NewMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, mh) {
		t.Fatalf("expected hash %x, got %x", expected, mh)
	}
	t.Logf("%x", mh)
}

func TestCheckMultihash(t *testing.T) {
	h := GetHasher("bar")
	l, _ := GetLength(expected)
	if l != h.Size() {
		t.Fatalf("expected length %d, got %d", h.Size(), l)
	}
	dh, _ := GetHash(expected)
	if !bytes.Equal(dh, dataHash) {
		t.Fatalf("expected content hash %x, got %x", dataHash, dh)
	}
}
