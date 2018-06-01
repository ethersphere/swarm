// Copyright 2017 The go-ethereum Authors
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

// +build linux darwin freebsd

package fuse

import (
	"bytes"
	"crypto/rand"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/storage"

	"github.com/ethereum/go-ethereum/log"

	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 4, "verbosity of logs")
	rawlog   = flag.Bool("rawlog", false, "turn off terminal formatting in logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(!*rawlog))))
}

type fileInfo struct {
	perm     uint64
	uid      int
	gid      int
	contents []byte
}

func createTestFilesAndUploadToSwarm(t *testing.T, api *api.Api, files map[string]fileInfo, uploadDir string, toEncrypt bool) string {

	for fname, finfo := range files {
		actualPath := filepath.Join(uploadDir, fname)
		filePath := filepath.Dir(actualPath)

		err := os.MkdirAll(filePath, 0777)
		if err != nil {
			t.Fatalf("Error creating directory '%v' : %v", filePath, err)
		}

		fd, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(finfo.perm))
		if err1 != nil {
			t.Fatalf("Error creating file %v: %v", actualPath, err1)
		}

		fd.Write(finfo.contents)
		fd.Chown(finfo.uid, finfo.gid)
		fd.Chmod(os.FileMode(finfo.perm))
		fd.Sync()
		fd.Close()
	}

	bzzhash, err := api.Upload(uploadDir, "", toEncrypt)
	if err != nil {
		t.Fatalf("Error uploading directory %v: %vm encryption: %v", uploadDir, err, toEncrypt)
	}

	return bzzhash
}

func mountDir(t *testing.T, api *api.Api, files map[string]fileInfo, bzzHash string, mountDir string) *SwarmFS {
	swarmfs := NewSwarmFS(api)
	_, err := swarmfs.Mount(bzzHash, mountDir)
	if isFUSEUnsupportedError(err) {
		t.Skip("FUSE not supported:", err)
	} else if err != nil {
		t.Fatalf("Error mounting hash %v: %v", bzzHash, err)
	}

	found := false
	mi := swarmfs.Listmounts()
	for _, minfo := range mi {
		minfo.lock.RLock()
		if minfo.MountPoint == mountDir {
			if minfo.StartManifest != bzzHash ||
				minfo.LatestManifest != bzzHash ||
				minfo.fuseConnection == nil {
				minfo.lock.RUnlock()
				t.Fatalf("Error mounting: exp(%s): act(%s)", bzzHash, minfo.StartManifest)
			}
			found = true
		}
		minfo.lock.RUnlock()
	}

	// Test listMounts
	if !found {
		t.Fatalf("Error getting mounts information for %v: %v", mountDir, err)
	}

	// Check if file and their attributes are as expected
	compareGeneratedFileWithFileInMount(t, files, mountDir)

	return swarmfs
}

func compareGeneratedFileWithFileInMount(t *testing.T, files map[string]fileInfo, mountDir string) {
	err := filepath.Walk(mountDir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		fname := path[len(mountDir)+1:]
		if _, ok := files[fname]; !ok {
			t.Fatalf(" file %v present in mount dir and is not expected", fname)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error walking dir %v", mountDir)
	}

	for fname, finfo := range files {
		destinationFile := filepath.Join(mountDir, fname)

		dfinfo, err := os.Stat(destinationFile)
		if err != nil {
			t.Fatalf("Destination file %v missing in mount: %v", fname, err)
		}

		if int64(len(finfo.contents)) != dfinfo.Size() {
			t.Fatalf("file %v Size mismatch  source (%v) vs destination(%v)", fname, int64(len(finfo.contents)), dfinfo.Size())
		}

		if dfinfo.Mode().Perm().String() != "-rwx------" {
			t.Fatalf("file %v Permission mismatch source (-rwx------) vs destination(%v)", fname, dfinfo.Mode().Perm())
		}

		fileContents, err := ioutil.ReadFile(filepath.Join(mountDir, fname))
		if err != nil {
			t.Fatalf("Could not readfile %v : %v", fname, err)
		}
		if !bytes.Equal(fileContents, finfo.contents) {
			t.Fatalf("File %v contents mismatch: %v , %v", fname, fileContents, finfo.contents)

		}
		// TODO: check uid and gid
	}
}

func checkFile(t *testing.T, testMountDir, fname string, contents []byte) {
	destinationFile := filepath.Join(testMountDir, fname)
	dfinfo, err1 := os.Stat(destinationFile)
	if err1 != nil {
		t.Fatalf("Could not stat file %v", destinationFile)
	}
	if dfinfo.Size() != int64(len(contents)) {
		t.Fatalf("Mismatch in size  actual(%v) vs expected(%v)", dfinfo.Size(), int64(len(contents)))
	}

	fd, err2 := os.OpenFile(destinationFile, os.O_RDONLY, os.FileMode(0665))
	if err2 != nil {
		t.Fatalf("Could not open file %v", destinationFile)
	}
	newcontent := make([]byte, len(contents))
	fd.Read(newcontent)
	fd.Close()

	if !bytes.Equal(contents, newcontent) {
		t.Fatalf("File content mismatch expected (%v): received (%v) ", contents, newcontent)
	}
}

func getRandomBytes(size int) []byte {
	contents := make([]byte, size)
	rand.Read(contents)
	return contents
}

func isDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)

	return err == io.EOF
}

