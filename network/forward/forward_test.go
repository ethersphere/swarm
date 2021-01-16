package forward

import (
	"bytes"
	"testing"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

// TestNew tests that the SessionManager constructor creates Session object with expected values
func TestNew(t *testing.T) {
	addr := make([]byte, 32)
	addr[31] = 0x01
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr, kadParams)
	kadLB := network.NewKademliaLoadBalancer(kad, false)
	defer kadLB.Stop()

	mgr := NewSessionManager(kadLB)
	fwdBase := mgr.New("", nil)
	defer mgr.Reap(fwdBase.Id())
	if !bytes.Equal(fwdBase.base, addr) {
		t.Fatalf("base base; expected %x, got %x", addr, fwdBase.base)
	}
	if fwdBase.id != 1 {
		t.Fatalf("sessionId; expected %d, got %d", 1, fwdBase.id)
	}

	bytesNear := pot.NewAddressFromString("00000001")
	capabilityIndex := "foo"
	fwdExplicit := mgr.New(capabilityIndex, bytesNear)
	if !bytes.Equal(fwdExplicit.base, bytesNear) {
		t.Fatalf("base explicit; expected %x, got %x", bytesNear, fwdExplicit.base)
	}
	if fwdExplicit.id != 2 {
		t.Fatalf("sessionId; expected %d, got %d", 2, fwdExplicit.id)
	}
	if fwdExplicit.capabilityIndex != capabilityIndex {
		t.Fatalf("capabilityindex, expected %s, got %s", capabilityIndex, fwdExplicit.capabilityIndex)
	}
	if len(mgr.sessions) != 2 {
		t.Fatalf("sessions array; expected %d, got %d", 2, len(mgr.sessions))
	}
}

// TestManagerContext tests that the SessionManager's context translations creates Session objects with expected values, and retrieves existing matching Session objects
func TestManagerContext(t *testing.T) {
	addr := make([]byte, 32)
	addr[31] = 0x01
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr, kadParams)
	kadLB := network.NewKademliaLoadBalancer(kad, false)
	defer kadLB.Stop()

	mgr := NewSessionManager(kadLB)
	fwdVoid := mgr.New("", nil) // id 1
	defer mgr.Reap(fwdVoid.Id())
	fwdOne := mgr.New("", nil) // id 2
	defer mgr.Reap(fwdOne.Id())
	if len(mgr.sessions) != 2 {
		t.Fatalf("mgr session length; expected 2, got %d", len(mgr.sessions))
	}
	if mgr.sessions[2] != fwdOne {
		t.Fatalf("fromcontext; expected %p, got %p", fwdOne, mgr.sessions[2])
	}

	newAddr := make([]byte, 32)
	newAddr[31] = 0x02
	fwdTwo := mgr.New("foo", newAddr) // id 3
	defer mgr.Reap(fwdTwo.Id())
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
	if !bytes.Equal(fwdTwo.base, sctx.Address) {
		t.Fatalf("to context id; expected %x, got %x", fwdTwo.base, sctx.Address)
	}

	sctx = NewSessionContext("", nil)
	sctx.SessionId = 3
	fwdThree, err := mgr.FromContext(sctx)
	if err != nil {
		t.Fatal(err)
	}
	if fwdThree != fwdTwo {
		t.Fatalf("from new context; expected %p, got %p", fwdTwo, fwdThree)
	}
}

// TestGet verifies that the synchronous Get method retrieves peers in the correct order
func TestGet(t *testing.T) {
	bytesOwn := pot.NewAddressFromString("00000000")
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(bytesOwn, kadParams)
	kadLB := network.NewKademliaLoadBalancer(kad, false)
	defer kadLB.Stop()
	cp := capability.NewCapability(4, 2)
	kad.RegisterCapabilityIndex("foo", *cp)

	bytesFar := pot.NewAddressFromString("10000000")
	bytesNear := pot.NewAddressFromString("00000001")
	addrFar := network.NewBzzAddr(bytesFar, []byte{})
	addrNear := network.NewBzzAddr(bytesNear, []byte{})
	addrFar.Capabilities.Add(cp)
	addrNear.Capabilities.Add(cp)
	peerFar := network.NewPeer(&network.BzzPeer{BzzAddr: addrFar}, kad)
	peerNear := network.NewPeer(&network.BzzPeer{BzzAddr: addrNear}, kad)
	kad.Register(addrFar)
	kad.Register(addrNear)
	kad.On(peerFar)
	kad.On(peerNear)

	mgr := NewSessionManager(kadLB)
	fwd := mgr.New("foo", nil)
	defer mgr.Reap(fwd.Id())
	p, err := fwd.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 {
		t.Fatalf("get first count; expected 1, got %d", len(p))
	}
	if !bytes.Equal(p[0].Address(), bytesNear) {
		t.Fatalf("get first address; expected %x, got %x", bytesNear, p[0].Address())
	}

	p, err = fwd.Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 {
		t.Fatalf("get peers count; expected 1, got %d", len(p))
	}
	if !bytes.Equal(p[0].Address(), bytesFar) {
		t.Fatalf("get second address; expected %x, got %x", bytesFar, p[0].Address())
	}
	log.Trace("peer", "peer", p)

}
