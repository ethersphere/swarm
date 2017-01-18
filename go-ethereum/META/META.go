package META

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sync"
	
	//"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	//"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	METAapi "github.com/ethereum/go-ethereum/META/api"
	"github.com/ethereum/go-ethereum/META/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	
	//httpapi "github.com/ethereum/go-ethereum/swarm/api/http"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

// the meta stack
type META struct {
	config      *METAapi.Config            // meta configuration
	api         *METAapi.Api               // high level api layer 
	privateKey  *ecdsa.PrivateKey
	//server		*p2p.Server					// temporary pointer to server		
	protopeers	*network.PeerCollection				// protocol access to sending through peers, exposes the Messenger
	consolechan chan string
}

type METAAPI struct {
	Api     *METAapi.Api
	PrvKey  *ecdsa.PrivateKey
}

func (self *META) API() *METAAPI {
	return &METAAPI{
		Api:     self.api,
		PrvKey:  self.privateKey,
	}
}

// creates a new  meta service instance
// implements node.Service
func NewMETA(ctx *node.ServiceContext, config *METAapi.Config) (self *META, err error) {
	if bytes.Equal(common.FromHex(config.PublicKey), storage.ZeroKey) {
		return nil, fmt.Errorf("empty public key")
	}
	if bytes.Equal(common.FromHex(config.METAKey), storage.ZeroKey) {
		return nil, fmt.Errorf("empty meta key")
	}

	self = &META{
		config:      config,
		privateKey:  config.PrivateKey,
		
	}
	glog.V(logger.Debug).Infof("Setting up META service components")

	self.api = METAapi.NewApi()
	
	self.protopeers = network.NewPeerCollection()
	
	self.consolechan = make(chan string)
	
	return self, nil
}


/*
Start is called when the stack is started
*/
// implements the node.Service interface
func (self *META) Start(net *p2p.Server) error {
	/*connectPeer := func(url string) error {
		node, err := discover.ParseNode(url)
		if err != nil {
			return fmt.Errorf("invalid node URL: %v", err)
		}
		net.AddPeer(node)
		return nil
	}*/

	glog.V(logger.Warn).Infof("Starting META service")

	return nil
}

// implements the node.Service interface
// stops all component services.
func (self *META) Stop() error {

	return self.config.Save()
}

func (self *META) APIs() []rpc.API {
	return []rpc.API{
		// public APIs
		{
			Namespace: "mw",
			Version:   "0.1",
			Service:   METAapi.NewInfo(self.config),
			Public:    true,
		},
		{
			Namespace: "mw",
			Version:   "0.1",
			Service:   METAapi.NewParrotNode(self.protopeers, self.consolechan),
			Public:    true,
		},
		{
			Namespace: "mw",
			Version:   "0.1",
			Service:   METAapi.NewParrotCrowd(self.protopeers, self.consolechan),
			Public:    true,
		},
		{
			Namespace: "mw",
			Version:   "0.1",
			Service:   METAapi.NewPeerBroadcastSwitch(self.protopeers),
			Public:    true,
		},
		{
			Namespace: "mw",
			Version:   "0.1",
			Service:   METAapi.NewWhoAreYou(self.protopeers),
			Public:    true,
		},
	}
}

func (self *META) Protocols() []p2p.Protocol {
	wg := sync.WaitGroup{}
	return []p2p.Protocol{
		p2p.Protocol(network.METAProtocol1(self.protopeers, &wg, self.consolechan)),
		p2p.Protocol(network.METAProtocol2(self.protopeers)),
	}
}

// API reflection for RPC (?)



