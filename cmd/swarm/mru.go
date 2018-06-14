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

// swarm resource [--rawmru] create <name> <frequency> <0x Hexdata>
// swarm resource update <Manifest Address or ENS domain> <0x Hexdata>
// swarm resource info <Manifest Address or ENS domain>

func resource(ctx *cli.Context) {

	args := ctx.Args()
	var (
		bzzapi      = strings.TrimRight(ctx.GlobalString(SwarmApiFlag.Name), "/")
		client      = swarm.NewClient(bzzapi)
		rawResource = ctx.Bool(SwarmResourceRawFlag.Name)
	)

	if len(args) < 1 {
		utils.Fatalf("Need create, update or info as first argument")
		return
	}

	switch args[0] {
	case "create":
		if len(args) < 4 {
			utils.Fatalf("Incorrect number of arguments. Syntax: swarm resource [--rawmru] create <name> <frequency> <0x Hexdata>")
			return
		}
		signer := mru.NewGenericSigner(getClientAccount(ctx))

		name := args[1]
		frequency, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			utils.Fatalf("Frequency formatting error: %s", err.Error())
			return
		}

		data, err := hexutil.Decode(args[3])
		if err != nil {
			utils.Fatalf("Error parsing data: %s", err.Error())
			return
		}
		manifestAddress, err := client.CreateResource(name, frequency, 0, data, !rawResource, signer)
		if err != nil {
			utils.Fatalf("Error creating resource: %s", err.Error())
			return
		}
		fmt.Println(manifestAddress) // output address to the user in a single line (useful for other commands to pick up)
	case "update":
		if len(args) < 3 {
			utils.Fatalf("Incorrect number of arguments. Syntax:swarm resource update <Manifest Address or ENS domain> <0x Hexdata>")
			return
		}
		signer := mru.NewGenericSigner(getClientAccount(ctx))
		manifestAddressOrDomain := args[1]
		data, err := hexutil.Decode(args[2])
		if err != nil {
			utils.Fatalf("Error parsing data: %s", err.Error())
			return
		}
		err = client.UpdateResource(manifestAddressOrDomain, data, signer)
		if err != nil {
			utils.Fatalf("Error updating resource: %s", err.Error())
			return
		}
	case "info":
		if len(args) < 2 {
			utils.Fatalf("Incorrect number of arguments. Syntax: swarm resource info <Manifest Address or ENS domain>")
			return
		}
		manifestAddressOrDomain := args[1]
		metadata, err := client.GetResourceMetadata(manifestAddressOrDomain)
		if err != nil {
			utils.Fatalf("Error retrieving resource metadata: %s", err.Error())
			return
		}
		fmt.Printf("Name: %s\n", metadata.Name())
		fmt.Printf("Update frequency: %d\n", metadata.Frequency())
		fmt.Printf("Start time: %d\n", metadata.StartTime())
		fmt.Printf("Owner: %s\n", metadata.OwnerAddr().Hex())
		fmt.Printf("Raw: %t\n", !metadata.Multihash())
		fmt.Printf("Next period: %d\n", metadata.Period())
		fmt.Printf("Next version: %d\n", metadata.Version())

	default:
		utils.Fatalf("invalid resource operation")
		return
	}
}
