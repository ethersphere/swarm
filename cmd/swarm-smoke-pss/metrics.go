// Copyright 2019 The The Swarm Authors
// This file is part of The Swarm.
//
// The Swarm is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with The Swarm. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	MetricsEnableInfluxDBExportFlag = cli.BoolFlag{
		Name:  "metrics.influxdb.export",
		Usage: "Enable metrics export/push to an external InfluxDB database",
	}
	MetricsInfluxDBEndpointFlag = cli.StringFlag{
		Name:  "metrics.influxdb.endpoint",
		Usage: "Metrics InfluxDB endpoint",
		Value: "http://127.0.0.1:8086",
	}
	MetricsInfluxDBDatabaseFlag = cli.StringFlag{
		Name:  "metrics.influxdb.database",
		Usage: "Metrics InfluxDB database",
		Value: "metrics",
	}
	MetricsInfluxDBUsernameFlag = cli.StringFlag{
		Name:  "metrics.influxdb.username",
		Usage: "Metrics InfluxDB username",
		Value: "",
	}
	MetricsInfluxDBPasswordFlag = cli.StringFlag{
		Name:  "metrics.influxdb.password",
		Usage: "Metrics InfluxDB password",
		Value: "",
	}
	// Tags are part of every measurement sent to InfluxDB. Queries on tags are faster in InfluxDB.
	// For example `host` tag could be used so that we can group all nodes and average a measurement
	// across all of them, but also so that we can select a specific node and inspect its measurements.
	// https://docs.influxdata.com/influxdb/v1.4/concepts/key_concepts/#tag-key
	MetricsInfluxDBTagsFlag = cli.StringFlag{
		Name:  "metrics.influxdb.tags",
		Usage: "Comma-separated InfluxDB tags (key/values) attached to all measurements",
		Value: "host=localhost",
	}
)

// MetricsFlags holds all command-line flags required for metrics collection.
var MetricsFlags = []cli.Flag{
	utils.MetricsEnabledFlag,
	MetricsEnableInfluxDBExportFlag,
	MetricsInfluxDBEndpointFlag,
	MetricsInfluxDBDatabaseFlag,
	MetricsInfluxDBUsernameFlag,
	MetricsInfluxDBPasswordFlag,
	MetricsInfluxDBTagsFlag,
}
