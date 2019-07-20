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

package pin

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/ethersphere/swarm/storage/feed"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/testutil"
)

// Pin a RAW file and unpin it multiple times
func TestPinRawUpload(t *testing.T) {

	p, f := getPinApiAndFileStore(t)
	data := testutil.RandomBytes(1, 10000)
	hash := uploadFile(t, f, data, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hex.EncodeToString(hash), 5, true)

}

// Pin a encrypted RAW file and unpin it multiple times
func TestPinRawUploadEncrypted(t *testing.T) {

	p, f := getPinApiAndFileStore(t)
	data := testutil.RandomBytes(1, 10000)
	hash := uploadFile(t, f, data, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hex.EncodeToString(hash), 5, true)

}

// Pin a simple collection and unpin it multiple times
func TestPinCollectionUpload(t *testing.T) {

	p, f := getPinApiAndFileStore(t)
	hash := uploadCollection(t, p, f, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hex.EncodeToString(hash), 1, false)

}

// Pin a encrypted simple collection and unpin it multiple times
func TestPinCollectionUploadEncrypted(t *testing.T) {

	p, f := getPinApiAndFileStore(t)
	hash := uploadCollection(t, p, f, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hex.EncodeToString(hash), 5, false)

}

// Test if the uploaded collection shows up in the ListPinFiles command
func TestListPinInfo(t *testing.T) {

	p, f := getPinApiAndFileStore(t)
	hash := uploadCollection(t, p, f, false)

	// Pin the hash
	err := p.PinFiles(hex.EncodeToString(hash), false, "")
	if err != nil {
		t.Fatalf("Could not pin " + err.Error())
	}

	// Gte the list of file pinned
	pinInfo := p.ListPinFiles()

	// Check if the uploaded collection is in the list files
	if _, ok := pinInfo[hex.EncodeToString(hash)]; !ok {
		t.Fatalf("uploaded collection not pinned")
	}

}

func getPinApiAndFileStore(t *testing.T) (*PinAPI, *storage.FileStore) {

	t.Helper()

	swarmDir, err := ioutil.TempDir("", "swarm-storage-test")
	if err != nil {
		t.Fatalf("could not create temp dir. Error: %s", err.Error())
	}

	localStore, err := localstore.New(swarmDir, make([]byte, 32), nil)
	if err != nil {
		os.RemoveAll(swarmDir)
		t.Fatalf("could not create localstore. Error: %s", err.Error())
	}

	tags := chunk.NewTags()
	fileStore := storage.NewFileStore(localStore, storage.NewFileStoreParams(), tags)

	// Swarm feeds test setup
	feedsDir, err := ioutil.TempDir("", "swarm-feeds-test")
	if err != nil {
		t.Fatal(err)
	}

	feeds, err := feed.NewTestHandler(feedsDir, &feed.HandlerParams{})
	if err != nil {
		t.Fatal(err)
	}

	swarmApi := api.NewAPI(fileStore, nil, feeds.Handler, nil, tags)

	pinAPI := NewPinApi(localStore, nil, tags, swarmApi)

	return pinAPI, fileStore

}

func uploadFile(t *testing.T, f *storage.FileStore, data []byte, toEncrypt bool) storage.Address {

	t.Helper()

	size := int64(len(data))
	ctx := context.TODO()
	addr, wait, err := f.Store(ctx, bytes.NewReader(data), size, toEncrypt)
	if err != nil {
		t.Fatalf("could not store file. Error: %s", err.Error())
	}

	err = wait(ctx)
	if err != nil {
		t.Fatalf("Store wait error: %v", err.Error())
	}

	return addr
}

type fileInfo struct {
	fileName string
	fileHash storage.Address
}

func uploadCollection(t *testing.T, p *PinAPI, f *storage.FileStore, toEncrypt bool) storage.Address {

	file1hash := uploadFile(t, f, []byte("file1.txt"), toEncrypt)
	file2hash := uploadFile(t, f, []byte("file2.txt"), toEncrypt)
	file3hash := uploadFile(t, f, []byte("dir1/file3.txt"), toEncrypt)
	file4hash := uploadFile(t, f, []byte("dir1/file4.txt"), toEncrypt)
	file5hash := uploadFile(t, f, []byte("dir2/file5.txt"), toEncrypt)
	file6hash := uploadFile(t, f, []byte("dir2/dir3/file6.txt"), toEncrypt)
	file7hash := uploadFile(t, f, []byte("dir2/dir4/file7.txt"), toEncrypt)
	file8hash := uploadFile(t, f, []byte("dir2/dir4/file8.txt"), toEncrypt)

	var testDirFiles = []fileInfo{
		{"file1.txt", file1hash},
		{"file2.txt", file2hash},
		{"dir1/file3.txt", file3hash},
		{"dir1/file4.txt", file4hash},
		{"dir2/file5.txt", file5hash},
		{"dir2/dir3/file6.txt", file6hash},
		{"dir2/dir4/file7.txt", file7hash},
		{"dir2/dir4/file8.txt", file8hash},
	}

	newAddr, err := p.api.NewManifest(context.TODO(), toEncrypt)
	if err != nil {
		t.Fatalf("could not create new manifest error: %v", err.Error())
	}

	newAddr, err = p.api.UpdateManifest(context.TODO(), newAddr, func(mw *api.ManifestWriter) error {

		for _, fileInfo := range testDirFiles {
			entry := &api.ManifestEntry{
				Hash:        hex.EncodeToString(fileInfo.fileHash),
				Path:        fileInfo.fileName,
				ContentType: mime.TypeByExtension(filepath.Ext(fileInfo.fileName)),
			}
			_, err = mw.AddEntry(context.TODO(), nil, entry)
			if err != nil {
				t.Fatalf("could not create new manifest error: %v", err.Error())
			}
		}
		return nil
	})

	return newAddr
}

