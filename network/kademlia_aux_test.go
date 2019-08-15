package network

import (
	"bytes"
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
)

func testKadPeerFromAddress(bzzAddr *BzzAddr, k *Kademlia) (*Peer, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	enodeId := enode.PubkeyToIDV4(&privKey.PublicKey)
	p2pPeer := p2p.NewPeer(enodeId, enodeId.String(), []p2p.Cap{})
	protoPeer := protocols.NewPeer(p2pPeer, nil, nil)

	return NewPeer(&BzzPeer{
		BzzAddr:   bzzAddr,
		LightNode: false,
		Peer:      protoPeer,
	}, k), err
}

func testKadBzzAddrFromAddress(addr pot.Address) *BzzAddr {
	return &BzzAddr{
		OAddr: addr.Bytes(),
		UAddr: addr.Bytes(),
	}
}

// TODO: return po from rpc
func TestKademliaGet(t *testing.T) {
	addr := pot.RandomAddress()
	addrBytes := addr.Bytes()

	bzzConfig := &BzzConfig{
		OverlayAddr:  addrBytes,
		UnderlayAddr: addrBytes,
		HiveParams:   NewHiveParams(),
		NetworkID:    42,
	}
	bzz := NewBzz(bzzConfig, nil)
	k := bzz.Kademlia

	var bzzAddrFill []*BzzAddr
	var peerFill []*Peer
	for i := 0; i < 8; i++ {
		bzzAddrFill = append(bzzAddrFill, testKadBzzAddrFromAddress(pot.RandomAddressAt(addr, i)))
		p, err := testKadPeerFromAddress(bzzAddrFill[i], k)
		if err != nil {
			t.Fatal(err)
		}
		peerFill = append(peerFill, p)
		k.Register(bzzAddrFill[i])
		k.On(peerFill[i])
	}

	rpcSrv := rpc.NewServer()
	rpcClient := rpc.DialInProc(rpcSrv)
	rpcSrv.RegisterName("bzz", bzz)
	peersRpc := []*Peer{}
	err := rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 0, 8)
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
	peerExtra, err := testKadPeerFromAddress(bzzAddrExtra, k)
	if err != nil {
		t.Fatal(err)
	}
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
	err = rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 0, 7)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range peersRpc {
		if !bytes.Equal(p.BzzAddr.OAddr, bzzAddrExtra.OAddr) && !bytes.Equal(p.BzzAddr.OAddr, bzzAddrFill[7].OAddr) {
			t.Fatalf("Unexpected peer %s", p.BzzAddr.OAddr)
		}
	}

	peersRpc = []*Peer{}
	err = rpcClient.Call(&peersRpc, "bzz_getConnsBin", addrBytes, 0, 6)
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
