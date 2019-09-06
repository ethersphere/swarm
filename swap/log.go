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

package swap

import (
	l "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

const (
	// CallDepth is set to 1 in order to influence to reported line number of
	// the log message with 1 skipped stack frame of calling l.Output()
	CallDepth = 1
	prefix    = "swap_"
)

// Warn is a convenient alias for log.Warn with stats
func Warn(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"warn", nil).Inc(1)
	l.Output(msg, l.LvlWarn, CallDepth, ctx...)
}

// Error is a convenient alias for log.Error with stats
func Error(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"error", nil).Inc(1)
	l.Output(msg, l.LvlError, CallDepth, ctx...)
}

// Crit is a convenient alias for log.Crit with stats
func Crit(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"crit", nil).Inc(1)
	l.Output(msg, l.LvlCrit, CallDepth, ctx...)
}

// Info is a convenient alias for log.Info with stats
func Info(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"info", nil).Inc(1)
	l.Output(msg, l.LvlInfo, CallDepth, ctx...)
}

// Debug is a convenient alias for log.Debug with stats
func Debug(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"debug", nil).Inc(1)
	l.Output(msg, l.LvlDebug, CallDepth, ctx...)
}

// Trace is a convenient alias for log.Trace with stats
func Trace(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter(prefix+"trace", nil).Inc(1)
	l.Output(msg, l.LvlTrace, CallDepth, ctx...)
}
