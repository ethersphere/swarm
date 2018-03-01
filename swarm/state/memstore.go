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

package state

import (
	"encoding"
	"sync"
)

// MemStore is the reference implementation of Store interface that is supposed
// to be used in tests.
type MemStore struct {
	db map[string][]byte
	mu sync.RWMutex
}

// NewMemStore returns a new instance of MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		db: make(map[string][]byte),
	}
}

// Get retrieves Intervals for a specific key. If there is no Intervals
// ErrNotFound is returned.
func (s *MemStore) Get(key string, i interface{}) (err error) {
	_, ok := i.(encoding.BinaryUnmarshaler)
	if !ok {
		return ErrInvalidArgument
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bytes, ok := s.db[key]
	if !ok {
		return ErrNotFound
	}

	return i.(encoding.BinaryUnmarshaler).UnmarshalBinary(bytes)
}

// Put stores Intervals for a specific key.
func (s *MemStore) Put(key string, i interface{}) (err error) {
	_, ok := i.(encoding.BinaryMarshaler)
	if !ok {
		return ErrInvalidArgument
	}

	bytes, err := i.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = bytes
	return nil
}

// Delete removes Intervals stored under a specific key.
func (s *MemStore) Delete(key string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.db[key]; !ok {
		return ErrNotFound
	}
	delete(s.db, key)
	return nil
}

// Close doesnot do anything.
func (s *MemStore) Close() error {
	return nil
}
