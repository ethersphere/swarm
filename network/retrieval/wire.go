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

package retrieval

import "github.com/ethersphere/swarm/storage"

// RetrieveRequest is the protocol msg for chunk retrieve requests
type RetrieveRequest struct {
	Ruid  uint            // unique identifier, to protect agains unsollicited chunks
	Price uint            // the best-effort price of the requested ChunkDelivery
	Addr  storage.Address // the address of the requested chunk
}

// ChunkDelivery is the protocol msg for delivering a solicited chunk to a peer
type ChunkDelivery struct {
	Ruid  uint            // unique identifier, to protect agains unsollicited chunks
	price uint            // the agreed-upon price of the ChunkDelivery
	Addr  storage.Address // the address of the chunk
	SData []byte          // the chunk
}
