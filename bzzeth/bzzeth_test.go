// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bzzeth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

func newBzzEthTester(t *testing.T, prvkey *ecdsa.PrivateKey, netStore *storage.NetStore) (*p2ptest.ProtocolTester, *BzzEth, func(), error) {
	t.Helper()

	if prvkey == nil {
		key, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("Could not generate key")
		}
		prvkey = key
	}

	b := New(netStore, nil)
	protocolTester := p2ptest.NewProtocolTester(prvkey, 1, b.Run)
	teardown := func() {
		protocolTester.Stop()
	}

	return protocolTester, b, teardown, nil
}

func newTestNetworkStore(t *testing.T) (prvkey *ecdsa.PrivateKey, netStore *storage.NetStore, cleanup func()) {
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Could not generate key")
	}

	dir, err := ioutil.TempDir("", "localstore-")
	if err != nil {
		t.Fatalf("Could not create localStore temp dir")
	}

	bzzAddr := network.PrivateKeyToBzzKey(prvkey)
	localStore, err := localstore.New(dir, bzzAddr, nil)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Could not create localStore")
	}

	netStore = storage.NewNetStore(localStore, bzzAddr,  enode.ID{})

	cleanup = func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatalf("Could not remove localstore dir")
		}
		err = netStore.Close()
		if err != nil {
			t.Fatalf("Could not close netStore")
		}

	}
	return prvkey, netStore, cleanup
}

func handshakeExchange(tester *p2ptest.ProtocolTester, peerID enode.ID, serveHeadersPeer, serveHeadersPivot bool) error {
	return tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Handshake",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: Handshake{
						ServeHeaders: serveHeadersPeer,
					},
					Peer: peerID,
				},
			},
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg: Handshake{
						ServeHeaders: serveHeadersPivot,
					},
					Peer: peerID,
				},
			},
		})
}

// This message is exchanged between two Swarm nodes to check if the connection drops
func dummyHandshakeMessage(tester *p2ptest.ProtocolTester, peerID enode.ID) error {
	return tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Handshake",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: Handshake{
						ServeHeaders: true,
					},
					Peer: peerID,
				},
			},
		})
}

// tests handshake between eth node and swarm node
// on successful handshake the protocol does not go idle
// peer added to the pool and serves headers is registered
func TestBzzEthHandshake(t *testing.T) {
	tester, b, teardown, err := newBzzEthTester(t, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	node := tester.Nodes[0]
	err = handshakeExchange(tester, node.ID(), true, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// after successful handshake, expect peer added to peer pool
	p := getPeerAfterConnection(node.ID(), b)
	if p == nil {
		t.Fatal("bzzeth peer not added")
	}
	if !p.serveHeaders {
		t.Fatal("bzzeth peer serveHeaders not set")
	}

	close(b.quit)
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("?")})
	if err == nil || err.Error() != "timed out waiting for peers to disconnect" {
		t.Fatal(err)
	}
}

// TestBzzBzzHandshake tests that a handshake between two Swarm nodes
func TestBzzBzzHandshake(t *testing.T) {
	tester, b, teardown, err := newBzzEthTester(t, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	node := tester.Nodes[0]
	err = handshakeExchange(tester, node.ID(), false, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// after successful handshake, expect peer added to peer pool
	p := getPeerAfterConnection(node.ID(), b)
	if p == nil {
		t.Fatal("bzzeth peer not added")
	}

	// after closing the protocol, expect disconnect
	close(b.quit)
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("?")})
	if err == nil || err.Error() != "timed out waiting for peers to disconnect" {
		t.Fatal(err)
	}
}

