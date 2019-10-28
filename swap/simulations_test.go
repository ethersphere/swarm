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

package swap

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	contractFactory "github.com/ethersphere/go-sw3/contracts-v0-1-1/simpleswapfactory"
	cswap "github.com/ethersphere/swarm/contracts/swap"
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
// swapSimulationParams allows to avoid global variables for the test
type swapSimulationParams struct {
	swaps       map[int]*Swap
	dirs        map[int]string
	count       int
	maxMsgPrice int
	minMsgPrice int
	nodeCount   int
	backend     *swapTestBackend
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

func (m *testMsgBySender) Price() *protocols.Price {
	return &protocols.Price{
		Value:   1000, // arbitrary price for now
		PerByte: true,
		Payer:   protocols.Sender,
	}
}

func (m *testMsgByReceiver) Price() *protocols.Price {
	return &protocols.Price{
		Value:   100, // arbitrary price for now
		PerByte: false,
		Payer:   protocols.Receiver,
	}
}

func (m *testMsgBigPrice) Price() *protocols.Price {
	return &protocols.Price{
		Value:   DefaultPaymentThreshold + 1,
		PerByte: false,
		Payer:   protocols.Sender,
	}
}

// testService encapsulates objects needed for the simulation
type testService struct {
	lock  sync.Mutex
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
		// we use a swap service
		"swap": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			// every simulation node has an instance of a `testService`
			ts := newTestService()
			// balance is the interface for `NewAccounting`; it is a Swap
			balance := params.swaps[params.count]
			dir := params.dirs[params.count]
			// every node is a different instance of a Swap and gets a different entry in the map
			params.count++
			ts.spec = newTestSpec()
			// create the accounting instance and assign to the spec
			ts.spec.Hook = protocols.NewAccounting(balance)
			ts.swap = balance
			// deploy the accounting to the `SimulatedBackend`
			err = testDeploy(context.Background(), balance)
			if err != nil {
				return nil, nil, err
			}

			cleanup = func() {
				ts.swap.store.Close()
				os.RemoveAll(dir)
			}

			return ts, cleanup, nil
		},
	}
	return simServiceMap
}

// newSharedBackendSwaps pre-loads each simulated node account with "funds"
// so that later in the simulation all operations have sufficient gas
func newSharedBackendSwaps(t *testing.T, nodeCount int) (*swapSimulationParams, error) {
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
		alloc[addrs[i]] = core.GenesisAccount{Balance: big.NewInt(10000 * int64(RetrieveRequestPrice))}
		dir, err := ioutil.TempDir("", fmt.Sprintf("swap_test_store_%x", addrs[i].Hex()))
		if err != nil {
			return nil, err
		}
		stateStore, err := state.NewDBStore(dir)
		if err != nil {
			return nil, err
		}
		params.dirs[i] = dir
		stores[i] = stateStore
	}
	// then create the single SimulatedBackend
	gasLimit := uint64(8000000000)
	defaultBackend := backends.NewSimulatedBackend(alloc, gasLimit)
	defaultBackend.Commit()

	factoryAddress, _, _, err := contractFactory.DeploySimpleSwapFactory(bind.NewKeyedTransactor(keys[0]), defaultBackend)
	if err != nil {
		t.Fatalf("Error while deploying factory: %v", err)
	}
	defaultBackend.Commit()

	testBackend := &swapTestBackend{SimulatedBackend: defaultBackend, factoryAddress: factoryAddress}
	// finally, create all Swap instances for each node, which share the same backend
	var owner *Owner
	defParams := newDefaultParams(t)
	for i := 0; i < nodeCount; i++ {
		owner = createOwner(keys[i])
		factory, err := cswap.FactoryAt(testBackend.factoryAddress, testBackend)
		if err != nil {
			t.Fatal(err)
		}
		params.swaps[i] = newSwapInstance(stores[i], owner, testBackend, big.NewInt(10), defParams, factory)
	}

	params.backend = testBackend
	return params, nil
}

