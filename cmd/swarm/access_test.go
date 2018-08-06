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
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
	swarm "github.com/ethereum/go-ethereum/swarm/api/client"
)

func TestAccess(t *testing.T) {
	cluster := newTestCluster(t, 1)
	defer cluster.Shutdown()

	// create a tmp file
	tmp, err := ioutil.TempDir("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	// write data to file
	data := "notsorandomdata"
	dataFilename := filepath.Join(tmp, "data.txt")

	err = ioutil.WriteFile(dataFilename, []byte(data), 0666)
	if err != nil {
		t.Fatal(err)
	}

	hashRegexp := `[a-f\d]{128}`

	// upload the file with 'swarm up' and expect a hash
	log.Info(fmt.Sprintf("uploading file with 'swarm up'"))
	up := runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"up",
		"--encrypt",
		dataFilename)
	_, matches := up.ExpectRegexp(hashRegexp)
	up.ExpectExit()

	if len(matches) < 1 {
		t.Fatal("no matches found")
	}

	ref := matches[0]

	password := "smth"

	passwordFilename := filepath.Join(tmp, "password.txt")

	err = ioutil.WriteFile(passwordFilename, []byte(password), 0666)
	if err != nil {
		t.Fatal(err)
	}

	up = runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"access",
		"new",
		"pass",
		"--dry-run",
		"--password",
		passwordFilename,
		ref,
	)

	_, matches = up.ExpectRegexp(".+")
	up.ExpectExit()

	if len(matches) == 0 {
		t.Fatalf("stdout not matched")
	}

	var m api.Manifest

	err = json.Unmarshal([]byte(matches[0]), &m)
	if err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if len(m.Entries) != 1 {
		t.Fatalf("expected one manifest entry, got %v", len(m.Entries))
	}

	e := m.Entries[0]

	ct := "application/bzz-manifest+json"
	if e.ContentType != ct {
		t.Errorf("expected %q content type, got %q", ct, e.ContentType)
	}

	if e.Access == nil {
		t.Fatal("manifest access is nil")
	}

	a := e.Access

	if a.Type != "pass" {
		t.Errorf(`got access type %q, expected "pass"`, a.Type)
	}
	if len(a.Salt) < 32 {
		t.Errorf(`got salt with length %v, expected not less the 32 bytes`, len(a.Salt))
	}
	if a.KdfParams == nil {
		t.Fatal("manifest access kdf params is nil")
	}

	client := swarm.NewClient(cluster.Nodes[0].URL)

	hash, err := client.UploadManifest(&m, false)
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{}

	url := cluster.Nodes[0].URL + "/" + "bzz:/" + hash
	response, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatal("should be a 401")
	}
	authHeader := response.Header.Get("WWW-Authenticate")
	if authHeader == "" {
		t.Fatal("should be something here")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("", password)

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, response.StatusCode)
	}
	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != data {
		t.Errorf("expected decrypted data %q, got %q", data, string(d))
	}

	log.Info("download file with 'swarm down'")
	up = runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"down",
		"bzz:/"+hash,
		tmp,
		"--password",
		passwordFilename)

	up.ExpectExit()

	wrongPasswordFilename := filepath.Join(tmp, "password-wrong.txt")

	err = ioutil.WriteFile(wrongPasswordFilename, []byte("just wr0ng"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	log.Info("download file with 'swarm down' with wrong password")
	up = runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"down",
		"bzz:/"+hash,
		tmp,
		"--password",
		wrongPasswordFilename)

	_, matches = up.ExpectRegexp("unauthorized")
	if len(matches) != 1 && matches[0] != "unauthorized" {
		t.Fatal(`"unauthorized" not found in output"`)
	}
	up.ExpectExit()
}

