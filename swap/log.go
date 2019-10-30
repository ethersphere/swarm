package swap

import (
	log "github.com/ethereum/go-ethereum/log"
)

const (
	// DefaultAction is the default action filter for swap logs
	DefaultAction string = "undefined"
)

// Logger wraps the ethereum logger with specific information for swap logging
// this struct contains an action string that is used for grouping similar logs together
// each log contains a context which will be printed on each message
type Logger struct {
	action string
	logger log.Logger
}

func wrapCtx(sl Logger, ctx ...interface{}) []interface{} {
	// check for already-existing swap action in context
	for _, elem := range ctx {
		if elem == "swap_action" {
			return ctx
		}
	}
	// append otherwise
	return append([]interface{}{"swap_action", sl.action}, ctx...)
}

// Warn is a convenient alias for log.Warn with a defined action context
func (sl Logger) Warn(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Warn(msg, ctx...)
}

// Error is a convenient alias for log.Error with a defined action context
func (sl Logger) Error(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Error(msg, ctx...)
}

//Crit is a convenient alias for log.Crit with a defined action context
func (sl Logger) Crit(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Crit(msg, ctx...)
}

//Info is a convenient alias for log.Info with a defined action context
func (sl Logger) Info(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Info(msg, ctx...)
}

//Debug is a convenient alias for log.Debug with a defined action context
func (sl Logger) Debug(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Debug(msg, ctx...)
}

// Trace is a convenient alias for log.Trace with a defined action context
func (sl Logger) Trace(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Trace(msg, ctx...)
}

// SetLogAction sets the current log action prefix
func (sl *Logger) SetLogAction(action string) {
	sl.action = action
}

// newLogger return a new SwapLogger Instance with ctx loaded for swap
func newLogger(logPath string, ctx []interface{}) (swapLogger Logger) {
	swapLogger = Logger{
		action: DefaultAction,
	}
	swapLogger.logger = log.New(ctx...)
	setLoggerHandler(logPath, swapLogger.logger)
	return swapLogger
}
