// Copyright 2019 The go-ethereum Authors
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
package swap

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

/*
This file adds some in-process simulations to Swap.

It is NOT an integration test; it does not test integration of Swap with other
protocols like stream or retrieval; it is independent of backends and blockchains,
and is purely meant for testing the accounting functionality across nodes.

For integration tests, run test cluster deployments with all integration modueles
(blockchains, oracles, etc.)
*/

var bucketKeySwap = simulation.BucketKey("swap")

// swapSimulationParams allows to avoid global variables for the test
type swapSimulationParams struct {
	swaps       map[int]*Swap
	dirs        map[int]string
	count       int
	maxMsgPrice int
	minMsgPrice int
	nodeCount   int
	backend     *backends.SimulatedBackend
}

// define test message types
type testMsgBySender struct{}
type testMsgByReceiver struct{}
type testMsgBigPrice struct{}

// create a test Spec; every node has its Spec and its accounting Hook
func newTestSpec() *protocols.Spec {
	return &protocols.Spec{
		Name:       "testSpec",
		Version:    1,
		MaxMsgSize: 10 * 1024 * 1024,
		Messages: []interface{}{
			testMsgBySender{},
			testMsgByReceiver{},
			testMsgBigPrice{},
		},
	}
}

// testPrices holds prices for these test messages
type testPrices struct {
	priceMatrix map[reflect.Type]*protocols.Price
}

// assign prices for the test messages
func (tp *testPrices) newTestPriceMatrix() {
	tp.priceMatrix = map[reflect.Type]*protocols.Price{
		reflect.TypeOf(testMsgBySender{}): {
			Value:   1000, // arbitrary price for now
			PerByte: true,
			Payer:   protocols.Sender,
		},
		reflect.TypeOf(testMsgByReceiver{}): {
			Value:   100, // arbitrary price for now
			PerByte: false,
			Payer:   protocols.Receiver,
		},
		reflect.TypeOf(testMsgBigPrice{}): {
			Value:   DefaultPaymentThreshold + 1,
			PerByte: false,
			Payer:   protocols.Sender,
		},
	}
}

// Price returns the price for a (test) message
func (tp *testPrices) Price(msg interface{}) *protocols.Price {
	t := reflect.TypeOf(msg).Elem()
	return tp.priceMatrix[t]
}

// testService encapsulates objects needed for the simulation
type testService struct {
	swap  *Swap
	spec  *protocols.Spec
	peers map[enode.ID]*testPeer
}

func newTestService() *testService {
	return &testService{
		peers: make(map[enode.ID]*testPeer),
	}
}

// testPeer is our object for the test protocol; we can use it to handle our own messages
type testPeer struct {
	*protocols.Peer
}

// handle our own messages; we don't need to do anything (yet), we only
// want messages to be sent and received, and we need this function for the protocol spec
func (tp *testPeer) handleMsg(ctx context.Context, msg interface{}) error {
	return nil
}

// newSimServiceMap creates the `ServiceFunc` map for node initialization.
// The trick we need to apply is that we need to create a `SimulatedBackend`
// with all accounts for every simulation node pre-loaded with "funds".
// To do this, we pass a `swapSimulationParams` object to this function,
// which contains the shared objects needed to initialize the `SimulatedBackend`
func newSimServiceMap(params *swapSimulationParams) map[string]simulation.ServiceFunc {
	simServiceMap := map[string]simulation.ServiceFunc{
		// we need the bzz service in order to build up a kademlia
		"bzz": func(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
			addr := network.NewAddr(ctx.Config.Node())
			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			kad := network.NewKademlia(addr.Over(), network.NewKadParams())
			return network.NewBzz(config, kad, nil, nil, nil), nil, nil
		},
		// and we also use a swap service
		"swap": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			// every simulation node has an instance of a `testService`
			ts := newTestService()
			// balance is the interface for `NewAccounting`; it is a Swap
			balance := params.swaps[params.count]
			// every node is a different instance of a Swap and gets a different entry in the map
			params.count++
			// to create the accounting, we also need a `Price` instance
			prices := &testPrices{}
			// create the matrix of test prices
			prices.newTestPriceMatrix()
			ts.spec = newTestSpec()
			// create the accounting instance and assign to the spec
			ts.spec.Hook = protocols.NewAccounting(balance, prices)
			ts.swap = balance
			// deploy the accounting to the `SimulatedBackend`
			testDeploy(context.Background(), balance.backend, balance)
			params.backend.Commit()
			// store the testService into the bucket
			bucket.Store(bucketKeySwap, ts)

			cleanup = func() {
				for _, dir := range params.dirs {
					os.RemoveAll(dir)
				}
			}

			return ts, cleanup, nil
		},
	}
	return simServiceMap
}

