// Copyright 2016 The go-ethereum Authors
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

// Contains all the wrappers from the node package to support client side node
// management on mobile platforms.

package geth

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethstats"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/les"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/swarm"
	swarmapi "github.com/ethereum/go-ethereum/swarm/api"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

// NodeConfig represents the collection of configuration values to fine tune the Geth
// node embedded into a mobile process. The available values are a subset of the
// entire API provided by go-ethereum to reduce the maintenance surface and dev
// complexity.
type NodeConfig struct {
	// Bootstrap nodes used to establish connectivity with the rest of the network.
	BootstrapNodes *Enodes

	// MaxPeers is the maximum number of peers that can be connected. If this is
	// set to zero, then only the configured static and trusted peers can connect.
	MaxPeers int

	// EthereumEnabled specifies whether the node should run the Ethereum protocol.
	EthereumEnabled bool

	// EthereumNetworkID is the network identifier used by the Ethereum protocol to
	// decide if remote peers should be accepted or not.
	EthereumNetworkID int64 // uint64 in truth, but Java can't handle that...

	// EthereumGenesis is the genesis JSON to use to seed the blockchain with. An
	// empty genesis state is equivalent to using the mainnet's state.
	EthereumGenesis string

	// EthereumDatabaseCache is the system memory in MB to allocate for database caching.
	// A minimum of 16MB is always reserved.
	EthereumDatabaseCache int

	// EthereumNetStats is a netstats connection string to use to report various
	// chain, transaction and node stats to a monitoring server.
	//
	// It has the form "nodename:secret@host:port"
	EthereumNetStats string

	// WhisperEnabled specifies whether the node should run the Whisper protocol.
	WhisperEnabled bool

	// Listening address of pprof server.
	PprofAddress string
	// NB: this is a hack, and very likely not a permanent solution
	// PssEnabled specifies whether the node should run pss
	PssEnabled  bool
	PssAccount  string
	PssPassword string
}

// defaultNodeConfig contains the default node configuration values to use if all
// or some fields are missing from the user's specified list.
var defaultNodeConfig = &NodeConfig{
	BootstrapNodes:        FoundationBootnodes(),
	MaxPeers:              25,
	EthereumEnabled:       true,
	EthereumNetworkID:     1,
	EthereumDatabaseCache: 16,
}

// NewNodeConfig creates a new node option set, initialized to the default values.
func NewNodeConfig() *NodeConfig {
	config := *defaultNodeConfig
	return &config
}

// Node represents a Geth Ethereum node instance.
type Node struct {
	node *node.Node
	Ps   *Pss
}

// NewNode creates and configures a new Geth node.

func NewNode(datadir string, config *NodeConfig, ks *keystore.KeyStore) (stack *Node, _ error) {
	return NewNodeWithKeystore(datadir, config, nil)
}

