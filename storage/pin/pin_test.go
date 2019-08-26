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
	"errors"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/testutil"
)

// TestPinRawUpload pins a RAW file and unpin it multiple times
func TestPinRawUpload(t *testing.T) {
	p, f, closeFunc := getPinApiAndFileStore(t)
	defer closeFunc()

	data := testutil.RandomBytes(1, 10000)
	hash := uploadFile(t, f, data, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hash, 3, true)
}

// TestPinRawUploadEncrypted pins a encrypted RAW file and unpin it multiple times
func TestPinRawUploadEncrypted(t *testing.T) {
	p, f, closeFunc := getPinApiAndFileStore(t)
	defer closeFunc()

	data := testutil.RandomBytes(2, 10000)
	hash := uploadFile(t, f, data, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hash, 3, true)
}

// TestPinCollectionUpload pins a simple collection and unpin it multiple times
func TestPinCollectionUpload(t *testing.T) {
	p, f, closeFunc := getPinApiAndFileStore(t)
	defer closeFunc()

	hash := uploadCollection(t, p, f, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hash, 3, false)
}

// TestPinCollectionUploadEncrypted pins a encrypted simple collection and unpin it multiple times
func TestPinCollectionUploadEncrypted(t *testing.T) {
	p, f, closeFunc := getPinApiAndFileStore(t)
	defer closeFunc()

	hash := uploadCollection(t, p, f, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, p, hash, 3, false)
}

// TestWalker tests the walkChunksFromRootHash function which is the crux of
// commands like pin, unpin & list.
func TestWalker(t *testing.T) {
	sizes := []int{1, 4095, 4096, 4097, 123456}
	for i := range sizes {
		p, f, closeFunc := getPinApiAndFileStore(t)
		defer closeFunc()

		data := testutil.RandomBytes(1, sizes[i])
		hash := uploadFile(t, f, data, false)

		addrs, err := f.GetAllReferences(context.TODO(), bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Error getting original chunks of the file")
		}

		// Function to collect the walked chunks
		walkedChunks := make(map[string]uint64)
		var lock = sync.RWMutex{}
		walkerFunction := func(ref storage.Reference) error {
			chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
			lock.Lock()
			defer lock.Unlock()
			walkedChunks[hex.EncodeToString(chunkAddr)] = 0
			return nil
		}
		err = p.walkChunksFromRootHash(hash, true, "", walkerFunction)
		if err != nil {
			t.Fatalf("Walker error for hash %s", hash)
		}

		// Check if the number of chunks and chunk addresses match
		if len(addrs) != len(walkedChunks) {
			t.Fatalf("Expected number of chunks to be %d, but is %d", len(addrs), len(walkedChunks))
		}
		for _, hash := range addrs {
			if _, ok := walkedChunks[hex.EncodeToString(hash)]; !ok {
				t.Fatalf("Expected chunk %s not present", hash)
			}
		}
	}
}

// TestListPinInfo tests the ListPins command by pinning and unpinning a collection
// twice and check if this gets reflected properly in the data structure
func TestListPinInfo(t *testing.T) {
	p, f, closeFunc := getPinApiAndFileStore(t)
	defer closeFunc()

	hash := uploadCollection(t, p, f, false)

	// Pin the hash for the first time
	err := p.PinFiles(hash, false, "")
	if err != nil {
		t.Fatalf("Could not pin " + err.Error())
	}

	// Get the list of pinned files by calling the ListPins command
	pinsInfo, err := p.ListPins()
	if err != nil {
		t.Fatalf("Error executing ListPins command")
	}

	// Check if the uploaded collection is in the list files data structure
	pinInfo, err := getPinInfo(pinsInfo, hash)
	if err != nil {
		t.Fatalf("uploaded collection not pinned")
	}
	if pinInfo.PinCounter != 1 {
		t.Fatalf("pincounter expected is 1 got is %d", pinInfo.PinCounter)
	}
	if pinInfo.IsRaw {
		t.Fatalf("IsRaw expected is false got is true")
	}

	// Pin it once more and check if the counters increases
	err = p.PinFiles(hash, false, "")
	if err != nil {
		t.Fatalf("Could not pin " + err.Error())
	}

	// Get the list of pinned files by calling the ListPins command
	pinsInfo, err = p.ListPins()
	if err != nil {
		t.Fatalf("Error executing ListPins command")
	}
	pinInfo, err = getPinInfo(pinsInfo, hash)
	if err != nil {
		t.Fatalf("hash not pinned ")
	}
	if pinInfo.PinCounter != 2 {
		t.Fatalf("pincounter expected is 2 got is %d", pinInfo.PinCounter)
	}

	// Unpin it and check if the counter decrements
	err = p.UnpinFiles(hash, "")
	if err != nil {
		t.Fatalf("Could not unpin " + err.Error())
	}

	// Get the list of pinned files by calling the ListPins command
	pinsInfo, err = p.ListPins()
	if err != nil {
		t.Fatalf("Error executing ListPins command")
	}
	pinInfo, err = getPinInfo(pinsInfo, hash)
	if err != nil {
		t.Fatalf("collection totally unpinned")
	}
	if pinInfo.PinCounter != 1 {
		t.Fatalf("pincounter expected is 1 got is %d", pinInfo.PinCounter)
	}

	// Unpin it final time and the entry should not be there
	err = p.UnpinFiles(hash, "")
	if err != nil {
		t.Fatalf("Could not unpin " + err.Error())
	}

	// Get the list of pinned files by calling the ListPins command
	pinsInfo, err = p.ListPins()
	if err != nil {
		t.Fatalf("Error executing ListPins command")
	}
	_, err = getPinInfo(pinsInfo, hash)
	if err == nil {
		t.Fatalf("uploaded collection is still pinned")
	}
}

