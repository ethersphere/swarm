package network

import (
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/pot"
)

func testKadPeerFromAddress(bzzAddr *BzzAddr, k *Kademlia) *Peer {
	return NewPeer(&BzzPeer{
		BzzAddr:   bzzAddr,
		LightNode: false,
	}, k)
}

func testKadBzzAddrFromAddress(addr pot.Address) *BzzAddr {
	return &BzzAddr{
		OAddr: addr.Bytes(),
		UAddr: addr.Bytes(),
	}
}

func TestKademliaGet(t *testing.T) {
	addr := pot.RandomAddress()
	addrBytes := addr.Bytes()
	kp := NewKadParams()
	k := NewKademlia(addrBytes, kp)

	bzzAddrOne := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 8))
	bzzAddrTwoFirst := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 16))
	bzzAddrTwoSecond := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 16))
	bzzAddrFourFirst := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 64))
	bzzAddrFourSecond := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 64))

	peerOne := testKadPeerFromAddress(bzzAddrOne, k)
	peerFourFirst := testKadPeerFromAddress(bzzAddrTwoFirst, k)
	peerFourSecond := testKadPeerFromAddress(bzzAddrTwoSecond, k)
	peerTwoFirst := testKadPeerFromAddress(bzzAddrFourFirst, k)
	peerTwoSecond := testKadPeerFromAddress(bzzAddrFourSecond, k)

	k.Register(bzzAddrOne)
	k.Register(bzzAddrFourFirst)
	k.Register(bzzAddrFourSecond)
	k.Register(bzzAddrTwoFirst)
	k.Register(bzzAddrTwoSecond)

	k.On(peerOne)
	k.On(peerTwoFirst)
	k.On(peerTwoSecond)
	k.On(peerFourFirst)
	k.On(peerFourSecond)

	//	peers, _ := k.GetConnsBin(addrBytes, 255)
	//	for _, p := range peers {
	//		t.Logf("peer %x", p.Address())
	//	}
	//
	//	peers, _ = k.GetConnsBin(addrBytes, 64)
	//	for _, p := range peers {
	//		t.Logf("peer %x", p.Address())
	//	}
	//
	//	peers, _ = k.GetConnsBin(addrBytes, 63)
	//	for _, p := range peers {
	//		t.Logf("peer %x", p.Address())
	//	}
	//
	//	peers, _ = k.GetConnsBin(addrBytes[:1], 64)
	//	for _, p := range peers {
	//		t.Logf("peer %x", p.Address())
	//	}

	bzzConfig := &BzzConfig{
		OverlayAddr:  addrBytes,
		UnderlayAddr: addrBytes,
		HiveParams:   NewHiveParams(),
		NetworkID:    42,
	}
	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("bzz", NewBzz(bzzConfig, k, nil, nil, nil))
	var peersRpc []*Peer
	err := rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 63)
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range peersRpc {
		t.Logf("peer %x", p.Address())
	}
}
