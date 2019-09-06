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

package swarm

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/api"
	httpapi "github.com/ethersphere/swarm/api/http"
	"github.com/ethersphere/swarm/bzzeth"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/contracts/ens"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/fuse"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/retrieval"
	"github.com/ethersphere/swarm/network/stream/v2"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pss"
	pssmessage "github.com/ethersphere/swarm/pss/message"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
	"github.com/ethersphere/swarm/storage/pin"
	"github.com/ethersphere/swarm/swap"
	"github.com/ethersphere/swarm/tracing"
)

var (
	updateGaugesPeriod = 5 * time.Second
	startCounter       = metrics.NewRegisteredCounter("stack,start", nil)
	stopCounter        = metrics.NewRegisteredCounter("stack,stop", nil)
	uptimeGauge        = metrics.NewRegisteredGauge("stack.uptime", nil)
)

// Swarm abstracts the complete Swarm stack
type Swarm struct {
	config            *api.Config        // swarm configuration
	api               *api.API           // high level api layer (fs/manifest)
	dns               api.Resolver       // DNS registrar
	fileStore         *storage.FileStore // distributed preimage archive, the local API to the storage with document level storage/retrieval support
	streamer          *stream.Registry
	retrieval         *retrieval.Retrieval
	bzz               *network.Bzz // the logistic manager
	bzzEth            *bzzeth.BzzEth
	backend           cswap.Backend
	privateKey        *ecdsa.PrivateKey
	netStore          *storage.NetStore
	sfs               *fuse.SwarmFS // need this to cleanup all the active mounts on node exit
	ps                *pss.Pss
	swap              *swap.Swap
	stateStore        *state.DBStore
	tags              *chunk.Tags
	accountingMetrics *protocols.AccountingMetrics
	cleanupFuncs      []func() error
	pinAPI            *pin.API // API object implements all pinning related commands

	tracerClose io.Closer
}

