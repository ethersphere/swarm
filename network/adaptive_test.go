package network

import (
	"testing"
)

var (
	objectprefix = []byte{10, 255, 129, 6, 1, 2, 255, 132, 0, 0, 0, 9, 255, 130, 0, 5, 16, 1, 1, 2, 0}

	changes = [][]bool{
		{
			true, false, false, false, false, false, true, true,
			false, false, false, false, false, true, true, false,
		}, // 0x8306
		{
			false, false, false, false, false, false, false, true,
			false, false, false, false, false, false, true, false,
		}, // 0x0102
	}

	expects = [][]bool{
		{
			true, false, false, false, false, false, true, true,
			false, false, false, false, false, true, true, false,
		}, // 0x8306
		{
			true, false, false, false, false, false, true, false,
			false, false, false, false, false, true, false, false,
		}, // 0x8204
	}
)

// TestCapabilitiesControl tests that the methods for manipulating the capabilities bitvectors set values correctly and return errors when they should
func TestCapabilitiesControl(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities()

	// Register module. Should succeed
	c1 := NewCapability(1, 16)
	err := caps.add(c1)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule fail: %v", err)
	}

	// Fail if capability id already exists
	c2 := NewCapability(1, 1)
	err = caps.add(c2)
	if err == nil {
		t.Fatalf("Expected RegisterCapabilityModule call with existing id to fail")
	}

	// More than one capabilities flag vector should be possible
	c3 := NewCapability(2, 1)
	err = caps.add(c3)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule (second) fail: %v", err)
	}

	// Set initial flags
	c4 := caps.get(1)
	for i, b := range changes[0] {
		if b {
			c4.Set(i)
		}
	}
	// verify value
	c4 = caps.get(1)
	for i, b := range c4.cap {
		if b != expects[0][i] {
			t.Fatalf("Expected capability flags after first SetCapability %v, got: %v", expects[0], c4.cap)
		}
	}

	// Consecutive Set should only set specified bytes, leave others alone
	c5 := caps.get(1)
	for i, b := range changes[1] {
		if b {
			c5.Unset(i)
		}
	}
	// verify value
	c5 = caps.get(1)
	for i, b := range c4.cap {
		if b != expects[1][i] {
			t.Fatalf("Expected capability flags after first SetCapability %v, got: %v", expects[0], c4.cap)
		}
	}
}
