// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

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

// hash and wrap as mulithash
func TestNewMultihash(t *testing.T) {
	mh, err := NewMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, mh) {
		t.Fatalf("expected hash %x, got %x", expected, mh)
	}
}

// parse multihash, and check that invalid multihashes fail
func TestCheckMultihash(t *testing.T) {
	h := GetHash()
	l, _ := GetLength(expected)
	if l != h.Size() {
		t.Fatalf("expected length %d, got %d", h.Size(), l)
	}
	if _, err := GetLength(expected[1:]); err == nil {
		t.Fatalf("expected failure on corrupt header")
	}
	if _, err := GetLength(expected[:len(expected)-2]); err == nil {
		t.Fatalf("expected failure on short content")
	}
	dh, _ := FromMultihash(expected)
	if !bytes.Equal(dh, dataHash) {
		t.Fatalf("expected content hash %x, got %x", dataHash, dh)
	}
}
