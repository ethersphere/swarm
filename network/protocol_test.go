// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"crypto/ecdsa"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/p2p/protocols"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/testutil"
)

const (
	TestProtocolVersion = 12
)

var TestProtocolNetworkID = DefaultTestNetworkID

func init() {
	testutil.Init()
}

func HandshakeMsgExchange(lhs, rhs *HandshakeMsg, id enode.ID) []p2ptest.Exchange {
	return []p2ptest.Exchange{
		{
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg:  lhs,
					Peer: id,
				},
			},
		},
		{
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg:  rhs,
					Peer: id,
				},
			},
		},
	}
}

func newBzzHandshakeMsg(version uint64, networkId uint64, addr *BzzAddr, lightNode bool) *HandshakeMsg {
	capabilities := capability.NewCapabilities()
	var cap *capability.Capability
	if lightNode {
		cap = newLightCapability()
	} else {
		cap = newFullCapability()
	}
	capabilities.Add(cap)
	msg := &HandshakeMsg{
		Version:      version,
		NetworkID:    networkId,
		Addr:         addr,
		Capabilities: capabilities,
	}

	return msg
}

func newBzzBaseTester(n int, prvkey *ecdsa.PrivateKey, spec *protocols.Spec, run func(*BzzPeer) error) (*bzzTester, error) {
	var addrs [][]byte
	for i := 0; i < n; i++ {
		addr := pot.RandomAddress()
		addrs = append(addrs, addr[:])
	}
	pt, _, err := newBzzBaseTesterWithAddrs(prvkey, addrs, spec, run)
	return pt, err
}

func newBzzBaseTesterWithAddrs(prvkey *ecdsa.PrivateKey, addrs [][]byte, spec *protocols.Spec, run func(*BzzPeer) error) (*bzzTester, [][]byte, error) {
	n := len(addrs)
	cs := make(map[enode.ID]chan bool)
	var csMu sync.Mutex

	srv := func(p *BzzPeer) error {
		defer func() {
			csMu.Lock()
			defer csMu.Unlock()
			if cs[p.ID()] != nil {
				close(cs[p.ID()])
			}
		}()
		return run(p)
	}
	mu := &sync.Mutex{}
	nodeToAddr := make(map[enode.ID][]byte)
	protocol := func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		mu.Lock()
		nodeToAddr[p.ID()] = addrs[0]
		mu.Unlock()
		bzzAddr := &BzzAddr{addrs[0], []byte(p.Node().String())}
		addrs = addrs[1:]
		return srv(&BzzPeer{Peer: protocols.NewPeer(p, rw, spec), BzzAddr: bzzAddr})
	}

	s := p2ptest.NewProtocolTester(prvkey, n, protocol)
	var record enr.Record
	bzzKey := PrivateKeyToBzzKey(prvkey)
	record.Set(NewENRAddrEntry(bzzKey))
	err := enode.SignV4(&record, prvkey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate ENR: %v", err)
	}
	nod, err := enode.New(enode.V4ID{}, &record)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create enode: %v", err)
	}
	addr := getENRBzzAddr(nod)

	csMu.Lock()
	for _, node := range s.Nodes {
		log.Warn("node", "node", node)
		cs[node.ID()] = make(chan bool)
	}
	csMu.Unlock()

	var nodeAddrs [][]byte
	pt := &bzzTester{
		addr:           addr,
		ProtocolTester: s,
		cs:             cs,
	}
	mu.Lock()
	for _, n := range pt.Nodes {
		nodeAddrs = append(nodeAddrs, nodeToAddr[n.ID()])
	}
	mu.Unlock()

	return pt, nodeAddrs, nil
}

type bzzTester struct {
	*p2ptest.ProtocolTester
	addr *BzzAddr
	cs   map[enode.ID]chan bool
	bzz  *Bzz
}

func newBzz(addr *BzzAddr, lightNode bool) *Bzz {
	config := &BzzConfig{
		OverlayAddr:  addr.Over(),
		UnderlayAddr: addr.Under(),
		HiveParams:   NewHiveParams(),
		NetworkID:    DefaultTestNetworkID,
		LightNode:    lightNode,
	}
	kad := NewKademlia(addr.OAddr, NewKadParams())
	bzz := NewBzz(config, kad, nil, nil, nil)
	return bzz
}

func newBzzHandshakeTester(n int, prvkey *ecdsa.PrivateKey, lightNode bool) (*bzzTester, error) {

	var record enr.Record
	bzzkey := PrivateKeyToBzzKey(prvkey)
	record.Set(NewENRAddrEntry(bzzkey))
	record.Set(ENRLightNodeEntry(lightNode))
	err := enode.SignV4(&record, prvkey)
	if err != nil {
		return nil, err
	}
	nod, err := enode.New(enode.V4ID{}, &record)
	addr := getENRBzzAddr(nod)

	bzz := newBzz(addr, lightNode)

	pt := p2ptest.NewProtocolTester(prvkey, n, bzz.runBzz)

	return &bzzTester{
		addr:           addr,
		ProtocolTester: pt,
		bzz:            bzz,
	}, nil
}

