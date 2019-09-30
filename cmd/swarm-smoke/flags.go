// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	swarmmetrics "github.com/ethersphere/swarm/metrics"
	"github.com/ethersphere/swarm/tracing"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	flags         []cli.Flag
	allhosts      string
	hosts         []string
	filesize      int
	syncDelay     bool
	pushsyncDelay bool
	syncMode      string
	inputSeed     int
	httpPort      int
	wsPort        int
	verbosity     int
	timeout       int
	single        bool
	onlyUpload    bool
	debug         bool
	bail          bool
)

func init() {
	flags = []cli.Flag{
		cli.StringFlag{
			Name:        "hosts",
			Value:       "",
			Usage:       "comma-separated list of swarm hosts",
			Destination: &allhosts,
		},
		cli.IntFlag{
			Name:        "http-port",
			Value:       80,
			Usage:       "http port",
			Destination: &httpPort,
		},
		cli.IntFlag{
			Name:        "ws-port",
			Value:       8546,
			Usage:       "ws port",
			Destination: &wsPort,
		},
		cli.IntFlag{
			Name:        "seed",
			Value:       0,
			Usage:       "input seed in case we need deterministic upload",
			Destination: &inputSeed,
		},
		cli.IntFlag{
			Name:        "filesize",
			Value:       1024,
			Usage:       "file size for generated random file in KB",
			Destination: &filesize,
		},
		cli.StringFlag{
			Name:        "sync-mode",
			Value:       "pullsync",
			Usage:       "sync mode - pushsync or pullsync or both",
			Destination: &syncMode,
		},
		cli.BoolFlag{
			Name:        "pushsync-delay",
			Usage:       "wait for content to be push synced",
			Destination: &pushsyncDelay,
		},
		cli.BoolFlag{
			Name:        "sync-delay",
			Usage:       "wait for content to be synced",
			Destination: &syncDelay,
		},
		cli.IntFlag{
			Name:        "verbosity",
			Value:       1,
			Usage:       "verbosity",
			Destination: &verbosity,
		},
		cli.IntFlag{
			Name:        "timeout",
			Value:       180,
			Usage:       "timeout in seconds after which kill the process",
			Destination: &timeout,
		},
		cli.BoolFlag{
			Name:        "single",
			Usage:       "whether to fetch content from a single node or from all nodes",
			Destination: &single,
		},
		cli.BoolFlag{
			Name:        "only-upload",
			Usage:       "whether to only upload content to a single node without fetching",
			Destination: &onlyUpload,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "whether to call debug APIs as part of the smoke test",
			Destination: &debug,
		},
		cli.BoolFlag{
			Name:        "bail",
			Usage:       "whether to fail the smoke test on any intermediate errors (such as chunks not found on max prox)",
			Destination: &bail,
		},
	}

	flags = append(flags, []cli.Flag{
		utils.MetricsEnabledFlag,
		swarmmetrics.MetricsInfluxDBEndpointFlag,
		swarmmetrics.MetricsInfluxDBDatabaseFlag,
		swarmmetrics.MetricsInfluxDBUsernameFlag,
		swarmmetrics.MetricsInfluxDBPasswordFlag,
		swarmmetrics.MetricsInfluxDBTagsFlag,
	}...)

	flags = append(flags, tracing.Flags...)

}
