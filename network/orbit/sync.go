package orbit

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
)

// Orb implements node.Service
var _ node.Service = (*Orb)(nil)
var pollTime = 1 * time.Second

// Orb is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type Orb struct {
	mtx            sync.RWMutex
	addr           enode.ID
	intervalsStore state.Store //every protocol would make use of this
	peers          map[enode.ID]*Peer
	netStore       *storage.NetStore
	kad            *network.Kademlia
	started        bool
	//getPeer        func(enode.ID) *Peer

	spec    *protocols.Spec   //this protocol's spec
	balance protocols.Balance //implements protocols.Balance, for accounting
	prices  protocols.Prices  //implements protocols.Prices, provides prices to accounting

	quit chan struct{} // terminates registry goroutines
}

func NewOrb(me enode.ID, intervalsStore state.Store, kad *network.Kademlia, ns *storage.NetStore) *Orb {
	orb := &Orb{
		addr:           me,
		intervalsStore: intervalsStore,
		peers:          make(map[enode.ID]*Peer),
		kad:            kad,
		netStore:       ns,
		quit:           make(chan struct{}),
	}

	var spec = &protocols.Spec{
		Name:       "orb",
		Version:    8,
		MaxMsgSize: 10 * 1024 * 1024,
		Messages: []interface{}{
			StreamMsg{},
		},
	}
	orb.spec = spec

	return orb
}

func (o *Orb) addPeer(p *Peer) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	o.peers[p.ID()] = p
}

func (o *Orb) removePeer(p *Peer) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if _, found := o.peers[p.ID()]; found {
		delete(o.peers, p.ID())
		p.Left()

	} else {
		log.Warn("peer was marked for removal but not found")
		panic("shouldnt happen")
	}
}

func (o *Orb) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, o.spec)
	bp := network.NewBzzPeer(peer)
	np := network.NewPeer(bp, o.kad)
	o.kad.On(np)
	defer o.kad.Off(np)

	sp := NewPeer(bp)
	o.addPeer(sp)
	defer o.removePeer(sp)
	go func() {
		time.Sleep(50 * time.Millisecond)
		if err := sp.Send(context.TODO(), StreamMsg{}); err != nil {
			log.Error("err sending", "err", err)
		}

	}()

	return peer.Run(sp.HandleMsg)
}

func (o *Orb) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "orb",
			Version: 1,
			Length:  10 * 1024 * 1024,
			Run:     o.Run,
		},
	}
}

func (r *Orb) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "orb",
			Version:   "1.0",
			Service:   NewAPI(r),
			Public:    false,
		},
	}
}

// Additional public methods accessible through API for pss
type API struct {
	*Orb
}

func NewAPI(o *Orb) *API {
	return &API{Orb: o}
}

func (o *Orb) Start(server *p2p.Server) error {
	log.Info("started getting this done")
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.started {
		panic("shouldnt happen")
	}
	o.started = true
	go func() {
		o.started = true
		for {
			select {
			case <-o.quit:
				return
			case <-time.After(pollTime):
				// go over each peer and for each subprotocol check that each stream is working
				// i.e. for each peer, for each subprotocol, a client should be created (with an infinite loop)
				// fetching the stream
			}
		}
	}()
	return nil
}

func (o *Orb) Stop() error {
	log.Info("shutting down")
	o.mtx.Lock()
	defer o.mtx.Unlock()
	close(o.quit)
	return nil
}
