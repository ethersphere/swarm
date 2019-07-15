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

package testutil

import (
	"fmt"
	swarmhttp "github.com/ethersphere/swarm/api/http"
	"testing"
)



func CheckIfPinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string, data []byte, pinCounter uint64, isRaw bool) {


	//fmt.Println(rootHash)
	//pf := srv.PinAPI.GetPinnedFiles()
	//for k,v := range pf {
	//	fmt.Println(k,v)
	//}

	// Check if the root hash is in the pinFilesIndex
	pinnedFiles := srv.PinAPI.GetPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; !ok {
		t.Fatalf("File %s not pinned in pinFilesIndex", rootHash)
	}

	// Get all the chunks from DB
	//chunksInDB, err := srv.FileStore.GetAllReferences(context.Background(), bytes.NewReader(data))
	//if err != nil {
	//	t.Fatal(err)
	//}
	chunksInDB := srv.PinAPI.GetAllChunksFromDB()
	//for k, _ := range chunksInDB {
	//	fmt.Println("Chunks in DB", k)
	//}


	// Get pinned chunks details from pinning indexes
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash,"")
	//for k,v := range pinnedChunks {
	//	fmt.Println("Pinned chunks", k,v)
	//}

	if !isRaw {
		// Add the empty manifest chunk
		pinnedChunks["8b634aea26eec353ac0ecbec20c94f44d6f8d11f38d4578a4c207a84c74ef731"] = pinCounter
	}


	// Check if number of chunk hashes are same
	if len(chunksInDB) != len(pinnedChunks) {
		t.Fatalf("Expected number of chunks to be %d, but is %d", len(chunksInDB), len(pinnedChunks))
	}


	// don't check for chunk correctness for encrypted files as the encryption key is random
	//if !isEncrypted {
		// Check if all the chunk address are same
		noOfChunksMissing := 0
		for hash,_ := range chunksInDB {
			if _, ok := pinnedChunks[hash]; !ok {
				if !isRaw && noOfChunksMissing == 0{
					noOfChunksMissing = 1
					continue
				}
				t.Fatalf("Expected chunk %s not present", hash)
			}
		}
	//}

	// Check for pin counter correctness
	if pinCounter != 0 {
		for _, pc := range pinnedChunks {
			if pc != pinCounter {
				t.Fatalf("Expected pin counter %d got %d", pinCounter, pc)
			}
		}
	}

}

func IsNoChunksPinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string) {

	// Get pinned chunks details from pinning indexes
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash,"")

	// Check if number of chunk hashes are same
	if len(pinnedChunks) != 0 {
		t.Fatalf("Expected empty pinIndex but %d chunks found", len(pinnedChunks))
	}
}


func CheckIfUnpinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string) {

	// root hash should not be in pinFilesIndex
	pinnedFiles := srv.PinAPI.GetPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; ok {
		t.Fatalf("File %s pinned in pinFilesIndex", rootHash)
	}

	// The chunks of this file should not be in pinIndex too
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash,"")
	if len(pinnedChunks) != 0 {
		t.Fatalf("Chunks of this file present in pinIndex")
	}
}


func PrintPinIndexes(srv *swarmhttp.TestSwarmServer) {

	pinnedFiles := srv.PinAPI.GetPinnedFiles()
	for fileName,treeSize := range pinnedFiles {
		fmt.Println("roothash = ", fileName,"treeSize = ", treeSize)
	}

	pinnedChunks := srv.PinAPI.GetPinnedChunks()
	for chunkHash, pinCounter := range pinnedChunks {
		fmt.Println("chunkHash = ", chunkHash, "pinCounter = ", pinCounter)
	}

}