// This function is called from test after a file is pinned.
// It check if the file's chunks are properly pinned.
// Assumption is that the file is uploaded in an empty database so that it can be easily tested.
// It takes the root hash and expected pinCounter value.
// It also has some hacks to take care of existing issues in the way we upload.
//
// The check process is as follows
//   1) Check if the root hash is pinned in pinFilesIndex
//         - getPinnedFiles gives all the root chunks pinned in pinFilesIndex
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
//         -  This is just a simple go through of all the pinned chunks list and check if the counter is
//            equal to the pin counter given as argument
func failIfNotPinned(t *testing.T, p *PinAPI, rootHash string, pinCounter uint64, isRaw bool) {

	t.Helper()

	// 1 - Check if the root hash is pinned in pinFilesIndex
	pinnedFiles := p.getPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; !ok {
		t.Fatalf("File %s not pinned in pinFilesIndex", rootHash)
	}

	// 2a - Get all the chunks of the file from retrievalDataIndex (since the file is uploaded in an empty database)
	chunksInDB := p.getAllChunksFromDB()

	// 2b - Get pinned chunks details from pinning indexes
	pinnedChunks := p.collectPinnedChunks(rootHash, "")
	if !isRaw {
		// Add the empty manifest chunk
		pinnedChunks["8b634aea26eec353ac0ecbec20c94f44d6f8d11f38d4578a4c207a84c74ef731"] = pinCounter
	}

	// 2c - Check if number of chunk hashes are same
	if len(chunksInDB) != len(pinnedChunks) {
		t.Fatalf("Expected number of chunks to be %d, but is %d", len(chunksInDB), len(pinnedChunks))
	}

	// 2c - Check if all the chunk address are same
	// tolerate 1 chunk failure for dummy manifest on encrypted collection
	tolerateOneChunkFailure := true
	for hash := range chunksInDB {
		if _, ok := pinnedChunks[hash]; !ok {
			if !isRaw && tolerateOneChunkFailure {
				tolerateOneChunkFailure = false
				continue
			}
			t.Fatalf("Expected chunk %s not present", hash)
		}
	}

	// 3 - Check for pin counter correctness
	if pinCounter != 0 {
		for _, pc := range pinnedChunks {
			if pc != pinCounter {
				t.Fatalf("Expected pin counter %d got %d", pinCounter, pc)
			}
		}
	}
}

func failIfUnpinned(t *testing.T, p *PinAPI, rootHash string) {

	t.Helper()

	// root hash should not be in pinFilesIndex
	pinnedFiles := p.getPinnedFiles()
	if _, ok := pinnedFiles[rootHash]; ok {
		t.Fatalf("File %s pinned in pinFilesIndex", rootHash)
	}

	// The chunks of this file should not be in pinIndex too
	pinnedChunks := p.collectPinnedChunks(rootHash, "")
	if len(pinnedChunks) != 0 {
		t.Fatalf("Chunks of this file present in pinIndex")
	}
}

func pinUnpinAndFailIfError(t *testing.T, p *PinAPI, rootHash string, noOfPinUnpin int, isRaw bool) {

	t.Helper()

	// Pin the file and check if it is pinned
	for i := 0; i < noOfPinUnpin; i++ {
		err := p.PinFiles(rootHash, isRaw, "")
		if err != nil {
			t.Fatalf("Could not pin " + err.Error())
		}
		pinCounter := uint64(i + 1)
		failIfNotPinned(t, p, rootHash, pinCounter, isRaw)
	}

	// Unpin and see if the file is unpinned
	for i := noOfPinUnpin; i > 0; i-- {
		err := p.UnpinFiles(rootHash, "")
		if err != nil {
			t.Fatalf("Could not unpin " + err.Error())
		}

		if i == 1 {
			// Final unpinning
			failIfUnpinned(t, p, rootHash)
		} else {
			// Check if the pin counter is decremented
			pinCounter := uint64(i - 1)
			failIfNotPinned(t, p, rootHash, pinCounter, isRaw)
		}
	}
}
