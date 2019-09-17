package network

import (
	"bytes"
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
	if !addr.Match(&addrRecovered) {
		t.Fatalf("bzzaddr mismatch, expected %v, got %v", addr, addrRecovered)
	}
}

// Match returns true if the passed BzzAddr is identical to the receiver
func (b *BzzAddr) Match(bcmp *BzzAddr) bool {
	if !bytes.Equal(b.OAddr, bcmp.OAddr) {
		return false
	}
	if !bytes.Equal(b.UAddr, bcmp.UAddr) {
		return false
	}
	if !b.Capabilities.Match(bcmp.Capabilities) {
		return false
	}
	return true
}