// newSharedBackendSwaps pre-loads each simulated node account with "funds"
// so that later in the simulation all operations have sufficient gas
func newSharedBackendSwaps(nodeCount int) (*swapSimulationParams, error) {
	params := &swapSimulationParams{
		swaps:       make(map[int]*Swap),
		dirs:        make(map[int]string),
		maxMsgPrice: 10000,
		minMsgPrice: 100,
		nodeCount:   nodeCount,
	}
	keys := make(map[int]*ecdsa.PrivateKey)
	addrs := make(map[int]common.Address)
	alloc := core.GenesisAlloc{}
	stores := make(map[int]*state.DBStore)

	// for each node, generate keys, a GenesisAccount and a state store
	for i := 0; i < nodeCount; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		keys[i] = key
		addrs[i] = crypto.PubkeyToAddress(key.PublicKey)
		alloc[addrs[i]] = core.GenesisAccount{Balance: big.NewInt(10000000000)}
		dir, err := ioutil.TempDir("", fmt.Sprintf("swap_test_store_%d", i))
		if err != nil {
			return nil, err
		}
		stateStore, err2 := state.NewDBStore(dir)
		if err2 != nil {
			return nil, err
		}
		params.dirs[i] = dir
		stores[i] = stateStore
	}
	// then create the single SimulatedBackend
	gasLimit := uint64(8000000000)
	defaultBackend := backends.NewSimulatedBackend(alloc, gasLimit)
	// finally, create all Swap instances for each node, which share the same backend
	for i := 0; i < nodeCount; i++ {
		params.swaps[i] = New(stores[i], keys[i], defaultBackend)
	}

	params.backend = defaultBackend
	return params, nil
}

