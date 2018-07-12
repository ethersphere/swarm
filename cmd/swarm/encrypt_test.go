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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
	swarm "github.com/ethereum/go-ethereum/swarm/api/client"
)

func TestEncrypt(t *testing.T) {
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

	t.Log("ref", ref)

	// upload the file with 'swarm up' and expect a hash
	log.Info(fmt.Sprintf("uploading file with 'swarm up'"))
	up = runSwarm(t,
		"--bzzapi",
		cluster.Nodes[0].URL,
		"encrypt",
		"--password",
		"smth",
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

}
