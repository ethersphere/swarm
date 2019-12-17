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

	mgr := NewSessionManager(kad)
	fwdBase := mgr.New("", nil)
	if !bytes.Equal(fwdBase.pivot, addr) {
		t.Fatalf("pivot base; expected %x, got %x", addr, fwdBase.pivot)
	}
	if fwdBase.id != 0 {
		t.Fatalf("sessionId; expected %d, got %d", 42, fwdBase.id)
	}

	bytesNear := pot.NewAddressFromString("00000001")
	capabilityIndex := "foo"
	fwdExplicit := mgr.New(capabilityIndex, bytesNear)
	if !bytes.Equal(fwdExplicit.pivot, bytesNear) {
		t.Fatalf("pivot explicit; expected %x, got %x", bytesNear, fwdExplicit.pivot)
	}
	if fwdExplicit.id != 1 {
		t.Fatalf("sessionId; expected %d, got %d", 43, fwdExplicit.id)
	}
	if fwdExplicit.capabilityIndex != capabilityIndex {
		t.Fatalf("capabilityindex, expected %s, got %s", capabilityIndex, fwdExplicit.capabilityIndex)
	}
	if len(mgr.sessions) != 2 {
		t.Fatalf("sessions array; expected %d, got %d", 2, len(mgr.sessions))
	}
}

func TestManagerContext(t *testing.T) {
	addr := make([]byte, 32)
	addr[31] = 0x01
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr, kadParams)

	mgr := NewSessionManager(kad)
	_ = mgr.New("", nil)       // id 1
	fwdOne := mgr.New("", nil) // id 2
	if len(mgr.sessions) != 2 {
		t.Fatalf("mgr session length; expected 2, got %d", len(mgr.sessions))
	}
	if mgr.sessions[2] != fwdOne {
		t.Fatalf("fromcontext; expected %p, got %p", fwdOne, mgr.sessions[2])
	}

	newAddr := make([]byte, 32)
	newAddr[31] = 0x02
	fwdTwo := mgr.New("foo", newAddr) // id 3
	sctx, err := mgr.ToContext(3)
	if err != nil {
		t.Fatal(err)
	}
	if fwdTwo.id != sctx.SessionId {
		t.Fatalf("to context id; expected %d, got %d", fwdTwo.id, sctx.SessionId)
	}
	if fwdTwo.capabilityIndex != sctx.CapabilityIndex {
		t.Fatalf("to context id; expected %s, got %s", fwdTwo.capabilityIndex, sctx.CapabilityIndex)
	}
	if !bytes.Equal(fwdTwo.pivot, sctx.Address) {
		t.Fatalf("to context id; expected %x, got %x", fwdTwo.pivot, sctx.Address)
	}

	sctx = NewSessionContext("bar", newAddr)
	fwdThree, err := mgr.FromContext(sctx)
	if err != nil {
		t.Fatal(err)
	}
	if fwdThree.id != 3 {
		t.Fatalf("from new context id; expected %d, got %d", 3, fwdThree.id)
	}
	if fwdThree.capabilityIndex != sctx.CapabilityIndex {
		t.Fatalf("to context id; expected %s, got %s", fwdThree.capabilityIndex, sctx.CapabilityIndex)
	}
	if !bytes.Equal(fwdThree.pivot, sctx.Address) {
		t.Fatalf("to context id; expected %x, got %x", fwdThree.pivot, sctx.Address)
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
