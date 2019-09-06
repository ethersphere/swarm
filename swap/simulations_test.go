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

var bucketKeySwap = simulation.BucketKey("swap")

func newSimServiceMap(params *swapSimulationParams) map[string]simulation.ServiceFunc {
	simServiceMap := map[string]simulation.ServiceFunc{
		"bzz": func(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
			addr := network.NewAddr(ctx.Config.Node())
			hp := network.NewHiveParams()
			hp.Discovery = false
			//assign the network ID
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			kad := network.NewKademlia(addr.Over(), network.NewKadParams())
			return network.NewBzz(config, kad, nil, nil, nil), nil, nil
		},
		"swap": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {

			ts := newTestService()

			balance := params.swaps[params.count]
			params.count++
			prices := &testPrices{}
			prices.newTestPriceMatrix()
			ts.spec = newTestSpec()
			ts.spec.Hook = protocols.NewAccounting(balance, prices)
			ts.swap = balance
			testDeploy(context.Background(), balance.backend, balance)
			balance.backend.(*backends.SimulatedBackend).Commit()

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

type swapSimulationParams struct {
	swaps       map[int]*Swap
	dirs        map[int]string
	count       int
	maxMsgPrice int
	minMsgPrice int
	nodeCount   int
}

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

	for i := 0; i < nodeCount; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		keys[i] = key
		addrs[i] = crypto.PubkeyToAddress(key.PublicKey)
		alloc[addrs[i]] = core.GenesisAccount{Balance: big.NewInt(1000000000)}
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
	gasLimit := uint64(8000000)
	defaultBackend := backends.NewSimulatedBackend(alloc, gasLimit)
	for i := 0; i < nodeCount; i++ {
		params.swaps[i] = New(stores[i], keys[i], common.Address{}, defaultBackend)
	}

	return params, nil
}

func TestPingPongChequeSimulation(t *testing.T) {
	nodeCount := 2
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
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

		debitor := sim.UpNodeIDs()[0]
		creditor := sim.UpNodeIDs()[1]

		maxCheques := 42

		item, ok := sim.NodeItem(debitor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		debitorSvc := item.(*testService)

		peerItem, ok := sim.NodeItem(creditor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		creditorSvc := peerItem.(*testService)

		creditorPeer := debitorSvc.peers[creditor]
		debitorPeer := creditorSvc.peers[debitor]

		for i := 0; i < maxCheques; i++ {
			if i%2 == 0 {
				creditorPeer.Send(ctx, &testMsgBigPrice{})
			} else {
				debitorPeer.Send(ctx, &testMsgBigPrice{})
			}
			time.Sleep(50 * time.Millisecond)
		}

		fmt.Println(creditorSvc.swap.getBalance(debitor))
		fmt.Println(debitorSvc.swap.getBalance(creditor))
		ch1, ok := creditorSvc.swap.getCheque(debitor)
		if err != nil {
			return errors.New("peer not found")
		}
		fmt.Println(ch1.CumulativePayout)
		ch2, ok := debitorSvc.swap.getCheque(creditor)
		if err != nil {
			return errors.New("peer not found")
		}
		fmt.Println(ch2.CumulativePayout)

		return nil

	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	log.Info("Simulation ended")
}

func TestMultiChequeSimulation(t *testing.T) {
	nodeCount := 2
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
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

		debitor := sim.UpNodeIDs()[0]
		creditor := sim.UpNodeIDs()[1]
		maxCheques := 6

		item, ok := sim.NodeItem(debitor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		debitorSvc := item.(*testService)

		peerItem, ok := sim.NodeItem(creditor, bucketKeySwap)
		if !ok {
			return errors.New("no swap in simulation bucket")
		}
		creditorSvc := peerItem.(*testService)

		creditorPeer := debitorSvc.peers[creditor]

		for i := 0; i < maxCheques; i++ {
			creditorPeer.Send(ctx, &testMsgBigPrice{})
			time.Sleep(100 * time.Millisecond)
		}

		b1, _ := debitorSvc.swap.getBalance(creditor)
		b2, _ := creditorSvc.swap.getBalance(debitor)

		if b1 != -b2 {
			return fmt.Errorf("Expected symmetric balances, but they are not: %d vs %d", b1, b2)
		}

		var cheque1, cheque2 *Cheque
		if cheque1, ok = debitorSvc.swap.getCheque(creditor); !ok {
			return errors.New("expected cheques with creditor, but none found")
		}
		creditorSvc.swap.store.Get(receivedChequeKey(debitor), &cheque2)
		if cheque2 == nil {
			return errors.New("expected cheques with debitor, but none found")
		}

		if cheque1.CumulativePayout != cheque2.CumulativePayout {
			return fmt.Errorf("Expected symmetric cheques payout, but they are not: %d vs %d", cheque1.CumulativePayout, cheque2.CumulativePayout)
		}

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

func TestSimpleSimulation(t *testing.T) {

	nodeCount := 16
	params, err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		t.Fatal(err)
	}
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

		nodes := sim.UpNodeIDs()
		maxMsgs := (DefaultPaymentThreshold / params.maxMsgPrice) * (nodeCount - 1)
		msgCount := 0

	ITER:
		for {
			for _, node := range nodes {
				for k, p := range nodes {
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

		time.Sleep(1 * time.Second)

		for i, node := range nodes {

			//now iterate
			//and check that every node k has the same
			//balance with a peer as that peer with the node,
			//but in inverted signs

			//iterate the map
			p := i + 1
			if i == len(nodes)-1 {
				p = 0
			}
			item, ok := sim.NodeItem(node, bucketKeySwap)
			if !ok {
				return errors.New("no swap in simulation bucket")
			}
			ts := item.(*testService)

			peerItem, ok := sim.NodeItem(nodes[p], bucketKeySwap)
			if !ok {
				return errors.New("no swap in simulation bucket")
			}
			peerTs := peerItem.(*testService)

			nodeBalanceWithP, ok := ts.swap.getBalance(nodes[p])
			if !ok {
				return fmt.Errorf("expected balance for peer %v to be found, but not found", nodes[p])
			}
			pBalanceWithNode, ok := peerTs.swap.getBalance(node)
			if !ok {
				return fmt.Errorf("expected counter balance for node %v to be found, but not found", node)
			}
			if nodeBalanceWithP != -pBalanceWithNode {
				return fmt.Errorf("Expected symmetric balances, but they are not: %d vs %d", nodeBalanceWithP, pBalanceWithNode)
			}
		}

		return nil

	})

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	log.Info("Simulation ended")
}

type testMsgBySender struct{}
type testMsgByReceiver struct{}
type testMsgBigPrice struct{}

type testPeer struct {
	*protocols.Peer
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

type testPrices struct {
	priceMatrix map[reflect.Type]*protocols.Price
}

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

func (tp *testPrices) Price(msg interface{}) *protocols.Price {
	t := reflect.TypeOf(msg).Elem()
	return tp.priceMatrix[t]
}

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

func (ts *testService) runProtocol(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, ts.spec)
	tp := &testPeer{Peer: peer}
	ts.peers[tp.ID()] = tp
	//peer.Send(context.Background(), &testMsgByReceiver{})
	return peer.Run(tp.handleMsg)
}

func (tp *testPeer) handleMsg(ctx context.Context, msg interface{}) error {

	switch msg.(type) {

	case *testMsgBySender:

	case *testMsgByReceiver:
	}
	return nil
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
