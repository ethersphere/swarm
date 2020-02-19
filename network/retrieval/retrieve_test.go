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

package retrieval

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/chunk"
	chunktesting "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/p2p/protocols"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

var (
	bucketKeyFileStore = simulation.BucketKey("filestore")
	bucketKeyNetstore  = simulation.BucketKey("netstore")

	hash0 = sha3.Sum256([]byte{0})
)

func init() {
	testutil.Init()
}

// TestChunkDelivery brings up two nodes, stores a few chunks on the first node, then tries to retrieve them from the second node
func TestChunkDelivery(t *testing.T) {
	chunkCount := 10
	filesize := chunkCount * 4096

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-retrieve": newBzzRetrieveWithLocalstore,
	}, true)
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		log.Debug("uploader node", "enode", nodeIDs[0])

		fs := sim.MustNodeItem(nodeIDs[0], bucketKeyFileStore).(*storage.FileStore)

		//put some data into just the first node
		data := make([]byte, filesize)
		if _, err := io.ReadFull(rand.Reader, data); err != nil {
			t.Fatalf("reading from crypto/rand failed: %v", err.Error())
		}
		refs, err := getAllRefs(data)
		if err != nil {
			return err
		}
		log.Trace("got all refs", "refs", refs)
		_, wait, err := fs.Store(context.Background(), bytes.NewReader(data), int64(filesize), false)
		if err != nil {
			return err
		}
		if err := wait(context.Background()); err != nil {
			return err
		}

		id, err := sim.AddNode()
		if err != nil {
			return err
		}
		err = sim.Net.Connect(id, nodeIDs[0])
		if err != nil {
			return err
		}
		nodeIDs = sim.UpNodeIDs()
		if len(nodeIDs) != 2 {
			return fmt.Errorf("wrong number of nodes, expected %d got %d", 2, len(nodeIDs))
		}

		// allow the two nodes time to set up the protocols otherwise kademlias will be empty when retrieve requests happen
		time.Sleep(50 * time.Millisecond)
		log.Debug("fetching through node", "enode", nodeIDs[1])
		ns := sim.MustNodeItem(nodeIDs[1], bucketKeyNetstore).(*storage.NetStore)
		ctr := 0
		for _, ch := range refs {
			ctr++
			_, err := ns.Get(context.Background(), chunk.ModeGetRequest, storage.NewRequest(ch))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestNoSuitablePeer brings up two nodes, tries to retrieve a chunk which is never
// found, expecting a NoSuitablePeer error from netstore
func TestNoSuitablePeer(t *testing.T) {
	nodes := 2

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-retrieve": newBzzRetrieveWithLocalstore,
	}, true)
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := sim.AddNodesAndConnectFull(nodes)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodes {
			t.Fatal("not enough nodes up")
		}
		// allow the two nodes time to set up the protocols otherwise kademlias will be empty when retrieve requests happen
		i := 0
		for iterate := true; iterate; {
			kinfo := sim.MustNodeItem(nodeIDs[1], simulation.BucketKeyKademlia).(*network.Kademlia).KademliaInfo()
			if kinfo.TotalConnections != 1 {
				i++
			} else {
				break
			}
			time.Sleep(50 * time.Millisecond)
			if i == 5 {
				t.Fatal("timed out waiting for 1 connections")
			}
		}

		log.Debug("fetching through node", "enode", nodeIDs[1])
		ns := sim.MustNodeItem(nodeIDs[1], bucketKeyNetstore).(*storage.NetStore)
		c := chunktesting.GenerateTestRandomChunk()

		ref := c.Address()
		_, err := ns.Get(context.Background(), chunk.ModeGetRequest, storage.NewRequest(ref))
		if err == nil {
			return errors.New("expected netstore retrieval error but got none")
		}
		if err != storage.ErrNoSuitablePeer {
			return fmt.Errorf("expected ErrNoSuitablePeer but got %v instead", err)
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestUnsolicitedChunkDelivery tests that a node is dropped in response to an unsolicited chunk delivery
// this case covers a chunk Ruid that was not previously known to the downstream peer
func TestUnsolicitedChunkDelivery(t *testing.T) {
	pk, ns, cleanup := newTestNetstore(t)
	defer cleanup()
	bzzAddr := network.PrivateKeyToBzzKey(pk)

	kad := network.NewKademlia(bzzAddr, network.NewKadParams())

	tester, _, teardown, err := newRetrievalTester(t, pk, ns, kad)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	node := tester.Nodes[0]

	// deliver with a RUID which cannot be found
	tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Non-existent RUID chunk delivery",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: &ChunkDelivery{
						Ruid:  1234,
						Addr:  nil,
						SData: nil,
					},
					Peer: node.ID(),
				},
			},
		})

	// expect peer disconnection
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("subprotocol error")})

	if err != nil {
		t.Fatal(err)
	}
}