// should test handshakes in one exchange? parallelisation
func (s *bzzTester) testHandshake(lhs, rhs *HandshakeMsg, disconnects ...*p2ptest.Disconnect) error {
	if err := s.TestExchanges(HandshakeMsgExchange(lhs, rhs, rhs.Addr.ID())...); err != nil {
		return err
	}

	if len(disconnects) > 0 {
		return s.TestDisconnected(disconnects...)
	}

	// If we don't expect disconnect, ensure peers remain connected
	err := s.TestDisconnected(&p2ptest.Disconnect{
		Peer:  s.Nodes[0].ID(),
		Error: nil,
	})

	if err == nil {
		return fmt.Errorf("Unexpected peer disconnect")
	}

	if err.Error() != "timed out waiting for peers to disconnect" {
		return err
	}

	return nil
}

func correctBzzHandshake(addr *BzzAddr, lightNode bool) *HandshakeMsg {
	return newBzzHandshakeMsg(TestProtocolVersion, TestProtocolNetworkID, addr, lightNode)
}

func TestBzzHandshakeNetworkIDMismatch(t *testing.T) {
	lightNode := false
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	s, err := newBzzHandshakeTester(1, prvkey, lightNode)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	node := s.Nodes[0]

	err = s.testHandshake(
		correctBzzHandshake(s.addr, lightNode),
		newBzzHandshakeMsg(TestProtocolVersion, 321, NewAddr(node), false),
		&p2ptest.Disconnect{Peer: node.ID(), Error: fmt.Errorf("Handshake error: Message handler error: (msg code 0): network id mismatch 321 (!= %v)", TestProtocolNetworkID)},
	)

	if err != nil {
		t.Fatal(err)
	}
}

func TestBzzHandshakeVersionMismatch(t *testing.T) {
	lightNode := false
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	s, err := newBzzHandshakeTester(1, prvkey, lightNode)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	node := s.Nodes[0]

	err = s.testHandshake(
		correctBzzHandshake(s.addr, lightNode),
		newBzzHandshakeMsg(0, TestProtocolNetworkID, NewAddr(node), false),
		&p2ptest.Disconnect{Peer: node.ID(), Error: fmt.Errorf("Handshake error: Message handler error: (msg code 0): version mismatch 0 (!= %d)", TestProtocolVersion)},
	)

	if err != nil {
		t.Fatal(err)
	}
}

// Checks the temporary checkHandshake criteria for legacy full/light capabilities presets
func TestBzzHandshakeInvalidCapabilities(t *testing.T) {
	lightNode := false
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	s, err := newBzzHandshakeTester(1, prvkey, lightNode)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	node := s.Nodes[0]

	msg := newBzzHandshakeMsg(TestProtocolVersion, TestProtocolNetworkID, NewAddr(node), false)
	cap := msg.Capabilities.Get(0)
	cap.Set(14)
	err = s.testHandshake(
		correctBzzHandshake(s.addr, lightNode),
		msg,
		&p2ptest.Disconnect{Peer: node.ID(), Error: fmt.Errorf("Handshake error: Message handler error: (msg code 0): invalid capabilities setting: %s", msg.Capabilities)},
	)

	if err != nil {
		t.Fatal(err)
	}
}
func TestBzzHandshakeSuccess(t *testing.T) {
	lightNode := false
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	s, err := newBzzHandshakeTester(1, prvkey, lightNode)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	node := s.Nodes[0]

	err = s.testHandshake(
		correctBzzHandshake(s.addr, lightNode),
		newBzzHandshakeMsg(TestProtocolVersion, TestProtocolNetworkID, NewAddr(node), false),
	)

	if err != nil {
		t.Fatal(err)
	}
}

func TestBzzHandshakeLightNode(t *testing.T) {
	var lightNodeTests = []struct {
		name      string
		lightNode bool
	}{
		{"on", true},
		{"off", false},
	}

	for _, test := range lightNodeTests {
		t.Run(test.name, func(t *testing.T) {
			prvkey, err := crypto.GenerateKey()
			if err != nil {
				t.Fatal(err)
			}
			pt, err := newBzzHandshakeTester(1, prvkey, false)
			if err != nil {
				t.Fatal(err)
			}
			defer pt.Stop()

			node := pt.Nodes[0]
			addr := NewAddr(node)

			err = pt.testHandshake(
				correctBzzHandshake(pt.addr, false),
				newBzzHandshakeMsg(TestProtocolVersion, TestProtocolNetworkID, addr, test.lightNode),
			)

			if err != nil {
				t.Fatal(err)
			}

			var nodeCapability *capability.Capability
			if test.lightNode {
				nodeCapability = lightCapability
			} else {
				nodeCapability = fullCapability
			}
			select {

			case <-pt.bzz.handshakes[node.ID()].done:
				for _, cp := range pt.bzz.handshakes[node.ID()].Capabilities.Caps {
					if cp.String() != nodeCapability.String() {
						t.Fatalf("peer LightNode flag is %v, should be %v", cp.String(), nodeCapability.String())
					}
				}
			case <-time.After(10 * time.Second):
				t.Fatal("test timeout")
			}
		})
	}
}