// TestPingPongChequeSimulation just launches two nodes and sends each other
// messages which immediately crosses the PaymentThreshold and triggers cheques
// to each other. Checks that accounting and cheque handling works across multiple
// cheques and also if cheques are mutually sent
func TestPingPongChequeSimulation(t *testing.T) {
	nodeCount := 2
	// create the shared backend and params
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// initialize the simulation
	sim := simulation.NewInProc(newSimServiceMap(params))
	defer sim.Close()

	log.Info("Initializing")

	ctx, cancelSimRun := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelSimRun()

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("starting simulation...")

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		log.Info("simulation running")
		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		_, err = sim.WaitTillHealthy(ctx)
		if err != nil {
			t.Fatal(err)
		}

		p1 := sim.UpNodeIDs()[0]
		p2 := sim.UpNodeIDs()[1]

		maxCheques := 42

		item, ok := sim.NodeItem(p1, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		p1Svc := item.(*testService)

		peerItem, ok := sim.NodeItem(p2, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		p2Svc := peerItem.(*testService)

		p2Peer := p1Svc.peers[p2]
		p1Peer := p2Svc.peers[p1]

		for i := 0; i < maxCheques; i++ {
			if i%2 == 0 {
				p2Peer.Send(ctx, &testMsgBigPrice{})
			} else {
				p1Peer.Send(ctx, &testMsgBigPrice{})
			}
			time.Sleep(50 * time.Millisecond)
		}

		// we need to synchronize when we can actually go check that all values are ok
		// (all cheques arrived). Without it, specifically on CI (travis) the tests are flaky
		chequesArrived := make(chan struct{})

		// periodically check that all cheques have arrived
		go func() {
			var ch1, ch2 *Cheque
			for {
				time.Sleep(10 * time.Millisecond)
				select {
				case <-ctx.Done():
					return
				default:
				}
				p1Svc.swap.store.Get(receivedChequeKey(p2), &ch1)
				p2Svc.swap.store.Get(receivedChequeKey(p1), &ch2)
				if ch1 == nil || ch2 == nil {
					continue
				}
				// every peer gets maxCheques/2 messages, thus we can check that the CumulativePayout corresponds
				// (NOTE: DefaultPaymentThreshold + 1 is assumed to be the price for `testMsgBigPrice`)
				if ch1.CumulativePayout == uint64(maxCheques/2*(DefaultPaymentThreshold+1)) &&
					ch2.CumulativePayout == uint64(maxCheques/2*(DefaultPaymentThreshold+1)) {
					log.Debug("expected payout reached. going to check values now")
					close(chequesArrived)
					return
				}
			}
		}()

		log.Debug("waiting for cheque to arrive....")
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for cheques, aborting.")
		case <-chequesArrived:
		}
		log.Debug("all good.")

		ch1, ok := p2Svc.swap.getCheque(p1)
		if !ok {
			return errors.New("peer not found")
		}
		ch2, ok := p1Svc.swap.getCheque(p2)
		if !ok {
			return errors.New("peer not found")
		}

		expected := uint64(maxCheques / 2 * (DefaultPaymentThreshold + 1))
		if ch1.CumulativePayout != expected {
			return fmt.Errorf("expected cumulative payout to be %d, but is %d", expected, ch1.CumulativePayout)
		}
		if ch2.CumulativePayout != expected {
			return fmt.Errorf("expected cumulative payout to be %d, but is %d", expected, ch2.CumulativePayout)
		}

		return nil

	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	log.Info("Simulation ended")
}

// TestMultiChequeSimulation just launches two nodes, and sends multiple cheques
// to the same node; checks that accounting still works properly afterwards and that
// cheque cumulation values add up correctly
func TestMultiChequeSimulation(t *testing.T) {
	nodeCount := 2
	// create the shared backend and params
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// initialize the simulation
	sim := simulation.NewInProc(newSimServiceMap(params))
	defer sim.Close()

	log.Info("Initializing")

	ctx, cancelSimRun := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelSimRun()

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("starting simulation...")

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		log.Info("simulation running")
		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		_, err = sim.WaitTillHealthy(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// define the nodes
		debitor := sim.UpNodeIDs()[0]
		creditor := sim.UpNodeIDs()[1]
		// we will send just maxCheques number of cheques
		maxCheques := 6

		// get the testService for the debitor
		item, ok := sim.NodeItem(debitor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		debitorSvc := item.(*testService)

		// get the testService for the creditor
		peerItem, ok := sim.NodeItem(creditor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		creditorSvc := peerItem.(*testService)

		// the peer object used for sending
		creditorPeer := debitorSvc.peers[creditor]

		// send maxCheques number of cheques
		for i := 0; i < maxCheques; i++ {
			// use a price which will trigger a cheque each time
			creditorPeer.Send(ctx, &testMsgBigPrice{})
			// we need to sleep a bit in order to give time for the cheque to be processed
			time.Sleep(50 * time.Millisecond)
		}

		// we need some synchronization, or tests get flaky, especially on CI (travis)
		chequesArrived := make(chan struct{})

		// check periodically that the peer has all cheques
		go func() {
			for {
				time.Sleep(10 * time.Millisecond)
				select {
				case <-ctx.Done():
					return
				default:
				}
				var ch *Cheque
				creditorSvc.swap.store.Get(receivedChequeKey(debitor), &ch)
				if ch == nil {
					continue
				}
				// the peer should have a CumulativePayout correspondent to the amount of cheques emitted
				if ch.CumulativePayout == uint64(maxCheques*(DefaultPaymentThreshold+1)) {
					log.Debug("expected payout reached. going to check values now")
					close(chequesArrived)
				}
			}
		}()

		log.Debug("waiting for cheque to arrive....")
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for cheques, aborting.")
		case <-chequesArrived:
		}
		log.Debug("all good.")

		// check balances:
		b1, _ := debitorSvc.swap.getBalance(creditor)
		b2, _ := creditorSvc.swap.getBalance(debitor)

		if b1 != -b2 {
			return fmt.Errorf("Expected symmetric balances, but they are not: %d vs %d", b1, b2)
		}
		// check cheques
		var cheque1, cheque2 *Cheque
		if cheque1, ok = debitorSvc.swap.getCheque(creditor); !ok {
			return errors.New("expected cheques with creditor, but none found")
		}
		creditorSvc.swap.store.Get(receivedChequeKey(debitor), &cheque2)
		if cheque2 == nil {
			return errors.New("expected cheques with debitor, but none found")
		}

		// both cheques (at issuer and beneficiary) should have same cumulative value
		if cheque1.CumulativePayout != cheque2.CumulativePayout {
			return fmt.Errorf("Expected symmetric cheques payout, but they are not: %d vs %d", cheque1.CumulativePayout, cheque2.CumulativePayout)
		}

		// check also the actual expected amount
		expectedPayout := uint64(maxCheques * (DefaultPaymentThreshold + 1))

		if cheque2.CumulativePayout != expectedPayout {
			return fmt.Errorf("Expected %d in cumulative payout, got %d", expectedPayout, cheque1.CumulativePayout)
		}

		return nil

	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	log.Info("Simulation ended")
}

// TestSimpleSimulation starts 16 nodes, then in a simple round robin fashion sends messages to each other.
// Then checks that accounting is ok. It checks the actual amount of balances without any cheques sent,
// in order to verify that the most basic accounting works.
func TestSimpleSimulation(t *testing.T) {
	nodeCount := 16
	// create the shared backend and params
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// initialize the simulation
	sim := simulation.NewInProc(newSimServiceMap(params))
	defer sim.Close()

	log.Info("Initializing")

	ctx, cancelSimRun := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelSimRun()

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("starting simulation...")

	// setup a filter for all received messages
	// we count all received messages by any peer in order to know when the last
	// peer has actually received and processed the message
	msgs := sim.PeerEvents(
		context.Background(),
		sim.NodeIDs(),
		// Watch when bzz messages 1 and 4 are received.
		simulation.NewPeerEventsFilter().ReceivedMessages().Protocol("testSpec").MsgCode(0),
		simulation.NewPeerEventsFilter().ReceivedMessages().Protocol("testSpec").MsgCode(1),
	)

	// we don't want any cheques to be issued for this test, we only want to test accounting across nodes
	// for this we define a "global" maximum amount of messages to be sent;
	// this formula should ensure that we trigger enough messages but not enough to trigger cheques
	maxMsgs := (DefaultPaymentThreshold / params.maxMsgPrice) * (nodeCount - 1)
	// need some syncrhonization to make sure we wait enough before check all balances:
	// all messages should have been received
	allMessagesArrived := make(chan struct{})
	// count all messages received in the simulation
	recvCount := 0

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for m := range msgs {
			if m.Error != nil {
				log.Error("bzz message", "err", m.Error)
				continue
			}
			// received a message
			recvCount++
			// all messages have been received
			if recvCount == maxMsgs {
				close(allMessagesArrived)
				return
			}
		}
	}()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		log.Info("simulation running")
		disconnected := watchDisconnections(ctx, sim)
		defer func() {
			if err != nil && disconnected.bool() {
				err = errors.New("disconnect events received")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		_, err = sim.WaitTillHealthy(ctx)
		if err != nil {
			t.Fatal(err)
		}

		nodes := sim.UpNodeIDs()
		msgCount := 0

		// iterate all nodes, then send each other test messages
	ITER:
		for {
			for _, node := range nodes {
				for k, p := range nodes {
					// don't send to self
					if node == p {
						continue
					}
					if msgCount < maxMsgs {
						item, ok := sim.NodeItem(node, bucketKeySwap)
						if !ok {
							return errors.New("no swap in simulation bucket")
						}
						ts := item.(*testService)

						tp := ts.peers[p]
						if tp == nil {
							return errors.New("peer is nil")
						}
						// also alternate between Sender paid and Receiver paid messages
						if k%2 == 0 {
							tp.Send(ctx, &testMsgByReceiver{})
						} else {
							tp.Send(ctx, &testMsgBySender{})
						}
						msgCount++
					} else {
						break ITER
					}
				}
			}
		}

		// before we can check the balances, we need to wait a bit, as the last messages
		// may still be processed
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting for all messages to arrive, aborting")
		case <-allMessagesArrived:
		}
		log.Debug("all messages arrived")

		//now iterate again and check that every node has the same
		//balance with a peer as that peer with the same node,
		//but in inverted signs
		for _, node := range nodes {
			item, ok := sim.NodeItem(node, bucketKeySwap)
			if !ok {
				return errors.New("no swap in simulation bucket")
			}
			ts := item.(*testService)
			// for each node look up the peers
			for _, p := range nodes {
				// no need to check self
				if p == node {
					continue
				}

				peerItem, ok := sim.NodeItem(p, bucketKeySwap)
				if !ok {
					return errors.New("no swap in simulation bucket")
				}
				peerTs := peerItem.(*testService)

				// balance of the node with peer p
				nodeBalanceWithP, ok := ts.swap.getBalance(p)
				if !ok {
					return fmt.Errorf("expected balance for peer %v to be found, but not found", p)
				}
				// balance of the peer with node
				pBalanceWithNode, ok := peerTs.swap.getBalance(node)
				if !ok {
					return fmt.Errorf("expected counter balance for node %v to be found, but not found", node)
				}
				if nodeBalanceWithP != -pBalanceWithNode {
					return fmt.Errorf("Expected symmetric balances, but they are not: %d vs %d", nodeBalanceWithP, pBalanceWithNode)
				}
			}
		}

		return nil

	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	log.Info("Simulation ended")
}

// watchDisconnections receives simulation peer events in a new goroutine and sets atomic value
// disconnected to true in case of a disconnect event.
func watchDisconnections(ctx context.Context, sim *simulation.Simulation) (disconnected *boolean) {
	log.Debug("Watching for disconnections")
	disconnections := sim.PeerEvents(
		ctx,
		sim.NodeIDs(),
		simulation.NewPeerEventsFilter().Drop(),
	)
	disconnected = new(boolean)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d := <-disconnections:
				if d.Error != nil {
					log.Error("peer drop event error", "node", d.NodeID, "peer", d.PeerID, "err", d.Error)
				} else {
					log.Error("peer drop", "node", d.NodeID, "peer", d.PeerID)
				}
				disconnected.set(true)
			}
		}
	}()
	return disconnected
}

