package network

import (
	"bytes"
	"context"
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

	var bzzAddrFill []*BzzAddr
	var peerFill []*Peer
	for i := 0; i < 8; i++ {
		bzzAddrFill = append(bzzAddrFill, testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, i)))
		peerFill = append(peerFill, testKadPeerFromAddress(bzzAddrFill[i], k))
		k.Register(bzzAddrFill[i])
		k.On(peerFill[i])
	}

	bzzConfig := &BzzConfig{
		OverlayAddr:  addrBytes,
		UnderlayAddr: addrBytes,
		HiveParams:   NewHiveParams(),
		NetworkID:    42,
	}
	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("bzz", NewBzz(bzzConfig, k, nil, nil, nil))
	peersRpc := []*Peer{}
	err := rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 8)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range peersRpc {
		t.Logf("peer %x", p.Address())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msgC := make(chan KademliaNotification)
	sub, err := rpcClient.Subscribe(ctx, "bzz", msgC, "receive")
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()

	bzzAddrExtra := testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, 7))
	peerExtra := testKadPeerFromAddress(bzzAddrExtra, k)
	k.Register(bzzAddrExtra)
	notification := <-msgC
	if notification.Depth != 6 || notification.Serial != 6 {
		t.Fatalf("Expected depth/serial 6/6, got %d/%d", notification.Depth, notification.Serial)
	}

	k.On(peerExtra)
	notification = <-msgC
	if notification.Depth != 7 || notification.Serial != 7 {
		t.Fatalf("Expected depth/serial 7/7, got %d/%d", notification.Depth, notification.Serial)
	}

	peersRpc = []*Peer{}
	err = rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 7)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range peersRpc {
		if !bytes.Equal(p.BzzAddr.OAddr, bzzAddrExtra.OAddr) && !bytes.Equal(p.BzzAddr.OAddr, bzzAddrFill[7].OAddr) {
			t.Fatalf("Unexpected peer %s", p.BzzAddr.OAddr)
		}
	}

	peersRpc = []*Peer{}
	err = rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 6)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range peersRpc {
		if !bytes.Equal(p.BzzAddr.OAddr, bzzAddrFill[6].OAddr) {
			t.Fatalf("Unexpected peer %s", p.BzzAddr.OAddr)
		}
	}
	t.Log(k)

	k.Off(peerFill[2])
	notification = <-msgC
	if notification.Depth != 2 || notification.Serial != 8 {
		t.Fatalf("Expected depth/serial 7/7, got %d/%d", notification.Depth, notification.Serial)
	}
}
