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
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

func download(ctx *cli.Context) {
	isRecursive, isRaw := false
	log.Debug("swarm download")
	args := ctx.Args()
	if len(args) < 1 {
		utils.Fatalf("Usage: swarm download <bzz locator> [<destination path>]")
	}

	newArgs := []string{}

	for _, v := range args {
		if v == "--recursive" {
			isRecursive = true
			log.Debug("swarm download: is recursive")

		}
		if !strings.HasPrefix(v, "--") {
			newArgs = append(newArgs, v)
		}
	}
	args = newArgs

	dir := ""
	filename := ""

	if len(args) == 1 {
		// no destination arg - assume current terminal working dir
		workingDir, err := filepath.Abs("./")
		log.Trace(fmt.Sprintf("swarm download: no destination path - assuming working dir: %s", workingDir))

		if err != nil {
			utils.Fatalf("Fatal: could not get current working directory")
		}
		dir = workingDir
	} else {
		log.Trace(fmt.Sprintf("swarm download: destination path arg: %s", args[1]))
		dir = args[1]
	}

	log.Debug(fmt.Sprintf("working dir: %s", dir))

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

	// assume behaviour accoridng to --recursive switch
	if isRecursive {
		//we are downloading a directory
	} else {
		// we are downloading a file
	}

	// && strings.endsWith(uri.path,'/') == false ??

	manifestList, err := client.List(uri.Addr, uri.Path)
	if err != nil {
		utils.Fatalf("could not list manifest: %v", err)
	}
	manifestLen := len(manifestList.Entries)
	if manifestLen == 0 {
		//err
		utils.Fatalf("could not stat path at address")
	} else if manifestLen == 1 {
		//single file
		v := manifestList.Entries[0]
	} else {
		//multiple files
	}

	if uri.Path != "" {
		// we are downloading a file/path from a manifest
		log.Debug(fmt.Sprintf("swarm download: downloading file/path from a manifest. hash: %s, path:%s", uri.Addr, uri.Path))
		file, err := client.Download(uri.Addr, uri.Path)
		if err != nil {
			utils.Fatalf("could not download %s from given address: %s. error: %v", uri.Path, uri.Addr, err)
		}
		log.Debug(fmt.Sprintf("swarm download: downloaded successfully"))

		re := regexp.MustCompile("[^/]+$") //everything after last slash
		if results := re.FindAllString(uri.Path, -1); len(results) > 0 {
			filename = results[len(results)]
		} else {
			filename = uri.Path
		}
		fileToCreate := path.Join(dir, filename)

		log.Debug(fmt.Sprintf("swarm download: creating outfile at: %s", fileToCreate))
		outFile, err := os.Create(fileToCreate)
		if err != nil {
			utils.Fatalf("could not create file: %v", err)
		}
		defer outFile.Close()

		log.Debug(fmt.Sprintf("swarm download: copying to outfile"))
		_, err = io.Copy(outFile, file)
		if err != nil {
			utils.Fatalf("could not copy response to file: %v", err)
		}
	} else {
		// we are downloading a directory
		log.Debug(fmt.Sprintf("swarm download: downloading directory"))
		err := client.DownloadDirectory(uri.Addr, uri.Path, dir)
		if err != nil {
			utils.Fatalf("could not download directory error: %v", err)
		}
	}
}
