// Copyright 2018 The go-ethereum Authors
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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/cmd/swarm/testdata"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm"
	"github.com/ethereum/go-ethereum/swarm/testutil"
)

const (
	DATABASE_FIXTURE_BZZ_ACCOUNT = "0aa159029fa13ffa8fa1c6fff6ebceface99d6a4"
	DATABASE_FIXTURE_PASSWORD    = "pass"
	FIXTURE_DATADIR_PREFIX       = "swarm/bzz-0aa159029fa13ffa8fa1c6fff6ebceface99d6a4"
	FixtureBaseKey               = "a9f22b3d77b4bdf5f3eefce995d6c8e7cecf2636f20956f08a0d1ed95adb52ad"
)

// TestCLISwarmExportImport perform the following test:
// 1. runs swarm node
// 2. uploads a random file
// 3. runs an export of the local datastore
// 4. runs a second swarm node
// 5. imports the exported datastore
// 6. fetches the uploaded random file from the second node
func TestCLISwarmExportImport(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	cluster := newTestCluster(t, 1)

	// generate random 1mb file
	content := testutil.RandomBytes(1, 1000000)
	fileName := testutil.TempFileWithContent(t, string(content))
	defer os.Remove(fileName)

	// upload the file with 'swarm up' and expect a hash
	up := runSwarm(t, "--bzzapi", cluster.Nodes[0].URL, "up", fileName)
	_, matches := up.ExpectRegexp(`[a-f\d]{64}`)
	up.ExpectExit()
	hash := matches[0]

	var info swarm.Info
	if err := cluster.Nodes[0].Client.Call(&info, "bzz_info"); err != nil {
		t.Fatal(err)
	}
	cluster.Stop()
	defer cluster.Cleanup()

	// generate an export.tar
	exportCmd := runSwarm(t, "db", "export", info.Path+"/chunks", info.Path+"/export.tar", strings.TrimPrefix(info.BzzKey, "0x"))
	exportCmd.ExpectExit()

	// start second cluster
	cluster2 := newTestCluster(t, 1)

	var info2 swarm.Info
	if err := cluster2.Nodes[0].Client.Call(&info2, "bzz_info"); err != nil {
		t.Fatal(err)
	}

	// stop second cluster, so that we close LevelDB
	cluster2.Stop()
	defer cluster2.Cleanup()

	// import the export.tar
	importCmd := runSwarm(t, "db", "import", info2.Path+"/chunks", info.Path+"/export.tar", strings.TrimPrefix(info2.BzzKey, "0x"))
	importCmd.ExpectExit()

	// spin second cluster back up
	cluster2.StartExistingNodes(t, 1, strings.TrimPrefix(info2.BzzAccount, "0x"))

	// try to fetch imported file
	res, err := http.Get(cluster2.Nodes[0].URL + "/bzz:/" + hash)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("expected HTTP status %d, got %s", 200, res.Status)
	}

	// compare downloaded file with the generated random file
	mustEqualFiles(t, bytes.NewReader(content), res.Body)
}

