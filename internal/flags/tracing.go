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

package flags

import "gopkg.in/urfave/cli.v1"

// Tracing holds all command-line flags required for tracing collection.
var Tracing = []cli.Flag{
	TracingEnabledFlag,
	TracingEndpointFlag,
	TracingSvcFlag,
}

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
