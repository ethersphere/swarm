package network

import (
	"bytes"
	"fmt"
	"testing"
)

// TestCapabilitiesString tests the correctness
func TestCapabilitiesString(t *testing.T) {
	caps := Capabilities{}

	// set up capabilities with arbitary content
	cOne := newCapability(0, 2)
	controlFlags := []byte{0x08, 0x2a}
	cOne.set(controlFlags)
	caps.add(cOne)

	cTwo := newCapability(127, 3)
	controlFlags = []byte{0x00, 0x02, 0x6e}
	cTwo.set(controlFlags)
	caps.add(cTwo)

	controlString := "00:0000100000101010,7f:000000000000001001101110"

	capstring := fmt.Sprintf("%v", caps)
	if capstring != controlString {
		t.Fatalf("capabilities string mismatch, expected: '%s', got '%s'", controlString, capstring)
	}
}

// TestCapabilitiesAPI tests that the API alters the capabilities as they should, and throws errors when it should
func TestCapabilitiesAPI(t *testing.T) {

	changes := [][]byte{
		{0x01, 0x02},
		{0x82, 0x04},
		{0x01, 0x02},
	}
	expects := [][]byte{
		{0x01, 0x02},
		{0x83, 0x06},
		{0x82, 0x04},
	}

	// Initialize capability
	// Set explicitly with builtin bzz value
	caps := NewCapabilities()
	id, changeC := caps.subscribe()

	errC := make(chan error, len(expects))
	go func() {
		i := 0
		for {
			select {
			case f, ok := <-changeC:
				if !ok {
					close(errC)
					return
				}
				if !bytes.Equal(expects[i], f[2:]) {
					errC <- fmt.Errorf("notify (%d) failed: got %v, expect %v", i, f, expects[i])
				}
				i++
			}
		}
	}()

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
	err = caps.SetCapability(1, changes[0])
	if err != nil {
		t.Fatalf("SetCapability (1) fail: %v", err)
	}

	// check set correctly
	if !bytes.Equal(caps.Flags[0][2:], expects[0]) {
		t.Fatalf("Expected capability flags after first SetCapability %v, got: %v", expects[0], caps.Flags[0][2:])
	}

	// Consecutive setcapability should only set specified bytes, leave others alone
	err = caps.SetCapability(1, changes[1])
	if err != nil {
		t.Fatalf("SetCapability (2) fail: %v", err)
	}
	if !bytes.Equal(caps.Flags[0][2:], expects[1]) {
		t.Fatalf("Expected capability flags after second SetCapability %v, got: %v", expects[1], caps.Flags[0][2:])
	}

	// Removecapability should only remove specified bytes, leave others alone
	err = caps.RemoveCapability(1, changes[2])
	if err != nil {
		t.Fatalf("RemoveCapability fail: %v", err)
	}
	if !bytes.Equal(caps.Flags[0][2:], expects[2]) {
		t.Fatalf("Expected capability flags after second SetCapability %v, got: %v", expects[2], caps.Flags[0][2:])
	}

	err = caps.unsubscribe(id)
	if err != nil {
		t.Fatal(err)
	}
	for {
		err, ok := <-errC
		if !ok {
			break
		}
		t.Fatal(err)
	}
}
