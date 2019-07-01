package network

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCapabilitiesString(t *testing.T) {
	caps := Capabilities{}

	// set up capabilities with arbitary content
	cOne := NewCapability(0, 2)
	controlFlags := []byte{0x08, 0x2a}
	cOne.Set(controlFlags)
	caps.add(cOne)

	cTwo := NewCapability(127, 3)
	controlFlags = []byte{0x00, 0x02, 0x6e}
	cTwo.Set(controlFlags)
	caps.add(cTwo)

	controlString := "00:0000100000101010,7f:000000000000001001101110"

	capstring := fmt.Sprintf("%v", caps)
	if capstring != controlString {
		t.Fatalf("capabilities string mismatch, expected: '%s', got '%s'", controlString, capstring)
	}
}

// TestCapabilitiesAPI tests that the API alters the capabilities as they should, and throws errors when it should
func TestCapabilitiesAPI(t *testing.T) {

	// Initialize capability
	// Set explicitly with builtin bzz value
	caps := Capabilities{}

	// Register module. Should succeeed
	err := caps.RegisterCapabilityModule(1, 2)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule fail: %v", err)
	}

	// Fail if capability id already exists
	err = caps.RegisterCapabilityModule(1, 1)
	if err == nil {
		t.Fatalf("Expected RegisterCapabilityModule call with existing id to fail")
	}

	// Move than one capabilities flag vector should be possible
	err = caps.RegisterCapabilityModule(2, 1)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule (second) fail: %v", err)
	}

	// Set on non-existing capability should fail
	err = caps.SetCapability(0, []byte{0x12})
	if err == nil {
		t.Fatalf("Expected SetCapability call with non-existing id to fail")
	}

	// Set on non-existing capability should fail
	err = caps.RemoveCapability(0, []byte{0x12})
	if err == nil {
		t.Fatalf("Expected RemoveCapability call with non-existing id to fail")
	}

	// Wrong flag byte length should fail
	err = caps.SetCapability(1, []byte{0x12, 0x34, 0x56})
	if err == nil {
		t.Fatalf("Expected SetCapability call with wrong length id to fail")
	}

	// Correct flag byte and capability id should succeed
	err = caps.SetCapability(1, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("SetCapability (1) fail: %v", err)
	}

	// check set correctly
	expected := []byte{0x01, 0x02}
	if !bytes.Equal(caps[0][2:], expected) {
		t.Fatalf("Expected capability flags after first SetCapability %v, got: %v", expected, caps[0][2:])
	}

	// Consecutive setcapability should only set specified bytes, leave others alone
	err = caps.SetCapability(1, []byte{0x82, 0x04})
	if err != nil {
		t.Fatalf("SetCapability (2) fail: %v", err)
	}
	expected = []byte{0x83, 0x06}
	if !bytes.Equal(caps[0][2:], expected) {
		t.Fatalf("Expected capability flags after second SetCapability %v, got: %v", expected, caps[0][2:])
	}

	// Removecapability should only remove specified bytes, leave others alone
	err = caps.RemoveCapability(1, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("RemoveCapability fail: %v", err)
	}
	expected = []byte{0x82, 0x04}
	if !bytes.Equal(caps[0][2:], expected) {
		t.Fatalf("Expected capability flags after second SetCapability %v, got: %v", expected, caps[0][2:])
	}

}
