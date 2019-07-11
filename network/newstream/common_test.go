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

package newstream

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/mock"
)

var (
	loglevel = flag.Int("loglevel", 5, "verbosity of logs")
)

func init() {
	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

func newTestLocalStore(id enode.ID, addr *network.BzzAddr, globalStore mock.GlobalStorer) (localStore *localstore.DB, cleanup func(), err error) {
	dir, err := ioutil.TempDir(tmpDir, "localstore-")
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		os.RemoveAll(dir)
	}

	var mockStore *mock.NodeStore
	if globalStore != nil {
		mockStore = globalStore.NewNodeStore(common.BytesToAddress(id.Bytes()))
	}

	localStore, err = localstore.New(dir, addr.Over(), &localstore.Options{
		MockStore: mockStore,
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return localStore, cleanup, nil
}

// Test run global tmp dir. Please, use it as the first argument
// to ioutil.TempDir function calls in this package tests.
var tmpDir string

func TestMain(m *testing.M) {
	// Tests in this package generate a lot of temporary directories
	// that may not be removed if tests are interrupted with SIGINT.
	// This function constructs a single top-level directory to be used
	// to store all data from a test execution. It removes the
	// tmpDir with defer, or by catching keyboard interrupt signal,
	// so that all data will be removed even on forced termination.

	var err error
	tmpDir, err = ioutil.TempDir("", "swarm-stream-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	go func() {
		first := true
		for range c {
			fmt.Fprintln(os.Stderr, "signal: interrupt")
			if first {
				fmt.Fprintln(os.Stderr, "removing swarm stream tmp directory", tmpDir)
				os.RemoveAll(tmpDir)
				os.Exit(1)
			}
		}
	}()
	os.Exit(m.Run())
}