type testAPI struct {
	api *api.API
}

func (ta *testAPI) mountListAndUnmountEncrypted(t *testing.T) {
	log.Info("Starting mountListAndUnmountEncrypted test")
	ta.mountListAndUnmount(t, true)
	log.Info("Test mountListAndUnmountEncrypted terminated")
}

func (ta *testAPI) mountListAndUnmountNonEncrypted(t *testing.T) {
	log.Info("Starting mountListAndUnmountNonEncrypted test")
	ta.mountListAndUnmount(t, false)
	log.Info("Test mountListAndUnmountNonEncrypted terminated")
}

func (ta *testAPI) mountListAndUnmount(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testDir, err := ioutil.TempDir(os.TempDir(), "mountListAndUnmount")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)
	testUploadDir := filepath.Join(testDir, "fuse-source")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}

	testMountDir := filepath.Join(testDir, "fuse-dest")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir: %v", err)
	}

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["2.txt"] = fileInfo{0711, 333, 444, getRandomBytes(10)}
	files["3.txt"] = fileInfo{0622, 333, 444, getRandomBytes(100)}
	files["4.txt"] = fileInfo{0533, 333, 444, getRandomBytes(1024)}
	files["5.txt"] = fileInfo{0544, 333, 444, getRandomBytes(10)}
	files["6.txt"] = fileInfo{0555, 333, 444, getRandomBytes(10)}
	files["7.txt"] = fileInfo{0666, 333, 444, getRandomBytes(10)}
	files["8.txt"] = fileInfo{0777, 333, 333, getRandomBytes(10)}
	files["11.txt"] = fileInfo{0777, 333, 444, getRandomBytes(10)}
	files["111.txt"] = fileInfo{0777, 333, 444, getRandomBytes(10)}
	files["two/2.txt"] = fileInfo{0777, 333, 444, getRandomBytes(10)}
	files["two/2/2.txt"] = fileInfo{0777, 333, 444, getRandomBytes(10)}
	files["two/2./2.txt"] = fileInfo{0777, 444, 444, getRandomBytes(10)}
	files["twice/2.txt"] = fileInfo{0777, 444, 333, getRandomBytes(200)}
	files["one/two/three/four/five/six/seven/eight/nine/10.txt"] = fileInfo{0777, 333, 444, getRandomBytes(10240)}
	files["one/two/three/four/five/six/six"] = fileInfo{0777, 333, 444, getRandomBytes(10)}

	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to Swarm")

	swarmfs := mountDir(t, ta.api, files, bzzHash, testMountDir)
	defer swarmfs.Stop()
	log.Info("Mounted swarm fs")

	// Check unmount
	_, err = swarmfs.Unmount(testMountDir)
	if err != nil {
		t.Fatalf("could not unmount  %v", bzzHash)
	}
	log.Info("Unmount successful")
	if !isDirEmpty(testMountDir) {
		t.Fatalf("unmount didnt work for %v", testMountDir)
	}
	log.Info("mountListAndUnmount terminated")
}

func (ta *testAPI) maxMountsEncrypted(t *testing.T) {
	log.Info("Starting maxMountsEncrypted test")
	ta.runMaxMounts(t, true)
	log.Info("Test maxMountsEncrypted terminated")
}

func (ta *testAPI) maxMountsNonEncrypted(t *testing.T) {
	log.Info("Starting maxMountsNonEncrypted test")
	ta.runMaxMounts(t, false)
	log.Info("Test maxMountsNonEncrypted terminated")
}

