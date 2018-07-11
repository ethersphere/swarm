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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto/randentropy"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/api"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/urfave/cli.v1"
)

func encrypt(ctx *cli.Context) {
	args := ctx.Args()

	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
		return
	}

	ref := args[0]

	pass := readPassword()

	fmt.Println(ref)
	fmt.Println(pass)

	salt := randentropy.GetEntropyCSPRNG(32)

	ae, err := api.NewAccessEntryPassword(salt, api.DefaultKdfParams)
	if err != nil {
		utils.Fatalf("Error: %v", err)
		return
	}

	derivedKey, err := api.NewAccessSessionKey(pass, ae)
	if err != nil {
		utils.Fatalf("Error: %v", err)
		return
	}

	encryptKey := derivedKey[:16]

	m := api.ManifestEntry{
		Hash:        hex.EncodeToString(encryptKey),
		ContentType: api.ManifestType,
		//Size:        123, // ?
		ModTime: time.Now(),
		Access:  ae,
	}

	js, err := json.Marshal(m)
	if err != nil {
		utils.Fatalf("Error: %v", err)
		return
	}

	fmt.Println(string(js))
}

// readPassword reads a single line from stdin, trimming it from the trailing new
// line and returns it. The input will not be echoed.
func readPassword() string {
	fmt.Println("Enter password:")
	fmt.Printf("> ")
	text, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Crit("Failed to read password", "err", err)
	}
	fmt.Println()
	return string(text)
}
