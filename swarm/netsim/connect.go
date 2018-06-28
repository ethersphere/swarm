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
	"strings"

	"github.com/ethereum/go-ethereum/p2p/discover"
)

func (s *Simulation) ConnectToPivotNode(id discover.NodeID) (err error) {
	pid := s.PivotNodeID()
	if pid == nil {
		return ErrNoPivotNode
	}
	return s.connect(*pid, id)
}

func (s *Simulation) ConnectToLastNode(id discover.NodeID) (err error) {
	ids := s.UpNodeIDs()
	l := len(ids)
	if l < 2 {
		return ErrNodeNotFound
	}
	lid := ids[l-1]
	if lid == id {
		lid = ids[l-2]
	}
	return s.connect(lid, id)
}

func (s *Simulation) ConnectToRandomNode(id discover.NodeID) (err error) {
	n := s.randomNode(id)
	if n == nil {
		return ErrNodeNotFound
	}
	return s.connect(n.ID, id)
}

func (s *Simulation) ConnectNodesFull() (err error) {
	ids := s.UpNodeIDs()
	l := len(ids)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			err = s.connect(ids[i], ids[j])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Simulation) ConnectNodesChain() (err error) {
	ids := s.UpNodeIDs()
	l := len(ids)
	for i := 0; i < l-1; i++ {
		err = s.connect(ids[i], ids[i+1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Simulation) ConnectNodesRing() (err error) {
	ids := s.UpNodeIDs()
	l := len(ids)
	if l < 2 {
		return nil
	}
	for i := 0; i < l-1; i++ {
		err = s.connect(ids[i], ids[i+1])
		if err != nil {
			return err
		}
	}
	return s.connect(ids[l-1], ids[0])
}

func (s *Simulation) ConnectNodesStar(id discover.NodeID) (err error) {
	ids := s.UpNodeIDs()
	l := len(ids)
	for i := 0; i < l; i++ {
		if id == ids[i] {
			continue
		}
		err = s.connect(id, ids[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Simulation) ConnectNodesStarPivot() (err error) {
	id := s.PivotNodeID()
	if id == nil {
		return ErrNoPivotNode
	}
	return s.ConnectNodesStar(*id)
}

func (s *Simulation) connect(oneID, otherID discover.NodeID) error {
	return ignoreAlreadyConnectedErr(s.Net.Connect(oneID, otherID))
}

func ignoreAlreadyConnectedErr(err error) error {
	if err == nil || strings.Contains(err.Error(), "already connected") {
		return nil
	}
	return err
}