// NewSwarm creates a new swarm service instance
// implements node.Service
// If mockStore is not nil, it will be used as the storage for chunk data.
// MockStore should be used only for testing.
func NewSwarm(config *api.Config, mockStore *mock.NodeStore) (self *Swarm, err error) {
	if bytes.Equal(common.FromHex(config.PublicKey), storage.ZeroAddr) {
		return nil, fmt.Errorf("empty public key")
	}
	if bytes.Equal(common.FromHex(config.BzzKey), storage.ZeroAddr) {
		return nil, fmt.Errorf("empty bzz key")
	}

	self = &Swarm{
		config:       config,
		privateKey:   config.ShiftPrivateKey(),
		cleanupFuncs: []func() error{},
	}
	log.Debug("Setting up Swarm service components")

	// Swap initialization
	if config.SwapEnabled {
		// for now, Swap can only be enabled in a whitelisted network
		if self.config.NetworkID != swap.AllowedNetworkID {
			return nil, fmt.Errorf("swap can only be enabled under Network ID %d, found Network ID %d instead", swap.AllowedNetworkID, self.config.NetworkID)
		}
		// if Swap is enabled, we MUST have a contract API
		if self.config.SwapBackendURL == "" {
			return nil, errors.New("swap enabled but no contract address given; fatal error condition, aborting")
		}
		log.Info("connecting to SWAP API", "url", self.config.SwapBackendURL)
		self.backend, err = ethclient.Dial(self.config.SwapBackendURL)
		if err != nil {
			return nil, fmt.Errorf("error connecting to SWAP API %s: %s", self.config.SwapBackendURL, err)
		}

		// initialize the balances store
		swapStore, err := state.NewDBStore(filepath.Join(config.Path, "swap.db"))
		if err != nil {
			return nil, err
		}
		// create the accounting objects
		self.swap = swap.New(swapStore, self.privateKey, self.backend)
		// start anonymous metrics collection
		self.accountingMetrics = protocols.SetupAccountingMetrics(10*time.Second, filepath.Join(config.Path, "metrics.db"))
	}

	config.HiveParams.Discovery = true

	if config.DisableAutoConnect {
		config.HiveParams.DisableAutoConnect = true
	}

	bzzconfig := &network.BzzConfig{
		NetworkID:    config.NetworkID,
		OverlayAddr:  common.FromHex(config.BzzKey),
		HiveParams:   config.HiveParams,
		LightNode:    config.LightNodeEnabled,
		BootnodeMode: config.BootnodeMode,
	}

	self.stateStore, err = state.NewDBStore(filepath.Join(config.Path, "state-store.db"))
	if err != nil {
		return
	}

	// set up high level api
	var resolver *api.MultiResolver
	if len(config.EnsAPIs) > 0 {
		opts := []api.MultiResolverOption{}
		for _, c := range config.EnsAPIs {
			tld, endpoint, addr := parseEnsAPIAddress(c)
			r, err := newEnsClient(endpoint, addr, config, self.privateKey)
			if err != nil {
				return nil, err
			}
			opts = append(opts, api.MultiResolverOptionWithResolver(r, tld))

		}
		resolver = api.NewMultiResolver(opts...)
		self.dns = resolver
	}
	// check that we are not in the old database schema
	// if so - fail and exit
	isLegacy := localstore.IsLegacyDatabase(config.ChunkDbPath)

	if isLegacy {
		return nil, errors.New("Legacy database format detected! Please read the migration announcement at: https://github.com/ethersphere/swarm/blob/master/docs/Migration-v0.3-to-v0.4.md")
	}

	var feedsHandler *feed.Handler
	fhParams := &feed.HandlerParams{}

	feedsHandler = feed.NewHandler(fhParams)

	localStore, err := localstore.New(config.ChunkDbPath, config.BaseKey, &localstore.Options{
		MockStore: mockStore,
		Capacity:  config.DbCapacity,
	})
	if err != nil {
		return nil, err
	}
	lstore := chunk.NewValidatorStore(
		localStore,
		storage.NewContentAddressValidator(storage.MakeHashFunc(storage.DefaultHash)),
		feedsHandler,
	)

	nodeID := config.Enode.ID()
	self.netStore = storage.NewNetStore(lstore, bzzconfig.OverlayAddr, nodeID)

	to := network.NewKademlia(
		common.FromHex(config.BzzKey),
		network.NewKadParams(),
	)
	self.retrieval = retrieval.New(to, self.netStore, bzzconfig.OverlayAddr) // nodeID.Bytes())
	self.netStore.RemoteGet = self.retrieval.RequestFromPeers

	feedsHandler.SetStore(self.netStore)

	syncing := true
	if !config.SyncEnabled || config.LightNodeEnabled || config.BootnodeMode {
		syncing = false
	}

	syncProvider := stream.NewSyncProvider(self.netStore, to, syncing, false)
	self.streamer = stream.New(self.stateStore, bzzconfig.OverlayAddr, syncProvider)
	self.tags = chunk.NewTags() //todo load from state store

	// Swarm Hash Merklised Chunking for Arbitrary-length Document/File storage
	lnetStore := storage.NewLNetStore(self.netStore)
	self.fileStore = storage.NewFileStore(lnetStore, localStore, self.config.FileStoreParams, self.tags)

	log.Debug("Setup local storage")

	self.bzz = network.NewBzz(bzzconfig, to, self.stateStore, stream.Spec, retrieval.Spec, self.streamer.Run, self.retrieval.Run)

	self.bzzEth = bzzeth.New()

	// Pss = postal service over swarm (devp2p over bzz)
	self.ps, err = pss.New(to, config.Pss)
	if err != nil {
		return nil, err
	}
	if pss.IsActiveHandshake {
		pss.SetHandshakeController(self.ps, pss.NewHandshakeParams())
	}

	self.api = api.NewAPI(self.fileStore, self.dns, feedsHandler, self.privateKey, self.tags)

	// Instantiate the pinAPI object with the already opened localstore
	self.pinAPI = pin.NewAPI(localStore, self.stateStore, self.config.FileStoreParams, self.tags, self.api)

	self.sfs = fuse.NewSwarmFS(self.api)
	log.Debug("Initialized FUSE filesystem")

	return self, nil
}

// parseEnsAPIAddress parses string according to format
// [tld:][contract-addr@]url and returns ENSClientConfig structure
// with endpoint, contract address and TLD.
func parseEnsAPIAddress(s string) (tld, endpoint string, addr common.Address) {
	isAllLetterString := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}
	endpoint = s
	if i := strings.Index(endpoint, ":"); i > 0 {
		if isAllLetterString(endpoint[:i]) && len(endpoint) > i+2 && endpoint[i+1:i+3] != "//" {
			tld = endpoint[:i]
			endpoint = endpoint[i+1:]
		}
	}
	if i := strings.Index(endpoint, "@"); i > 0 {
		addr = common.HexToAddress(endpoint[:i])
		endpoint = endpoint[i+1:]
	}
	return
}

