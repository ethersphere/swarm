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
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console"
	crypto "github.com/ethereum/go-ethereum/crypto"
	ecies "github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/crypto/randentropy"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

// var defaultNodeConfig = node.DefaultConfig

// This init function sets defaults so cmd/swarm can run alongside geth.
func init() {
	defaultNodeConfig.Name = clientIdentifier
	defaultNodeConfig.Version = params.VersionWithCommit(gitCommit)
	defaultNodeConfig.P2P.ListenAddr = ":30399"
	defaultNodeConfig.IPCPath = "bzzd.ipc"
	// Set flag defaults for --help display.
	utils.ListenPortFlag.Value = 30399
}

func accessNew(ctx *cli.Context) {
	var (
		ae         *api.AccessEntry
		sessionKey []byte
		err        error
		salt       = randentropy.GetEntropyCSPRNG(32)
		args       = ctx.Args()
		dryRun     = ctx.Bool(SwarmDryRunFlag.Name)
		pk         = ctx.Bool(SwarmAccessPKFlag.Name)
	)

	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
		return
	}

	ref := args[0]

	if pk {
		sessionKey, ae, err = doPKNew(ctx, salt)
		if err != nil {
			utils.Fatalf("error getting session key: %v", err)
		}
	} else {
		sessionKey, ae, err = doPasswordNew(ctx, salt)
		if err != nil {
			utils.Fatalf("error getting session key: %v", err)
		}
	}
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

// readPassword reads a single line from stdin, trimming it from the trailing new
// line and returns it. The input will not be echoed.
func readPassword() string {
	test, err := console.Stdin.PromptPassword("Enter password: ")
	if err != nil {
		log.Crit("Failed to read password", "err", err)
	}

	return string(test)
}

func getSessionKeyPK(publisherPrivKey *ecdsa.PrivateKey, granteePubKey *ecdsa.PublicKey, salt []byte) ([]byte, error) {
	granteePubEcies := ecies.ImportECDSAPublic(granteePubKey)
	privateKey := ecies.ImportECDSA(publisherPrivKey)

	bytes, err := privateKey.GenerateShared(granteePubEcies, 16, 16)
	if err != nil {
		return nil, err
	}
	bytes = append(salt, bytes...)
	sessionKey := crypto.Keccak256(bytes)
	return sessionKey, nil
}

func doPKNew(ctx *cli.Context, salt []byte) (sessionKey []byte, ae *api.AccessEntry, err error) {
	bzzconfig, err := buildConfig(ctx)
	if err != nil {
		utils.Fatalf("unable to configure swarm: %v", err)
	}
	cfg := defaultNodeConfig

	//geth only supports --datadir via command line
	//in order to be consistent within swarm, if we pass --datadir via environment variable
	//or via config file, we get the same directory for geth and swarm
	if _, err := os.Stat(bzzconfig.Path); err == nil {
		cfg.DataDir = bzzconfig.Path
	}
	//setup the ethereum node
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

	sessionKey, err = getSessionKeyPK(privateKey, granteePub, salt)
	if err != nil {
		log.Error("error getting session key", "err", err)
		return nil, nil, err
	}

	ae, err = api.NewAccessEntryPK(hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey)), salt)
	if err != nil {
		log.Error("error generating access entry", "err", err)
		return nil, nil, err
	}

	return sessionKey, ae, nil
}

func doPasswordNew(ctx *cli.Context, salt []byte) (sessionKey []byte, ae *api.AccessEntry, err error) {
	pass := ctx.String(SwarmAccessPasswordFlag.Name)
	if pass == "" {
		pass = readPassword()
	}
	ae, err = api.NewAccessEntryPassword(salt, api.DefaultKdfParams)
	if err != nil {
		return nil, nil, err
	}

	sessionKey, err = api.NewSessionKeyPassword(pass, ae)
	if err != nil {
		return nil, nil, err
	}
	return sessionKey, ae, nil
}
