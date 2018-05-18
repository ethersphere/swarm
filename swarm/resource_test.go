package swarm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/storage/resource"
)

const (
	resourceName      = "foo.eth"
	resourceFrequency = 2
)

func TestResourceNotifyWithSwarm(t *testing.T) {
	dir, err := ioutil.TempDir("", "swarm-resource-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	swarms := make(map[discover.NodeID]*Swarm)
	services := map[string]adapters.ServiceFunc{
		"swarm": func(ctx *adapters.ServiceContext) (node.Service, error) {
			config := api.NewConfig()

			dir, err := ioutil.TempDir(dir, "node")
			if err != nil {
				return nil, err
			}

			config.Path = dir

			privkey, err := crypto.GenerateKey()
			if err != nil {
				return nil, err
			}

			config.Init(privkey)
			s, err := NewSwarm(nil, nil, config, nil)
			if err != nil {
				return nil, err
			}
			log.Info("new swarm", "bzzKey", config.BzzKey, "baseAddr", fmt.Sprintf("%x", s.bzz.BaseAddr()))
			swarms[ctx.Config.ID] = s
			return s, nil
		},
	}

	a := adapters.NewSimAdapter(services)
	net := simulations.NewNetwork(a, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "swarm",
	})
	defer net.Shutdown()

	l_nodeconf := adapters.RandomNodeConfig()
	l_node, err := net.NewNodeWithConfig(l_nodeconf)
	if err != nil {
		t.Fatal(err)
	}
	err = net.Start(l_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	r_nodeconf := adapters.RandomNodeConfig()
	r_node, err := net.NewNodeWithConfig(r_nodeconf)
	if err != nil {
		t.Fatal(err)
	}
	err = net.Start(r_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	err = net.Connect(l_node.ID(), r_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	l_rpc, err := l_node.Client()
	if err != nil {
		t.Fatal(err)
	}
	r_rpc, err := r_node.Client()
	if err != nil {
		t.Fatal(err)
	}

	// create the resource
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = swarms[l_node.ID()].api.ResourceCreate(ctx, resourceName, 2)
	if err != nil {
		t.Fatal(err)
	}
	// update the resource
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, lastPeriod, lastVersion, err := swarms[l_node.ID()].api.ResourceUpdate(ctx, resourceName, []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}

	// turn on notifications
	err = l_rpc.Call(nil, "resource_enableNotifications", resourceName)
	if err != nil {
		t.Fatal(err)
	}

	// get address and pubkey of updater node
	// and add it to the client node address book
	var l_addr string
	err = l_rpc.Call(&l_addr, "pss_baseAddr")
	if err != nil {
		t.Fatal(err)
	}

	var l_pubkey string
	err = l_rpc.Call(&l_pubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	nC := make(chan resource.Notification)
	r_sub_rsrc, err := r_rpc.Subscribe(ctx, "resource", nC, "requestNotification", resourceName, l_pubkey, l_addr)
	if err != nil {
		t.Fatal(err)
	}
	defer r_sub_rsrc.Unsubscribe()

	nsg := <-nC

	time.Sleep(time.Second)
	// update the resource
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, lastPeriod, lastVersion, err = swarms[l_node.ID()].api.ResourceUpdate(ctx, resourceName, []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}

	nsg = <-nC
	if nsg.Period != lastPeriod || nsg.Version != lastVersion {
		t.Fatalf("Expected period/version %d.%d, got %d.%d", lastPeriod, lastVersion, nsg.Period, nsg.Version)

	}
}
