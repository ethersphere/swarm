package network

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	changes = [][]byte{
		{0x01, 0x02},
		{0x82, 0x04},
		{0x01, 0x02},
	}
	expects = [][]byte{
		{0x01, 0x02},
		{0x83, 0x06},
		{0x82, 0x04},
	}
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

	// Initialize capability
	caps := NewCapabilities(nil)

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

}

func TestCapabilitiesNotification(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities(nil)

	rpcSrv := rpc.NewServer()
	rpcSrv.RegisterName("cap", caps)
	rpcClient := rpc.DialInProc(rpcSrv)

	changeC := make(chan capability)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := rpcClient.Subscribe(ctx, "cap", changeC, "subscribeChange")
	if err != nil {
		t.Fatalf("Capabilities change subscription fail: %v", err)
	}

	errC := make(chan error)
	go func() {
		i := 0
		for {
			select {
			case c, ok := <-changeC:
				if !ok {
					log.Error("closed")
					close(errC)
					return
				}
				if !bytes.Equal(c[2:], expects[i]) {
					errC <- fmt.Errorf("subscribe return fail, got: %v, expected %v", c[2:], expects[i])
				}
			}
			i = i + 1
		}
	}()

	// register capability
	err = rpcClient.Call(nil, "cap_registerCapabilityModule", 1, 2)
	if err != nil {
		t.Fatalf("RegisterCapabilityModule fail: %v", err)
	}

	// Correct flag byte and capability id should succeed
	err = caps.SetCapability(1, changes[0])
	if err != nil {
		t.Fatalf("SetCapability (1) fail: %v", err)
	}

	// Consecutive setcapability should only set specified bytes, leave others alone
	err = caps.SetCapability(1, changes[1])
	if err != nil {
		t.Fatalf("SetCapability (2) fail: %v", err)
	}

	// Removecapability should only remove specified bytes, leave others alone
	err = caps.RemoveCapability(1, changes[2])
	if err != nil {
		t.Fatalf("RemoveCapability fail: %v", err)
	}

	sub.Unsubscribe()
	close(changeC)

	err, ok := <-errC
	if ok {
		t.Fatal(err)
	}
}