func (ta *testAPI) runMaxMounts(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testDir, err := ioutil.TempDir(os.TempDir(), "runMaxMounts")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "max-upload1")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	bzzHash1 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir1")
	testMountDir := filepath.Join(testDir, "max-mount1")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir: %v", err)
	}
	swarmfs1 := mountDir(t, ta.api, files, bzzHash1, testMountDir)
	defer swarmfs1.Stop()

	testUploadDir2 := filepath.Join(testDir, "max-upload2")
	err = os.MkdirAll(testUploadDir2, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir 2: %v", err)
	}
	files["2.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testMountDir2 := filepath.Join(testDir, "max-mount2")
	err = os.MkdirAll(testMountDir2, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	bzzHash2 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir2, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir2")
	_ = mountDir(t, ta.api, files, bzzHash2, testMountDir2)

	testUploadDir3 := filepath.Join(testDir, "max-upload3")
	err = os.MkdirAll(testUploadDir3, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir 3: %v", err)
	}
	files["3.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testMountDir3 := filepath.Join(testDir, "max-mount3")
	err = os.MkdirAll(testMountDir3, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 3: %v", err)
	}
	bzzHash3 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir3, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir3")
	_ = mountDir(t, ta.api, files, bzzHash3, testMountDir3)

	testUploadDir4 := filepath.Join(testDir, "max-upload4")
	err = os.MkdirAll(testUploadDir4, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir 4: %v", err)
	}
	files["4.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testMountDir4 := filepath.Join(testDir, "max-mount4")
	err = os.MkdirAll(testMountDir4, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 4: %v", err)
	}
	bzzHash4 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir4, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir4")
	_ = mountDir(t, ta.api, files, bzzHash4, testMountDir4)

	testUploadDir5 := filepath.Join(testDir, "max-upload5")
	err = os.MkdirAll(testUploadDir5, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir 5: %v", err)
	}
	files["5.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testMountDir5 := filepath.Join(testDir, "max-mount5")
	err = os.MkdirAll(testMountDir5, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 5: %v", err)
	}
	bzzHash5 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir5, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir5")
	_ = mountDir(t, ta.api, files, bzzHash5, testMountDir5)

	testUploadDir6 := filepath.Join(testDir, "max-upload6")
	err = os.MkdirAll(testUploadDir6, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir 6: %v", err)
	}
	files["6.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	testMountDir6 := filepath.Join(testDir, "max-mount6")
	err = os.MkdirAll(testMountDir6, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 5: %v", err)
	}
	bzzHash6 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir6, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir6")
	_, err = swarmfs.Mount(bzzHash6, testMountDir6)
	if err == nil {
		t.Fatalf("Error: Going beyond max mounts  %v", bzzHash6)
	}
	log.Info("Maximum mount reached, additional mount failed. Correct.")
}

func (ta *testAPI) remountEncrypted(t *testing.T) {
	log.Info("Starting remountEncrypted test")
	ta.remount(t, true)
	log.Info("Test remountEncrypted terminated")
}
func (ta *testAPI) remountNonEncrypted(t *testing.T) {
	log.Info("Starting remountNonEncrypted test")
	ta.remount(t, false)
	log.Info("Test remountNonEncrypted terminated")
}

func (ta *testAPI) remount(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}

	testDir, err := ioutil.TempDir(os.TempDir(), "remount")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "remount-upload1")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir := filepath.Join(testDir, "remount-mount1")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	bzzHash1 := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir1")
	swarmfs := mountDir(t, ta.api, files, bzzHash1, testMountDir)
	defer swarmfs.Stop()

	uploadDir2 := filepath.Join(testDir, "remount-upload2")
	err = os.MkdirAll(uploadDir2, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir2 := filepath.Join(testDir, "remount-mount2")
	err = os.MkdirAll(testMountDir2, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	bzzHash2 := createTestFilesAndUploadToSwarm(t, ta.api, files, uploadDir2, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir2")
	_, err = swarmfs.Mount(bzzHash1, testMountDir2)
	if err != nil {
		t.Fatalf("Error mounting hash  %v", bzzHash1)
	}
	swarmfs.Unmount(testMountDir2)
	log.Info("Remount hash1 successful")

	// mount a different hash in already mounted point
	_, err = swarmfs.Mount(bzzHash2, testMountDir)
	if err == nil {
		t.Fatalf("Error mounting hash  %v", bzzHash2)
	}
	log.Info("Mount on existing mount point failed. Correct.")

	// mount nonexistent hash
	_, err = swarmfs.Mount("0xfea11223344", testMountDir)
	if err == nil {
		t.Fatalf("Error mounting hash  %v", bzzHash2)
	}
	log.Info("Nonexistent hash hasn't been mounted. Correct.")
}

func (ta *testAPI) unmountEncrypted(t *testing.T) {
	log.Info("Starting unmountEncrypted test")
	ta.unmount(t, true)
	log.Info("Test unmountEncrypted terminated")
}

func (ta *testAPI) unmountNonEncrypted(t *testing.T) {
	log.Info("Starting unmountNonEncrypted test")
	ta.unmount(t, false)
	log.Info("Test unmountNonEncrypted terminated")
}

func (ta *testAPI) unmount(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testDir, err := ioutil.TempDir(os.TempDir(), "unmount")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "ex-upload")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir := filepath.Join(testDir, "ex-mount")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")
	swarmfs := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Mounted Dir")
	defer swarmfs.Stop()

	swarmfs.Unmount(testMountDir)
	log.Info("Unmounted Dir")

	mi := swarmfs.Listmounts()
	log.Debug("Going to list mounts")
	for _, minfo := range mi {
		log.Debug("Mount point in list: ", "point", minfo.MountPoint)
		if minfo.MountPoint == testMountDir {
			t.Fatalf("mount state not cleaned up in unmount case %v", testMountDir)
		}
	}
	log.Debug("subtest terminated")
}

func (ta *testAPI) unmountWhenResourceBusyEncrypted(t *testing.T) {
	log.Info("Starting unmountWhenResourceBusyEncrypted test")
	ta.unmountWhenResourceBusy(t, true)
	log.Info("Test unmountWhenResourceBusyEncrypted terminated")
}
func (ta *testAPI) unmountWhenResourceBusyNonEncrypted(t *testing.T) {
	log.Info("Starting unmountWhenResourceBusyNonEncrypted test")
	ta.unmountWhenResourceBusy(t, false)
	log.Info("Test unmountWhenResourceBusyNonEncrypted terminated")
}

func (ta *testAPI) unmountWhenResourceBusy(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testDir, err := ioutil.TempDir(os.TempDir(), "unmountResourceBusy")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "ex-upload")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir := filepath.Join(testDir, "ex-mount")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")
	swarmfs := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")
	defer swarmfs.Stop()

	actualPath := filepath.Join(testMountDir, "2.txt")
	//d, err := os.OpenFile(actualPath, os.O_RDWR, os.FileMode(0700))
	d, err := os.Create(actualPath)
	if err != nil {
		t.Fatalf("Couldn't create new file: %v", err)
	}
	_, err = d.Write(getRandomBytes(10))
	if err != nil {
		t.Fatalf("Couldn't write to file: %v", err)
	}
	log.Debug("Bytes written")

	_, err = swarmfs.Unmount(testMountDir)
	if err == nil {
		t.Fatalf("Expected mount to fail due to resource busy, but it succeeded...")
	}
	err = d.Close()
	if err != nil {
		t.Fatalf("Couldn't close file!  %v", bzzHash)
	}
	log.Debug("File closed")

	_, err = swarmfs.Unmount(testMountDir)
	if err != nil {
		t.Fatalf("Expected mount to succeed after freeing resource, but it failed: %v", err)
	}
	mi := swarmfs.Listmounts()
	log.Debug("Going to list mounts")
	for _, minfo := range mi {
		log.Debug("Mount point in list: ", "point", minfo.MountPoint)
		if minfo.MountPoint == testMountDir {
			t.Fatalf("mount state not cleaned up in unmount case %v", testMountDir)
		}
	}
	log.Debug("subtest terminated")
}

func (ta *testAPI) seekInMultiChunkFileEncrypted(t *testing.T) {
	log.Info("Starting seekInMultiChunkFileEncrypted test")
	ta.seekInMultiChunkFile(t, true)
	log.Info("Test seekInMultiChunkFileEncrypted terminated")
}

func (ta *testAPI) seekInMultiChunkFileNonEncrypted(t *testing.T) {
	log.Info("Starting seekInMultiChunkFileNonEncrypted test")
	ta.seekInMultiChunkFile(t, false)
	log.Info("Test seekInMultiChunkFileNonEncrypted terminated")
}

func (ta *testAPI) seekInMultiChunkFile(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testDir, err := ioutil.TempDir(os.TempDir(), "seekInMultiChunkFile")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "seek-upload")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir := filepath.Join(testDir, "seek-mount")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10240)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")
	defer swarmfs.Stop()

	// Create a new file seek the second chunk
	actualPath := filepath.Join(testMountDir, "1.txt")
	d, err := os.OpenFile(actualPath, os.O_RDONLY, os.FileMode(0700))
	if err != nil {
		t.Fatalf("Couldn't open file: %v", err)
	}
	log.Debug("Opened file")
	defer d.Close()

	_, err = d.Seek(5000, 0)
	if err != nil {
		t.Fatalf("Error seeking in file: %v", err)
	}

	contents := make([]byte, 1024)
	_, err = d.Read(contents)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	log.Debug("Read contents")
	finfo := files["1.txt"]

	if !bytes.Equal(finfo.contents[:6024][5000:], contents) {
		t.Fatalf("File seek contents mismatch")
	}
	log.Debug("subtest terminated")
}

func (ta *testAPI) createNewFileEncrypted(t *testing.T) {
	log.Info("Starting createNewFileEncrypted test")
	ta.createNewFile(t, true)
	log.Info("Test createNewFileEncrypted terminated")
}

func (ta *testAPI) createNewFileNonEncrypted(t *testing.T) {
	log.Info("Starting createNewFileNonEncrypted test")
	ta.createNewFile(t, false)
	log.Info("Test createNewFileNonEncrypted terminated")
}

func (ta *testAPI) createNewFile(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testDir, err := ioutil.TempDir(os.TempDir(), "seekInMultiChunkFile")
	if err != nil {
		t.Fatalf("Couldn't create test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	testUploadDir := filepath.Join(testDir, "seek-upload")
	err = os.MkdirAll(testUploadDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create upload dir: %v", err)
	}
	testMountDir := filepath.Join(testDir, "seek-mount")
	err = os.MkdirAll(testMountDir, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	defer swarmfs1.Stop()
	log.Info("Directory mounted")

	// Create a new file in the root dir and check
	actualPath := filepath.Join(testMountDir, "2.txt")
	d, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(0665))
	if err1 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err1)
	}
	log.Debug("Opened file")
	contents := make([]byte, 11)
	_, err = rand.Read(contents)
	if err != nil {
		t.Fatalf("Could not rand read contents %v", err)
	}
	log.Debug("content read")
	_, err = d.Write(contents)
	if err != nil {
		t.Fatalf("Couldn't write contents: %v", err)
	}
	log.Debug("content written")
	err = d.Close()
	if err != nil {
		t.Fatalf("Couldn't close file: %v", err)
	}
	log.Debug("file closed")

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")

	testMountDir2 := filepath.Join(testDir, "create-mount2")
	err = os.MkdirAll(testMountDir2, 0777)
	if err != nil {
		t.Fatalf("Couldn't create mount dir 2: %v", err)
	}
	// mount again and see if things are okay
	files["2.txt"] = fileInfo{0700, 333, 444, contents}
	_ = mountDir(t, ta.api, files, mi.LatestManifest, testMountDir2)
	log.Info("Directory mounted again")

	checkFile(t, testMountDir2, "2.txt", contents)
	log.Debug("subtest terminated")
}

func (ta *testAPI) createNewFileInsideDirectoryEncrypted(t *testing.T) {
	log.Info("Starting createNewFileInsideDirectoryEncrypted test")
	ta.createNewFileInsideDirectory(t, true)
	log.Info("Test createNewFileInsideDirectoryEncrypted terminated")
}

func (ta *testAPI) createNewFileInsideDirectoryNonEncrypted(t *testing.T) {
	log.Info("Starting createNewFileInsideDirectoryNonEncrypted test")
	ta.createNewFileInsideDirectory(t, false)
	log.Info("Test createNewFileInsideDirectoryNonEncrypted terminated")
}

func (ta *testAPI) createNewFileInsideDirectory(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "createinsidedir-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "createinsidedir-mount")
	defer os.RemoveAll(testMountDir)

	files["one/1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	// Create a new file inside a existing dir and check
	dirToCreate := filepath.Join(testMountDir, "one")
	actualPath := filepath.Join(dirToCreate, "2.txt")
	d, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(0665))
	if err1 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err1)
	}
	log.Debug("File opened")
	contents := make([]byte, 11)
	rand.Read(contents)
	log.Debug("Content read")
	d.Write(contents)
	log.Debug("Content written")
	d.Close()
	log.Debug("File closed")

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	testMountDir2, _ := ioutil.TempDir(os.TempDir(), "createinsidedir-mount2")
	defer os.RemoveAll(testMountDir2)
	// mount again and see if things are okay
	files["one/2.txt"] = fileInfo{0700, 333, 444, contents}
	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir2)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()

	checkFile(t, testMountDir2, "one/2.txt", contents)
	log.Debug("subtest terminated")
}

