package simulations

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
}

type dispatcher struct {
	network *Network
	events  map[discover.NodeID]chan *Event
	startC  chan error
	quitC   chan struct{}
}

func newDispatcher(network *Network, quitC chan struct{}) *dispatcher {
	return &dispatcher{
		network: network,
		startC:  make(chan error),
		quitC:   quitC,
		events:  make(map[discover.NodeID]chan *Event),
	}
}

func (self *dispatcher) run() {
	events := make(chan *Event)
	sub := self.network.Events().Subscribe(events)
	defer sub.Unsubscribe()

	wg := sync.WaitGroup{}
	nodes := self.network.GetNodes()
	wg.Add(1)
	for _, n := range nodes {
		self.events[n.ID()] = make(chan *Event)
	}
	go func() {
		for {
			select {
			case ev := <-events:
				if ev == nil {
					log.Warn("dispatcher got nil event")
					wg.Done()
					return
				}
				log.Warn("dispatcher event", "event", ev.Type)
				if ev.Type == "msg" {
					continue
				} else if ev.Type == EventTypeConn {
					self.events[ev.Conn.One] <- ev
				} else if ev.Type == EventTypeNode {
					self.events[ev.Node.Config.ID] <- ev
				}
			}
		}
	}()
	self.startC <- nil
	wg.Wait()
	<-self.quitC
	for _, n := range nodes {
		close(self.events[n.ID()])
	}
}

func TestNodeUpAndConn(t *testing.T) {
	// create simulation network with 20 testService nodes
	adapter := adapters.NewSimAdapter(adapters.Services{
		"test": newTestService,
	})
	network := NewNetwork(adapter, &NetworkConfig{
		DefaultService: "test",
	})
	defer network.Shutdown()
	nodeCount := 3
	ids := make([]discover.NodeID, nodeCount)

	for i := 0; i < nodeCount; i++ {
		conf := adapters.RandomNodeConfig()
		node, err := network.NewNodeWithConfig(conf)
		ids[i] = node.ID()
		if err != nil {
			t.Fatalf("error creating node: %s", err)
		}
	}

	trigger := make(chan discover.NodeID)
	events := make(chan *Event)
	sub := self.network.Events().Subscribe(events)
	defer sub.Unsubscribe()

	action := func(ctx context.Context) error {
		go func() {
			for {
				select {
				case ev := <-events:
					if ev == nil {
						log.Warn("got nil event", "node", n)
						return
					}
					if ev.Type == EventTypeNode {
						if ev.Node.Up {
							log.Info(fmt.Sprintf("got node up event %v", ev))
							trigger <- ev.Node.Config.ID
							return
						}
					}
				}
			}
		}()
		for _, n := range ids {
			if err := network.Start(n); err != nil {
				t.Fatalf("error starting node: %s", err)
			}
			log.Info("network start returned", "node", n)
		}
		return nil
	}

	check := func(ctx context.Context, nodeId discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("check up", "node", nodeId)
		return true, nil
	}

	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := NewSimulation(network).Run(ctx, &Step{
		Action:  action,
		Trigger: trigger,
		Expect: &Expectation{
			Nodes: ids,
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatalf("simulation failed: %s", result.Error)
	}

	action = func(ctx context.Context) error {
		go func() {
			for {
				select {
				case ev := <-events:
					if ev.Type == EventTypeConn {
						if ev.Conn.Up {
							log.Info(fmt.Sprintf("got conn up event %v", ev))
							//if (ev.Conn.One == ids[i] && ev.Conn.Other == ids[j]) || (ev.Conn.One == ids[j] && ev.Conn.Other == ids[i]) {
							//if ev.Conn.One == ids[i] && ev.Conn.Other == ids[j] {
							trigger <- ev.Conn.One
						}
					}
				case <-quitC:
					return
				}
			}
		}()
		for i, n := range ids {
			j := i - 1
			if i == 0 {
				j = len(ids) - 1
			}

			if err := network.Connect(ids[i], ids[j]); err != nil {
				t.Fatalf("error connecting nodes %x => %x: %s", ids[i], ids[j], err)
			}
			log.Info("network connect returned", "one", ids[i], "other", ids[j])
		}
		return nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	check = func(ctx context.Context, id discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("trigger expect", "node", id)
		return true, nil
	}
	result = NewSimulation(network).Run(ctx, &Step{
		Action:  action,
		Trigger: trigger,
		Expect: &Expectation{
			Nodes: ids,
			Check: check,
		},
	})
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
