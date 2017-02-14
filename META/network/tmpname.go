package network

import (
	"fmt"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/p2p/adapters"
)

type METATmpName struct {
	METAHeader
	Name string
	Swarmhash storage.Key
	Node adapters.NodeId
}

func NewMETATmpName() (mtn *METATmpName) {
	mtn = &METATmpName{
		METAHeader: NewMETAEnvelope(),
		Node: adapters.NodeId{},
	}
	return
}

func (mtn *METATmpName) AsString() string {
	return fmt.Sprintf("METATmpName '%s' is node pointing to swarmhash '%v'", mtn.Name, mtn.Swarmhash) 
}

