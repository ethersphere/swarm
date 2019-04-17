// Copyright 2019 The go-ethereum Authors
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

package chunk

import (
	"math/rand"
	"sync"
	"time"
)

// tags holds the tag infos indexed by name
type tags struct {
	tags *sync.Map
	rng  *rand.Rand
}

// NewTags creates a tags object
func newTags() *tags {

	return &tags{
		tags: &sync.Map{},
		rng:  rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// New creates a new tag, stores it by the name and returns it
// it returns an error if the tag with this name already exists
func (ts *tags) New(s string, total int) (*Tag, error) {
	t := &Tag{
		uid:       ts.rng.Uint32(),
		Name:      s,
		startedAt: time.Now(),
		total:     uint32(total),
	}
	if _, loaded := ts.tags.LoadOrStore(s, t); loaded {
		return nil, errExists
	}
	return t, nil
}

// Inc increments the state count for a tag if tag is found
func (ts *tags) Inc(s string, f State) {
	t, ok := ts.tags.Load(s)
	if !ok {
		return
	}
	t.(*Tag).Inc(f)
}

// Get returns the state count for a tag
func (ts *tags) Get(s string, f State) int {
	t, _ := ts.tags.Load(s)
	return t.(*Tag).Get(f)
}
