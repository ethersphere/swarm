// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swarmdb_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"swarmdb"
	"testing"
)

func TestIssueReceive(t *testing.T) {
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	ns := swarmdb.NewNetstats(config)
	swapdbstore, err := swarmdb.NewSwapDBStore(config, ns)
	if err != nil {
		t.Fatal("Failure to open NewSwapDBStore")
	}

	localAddress := common.HexToAddress("9982ad7bfbe62567287dafec879d20687e4b76f5")
	peerAddress := common.HexToAddress("0082ad7bfbe62567287dafec879d20687e4b76aa")
	amount := 17

	// Test Issue
	ch, err := swapdbstore.Issue(amount, localAddress, peerAddress)
	if err != nil {
		t.Fatalf("[swapdb_test:TestIssueReceive] Issue %s", err.Error())
	}
	fmt.Printf("Issued check %v\n", ch)

	// TODO: make a check signed by peer to test this correctly
	// Test Receive
	err = swapdbstore.Receive(-amount, ch)
	if err != nil {
		t.Fatalf("[swapdb_test:TestIssueReceive] Receive %s", err.Error())
	}
	fmt.Printf("Received check %v\n", ch)
	// TODO: test Receive with an *incorrect* signature

	// Test GenerateSwapLog
	startts := int64(0)
	endts := int64(1)
	log, err := swapdbstore.GenerateSwapLog(startts, endts)
	if err != nil {
		t.Fatalf("[swapdb_test:TestIssueReceive] GenerateSwapLog %s", err.Error())
	}
	fmt.Printf("Log: %s\n", log)
	// TODO: test that GenerateSwapLog has the above issue and receive
}