func TestExportLegacyToNew(t *testing.T) {
	// changes to the swarm code:
	/*
		bzz account 0aa159029fa13ffa8fa1c6fff6ebceface99d6a4
		password qwerty
		uploaded file hash http://localhost:8500/bzz:/67a86082ee0ea1bc7dd8d955bb1e14d04f61d55ae6a4b37b3d0296a3a95e454a/

		open LDB on startup, try to get old DB schema key from the db.
			if its valid - ask for migration.
			if not - check new db schema key validity.
			migration should be also correct with last shcema name that requires the export+import migration
			and check that it was done successfully
			user should delete the folder after export so that the new db gets created
		v0. create old database tarball as a fixture, push to codebase
		v1. unpack fixture to tmp dir
		v2. try to open new swarm that should complain about schema changes but with same datadir
		v3. catch the complaint
		v4. do the export
		v5. remove the folder
		vx6. try to reopen with new swarm - file should not be retrievable
		-7. close
			8. try to import the dump
			9. file should be accessible
	*/

	const UPLOADED_HASH = "67a86082ee0ea1bc7dd8d955bb1e14d04f61d55ae6a4b37b3d0296a3a95e454a"
	tmpdir, err := ioutil.TempDir("", "swarm-test")
	fmt.Println(tmpdir)
	//	defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	deflateBase64Gzip(t, testdata.DATADIR_MIGRATION_FIXTURE, tmpdir)

	tmpPassword := testutil.TempFileWithContent(t, DATABASE_FIXTURE_PASSWORD)
	defer os.Remove(tmpPassword)

	flags := []string{
		"--datadir", tmpdir,
		"--bzzaccount", DATABASE_FIXTURE_BZZ_ACCOUNT,
		"--password", tmpPassword,
	}

	newSwarmOldDb := runSwarm(t, flags...)
	_, matches := newSwarmOldDb.ExpectRegexp(".+")
	newSwarmOldDb.ExpectExit()

	if len(matches) == 0 {
		t.Fatalf("stdout not matched")
	}

	if newSwarmOldDb.ExitStatus() == 0 {
		t.Fatal("should error")
	}
	t.Log("exporting legacy database")
	actualDataDir := path.Join(tmpdir, FIXTURE_DATADIR_PREFIX)
	exportCmd := runSwarm(t, "--verbosity", "5", "db", "export", actualDataDir+"/chunks", tmpdir+"/export.tar", FixtureBaseKey)
	exportCmd.ExpectExit()

	stat, err := os.Stat(tmpdir + "/export.tar")
	if err != nil {
		t.Fatal(err)
	}

	// make some silly size assumption
	if stat.Size() < 90000 {
		t.Fatal("export size too small")
	}
	t.Log("removing chunk datadir")
	err = os.RemoveAll(path.Join(actualDataDir, "chunks"))
	if err != nil {
		t.Fatal(err)
	}

	// start second cluster
	cluster2 := newTestCluster(t, 1)
	var info2 swarm.Info
	if err := cluster2.Nodes[0].Client.Call(&info2, "bzz_info"); err != nil {
		t.Fatal(err)
	}

	// stop second cluster, so that we close LevelDB
	cluster2.Stop()
	defer cluster2.Cleanup()

	t.Log("importing from legacy db")
	// import the export.tar
	importCmd := runSwarm(t, "db", "import", info2.Path+"/chunks", tmpdir+"/export.tar", strings.TrimPrefix(info2.BzzKey, "0x"))
	importCmd.ExpectExit()

	// spin second cluster back up
	cluster2.StartExistingNodes(t, 1, strings.TrimPrefix(info2.BzzAccount, "0x"))
	t.Log("trying to http get the file")
	// try to fetch imported file
	res, err := http.Get(cluster2.Nodes[0].URL + "/bzz:/" + UPLOADED_HASH)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("expected HTTP status %d, got %s", 200, res.Status)
	}

	// compare downloaded file with the generated random file
	//mustEqualFiles(t, bytes.NewReader(content), res.Body)

}

func mustEqualFiles(t *testing.T, up io.Reader, down io.Reader) {
	h := md5.New()
	upLen, err := io.Copy(h, up)
	if err != nil {
		t.Fatal(err)
	}
	upHash := h.Sum(nil)
	h.Reset()
	downLen, err := io.Copy(h, down)
	if err != nil {
		t.Fatal(err)
	}
	downHash := h.Sum(nil)

	if !bytes.Equal(upHash, downHash) || upLen != downLen {
		t.Fatalf("downloaded imported file md5=%x (length %v) is not the same as the generated one mp5=%x (length %v)", downHash, downLen, upHash, upLen)
	}
}

func deflateBase64Gzip(t *testing.T, base64File, directory string) {
	t.Helper()

	f := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64File))

	gzf, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}

	tarReader := tar.NewReader(gzf)

	i := 0
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatal(err)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			log.Debug("deflating tgz", "creating directory", path.Join(directory, name))
			err := os.Mkdir(path.Join(directory, name), os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			log.Debug("deflating tgz", "creating file", path.Join(directory, name))
			file, err := os.Create(path.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("shouldn't happen")
		}

		i++
	}

}
