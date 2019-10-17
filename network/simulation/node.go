// Copyright 2018 The go-ethereum Authors
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

package simulation

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/network"
)

var (
	BucketKeyBzzPrivateKey BucketKey = "bzzprivkey"

	// PropertyBootnode is a property string for NodeConfig, representing that a node is a bootnode
	PropertyBootnode = "bootnode"
)

// NodeIDs returns NodeIDs for all nodes in the network.
func (s *Simulation) NodeIDs() (ids []enode.ID) {
	nodes := s.Net.GetNodes()
	ids = make([]enode.ID, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID()
	}
	return ids
}

// UpNodeIDs returns NodeIDs for nodes that are up in the network.
func (s *Simulation) UpNodeIDs() (ids []enode.ID) {
	nodes := s.Net.GetNodes()
	for _, node := range nodes {
		if node.Up() {
			ids = append(ids, node.ID())
		}
	}
	return ids
}

// DownNodeIDs returns NodeIDs for nodes that are stopped in the network.
func (s *Simulation) DownNodeIDs() (ids []enode.ID) {
	nodes := s.Net.GetNodes()
	for _, node := range nodes {
		if !node.Up() {
			ids = append(ids, node.ID())
		}
	}
	return ids
}

// AddNodeOption defines the option that can be passed
// to Simulation.AddNode method.
type AddNodeOption func(*adapters.NodeConfig)

// AddNodeWithMsgEvents sets the EnableMsgEvents option
// to NodeConfig.
func AddNodeWithMsgEvents(enable bool) AddNodeOption {
	return func(o *adapters.NodeConfig) {
		o.EnableMsgEvents = enable
	}
}

// AddNodeWithProperty specifies a property that this node should hold
// in the running services. (e.g. "bootnode", etc)
func AddNodeWithProperty(propertyName string) AddNodeOption {
	return func(o *adapters.NodeConfig) {
		o.Properties = append(o.Properties, propertyName)
	}
}

// AddNodeWithService specifies a service that should be
// started on a node. This option can be repeated as variadic
// argument toe AddNode and other add node related methods.
// If AddNodeWithService is not specified, all services will be started.
func AddNodeWithService(serviceName string) AddNodeOption {
	return func(o *adapters.NodeConfig) {
		o.Services = append(o.Services, serviceName)
	}
}

// AddNode creates a new node with random configuration,
// applies provided options to the config and adds the node to network.
// By default all services will be started on a node. If one or more
// AddNodeWithService option are provided, only specified services will be started.
func (s *Simulation) AddNode(opts ...AddNodeOption) (id enode.ID, err error) {
	conf := adapters.RandomNodeConfig()
	for _, o := range opts {
		o(conf)
	}
	if len(conf.Services) == 0 {
		conf.Services = s.serviceNames
	}

	// add ENR records to the underlying node
	// most importantly the bzz overlay address
	bzzPrivateKey, err := BzzPrivateKeyFromConfig(conf)
	if err != nil {
		return enode.ID{}, err
	}

	enodeParams := &network.EnodeParams{
		PrivateKey: bzzPrivateKey,
	}

	// Check for any properties relevant to the creation of the Enode Record
	for _, property := range conf.Properties {
		if property == PropertyBootnode {
			enodeParams.Bootnode = true
		}
	}

	record, err := network.NewEnodeRecord(enodeParams)
	conf.Record = *record

	// Add the bzz address to the node config
	node, err := s.Net.NewNodeWithConfig(conf)
	if err != nil {
		return id, err
	}
	s.buckets[node.ID()] = new(sync.Map)
	s.SetNodeItem(node.ID(), BucketKeyBzzPrivateKey, bzzPrivateKey)

	return node.ID(), s.Net.Start(node.ID())
}