func getPinApiAndFileStore(t *testing.T) (*API, *storage.FileStore, func()) {
	t.Helper()

	swarmDir, err := ioutil.TempDir("", "swarm-storage-test")
	if err != nil {
		t.Fatalf("could not create temp dir. Error: %s", err.Error())
	}

	stateStore, err := state.NewDBStore(filepath.Join(swarmDir, "state-store.db"))
	if err != nil {
		t.Fatalf("could not create state store. Error: %s", err.Error())
	}

	lStore, err := localstore.New(swarmDir, make([]byte, 32), nil)
	if err != nil {
		t.Fatalf("could not create localstore. Error: %s", err.Error())
	}
	tags := chunk.NewTags()
	fileStore := storage.NewFileStore(lStore, storage.NewFileStoreParams(), tags)

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
	pinAPI := NewAPI(lStore, stateStore, nil, tags, swarmApi)

	closeFunc := func() {
		err := stateStore.Close()
		if err != nil {
			t.Fatalf("Could not close state store")
		}
		err = lStore.Close()
		if err != nil {
			t.Fatalf("Could not close localStore")
		}
		feeds.Close()
		err = os.RemoveAll(feedsDir)
		if err != nil {
			t.Fatalf("Could not remove swarm feeds dir")
		}
		err = os.RemoveAll(swarmDir)
		if err != nil {
			t.Fatalf("Could not remove swarm temp dir")
		}
	}

	return pinAPI, fileStore, closeFunc
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

type testFileInfo struct {
	fileName string
	fileHash storage.Address
}

func uploadCollection(t *testing.T, p *API, f *storage.FileStore, toEncrypt bool) storage.Address {
	file1hash := uploadFile(t, f, testutil.RandomBytes(1, 10000), toEncrypt)
	file2hash := uploadFile(t, f, testutil.RandomBytes(2, 10000), toEncrypt)
	file3hash := uploadFile(t, f, testutil.RandomBytes(3, 10000), toEncrypt)
	file4hash := uploadFile(t, f, testutil.RandomBytes(4, 10000), toEncrypt)
	file5hash := uploadFile(t, f, testutil.RandomBytes(5, 10000), toEncrypt)
	file6hash := uploadFile(t, f, testutil.RandomBytes(6, 10000), toEncrypt)
	file7hash := uploadFile(t, f, testutil.RandomBytes(7, 10000), toEncrypt)
	file8hash := uploadFile(t, f, testutil.RandomBytes(8, 10000), toEncrypt)

	var testDirFiles = []testFileInfo{
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

	// Simulate "swarm --recursive up"
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
// It takes the root hash and expected PinCounter value.
// It also has some hacks to take care of existing issues in the way we upload.
//
// The check process is as follows
//   1) Check if the root hash is present in the pinnedFile map
//   2) Check if all the files's chunks are in pinIndex
//         a) Get all the chunks
//            get it from retrievalDataIndex
//            since the assumption is the DB has only this file, it gives all the file's chunks.
//            getAllRefs cannot be used here as it does not give the chunks that belong to manifests.
//         b) Get all chunks that are pinned (from pinIndex)
//            In every upload.. an empty manifest is uploaded. that why add this hash to this list
//         c) Check if both the above lists are equal
//   3) Check if all the chunks pinned have the proper PinCounter
//         -  This is just a simple go through of all the pinned chunks list and check if the counter is
//            equal to the pin counter given as argument
func failIfNotPinned(t *testing.T, p *API, rootHash []byte, pinCounter uint64, isRaw bool) {
	t.Helper()

	// 1 - Check if the root hash is pinned in state store
	_, err := p.getPinnedFile(rootHash)
	if err != nil {
		t.Fatalf("File %s not pinned in state store", rootHash)
	}

	// 2a - Get all the chunks of the file from retrievalDataIndex (since the file is uploaded in an empty database)
	chunksInDB := p.getAllChunksFromDB(t)

	// 2b - Get pinned chunks details from pinning indexes
	pinnedChunks := p.collectPinnedChunks(t, rootHash, "", isRaw)
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

	pinnedFiles, err := p.ListPins()
	if err != nil {
		t.Fatalf("Could not load pin state from state store")
	}

	fileInfo, err := getPinInfo(pinnedFiles, rootHash)
	if err != nil {
		t.Fatalf("Fileinfo not present in state store")
	}

	if fileInfo.IsRaw != isRaw {
		t.Fatalf("Invalid IsRaw state in fileInfo")
	}

	if fileInfo.PinCounter != pinCounter {
		t.Fatalf("Invalid pincounter expected %d got %d", pinCounter, fileInfo.PinCounter)
	}
}

func failIfNotUnpinned(t *testing.T, p *API, rootHash []byte, isRaw bool) {
	t.Helper()

	// root hash should not be in state DB
	_, err := p.getPinnedFile(rootHash)
	if err == nil {
		t.Fatalf("File %s pinned in pinFilesIndex", rootHash)
	}

	// The chunks of this file should not be in pinIndex too
	pinnedChunks := p.collectPinnedChunks(t, rootHash, "", isRaw)
	if len(pinnedChunks) != 0 {
		t.Fatalf("Chunks of this file present in pinIndex")
	}
}

func pinUnpinAndFailIfError(t *testing.T, p *API, rootHash []byte, noOfPinUnpin int, isRaw bool) {
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
			failIfNotUnpinned(t, p, rootHash, isRaw)
		} else {
			// Check if the pin counter is decremented
			pinCounter := uint64(i - 1)
			failIfNotPinned(t, p, rootHash, pinCounter, isRaw)
		}
	}
}

// collectPinnedChunks is used to collect all the chunks that are pinned as part of the
// given root hash.
func (p *API) collectPinnedChunks(t *testing.T, rootHash []byte, credentials string, isRaw bool) map[string]uint64 {
	t.Helper()

	pinnedChunks := make(map[string]uint64)
	var lock = sync.RWMutex{}
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		pinCounter, err := p.getPinCounterOfChunk(chunk.Address(chunkAddr))
		if err != nil {
			if err == chunk.ErrChunkNotFound {
				return nil
			} else {
				return err
			}
		}
		lock.Lock()
		defer lock.Unlock()
		pinnedChunks[hex.EncodeToString(chunkAddr)] = pinCounter
		return nil
	}
	err := p.walkChunksFromRootHash(rootHash, isRaw, credentials, walkerFunction)
	if err != nil {
		t.Fatal("Error during walking")
	}

	return pinnedChunks
}

