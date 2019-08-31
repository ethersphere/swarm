// Copyright 2019 The Swarm authors
// This file is part of the swarm library.
//
// The swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the swarm library. If not, see <http://www.gnu.org/licenses/>.

package capability

import (
	"fmt"

	"github.com/ethersphere/swarm/log"
)

// API abstracts RPC API access to capabilities controls
// will in the future provide notifications of capability changes
type API struct {
	*Capabilities
}

// NetAPI creates new API abstraction around provided Capabilities object
func NewAPI(c *Capabilities) *API {
	return &API{
		Capabilities: c,
	}
}

// RegisterCapability adds the given capability object to the Capabilities collection
// If the Capability is already registered an error will be returned
func (a *API) RegisterCapability(cp *Capability) error {
	log.Debug("Registering capability", "cp", cp)
	return a.Add(cp)
}

// IsRegisteredCapability returns true if a Capability with the given id is registered
func (a *API) IsRegisteredCapability(id CapabilityID) (bool, error) {
	return a.Get(id) != nil, nil
}

// MatchCapability returns true if the Capability flag at the given index is set
// Fails with error if the Capability is not registered, or if the index is out of bounds
func (a *API) MatchCapability(id CapabilityID, idx int) (bool, error) {
	c := a.Get(id)
	if c == nil {
		return false, fmt.Errorf("Capability %d not registered", id)
	}
	if idx > len(c.Cap)-1 {
		return false, fmt.Errorf("Capability %d idx %d out of range", id, idx)
	}
	return c.Cap[idx], nil
}
