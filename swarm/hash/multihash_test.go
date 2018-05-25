package swarmhash

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
	mh, err := NewMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, mh) {
		t.Fatalf("expected hash %x, got %x", expected, mh)
	}
}

func TestCheckMultihash(t *testing.T) {
	h := GetHash()
	l, _ := GetLength(expected)
	if l != h.Size() {
		t.Fatalf("expected length %d, got %d", h.Size(), l)
	}
	dh, _ := FromMultihash(expected)
	if !bytes.Equal(dh, dataHash) {
		t.Fatalf("expected content hash %x, got %x", dataHash, dh)
	}
}
