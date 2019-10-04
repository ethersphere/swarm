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

// Command bzzup uploads files to the swarm HTTP API.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethersphere/swarm/api/client"
	swarm "github.com/ethersphere/swarm/api/client"
	"github.com/ethersphere/swarm/chunk"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	upCommand = cli.Command{
		Action:             upload,
		CustomHelpTemplate: helpTemplate,
		Name:               "up",
		Usage:              "uploads a file or directory to swarm using the HTTP API",
		ArgsUsage:          "<file>",
		Flags:              []cli.Flag{SwarmEncryptedFlag, SwarmPinFlag, SwarmNoTrackUploadFlag, SwarmVerboseFlag},
		Description:        "uploads a file or directory to swarm using the HTTP API and prints the root hash",
	}

	pollDelay   = 200 * time.Millisecond
	chunkStates = []struct {
		name  string
		state chunk.State
	}{
		{"Split", chunk.StateSplit},
		{"Stored", chunk.StateStored},
		{"Sent", chunk.StateSent},
		{"Synced", chunk.StateSynced},
	}
)

func upload(ctx *cli.Context) {
	args := ctx.Args()
	var (
		bzzapi          = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		recursive       = ctx.GlobalBool(SwarmRecursiveFlag.Name)
		wantManifest    = ctx.GlobalBoolT(SwarmWantManifestFlag.Name)
		defaultPath     = ctx.GlobalString(SwarmUploadDefaultPath.Name)
		fromStdin       = ctx.GlobalBool(SwarmUpFromStdinFlag.Name)
		mimeType        = ctx.GlobalString(SwarmUploadMimeType.Name)
		verbose         = ctx.Bool(SwarmVerboseFlag.Name)
		client          = swarm.NewClient(bzzapi)
		toEncrypt       = ctx.Bool(SwarmEncryptedFlag.Name)
		toPin           = ctx.Bool(SwarmPinFlag.Name)
		notrack         = ctx.Bool(SwarmNoTrackUploadFlag.Name)
		autoDefaultPath = false
		file            string
	)
	if !verbose {
		chunkStates = chunkStates[3:] // just poll Synced state
	}
	if autoDefaultPathString := os.Getenv(SwarmAutoDefaultPath); autoDefaultPathString != "" {
		b, err := strconv.ParseBool(autoDefaultPathString)
		if err != nil {
			utils.Fatalf("invalid environment variable %s: %v", SwarmAutoDefaultPath, err)
		}
		autoDefaultPath = b
	}
	if len(args) != 1 {
		if fromStdin {
			tmp, err := ioutil.TempFile("", "swarm-stdin")
			if err != nil {
				utils.Fatalf("error create tempfile: %s", err)
			}
			defer os.Remove(tmp.Name())
			n, err := io.Copy(tmp, os.Stdin)
			if err != nil {
				utils.Fatalf("error copying stdin to tempfile: %s", err)
			} else if n == 0 {
				utils.Fatalf("error reading from stdin: zero length")
			}
			file = tmp.Name()
		} else {
			utils.Fatalf("Need filename as the first and only argument")
		}
	} else {
		file = expandPath(args[0])
	}

	if !wantManifest {
		f, err := swarm.Open(file)
		if err != nil {
			utils.Fatalf("Error opening file: %s", err)
		}
		defer f.Close()
		hash, err := client.UploadRaw(f, f.Size, toEncrypt, toPin, true)
		if err != nil {
			utils.Fatalf("Upload failed: %s", err)
		}
		fmt.Println(hash)
		return
	}

	stat, err := os.Stat(file)
	if err != nil {
		utils.Fatalf("Error opening file: %s", err)
	}

	// define a function which either uploads a directory or single file
	// based on the type of the file being uploaded
	var doUpload func() (hash string, err error)
	if stat.IsDir() {
		doUpload = func() (string, error) {
			if !recursive {
				return "", errors.New("Argument is a directory and recursive upload is disabled")
			}
			if autoDefaultPath && defaultPath == "" {
				defaultEntryCandidate := path.Join(file, "index.html")
				log.Debug("trying to find default path", "path", defaultEntryCandidate)
				defaultEntryStat, err := os.Stat(defaultEntryCandidate)
				if err == nil && !defaultEntryStat.IsDir() {
					log.Debug("setting auto detected default path", "path", defaultEntryCandidate)
					defaultPath = defaultEntryCandidate
				}
			}
			if defaultPath != "" {
				// construct absolute default path
				absDefaultPath, _ := filepath.Abs(defaultPath)
				absFile, _ := filepath.Abs(file)
				// make sure absolute directory ends with only one "/"
				// to trim it from absolute default path and get relative default path
				absFile = strings.TrimRight(absFile, "/") + "/"
				if absDefaultPath != "" && absFile != "" && strings.HasPrefix(absDefaultPath, absFile) {
					defaultPath = strings.TrimPrefix(absDefaultPath, absFile)
				}
			}
			return client.UploadDirectory(file, defaultPath, "", toEncrypt, toPin, true)
		}
	} else {
		doUpload = func() (string, error) {
			f, err := swarm.Open(file)
			if err != nil {
				return "", fmt.Errorf("error opening file: %s", err)
			}
			defer f.Close()
			if mimeType != "" {
				f.ContentType = mimeType
			}
			return client.Upload(f, "", toEncrypt, toPin, true)
		}
	}
	hash, err := doUpload()
	if err != nil {
		utils.Fatalf("Upload failed: %s", err)
	}

	// dont show the progress bar (machine readable output)
	if notrack {
		fmt.Println(hash)
		return
	}

	// this section renders the cli UI for showing the progress bars
	tag, err := client.TagByHash(hash)
	if err != nil {
		utils.Fatalf("failed to get tag data for hash: %v", err)
	}
	fmt.Println("Swarm Hash:", hash)
	fmt.Println("Tag UID:", tag.Uid)
	// check if the user uploaded something that was already completely stored
	// in the local store (otherwise we hang forever because there's nothing to sync)
	// as the chunks are already supposed to be synced
	seen, total, err := tag.Status(chunk.StateSeen)
	if total-seen > 0 {
		fmt.Println("Upload status:")
		bars := createTagBars(tag, verbose)
		pollTag(client, tag, bars)
	}

	fmt.Println("Done! took", time.Since(tag.StartedAt))
	fmt.Println("Your Swarm hash should now be retrievable from other nodes!")
}