// TestPingPongChequeSimulation just launches two nodes and sends each other
// messages which immediately crosses the PaymentThreshold and triggers cheques
// to each other. Checks that accounting and cheque handling works across multiple
// cheques and also if cheques are mutually sent
func TestPingPongChequeSimulation(t *testing.T) {
	nodeCount := 2
	// create the shared backend and params
	params, err := newSharedBackendSwaps(t, nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	params.backend.cashDone = make(chan struct{}, 1)
	defer close(params.backend.cashDone)
	// initialize the simulation
	sim := simulation.NewBzzInProc(newSimServiceMap(params), false)
	defer sim.Close()

	log.Info("Initializing")

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	p1 := sim.UpNodeIDs()[0]
	p2 := sim.UpNodeIDs()[1]
	ts1 := sim.Service("swap", p1).(*testService)
	ts2 := sim.Service("swap", p2).(*testService)

	var ts1Len, ts2Len int
	timeout := time.After(10 * time.Second)

	for {
		// let's always be nice and allow a time out to be catched
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for all swap peer connections to be established")
		default:
		}
		// the node has all other peers in its peer list
		ts1.swap.peersLock.Lock()
		ts1Len = len(ts1.swap.peers)
		ts1.swap.peersLock.Unlock()

		ts2.swap.peersLock.Lock()
		ts2Len = len(ts2.swap.peers)
		ts2.swap.peersLock.Unlock()

		if ts1Len == 1 && ts2Len == 1 {
			break
		}
		// don't overheat the CPU...
		time.Sleep(5 * time.Millisecond)
	}

	maxCheques := 42

	ts1.lock.Lock()
	p2Peer := ts1.peers[p2]
	ts1.lock.Unlock()

	ts2.lock.Lock()
	p1Peer := ts2.peers[p1]
	ts2.lock.Unlock()

	for i := 0; i < maxCheques; i++ {
		if i%2 == 0 {
			if err := p2Peer.Send(context.Background(), &testMsgBigPrice{}); err != nil {
				t.Fatal(err)
			}
			if err := waitForChequeProcessed(t, ts2); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := p1Peer.Send(context.Background(), &testMsgBigPrice{}); err != nil {
				t.Fatal(err)
			}
			if err := waitForChequeProcessed(t, ts1); err != nil {
				t.Fatal(err)
			}
		}
	}

	ch1, err := ts2.swap.loadLastReceivedCheque(p1)
	if err != nil {
		t.Fatal(err)
	}
	ch2, err := ts1.swap.loadLastReceivedCheque(p2)
	if err != nil {
		t.Fatal(err)
	}

	expected := uint64(maxCheques) / 2 * (DefaultPaymentThreshold + 1)
	if ch1.CumulativePayout != expected {
		t.Fatalf("expected cumulative payout to be %d, but is %d", expected, ch1.CumulativePayout)
	}
	if ch2.CumulativePayout != expected {
		t.Fatalf("expected cumulative payout to be %d, but is %d", expected, ch2.CumulativePayout)
	}

	log.Info("Simulation ended")
}