// ensClient provides functionality for api.ResolveValidator
type ensClient struct {
	*ens.ENS
	*ethclient.Client
}

// newEnsClient creates a new ENS client for that is a consumer of
// a ENS API on a specific endpoint. It is used as a helper function
// for creating multiple resolvers in NewSwarm function.
func newEnsClient(endpoint string, addr common.Address, config *api.Config, privkey *ecdsa.PrivateKey) (*ensClient, error) {
	log.Info("connecting to ENS API", "url", endpoint)
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error connecting to ENS API %s: %s", endpoint, err)
	}
	ethClient := ethclient.NewClient(client)

	ensRoot := config.EnsRoot
	if addr != (common.Address{}) {
		ensRoot = addr
	} else {
		a, err := detectEnsAddr(client)
		if err == nil {
			ensRoot = a
		} else {
			log.Warn(fmt.Sprintf("could not determine ENS contract address, using default %s", ensRoot), "err", err)
		}
	}
	transactOpts := bind.NewKeyedTransactor(privkey)
	dns, err := ens.NewENS(transactOpts, ensRoot, ethClient)
	if err != nil {
		return nil, err
	}
	log.Debug(fmt.Sprintf("-> Swarm Domain Name Registrar %v @ address %v", endpoint, ensRoot.Hex()))
	return &ensClient{
		ENS:    dns,
		Client: ethClient,
	}, err
}

// detectEnsAddr determines the ENS contract address by getting both the
// version and genesis hash using the client and matching them to either
// mainnet or testnet addresses
func detectEnsAddr(client *rpc.Client) (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	if err := client.CallContext(ctx, &version, "net_version"); err != nil {
		return common.Address{}, err
	}

	block, err := ethclient.NewClient(client).BlockByNumber(ctx, big.NewInt(0))
	if err != nil {
		return common.Address{}, err
	}

	switch {

	case version == "1" && block.Hash() == params.MainnetGenesisHash:
		log.Info("using Mainnet ENS contract address", "addr", ens.MainNetAddress)
		return ens.MainNetAddress, nil

	case version == "3" && block.Hash() == params.TestnetGenesisHash:
		log.Info("using Testnet ENS contract address", "addr", ens.TestNetAddress)
		return ens.TestNetAddress, nil

	default:
		return common.Address{}, fmt.Errorf("unknown version and genesis hash: %s %s", version, block.Hash())
	}
}

/*
Start is called when the stack is started
* starts the network kademlia hive peer management
* (starts netStore level 0 api)
* starts DPA level 1 api (chunking -> store/retrieve requests)
* (starts level 2 api)
* starts http proxy server
* registers url scheme handlers for bzz, etc
* TODO: start subservices like sword, swear, swarmdns
*/
// implements the node.Service interface
func (s *Swarm) Start(srv *p2p.Server) error {
	startTime := time.Now()

	s.tracerClose = tracing.Closer

	// update uaddr to correct enode
	newaddr := s.bzz.UpdateLocalAddr([]byte(srv.Self().URLv4()))
	log.Info("Updated bzz local addr", "oaddr", fmt.Sprintf("%x", newaddr.OAddr), "uaddr", fmt.Sprintf("%s", newaddr.UAddr))

	if s.config.SwapEnabled {
		if err := s.swap.StartChequebook(s.config.Contract); err != nil {
			return err
		}
	} else {
		log.Info("SWAP disabled: no chequebook set")
	}

	log.Info("Starting bzz service")

	err := s.bzz.Start(srv)
	if err != nil {
		log.Error("bzz failed", "err", err)
		return err
	}
	log.Info("Swarm network started", "bzzaddr", fmt.Sprintf("%x", s.bzz.Hive.BaseAddr()))

	err = s.bzzEth.Start(srv)
	if err != nil {
		return err
	}

	if s.ps != nil {
		s.ps.Start(srv)
	}
	// start swarm http proxy server
	if s.config.Port != "" {
		addr := net.JoinHostPort(s.config.ListenAddr, s.config.Port)
		server := httpapi.NewServer(s.api, s.pinAPI, s.config.Cors)

		if s.config.Cors != "" {
			log.Info("Swarm HTTP proxy CORS headers", "allowedOrigins", s.config.Cors)
		}

		go func() {
			// We need to use net.Listen because the addr could be on port '0',
			// which means that the OS will allocate a port for us
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				log.Error("Could not open a port for Swarm HTTP proxy", "err", err.Error())
				return
			}
			s.config.Port = strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
			log.Info("Starting Swarm HTTP proxy", "port", s.config.Port)

			err = http.Serve(listener, server)
			if err != nil {
				log.Error("Could not start Swarm HTTP proxy", "err", err.Error())
			}
		}()
	}

	doneC := make(chan struct{})

	s.cleanupFuncs = append(s.cleanupFuncs, func() error {
		close(doneC)
		return nil
	})

	go func(time.Time) {
		for {
			select {
			case <-time.After(updateGaugesPeriod):
				uptimeGauge.Update(time.Since(startTime).Nanoseconds())
			case <-doneC:
				return
			}
		}
	}(startTime)

	startCounter.Inc(1)
	if err := s.streamer.Start(srv); err != nil {
		return err
	}
	return s.retrieval.Start(srv)
}

