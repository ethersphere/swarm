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

package testutil

import (
	"flag"

	"github.com/ethereum/go-ethereum/log"
	"github.com/mattn/go-colorable"
)

// Common flags used in Swarm tests.
var (
	Loglevel    = flag.Int("loglevel", 2, "verbosity of logs")
	Longrunning = flag.Bool("longrunning", false, "do run long-running tests")

	rawlog = flag.Bool("rawlog", false, "remove terminal formatting from logs")
)

// Init ensures that testing.Init is called before flag.Parse and sets common
// logging options.
func Init() {
	testInit()

	flag.Parse()

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*Loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(!*rawlog))))
}

// This function is set to testing.Init for go 1.13.
var testInit = func() {}