func (ta *testAPI) createNewFileInsideNewDirectoryEncrypted(t *testing.T) {
	log.Info("Starting createNewFileInsideNewDirectoryEncrypted test")
	ta.createNewFileInsideNewDirectory(t, true)
	log.Info("Test createNewFileInsideNewDirectoryEncrypted terminated")
}

func (ta *testAPI) createNewFileInsideNewDirectoryNonEncrypted(t *testing.T) {
	log.Info("Starting createNewFileInsideNewDirectoryNonEncrypted test")
	ta.createNewFileInsideNewDirectory(t, false)
	log.Info("Test createNewFileInsideNewDirectoryNonEncrypted terminated")
}

func (ta *testAPI) createNewFileInsideNewDirectory(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "createinsidenewdir-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "createinsidenewdir-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	// Create a new file inside a existing dir and check
	dirToCreate := filepath.Join(testMountDir, "one")
	os.MkdirAll(dirToCreate, 0777)
	defer os.RemoveAll(dirToCreate)
	actualPath := filepath.Join(dirToCreate, "2.txt")
	d, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(0665))
	if err1 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err1)
	}
	log.Debug("File opened")
	contents := make([]byte, 11)
	rand.Read(contents)
	log.Debug("content read")
	d.Write(contents)
	log.Debug("content written")
	d.Close()
	log.Debug("File closed")

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	files["one/2.txt"] = fileInfo{0700, 333, 444, contents}
	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()

	checkFile(t, testMountDir, "one/2.txt", contents)
	log.Debug("subtest terminated")
}

