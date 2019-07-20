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

package client

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethersphere/swarm/api"
	swarmhttp "github.com/ethersphere/swarm/api/http"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/feed"
	"github.com/ethersphere/swarm/storage/localstore"
	"github.com/ethersphere/swarm/storage/pin"
	"github.com/ethersphere/swarm/testutil"
)

//
//   isRaw       toPin        encrypted     noPfPins
//
//  Uploaded files are RAW
//    true        true           false         1          TestPinWithRawUpload
//    true        false          false         1          TestPinAfterRawUpload
//    true        false          false         2          TestPinAfterRawUploadPinMultipleTimes
//    true        true           true          1          TestPinUploadRawEncrypted
//    true        false          true          1          TestPinAfterUploadRawEncrypted
//
//  Collection formed by creating multiple files and modifying manifest entries
//    false       false          false         2          TestPinCollectionMultipleTimesAfterUpload
//    false       false          true          2          TestPinEncryptedCollectionMultipleTimesAfterUpload
//    false       true           false         2          TestPinCollectionDuringUploadMultipleTimes
//    false       true           true          2          TestPinEncryptedCollectionDuringUploadMultipleTimes
//
//  Collection formed by creating dir tree in FS and then uploading the directory
//    false        true           false        1          TestPinDuringUploadDirectory
//    false        true           true         1          TestPinDuringUploadEncryptedDirectory
//    false        false          false        1          TestPinAfterUploadDirectory
//    false        false          true         1          TestPinAfterUploadEncryptedDirectory
//
//  Multipart upload pinning
//    false        true           ---          1          TestPinningWithMultipartUpload

// Pin a file while uploading of a RAW file and unpin it
func TestPinWithRawUpload(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	data := testutil.RandomBytes(1, 10000)
	hash := testClientUploadRaw(srv, false, t, data, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, hash, true, 1)

}

// Pin a file separately after uploading a RAW file and unpin it
func TestPinAfterRawUpload(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	data := testutil.RandomBytes(1, 10000)
	hash := testClientUploadRaw(srv, false, t, data, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, hash, true, 1)
}

// Upload a RAW file then pin and unpin multiple times
func TestPinAfterRawUploadPinMultipleTimes(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	data := testutil.RandomBytes(1, 10000)
	hash := testClientUploadRaw(srv, false, t, data, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, hash, true, 2)
}

// Pin a file during upload of a RAW encrypted file and then unpin it
func TestPinUploadRawEncrypted(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	data := testutil.RandomBytes(1, 10000)
	hash := testClientUploadRaw(srv, true, t, data, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, hash, true, 1)
}

// Pin a file during upload of a RAW encrypted file and then unpin it
func TestPinAfterUploadRawEncrypted(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	data := testutil.RandomBytes(1, 10000)
	hash := testClientUploadRaw(srv, true, t, data, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, hash, true, 1)
}

// Pin collection after the file is uploaded and pin multiple times
func TestPinCollectionMultipleTimesAfterUpload(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	pinRoot := testClientUploadCollection(srv, false, t, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, pinRoot, false, 2)
}

// Pin encrypted collection after the file is uploaded and pin multiple times
func TestPinEncryptedCollectionMultipleTimesAfterUpload(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	pinRoot := testClientUploadCollection(srv, true, t, false)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, pinRoot, false, 2)
}

// Pin collection during file upload and once after that file is uploaded
func TestPinCollectionDuringUploadMultipleTimes(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	pinRoot := testClientUploadCollection(srv, false, t, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, pinRoot, false, 2)
}

// Pin encrypted collection during file upload and once after that file is uploaded
func TestPinEncryptedCollectionDuringUploadMultipleTimes(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	pinRoot := testClientUploadCollection(srv, true, t, true)

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, pinRoot, false, 2)
}

