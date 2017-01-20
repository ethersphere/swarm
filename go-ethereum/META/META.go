package META

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sync"
	
	//"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	//"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	METAapi "github.com/ethereum/go-ethereum/META/api"
	"github.com/ethereum/go-ethereum/META/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	
	//httpapi "github.com/ethereum/go-ethereum/swarm/api/http"
)




// the meta stack
type META struct {
	config      *METAapi.Config            // meta configuration
	api         *METAapi.Api               // high level api layer 
	privateKey  *ecdsa.PrivateKey
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
	
	// set up high level api
	// we probably need this, but comment out for now
	// seems to have to do with contracts
	//transactOpts := bind.NewKeyedTransactor(self.privateKey)

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
	}
}

func (self *META) Protocols() []p2p.Protocol {
	wg := sync.WaitGroup{}
	return []p2p.Protocol{p2p.Protocol(network.METAProtocol(&wg))}
}

// API reflection for RPC (?)