func TestAccessPK(t *testing.T) {
	// Setup Swarm and upload a test file to it
	cluster := newTestCluster(t, 1)
	defer cluster.Shutdown()

	// create a tmp file
	tmp, err := ioutil.TempFile("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	// write data to file
	data := "notsorandomdata"
	_, err = io.WriteString(tmp, data)
	if err != nil {
		t.Fatal(err)
	}

	hashRegexp := `[a-f\d]{128}`

	// upload the file with 'swarm up' and expect a hash
	log.Info(fmt.Sprintf("uploading file with 'swarm up'"))
	up := runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"up",
		"--encrypt",
		tmp.Name())
	_, matches := up.ExpectRegexp(hashRegexp)
	up.ExpectExit()

	if len(matches) < 1 {
		t.Fatal("no matches found")
	}

	ref := matches[0]

	pk := cluster.Nodes[0].PrivateKey
	granteePubKey := crypto.FromECDSAPub(&pk.PublicKey)

	publisherDir, err := ioutil.TempDir("", "swarm-account-dir-temp")
	if err != nil {
		t.Fatal(err)
	}

	passFile, err := ioutil.TempFile("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer passFile.Close()
	defer os.Remove(passFile.Name())
	_, err = io.WriteString(passFile, testPassphrase)
	if err != nil {
		t.Fatal(err)
	}

	_, publisherAccount := getTestAccount(t, publisherDir)
	up = runSwarm(t,
		"--bzzaccount",
		publisherAccount.Address.String(),
		"--password",
		passFile.Name(),
		"--datadir",
		publisherDir,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"access",
		"new",
		"pk",
		"--dry-run",
		"--grant-pk",
		hex.EncodeToString(granteePubKey),
		ref,
	)

	_, matches = up.ExpectRegexp(".+")
	up.ExpectExit()

	if len(matches) == 0 {
		t.Fatalf("stdout not matched")
	}

	var m api.Manifest

	err = json.Unmarshal([]byte(matches[0]), &m)
	if err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if len(m.Entries) != 1 {
		t.Fatalf("expected one manifest entry, got %v", len(m.Entries))
	}

	e := m.Entries[0]

	ct := "application/bzz-manifest+json"
	if e.ContentType != ct {
		t.Errorf("expected %q content type, got %q", ct, e.ContentType)
	}

	if e.Access == nil {
		t.Fatal("manifest access is nil")
	}

	a := e.Access

	if a.Type != "pk" {
		t.Errorf(`got access type %q, expected "pk"`, a.Type)
	}
	if len(a.Salt) < 32 {
		t.Errorf(`got salt with length %v, expected not less the 32 bytes`, len(a.Salt))
	}
	if a.KdfParams != nil {
		t.Fatal("manifest access kdf params should be nil")
	}

	client := swarm.NewClient(cluster.Nodes[0].URL)

	hash, err := client.UploadManifest(&m, false)
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{}

	url := cluster.Nodes[0].URL + "/" + "bzz:/" + hash
	response, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatal("should be a 200")
	}
	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != data {
		t.Errorf("expected decrypted data %q, got %q", data, string(d))
	}

}
func TestAccessACT(t *testing.T) {
	// Setup Swarm and upload a test file to it
	cluster := newTestCluster(t, 3)
	defer cluster.Shutdown()

	// create a tmp file
	tmp, err := ioutil.TempFile("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	// write data to file
	data := "notsorandomdata"
	_, err = io.WriteString(tmp, data)
	if err != nil {
		t.Fatal(err)
	}

	hashRegexp := `[a-f\d]{128}`

	// upload the file with 'swarm up' and expect a hash
	log.Info(fmt.Sprintf("uploading file with 'swarm up'"))
	up := runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"up",
		"--encrypt",
		tmp.Name())
	_, matches := up.ExpectRegexp(hashRegexp)
	up.ExpectExit()

	if len(matches) < 1 {
		t.Fatal("no matches found")
	}

	ref := matches[0]

	grantees := []string{}
	for _, v := range cluster.Nodes {
		pk := v.PrivateKey
		granteePubKey := crypto.CompressPubkey(&pk.PublicKey)
		grantees = append(grantees, hex.EncodeToString(granteePubKey))
	}

	granteesPubkeyListFile, err := ioutil.TempFile("", "grantees-pubkey-list.csv")
	if err != nil {
		t.Fatal(err)
	}

	_, err = granteesPubkeyListFile.WriteString(strings.Join(grantees, "\n"))
	if err != nil {
		t.Fatal(err)
	}

	defer granteesPubkeyListFile.Close()
	defer os.Remove(granteesPubkeyListFile.Name())

	publisherDir, err := ioutil.TempDir("", "swarm-account-dir-temp")
	if err != nil {
		t.Fatal(err)
	}

	passFile, err := ioutil.TempFile("", "swarm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer passFile.Close()
	defer os.Remove(passFile.Name())
	_, err = io.WriteString(passFile, testPassphrase)
	if err != nil {
		t.Fatal(err)
	}

	_, publisherAccount := getTestAccount(t, publisherDir)
	up = runSwarm(t,
		"--bzzaccount",
		publisherAccount.Address.String(),
		"--password",
		passFile.Name(),
		"--datadir",
		publisherDir,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"access",
		"new",
		"act",
		"--dry-run",
		"--grant-keys",
		granteesPubkeyListFile.Name(),
		ref,
	)

	_, matches = up.ExpectRegexp(".+")
	up.ExpectExit()

	if len(matches) == 0 {
		t.Fatalf("stdout not matched")
	}

	var m api.Manifest

	err = json.Unmarshal([]byte(matches[0]), &m)
	if err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if len(m.Entries) != 1 {
		t.Fatalf("expected one manifest entry, got %v", len(m.Entries))
	}

	e := m.Entries[0]

	ct := "application/bzz-manifest+json"
	if e.ContentType != ct {
		t.Errorf("expected %q content type, got %q", ct, e.ContentType)
	}

	if e.Access == nil {
		t.Fatal("manifest access is nil")
	}

	a := e.Access

	if a.Type != "act" {
		t.Fatalf(`got access type %q, expected "act"`, a.Type)
	}
	if len(a.Salt) < 32 {
		t.Fatalf(`got salt with length %v, expected not less the 32 bytes`, len(a.Salt))
	}
	if a.KdfParams != nil {
		t.Fatal("manifest access kdf params should be nil")
	}

	client := swarm.NewClient(cluster.Nodes[0].URL)

	hash, err := client.UploadManifest(&m, false)
	if err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{}

	url := cluster.Nodes[0].URL + "/" + "bzz:/" + hash
	response, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatal("should be a 200")
	}
	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != data {
		t.Errorf("expected decrypted data %q, got %q", data, string(d))
	}

}

