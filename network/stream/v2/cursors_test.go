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

package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/state"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// TestNodesExchangeCorrectBinIndexes tests that two nodes exchange the correct cursors for all streams
// it tests that all streams are exchanged
func TestNodesExchangeCorrectBinIndexes(t *testing.T) {
	const (
		nodeCount  = 2
		chunkCount = 1000
	)

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		serviceNameStream: newSyncSimServiceFunc(&SyncSimServiceOptions{
			InitialChunkCount: chunkCount,
		}),
	}, true)
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	getCursorsCopy := func(sim *simulation.Simulation, idOne, idOther enode.ID) map[string]uint64 {
		r := nodeRegistry(sim, idOne)
		if r == nil {
			return nil
		}
		p := r.getPeer(idOther)
		if p == nil {
			return nil
		}
		return p.getCursorsCopy()
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			return errors.New("not enough nodes up")
		}

		// periodically check for cursors
		for i := 0; i < 100; i++ {
			// wait for the nodes to exchange StreamInfo messages
			time.Sleep(10 * time.Millisecond)

			idOne := nodeIDs[0]
			idOther := nodeIDs[1]
			onesCursors := getCursorsCopy(sim, idOne, idOther)
			othersCursors := getCursorsCopy(sim, idOther, idOne)

			onesBins := nodeInitialBinIndexes(sim, idOne)
			othersBins := nodeInitialBinIndexes(sim, idOther)

			err1 := compareNodeBinsToStreams(t, onesCursors, othersBins)
			if err1 != nil {
				err = err1 // set the resulting error when the loop is done
			}
			err2 := compareNodeBinsToStreams(t, othersCursors, onesBins)
			if err2 != nil {
				err = err2 // set the resulting error when the loop is done
			}
			if err1 == nil && err2 == nil {
				return nil
			}
		}

		return err
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