// TestMultiChequeSimulation just launches two nodes, a creditor and a debitor.
// The debitor is the one owing to the creditor, so the debitor is the one sending cheques
// It sends multiple cheques in a sequence to the same node.
// The test checks that accounting still works properly afterwards and that
// cheque cumulation values add up correctly
func TestMultiChequeSimulation(t *testing.T) {
	nodeCount := 2
	// create the shared backend and params
	params, err := newSharedBackendSwaps(t, nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	params.backend.cashDone = make(chan struct{}, 1)
	defer close(params.backend.cashDone)
	// initialize the simulation
	sim := simulation.NewBzzInProc(newSimServiceMap(params), false)
	defer sim.Close()

	log.Info("Initializing")

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	// define the nodes
	debitor := sim.UpNodeIDs()[0]
	creditor := sim.UpNodeIDs()[1]
	// get the testService for the debitor
	debitorSvc := sim.Service("swap", debitor).(*testService)
	// get the testService for the creditor
	creditorSvc := sim.Service("swap", creditor).(*testService)

	var debLen, credLen int
	timeout := time.After(10 * time.Second)
	for {
		// let's always be nice and allow a time out to be catched
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for all swap peer connections to be established")
		default:
		}
		// the node has all other peers in its peer list
		debitorSvc.swap.peersLock.Lock()
		debLen = len(debitorSvc.swap.peers)
		debitorSvc.swap.peersLock.Unlock()

		creditorSvc.swap.peersLock.Lock()
		credLen = len(creditorSvc.swap.peers)
		creditorSvc.swap.peersLock.Unlock()

		if debLen == 1 && credLen == 1 {
			break
		}
		// don't overheat the CPU...
		time.Sleep(5 * time.Millisecond)
	}

	// we will send just maxCheques number of cheques
	maxCheques := 10

	// the peer object used for sending
	debitorSvc.lock.Lock()
	creditorPeer := debitorSvc.peers[creditor]
	debitorSvc.lock.Unlock()

	// send maxCheques number of cheques
	for i := 0; i < maxCheques; i++ {
		// use a price which will trigger a cheque each time
		if err := creditorPeer.Send(context.Background(), &testMsgBigPrice{}); err != nil {
			t.Fatal(err)
		}
		// we need to wait a bit in order to give time for the cheque to be processed
		if err := waitForChequeProcessed(t, creditorSvc); err != nil {
			t.Fatal(err)
		}
	}

	// check balances:
	b1, err := debitorSvc.swap.loadBalance(creditor)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := creditorSvc.swap.loadBalance(debitor)
	if err != nil {
		t.Fatal(err)
	}

	if b1 != -b2 {
		t.Fatalf("Expected symmetric balances, but they are not: %d vs %d", b1, b2)
	}
	// check cheques
	var cheque1, cheque2 *Cheque
	if cheque1, err = debitorSvc.swap.loadLastSentCheque(creditor); err != nil {
		t.Fatalf("expected cheques with creditor, but none found")
	}
	if cheque2, err = creditorSvc.swap.loadLastReceivedCheque(debitor); err != nil {
		t.Fatalf("expected cheques with debitor, but none found")
	}

	// both cheques (at issuer and beneficiary) should have same cumulative value
	if cheque1.CumulativePayout != cheque2.CumulativePayout {
		t.Fatalf("Expected symmetric cheques payout, but they are not: %d vs %d", cheque1.CumulativePayout, cheque2.CumulativePayout)
	}

	// check also the actual expected amount
	expectedPayout := uint64(maxCheques) * (DefaultPaymentThreshold + 1)

	if cheque2.CumulativePayout != expectedPayout {
		t.Fatalf("Expected %d in cumulative payout, got %d", expectedPayout, cheque1.CumulativePayout)
	}

	log.Info("Simulation ended")
}

