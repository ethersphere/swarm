package bzz

import (
	"bytes"
	"fmt"
	"net"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/registrar"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

// the swarm stack
type Swarm struct {
	config    *Config
	chunker   *TreeChunker
	registrar registrar.VersionedRegistrar
	dpa       *DPA
	hive      *hive
	netStore  *netStore
	api       *Api
}

//
func NewSwarm(config *Config) (self *Swarm, err error) {

	self = &Swarm{
		config: config,
	}

	self.hive, err = newHive()
	if err != nil {
		return
	}

	self.netStore, err = newNetStore(config.Path, self.hive)
	if err != nil {
		return
	}

	dpa := newDPA(self.netStore)

	self.api = NewApi(dpa)

	return
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

/*
Start is called when the ethereum stack is started
- calls Init() on treechunker
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
	} else if bytes.Equal(node.ID[:], zeroKey) {
		err = fmt.Errorf("zero ID invalid")
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
					ID:   node.ID[:],
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
	self.chunker.Init()
	self.dpa.Start()
	if self.config.Port != "" {
		fmt.Printf("PORT: %v\n", self.config.Port)
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

// Bzz returns the bzz protocol class instances of which run on every peer
func (self *Swarm) Bzz() (p2p.Protocol, error) {
	return BzzProtocol(self.netStore, self.config.Swap)
}

func newDPA(store ChunkStore) *DPA {
	chunker := &TreeChunker{}
	chunker.Init()
	return &DPA{
		Chunker:    chunker,
		ChunkStore: store,
	}
}
