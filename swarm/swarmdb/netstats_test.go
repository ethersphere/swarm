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
	"math/rand"
	"swarmdb"
	"swarmdb/ash"
	"testing"
	"time"
)

var (
	testDBPath = "chunks.db"
	chunkTotal = 200
)

func TestNetstatsBasic(t *testing.T) {
	//General Connection
	config, _ := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	swarmdb.NewKeyManager(config)
	u := config.GetSWARMDBUser()
	netstats := swarmdb.NewNetstats(config)

	store, err := swarmdb.NewDBChunkStore(config, netstats)
	if err != nil {
		fmt.Printf("%s", err)
		t.Fatal("[FAILURE] to open DBChunkStore\n")
	} else {
		fmt.Printf("[SUCCESS] open DBChunkStore\n")
	}
	ts := int64(time.Now().Unix())
	t.Run("Write=0", func(t *testing.T) {
		//Simulate chunk writes w/ n chunkTotal
		for j := 0; j < chunkTotal; j++ {
			simdata := make([]byte, 4096)
			tmp := fmt.Sprintf("%s%d", "randombytes", j)
			copy(simdata, tmp)
			enc := rand.Intn(2)
			simh, err := store.StoreChunk(u, simdata, enc)
			if err != nil {
				t.Fatal("[FAILURE] writting record #%v [%x] => %v\n", j, simh, string(simdata[:]))
			} else if j%5 == 0 {
				fmt.Printf("Generating record [%x] => %v ... ", simh, string(simdata[:]))
				fmt.Printf("[SUCCESS] writing #%v chunk | Encryption: %v\n", j, enc)
			}
			secret := make([]byte, 32)
			rand.Read(secret)
			proofRequired := rand.Intn(2) != 0
			auditIndex := rand.Intn(128)

			response, err := store.RetrieveAsh(simh, secret, proofRequired, int8(auditIndex))
			if err != nil {
				t.Fatal("[FAILURE] Generating record [%x] %s\n", simh, err.Error())
			} else if j%5 == 0 {
				if proofRequired {
					ok, mr, err := ash.CheckProof(response.Proof.Root, response.Proof.Path, response.Proof.Index)
					if err == nil {
						fmt.Printf("Proof Verified: %t | Root: %x\n", ok, mr)
					} else {
						t.Fatal(err.Error())
					}
				}
				if j%5 == 0 {
					output, _ := json.Marshal(response)
					fmt.Printf("ProofRequired: %t | Index: %d | Seed: [%x]\n", proofRequired, auditIndex, secret)
					fmt.Printf("Generating record [%x]\n %v\n\n", simh, string(output))
				}
			}
		}
		err = netstats.Save()
		if err != nil {
			t.Fatalf("%s\n", err)
		}
	})

	t.Run("EFarmLog=1", func(t *testing.T) {
		log, err := store.GenerateFarmerLog(ts, ts+60)
		if err != nil {
			t.Fatal("[FAILURE] Farmer log Error\n")
		} else {
			fmt.Printf("\n%s\n", log)
			fmt.Printf("[SUCCESS] Farmer Operation completed\n")
		}
		err = netstats.Save()
		if err != nil {
			t.Fatalf("%s\n", err)
		}
	})
}
