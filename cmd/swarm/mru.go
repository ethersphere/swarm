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

// Command resource allows the user to create and update signed mutable resource updates
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/cmd/utils"
	swarm "github.com/ethereum/go-ethereum/swarm/api/client"
	"github.com/ethereum/go-ethereum/swarm/storage/mru"
	"gopkg.in/urfave/cli.v1"
)

// swarm resource create <name> <frequency> [--rawmru] <0x Hexdata>
// swarm resource update <Manifest Address or ENS domain> <0x Hexdata>
// swarm resource info <Manifest Address or ENS domain>

func resourceCreate(ctx *cli.Context) {
	args := ctx.Args()
	var (
		bzzapi      = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client      = swarm.NewClient(bzzapi)
		rawResource = ctx.Bool(SwarmResourceRawFlag.Name)
	)

	if len(args) < 3 {
		fmt.Println("Incorrect number of arguments")
		cli.ShowCommandHelpAndExit(ctx, "create", 1)
		return
	}
	signer := mru.NewGenericSigner(getClientAccount(ctx))

	name := args[0]
	frequency, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		utils.Fatalf("Frequency formatting error: %s", err.Error())
		return
	}

	data, err := hexutil.Decode(args[2])
	if err != nil {
		utils.Fatalf("Error parsing data: %s", err.Error())
		return
	}

	newResourceRequest, err := mru.NewCreateRequest(&mru.ResourceMetadata{
		Name:      name,
		Frequency: frequency,
		StartTime: mru.Timestamp{Time: 0},
		OwnerAddr: signer.Address(),
	})

	if err != nil {
		utils.Fatalf("Error creating new resource request: %s", err)
	}

	newResourceRequest.SetData(data, !rawResource)
	if err = newResourceRequest.Sign(signer); err != nil {
		utils.Fatalf("Error signing resource update: %s", err.Error())
	}

	manifestAddress, err := client.CreateResource(newResourceRequest)
	if err != nil {
		utils.Fatalf("Error creating resource: %s", err.Error())
		return
	}
	fmt.Println(manifestAddress) // output address to the user in a single line (useful for other commands to pick up)

}

func resourceUpdate(ctx *cli.Context) {
	args := ctx.Args()
	var (
		bzzapi = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client = swarm.NewClient(bzzapi)
	)
	if len(args) < 2 {
		fmt.Println("Incorrect number of arguments")
		cli.ShowCommandHelpAndExit(ctx, "update", 1)
		return
	}
	signer := mru.NewGenericSigner(getClientAccount(ctx))
	manifestAddressOrDomain := args[0]
	data, err := hexutil.Decode(args[1])
	if err != nil {
		utils.Fatalf("Error parsing data: %s", err.Error())
		return
	}

	// Retrieve resource status and metadata out of the manifest
	updateRequest, err := client.GetResourceMetadata(manifestAddressOrDomain)
	if err != nil {
		utils.Fatalf("Error retrieving resource status: %s", err.Error())
	}

	// set the new data
	updateRequest.SetData(data, updateRequest.Multihash()) // set data, keep current multihash setting

	// sign update
	if err = updateRequest.Sign(signer); err != nil {
		utils.Fatalf("Error signing resource update: %s", err.Error())
	}

	// post update
	err = client.UpdateResource(updateRequest)
	if err != nil {
		utils.Fatalf("Error updating resource: %s", err.Error())
		return
	}
}

func resourceInfo(ctx *cli.Context) {
	var (
		bzzapi = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client = swarm.NewClient(bzzapi)
	)
	args := ctx.Args()
	if len(args) < 1 {
		fmt.Println("Incorrect number of arguments.")
		cli.ShowCommandHelpAndExit(ctx, "info", 1)
		return
	}
	manifestAddressOrDomain := args[0]
	metadata, err := client.GetResourceMetadata(manifestAddressOrDomain)
	if err != nil {
		utils.Fatalf("Error retrieving resource metadata: %s", err.Error())
		return
	}
	encodedMetadata, err := mru.EncodeUpdateRequest(metadata)
	if err != nil {
		utils.Fatalf("Error encoding metadata to JSON for display:%s", err)
	}
	fmt.Println(string(encodedMetadata))
}