// TestUnsolicitedChunkDeliveryFaultyAddr tests that a misbehaving node cannot send a chunk delivery
// over a known retrieve request Ruid with a chunk address that does not match the requested address
func TestUnsolicitedChunkDeliveryFaultyAddr(t *testing.T) {
	pk, ns, cleanup := newTestNetstore(t)
	defer cleanup()
	bzzAddr := network.PrivateKeyToBzzKey(pk)

	kad := network.NewKademlia(bzzAddr, network.NewKadParams())

	tester, r, teardown, err := newRetrievalTester(t, pk, ns, kad)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()
	ns.RemoteGet = func(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
		return &enode.ID{}, func() {}, nil
	}
	node := tester.Nodes[0]

	// this exchange is needed so that the protocol peer gets created and that r.peers is not nil
	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "A bogus retrieve request",
			Triggers: []p2ptest.Trigger{
				{
					Code: 1,
					Msg: &RetrieveRequest{
						Ruid: 9876,
						Addr: []byte{5, 4, 3, 2},
					},
					Peer: node.ID(),
				},
			},
		},
	)
	for i := 0; i < 1000; i++ {
		if r.getPeer(node.ID()) != nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	// inject a supposed retrieve request that was sent to that peer
	r.getPeer(node.ID()).addRetrieval(1234, []byte{0, 1, 2, 3})

	// respond with a chunk delivery with the same Ruid but with a different chunk address
	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Ruid accepted but chunk address invalid",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: &ChunkDelivery{
						Ruid:  1234,
						Addr:  []byte{0, 2, 1, 3},
						SData: nil,
					},
					Peer: node.ID(),
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	// expect disconnection
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("subprotocol error")})

	if err != nil {
		t.Fatal(err)
	}
}

// TestUnsolicitedChunkDeliveryDouble tests that a misbehaving node cannot send a chunk delivery
// twice over a known retrieve request Ruid
func TestUnsolicitedChunkDeliveryDouble(t *testing.T) {
	pk, ns, cleanup := newTestNetstore(t)
	defer cleanup()
	bzzAddr := network.PrivateKeyToBzzKey(pk)

	kad := network.NewKademlia(bzzAddr, network.NewKadParams())

	tester, r, teardown, err := newRetrievalTester(t, pk, ns, kad)
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()
	ns.RemoteGet = func(ctx context.Context, req *storage.Request, localID enode.ID) (*enode.ID, func(), error) {
		return &enode.ID{}, func() {}, nil
	}
	node := tester.Nodes[0]

	// this exchange is needed so that the protocol peer gets created and that r.peers is not nil
	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "A bogus retrieve request",
			Triggers: []p2ptest.Trigger{
				{
					Code: 1,
					Msg: &RetrieveRequest{
						Ruid: 9876,
						Addr: []byte{5, 4, 3, 2},
					},
					Peer: node.ID(),
				},
			},
		},
	)
	for i := 0; i < 1000; i++ {
		if r.getPeer(node.ID()) != nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	// inject a supposed retrieve request that was sent to that peer
	r.getPeer(node.ID()).addRetrieval(1234, []byte{0, 1, 2, 3})

	// respond with a chunk delivery with the same Ruid and the matching chunk address
	err = tester.TestExchanges(
		p2ptest.Exchange{
			Label: "Ruid & Chunk address correct",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: &ChunkDelivery{
						Ruid:  1234,
						Addr:  []byte{0, 1, 2, 3},
						SData: nil,
					},
					Peer: node.ID(),
				},
			},
		},
		p2ptest.Exchange{
			Label: "Ruid seen, Chunk Address incorrect",
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg: &ChunkDelivery{
						Ruid:  1234,
						Addr:  []byte{0, 2, 1, 3},
						SData: nil,
					},
					Peer: node.ID(),
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	// expect disconnection
	err = tester.TestDisconnected(&p2ptest.Disconnect{Peer: node.ID(), Error: errors.New("subprotocol error")})

	if err != nil {
		t.Fatal(err)
	}
}

