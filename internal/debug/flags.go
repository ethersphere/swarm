// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package debug

import (
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/metrics/exp"
	"github.com/fjl/memsize/memsizeui"
	colorable "github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var Memsize memsizeui.Handler

var (
	ostream log.Handler
	glogger *log.GlogHandler
)

func init() {
	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	ostream = log.StreamHandler(output, log.TerminalFormat(usecolor))
	glogger = log.NewGlogHandler(ostream)
}

// rotatingFileHandler returns a RotatingFileHandler this will split the logs into multiple files.
// the files are split based on the limit parameter expressed in bytes
func rotatingFileHandler(logdir string) (log.Handler, error) {
	return log.RotatingFileHandler(
		logdir,
		262144,
		log.JSONFormatOrderedEx(false, true),
	)
}

type Options struct {
	Debug            bool
	Verbosity        int
	Vmodule          string
	BacktraceAt      string
	LogDirectory     string
	MemProfileRate   int
	BlockProfileRate int
	TraceFile        string
	CPUProfileFile   string
	PprofEnabled     bool
	PprofAddr        string
	PprofPort        int
}

// Setup initializes profiling and logging based on the CLI flags.
// It should be called as early as possible in the program.
func Setup(o Options) error {
	// logging
	log.PrintOrigins(o.Debug)
	if o.LogDirectory != "" {
		rfh, err := rotatingFileHandler(o.LogDirectory)
		if err != nil {
			return err
		}
		glogger.SetHandler(log.MultiHandler(ostream, rfh))
	}
	glogger.Verbosity(log.Lvl(o.Verbosity))
	glogger.Vmodule(o.Vmodule)
	glogger.BacktraceAt(o.BacktraceAt)
	log.Root().SetHandler(glogger)

	// profiling, tracing
	runtime.MemProfileRate = o.MemProfileRate
	Handler.SetBlockProfileRate(o.BlockProfileRate)
	if o.TraceFile != "" {
		if err := Handler.StartGoTrace(o.TraceFile); err != nil {
			return err
		}
	}
	if o.CPUProfileFile != "" {
		if err := Handler.StartCPUProfile(o.CPUProfileFile); err != nil {
			return err
		}
	}

	// pprof server
	if o.PprofEnabled {
		address := fmt.Sprintf("%s:%d", o.PprofAddr, o.PprofPort)
		StartPProf(address)
	}
	return nil
}

func StartPProf(address string) {
	// Hook go-metrics into expvar on any /debug/metrics request, load all vars
	// from the registry into expvar, and execute regular expvar handler.
	exp.Exp(metrics.DefaultRegistry)
	http.Handle("/memsize/", http.StripPrefix("/memsize", &Memsize))
	log.Info("Starting pprof server", "addr", fmt.Sprintf("http://%s/debug/pprof", address))
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error("Failure in running pprof server", "err", err)
		}
	}()
}

// Exit stops all running profiles, flushing their output to the
// respective file.
func Exit() {
	Handler.StopCPUProfile()
	Handler.StopGoTrace()
}
