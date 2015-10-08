package bzz

import (
	"bytes"
	"fmt"
	"net"
	"path/filepath"
	"strconv"

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
func NewSwarm(config *Config) (self *Swarm, proto p2p.Protocol, err error) {

	self = &Swarm{
		config: config,
	}

	self.hive, err = newHive()
	if err != nil {
		return
	}

	self.netStore, err = newNetStore(filepath.Join(config.Path, "db"), self.hive)
	if err != nil {
		return
	}

	self.dpa = newDPA(self.netStore)

	self.api = NewApi(self.dpa)

	proto, err = BzzProtocol(self.netStore, self.config.Swap)
	return
}

/*
Start is called when the ethereum stack is started
- launches the dpa (listening for chunk store/retrieve requests)
- launches the netStore (starts kademlia hive peer management)
- starts an http server
*/
func (self *Swarm) Start(node *discover.Node, listenAddr string, connectPeer func(string) error) {
	var err error
	if node == nil {
		err = fmt.Errorf("basenode nil")
	} else if self.netStore == nil {
		err = fmt.Errorf("netStore is nil")
	} else if connectPeer == nil {
		err = fmt.Errorf("no connect peer function")
	} else if bytes.Equal(common.FromHex(self.config.PublicKey), zeroKey) {
		err = fmt.Errorf("empty public key")
	} else if bytes.Equal(common.FromHex(self.config.BzzKey), zeroKey) {
		err = fmt.Errorf("empty bzz key")
	} else { // this is how we calculate the bzz address of the node
		// ideally this should be using the swarm hash function

		var port string
		_, port, err = net.SplitHostPort(listenAddr)
		if err == nil {
			intport, err := strconv.Atoi(port)
			if err != nil {
				err = fmt.Errorf("invalid port in '%s'", listenAddr)
			} else {
				baseAddr := &peerAddr{
					ID:   common.FromHex(self.config.PublicKey),
					IP:   node.IP,
					Port: uint16(intport),
				}
				baseAddr.new()
				err = self.hive.start(baseAddr, filepath.Join(self.config.Path, "bzz-peers.json"), connectPeer)
				if err == nil {
					err = self.netStore.start(baseAddr)
					if err == nil {
						glog.V(logger.Info).Infof("[BZZ] Swarm network started on bzz address: %064x", baseAddr.hash[:])
					}
				}
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
		api:    NewApi(dpa),
		config: config,
	}

	return
}

func newLocalDPA(datadir string) (*DPA, error) {

	dbStore, err := newDbStore(datadir)
	// dbStore.setCapacity(50000)
	if err != nil {
		return nil, err
	}

	return newDPA(&localStore{
		newMemStore(dbStore),
		dbStore,
	}), nil
}

func newDPA(store ChunkStore) *DPA {
	chunker := &TreeChunker{}
	chunker.Init()
	return &DPA{
		Chunker:    chunker,
		ChunkStore: store,
	}
}
