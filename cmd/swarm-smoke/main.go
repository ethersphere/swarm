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
	"fmt"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/cmd/utils"
	gethmetrics "github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/metrics/influxdb"
	cliflags "github.com/ethersphere/swarm/internal/flags"
	"github.com/ethersphere/swarm/tracing"

	"github.com/ethereum/go-ethereum/log"

	cli "gopkg.in/urfave/cli.v1"
)

var (
	gitCommit string // Git SHA1 commit hash of the release (set via linker flags)
)

func main() {
	app := cli.NewApp()
	app.Name = "smoke-test"
	app.Usage = ""

	app.Flags = flags
	app.Commands = []cli.Command{
		{
			Name:    "upload_and_sync",
			Aliases: []string{"c"},
			Usage:   "upload and sync",
			Action:  wrapCliCommand("upload-and-sync", uploadAndSyncCmd),
		},
		{
			Name:    "feed_sync",
			Aliases: []string{"f"},
			Usage:   "feed update generate, upload and sync",
			Action:  wrapCliCommand("feed-and-sync", feedUploadAndSyncCmd),
		},
		{
			Name:    "upload_speed",
			Aliases: []string{"u"},
			Usage:   "measure upload speed",
			Action:  wrapCliCommand("upload-speed", uploadSpeedCmd),
		},
		{
			Name:    "sliding_window",
			Aliases: []string{"s"},
			Usage:   "measure network aggregate capacity",
			Action:  wrapCliCommand("sliding-window", slidingWindowCmd),
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Before = func(ctx *cli.Context) error {
		tracing.Setup(ctx)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return emitMetrics(ctx)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())

		os.Exit(1)
	}
}

func emitMetrics(ctx *cli.Context) error {
	if gethmetrics.Enabled {
		var (
			endpoint = ctx.GlobalString(cliflags.MetricsInfluxDBEndpointFlag.Name)
			database = ctx.GlobalString(cliflags.MetricsInfluxDBDatabaseFlag.Name)
			username = ctx.GlobalString(cliflags.MetricsInfluxDBUsernameFlag.Name)
			password = ctx.GlobalString(cliflags.MetricsInfluxDBPasswordFlag.Name)
			tags     = ctx.GlobalString(cliflags.MetricsInfluxDBTagsFlag.Name)
		)

		tagsMap := utils.SplitTagsFlag(tags)
		tagsMap["version"] = gitCommit
		tagsMap["filesize"] = fmt.Sprintf("%v", filesize)

		return influxdb.InfluxDBWithTagsOnce(gethmetrics.DefaultRegistry, endpoint, database, username, password, "swarm-smoke.", tagsMap)
	}

	return nil
}
