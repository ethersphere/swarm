package swarmdb

import (
	"fmt"

	"github.com/ethereum/go-ethereum/p2p"
)

type API struct {
	*SwarmDB
}

func NewAPI(swdb *SwarmDB) *API {
	return &API{SwarmDB: swdb}
}

/////////////////////////////////////////////////////////////////////
// SECTION: node.Service interface
/////////////////////////////////////////////////////////////////////

func (self *SwarmDB) Start(srv *p2p.Server) error {
	return nil
}

func (self *SwarmDB) Stop() error {
	return nil
}

func (self *SwarmDB) Protocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "swarmdb",
			Version: 1, //TODO: SWARMDBVersion
			Length:  1,
			Run:     self.Run,
		},
	}
}

//Run when another node connects
func (self *SwarmDB) Run(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	//Taken from: https://github.com/nolash/go-ethereum-p2p-demo/blob/master/A4_Message.go

	// simplest payload possible; a byte slice
	outmsg := "foobar"

	// send the message
	err := p2p.Send(rw, 0, outmsg)
	if err != nil {
		return fmt.Errorf("Send p2p message fail: %v", err)
	}
	demo.Log.Info("sending message", "peer", p, "msg", outmsg)

	// wait for the message to come in from the other side
	// note that receive message event doesn't get emitted until we ReadMsg()
	inmsg, err := rw.ReadMsg()
	if err != nil {
		return fmt.Errorf("Receive p2p message fail: %v", err)
	}
	demo.Log.Info("received message", "peer", p, "msg", inmsg)

	// terminate the protocol
	return nil
}