func (ts *testService) Protocols() []p2p.Protocol {
	spec := newTestSpec()
	return []p2p.Protocol{
		{
			Name:    spec.Name,
			Version: spec.Version,
			Length:  spec.Length(),
			Run:     ts.runProtocol,
		},
		{
			Name:    Spec.Name,
			Version: Spec.Version,
			Length:  Spec.Length(),
			Run:     ts.swap.run,
		},
	}
}

// APIs retrieves the list of RPC descriptors the service provides
func (ts *testService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "testAccounted",
			Version:   "1.0",
			Service:   ts,
			Public:    false,
		},
	}
}

// runProtocol for the test spec
func (ts *testService) runProtocol(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, ts.spec)
	tp := &testPeer{Peer: peer}
	ts.peers[tp.ID()] = tp
	return peer.Run(tp.handleMsg)
}

// Start is called after all services have been constructed and the networking
// layer was also initialized to spawn any goroutines required by the service.
func (ts *testService) Start(server *p2p.Server) error {
	return nil
}

// Stop terminates all goroutines belonging to the service, blocking until they
// are all terminated.
func (ts *testService) Stop() error {
	return nil
}

// boolean is used to concurrently set
// and read a boolean value.
type boolean struct {
	v  bool
	mu sync.RWMutex
}

// set sets the value.
func (b *boolean) set(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.v = v
}

// bool reads the value.
func (b *boolean) bool() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.v
}
