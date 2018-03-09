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
	"encoding/json"
	"fmt"
	"strings"
	"swarmdb"
	"testing"
)

func TestConfig(t *testing.T) {
	config := swarmdb.GenerateSampleSWARMDBConfig("4b0d79af51456172dfcc064c1b4b8f45f363a80a434664366045165ba5217d53", "9982ad7bfbe62567287dafec879d20687e4b76f5", "wolkwolkwolk")
	err := swarmdb.SaveSWARMDBConfig(config, swarmdb.SWARMDBCONF_FILE)
	if err != nil {
		t.Fatal("Did not save config", err)
	}

	config2, err1 := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	if err1 != nil {
	}
	targ := `{"listenAddrTCP":"0.0.0.0","portTCP":2001,"listenAddrHTTP":"0.0.0.0","portHTTP":8501,"address":"9982ad7bfbe62567287dafec879d20687e4b76f5","privateKey":"4b0d79af51456172dfcc064c1b4b8f45f363a80a434664366045165ba5217d53","chunkDBPath":"/usr/local/swarmdb/data","usersKeysPath":"/usr/local/swarmdb/data/keystore","authentication":1,"users":[{"address":"9982ad7bfbe62567287dafec879d20687e4b76f5","passphrase":"wolkwolkwolk","minReplication":3,"maxReplication":5,"autoRenew":1}],"currency":"WLK","targetCostStorage":2.71828,"targetCostBandwidth":3.14159} {"listenAddrTCP":"127.0.0.1","portTCP":2000,"listenAddrHTTP":"127.0.0.1","portHTTP":8500,"address":"9982ad7bfbe62567287dafec879d20687e4b76f5","privateKey":"4b0d79af51456172dfcc064c1b4b8f45f363a80a434664366045165ba5217d53","chunkDBPath":"/swarmdb/data/keystore","authentication":1,"usersKeysPath":"/swarmdb/data/keystore","users":[{"address":"9982ad7bfbe62567287dafec879d20687e4b76f5","passphrase":"wolkwolkwolk","minReplication":3,"maxReplication":5,"autoRenew":1}],"currency":"WLK","targetCostStorage":2.71828,"targetCostBandwidth":3.14159}`

	cout, _ := json.Marshal(config2)
	if strings.Contains(string(cout), "wolkwolkwolk") {
		fmt.Printf("PASS Config: %s\n", cout)
	} else {
		t.Fatal("Mismatched output", string(cout), targ)
	}
}