func pollTag(client *client.Client, tag *chunk.Tag, bars map[string]*mpb.Bar) {
	oldTag := *tag
	lastTime := time.Now()

	for {
		time.Sleep(pollDelay)
		newTag, err := client.TagByHash(tag.Address.String())
		if err != nil {
			utils.Fatalf("had an error polling the tag for address %s, err %v", tag.Address.String(), err)
		}
		done := true
		for _, state := range chunkStates {
			// calculate the difference that we need to increment for each bar
			count, _, err := oldTag.Status(state.state)
			if err != nil {
				utils.Fatalf("error while getting tag status: %v", err)
			}
			newCount, total, err := newTag.Status(state.state)
			if err != nil {
				utils.Fatalf("error while getting tag status: %v", err)
			}
			d := int(newCount - count)
			if newCount != total {
				done = false
			}
			bars[state.name].SetTotal(total, done)
			bars[state.name].IncrBy(d, time.Since(lastTime))
		}
		if done {
			return
		}

		oldTag = *newTag
		lastTime = time.Now()
	}
}

func createTagBars(tag *chunk.Tag, verbose bool) map[string]*mpb.Bar {
	p := mpb.New(mpb.WithWidth(64))
	bars := make(map[string]*mpb.Bar)
	for _, state := range chunkStates {
		count, total, err := tag.Status(state.state)
		if err != nil {
			utils.Fatalf("could not get tag status: %v", err)
		}
		title := state.name
		var barElement *mpb.Bar
		width := 10
		if verbose {
			barElement = p.AddBar(total,
				mpb.PrependDecorators(
					// align the elements with a constant size (10 chars)
					decor.Name(title, decor.WC{W: width, C: decor.DidentRight}),
					// add unit counts
					decor.CountersNoUnit("%d / %d", decor.WCSyncSpace),
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60, and width reservation of 4
						decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 6}), "done",
					),
				),
				mpb.AppendDecorators(decor.Percentage()),
			)
		} else {
			title = fmt.Sprintf("Syncing %d chunks", total)
			width = len(title) + 3
			barElement = p.AddBar(total,
				mpb.PrependDecorators(
					// align the elements with a constant size (10 chars)
					decor.Name(title, decor.WC{W: width, C: decor.DidentRight}),
					// replace ETA decorator with "done" message, OnComplete event
					decor.OnComplete(
						// ETA decorator with ewma age of 60, and width reservation of 4
						decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 6}), "done",
					),
				),
				mpb.AppendDecorators(decor.Percentage()),
			)
		}
		// increment the bar with the initial value from the tag
		barElement.IncrBy(int(count))
		bars[state.name] = barElement
	}
	return bars
}

// Expands a file path
// 1. replace tilde with users home dir
// 2. expands embedded environment variables
// 3. cleans the path, e.g. /a/b/../c -> /a/c
// Note, it has limitations, e.g. ~someuser/tmp will not be expanded
func expandPath(p string) string {
	if i := strings.Index(p, ":"); i > 0 {
		return p
	}
	if i := strings.Index(p, "@"); i > 0 {
		return p
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := homeDir(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
