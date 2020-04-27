// Copyright 2020 The Swarm Authors
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

package prod

import (
	"context"

	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
)

// Sender defines code to be executed upon trigger of lost TODO: chunk????
type Sender func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error)

// Recover invokes underlying pss.Send as the first step of global pinning
func Recover(send Sender) {

}

// func Register?
// based of topic?
// maybe the sender
