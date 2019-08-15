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
	"math/big"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/contracts/chequebook"
	"github.com/ethersphere/swarm/contracts/ens"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/stream"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
)

var (
	updateGaugesPeriod = 5 * time.Second
	startCounter       = metrics.NewRegisteredCounter("stack,start", nil)
	stopCounter        = metrics.NewRegisteredCounter("stack,stop", nil)
	uptimeGauge        = metrics.NewRegisteredGauge("stack.uptime", nil)
)

// NewSwarm creates a new swarm service instance
// implements node.Service
// If mockStore is not nil, it will be used as the storage for chunk data.
// MockStore should be used only for testing.
func NewSwarm(config *api.Config, mockStore *mock.NodeStore) (svcs []node.ServiceConstructor, err error) {
	if bytes.Equal(common.FromHex(config.PublicKey), storage.ZeroAddr) {
		return nil, fmt.Errorf("empty public key")
	}
	if bytes.Equal(common.FromHex(config.BzzKey), storage.ZeroAddr) {
		return nil, fmt.Errorf("empty bzz key")
	}

	rpcInprocServer := rpc.NewServer()
	rpcInprocDialer := func() (*rpc.Client, error) {
		return rpc.DialInProc(rpcInprocServer), nil
	}

	var swap protocols.Balance = nil

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

	stateStore, err := state.NewDBStore(filepath.Join(config.Path, "state-store.db"))
	if err != nil {
		return
	}

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
	netStore := storage.NewNetStore(lstore, nodeID)

	delivery := stream.NewDelivery(nil, netStore)
	netStore.RemoteGet = delivery.RequestFromPeers

	feedsHandler.SetStore(netStore)

	syncing := stream.SyncingAutoSubscribe
	if !config.SyncEnabled || config.LightNodeEnabled {
		syncing = stream.SyncingDisabled
	}

	registryOptions := &stream.RegistryOptions{
		SkipCheck:       config.DeliverySkipCheck,
		Syncing:         syncing,
		SyncUpdateDelay: config.SyncUpdateDelay,
		MaxPeerServers:  config.MaxStreamPeerServers,
	}
	streamerFunc := func(ctx *node.ServiceContext) (node.Service, error) {
		streamer := stream.NewRegistry(nodeID, delivery, netStore, stateStore, registryOptions, swap)
		return streamer, nil
	}
	_ = streamerFunc

	log.Debug("Setup local storage")

	bzzFunc := func(ctx *node.ServiceContext) (node.Service, error) {
		bzz := network.NewBzz(bzzconfig, stateStore)
		return bzz, nil
	}
	svcs = append(svcs, bzzFunc)

	// Pss = postal service over swarm (devp2p over bzz)
	psFunc := func(ctx *node.ServiceContext) (node.Service, error) {
		config.Pss.RPCDialer = rpcInprocDialer
		ps, err := pss.New(config.Pss)
		rpcInprocServer.RegisterName("pss", ps)
		if err != nil {
			return nil, err
		}
		return ps, nil
	}
	svcs = append(svcs, psFunc)

	return svcs, nil
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

// serialisable info about swarm
type Info struct {
	*api.Config
	*chequebook.Params
}

func (s *Info) Info() *Info {
	return s
}
