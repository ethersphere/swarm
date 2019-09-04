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

func init() {
	err := newSharedBackendSwaps(nodeCount)
	if err != nil {
		fmt.Println(err)
	}
}

var simServiceMap = map[string]simulation.ServiceFunc{
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

		balance := swaps[count]
		count++
		prices := &testPrices{}
		prices.newTestPriceMatrix()
		ts.spec = newTestSpec()
		ts.spec.Hook = protocols.NewAccounting(balance, prices)
		ts.swap = balance
		testDeploy(context.Background(), balance.backend, balance)
		balance.backend.(*backends.SimulatedBackend).Commit()

		bucket.Store(bucketKeySwap, ts)

		cleanup = func() {
			//os.RemoveAll(dirs[count])
		}

		return ts, cleanup, nil
	},
}

var swaps map[int]*Swap
var dirs map[int]string
var count int

var nodeCount = 4

func newSharedBackendSwaps(nodeCount int) error {
	swaps = make(map[int]*Swap)
	dirs = make(map[int]string)
	keys := make(map[int]*ecdsa.PrivateKey)
	addrs := make(map[int]common.Address)
	alloc := core.GenesisAlloc{}
	stores := make(map[int]*state.DBStore)

	for i := 0; i < nodeCount; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			return err
		}
		keys[i] = key
		addrs[i] = crypto.PubkeyToAddress(key.PublicKey)
		alloc[addrs[i]] = core.GenesisAccount{Balance: big.NewInt(1000000000)}
		dir, err := ioutil.TempDir("", fmt.Sprintf("swap_test_store_%d", i))
		if err != nil {
			return err
		}
		stateStore, err2 := state.NewDBStore(dir)
		if err2 != nil {
			return err
		}
		dirs[i] = dir
		stores[i] = stateStore
	}
	gasLimit := uint64(8000000)
	defaultBackend := backends.NewSimulatedBackend(alloc, gasLimit)
	for i := 0; i < nodeCount; i++ {
		swaps[i] = New(stores[i], keys[i], common.Address{}, defaultBackend)
	}

	return nil

}

func TestSimpleSimulation(t *testing.T) {

	sim := simulation.NewInProc(simServiceMap)
	defer sim.Close()

	log.Info("Initializing")

	ctx, cancelSimRun := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelSimRun()

	_, err := sim.AddNodesAndConnectFull(nodeCount)
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
		ill, err := sim.WaitTillHealthy(ctx)
		if err != nil {
			// inspect the latest detected not healthy kademlias
			for id, kad := range ill {
				fmt.Println("Node", id)
				fmt.Println(kad.String())
			}
			// handle error...
			t.Fatal(err)
		}

		pivot := sim.UpNodeIDs()[0]
		item, ok := sim.NodeItem(pivot, bucketKeySwap)
		if !ok {
			return errors.New("no store in simulation bucket")
		}
		ts := item.(*testService)

		p := sim.UpNodeIDs()[1]
		tp := ts.peers[p]
		fmt.Println(ts.swap.peers)
		for {
			if ts.swap.balances[p] > -DefaultPaymentThreshold {
				tp.Send(ctx, &testMsgBySender{})
			} else {
				break
			}
			fmt.Println(ts.swap.balances[p])
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
	fmt.Println("run protocol")
	peer := protocols.NewPeer(p, rw, ts.spec)
	tp := &testPeer{Peer: peer}
	ts.peers[tp.ID()] = tp
	peer.Send(context.Background(), &testMsgByReceiver{})
	return peer.Run(tp.handleMsg)
}

func (tp *testPeer) handleMsg(ctx context.Context, msg interface{}) error {

	switch msg.(type) {

	case *testMsgBySender:
		fmt.Println("testMsgBySender")
		// go tp.Send(context.Background(), &testMsgByReceiver{})

	case *testMsgByReceiver:
		fmt.Println("testMsgByReceiver")
		//go tp.Send(context.Background(), &testMsgBySender{})
	}
	return nil
}

// Start is called after all services have been constructed and the networking
// layer was also initialized to spawn any goroutines required by the service.
func (ts *testService) Start(server *p2p.Server) error {
	fmt.Println("starting testService")
	return nil
}

// Stop terminates all goroutines belonging to the service, blocking until they
// are all terminated.
func (ts *testService) Stop() error {
	fmt.Println("stopping testService")
	return nil
}
