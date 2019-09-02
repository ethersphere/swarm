// Copyright 2019 The Swarm authors
// This file is part of the swarm library.
//
// The swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the swarm library. If not, see <http://www.gnu.org/licenses/>.

package capability

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
)

// TestCapabilitySetUnset tests that setting and unsetting bits yield expected results
func TestCapabilitySetUnset(t *testing.T) {
	firstSet := []bool{
		true, false, false, false, false, false, true, true, false,
	} // 1000 0011 0
	firstResult := firstSet
	secondSet := []bool{
		false, true, false, true, false, false, true, false, true,
	} // 0101 0010 1
	secondResult := []bool{
		true, true, false, true, false, false, true, true, true,
	} // 1101 0011 1
	thirdUnset := []bool{
		true, false, true, true, false, false, true, false, true,
	} // 1011 0010 1
	thirdResult := []bool{
		false, true, false, false, false, false, false, true, false,
	} // 0100 0001 0

	c := NewCapability(42, 9)
	for i, b := range firstSet {
		if b {
			c.Set(i)
		}
	}
	if !isSameBools(c.Cap, firstResult) {
		t.Fatalf("first set result mismatch, expected %v, got %v", firstResult, c.Cap)
	}

	for i, b := range secondSet {
		if b {
			c.Set(i)
		}
	}
	if !isSameBools(c.Cap, secondResult) {
		t.Fatalf("second set result mismatch, expected %v, got %v", secondResult, c.Cap)
	}

	for i, b := range thirdUnset {
		if b {
			c.Unset(i)
		}
	}
	if !isSameBools(c.Cap, thirdResult) {
		t.Fatalf("second set result mismatch, expected %v, got %v", thirdResult, c.Cap)
	}
}

// TestCapabilitiesControl tests that the methods for manipulating the capabilities bitvectors set values correctly and return errors when they should
func TestCapabilitiesControl(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities()

	// Register module. Should succeed
	c1 := NewCapability(1, 16)
	err := caps.Add(c1)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule fail: %v", err)
	}

	// Fail if capability id already exists
	c2 := NewCapability(1, 1)
	err = caps.Add(c2)
	if err == nil {
		t.Fatalf("Expected RegisterCapabilityModule call with existing id to fail")
	}

	// More than one capabilities flag vector should be possible
	c3 := NewCapability(2, 1)
	err = caps.Add(c3)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule (second) fail: %v", err)
	}
}

// TestCapabilitiesString checks that the string representation of the capabilities is correct
func TestCapabilitiesString(t *testing.T) {
	sets1 := []bool{
		false, false, true,
	}
	c1 := NewCapability(42, len(sets1))
	for i, b := range sets1 {
		if b {
			c1.Set(i)
		}
	}
	sets2 := []bool{
		true, false, false, false, true, false, true, false, true,
	}
	c2 := NewCapability(666, len(sets2))
	for i, b := range sets2 {
		if b {
			c2.Set(i)
		}
	}

	caps := NewCapabilities()
	caps.Add(c1)
	caps.Add(c2)

	correctString := "42:001,666:100010101"
	if correctString != caps.String() {
		t.Fatalf("Capabilities string mismatch; expected %s, got %s", correctString, caps)
	}
}

// TestCapabilitiesRLP ensures that a round of serialization and deserialization of Capabilities object
// results in the correct data
func TestCapabilitiesRLP(t *testing.T) {
	c := NewCapabilities()
	cap1 := &Capability{
		Id:  42,
		Cap: []bool{true, false, true},
	}
	c.Add(cap1)
	cap2 := &Capability{
		Id:  666,
		Cap: []bool{true, false, true, false, true, true, false, false, true},
	}
	c.Add(cap2)
	buf := bytes.NewBuffer(nil)
	err := rlp.Encode(buf, &c)
	if err != nil {
		t.Fatal(err)
	}

	cRestored := NewCapabilities()
	err = rlp.Decode(buf, &cRestored)
	if err != nil {
		t.Fatal(err)
	}

	cap1Restored := cRestored.Get(cap1.Id)
	if cap1Restored.Id != cap1.Id {
		t.Fatalf("cap 1 id not correct, expected %d, got %d", cap1.Id, cap1Restored.Id)
	}
	if !cap1.IsSameAs(cap1Restored) {
		t.Fatalf("cap 1 caps not correct, expected %v, got %v", cap1.Cap, cap1Restored.Cap)
	}

	cap2Restored := cRestored.Get(cap2.Id)
	if cap2Restored.Id != cap2.Id {
		t.Fatalf("cap 1 id not correct, expected %d, got %d", cap2.Id, cap2Restored.Id)
	}
	if !cap2.IsSameAs(cap2Restored) {
		t.Fatalf("cap 1 caps not correct, expected %v, got %v", cap2.Cap, cap2Restored.Cap)
	}
}