// TestBasicSwapSimulation starts 16 nodes, then in a simple round robin fashion sends messages to each other.
// Then checks that accounting is ok. It checks the actual amount of balances without any cheques sent,
// in order to verify that the most basic accounting works.
func TestBasicSwapSimulation(t *testing.T) {
	nodeCount := 16
	// create the shared backend and params
	params, err := newSharedBackendSwaps(t, nodeCount)
	if err != nil {
		t.Fatal(err)
	}
	// cleanup backend
	defer params.backend.Close()

	// initialize the simulation
	sim := simulation.NewBzzInProc(newSimServiceMap(params), false)
	defer sim.Close()

	log.Info("Initializing")

	_, err = sim.AddNodesAndConnectFull(nodeCount)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("Wait for all connections to be established")

	simulations.VerifyFull(t, sim.Net, sim.NodeIDs())

	log.Info("starting simulation...")

	// we don't want any cheques to be issued for this test, we only want to test accounting across nodes
	// for this we define a "global" maximum amount of messages to be sent;
	maxMsgs := 1500

	// need some synchronization to make sure we wait enough before checking all balances:
	// all messages should have been received, otherwise there may be some imbalances!
	allMessagesArrived := make(chan struct{})

	metricsReg := metrics.AccountingRegistry
	cter := metricsReg.Get("account.msg.credit")
	counter := cter.(metrics.Counter)
	counter.Clear()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		for {
			maxMsgsInt64 := int64(maxMsgs)
			select {
			case <-ctx.Done():
				return
			default:
			}
			// all messages have been received
			if counter.Count() == maxMsgsInt64 {
				close(allMessagesArrived)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	result := sim.Run(ctx, func(ctx context.Context, sim *simulation.Simulation) (err error) {
		log.Info("simulation running")

		nodes := sim.UpNodeIDs()
		msgCount := 0
		// unfortunately, before running the actual simulation, we need an additional check (...).
		// If we start sending right away, it can happen that devp2p did **not yet finish connecting swap peers**
		// (verified through multiple runs). This would then fail the test because on Swap.Add the peer is not (yet) found...
		// Thus this iteration here makes sure that all swap peers actually have been added on the Swap protocol as well.
	ALL_SWAP_PEERS:
		for _, node := range nodes {
			for {
				// let's always be nice and allow a time out to be catched
				select {
				case <-ctx.Done():
					return errors.New("Timed out waiting for all swap peer connections to be established")
				default:
				}
				ts := sim.Service("swap", node).(*testService)
				// the node has all other peers in its peer list
				ts.swap.peersLock.Lock()
				tsLen := len(ts.swap.peers)
				ts.swap.peersLock.Unlock()
				if tsLen == nodeCount-1 {
					// so let's take the next node
					continue ALL_SWAP_PEERS
				}
				// don't overheat the CPU...
				time.Sleep(5 * time.Millisecond)
			}
		}

		// iterate all nodes, then send each other test messages
	ITER:
		for {
			for _, node := range nodes {
				ts := sim.Service("swap", node).(*testService)
				for k, p := range nodes {
					// don't send to self
					if node == p {
						continue
					}
					if msgCount < maxMsgs {
						ts.lock.Lock()
						tp := ts.peers[p]
						ts.lock.Unlock()
						if tp == nil {
							return errors.New("peer is nil")
						}
						// also alternate between Sender paid and Receiver paid messages
						if k%2 == 0 {
							err := tp.Send(context.Background(), &testMsgByReceiver{})
							if err != nil {
								return err
							}
						} else {
							err := tp.Send(context.Background(), &testMsgBySender{})
							if err != nil {
								return err
							}
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
			nodeTs := sim.Service("swap", node).(*testService)
			// for each node look up the peers
			for _, p := range nodes {
				// no need to check self
				if p == node {
					continue
				}

				peerTs := sim.Service("swap", p).(*testService)

				// balance of the node with peer p
				nodeBalanceWithP, err := nodeTs.swap.loadBalance(p)
				if err != nil {
					return fmt.Errorf("expected balance for peer %v to be found, but not found", p)
				}
				// balance of the peer with node
				pBalanceWithNode, err := peerTs.swap.loadBalance(node)
				if err != nil {
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
	counter.Clear()
	log.Info("Simulation ended")
}

func waitForChequeProcessed(t *testing.T, ts *testService) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	backend := ts.swap.backend.(*swapTestBackend)

	select {
	case <-ctx.Done():
		t.Fatal("timed out waiting for cheque to be processed")
	case <-backend.cashDone:
	}
	return nil
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
	ts.lock.Lock()
	ts.peers[tp.ID()] = tp
	ts.lock.Unlock()
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
