package log

import (
	l "github.com/ethereum/go-ethereum/log"
	metrics "github.com/rcrowley/go-metrics"
)

// Warn is a convenient alias for log.Warn with stats
func Warn(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("warn", nil).Inc(1)
	l.Output(msg, l.LvlWarn, 3, ctx...)
}

// Error is a convenient alias for log.Error with stats
func Error(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("error", nil).Inc(1)
	l.Output(msg, l.LvlError, 3, ctx...)
}

// Crit is a convenient alias for log.Crit with stats
func Crit(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("crit", nil).Inc(1)
	l.Output(msg, l.LvlCrit, 3, ctx...)
}

// Info is a convenient alias for log.Info with stats
func Info(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("info", nil).Inc(1)
	l.Output(msg, l.LvlInfo, 3, ctx...)
}

// Debug is a convenient alias for log.Debug with stats
func Debug(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("debug", nil).Inc(1)
	l.Output(msg, l.LvlDebug, 3, ctx...)
}

// Trace is a convenient alias for log.Trace with stats
func Trace(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("trace", nil).Inc(1)
	l.Output(msg, l.LvlTrace, 3, ctx...)
}
