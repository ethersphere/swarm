package storage

import (
	"bytes"
	"testing"
)

// tests the hashing of an empty slice; not even a valid chunk
func TestEmpty(t *testing.T) {
	// obtained from
	// true | sha256sum
	expected := []byte{
		0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14, 0x9a, 0xfb, 0xf4, 0xc8,
		0x99, 0x6f, 0xb9, 0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
		0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55,
	}
	empty := []byte{}
	hasher := MakeHashFunc(NewChunkerParams().Hash)
	hash := BinaryMerkle(empty, hasher)
	t.Logf("Empty slice hash: %x", hash)
	if !bytes.Equal(hash, expected) {
		t.Errorf("Expected: %x", expected)
	}
}

// tests the hashing of an empty chunk
func TestEmptyChunk(t *testing.T) {
	// obtained from
	// head -c8 /dev/zero |sha256sum
	expected := []byte{
		0xaf, 0x55, 0x70, 0xf5, 0xa1, 0x81, 0x0b, 0x7a, 0xf7, 0x8c, 0xaf, 0x4b,
		0xc7, 0x0a, 0x66, 0x0f, 0x0d, 0xf5, 0x1e, 0x42, 0xba, 0xf9, 0x1d, 0x4d,
		0xe5, 0xb2, 0x32, 0x8d, 0xe0, 0xe8, 0x3d, 0xfc,
	}
	empty := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	hasher := MakeHashFunc(NewChunkerParams().Hash)
	hash := BinaryMerkle(empty, hasher)
	t.Logf("Empty chunk hash: %x", hash)
	if !bytes.Equal(hash, expected) {
		t.Errorf("Expected: %x", expected)
	}
}

// tests the hashing of a chunk containing 96 (=32+64) zeroes
func Test96ZeroesChunk(t *testing.T) {
	// obtained from
	// ((echo -en '\0140' ; head -c39 /dev/zero )|
	// openssl sha256 -binary ; head -c64 /dev/zero |
	// openssl sha256 -binary)|sha256sum
	expected := []byte{
		0xeb, 0xfb, 0x09, 0xe6, 0xcf, 0xc9, 0x90, 0x46, 0x6b, 0x23, 0xeb, 0x54,
		0xd4, 0x86, 0xf6, 0xfb, 0x23, 0xbd, 0x5d, 0x12, 0x71, 0x6e, 0xe7, 0x3b,
		0xb5, 0x2c, 0xec, 0x91, 0x39, 0x58, 0x6e, 0x1e,
	}
	empty := []byte{
		0x60, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	hasher := MakeHashFunc(NewChunkerParams().Hash)
	hash := BinaryMerkle(empty, hasher)
	t.Logf("96 zeroes chunk hash: %x", hash)
	if !bytes.Equal(hash, expected) {
		t.Errorf("Expected: %x", expected)
	}
}