// Pin directory after the file is uploaded and pin multiple times
func TestPinDuringUploadDirectory(t *testing.T) {

	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	dir := newTestDirectory(t)
	defer os.RemoveAll(dir)

	// upload the directory
	client := NewClient(srv.URL)

	// dont set default path.. otherwise that file will be pinned twice
	hash, err := client.UploadDirectory(dir, "", "", false, true)
	if err != nil {
		t.Fatalf("error uploading directory: %s", err)
	}

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, hash, false, 1)
}

// Pin directory collection after the file is uploaded and pin multiple times
func TestPinDuringUploadEncryptedDirectory(t *testing.T) {

	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	dir := newTestDirectory(t)
	defer os.RemoveAll(dir)

	// upload the directory
	client := NewClient(srv.URL)

	// dont set default path.. otherwise that file will be pinned twice
	hash, err := client.UploadDirectory(dir, "", "", true, true)
	if err != nil {
		t.Fatalf("error uploading directory: %s", err)
	}

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, true, hash, false, 1)
}

// Pin directory during file upload and once after that file is uploaded
func TestPinAfterUploadDirectory(t *testing.T) {

	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	dir := newTestDirectory(t)
	defer os.RemoveAll(dir)

	// upload the directory
	client := NewClient(srv.URL)

	// dont set default path.. otherwise that file will be pinned twice
	hash, err := client.UploadDirectory(dir, "", "", false, false)
	if err != nil {
		t.Fatalf("error uploading directory: %s", err)
	}

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, hash, false, 1)
}

// Pin directory collection during file upload and once after that file is uploaded
func TestPinAfterUploadEncryptedDirectory(t *testing.T) {

	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	dir := newTestDirectory(t)
	defer os.RemoveAll(dir)

	// upload the directory
	client := NewClient(srv.URL)

	// dont set default path.. otherwise that file will be pinned twice
	hash, err := client.UploadDirectory(dir, "", "", true, false)
	if err != nil {
		t.Fatalf("error uploading directory: %s", err)
	}

	// test pin and unpin
	pinUnpinAndFailIfError(t, srv, false, hash, false, 1)
}

