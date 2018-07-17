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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/randentropy"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

var salt = randentropy.GetEntropyCSPRNG(32)

// This init function sets defaults so cmd/swarm can run alongside geth.
func init() {
	defaultNodeConfig.Name = clientIdentifier
	defaultNodeConfig.Version = params.VersionWithCommit(gitCommit)
	defaultNodeConfig.P2P.ListenAddr = ":30399"
	defaultNodeConfig.IPCPath = "bzzd.ipc"
	// Set flag defaults for --help display.
	utils.ListenPortFlag.Value = 30399
}

func accessNewPass(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
	}

	var (
		ae         *api.AccessEntry
		sessionKey []byte
		err        error
		ref        = args[0]
	)
	sessionKey, ae, err = doPasswordNew(ctx, salt)
	if err != nil {
		utils.Fatalf("error getting session key: %v", err)
	}
	generateAccessControlManifest(ctx, ref, sessionKey, ae)
}

func accessNewPK(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
	}

	var (
		ae         *api.AccessEntry
		sessionKey []byte
		err        error
		ref        = args[0]
	)
	sessionKey, ae, err = doPKNew(ctx, salt)
	if err != nil {
		utils.Fatalf("error getting session key: %v", err)
	}
	generateAccessControlManifest(ctx, ref, sessionKey, ae)
}

func generateAccessControlManifest(ctx *cli.Context, ref string, sessionKey []byte, ae *api.AccessEntry) {
	dryRun := ctx.Bool(SwarmDryRunFlag.Name)
	refBytes, err := hex.DecodeString(ref)
	if err != nil {
		utils.Fatalf("Error: %v", err)
	}
	// encrypt ref with sessionKey
	enc := api.NewRefEncryption(len(refBytes))
	encrypted, err := enc.Encrypt(refBytes, sessionKey)
	if err != nil {
		utils.Fatalf("Error: %v", err)
	}

	m := api.Manifest{
		Entries: []api.ManifestEntry{
			api.ManifestEntry{
				Hash:        hex.EncodeToString(encrypted),
				ContentType: api.ManifestType,
				//Size:        123, // ?
				ModTime: time.Now(),
				Access:  ae,
			},
		},
	}

	if dryRun {
		js, err := json.Marshal(m)
		if err != nil {
			utils.Fatalf("Error: %v", err)
		}

		fmt.Println(string(js))
	} else {
		bzzapi := strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client := client.NewClient(bzzapi)

		key, err := client.UploadManifest(&m, false)
		if err != nil {
			utils.Fatalf("Error uploading manifest: %v", err)
		}
		fmt.Println(key)
	}
}

func doPKNew(ctx *cli.Context, salt []byte) (sessionKey []byte, ae *api.AccessEntry, err error) {
	// booting up the swarm node just as we do in bzzd action
	bzzconfig, err := buildConfig(ctx)
	if err != nil {
		utils.Fatalf("unable to configure swarm: %v", err)
	}
	cfg := defaultNodeConfig
	if _, err := os.Stat(bzzconfig.Path); err == nil {
		cfg.DataDir = bzzconfig.Path
	}
	utils.SetNodeConfig(ctx, &cfg)
	stack, err := node.New(&cfg)
	if err != nil {
		utils.Fatalf("can't create node: %v", err)
	}
	initSwarmNode(bzzconfig, stack, ctx)
	privateKey := getAccount(bzzconfig.BzzAccount, ctx, stack)

	granteePublicKey := ctx.String(SwarmAccessGrantPKFlag.Name)

	if granteePublicKey == "" {
		return nil, nil, errors.New("need a grantee Public Key")
	}
	b, err := hex.DecodeString(granteePublicKey)
	if err != nil {
		log.Error("error decoding grantee public key", "err", err)
		return nil, nil, err
	}

	granteePub, err := crypto.UnmarshalPubkey(b)
	if err != nil {
		log.Error("error unmarshaling grantee public key", "err", err)
		return nil, nil, err
	}

	sessionKey, err = api.NewSessionKeyPK(privateKey, granteePub, salt)
	if err != nil {
		log.Error("error getting session key", "err", err)
		return nil, nil, err
	}

	ae, err = api.NewAccessEntryPK(hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey)), salt)
	if err != nil {
		log.Error("error generating access entry", "err", err)
		return nil, nil, err
	}

	return sessionKey, ae, nil
}

func doPasswordNew(ctx *cli.Context, salt []byte) (sessionKey []byte, ae *api.AccessEntry, err error) {
	password := getPassPhrase("", 0, makePasswordList(ctx))
	ae, err = api.NewAccessEntryPassword(salt, api.DefaultKdfParams)
	if err != nil {
		return nil, nil, err
	}

	sessionKey, err = api.NewSessionKeyPassword(password, ae)
	if err != nil {
		return nil, nil, err
	}
	return sessionKey, ae, nil
}

// makePasswordList reads password lines from the file specified by the global --password flag
// and also by the same subcommand --password flag.
// This function ia a fork of utils.MakePasswordList to lookup cli context for subcommand.
// Function ctx.SetGlobal is not setting the global flag value that can be accessed
// by ctx.GlobalString using the current version of cli package.
func makePasswordList(ctx *cli.Context) []string {
	path := ctx.GlobalString(utils.PasswordFileFlag.Name)
	if path == "" {
		path = ctx.String(utils.PasswordFileFlag.Name)
		if path == "" {
			return nil
		}
	}
	text, err := ioutil.ReadFile(path)
	if err != nil {
		utils.Fatalf("Failed to read password file: %v", err)
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}
