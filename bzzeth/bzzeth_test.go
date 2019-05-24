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
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/network"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
)

var (
	loglevel = flag.Int("loglevel", 0, "verbosity of logs")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

func newBzzEthTester() (*p2ptest.ProtocolTester, *BzzEth, func(), error) {
	b := New(nil, nil)

	prvkey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	protocolTester := p2ptest.NewProtocolTester(prvkey, 1, b.Run)
	teardown := func() {
		protocolTester.Stop()
	}

	return protocolTester, b, teardown, nil
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

// tests handshake between eth node and swarm node
// on successful handshake the protocol does not go idle
// peer added to the pool and serves headers is registered
func TestBzzEthHandshake(t *testing.T) {
	tester, b, teardown, err := newBzzEthTester()
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
	tester, b, teardown, err := newBzzEthTester()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	// redefine isSwarmNodeFunc to force recognise remote peer as swarm node
	defer func(f func(*Peer) bool) {
		isSwarmNodeFunc = f
	}(isSwarmNodeFunc)
	isSwarmNodeFunc = func(_ *Peer) bool { return true }

	node := tester.Nodes[0]
	err = handshakeExchange(tester, node.ID(), false, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// after handshake expect protocol to hang, peer not added to pool
	p := b.peers.get(node.ID())
	if p != nil {
		t.Fatal("bzzeth swarm peer incorrectly added")
	}

	// after closing the ptotocall, expect disconnect
	close(b.quit)
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("protocol returned")})
	if err != nil {
		t.Fatal(err)
	}

}

func newBlockHeaderExchange(tester *p2ptest.ProtocolTester, peerID enode.ID, requestID uint32, offered []HeaderHash, wanted [][]byte) error {
	return tester.TestExchanges(
		p2ptest.Exchange{
			Label: "NewBlockHeaders",
			Triggers: []p2ptest.Trigger{
				{
					Code: 1,
					Msg: NewBlockHeaders{
						Headers: offered,
					},
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

// Test bzzeth full eth node sends new block header hashes
// respond with a GetBlockHeaders requesting headers falling inte
func TestNewBlockHeaders(t *testing.T) {
	// bzz pivot - full eth node peer
	// NewBlockHeaders trigger, expect
	tester, _, teardown, err := newBzzEthTester()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	offered := make([]HeaderHash, 256)
	for i := 0; i < len(offered); i++ {
		offered[i] = HeaderHash{crypto.Keccak256([]byte{uint8(i)}), []byte{uint8(i)}}
	}

	// redefine wantHeadeFunc for this test
	defer func(f func([]byte, *network.Kademlia) bool) {
		wantHeaderFunc = f
	}(wantHeaderFunc)
	wantedIndexes := []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}
	wantHeaderFunc = func(hash []byte, _ *network.Kademlia) bool {
		for _, i := range wantedIndexes {
			if bytes.Equal(hash, offered[i].Hash) {
				return true
			}
		}
		return false
	}

	wanted := make([][]byte, len(wantedIndexes))
	for i, w := range wantedIndexes {
		wanted[i] = crypto.Keccak256([]byte{uint8(w)})
	}

	// overwrite newRequestIDFunc to be deterministic
	defer func(f func() uint32) {
		newRequestIDFunc = f
	}(newRequestIDFunc)

	newRequestIDFunc = func() uint32 {
		return 42
	}

	node := tester.Nodes[0]

	err = handshakeExchange(tester, node.ID(), true, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = newBlockHeaderExchange(tester, node.ID(), 42, offered, wanted)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: subsequent NewBlockHeaders msg are not allowed
	// TODO: headers found in localstore should not be requested
	// TODO: after requested headers arrive and are stored
	// TODO: unrequested header delivery drops
}
