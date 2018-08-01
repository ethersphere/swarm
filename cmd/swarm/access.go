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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto/sha3"

	"github.com/ethereum/go-ethereum/cmd/utils"
	crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
	"gopkg.in/urfave/cli.v1"
)

var salt = make([]byte, 32)

// This init function sets defaults so cmd/swarm can run alongside geth.
func init() {
	defaultNodeConfig.Name = clientIdentifier
	defaultNodeConfig.Version = params.VersionWithCommit(gitCommit)
	defaultNodeConfig.P2P.ListenAddr = ":30399"
	defaultNodeConfig.IPCPath = "bzzd.ipc"
	// Set flag defaults for --help display.
	utils.ListenPortFlag.Value = 30399

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
}

func accessNewPass(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
	}

	var (
		ae        *api.AccessEntry
		accessKey []byte
		err       error
		ref       = args[0]
	)
	accessKey, ae, err = doPasswordNew(ctx, salt)
	if err != nil {
		utils.Fatalf("error getting session key: %v", err)
	}
	generateAccessControlManifest(ctx, ref, accessKey, ae)
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

func accessNewACT(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		utils.Fatalf("Expected 1 argument - the ref", "")
	}

	var (
		ae        *api.AccessEntry
		accessKey []byte
		err       error
		ref       = args[0]
	)

	grantees := []string{}

	accessKey, ae, err = doACTNew(ctx, salt, grantees)
	if err != nil {
		utils.Fatalf("error generating ACT manifest: %v", err)
	}

	if err != nil {
		utils.Fatalf("error getting session key: %v", err)
	}
	generateAccessControlManifest(ctx, ref, accessKey, ae)
}

func generateAccessControlManifest(ctx *cli.Context, ref string, accessKey []byte, ae *api.AccessEntry) {
	dryRun := ctx.Bool(SwarmDryRunFlag.Name)
	refBytes, err := hex.DecodeString(ref)
	if err != nil {
		utils.Fatalf("Error: %v", err)
	}
	// encrypt ref with accessKey
	enc := api.NewRefEncryption(len(refBytes))
	encrypted, err := enc.Encrypt(refBytes, accessKey)
	if err != nil {
		utils.Fatalf("Error: %v", err)
	}

	m := api.Manifest{
		Entries: []api.ManifestEntry{
			api.ManifestEntry{
				Hash:        hex.EncodeToString(encrypted),
				ContentType: api.ManifestType,
				ModTime:     time.Now(),
				Access:      ae,
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

func doACTNew(ctx *cli.Context, salt []byte, granteesPublicKeys []string) (accessKey []byte, ae *api.AccessEntry, err error) {
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

	accessKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}

	lookupPathEncryptedAccessKeyMap := make(map[string]string)
	for _, v := range granteesPublicKeys {
		if v == "" {
			return nil, nil, errors.New("need a grantee Public Key")
		}
		b, err := hex.DecodeString(v)
		if err != nil {
			log.Error("error decoding grantee public key", "err", err)
			return nil, nil, err
		}

		granteePub, err := crypto.UnmarshalPubkey(b)
		if err != nil {
			log.Error("error unmarshaling grantee public key", "err", err)
			return nil, nil, err
		}
		sessionKey, err := api.NewSessionKeyPK(privateKey, granteePub, salt)

		hasher := sha3.NewKeccak256()
		hasher.Write(append(sessionKey, 0))
		lookupKey := hasher.Sum(nil)

		hasher.Reset()
		hasher.Write(append(sessionKey, 1))

		accessKeyEncryptionKey := hasher.Sum(nil)

		enc := api.NewRefEncryption(len(accessKey))
		encryptedAccessKey, err := enc.Encrypt(accessKey, accessKeyEncryptionKey)

		lookupPathEncryptedAccessKeyMap[hex.EncodeToString(lookupKey)] = hex.EncodeToString(encryptedAccessKey)

	}
	m := api.Manifest{
		Entries: []api.ManifestEntry{},
	}

	for k, v := range lookupPathEncryptedAccessKeyMap {
		m.Entries = append(m.Entries, api.ManifestEntry{
			Path:        k,
			Hash:        v,
			ContentType: api.ManifestType,
		})
	}

	bzzapi := strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
	client := client.NewClient(bzzapi)

	key, err := client.UploadManifest(&m, false)
	if err != nil {
		utils.Fatalf("Error uploading manifest: %v", err)
	}

	uri, err := api.Parse("bzz://" + key)
	if err != nil {
		log.Error("error creating swarm URI from key", "err", err)
		return nil, nil, err
	}

	ae, err = api.NewAccessEntryACT(hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey)), salt, uri.Addr)
	if err != nil {
		log.Error("error generating access entry", "err", err)
		return nil, nil, err
	}

	return accessKey, ae, nil

	//create session keys for each public/private keypair
	//construct manifest
	/*
				create slice of session struct
				create array of lookup keys by hashing 0 to all of these session keys
				sha3(session +"0") => lookup key = sha3(append(sessionkey,0)) => check that output is different, maybe create unit test with sanity checks for known hashes
				access key encryption key = sha3(append(sessionKey,1))
				create access key = random32Bytes
				create encrypted accesskeys: encrypt access key using access key encryption keys:
					enc := api.NewRefEncryption(len(encrypted accesskey))
					encryptedAccessKey, err := enc.Encrypt(random32Bytes,accesskey encryption key)
				construct manifest where the path i is the lookup key and the manifest entry
				sitting at that path contains the

				=========
		root access manifest with meta
		see that its act and match on it

		take private key + public key from the metadata (fallback to password when this doesnt work)
		create shared secret
		create sessionkey+lookup key by hashing 1+0 with it then concat the act url + lookup key
		this with bzz/bzz-raw
		create access key decryption key by hashing 1 to session key
		decrypt what's in the manifest with this
		then try to decrypt the reference whicnh was in the original manigfest

	*/

	// ae, err = api.NewAccessEntryACT(hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey)), salt)
	// if err != nil {
	// 	log.Error("error generating access entry", "err", err)
	// 	return nil, nil, err
	// }

	// return sessionKey, ae, nil
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