// TestBzzBzzHandshakeWithMessage tests that a handshake between two Swarm nodes and message exchange
// disconnects the peer
func TestBzzBzzHandshakeWithMessage(t *testing.T) {
	// redefine isSwarmNodeFunc to force recognise remote peer as swarm node
	defer func(f func(*Peer) bool) {
		isSwarmNodeFunc = f
	}(isSwarmNodeFunc)
	isSwarmNodeFunc = func(_ *Peer) bool { return true }

	tester, b, teardown, err := newBzzEthTester(t, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	node := tester.Nodes[0]
	err = handshakeExchange(tester, node.ID(), false, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// after successful handshake, expect peer added to peer pool
	var p *Peer
	for i := 0; i < 10; i++ {
		p = b.peers.get(node.ID())
		if p != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if p == nil {
		t.Fatal("bzzeth peer not added")
	}

	// Send a dummy handshake message, wait for sometime and check if peer is dropped
	err = dummyHandshakeMessage(tester, node.ID())
	if err != nil {
		t.Fatal(err)
	}
	// after successful handshake, expect peer added to peer pool
	p1 := isPeerDisconnected(node.ID(), b)
	if p1 != nil {
		t.Fatal("bzzeth peer still connected")
	}
}

func getPeerAfterConnection(id enode.ID, b *BzzEth) (p *Peer) {
	for i := 0; i < 10; i++ {
		p = b.peers.get(id)
		if p != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func isPeerDisconnected(id enode.ID, b *BzzEth) (p *Peer) {
	var p1 *Peer
	for i := 0; i < 10; i++ {
		p1 = b.peers.get(id)
		if p1 == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func newBlockHeaderExchange(tester *p2ptest.ProtocolTester, peerID enode.ID, requestID uint32, offered *NewBlockHeaders, wanted [][]byte) error {
	return tester.TestExchanges(
		p2ptest.Exchange{
			Label: "NewBlockHeaders",
			Triggers: []p2ptest.Trigger{
				{
					Code: 1,
					Msg: offered,
					Peer: peerID,
				},
			},
			Expects: []p2ptest.Expect{
				{
					Code: 2,
					Msg: GetBlockHeaders{
						ID:     requestID,
						Hashes: wanted,
					},
					Peer: peerID,
				},
			},
		})
}

func blockHeaderExchange(tester *p2ptest.ProtocolTester, peerID enode.ID, requestID uint32, wantedData []rlp.RawValue) error {
	return tester.TestExchanges(
		p2ptest.Exchange{
			Label: "BlockHeaders",
			Triggers: []p2ptest.Trigger{
				{
					Code: 3,
					Msg: BlockHeaders{
						ID:      requestID,
						Headers: wantedData,
					},
					Peer: peerID,
				},
			},
		})
}

// Test bzzeth full eth node sends new block header hashes
// respond with a GetBlockHeaders requesting headers falling into the proximity of this node
// Also test two other conditions
// - If a header is already present in localstore, dont request it in GetBlockHeaders
// - If a unsolicited header is received, dont store it on localstore
// Apart from that it also tests if all the headers are delivered and stored in localstore
func TestNewBlockHeaders(t *testing.T) {
	prvKey, netstore, cleanup := newTestNetworkStore(t)
	defer cleanup()

	// bzz pivot - full eth node peer
	// NewBlockHeaders trigger, expect
	tester, _, teardown, err := newBzzEthTester(t, prvKey, netstore)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	offered := make(NewBlockHeaders, 256)
	for i := 0; i < len(offered); i++ {
		offered[i].Hash = common.BytesToHash(crypto.Keccak256([]byte{uint8(i)}))
		offered[i].Number = uint64(i)
	}

	// redefine wantHeadeFunc for this test
	wantedIndexes := []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}
	IgnoreIndexes := []int{77}
	wantHeaderFunc = func(hash []byte, _ *network.Kademlia) bool {
		for _, i := range wantedIndexes {
			if bytes.Equal(hash, offered[i].Hash.Bytes()) {
				return true
			}
			// Add the ignore headers (headers in localstore already) to the valid list
			if bytes.Equal(hash, offered[IgnoreIndexes[0]].Hash.Bytes()) {
				return true
			}
		}
		return false
	}

	wanted := make([][]byte, len(wantedIndexes))
	wantedData := make([]rlp.RawValue, len(wantedIndexes)+1)
	for i, w := range wantedIndexes {
		wanted[i] = crypto.Keccak256([]byte{uint8(w)})
		wantedData[i] = rlp.RawValue{uint8(w)}
	}

	// overwrite newRequestIDFunc to be deterministic
	defer func(f func() uint32) {
		newRequestIDFunc = f
	}(newRequestIDFunc)

	newRequestIDFunc = func() uint32 {
		return 42
	}

	// overwrite finishStorageFunc to test deterministic storage of headers
	finishStorageTesting := func(chunks []chunk.Chunk) {
		checkStorage(t, wantedIndexes, wanted, wantedData, netstore)
	}
	finishStorageFunc = finishStorageTesting

	// overwrite finishDeliveryFunc to test deterministic delivery of headers
	finishDeliveryTesting := func(hashes map[string]bool) {
		checkDelivery(t,wantedIndexes, wanted, hashes)
	}
	finishDeliveryFunc = finishDeliveryTesting

	node := tester.Nodes[0]
	err = handshakeExchange(tester, node.ID(), true, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Add a header to localstore
	// this header should not be requested in GetBlockHeaders
	_, err = netstore.Store.Put(context.Background(), chunk.ModePutUpload, newChunk([]byte{uint8(IgnoreIndexes[0])}))
	if err != nil {
		t.Fatal(err)
	}

	err = newBlockHeaderExchange(tester, node.ID(), newRequestIDFunc(), &offered, wanted)
	if err != nil {
		t.Fatal(err)
	}

	// Add a unsolicited header
	wantedData[len(wantedIndexes)] = rlp.RawValue{uint8(255)}
	err = blockHeaderExchange(tester, node.ID(), newRequestIDFunc(), wantedData)
	if err != nil {
		t.Fatal(err)
	}
}

func checkStorage(t *testing.T, wantedIndexes []int, wanted [][]byte, wantedData []rlp.RawValue, netstore *storage.NetStore) {
	// Check if requested headers arrived and are stored in localstore
	for i, _ := range wantedIndexes {
		chunk, err := netstore.Store.Get(context.Background(), chunk.ModeGetLookup, wanted[i])
		if err != nil {
			t.Fatalf("chunk  %v not found %v", hex.EncodeToString(wanted[i]), wantedData[i])
		}
		if !bytes.Equal(wantedData[i], chunk.Data()) {
			t.Fatalf("expected %v, got %v", wanted[i], chunk.Data())
		}
	}

	// check if unsolicited header delivery is dropped and not in localstore
	hash := crypto.Keccak256(wantedData[len(wantedIndexes)])
	yes, err := netstore.Store.Has(context.Background(), hash)
	if err != nil {
		t.Fatal(err)
	}
	if yes {
		t.Fatalf("unsolicited header %v is not dropped", hex.EncodeToString(hash))
	}
	return
}

func checkDelivery(t *testing.T, wantedIndexes []int, wanted [][]byte, hashes map[string]bool) {
	for i, _ := range wantedIndexes {
		hash := hex.EncodeToString(wanted[i])
		if _, ok := hashes[hash]; !ok {
			t.Fatalf("Header  %v not delivered", hash)
		}
	}
}