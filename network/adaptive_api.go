package network

import (
	"fmt"

	"github.com/ethersphere/swarm/log"
)

// CapabilitiesAPI abstracts RPC API access to capabilities controls
// will in the future provide notifications of capability changes
type CapabilitiesAPI struct {
	*Capabilities
}

// NetCapabilitiesAPI creates new API abstraction around provided Capabilities object
func NewCapabilitiesAPI(c *Capabilities) *CapabilitiesAPI {
	return &CapabilitiesAPI{
		Capabilities: c,
	}
}

// RegisterCapability adds the given capability object to the Capabilities collection
// If the Capability is already registered an error will be returned
func (a *CapabilitiesAPI) RegisterCapability(cp *Capability) error {
	log.Debug("Registering capability", "cp", cp)
	return a.add(cp)
}

// IsRegisteredCapability returns true if a Capability with the given id is registered
func (a *CapabilitiesAPI) IsRegisteredCapability(id CapabilityID) (bool, error) {
	return a.get(id) != nil, nil
}

// MatchCapability returns true if the Capability flag at the given index is set
// Fails with error if the Capability is not registered, or if the index is out of bounds
func (a *CapabilitiesAPI) MatchCapability(id CapabilityID, idx int) (bool, error) {
	c := a.get(id)
	if c == nil {
		return false, fmt.Errorf("Capability %d not registered", id)
	} else if idx > len(c.Cap)-1 {
		return false, fmt.Errorf("Capability %d idx %d out of range", id, idx)
	}
	return c.Cap[idx], nil
}
