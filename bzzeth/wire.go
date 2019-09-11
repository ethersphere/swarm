// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package bzzeth

import "github.com/ethersphere/swarm/p2p/protocols"

// Spec is the protocol spec for bzzeth
var Spec = &protocols.Spec{
	Name:       "bzzeth",
	Version:    1,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		Handshake{},
		DummyMessage{},
	},
	DisableContext: true,
}

// Handshake is used in between the ethereum node and the Swarm node
type Handshake struct {
	ServeHeaders bool // indicates if this node is expected to serve requests for headers
}

// Dummy message to send from another Swarm node
// Used only in testing
type DummyMessage struct {
	test bool
}
