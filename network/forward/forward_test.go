package forward

import (
	"testing"

	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/pot"
)

func TestGet(t *testing.T) {
	addr := make([]byte, 32)
	kadParams := network.NewKadParams()
	kad := network.NewKademlia(addr, kadParams)
	cp := capability.NewCapability(4, 2)
	kad.RegisterCapabilityIndex("foo", *cp)
	sctx := NewSessionContext("foo")
	fwd := New(sctx)

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
}
