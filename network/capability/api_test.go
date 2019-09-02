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
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

// TestAPI tests that API calls stores and reports correctly
func TestAPI(t *testing.T) {

	// Initialize capability
	caps := NewCapabilities()

	// Set up faux rpc
	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("cap", NewAPI(caps))

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
