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

package netsim

import (
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
)

func (s *Simulation) Service(id discover.NodeID) node.Service {
	simNode, ok := s.Net.GetNode(id).Node.(*adapters.SimNode)
	if !ok {
		return nil
	}
	services := simNode.Services()
	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (s *Simulation) RandomService() node.Service {
	n := s.randomNode()
	if n == nil {
		return nil
	}
	services := n.Services()
	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (s *Simulation) Services() (services []node.Service) {
	nodes := s.Net.GetNodes()
	for _, node := range nodes {
		if !node.Up {
			continue
		}
		simNode, ok := node.Node.(*adapters.SimNode)
		if !ok {
			continue
		}
		nss := simNode.Services()
		if len(nss) == 0 {
			continue
		}
		services = append(services, nss[0])
	}
	return services
}
