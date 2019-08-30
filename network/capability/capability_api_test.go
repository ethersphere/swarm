package capability

import (
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

// TestCapabilitiesAPI tests that API calls stores and reports correctly
func TestCapabilitiesAPI(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities()

	// Set up faux rpc
	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("cap", NewCapabilitiesAPI(caps))

	// create the capability and register it
	c1 := NewCapability(42, 13)
	c1.Set(9)
	err := rpcClient.Call(nil, "cap_registerCapability", c1)
	if err != nil {
		t.Fatalf("Register fail: %v", err)
	}

	// check that the capability is registered
	c1.Set(9)
	err = rpcClient.Call(nil, "cap_isRegisteredCapability", c1.Id)
	if err != nil {
		t.Fatalf("Register fail: %v", err)
	}

	// check that isRegistered doesn't give false positives
	c2 := CapabilityID(13)
	err = rpcClient.Call(nil, "cap_isRegisteredCapability", c2)
	if err != nil {
		t.Fatalf("Register fail: %v", err)
	}

	// check that correct values have been stored
	var r bool
	err = rpcClient.Call(&r, "cap_matchCapability", c1.Id, 9)
	if err != nil {
		t.Fatalf("isSet fail: %v", err)
	} else if !r {
		t.Fatalf("isSet should be false, got %v", r)
	}

	err = rpcClient.Call(&r, "cap_matchCapability", c1.Id, 1)
	if err != nil {
		t.Fatalf("isSet fail: %v", err)
	} else if r {
		t.Fatalf("isSet should be true, got %v", r)
	}
}
