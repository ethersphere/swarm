package orbit

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/state"
)

// Enumerate options for syncing and retrieval
type SyncingOption int

// Syncing options
const (
	// Syncing disabled
	SyncingDisabled SyncingOption = iota
	// Register the client and the server but not subscribe
	SyncingRegisterOnly
	// Both client and server funcs are registered, subscribe sent automatically
	SyncingAutoSubscribe
)

// Orb implements node.Service
var _ = &Orb{}.(node.Service)

// Registry registry for outgoing and incoming streamer constructors
type Orb struct {
	addr           enode.ID
	peers          map[enode.ID]*Peer
	intervalsStore state.Store
	maxPeerServers int
	spec           *protocols.Spec   //this protocol's spec
	balance        protocols.Balance //implements protocols.Balance, for accounting
	prices         protocols.Prices  //implements protocols.Prices, provides prices to accounting
	quit           chan struct{}     // terminates registry goroutines
	syncMode       SyncingOption
}

func NewOrb(me enode.ID, intervalsStore state.Store) *Orb {
	orb := &Orb{
		addr:           me,
		peers:          make(map[enode.ID]*Peer),
		intervalsStore: intervalsStore,
		maxPeerServers: 16,
		quit:           make(chan struct{}),
	}

	return orb
}

func (o *Orb) entryPoint(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := protocols.NewPeer(p, rw, r.spec)
	bp := network.NewBzzPeer(peer)
	np := network.NewPeer(bp, o.kad)
	o.kad.On(np)
	defer o.kad.Off(np)
	return o.peerLoop(bp)
}

func (o *Orb) peerLoop(peer *network.BzzPeer) error {

}

func (o *Orb) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "orb",
			Version: "einz",
			Length:  10 * 1024 * 1024,
			Run:     o.entryPoint,
		},
	}
}

func (r *Orb) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "orb",
			Version:   "1.0",
			Service:   nil,
			Public:    false,
		},
	}
}

func (r *Orb) Start(server *p2p.Server) error {
	log.Info("started getting this done")
	return nil
}

func (r *Orb) Stop() error {
	log.Info("shutting down")
	return nil
}