func TestAccessPKUnit(t *testing.T) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	sharedSecret := "a85586744a1ddd56a7ed9f33fa24f40dd745b3a941be296a0d60e329dbdb896d"

	for i, v := range []struct {
		publisherPriv string
		granteePub    string
	}{
		{
			publisherPriv: "ec5541555f3bc6376788425e9d1a62f55a82901683fd7062c5eddcc373a73459",
			granteePub:    "0226f213613e843a413ad35b40f193910d26eb35f00154afcde9ded57479a6224a",
		},
		{
			publisherPriv: "70c7a73011aa56584a0009ab874794ee7e5652fd0c6911cd02f8b6267dd82d2d",
			granteePub:    "02e6f8d5e28faaa899744972bb847b6eb805a160494690c9ee7197ae9f619181db",
		},
	} {
		b, _ := hex.DecodeString(v.granteePub)
		granteePub, _ := crypto.DecompressPubkey(b)
		publisherPrivate, _ := crypto.HexToECDSA(v.publisherPriv)

		ssKey, err := api.NewSessionKeyPK(publisherPrivate, granteePub, salt)
		if err != nil {
			t.Fatal(err)
		}

		hasher := sha3.NewKeccak256()
		hasher.Write(salt)
		shared, err := hex.DecodeString(sharedSecret)
		if err != nil {
			t.Fatal(err)
		}
		hasher.Write(shared)
		sum := hasher.Sum(nil)

		if !bytes.Equal(ssKey, sum) {
			t.Fatalf("%d: got a session key mismatch", i)
		}
	}

}
