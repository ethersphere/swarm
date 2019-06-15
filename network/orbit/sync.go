package orbit

import (
	"context"
	"time"

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
//var _ = &Orb{}.(node.Service)

// Orb is the base type that handles all client/server operations on a node
// it is instantiated once per stream protocol instance, that is, it should have
// one instance per node
type Orb struct {
	addr           enode.ID
	intervalsStore state.Store //every protocol would make use of this
	peers          map[enode.ID]*Peer
	spec           *protocols.Spec   //this protocol's spec
	balance        protocols.Balance //implements protocols.Balance, for accounting
	prices         protocols.Prices  //implements protocols.Prices, provides prices to accounting

	netStore *storage.NetStore
	kad      *network.Kademlia
	getPeer  func(enode.ID) *Peer

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
	o.peers = append(o.peers, p)
}

func (o *Orb) removePeer(p *Peer) {
	for i, v := range o.peers {
		if v == p {
			o.peers = append(o.peers[:i], o.peers[i+1:])
		}
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
	defer sp.Left()
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

func (r *Orb) Start(server *p2p.Server) error {
	log.Info("started getting this done")
	return nil
}

func (r *Orb) Stop() error {
	log.Info("shutting down")
	return nil
}
