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

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/p2p/protocols"
)

// Spec is the protocol spec for bzzeth
var Spec = &protocols.Spec{
	Name:       "bzzeth",
	Version:    1,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		Handshake{},
		NewBlockHeaders{},
		GetBlockHeaders{},
		BlockHeaders{},
	},
	DisableContext: true,
}

// Handshake is used in between the ethereum node and the Swarm node
type Handshake struct {
	ServeHeaders bool // indicates if this node is expected to serve requests for headers
}

// NewBlockHeaders is sent from the Ethereum client to the Swarm node
type NewBlockHeaders []struct {
	Hash        common.Hash // block hash
	BlockHeight uint64      // block height
}

// GetBlockHeaders is used between a Swarm node and the Ethereum node in two cases:
// 1. When an Ethereum node asks the header corresponding to the hashes in the message (eth -> bzz)
// 2. When a Swarm node cannot find a particular header in the network, it asks the ethereum node for the header in order to push it to the network (bzz -> eth)
type GetBlockHeaders struct {
	Rid    uint32   // request id
	Hashes [][]byte // slice of hashes
}

// BlockHeaders encapsulates actual header blobs sent as a response to GetBlockHeaders
// multiple responses to the same request, whatever the node has it sends right away
type BlockHeaders struct {
	Rid     uint32         // request id
	Headers []rlp.RawValue // list of rlp encoded block headers
}
