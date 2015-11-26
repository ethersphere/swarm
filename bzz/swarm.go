package bzz

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/common/httpclient"
	"github.com/ethereum/go-ethereum/common/registrar/ethreg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

const (
	singletonSwarmDbCapacity    = 50000
	singletonSwarmCacheCapacity = 500
)

// the swarm stack
type Swarm struct {
	config   *Config                // swarm configuration
	dpa      *DPA                   // distributed preimage archive
	hive     *hive                  // the logistic manager
	netStore *netStore              // dht storage logic
	api      *Api                   // high level api layer (fs/manifest)
	client   *httpclient.HTTPClient // bzz capable light http client
}

// creates a new swarm instance
func NewSwarm(stack *node.ServiceContext, config *Config) (self *Swarm, err error) {

	if bytes.Equal(common.FromHex(config.PublicKey), zeroKey) {
		return nil, fmt.Errorf("empty public key")
	}
	if bytes.Equal(common.FromHex(config.BzzKey), zeroKey) {
		return nil, fmt.Errorf("empty bzz key")
	}

	self = &Swarm{
		config: config,
		client: stack.Service("eth").(*eth.Ethereum).HTTPClient(),
	}

	self.hive, err = newHive(
		common.HexToHash(self.config.BzzKey), // key to hive (kademlia base address)
		config.HiveParams,                    // configuration parameters
	)
	if err != nil {
		return
	}

	self.netStore, err = newNetStore(makeHashFunc(config.ChunkerParams.Hash), config.StoreParams, self.hive)
	if err != nil {
		return
	}

	self.dpa = newDPA(self.netStore, self.config.ChunkerParams)
	ethereum := stack.Service("eth").(*eth.Ethereum)
	backend := newEthApi(ethereum)

	self.api = NewApi(self.dpa, ethreg.New(backend), self.config)

	// set chequebook
	err = self.SetChequebook(backend)
	if err != nil {
		return nil, fmt.Errorf("Unable to set swarm backend: %v", err)
	}
	return self, nil
}

/*
Start is called when the stack is started
- launches the dpa (listening for chunk store/retrieve requests)
- launches the netStore (starts kademlia hive peer management)
- starts an http server
*/
// implements the node.Service interface
func (self *Swarm) Start(net *p2p.Server) error {
	var err error
	connectPeer := func(url string) error {
		node, err := discover.ParseNode(url)
		if err != nil {
			return fmt.Errorf("invalid node URL: %v", err)
		}
		net.AddPeer(node)
		return nil
	}

	err = self.hive.start(
		discover.PubkeyID(&net.PrivateKey.PublicKey),
		func() string { return net.ListenAddr },
		connectPeer,
	)
	if err != nil {
		glog.V(logger.Warn).Infof("[BZZ] Swarm hive could not be started: %v", err)
	} else {
		err = self.netStore.start()
		if err != nil {
			glog.V(logger.Info).Infof("[BZZ] Swarm netstore could not be started: %v", err)
		} else {
			glog.V(logger.Info).Infof("[BZZ] Swarm network started on bzz address: %v", self.hive.addr)
		}
	}

	if err != nil {
		glog.V(logger.Info).Infof("[BZZ] Swarm started offline: %v", err)
	}

	self.dpa.Start()

	// start swarm http proxy server
	if self.config.Port != "" {
		go startHttpServer(self.api, self.config.Port)
	}
	// register roundtripper (using proxy) as bzz scheme handler
	// for the ethereum http client
	self.client.RegisterScheme("bzz", &RoundTripper{
		Port: self.config.Port,
	})
	return nil
}

// implements the node.Service interface
// stops all component services.
func (self *Swarm) Stop() error {
	self.dpa.Stop()
	self.netStore.stop()
	self.hive.stop()
	if ch := self.config.Swap.chequebook(); ch != nil {
		ch.Stop()
		ch.Save()
	}
	return self.config.Save()
}

// implements the node.Service interfacec
func (self *Swarm) Protocols() []p2p.Protocol {
	proto, err := BzzProtocol(self.netStore, self.config.Swap, self.config.SyncParams)
	if err != nil {
		return nil
	}
	return []p2p.Protocol{proto}
}

func (self *Swarm) Api() *Api {
	return self.api
}

// Backend interface implemented by eth or JSON-IPC client
func (self *Swarm) SetChequebook(backend chequebook.Backend) (err error) {
	done, err := self.config.Swap.setChequebook(self.config.Path, backend)
	if err != nil {
		return err
	}
	go func() {
		ok := <-done
		if ok {
			glog.V(logger.Info).Infof("[BZZ] Swarm: new chequebook set (%v): saving config file, resetting all connections in the hive", self.config.Swap.Contract)
			self.config.Save()
			self.hive.dropAll()
		}
	}()
	return nil
}

// Local swarm without netStore
func NewLocalSwarm(datadir, port string) (self *Swarm, err error) {

	prvKey, err := crypto.GenerateKey()
	if err != nil {
		return
	}

	config, err := NewConfig(datadir, common.Address{}, prvKey)
	if err != nil {
		return
	}
	config.Port = port

	dpa, err := newLocalDPA(datadir)
	if err != nil {
		return
	}

	self = &Swarm{
		api:    NewApi(dpa, nil, config),
		config: config,
	}

	return
}

// for testing locally
func newLocalDPA(datadir string) (*DPA, error) {

	hash := makeHashFunc("SHA256")

	dbStore, err := newDbStore(datadir, hash, singletonSwarmDbCapacity, 0)
	if err != nil {
		return nil, err
	}

	return newDPA(&localStore{
		newMemStore(dbStore, singletonSwarmCacheCapacity),
		dbStore,
	}, NewChunkerParams()), nil
}

func newDPA(store ChunkStore, params *ChunkerParams) *DPA {
	chunker := NewTreeChunker(params)
	return &DPA{
		Chunker:    chunker,
		ChunkStore: store,
	}
}