func (ta *testAPI) removeExistingFileEncrypted(t *testing.T) {
	log.Info("Starting removeExistingFileEncrypted test")
	ta.removeExistingFile(t, true)
	log.Info("Test removeExistingFileEncrypted terminated")
}

func (ta *testAPI) removeExistingFileNonEncrypted(t *testing.T) {
	log.Info("Starting removeExistingFileNonEncrypted test")
	ta.removeExistingFile(t, false)
	log.Info("Test removeExistingFileNonEncrypted terminated")
}

func (ta *testAPI) removeExistingFile(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "remove-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "remove-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	// Remove a file in the root dir and check
	actualPath := filepath.Join(testMountDir, "five.txt")
	os.Remove(actualPath)

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	delete(files, "five.txt")
	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()
}

func (ta *testAPI) removeExistingFileInsideDirEncrypted(t *testing.T) {
	log.Info("Starting removeExistingFileInsideDirEncrypted test")
	ta.removeExistingFileInsideDir(t, true)
	log.Info("Test removeExistingFileInsideDirEncrypted terminated")
}

func (ta *testAPI) removeExistingFileInsideDirNonEncrypted(t *testing.T) {
	log.Info("Starting removeExistingFileInsideDirNonEncrypted test")
	ta.removeExistingFileInsideDir(t, false)
	log.Info("Test removeExistingFileInsideDirNonEncrypted terminated")
}