// TestNodesCorrectBinsDynamic adds nodes to a star topology, connecting new nodes to the pivot node
// after each connection is made, the cursors on the pivot are checked, to reflect the bins that we are
// currently still interested in. this makes sure that correct bins are of interest
// when nodes enter the kademlia of the pivot node
func TestNodesCorrectBinsDynamic(t *testing.T) {
	const (
		nodeCount  = 6
		chunkCount = 500
	)

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		serviceNameStream: newSyncSimServiceFunc(&SyncSimServiceOptions{
			InitialChunkCount: chunkCount,
		}),
	}, true)
	defer sim.Close()

	ctx, cancel := context.WithTimeout(context.Background(), simContextTimeout)
	defer cancel()
	_, err := sim.AddNodesAndConnectStar(2)
	if err != nil {
		t.Fatal(err)
	}

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) error {
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != 2 {
			return errors.New("not enough nodes up")
		}

		// wait for the nodes to exchange StreamInfo messages
		wantCursorsCount := 17
		for i := 499; i >= 0; i-- { // wait time 5s
			time.Sleep(10 * time.Millisecond)
			count1 := nodeRegistry(sim, nodeIDs[0]).getPeer(nodeIDs[1]).cursorsCount()
			count2 := nodeRegistry(sim, nodeIDs[1]).getPeer(nodeIDs[0]).cursorsCount()
			if count1 >= wantCursorsCount && count2 >= wantCursorsCount {
				break
			}
			if i == 0 {
				return fmt.Errorf("got cursors %v and %v, want %v", count1, count2, wantCursorsCount)
			}
		}

		idPivot := nodeIDs[0]
		pivotSyncer := nodeRegistry(sim, idPivot)
		pivotKademlia := nodeKademlia(sim, idPivot)
		pivotDepth := uint(pivotKademlia.NeighbourhoodDepth())

		for j := 2; j <= nodeCount; j++ {
			// append a node to the simulation
			id, err := sim.AddNodes(1)
			if err != nil {
				return err
			}
			err = sim.Net.ConnectNodesStar(id, nodeIDs[0])
			if err != nil {
				return err
			}
			nodeIDs := sim.UpNodeIDs()
			if len(nodeIDs) != j+1 {
				return fmt.Errorf("not enough nodes up. got %d, want %d", len(nodeIDs), j)
			}
			time.Sleep(50 * time.Millisecond)
			idPivot = nodeIDs[0]
			for i := 1; i < j; i++ {
				idOther := nodeIDs[i]
				otherKademlia := sim.MustNodeItem(idOther, simulation.BucketKeyKademlia).(*network.Kademlia)
				po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
				depth := pivotKademlia.NeighbourhoodDepth()
				pivotCursors := pivotSyncer.getPeer(idOther).getCursorsCopy()

				// check that the pivot node is interested just in bins >= depth
				if po >= depth {
					othersBins := nodeInitialBinIndexes(sim, idOther)
					if err := compareNodeBinsToStreamsWithDepth(t, pivotCursors, othersBins, pivotDepth); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

var reestablishCursorsSnapshotFilename = "testdata/reestablish-cursors-snapshot.json"

// TestNodesRemovesCursors creates a pivot network of 2 nodes where the pivot's depth = 0.
// test sequence:
// - select another node with po >= depth (of the pivot's kademlia)
// - add other nodes to the pivot until the depth goes above that peer's po (depth > peerPo)
// - asserts that the pivot does not maintain any cursors of the node that moved out of depth
// - start removing nodes from the simulation until that peer is again within depth
// - check that the cursors are being re-established
func TestNodeRemovesAndReestablishCursors(t *testing.T) {
	t.Skip("disable to find more optimal way to run it")

	if *update {
		generateReestablishCursorsSnapshot(t, 2)
	}

	const chunkCount = 1000

	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		serviceNameStream: newSyncSimServiceFunc(nil),
	}, true)
	defer sim.Close()

	// load the snapshot
	if err := sim.UploadSnapshot(context.Background(), reestablishCursorsSnapshotFilename); err != nil {
		t.Fatal(err)
	}
	// load additional test specific data from the snapshot
	d, err := ioutil.ReadFile(reestablishCursorsSnapshotFilename)
	if err != nil {
		t.Fatal(err)
	}
	var s reestablishCursorsState
	if err := json.Unmarshal(d, &s); err != nil {
		t.Fatal(err)
	}

	pivotEnode := s.PivotEnode
	pivotKademlia := sim.MustNodeItem(pivotEnode, simulation.BucketKeyKademlia).(*network.Kademlia)
	lookupEnode := s.LookupEnode
	lookupPO := s.PO
	nodeCount := len(sim.UpNodeIDs())

	mustUploadChunks(context.Background(), t, nodeFileStore(sim, pivotEnode), chunkCount)

	log.Debug("tracking enode", "enode", lookupEnode, "po", lookupPO)

	// expecting some cursors
	waitForCursors(t, sim, pivotEnode, lookupEnode, true)

	//append nodes to simulation until the node po moves out of the depth, then assert no subs from pivot to that node
	for i := float64(1); pivotKademlia.NeighbourhoodDepth() <= lookupPO; i++ {
		// calculate the number of nodes to add:
		// - logarithmically increase by the number of iterations
		//   - ensure that the logarithm is greater then 0 by starting the iteration from 1, not 0
		//   - ensure that the logarithm is greater then 0 by adding 1
		// - multiply by the difference between target and current depth
		//   - ensure that is greater then 0 by adding 1
		//   - multiply it by empirical constant 4
		newNodeCount := int(math.Logb(i)+1) * ((lookupPO-pivotKademlia.NeighbourhoodDepth())*4 + 1)
		id, err := sim.AddNodes(newNodeCount)
		if err != nil {
			t.Fatal(err)
		}
		err = sim.Net.ConnectNodesStar(id, pivotEnode)
		if err != nil {
			t.Fatal(err)
		}
		nodeCount += newNodeCount
		nodeIDs := sim.UpNodeIDs()
		if len(nodeIDs) != nodeCount {
			t.Fatalf("got %v up nodes, want %v", len(nodeIDs), nodeCount)
		}
		log.Debug("added new nodes to reach depth", "new nodes", newNodeCount, "current depth", pivotKademlia.NeighbourhoodDepth(), "target depth", lookupPO)
	}

	log.Debug("added nodes to sim, node moved out of depth", "depth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "lookupEnode", lookupEnode)

	// no cursors should exist at this point
	waitForCursors(t, sim, pivotEnode, lookupEnode, false)

	var removed int
	// remove nodes from the simulation until the peer moves again into depth
	log.Error("pulling the plug on some nodes to make the depth go up again", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "lookupEnode", lookupEnode)
	for pivotKademlia.NeighbourhoodDepth() > lookupPO {
		_, err := sim.StopRandomNode(pivotEnode, lookupEnode)
		if err != nil {
			t.Fatal(err)
		}
		removed++
		log.Debug("removed 1 node", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO)
	}
	log.Debug("done removing nodes", "pivotDepth", pivotKademlia.NeighbourhoodDepth(), "peerPo", lookupPO, "removed", removed)

	// expecting some new cursors
	waitForCursors(t, sim, pivotEnode, lookupEnode, true)
}

// data appended to reestablish cursors snapshot
type reestablishCursorsState struct {
	PivotEnode  enode.ID `json:"pivotEnode"`
	LookupEnode enode.ID `json:"lookupEnode"`
	PO          int      `json:"po"`
}

// function that generates a simulation and saves its snapshot for
// TestNodeRemovesAndReestablishCursors test.
func generateReestablishCursorsSnapshot(t *testing.T, tagetPO int) {
	sim, pivotEnode, lookupEnode := setupReestablishCursorsSimulation(t, tagetPO)
	defer sim.Close()

	waitForCursors(t, sim, pivotEnode, lookupEnode, true)

	s, err := sim.Net.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	d, err := json.Marshal(struct {
		*simulations.Snapshot
		reestablishCursorsState
	}{
		Snapshot: s,
		reestablishCursorsState: reestablishCursorsState{
			PivotEnode:  pivotEnode,
			LookupEnode: lookupEnode,
			PO:          tagetPO,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("save snapshot file")

	err = ioutil.WriteFile(reestablishCursorsSnapshotFilename, d, 0666)
	if err != nil {
		t.Fatal(err)
	}
}

// function that generates a simulation that can be used in TestNodeRemovesAndReestablishCursors
// test with a provided target po which is a depth to reach in the test by adding new nodes.
func setupReestablishCursorsSimulation(t *testing.T, tagetPO int) (sim *simulation.Simulation, pivotEnode, lookupEnode enode.ID) {
	// initial node count
	nodeCount := 5

	sim = simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		serviceNameStream: newSyncSimServiceFunc(nil),
	}, true)

	nodeIDs, err := sim.AddNodesAndConnectStar(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	pivotEnode = nodeIDs[0]
	log.Debug("simulation pivot node", "id", pivotEnode)
	pivotKademlia := sim.MustNodeItem(pivotEnode, simulation.BucketKeyKademlia).(*network.Kademlia)

	// make sure that we get a node with po <= depth
	for i := 1; i < nodeCount; i++ {
		log.Debug("looking for a peer", "i", i, "nodecount", nodeCount)
		otherKademlia := sim.MustNodeItem(nodeIDs[i], simulation.BucketKeyKademlia).(*network.Kademlia)
		po := chunk.Proximity(otherKademlia.BaseAddr(), pivotKademlia.BaseAddr())
		depth := pivotKademlia.NeighbourhoodDepth()
		if po > depth {
			if po != tagetPO {
				log.Debug("wrong depth to reach, generating new simulation", "depth", po)
				return setupReestablishCursorsSimulation(t, tagetPO)
			}
			lookupEnode = nodeIDs[i]
			return
		}
		// append a node to the simulation
		id, err := sim.AddNode()
		if err != nil {
			t.Fatal(err)
		}
		log.Debug("added node to simulation, connecting to pivot", "id", id, "pivot", pivotEnode)
		if err = sim.Net.Connect(id, pivotEnode); err != nil {
			t.Fatal(err)
		}
		nodeCount++
		// wait for node to be set
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("node with po<=depth not found")
	return
}

// waitForCursors checks if the pivot node has some cursors or not
// by periodically checking for them.
func waitForCursors(t *testing.T, sim *simulation.Simulation, pivotEnode, lookupEnode enode.ID, wantSome bool) {
	t.Helper()

	var got int
	for i := 0; i < 1000; i++ { // 10s total wait
		time.Sleep(10 * time.Millisecond)
		s, ok := sim.Service(serviceNameStream, pivotEnode).(*Registry)
		if !ok {
			continue
		}
		p := s.getPeer(lookupEnode)
		if p == nil {
			continue
		}
		got = len(p.getCursorsCopy())
		if got != 0 == wantSome {
			return
		}
	}
	if wantSome {
		t.Fatalf("got %v cursors, but want some", got)
	} else {
		t.Fatalf("got %v cursors, but want none", got)
	}
}

// compareNodeBinsToStreams checks that the values on `onesCursors` correlate to the values in `othersBins`
// onesCursors represents the stream cursors that node A knows about node B (i.e. they shoud reflect directly in this case
// the values which node B retrieved from its local store)
// othersBins is the array of bin indexes on node B's local store as they were inserted into the store
func compareNodeBinsToStreams(t *testing.T, onesCursors map[string]uint64, othersBins []uint64) (err error) {
	if len(onesCursors) == 0 {
		return errors.New("no cursors")
	}
	if len(othersBins) == 0 {
		return errors.New("no bins")
	}

	for nameKey, cur := range onesCursors {
		id, err := parseSyncKey(parseID(nameKey).Key)
		if err != nil {
			return err
		}
		if othersBins[id] != cur {
			return fmt.Errorf("bin indexes not equal. bin %d, got %d, want %d", id, cur, othersBins[id])
		}
	}
	return nil
}

func compareNodeBinsToStreamsWithDepth(t *testing.T, onesCursors map[string]uint64, othersBins []uint64, depth uint) (err error) {
	log.Debug("compareNodeBinsToStreamsWithDepth", "cursors", onesCursors, "othersBins", othersBins, "depth", depth)
	if len(onesCursors) == 0 || len(othersBins) == 0 {
		return errors.New("no cursors")
	}
	// inclusive test
	for nameKey, cur := range onesCursors {
		bin, err := parseSyncKey(parseID(nameKey).Key)
		if err != nil {
			return err
		}
		if uint(bin) < depth {
			return fmt.Errorf("cursor at bin %d should not exist. depth %d", bin, depth)
		}
		if othersBins[bin] != cur {
			return fmt.Errorf("bin indexes not equal. bin %d, got %d, want %d", bin, cur, othersBins[bin])
		}
	}

	// exclusive test
	for i := uint8(0); i < uint8(depth); i++ {
		// should not have anything shallower than depth
		id := NewID("SYNC", encodeSyncKey(i))
		if _, ok := onesCursors[id.String()]; ok {
			return fmt.Errorf("oneCursors contains id %s, but it should not", id)
		}
	}
	return nil
}

// TestCorrectCursorsExchangeRace brings up two nodes with a random config
// then generates a whole bunch of bogus nodes at different POs from the pivot node
// those POs are then turned On in the pivot Kademlia in order to trigger depth changes
// without creating real nodes which slow down the test and the CI execution. The depth changes
// trigger different cursor requests from the pivot to the other node, these requests are being intercepted
// by a mock Stream handler which later on sends the replies to those requests in different order.
// the test finishes after the random replies are processed and the correct cursors are asserted according to the
// real kademlia depth. This test is to accommodate for possible race conditions where multiple cursors requests are
// sent but the kademlia depth keeps changing. This in turn causes to possibly discard some contents of those requests which
// are still in flight and which responses' are not yet processed
func TestCorrectCursorsExchangeRace(t *testing.T) {
	bogusNodeCount := 15
	bogusNodes := []*network.Peer{}
	popRandomNode := func() *network.Peer {
		log.Debug("bogus peer array length", "len", len(bogusNodes))
		i := rand.Intn(len(bogusNodes))
		elem := bogusNodes[i]
		bogusNodes = append(bogusNodes[:i], bogusNodes[i+1:]...)
		return elem
	}
	streamInfoRes := []*StreamInfoRes{}
	infoReqHook := func(msg *StreamInfoReq) {
		log.Trace("mock got StreamInfoReq msg", "msg", msg)

		//create the response
		res := &StreamInfoRes{}
		for _, v := range msg.Streams {
			cur, err := parseSyncKey(v.Key)
			if err != nil {
				t.Fatal(err)
			}
			desc := StreamDescriptor{
				Stream:  v,
				Cursor:  uint64(cur),
				Bounded: false,
			}
			res.Streams = append(res.Streams, desc)
		}
		streamInfoRes = append(streamInfoRes, res)
	}

	popRandomResponse := func() *StreamInfoRes {
		log.Debug("responses array length", "len", len(streamInfoRes))
		i := rand.Intn(len(streamInfoRes))
		elem := streamInfoRes[i]
		streamInfoRes = append(streamInfoRes[:i], streamInfoRes[i+1:]...)
		return elem
	}
	opts := &SyncSimServiceOptions{
		StreamConstructorFunc: func(s state.Store, b []byte, p ...StreamProvider) node.Service {
			return New(s, b, p...)
		},
	}
	sim := simulation.NewBzzInProc(map[string]simulation.ServiceFunc{
		serviceNameStream: newSyncSimServiceFunc(opts),
	}, true)
	defer sim.Close()

	// create the first node with the non mock initialiser
	pivot, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}

	// second node should start with the mock protocol
	opts.StreamConstructorFunc = func(s state.Store, b []byte, p ...StreamProvider) node.Service {
		return newMock(infoReqHook)
	}

	other, err := sim.AddNode()
	if err != nil {
		t.Fatal(err)
	}
	err = sim.Net.Connect(pivot, other)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	pivotKad := nodeKademlia(sim, pivot)
	pivotAddr := pot.NewAddressFromBytes(pivotKad.BaseAddr())
	pivotStream := nodeRegistry(sim, pivot)

	otherBase := nodeKademlia(sim, other).BaseAddr()
	otherPeer := pivotStream.getPeer(other)

	log.Debug(pivotKad.String())
	// add a few fictional nodes at higher POs so that we get kademlia depth change and as a result a trigger
	// for a StreamInfoReq message between the two 'real' nodes

	for i := 0; i < bogusNodeCount; i++ {
		rw := &p2p.MsgPipeRW{}
		ptpPeer := p2p.NewPeer(enode.ID{}, "wu tang killa beez", []p2p.Cap{})
		protoPeer := protocols.NewPeer(ptpPeer, rw, &protocols.Spec{})
		peerAddr := pot.RandomAddressAt(pivotAddr, i)
		bzzPeer := &network.BzzPeer{
			Peer:    protoPeer,
			BzzAddr: network.NewBzzAddr(peerAddr.Bytes(), []byte(fmt.Sprintf("%x", peerAddr[:]))),
		}
		peer := network.NewPeer(bzzPeer, pivotKad)
		pivotKad.On(peer)
		bogusNodes = append(bogusNodes, peer)
		time.Sleep(50 * time.Millisecond)
	}
CHECKSTREAMS:
	pivotDepth := pivotKad.NeighbourhoodDepth()
	po := chunk.Proximity(otherBase, pivotKad.BaseAddr())
	sub, qui := syncSubscriptionsDiff(po, -1, pivotDepth, pivotKad.MaxProxDisplay, false) //s.syncBinsOnlyWithinDepth)
	log.Debug("got desired pivot cursor state", "depth", pivotDepth, "subs", sub, "quits", qui)

	for i := len(streamInfoRes); i > 0; i-- {
		v := popRandomResponse()
		pivotStream.clientHandleStreamInfoRes(context.Background(), otherPeer, v)
	}

	//get the pivot cursors for peer, assert equal to what is in `sub`
	for _, stream := range getAllSyncStreams() {
		cur, ok := otherPeer.getCursor(stream)
		keyInt, err := parseSyncKey(stream.Key)
		if err != nil {
			t.Fatal(err)
		}
		shouldExist := checkKeyInSlice(int(keyInt), sub)

		if shouldExist == ok {
			continue
		} else {
			t.Fatalf("got a cursor that should not exist. key %s, cur %d", stream.Key, cur)
		}
	}

	// repeat, until all of the bogus nodes are out of the way
	if len(bogusNodes) > 0 {
		p := popRandomNode()
		pivotKad.Off(p)
		time.Sleep(50 * time.Millisecond) // wait for the streamInfoReq to come through
		goto CHECKSTREAMS
	}
}

type slipStreamMock struct {
	spec              *protocols.Spec
	streamInfoReqHook func(*StreamInfoReq)
}

func newMock(infoReqHook func(*StreamInfoReq)) *slipStreamMock {
	return &slipStreamMock{
		spec:              Spec,
		streamInfoReqHook: infoReqHook,
	}
}

func (s *slipStreamMock) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "bzz-stream",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     s.Run,
		},
	}
}

func (s *slipStreamMock) APIs() []rpc.API {
	return nil
}

func (s *slipStreamMock) Close() {
}

func (s *slipStreamMock) Start(server *p2p.Server) error {
	return nil
}

func (s *slipStreamMock) Stop() error {
	return nil
}

func (s *slipStreamMock) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, s.spec)
	return peer.Run(s.HandleMsg)
}

func (s *slipStreamMock) HandleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {
	case *StreamInfoReq:
		s.streamInfoReqHook(msg)
	case *GetRange:
		return nil
	default:
		panic("unexpected")
	}
	return nil
}

func getAllSyncStreams() (streams []ID) {
	for i := uint8(0); i <= 16; i++ {
		streams = append(streams, ID{
			Name: syncStreamName,
			Key:  encodeSyncKey(i),
		})
	}
	return
}