// TestDeliveryForwarding tests that chunk delivery forwarding requests happen. It creates three nodes (fetching, forwarding and uploading)
// where po(fetching,forwarding) = 1 and po(forwarding,uploading) = 1, then uploads chunks to the uploading node, afterwards
// tries to retrieve the relevant chunks (ones with po = 0 to fetching i.e. no bits in common with fetching and with
// po >= 1 with uploading i.e. with 1 bit or more in common with the uploading)
func TestDeliveryForwarding(t *testing.T) {
	chunkCount := 100
	filesize := chunkCount * 4096
	sim, uploader, forwarder, fetcher := setupTestDeliveryForwardingSimulation(t)
	defer sim.Close()

	log.Debug("test delivery forwarding", "uploader", uploader, "forwarder", forwarder, "fetcher", fetcher)

	uploaderNodeStore := sim.MustNodeItem(uploader, bucketKeyFileStore).(*storage.FileStore)
	fetcherBase := sim.MustNodeItem(fetcher, simulation.BucketKeyKademlia).(*network.Kademlia).BaseAddr()
	uploaderBase := sim.MustNodeItem(fetcher, simulation.BucketKeyKademlia).(*network.Kademlia).BaseAddr()
	ctx := context.Background()
	_, wait, err := uploaderNodeStore.Store(ctx, testutil.RandomReader(101010, filesize), int64(filesize), false)
	if err != nil {
		t.Fatal(err)
	}
	if err = wait(ctx); err != nil {
		t.Fatal(err)
	}

	chunks, err := getChunks(uploaderNodeStore.ChunkStore)
	if err != nil {
		t.Fatal(err)
	}
	for c := range chunks {
		addr, err := hex.DecodeString(c)
		if err != nil {
			t.Fatal(err)
		}

		// try to retrieve all of the chunks which have no bits in common with the
		// fetcher, but have more than one bit in common with the uploader node
		if chunk.Proximity(addr, fetcherBase) == 0 && chunk.Proximity(addr, uploaderBase) >= 1 {
			req := storage.NewRequest(chunk.Address(addr))
			fetcherNetstore := sim.MustNodeItem(fetcher, bucketKeyNetstore).(*storage.NetStore)
			_, err := fetcherNetstore.Get(ctx, chunk.ModeGetRequest, req)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func setupTestDeliveryForwardingSimulation(t *testing.T) (sim *simulation.Simulation, uploader, forwarder, fetching enode.ID) {
	sim = simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		"bzz-retrieve": newBzzRetrieveWithLocalstore,
	}, true)

	fetching, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	fetcherBase := sim.MustNodeItem(fetching, simulation.BucketKeyKademlia).(*network.Kademlia).BaseAddr()

	override := func(o *adapters.NodeConfig) func(*adapters.NodeConfig) {
		return func(c *adapters.NodeConfig) {
			*o = *c
		}
	}

	// create a node that will be in po 1 from fetcher
	forwarderConfig := nodeConfigAtPo(t, fetcherBase, 1)
	forwarder, err = sim.AddNode(override(forwarderConfig))
	if err != nil {
		t.Fatal(err)
	}

	err = sim.Net.Connect(fetching, forwarder)
	if err != nil {
		t.Fatal(err)
	}

	forwarderBase := sim.MustNodeItem(forwarder, simulation.BucketKeyKademlia).(*network.Kademlia).BaseAddr()

	// create a node on which the files will be stored at po 1 in relation to the forwarding node
	uploaderConfig := nodeConfigAtPo(t, forwarderBase, 1)
	uploader, err = sim.AddNode(override(uploaderConfig))
	if err != nil {
		t.Fatal(err)
	}

	err = sim.Net.Connect(forwarder, uploader)
	if err != nil {
		t.Fatal(err)
	}

	return sim, uploader, forwarder, fetching
}

// if there is one peer in the Kademlia, RequestFromPeers should return it
func TestRequestFromPeers(t *testing.T) {
	dummyPeerID := enode.HexID("3431c3939e1ee2a6345e976a8234f9870152d64879f30bc272a074f6859e75e8")

	addr := network.RandomBzzAddr()
	to := network.NewKademlia(addr.OAddr, network.NewKadParams())
	protocolsPeer := protocols.NewPeer(p2p.NewPeer(dummyPeerID, "dummy", []p2p.Cap{{Name: "bzz-retrieve", Version: 1}}), nil, nil)
	peer := network.NewPeer(&network.BzzPeer{
		BzzAddr: network.RandomBzzAddr(),
		Peer:    protocolsPeer,
	}, to)

	to.On(peer)

	s := New(to, nil, addr, nil)

	req := storage.NewRequest(storage.Address(hash0[:]))
	id, err := s.findPeerLB(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if id.ID() != dummyPeerID {
		t.Fatalf("Expected an id, got %v", id)
	}
}

//TestHasPriceImplementation is to check that Retrieval provides priced messages
func TestHasPriceImplementation(t *testing.T) {
	price := (&ChunkDelivery{}).Price()
	if price == nil || price.Value == 0 {
		t.Fatal("No prices set for chunk delivery msg")
	}

	price = (&RetrieveRequest{}).Price()
	if price == nil || price.Value == 0 {
		t.Fatal("No prices set for retrieve requests")
	}
}

func newBzzRetrieveWithLocalstore(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
	n := ctx.Config.Node()
	addr := network.NewBzzAddrFromEnode(n)

	localStore, localStoreCleanup, err := newTestLocalStore(n.ID(), addr, nil)
	if err != nil {
		return nil, nil, err
	}

	k, _ := bucket.LoadOrStore(simulation.BucketKeyKademlia, network.NewKademlia(addr.Over(), network.NewKadParams()))
	kad := k.(*network.Kademlia)

	netStore := storage.NewNetStore(localStore, addr)
	lnetStore := storage.NewLNetStore(netStore)
	fileStore := storage.NewFileStore(lnetStore, lnetStore, storage.NewFileStoreParams(), chunk.NewTags())

	var store *state.DBStore
	// Use on-disk DBStore to reduce memory consumption in race tests.
	dir, err := ioutil.TempDir("", "statestore-")
	if err != nil {
		return nil, nil, err
	}
	store, err = state.NewDBStore(dir)
	if err != nil {
		return nil, nil, err
	}

	r := New(kad, netStore, addr, nil)
	netStore.RemoteGet = r.RequestFromPeers
	bucket.Store(bucketKeyFileStore, fileStore)
	bucket.Store(bucketKeyNetstore, netStore)
	bucket.Store(simulation.BucketKeyKademlia, kad)

	cleanup = func() {
		localStore.Close()
		localStoreCleanup()
		store.Close()
		os.RemoveAll(dir)
	}

	return r, cleanup, nil
}

func newTestLocalStore(id enode.ID, addr *network.BzzAddr, globalStore mock.GlobalStorer) (localStore *localstore.DB, cleanup func(), err error) {
	dir, err := ioutil.TempDir("", "localstore-")
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		os.RemoveAll(dir)
	}

	var mockStore *mock.NodeStore
	if globalStore != nil {
		mockStore = globalStore.NewNodeStore(common.BytesToAddress(id.Bytes()))
	}

	localStore, err = localstore.New(dir, addr.Over(), &localstore.Options{
		MockStore: mockStore,
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return localStore, cleanup, nil
}

func getAllRefs(testData []byte) (storage.AddressCollection, error) {
	datadir, err := ioutil.TempDir("", "chunk-debug")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(datadir)
	fileStore, cleanup, err := storage.NewLocalFileStore(datadir, make([]byte, 32), chunk.NewTags())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	reader := bytes.NewReader(testData)
	return fileStore.GetAllReferences(context.Background(), reader)
}

func getChunks(store chunk.Store) (chunks map[string]struct{}, err error) {
	chunks = make(map[string]struct{})
	for po := uint8(0); po <= chunk.MaxPO; po++ {
		last, err := store.LastPullSubscriptionBinID(po)
		if err != nil {
			return nil, err
		}
		if last == 0 {
			continue
		}
		ch, _ := store.SubscribePull(context.Background(), po, 0, last)
		for c := range ch {
			addr := c.Address.Hex()
			if _, ok := chunks[addr]; ok {
				return nil, fmt.Errorf("duplicate chunk %s", addr)
			}
			chunks[addr] = struct{}{}
		}
	}
	return chunks, nil
}

// nodeConfigAtPo brute forces a node config to create a node that has an overlay address at the provided po in relation to the given baseaddr
func nodeConfigAtPo(t *testing.T, baseaddr []byte, po int) *adapters.NodeConfig {
	foundPo := -1
	var conf *adapters.NodeConfig
	for foundPo != po {
		conf = adapters.RandomNodeConfig()
		ip := net.IPv4(127, 0, 0, 1)
		enrIP := enr.IP(ip)
		conf.Record.Set(&enrIP)
		enrTCPPort := enr.TCP(conf.Port)
		conf.Record.Set(&enrTCPPort)
		enrUDPPort := enr.UDP(0)
		conf.Record.Set(&enrUDPPort)

		err := enode.SignV4(&conf.Record, conf.PrivateKey)
		if err != nil {
			t.Fatalf("unable to generate ENR: %v", err)
		}
		nod, err := enode.New(enode.V4ID{}, &conf.Record)
		if err != nil {
			t.Fatalf("unable to create enode: %v", err)
		}

		n := network.NewBzzAddrFromEnode(nod)
		foundPo = chunk.Proximity(baseaddr, n.Over())
	}

	return conf
}

func newRetrievalTester(t *testing.T, prvkey *ecdsa.PrivateKey, netStore *storage.NetStore, kad *network.Kademlia) (*p2ptest.ProtocolTester, *Retrieval, func(), error) {
	t.Helper()

	if prvkey == nil {
		key, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("Could not generate key")
		}
		prvkey = key
	}

	r := New(kad, netStore, network.NewBzzAddr(kad.BaseAddr(), nil), nil)
	protocolTester := p2ptest.NewProtocolTester(prvkey, 1, r.runProtocol)

	return protocolTester, r, protocolTester.Stop, nil
}

func newTestNetstore(t *testing.T) (prvkey *ecdsa.PrivateKey, netStore *storage.NetStore, cleanup func()) {
	t.Helper()
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

	netStore = storage.NewNetStore(localStore, network.NewBzzAddr(bzzAddr, nil))

	cleanup = func() {
		err = netStore.Close()
		if err != nil {
			t.Fatalf("Could not close netStore")
		}
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatalf("Could not remove localstore dir")
		}
	}
	return prvkey, netStore, cleanup
}