// AddNodes creates new nodes with random configurations,
// applies provided options to the config and adds nodes to network.
func (s *Simulation) AddNodes(count int, opts ...AddNodeOption) (ids []enode.ID, err error) {
	ids = make([]enode.ID, 0, count)
	for i := 0; i < count; i++ {
		id, err := s.AddNode(opts...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// AddBootnode creates a bootnode using AddNode(opts) and appends it to Simulation.bootNodes
func (s *Simulation) AddBootnode(opts ...AddNodeOption) (id enode.ID, err error) {
	opts = append(opts, AddNodeWithProperty(PropertyBootnode))
	id, err = s.AddNode(opts...)
	if err != nil {
		return id, err
	}

	return id, nil
}

// AddNodesAndConnectFull is a helper method that combines
// AddNodes and ConnectNodesFull. Only new nodes will be connected.
func (s *Simulation) AddNodesAndConnectFull(count int, opts ...AddNodeOption) (ids []enode.ID, err error) {
	if count < 2 {
		return nil, errors.New("count of nodes must be at least 2")
	}
	ids, err = s.AddNodes(count, opts...)
	if err != nil {
		return nil, err
	}
	err = s.Net.ConnectNodesFull(ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// AddNodesAndConnectChain is a helpper method that combines
// AddNodes and ConnectNodesChain. The chain will be continued from the last
// added node, if there is one in simulation using ConnectToLastNode method.
func (s *Simulation) AddNodesAndConnectChain(count int, opts ...AddNodeOption) (ids []enode.ID, err error) {
	if count < 2 {
		return nil, errors.New("count of nodes must be at least 2")
	}
	id, err := s.AddNode(opts...)
	if err != nil {
		return nil, err
	}
	err = s.Net.ConnectToLastNode(id)
	if err != nil {
		return nil, err
	}
	ids, err = s.AddNodes(count-1, opts...)
	if err != nil {
		return nil, err
	}
	ids = append([]enode.ID{id}, ids...)
	err = s.Net.ConnectNodesChain(ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// AddNodesAndConnectRing is a helpper method that combines
// AddNodes and ConnectNodesRing.
func (s *Simulation) AddNodesAndConnectRing(count int, opts ...AddNodeOption) (ids []enode.ID, err error) {
	if count < 2 {
		return nil, errors.New("count of nodes must be at least 2")
	}
	ids, err = s.AddNodes(count, opts...)
	if err != nil {
		return nil, err
	}
	err = s.Net.ConnectNodesRing(ids)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// AddNodesAndConnectStar is a helper method that combines
// AddNodes and ConnectNodesStar.
func (s *Simulation) AddNodesAndConnectStar(count int, opts ...AddNodeOption) (ids []enode.ID, err error) {
	if count < 2 {
		return nil, errors.New("count of nodes must be at least 2")
	}
	ids, err = s.AddNodes(count, opts...)
	if err != nil {
		return nil, err
	}
	err = s.Net.ConnectNodesStar(ids[1:], ids[0])
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// AddNodesAndConnectToBootnode is a helper method that combines
// AddNodes, AddBootnode and ConnectNodesStar, where the center node is a new bootnode.
// The count parameter excludes the bootnode.
func (s *Simulation) AddNodesAndConnectToBootnode(count int, opts ...AddNodeOption) (ids []enode.ID, bootNodeID enode.ID, err error) {
	bootNodeID, err = s.AddBootnode(opts...)
	if err != nil {
		return nil, bootNodeID, err
	}

	ids, err = s.AddNodes(count, opts...)
	if err != nil {
		return nil, bootNodeID, err
	}

	err = s.Net.ConnectNodesStar(ids, bootNodeID)
	if err != nil {
		return nil, bootNodeID, err
	}

	return ids, bootNodeID, nil
}

// UploadSnapshot uploads a snapshot to the simulation
// This method tries to open the json file provided, applies the config to all nodes
// and then loads the snapshot into the Simulation network
func (s *Simulation) UploadSnapshot(ctx context.Context, snapshotFile string, opts ...AddNodeOption) error {
	f, err := os.Open(snapshotFile)
	if err != nil {
		return err
	}

	jsonbyte, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return err
	}
	var snap simulations.Snapshot
	if err := json.Unmarshal(jsonbyte, &snap); err != nil {
		return err
	}

	//the snapshot probably has the property EnableMsgEvents not set
	//set it to true (we need this to wait for messages before uploading)
	for i := range snap.Nodes {
		snap.Nodes[i].Node.Config.EnableMsgEvents = true
		snap.Nodes[i].Node.Config.Services = s.serviceNames
		for _, o := range opts {
			o(snap.Nodes[i].Node.Config)
		}
	}

	if err := s.Net.Load(&snap); err != nil {
		return err
	}
	return s.WaitTillSnapshotRecreated(ctx, &snap)
}

// StartNode starts a node by NodeID.
func (s *Simulation) StartNode(id enode.ID) (err error) {
	return s.Net.Start(id)
}

// StartRandomNode starts a random node.
// Nodes can be excluded by providing their enode.ID.
func (s *Simulation) StartRandomNode(excludeIDs ...enode.ID) (id enode.ID, err error) {
	n := s.Net.GetRandomDownNode(excludeIDs...)
	if n == nil {
		return id, ErrNodeNotFound
	}
	return n.ID(), s.Net.Start(n.ID())
}

// StartRandomNodes starts random nodes. Nodes
// Nodes can be excluded by providing their enode.ID.
func (s *Simulation) StartRandomNodes(count int, excludeIDs ...enode.ID) (ids []enode.ID, err error) {
	ids = make([]enode.ID, 0, count)
	for i := 0; i < count; i++ {
		n := s.Net.GetRandomDownNode(excludeIDs...)
		if n == nil {
			return nil, ErrNodeNotFound
		}
		err = s.Net.Start(n.ID())
		if err != nil {
			return nil, err
		}
		ids = append(ids, n.ID())
	}
	return ids, nil
}

// StopNode stops a node by NodeID.
func (s *Simulation) StopNode(id enode.ID) (err error) {
	return s.Net.Stop(id)
}

// StopRandomNode stops a random node.
// Nodes can be excluded by providing their enode.ID.
func (s *Simulation) StopRandomNode(excludeIDs ...enode.ID) (id enode.ID, err error) {
	n := s.Net.GetRandomUpNode(excludeIDs...)
	if n == nil {
		return id, ErrNodeNotFound
	}
	return n.ID(), s.Net.Stop(n.ID())
}

// StopRandomNodes stops random nodes.
// Nodes can be excluded by providing their enode.ID.
func (s *Simulation) StopRandomNodes(count int, excludeIDs ...enode.ID) (ids []enode.ID, err error) {
	ids = make([]enode.ID, 0, count)
	for i := 0; i < count; i++ {
		n := s.Net.GetRandomUpNode(excludeIDs...)
		if n == nil {
			return nil, ErrNodeNotFound
		}
		err = s.Net.Stop(n.ID())
		if err != nil {
			return nil, err
		}
		ids = append(ids, n.ID())
	}
	return ids, nil
}

// seed the random generator for Simulation.randomNode.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// BzzPrivateKeyFromConfig derives a private key for swarm for the node key
// returns the private key used to generate the bzz key
func BzzPrivateKeyFromConfig(conf *adapters.NodeConfig) (*ecdsa.PrivateKey, error) {
	// pad the seed key some arbitrary data as ecdsa.GenerateKey takes 40 bytes seed data
	privKeyBuf := append(crypto.FromECDSA(conf.PrivateKey), []byte{0x62, 0x7a, 0x7a, 0x62, 0x7a, 0x7a, 0x62, 0x7a}...)
	bzzPrivateKey, err := ecdsa.GenerateKey(crypto.S256(), bytes.NewReader(privKeyBuf))
	if err != nil {
		return nil, err
	}
	return bzzPrivateKey, nil
}