func (ta *testAPI) removeExistingFileInsideDir(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "remove-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "remove-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["one/five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["one/six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	// Remove a file in the root dir and check
	actualPath := filepath.Join(testMountDir, "one/five.txt")
	os.Remove(actualPath)

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	delete(files, "one/five.txt")
	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()
}

func (ta *testAPI) removeNewlyAddedFileEncrypted(t *testing.T) {
	log.Info("Starting removeNewlyAddedFileEncrypted test")
	ta.removeNewlyAddedFile(t, true)
	log.Info("Test removeNewlyAddedFileEncrypted terminated")
}

func (ta *testAPI) removeNewlyAddedFileNonEncrypted(t *testing.T) {
	log.Info("Starting removeNewlyAddedFileNonEncrypted test")
	ta.removeNewlyAddedFile(t, false)
	log.Info("Test removeNewlyAddedFileNonEncrypted terminated")
}

func (ta *testAPI) removeNewlyAddedFile(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "removenew-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "removenew-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")
	defer swarmfs1.Stop()

	// Adda a new file and remove it
	dirToCreate := filepath.Join(testMountDir, "one")
	os.MkdirAll(dirToCreate, os.FileMode(0665))
	defer os.RemoveAll(dirToCreate)

	actualPath := filepath.Join(dirToCreate, "2.txt")
	d, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(0665))
	if err1 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err1)
	}
	log.Debug("file opened")
	contents := make([]byte, 11)
	rand.Read(contents)
	log.Debug("content read")
	d.Write(contents)
	log.Debug("content written")
	d.Close()
	log.Debug("file closed")

	checkFile(t, testMountDir, "one/2.txt", contents)
	log.Debug("file checked")

	os.Remove(actualPath)

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	testMountDir2, _ := ioutil.TempDir(os.TempDir(), "removenew-mount2")
	defer os.RemoveAll(testMountDir2)
	// mount again and see if things are okay
	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir2)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()

	if bzzHash != mi.LatestManifest {
		t.Fatalf("same contents different hash orig(%v): new(%v)", bzzHash, mi.LatestManifest)
	}
}

func (ta *testAPI) addNewFileAndModifyContentsEncrypted(t *testing.T) {
	log.Info("Starting addNewFileAndModifyContentsEncrypted test")
	ta.addNewFileAndModifyContents(t, true)
	log.Info("Test addNewFileAndModifyContentsEncrypted terminated")
}

func (ta *testAPI) addNewFileAndModifyContentsNonEncrypted(t *testing.T) {
	log.Info("Starting addNewFileAndModifyContentsNonEncrypted test")
	ta.addNewFileAndModifyContents(t, false)
	log.Info("Test addNewFileAndModifyContentsNonEncrypted terminated")
}

func (ta *testAPI) addNewFileAndModifyContents(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "modifyfile-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "modifyfile-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	// Create a new file in the root dir and check
	actualPath := filepath.Join(testMountDir, "2.txt")
	d, err1 := os.OpenFile(actualPath, os.O_RDWR|os.O_CREATE, os.FileMode(0665))
	if err1 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err1)
	}
	log.Debug("file opened")
	line1 := []byte("Line 1")
	rand.Read(line1)
	log.Debug("line read")
	d.Write(line1)
	log.Debug("line written")
	d.Close()
	log.Debug("file closed")

	mi1, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	testMountDir2, _ := ioutil.TempDir(os.TempDir(), "modifyfile-mount2")
	defer os.RemoveAll(testMountDir2)
	// mount again and see if things are okay
	files["2.txt"] = fileInfo{0700, 333, 444, line1}
	swarmfs2 := mountDir(t, ta.api, files, mi1.LatestManifest, testMountDir2)
	log.Info("Directory mounted again")

	checkFile(t, testMountDir2, "2.txt", line1)
	log.Debug("file checked")

	mi2, err3 := swarmfs2.Unmount(testMountDir2)
	if err3 != nil {
		t.Fatalf("Could not unmount %v", err3)
	}
	log.Info("Directory unmounted again")
	swarmfs2.Stop()

	// mount again and modify
	swarmfs3 := mountDir(t, ta.api, files, mi2.LatestManifest, testMountDir)
	log.Info("Directory mounted yet again")

	fd, err4 := os.OpenFile(actualPath, os.O_RDWR|os.O_APPEND, os.FileMode(0665))
	if err4 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err4)
	}
	log.Debug("file opened")
	line2 := []byte("Line 2")
	rand.Read(line2)
	log.Debug("line read")
	fd.Seek(int64(len(line1)), 0)
	fd.Write(line2)
	log.Debug("line written")
	fd.Close()
	log.Debug("file closed")

	mi3, err5 := swarmfs3.Unmount(testMountDir)
	if err5 != nil {
		t.Fatalf("Could not unmount %v", err5)
	}
	log.Info("Directory unmounted yet again")
	swarmfs3.Stop()

	testMountDir4, _ := ioutil.TempDir(os.TempDir(), "modifyfile-mount4")
	defer os.RemoveAll(testMountDir4)
	// mount again and see if things are okay
	b := [][]byte{line1, line2}
	line1and2 := bytes.Join(b, []byte(""))
	files["2.txt"] = fileInfo{0700, 333, 444, line1and2}
	swarmfs4 := mountDir(t, ta.api, files, mi3.LatestManifest, testMountDir4)
	log.Info("Directory mounted final time")
	defer swarmfs4.Stop()

	checkFile(t, testMountDir4, "2.txt", line1and2)
	log.Debug("file checked")
}

func (ta *testAPI) removeEmptyDirEncrypted(t *testing.T) {
	log.Info("Starting removeEmptyDirEncrypted test")
	ta.removeEmptyDir(t, true)
	log.Info("Test removeEmptyDirEncrypted terminated")
}

