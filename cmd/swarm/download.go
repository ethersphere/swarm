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
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
	swarm "github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

func download(ctx *cli.Context) {
	log.Debug("swarm down")
	args := ctx.Args()
	if len(args) < 1 {
		utils.Fatalf("Usage: swarm down <bzz locator> [<destination path>]")
	}

	var (
		bzzapi      = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		isRecursive = ctx.GlobalBool(SwarmRecursiveUploadFlag.Name)
		client      = swarm.NewClient(bzzapi)
	)

	dir := ""

	if len(args) == 1 {
		// no destination arg - assume current terminal working dir
		workingDir, err := filepath.Abs("./")
		log.Trace(fmt.Sprintf("swarm down: no destination path - assuming working dir: %s", workingDir))

		if err != nil {
			utils.Fatalf("Fatal: could not get current working directory")
		}
		dir = workingDir
	} else {
		log.Trace(fmt.Sprintf("destination path arg: %s", args[1]))
		dir = args[1]
	}

	fi, err := os.Stat(dir)
	if err != nil {
		utils.Fatalf("could not stat path")
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		utils.Fatalf("destination path is not a directory!")
	}

	uri, err := api.Parse(args[0])

	// assume behaviour according to --recursive switch
	if isRecursive {
		if err := client.DownloadDirectory(uri.Addr, uri.Path, dir); err != nil {
			utils.Fatalf("encoutered a fatal error while downloading directory: %v", err)
		}
	} else {
		// we are downloading a file
		log.Debug(fmt.Sprintf("swarm down: downloading file/path from a manifest. hash: %s, path:%s", uri.Addr, uri.Path))

		err := client.DownloadFile(uri.Addr, uri.Path, dir)
		if err != nil {
			utils.Fatalf("could not download %s from given address: %s. error: %v", uri.Path, uri.Addr, err)
		}
	}
}