// getAllChunksFromDB is used in testing to generate the truth dataset about all the chunks
// that are present in the DB.
func (p *API) getAllChunksFromDB(t *testing.T) map[string]int {
	t.Helper()

	var addrLock sync.RWMutex
	addrs := make(map[string]int)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup

	for bin := uint8(0); bin < uint8(chunk.MaxPO); bin++ {
		ch, stop := p.db.SubscribePull(ctx, bin, 0, 0)
		defer stop()

		wg.Add(1)
		go getChunks(t, bin, addrs, &addrLock, ch, &wg, ctx)
	}

	wg.Wait()
	return addrs
}

func getChunks(t *testing.T, bin uint8, addrs map[string]int, addrLock *sync.RWMutex, ch <-chan chunk.Descriptor, wg *sync.WaitGroup, ctx context.Context) {
	t.Helper()

	defer wg.Done()
	for {
		select {
		case got, ok := <-ch:
			if !ok {
				return
			}
			addrLock.Lock()
			addrs[hex.EncodeToString(got.Address)] = 0
			addrLock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func getPinInfo(pinInfo []PinInfo, hash storage.Address) (PinInfo, error) {
	for _, fi := range pinInfo {
		if bytes.Equal(fi.Address, hash) {
			return fi, nil
		}
	}
	return PinInfo{}, errors.New("Pininfo not found")
}
