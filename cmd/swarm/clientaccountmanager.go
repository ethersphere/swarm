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

/*
//This module implements the bare minimum to allow client commands to access the keystore
*/
package main

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
)

type ClientAccountManagerProvider struct {
	manager *accounts.Manager
}

func (c *ClientAccountManagerProvider) AccountManager() *accounts.Manager {
	return c.manager
}

func getClientAccount(ctx *cli.Context) *ecdsa.PrivateKey {
	bzzconfig, err := buildConfig(ctx)
	if err != nil {
		utils.Fatalf("Cannot read configuration: %s", err.Error())
		return nil
	}
	nodeconfig := defaultNodeConfig
	if _, err := os.Stat(bzzconfig.Path); err == nil {
		nodeconfig.DataDir = bzzconfig.Path
	}

	accountManager, _, err := makeAccountManager(&nodeconfig)
	if err != nil {
		utils.Fatalf("Cannot create account manager: %s", err.Error())
		return nil
	}

	return getAccount(bzzconfig.BzzAccount, ctx, &ClientAccountManagerProvider{
		manager: accountManager,
	})

}

func makeAccountManager(conf *node.Config) (*accounts.Manager, string, error) {
	scryptN, scryptP, keydir, err := conf.AccountConfig()
	var ephemeral string
	if keydir == "" {
		// There is no datadir.
		keydir, err = ioutil.TempDir("", "go-ethereum-keystore")
		ephemeral = keydir
	}

	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}
	if !conf.NoUSB {
		// Start a USB hub for Ledger hardware wallets
		if ledgerhub, err := usbwallet.NewLedgerHub(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start Ledger hub, disabling: %v", err))
		} else {
			backends = append(backends, ledgerhub)
		}
		// Start a USB hub for Trezor hardware wallets
		if trezorhub, err := usbwallet.NewTrezorHub(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start Trezor hub, disabling: %v", err))
		} else {
			backends = append(backends, trezorhub)
		}
	}
	return accounts.NewManager(backends...), ephemeral, nil
}
