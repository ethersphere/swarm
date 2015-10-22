package bzz

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/common/registrar"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

// the swarm stack
type Swarm struct {
	config   *Config
	dpa      *DPA
	hive     *hive
	netStore *netStore
	api      *Api
}

//
func NewSwarm(id discover.NodeID, config *Config) (self *Swarm, proto p2p.Protocol, err error) {

	self = &Swarm{
		config: config,
	}

	self.hive, err = newHive(common.HexToHash(self.config.BzzKey), id, config.HiveParams)
	if err != nil {
		return
	}

	self.netStore, err = newNetStore(makeHashFunc(config.ChunkerParams.Hash), config.StoreParams, self.hive)
	if err != nil {
		return
	}

	self.dpa = newDPA(self.netStore, self.config.ChunkerParams)

	self.api = NewApi(self.dpa, self.config)

	proto, err = BzzProtocol(self.netStore, self.config.Swap)
	return
}

/*
Start is called when the ethereum stack is started
- launches the dpa (listening for chunk store/retrieve requests)
- launches the netStore (starts kademlia hive peer management)
- starts an http server
*/
func (self *Swarm) Start(listenAddr func() string, connectPeer func(string) error) {
	var err error
	if self.netStore == nil {
		err = fmt.Errorf("netStore is nil")
	} else if connectPeer == nil {
		err = fmt.Errorf("no connect peer function")
	} else if bytes.Equal(common.FromHex(self.config.PublicKey), zeroKey) {
		err = fmt.Errorf("empty public key")
	} else if bytes.Equal(common.FromHex(self.config.BzzKey), zeroKey) {
		err = fmt.Errorf("empty bzz key")
	} else { // this is how we calculate the bzz address of the node

		err = self.hive.start(listenAddr, connectPeer)
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
	}

	if err != nil {
		glog.V(logger.Info).Infof("[BZZ] Swarm started started offline: %v", err)
	}

	self.dpa.Start()

	if self.config.Port != "" {
		go startHttpServer(self.api, self.config.Port)
	}
}

// stops all component services.
func (self *Swarm) Stop() {
	self.dpa.Stop()
	self.netStore.stop()
	self.hive.stop()
	if ch := self.config.Swap.chequebook; ch != nil {
		ch.Stop()
		ch.Save()
	}
}

func (self *Swarm) Api() *Api {
	return self.api
}

func (self *Swarm) ProxyPort() string {
	return self.config.Port
}

func (self *Swarm) SetRegistrar(reg registrar.VersionedRegistrar) {
	self.api.registrar = reg
}

// Backend interface implemented by xeth.XEth or JSON-IPC client
func (self *Swarm) SetChequebook(backend chequebook.Backend) (err error) {
	return self.config.Swap.setChequebook(self.config.Path, backend)
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
		api:    NewApi(dpa, config),
		config: config,
	}

	return
}

// for testing locally
func newLocalDPA(datadir string) (*DPA, error) {

	hash := makeHashFunc("SHA256")

	dbStore, err := newDbStore(datadir, hash, 50000, 0)
	if err != nil {
		return nil, err
	}

	return newDPA(&localStore{
		newMemStore(dbStore, 500),
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