// Stop stops all component services.
// Implements the node.Service interface.
func (s *Swarm) Stop() error {
	if s.tracerClose != nil {
		err := s.tracerClose.Close()
		tracing.FinishSpans()
		if err != nil {
			return err
		}
	}

	if s.ps != nil {
		s.ps.Stop()
	}
	if s.swap != nil {
		s.swap.Stop()
	}
	if s.accountingMetrics != nil {
		s.accountingMetrics.Close()
	}

	if err := s.streamer.Stop(); err != nil {
		log.Error("streamer stop", "err", err)
	}
	if err := s.retrieval.Stop(); err != nil {
		log.Error("retrieval stop", "err", err)
	}

	if s.netStore != nil {
		s.netStore.Close()
	}
	s.sfs.Stop()
	stopCounter.Inc(1)

	err := s.bzzEth.Stop()
	if err != nil {
		log.Error("error during bzz-eth shutdown", "err", err)
	}

	err = s.bzz.Stop()
	if s.stateStore != nil {
		s.stateStore.Close()
	}

	for _, cleanF := range s.cleanupFuncs {
		err = cleanF()
		if err != nil {
			log.Error("encountered an error while running cleanup function", "err", err)
			break
		}
	}
	return err
}

// Protocols implements the node.Service interface
func (s *Swarm) Protocols() (protos []p2p.Protocol) {
	if s.config.BootnodeMode {
		protos = append(protos, s.bzz.Protocols()...)
	} else {
		protos = append(protos, s.bzz.Protocols()...)
		protos = append(protos, s.bzzEth.Protocols()...)
		if s.ps != nil {
			protos = append(protos, s.ps.Protocols()...)
		}

		if s.swap != nil {
			protos = append(protos, s.swap.Protocols()...)
		}
	}
	return
}

// APIs returns the RPC API descriptors the Swarm implementation offers
// implements node.Service
func (s *Swarm) APIs() []rpc.API {
	apis := []rpc.API{
		// public APIs
		{
			Namespace: "bzz",
			Version:   "3.0",
			Service:   &Info{s.config},
			Public:    true,
		},
		// admin APIs
		{
			Namespace: "bzz",
			Version:   "3.0",
			Service:   api.NewInspector(s.api, s.bzz.Hive, s.netStore, s.streamer),
			Public:    false,
		},
		{
			Namespace: "swarmfs",
			Version:   fuse.SwarmFSVersion,
			Service:   s.sfs,
			Public:    false,
		},
		{
			Namespace: "accounting",
			Version:   protocols.AccountingVersion,
			Service:   protocols.NewAccountingApi(s.accountingMetrics),
			Public:    false,
		},
	}

	apis = append(apis, s.bzz.APIs()...)
	apis = append(apis, s.streamer.APIs()...)
	apis = append(apis, s.bzzEth.APIs()...)

	if s.ps != nil {
		apis = append(apis, s.ps.APIs()...)
	}

	if s.config.SwapEnabled {
		apis = append(apis, s.swap.APIs()...)
	}

	return apis
}

// RegisterPssProtocol adds a devp2p protocol to the swarm node's Pss instance
func (s *Swarm) RegisterPssProtocol(topic *pssmessage.Topic, spec *protocols.Spec, targetprotocol *p2p.Protocol, options *pss.ProtocolParams) (*pss.Protocol, error) {
	return pss.RegisterProtocol(s.ps, topic, spec, targetprotocol, options)
}

// Info represents the current Swarm node's configuration
type Info struct {
	*api.Config
}

// Info returns the current Swarm configuration
func (i *Info) Info() *Info {
	return i
}
