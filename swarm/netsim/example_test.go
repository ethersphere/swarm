// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package netsim_test

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/swarm/netsim"
)

func ExampleSimulation_PeerEvents() {
	sim := netsim.NewSimulation(nil, &netsim.SimulationOptions{})
	defer sim.Close()

	events := sim.PeerEvents(context.Background(), sim.NodeIDs())

	go func() {
		for e := range events {
			if e.Error != nil {
				log.Error("peer event", "err", e.Error)
				continue
			}
			log.Info("peer event", "node", e.NodeID, "peer", e.Event.Peer, "msgcode", e.Event.MsgCode)
		}
	}()
}

func ExampleSimulation_PeerEvents_Disconnections() {
	sim := netsim.NewSimulation(nil, &netsim.SimulationOptions{})
	defer sim.Close()

	disconnections := sim.PeerEvents(
		context.Background(),
		sim.NodeIDs(),
		netsim.NewPeerEventsFilter().Type(p2p.PeerEventTypeDrop),
	)

	go func() {
		for d := range disconnections {
			if d.Error != nil {
				log.Error("peer drop", "err", d.Error)
				continue
			}
			log.Warn("peer drop", "node", d.NodeID, "peer", d.Event.Peer)
		}
	}()
}

func ExampleSimulation_PeerEvents_MultipleFilters() {
	sim := netsim.NewSimulation(nil, &netsim.SimulationOptions{})
	defer sim.Close()

	msgs := sim.PeerEvents(
		context.Background(),
		sim.NodeIDs(),
		// Watch when bzz messages 1 and 4 are received.
		netsim.NewPeerEventsFilter().Type(p2p.PeerEventTypeMsgRecv).Protocol("bzz").MsgCode(1),
		netsim.NewPeerEventsFilter().Type(p2p.PeerEventTypeMsgRecv).Protocol("bzz").MsgCode(4),
	)

	go func() {
		for m := range msgs {
			if m.Error != nil {
				log.Error("bzz message", "err", m.Error)
				continue
			}
			log.Info("bzz message", "node", m.NodeID, "peer", m.Event.Peer)
		}
	}()
}
