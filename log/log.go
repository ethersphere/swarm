package log

import (
	l "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

const (
	// CallDepth is set to 1 in order to influence to reported line number of
	// the log message with 1 skipped stack frame of calling l.Output()
	CallDepth = 1
)

var (
	logBaseAddr = false
)

// Export go-ethereum/log interface so that swarm/log can be used with it interchangeably
type Logger = l.Logger

// NewBaseAddressLogger creates a new logger with a `base` prefix
func NewBaseAddressLogger(baseAddr string, ctx ...interface{}) l.Logger {
	if logBaseAddr {
		return l.New(append([]interface{}{"base", baseAddr}, ctx...)...)
	}

	return l.New(ctx...)
}

// New creates new swarm logger
func New(ctx ...interface{}) Logger {
	return l.New(ctx)
}

// EnableBaseAddress enables the logging of the base address
// it is used for tests
func EnableBaseAddress() {
	logBaseAddr = true
}

// Warn is a convenient alias for log.Warn with stats
func Warn(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("warn", nil).Inc(1)
	l.Output(msg, l.LvlWarn, CallDepth, ctx...)
}

// Error is a convenient alias for log.Error with stats
func Error(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("error", nil).Inc(1)
	l.Output(msg, l.LvlError, CallDepth, ctx...)
}

// Crit is a convenient alias for log.Crit with stats
func Crit(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("crit", nil).Inc(1)
	l.Output(msg, l.LvlCrit, CallDepth, ctx...)
}

// Info is a convenient alias for log.Info with stats
func Info(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("info", nil).Inc(1)
	l.Output(msg, l.LvlInfo, CallDepth, ctx...)
}

// Debug is a convenient alias for log.Debug with stats
func Debug(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("debug", nil).Inc(1)
	l.Output(msg, l.LvlDebug, CallDepth, ctx...)
}

// Trace is a convenient alias for log.Trace with stats
func Trace(msg string, ctx ...interface{}) {
	metrics.GetOrRegisterCounter("trace", nil).Inc(1)
	l.Output(msg, l.LvlTrace, CallDepth, ctx...)
}

// GetHandler return the Handler assigned to root
func GetHandler() l.Handler {
	return l.Root().GetHandler()
}
