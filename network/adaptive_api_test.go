package network

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

func TestCapabilitiesAPINotifications(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities(nil)

	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("cap", NewCapabilitiesAPI(caps))

	changeRemoteC := make(chan capability)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := rpcClient.Subscribe(ctx, "cap", changeRemoteC, "subscribeChanges")
	if err != nil {
		t.Fatalf("Capabilities change subscription fail: %v", err)
	}

	errRemoteC := make(chan error)
	go func() {
		i := 0
		for {
			select {
			case c, ok := <-changeRemoteC:
				if !ok {
					close(errRemoteC)
					return
				}
				if !bytes.Equal(c[2:], expects[i]) {
					errRemoteC <- fmt.Errorf("subscribe remote return fail, got: %v, expected %v", c[2:], expects[i])
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
	err = rpcClient.Call(nil, "cap_setCapability", 1, changes[0])
	if err != nil {
		t.Fatalf("SetCapability (1) fail: %v", err)
	}

	// Consecutive setcapability should only set specified bytes, leave others alone
	err = rpcClient.Call(nil, "cap_setCapability", 1, changes[1])
	if err != nil {
		t.Fatalf("SetCapability (2) fail: %v", err)
	}

	// Removecapability should only remove specified bytes, leave others alone
	err = rpcClient.Call(nil, "cap_removeCapability", 1, changes[2])
	if err != nil {
		t.Fatalf("RemoveCapability fail: %v", err)
	}

	sub.Unsubscribe()
	close(changeRemoteC)
	err, ok := <-errRemoteC
	if ok {
		t.Fatal(err)
	}
}
