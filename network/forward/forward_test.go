package forward

import (
	"bytes"
	"testing"

	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

func TestNew(t *testing.T) {

	addr := make([]byte, 32)
	addr[31] = 0x01
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr, kadParams)

	sessionId = 42
	fwdBase := New(kad, "", nil)
	if !bytes.Equal(fwdBase.pivot, addr) {
		t.Fatalf("pivot base; expected %x, got %x", addr, fwdBase.pivot)
	}
	if fwdBase.id != 42 {
		t.Fatalf("sessionId; expected %d, got %d", 42, fwdBase.id)
	}

	bytesNear := pot.NewAddressFromString("00000001")
	capabilityIndex := "foo"
	fwdExplicit := New(kad, capabilityIndex, bytesNear)
	if !bytes.Equal(fwdExplicit.pivot, bytesNear) {
		t.Fatalf("pivot explicit; expected %x, got %x", bytesNear, fwdExplicit.pivot)
	}
	if fwdExplicit.id != 43 {
		t.Fatalf("sessionId; expected %d, got %d", 43, fwdExplicit.id)
	}
	if fwdExplicit.capabilityIndex != capabilityIndex {
		t.Fatalf("capabilityindex, expected %s, got %s", capabilityIndex, fwdExplicit.capabilityIndex)
	}
}

//func TestGet() {
//addr := make([]byte, 32)
//	kadParams := network.NewKadParams()
//	kad := network.NewKademlia(addr, kadParams)
//	cp := capability.NewCapability(4, 2)
//	kad.RegisterCapabilityIndex("foo", *cp)
//
//	bytesFar := pot.NewAddressFromString("10000000")
//	bytesNear := pot.NewAddressFromString("00000001")
//	addrFar := network.NewBzzAddr(bytesFar, []byte{})
//	addrNear := network.NewBzzAddr(bytesNear, []byte{})
//	addrFar.Capabilities.Add(cp)
//	addrNear.Capabilities.Add(cp)
//	peerFar := network.NewPeer(&network.BzzPeer{BzzAddr: addrFar}, kad)
//	peerNear := network.NewPeer(&network.BzzPeer{BzzAddr: addrNear}, kad)
//	kad.Register(addrFar)
//	kad.Register(addrNear)
//	kad.On(peerFar)
//	kad.On(peerNear)
//
//
//resultNear, err := fwdBase.Get(1)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if len(resultNear) != 1 {
//		t.Fatalf("peer missing, expected %d, got %d", 1, len(resultNear))
//	}
//	if !bytes.Equal(resultNear[0].Address(), addrNear.Address()) {
//		t.Fatalf("peer mismatch, expected %x, got %x", addrNear.Address(), resultNear[0].Address())
//	}

//}
