// Copyright 2019 The Swarm Authors
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

import "gopkg.in/urfave/cli.v1"

var (
	TracingEnabledFlag = cli.BoolFlag{
		Name:  "tracing",
		Usage: "Enable tracing",
	}
	TracingEndpointFlag = cli.StringFlag{
		Name:  "tracing.endpoint",
		Usage: "Tracing endpoint",
		Value: "0.0.0.0:6831",
	}
	TracingSvcFlag = cli.StringFlag{
		Name:  "tracing.svc",
		Usage: "Tracing service name",
		Value: "swarm",
	}
)

// Flags holds all command-line flags required for tracing collection.
var TracingFlags = []cli.Flag{
	TracingEnabledFlag,
	TracingEndpointFlag,
	TracingSvcFlag,
}
