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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
)

func TestNoEncrypt(t *testing.T) {
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

	// password := "smth"

	// up = runSwarm(t,
	// 	"--bzzapi",
	// 	cluster.Nodes[0].URL,
	// 	"encrypt",
	// 	"--password",
	// 	password,
	// 	ref,
	// )

	httpClient := &http.Client{}

	url := cluster.Nodes[0].URL + "/" + "bzz-raw:/" + ref
	response, err := httpClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	url = cluster.Nodes[0].URL + "/" + "bzz-raw:/"

	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(d))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", api.ManifestType)

	response, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	d, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	hash := string(bytes.TrimSpace(d))

	url = cluster.Nodes[0].URL + "/" + "bzz:/" + hash

	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, response.StatusCode)
	}
	d, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(d) != data {
		t.Errorf("expected decrypted data %q, got %q", data, string(d))
	}
}
