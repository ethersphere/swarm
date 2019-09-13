package network

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/network/capability"
)

// TestBzzAddrRLPSerialzation verifies reverisibility of RLP serialization of BzzAddr
func TestBzzAddrRLPSerialization(t *testing.T) {
	caps := capability.NewCapabilities()
	caps.Add(lightCapability)
	addr := RandomBzzAddr().WithCapabilities(caps)
	b, err := rlp.EncodeToBytes(addr)
	if err != nil {
		t.Fatal(err)
	}
	var addrRecovered BzzAddr
	err = rlp.DecodeBytes(b, &addrRecovered)
	if err != nil {
		t.Fatal(err)
	}
}