// Pin during a multipart upload
// upload with pinning and unpinning
func TestPinningWithMultipartUpload(t *testing.T) {
	srv := NewTestSwarmServer(t, pinServerFunc, nil, nil)
	defer srv.Close()

	// define an uploader which uploads testDirFiles with some data
	// note: this test should result in SEEN chunks. assert accordingly
	uploader := UploaderFunc(func(upload UploadFn) error {
		for _, name := range testDirFiles {
			data := []byte(name)
			file := &File{
				ReadCloser: ioutil.NopCloser(bytes.NewReader(data)),
				ManifestEntry: api.ManifestEntry{
					Path:        name,
					ContentType: "text/plain",
					Size:        int64(len(data)),
				},
			}
			if err := upload(file); err != nil {
				return err
			}
		}
		return nil
	})

	// upload the files as a multipart upload
	client := NewClient(srv.URL)
	hash, err := client.MultipartUpload("", uploader, true)
	if err != nil {
		t.Fatal(err)
	}

	// File pinned during upload.. check if this is pinned properly
	failIfNotPinned(t, srv, hash, 1, false)

}

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
//         -  This is just a simple go through of all the pinned chunks list and check if the counter is
//            equal to the pin counter given as argument
func failIfNotPinned(t *testing.T, srv *testSimpleSwarmServer, rootHash string, pinCounter uint64, isRaw bool) {

	t.Helper()

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

type testSimpleSwarmServer struct {
	*httptest.Server
	Hasher      storage.SwarmHash
	FileStore   *storage.FileStore
	Tags        *chunk.Tags
	PinAPI      *pin.PinAPI
	dir         string
	cleanup     func()
	CurrentTime uint64
}

func NewTestSwarmServer(t *testing.T, serverFunc func(*api.API, *pin.PinAPI) swarmhttp.TestServer, resolver api.Resolver,
	o *localstore.Options) *testSimpleSwarmServer {

	swarmDir, err := ioutil.TempDir("", "swarm-storage-test")
	if err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(swarmDir, make([]byte, 32), o)
	if err != nil {
		os.RemoveAll(swarmDir)
		t.Fatal(err)
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

	swarmApi := api.NewAPI(fileStore, resolver, feeds.Handler, nil, tags)
	pinAPI := pin.NewPinApi(localStore, nil, tags, swarmApi)
	apiServer := httptest.NewServer(serverFunc(swarmApi, pinAPI))

	tss := &testSimpleSwarmServer{
		Server:    apiServer,
		FileStore: fileStore,
		Tags:      tags,
		PinAPI:    pinAPI,
		dir:       swarmDir,
		Hasher:    storage.MakeHashFunc(storage.DefaultHash)(),
		cleanup: func() {
			apiServer.Close()
			fileStore.Close()
			feeds.Close()
			os.RemoveAll(swarmDir)
			os.RemoveAll(feedsDir)
		},
		CurrentTime: 42,
	}
	feed.TimestampProvider = tss
	return tss
}

func (t *testSimpleSwarmServer) Close() {
	t.cleanup()
}
func (t *testSimpleSwarmServer) Now() feed.Timestamp {
	return feed.Timestamp{Time: t.CurrentTime}
}

func failIfUnpinned(t *testing.T, srv *testSimpleSwarmServer, rootHash string) {

	t.Helper()

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

func pinUnpinAndFailIfError(t *testing.T, srv *testSimpleSwarmServer, pinnedDuringUpload bool,
	rootHash string, isRaw bool, noOfPinUnpin int) {

	t.Helper()

	for i := 0; i < noOfPinUnpin; i++ {
		if !pinnedDuringUpload {

			// Pin the file
			err := srv.PinAPI.PinFiles(rootHash, isRaw, "")
			if err != nil {
				t.Fatalf("Could not pin " + err.Error())
			}

			pinCounter := uint64(i + 1)
			failIfNotPinned(t, srv, rootHash, pinCounter, isRaw)
		} else {
			// If the file is already pinned during upload.. just check pinning
			pinnedDuringUpload = false
			failIfNotPinned(t, srv, rootHash, 1, isRaw)
		}
	}

	for i := noOfPinUnpin; i > 0; i-- {

		// Unpin and see if the file is unpinned
		err := srv.PinAPI.UnpinFiles(rootHash, "")
		if err != nil {
			t.Fatalf("Could not unpin " + err.Error())
		}

		if i == 1 {
			// Final unpinning
			failIfUnpinned(t, srv, rootHash)
		} else {
			// Check if the pin counter is decremented
			pinCounter := uint64(noOfPinUnpin - 1)
			failIfNotPinned(t, srv, rootHash, pinCounter, isRaw)
		}
	}
}

func testClientUploadRaw(srv *testSimpleSwarmServer, toEncrypt bool, t *testing.T, data []byte, toPin bool) string {
	clientGateway := NewClient(srv.URL)

	hash, err := clientGateway.UploadRaw(bytes.NewReader(data), int64(len(data)), toEncrypt, toPin)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func testClientUploadCollection(srv *testSimpleSwarmServer, toEncrypt bool, t *testing.T, pinDuringUpload bool) string {

	t.Helper()

	clientGateway := NewClient(srv.URL)
	upload := func(manifest, path string, data []byte, toPin bool) string {
		file := &File{
			ReadCloser: ioutil.NopCloser(bytes.NewReader(data)),
			ManifestEntry: api.ManifestEntry{
				Path:        path,
				ContentType: "text/plain",
				Size:        int64(len(data)),
			},
		}
		hash, err := clientGateway.Upload(file, manifest, toEncrypt, toPin)
		if err != nil {
			t.Fatal(err)
		}
		return hash
	}

	// upload a file to the root of a manifest
	rootData := testutil.RandomBytes(1, 10000)
	rootHash := upload("", "", rootData, pinDuringUpload)

	return rootHash
}

func pinServerFunc(api *api.API, pinAPI *pin.PinAPI) swarmhttp.TestServer {
	return swarmhttp.NewServer(api, "", pinAPI)
}