func NewNodeWithKeystoreString(datadir string, config *NodeConfig, ksstr string) (stack *Node, _ error) {
	ks := NewKeyStore(ksstr, keystore.LightScryptN, keystore.LightScryptP)
	return NewNodeWithKeystore(datadir, config, ks)
	//ks := rawStack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

func NewNodeWithKeystore(datadir string, config *NodeConfig, ks *KeyStore) (stack *Node, _ error) {

	resultNode := &Node{}
	// If no or partial configurations were specified, use defaults
	if config == nil {
		config = NewNodeConfig()
	}
	if config.MaxPeers == 0 {
		config.MaxPeers = defaultNodeConfig.MaxPeers
	}
	if config.BootstrapNodes == nil || config.BootstrapNodes.Size() == 0 {
		config.BootstrapNodes = defaultNodeConfig.BootstrapNodes
	}

	if config.PprofAddress != "" {
		debug.StartPProf(config.PprofAddress)
	}

	// Create the empty networking stack
	nodeConf := &node.Config{
		Name:        clientIdentifier,
		Version:     params.Version,
		DataDir:     datadir,
		KeyStoreDir: filepath.Join(datadir, "keystore"), // Mobile should never use internal keystores!
		P2P: p2p.Config{
			NoDiscovery:      true,
			DiscoveryV5:      true,
			BootstrapNodesV5: config.BootstrapNodes.nodes,
			ListenAddr:       ":0",
			NAT:              nat.Any(),
			MaxPeers:         config.MaxPeers,
		},
	}

	rawStack, err := node.New(nodeConf)
	if err != nil {
		return nil, err
	}

	debug.Memsize.Add("node", rawStack)

	var genesis *core.Genesis
	if config.EthereumGenesis != "" {
		// Parse the user supplied genesis spec if not mainnet
		genesis = new(core.Genesis)
		if err := json.Unmarshal([]byte(config.EthereumGenesis), genesis); err != nil {
			return nil, fmt.Errorf("invalid genesis spec: %v", err)
		}
		// If we have the testnet, hard code the chain configs too
		if config.EthereumGenesis == TestnetGenesis() {
			genesis.Config = params.TestnetChainConfig
			if config.EthereumNetworkID == 1 {
				config.EthereumNetworkID = 3
			}
		}
	}
	// Register the Ethereum protocol if requested
	if config.EthereumEnabled {
		ethConf := eth.DefaultConfig
		ethConf.Genesis = genesis
		ethConf.SyncMode = downloader.LightSync
		ethConf.NetworkId = uint64(config.EthereumNetworkID)
		ethConf.DatabaseCache = config.EthereumDatabaseCache
		if err := rawStack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
			return les.New(ctx, &ethConf)
		}); err != nil {
			return nil, fmt.Errorf("ethereum init: %v", err)
		}
		// If netstats reporting is requested, do it
		if config.EthereumNetStats != "" {
			if err := rawStack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
				var lesServ *les.LightEthereum
				ctx.Service(&lesServ)

				return ethstats.New(config.EthereumNetStats, nil, lesServ)
			}); err != nil {
				return nil, fmt.Errorf("netstats init: %v", err)
			}
		}
	}
	// Register the Whisper protocol if requested
	if config.WhisperEnabled {
		if err := rawStack.Register(func(*node.ServiceContext) (node.Service, error) {
			return whisper.New(&whisper.DefaultConfig), nil
		}); err != nil {
			return nil, fmt.Errorf("whisper init: %v", err)
		}
	}
	if config.PssEnabled && ks != nil {
		log.Debug("pss enabled")
		bzzSvc := func(ctx *node.ServiceContext) (node.Service, error) {
			//ks := rawStack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
			log.Warn("keystore", "ks", ks)
			var a accounts.Account
			var err error
			if common.IsHexAddress(config.PssAccount) {
				//a, err = ks.Find(accounts.Account{Address: common.HexToAddress(config.PssAccount)})
				a = ks.GetAccounts().accounts[0]
			} else if ix, ixerr := strconv.Atoi(config.PssAccount); ixerr == nil && ix > 0 {
				if accounts := ks.GetAccounts().accounts; len(accounts) > ix {
					a = accounts[ix]
				} else {
					err = fmt.Errorf("index %d higher than number of accounts %d", ix, len(accounts))
				}
			} else {
				return nil, fmt.Errorf("Can't find swarm account key %s", config.PssAccount)
			}
			if err != nil {
				return nil, fmt.Errorf("Can't find swarm account key: %v - Is the provided bzzaccount(%s) from the right datadir/Path?", err, config.PssAccount)
			}
			keyjson, err := ioutil.ReadFile(a.URL.Path)
			if err != nil {
				return nil, fmt.Errorf("Can't load swarm account key: %v", err)
			}
			var bzzkey *ecdsa.PrivateKey
			//for i := 0; i < 3; i++ {
			//	password := getPassPhrase(fmt.Sprintf("Unlocking swarm account %s [%d/3]", a.Address.Hex(), i+1), i, passwords)
			//key, err := keystore.DecryptKey(keyjson, password)
			key, err := keystore.DecryptKey(keyjson, config.PssPassword)
			if err == nil {
				bzzkey = key.PrivateKey
			}
			//}
			if bzzkey == nil {
				return nil, fmt.Errorf("Can't decrypt swarm account key")
			}
			bzzconfig := swarmapi.NewConfig()
			bzzconfig.SyncEnabled = false
			bzzconfig.Path = rawStack.DataDir()
			bzzconfig.Init(bzzkey)

			svc, err := swarm.NewSwarm(ctx, nil, bzzconfig, nil)
			resultNode.Ps = &Pss{ps: svc.Ps}
			if err != nil {
				log.Error("swarm svc", "err", err)
			}
			return svc, err
		}
		if err := rawStack.Register(bzzSvc); err != nil {
			return nil, fmt.Errorf("pss init: %v", err)
		}
	}
	resultNode.node = rawStack
	return resultNode, nil
}

// Start creates a live P2P node and starts running it.
func (n *Node) Start() error {
	return n.node.Start()
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (n *Node) Stop() error {
	return n.node.Stop()
}

// GetEthereumClient retrieves a client to access the Ethereum subsystem.
func (n *Node) GetEthereumClient() (client *EthereumClient, _ error) {
	rpc, err := n.node.Attach()
	if err != nil {
		return nil, err
	}
	return &EthereumClient{ethclient.NewClient(rpc)}, nil
}

// GetNodeInfo gathers and returns a collection of metadata known about the host.
func (n *Node) GetNodeInfo() *NodeInfo {
	return &NodeInfo{n.node.Server().NodeInfo()}
}

// GetPeersInfo returns an array of metadata objects describing connected peers.
func (n *Node) GetPeersInfo() *PeerInfos {
	return &PeerInfos{n.node.Server().PeersInfo()}
}
