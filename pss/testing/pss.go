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

package testing

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/localstore"
)

// NewMockLocalStore generates a mocked localstore
// this is used by test that need access to the localstore
// the tags are initialized on store init
func NewMockLocalStore(t *testing.T, tags *chunk.Tags) *localstore.DB {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{Tags: tags})
	if err != nil {
		t.Fatal(err)
	}

	return localStore
}

// TODO: later test could be a simulation test for 2 nodes, localstore + netstore
