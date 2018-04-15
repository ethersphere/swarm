package simulations

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
}

func TestNodeUpAndConn(t *testing.T) {

	var quitC []chan struct{}
	defer func() {
		for _, q := range quitC {
			if q != nil {
				close(q)
			}
		}
	}()

	// create simulation network with 20 testService nodes
	adapter := adapters.NewSimAdapter(adapters.Services{
		"test": newTestService,
	})
	network := NewNetwork(adapter, &NetworkConfig{
		DefaultService: "test",
	})
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
	sub := network.Events().Subscribe(events)
	defer sub.Unsubscribe()
	quitC = append(quitC, make(chan struct{}))

	action := func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				select {
				case ev := <-events:
					if ev == nil {
						panic("got nil event")
					}
					if ev.Type == EventTypeNode {
						if ev.Node.Up {
							log.Info("got node up event", "event", ev, "node", ev.Node.Config.ID)
							trigger <- ev.Node.Config.ID
						}
					}

				case <-quitC:
					log.Warn("got quit action 1")
					return
				}

			}
		}(quitC[len(quitC)-1])
		go func() {
			for _, n := range ids {
				if err := network.Start(n); err != nil {
					t.Fatalf("error starting node: %s", err)
				}
				log.Info("network start returned", "node", n)
			}
		}()
		return nil
	}

	check := func(ctx context.Context, nodeId discover.NodeID) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		log.Info("trigger expect up", "node", nodeId)
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
	close(quitC[len(quitC)-1])
	quitC[len(quitC)-1] = make(chan struct{})

	action = func(ctx context.Context) error {
		go func(quitC chan struct{}) {
			for {
				select {
				case ev := <-events:
					if ev.Type == EventTypeConn {
						if ev.Conn.Up {
							log.Info(fmt.Sprintf("got conn up event %v", ev))
							trigger <- ev.Conn.One
						}
					}
				case <-quitC:
					return
				}
			}
		}(quitC[len(quitC)-1])
		go func() {
			for i := range ids {
				j := i - 1
				if i == 0 {
					j = len(ids) - 1
				}

				if err := network.Connect(ids[i], ids[j]); err != nil {
					t.Fatalf("error connecting nodes %x => %x: %s", ids[i], ids[j], err)
				}
				log.Info("network connect returned", "one", ids[i], "other", ids[j])
			}
		}()
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
		log.Info("trigger expect conn", "node", id)
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
	log.Info("done")
}
