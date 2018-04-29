// Copyright 2016 The go-ethereum Authors
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
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

func download(ctx *cli.Context) {
	isRecursive := false

	args := ctx.Args()
	if len(args) < 1 {
		utils.Fatalf("Usage: swarm download <bzz locator> [<destination path>]")
	}

	newArgs := []string{}

	for _, v := range args {
		if v == "--recursive" {
			isRecursive = true
		}
		if !strings.HasPrefix(v, "--") {
			newArgs = append(newArgs, v)
		}
	}
	args = newArgs

	dir := ""
	if len(args) == 1 {
		// no destination arg - assume current terminal working dir ./
		workingDir, err := filepath.Abs("./")
		if err != nil {
			utils.Fatalf("Fatal: could not get current working directory")
		}
		dir = workingDir
	} else {
		dir = args[1]
	}

	fmt.Println(dir)

	fi, err := os.Stat(dir)
	if err != nil {
		utils.Fatalf("could not stat path")
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		utils.Fatalf("destination path is not a directory!")
	}

	uri, err := api.Parse(args[0])
	bzzapi := strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
	client := client.NewClient(bzzapi)

	//possible cases:
	// bzz:/addr -> download directory, possible recursive
	// bzz:/addr/path -> download file

	if uri.Path != "" {
		// we are downloading a file/path from a manifest
		file, err := client.Download(uri.Addr, uri.Path)
		if err != nil {
			utils.Fatalf("could not download %s from given address: %s. error: %v", uri.Path, uri.Addr, err)
		}

	} else {
		// we are downloading a directory
		err := client.DownloadDirectory(uri.Addr, uri.Path, dir)
		if err != nil {
			utils.Fatalf("could not download directory error: %v", err)
		}
	}

	if !isRecursive {

	}

}
