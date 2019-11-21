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

package forky

import (
	"github.com/ethersphere/swarm/chunk"
)

type metaTrie struct {
	byte     byte
	value    *Meta
	branches []*metaTrie
}

func (t *metaTrie) get(addr chunk.Address) (m *Meta) {
	v := addr[0]
	for _, b := range t.branches {
		if b.byte == v {
			if len(addr) == 1 {
				return b.value
			}
			return b.get(addr[1:])
		}
	}
	return nil
}

func (t *metaTrie) set(addr chunk.Address, m *Meta) (overwritten bool) {
	x := t
	overwritten = true
	for _, v := range addr {
		i := branchIndex(v, x.branches)
		if i < 0 {
			i = len(x.branches)
			x.branches = append(x.branches, &metaTrie{
				byte: v,
			})
			overwritten = false
		}
		x = x.branches[i]
	}
	if overwritten {
		overwritten = x.value != nil
	}
	x.value = m
	return overwritten
}

func (t *metaTrie) remove(addr chunk.Address) (removed bool) {
	v := addr[0]
	l := len(addr)
	for _, b := range t.branches {
		if b.byte == v {
			if l == 1 {
				if b.value != nil {
					b.value = nil
					return true
				}
				return false
			}
			return b.remove(addr[1:])
		}
	}
	return false
}

func branchIndex(v byte, branches []*metaTrie) (i int) {
	for i, b := range branches {
		if b.byte == v {
			return i
		}
	}
	return -1
}
