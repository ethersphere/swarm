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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/log"
)

func TestEncrypt(t *testing.T) {
	log.Info("starting 3 node cluster")
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
		ref,
	)
	up.InputLine("test")

	_, matches = up.ExpectRegexp(".*")

	fmt.Println(matches)
}
