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

package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"testing"

	swarmhttp "github.com/ethersphere/swarm/api/http"
)

// This function is called from test after a file is pinned.
// It check if the file's chunks are properly pinned.
// Assumption is that the file is uploaded in an empty database so that it can be easily tested.
// It takes the root hash and expected pinCounter value.
// It also has some hacks to take care of existing issues in the way we upload.
//
// The check process is as follows
//   1) Check if the root hash is pinned in pinFilesIndex
//         - GetPinnedFiles gives all the root chunks pinned in pinFilesIndex
//         - Check if this list contains our root hash
//   2) Check if all the files's chunks are in pinIndex
//         a) Get all the chunks
//            get it from retrievalDataIndex
//            since the assumption is the DB has only this file, it gives all the file's chunks.
//            getAllRefs cannot be used here as it does not give the chunks that belong to manifests.
//         b) Get all chunks that are pinned (from pinIndex)
//            In every upload.. an empty manifest is uploaded. that why add this hash to this list
//         c) Check if both the above lists are equal
//   3) Check if all the chunks pinned have the proper pinCounter
//         a) This is just a simple go through of all the pinned chunks list and check if the counter is
//            equal to the pin counter given as argument
//         b) The only hack in this is .. if the default path in the manifest is pointing to a file
//            then the pin counter of the chunks of that file will be pinned twice. This is taken care
//            by passing the root hash of that default file in defaultFileHash argument.
func checkIfPinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string, defaultFileHash []byte, pinCounter uint64, isRaw bool) {

	// 1 - Check if the root hash is pinned in pinFilesIndex
	pinnedFiles := srv.PinAPI.GetPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; !ok {
		t.Fatalf("File %s not pinned in pinFilesIndex", rootHash)
	}


	// 2a - Get all the chunks of the file from retrievalDataIndex (since the file is uploaded in an empty database)
	chunksInDB := srv.PinAPI.GetAllChunksFromDB()

	// 2b - Get pinned chunks details from pinning indexes
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash, "")
	if !isRaw {
		// Add the empty manifest chunk
		pinnedChunks["8b634aea26eec353ac0ecbec20c94f44d6f8d11f38d4578a4c207a84c74ef731"] = pinCounter
	}

	// 2c - Check if number of chunk hashes are same
	if len(chunksInDB) != len(pinnedChunks) {
		t.Fatalf("Expected number of chunks to be %d, but is %d", len(chunksInDB), len(pinnedChunks))
	}

	// 2c - Check if all the chunk address are same
	noOfChunksMissing := 0
	for hash := range chunksInDB {
		if _, ok := pinnedChunks[hash]; !ok {
			if !isRaw && noOfChunksMissing == 0 {
				noOfChunksMissing = 1
				continue
			}
			t.Fatalf("Expected chunk %s not present", hash)
		}
	}

	// 3 - Check for pin counter correctness
	if pinCounter != 0 {
		for hash, pc := range pinnedChunks {
			if pc != pinCounter {

				foundChunk := false

				// 3b - hack for default file in manifest
				// If "default path" is pointing to a file in the manifest...
				// that file's chunks would have been pinned twice
				if defaultFileHash != nil {

					defaultFilehash, err := srv.FileStore.GetAllReferences(context.Background(), bytes.NewReader(defaultFileHash))
					if err != nil {
						t.Fatal(err)
					}

					for _, defaultFileChunk := range defaultFilehash {
						if hash == hex.EncodeToString(defaultFileChunk) && pc == pinCounter+1 {
							foundChunk = true
							break
						}
					}
				}

				if !foundChunk {
					t.Fatalf("Expected pin counter %d got %d", pinCounter, pc)
				}
			}
		}
	}
}

func isNoChunksPinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string) {

	// Get pinned chunks details from pinning indexes
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash, "")

	// Check if number of chunk hashes are same
	if len(pinnedChunks) != 0 {
		t.Fatalf("Expected empty pinIndex but %d chunks found", len(pinnedChunks))
	}
}

func checkIfUnpinned(t *testing.T, srv *swarmhttp.TestSwarmServer, rootHash string) {

	// root hash should not be in pinFilesIndex
	pinnedFiles := srv.PinAPI.GetPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; ok {
		t.Fatalf("File %s pinned in pinFilesIndex", rootHash)
	}

	// The chunks of this file should not be in pinIndex too
	pinnedChunks := srv.PinAPI.CollectPinnedChunks(rootHash, "")
	if len(pinnedChunks) != 0 {
		t.Fatalf("Chunks of this file present in pinIndex")
	}
}
