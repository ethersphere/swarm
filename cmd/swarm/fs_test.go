// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	colorable "github.com/mattn/go-colorable"
)

func init() {
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

type testFile struct {
	filePath string
	content  string
	file     *os.File
}

// TestCLISwarmUp tests that running 'swarm up' makes the resulting file
// available from all nodes via the HTTP API
func TestCLISwarmFs(t *testing.T) {
	log.Info("starting 3 node cluster")
	cluster := newTestCluster(t, 3)
	defer cluster.Shutdown()

	// create a tmp file
	mountPoint, err := ioutil.TempDir("", "swarm-test")
	log.Debug(fmt.Sprintf("1st mount: %s", mountPoint))
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mountPoint)
	handlingNode := cluster.Nodes[0]
	mhash := doUploadFile(t, handlingNode)
	log.Debug("Mounting first run...")

	mount := runSwarm(t, []string{
		"fs",
		"mount",
		"--ipcpath", filepath.Join(handlingNode.Dir, handlingNode.IpcPath),
		mhash,
		mountPoint,
	}...)
	mount.ExpectExit()

	filesToAssert := []*testFile{}

	dirPath, err := createDirInDir(mountPoint, "testSubDir")
	if err != nil {
		t.Fatal(err)
	}
	dirPath2, err := createDirInDir(dirPath, "AnotherTestSubDir")

	dummyContent := "somerandomtestcontentthatshouldbeasserted"
	for _, d := range []string{mountPoint, dirPath, dirPath2} {
		for _, entry := range []string{"f1.tmp", "f2.tmp"} {
			tFile, err := createTestFileInPath(d, entry, dummyContent)
			if err != nil {
				t.Fatal(err)
			}
			filesToAssert = append(filesToAssert, tFile)
		}
	}
	if len(filesToAssert) != 6 {
		t.Fatalf("should have 4 files to assert now, got %d", len(filesToAssert))
	}
	hashRegexp := `[a-f\d]{64}`
	log.Debug("Unmounting first run...")

	unmount := runSwarm(t, []string{
		"fs",
		"unmount",
		"--ipcpath", filepath.Join(handlingNode.Dir, handlingNode.IpcPath),
		mountPoint,
	}...)
	_, matches := unmount.ExpectRegexp(hashRegexp)
	unmount.ExpectExit()

	hash := matches[0]
	if hash == mhash {
		t.Fatal("this should not be equal")
	}
	log.Debug("asserting no files in mount point")

	//check that there's nothing in the mount folder
	files, err := ioutil.ReadDir(mountPoint)
	if err != nil {
		t.Fatalf("had an error reading the directory: %v", err)
	}

	if len(files) != 0 {
		t.Fatal("there shouldn't be anything here")
	}
	log.Debug("Remounting, second run...")

	//remount, check files
	newMount := runSwarm(t, []string{
		"fs",
		"mount",
		"--ipcpath", filepath.Join(handlingNode.Dir, handlingNode.IpcPath),
		hash, // the latest hash
		mountPoint,
	}...)

	newMount.ExpectExit()
	time.Sleep(5 * time.Second)

	files, err = ioutil.ReadDir(mountPoint)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("there should be something here")
	}
	log.Debug("Traversing file tree to see it matches previous mount")

	for _, file := range filesToAssert {
		fmt.Printf("trying to read filepath: %s", file.filePath)
		fileBytes, err := ioutil.ReadFile(file.filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(fileBytes, bytes.NewBufferString(file.content).Bytes()) {
			t.Fatal("this should be equal")
		}
	}

	// cmd := exec.Command("lsof", "+f", "--", mountPoint)
	// log.Debug("Running command and waiting for it to finish...")
	// err = cmd.Run()
	// log.Debug(fmt.Sprintf("Command finished with error: %v", err))
	// if err != nil {
	// 	t.Fatalf("could not exec lsof: %v", err)
	// }

	unmountSec := runSwarm(t, []string{
		"fs",
		"unmount",
		"--ipcpath", filepath.Join(handlingNode.Dir, handlingNode.IpcPath),
		mountPoint,
	}...)

	_, matches = unmountSec.ExpectRegexp(hashRegexp)
	unmountSec.ExpectExit()

	if matches[0] != hash {
		t.Fatal("these should be equal - no changes made")
	}
}

func doUploadFile(t *testing.T, node *testNode) string {
	// create a tmp file
	tmp, err := ioutil.TempFile("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	// write data to file
	data := "randomdata"
	_, err = io.WriteString(tmp, data)
	if err != nil {
		t.Fatal(err)
	}

	hashRegexp := `[a-f\d]{64}`

	flags := []string{
		"--bzzapi", node.URL,
		"up",
		tmp.Name()}

	log.Info(fmt.Sprintf("uploading file with 'swarm up'"))
	up := runSwarm(t, flags...)
	_, matches := up.ExpectRegexp(hashRegexp)
	up.ExpectExit()
	hash := matches[0]
	log.Info("file uploaded", "hash", hash)
	return hash

}

func createDirInDir(createInDir string, dirToCreate string) (string, error) {
	fullpath := filepath.Join(createInDir, dirToCreate)
	err := os.MkdirAll(fullpath, 0777)
	if err != nil {
		return "", err
	}
	log.Debug(fmt.Sprintf("dirindir: %s", fullpath))
	return fullpath, nil
}

func createTestFileInPath(dir, filename, content string) (*testFile, error) {
	tFile := &testFile{}
	filePath := filepath.Join(dir, filename)
	if file, err := os.Create(filePath); err == nil {
		log.Debug(fmt.Sprintf("creatingfile: %s", filePath))

		tFile.file = file
		tFile.content = content
		tFile.filePath = filePath

		_, err = io.WriteString(file, content)
		if err != nil {
			return nil, err
		}
	}

	return tFile, nil
}
