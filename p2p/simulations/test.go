package simulations

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"
)

// NoopService is the service that does not do anything
// but implements node.Service interface.
type NoopService struct {
	c map[enode.ID]chan struct{}
}

func NewNoopService(ackC map[enode.ID]chan struct{}) *NoopService {
	return &NoopService{
		c: ackC,
	}
}

func (t *NoopService) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "noop",
			Version: 666,
			Length:  0,
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				if t.c != nil {
					t.c[peer.ID()] = make(chan struct{})
					close(t.c[peer.ID()])
				}
				rw.ReadMsg()
				return nil
			},
			NodeInfo: func() interface{} {
				return struct{}{}
			},
			PeerInfo: func(id enode.ID) interface{} {
				return struct{}{}
			},
			Attributes: []enr.Entry{},
		},
	}
}

func (t *NoopService) APIs() []rpc.API {
	return []rpc.API{}
}

func (t *NoopService) Start(server *p2p.Server) error {
	return nil
}

func (t *NoopService) Stop() error {
	return nil
}