func (ta *testAPI) removeEmptyDirNonEncrypted(t *testing.T) {
	log.Info("Starting removeEmptyDirNonEncrypted test")
	ta.removeEmptyDir(t, false)
	log.Info("Test removeEmptyDirNonEncrypted terminated")
}

func (ta *testAPI) removeEmptyDir(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "rmdir-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "rmdir-mount")
	defer os.RemoveAll(testMountDir)

	files["1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	defer swarmfs1.Stop()
	log.Info("Directory mounted")

	newdir := filepath.Join(testMountDir, "newdir")
	os.MkdirAll(newdir, 0777)
	defer os.RemoveAll(newdir)

	mi, err3 := swarmfs1.Unmount(testMountDir)
	if err3 != nil {
		t.Fatalf("Could not unmount %v", err3)
	}
	log.Info("Directory unmounted")
	if bzzHash != mi.LatestManifest {
		t.Fatalf("same contents different hash orig(%v): new(%v)", bzzHash, mi.LatestManifest)
	}
}

func (ta *testAPI) removeDirWhichHasFilesEncrypted(t *testing.T) {
	log.Info("Starting removeDirWhichHasFilesEncrypted test")
	ta.removeDirWhichHasFiles(t, true)
	log.Info("Test removeDirWhichHasFilesEncrypted terminated")
}
func (ta *testAPI) removeDirWhichHasFilesNonEncrypted(t *testing.T) {
	log.Info("Starting removeDirWhichHasFilesNonEncrypted test")
	ta.removeDirWhichHasFiles(t, false)
	log.Info("Test removeDirWhichHasFilesNonEncrypted terminated")
}

func (ta *testAPI) removeDirWhichHasFiles(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "rmdir-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "rmdir-mount")
	defer os.RemoveAll(testMountDir)

	files["one/1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/five.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/six.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	dirPath := filepath.Join(testMountDir, "two")
	os.RemoveAll(dirPath)

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v ", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	delete(files, "two/five.txt")
	delete(files, "two/six.txt")

	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir)
	log.Info("Directory mounted")
	defer swarmfs2.Stop()
}

func (ta *testAPI) removeDirWhichHasSubDirsEncrypted(t *testing.T) {
	log.Info("Starting removeDirWhichHasSubDirsEncrypted test")
	ta.removeDirWhichHasSubDirs(t, true)
	log.Info("Test removeDirWhichHasSubDirsEncrypted terminated")
}

func (ta *testAPI) removeDirWhichHasSubDirsNonEncrypted(t *testing.T) {
	log.Info("Starting removeDirWhichHasSubDirsNonEncrypted test")
	ta.removeDirWhichHasSubDirs(t, false)
	log.Info("Test removeDirWhichHasSubDirsNonEncrypted terminated")
}
func (ta *testAPI) removeDirWhichHasSubDirs(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "rmsubdir-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "rmsubdir-mount")
	defer os.RemoveAll(testMountDir)

	files["one/1.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/three/2.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/three/3.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/four/5.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/four/6.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}
	files["two/four/six/7.txt"] = fileInfo{0700, 333, 444, getRandomBytes(10)}

	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	dirPath := filepath.Join(testMountDir, "two")
	os.RemoveAll(dirPath)

	mi, err2 := swarmfs1.Unmount(testMountDir)
	if err2 != nil {
		t.Fatalf("Could not unmount %v ", err2)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	delete(files, "two/three/2.txt")
	delete(files, "two/three/3.txt")
	delete(files, "two/four/5.txt")
	delete(files, "two/four/6.txt")
	delete(files, "two/four/six/7.txt")

	swarmfs2 := mountDir(t, ta.api, files, mi.LatestManifest, testMountDir)
	log.Info("Directory mounted again")
	defer swarmfs2.Stop()
}

func (ta *testAPI) appendFileContentsToEndEncrypted(t *testing.T) {
	log.Info("Starting appendFileContentsToEndEncrypted test")
	ta.appendFileContentsToEnd(t, true)
	log.Info("Test appendFileContentsToEndEncrypted terminated")
}

func (ta *testAPI) appendFileContentsToEndNonEncrypted(t *testing.T) {
	log.Info("Starting appendFileContentsToEndNonEncrypted test")
	ta.appendFileContentsToEnd(t, false)
	log.Info("Test appendFileContentsToEndNonEncrypted terminated")
}

func (ta *testAPI) appendFileContentsToEnd(t *testing.T, toEncrypt bool) {
	files := make(map[string]fileInfo)
	testUploadDir, _ := ioutil.TempDir(os.TempDir(), "appendlargefile-upload")
	defer os.RemoveAll(testUploadDir)
	testMountDir, _ := ioutil.TempDir(os.TempDir(), "appendlargefile-mount")
	defer os.RemoveAll(testMountDir)

	line1 := make([]byte, 10)
	rand.Read(line1)
	files["1.txt"] = fileInfo{0700, 333, 444, line1}
	bzzHash := createTestFilesAndUploadToSwarm(t, ta.api, files, testUploadDir, toEncrypt)
	log.Info("Created test files and uploaded to swarm with uploadDir")

	swarmfs1 := mountDir(t, ta.api, files, bzzHash, testMountDir)
	log.Info("Directory mounted")

	actualPath := filepath.Join(testMountDir, "1.txt")
	fd, err4 := os.OpenFile(actualPath, os.O_RDWR|os.O_APPEND, os.FileMode(0665))
	if err4 != nil {
		t.Fatalf("Could not create file %s : %v", actualPath, err4)
	}
	log.Debug("file opened")
	line2 := make([]byte, 5)
	rand.Read(line2)
	log.Debug("line read")
	fd.Seek(int64(len(line1)), 0)
	fd.Write(line2)
	log.Debug("line written")
	fd.Close()
	log.Debug("file closed")

	mi1, err5 := swarmfs1.Unmount(testMountDir)
	if err5 != nil {
		t.Fatalf("Could not unmount %v ", err5)
	}
	log.Info("Directory unmounted")
	swarmfs1.Stop()

	// mount again and see if things are okay
	b := [][]byte{line1, line2}
	line1and2 := bytes.Join(b, []byte(""))
	files["1.txt"] = fileInfo{0700, 333, 444, line1and2}
	swarmfs2 := mountDir(t, ta.api, files, mi1.LatestManifest, testMountDir)
	log.Info("Directory mounted")
	defer swarmfs2.Stop()

	checkFile(t, testMountDir, "1.txt", line1and2)
	log.Debug("file checked")
}

func TestFUSE(t *testing.T) {
	t.Skip("disable fuse tests until they are stable")
	datadir, err := ioutil.TempDir("", "fuse")
	if err != nil {
		t.Fatalf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(datadir)

	fileStore, err := storage.NewLocalFileStore(datadir, make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	ta := &testAPI{api: api.NewAPI(fileStore, nil, nil)}

	t.Run("mountListAndUnmountEncrypted", ta.mountListAndUnmountEncrypted)
	t.Run("mountListAndUnmountNonEncrypted", ta.mountListAndUnmountNonEncrypted)
	t.Run("maxMountsEncrypted", ta.maxMountsEncrypted)
	t.Run("maxMountsNonEncrypted", ta.maxMountsNonEncrypted)
	t.Run("remountEncrypted", ta.remountEncrypted)
	t.Run("remountNonEncrypted", ta.remountNonEncrypted)
	t.Run("unmountEncrypted", ta.unmountEncrypted)
	t.Run("unmountNonEncrypted", ta.unmountNonEncrypted)
	t.Run("unmountWhenResourceBusyEncrypted", ta.unmountWhenResourceBusyEncrypted)
	t.Run("unmountWhenResourceBusyNonEncrypted", ta.unmountWhenResourceBusyNonEncrypted)
	t.Run("seekInMultiChunkFileEncrypted", ta.seekInMultiChunkFileEncrypted)
	t.Run("seekInMultiChunkFileNonEncrypted", ta.seekInMultiChunkFileNonEncrypted)
	t.Run("createNewFileEncrypted", ta.createNewFileEncrypted)
	t.Run("createNewFileNonEncrypted", ta.createNewFileNonEncrypted)
	t.Run("createNewFileInsideDirectoryEncrypted", ta.createNewFileInsideDirectoryEncrypted)
	t.Run("createNewFileInsideDirectoryNonEncrypted", ta.createNewFileInsideDirectoryNonEncrypted)
	t.Run("createNewFileInsideNewDirectoryEncrypted", ta.createNewFileInsideNewDirectoryEncrypted)
	t.Run("createNewFileInsideNewDirectoryNonEncrypted", ta.createNewFileInsideNewDirectoryNonEncrypted)
	t.Run("removeExistingFileEncrypted", ta.removeExistingFileEncrypted)
	t.Run("removeExistingFileNonEncrypted", ta.removeExistingFileNonEncrypted)
	t.Run("removeExistingFileInsideDirEncrypted", ta.removeExistingFileInsideDirEncrypted)
	t.Run("removeExistingFileInsideDirNonEncrypted", ta.removeExistingFileInsideDirNonEncrypted)
	t.Run("removeNewlyAddedFileEncrypted", ta.removeNewlyAddedFileEncrypted)
	t.Run("removeNewlyAddedFileNonEncrypted", ta.removeNewlyAddedFileNonEncrypted)
	t.Run("addNewFileAndModifyContentsEncrypted", ta.addNewFileAndModifyContentsEncrypted)
	t.Run("addNewFileAndModifyContentsNonEncrypted", ta.addNewFileAndModifyContentsNonEncrypted)
	t.Run("removeEmptyDirEncrypted", ta.removeEmptyDirEncrypted)
	t.Run("removeEmptyDirNonEncrypted", ta.removeEmptyDirNonEncrypted)
	t.Run("removeDirWhichHasFilesEncrypted", ta.removeDirWhichHasFilesEncrypted)
	t.Run("removeDirWhichHasFilesNonEncrypted", ta.removeDirWhichHasFilesNonEncrypted)
	t.Run("removeDirWhichHasSubDirsEncrypted", ta.removeDirWhichHasSubDirsEncrypted)
	t.Run("removeDirWhichHasSubDirsNonEncrypted", ta.removeDirWhichHasSubDirsNonEncrypted)
	t.Run("appendFileContentsToEndEncrypted", ta.appendFileContentsToEndEncrypted)
	t.Run("appendFileContentsToEndNonEncrypted", ta.appendFileContentsToEndNonEncrypted)
}
