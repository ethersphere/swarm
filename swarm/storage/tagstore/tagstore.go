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

package tagstore

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/state"
)

// DB implements chunk.Store.
var _ chunk.TagStore = &DB{}

var (
	// Default value for Capacity DB option.
	defaultCapacity uint64 = 5000000
)

// DB is the tag store implementation and holds
// database related objects.
type DB struct {
	// state store ref
	store state.Store

	// random number generator
	rng *rand.Rand

	batchMu sync.Mutex

	// this channel is closed when close function is called
	// to terminate other goroutines
	close chan struct{}
}

// New returns a new DB.  All fields and indexes are initialized
// and possible conflicts with schema from existing database is checked.
// One goroutine for writing batches is created.
func New(store state.Store) (db *DB, err error) {
	if store == nil {
		return nil, errors.New("provided store is nil, expecting a state.Store")
	}

	// set the store to the state store
	db = &DB{
		store: store,
	}

	// initialise the random number generator
	db.rng = rand.New(rand.NewSource(time.Now().Unix()))
	return db, nil
}

// Close closes the underlying database.
func (db *DB) Close() (err error) {
	close(db.close)

	return nil
}

// now is a helper function that returns a current unix timestamp
// in UTC timezone.
// It is set in the init function for usage in production, and
// optionally overridden in tests for data validation.
var now func() int64

func init() {
	// set the now function
	now = func() (t int64) {
		return time.Now().UTC().UnixNano()
	}
}
