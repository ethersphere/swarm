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

package leveldb_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/ethersphere/swarm/storage/fcds/leveldb"
	"github.com/ethersphere/swarm/storage/fcds/test"
)

func TestFCDS(t *testing.T) {
	test.Test(t, func(t *testing.T) (fcds.Interface, func()) {
		path, err := ioutil.TempDir("", "swarm-fcds-")
		if err != nil {
			t.Fatal(err)
		}

		metaStore, err := leveldb.NewMetaStore(filepath.Join(path, "meta"))
		if err != nil {
			t.Fatal(err)
		}

		return test.NewFCDSStore(t, path, metaStore)
	})
}
