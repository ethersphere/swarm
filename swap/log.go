package swap

import (
	"encoding/hex"

	log "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
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

// setLoggerHandler will set the logger handle to write logs to the specified path
// or use the default swarm logger in case this isn't specified or an error occurs
func setLoggerHandler(logpath string, logger log.Logger) {
	lh := log.Root().GetHandler()

	if logpath == "" {
		logger.SetHandler(lh)
		return
	}

	rfh, err := swapRotatingFileHandler(logpath)

	if err != nil {
		log.Warn("RotatingFileHandler was not initialized", "logdir", logpath, "err", err)
		// use the default swarm logger as a fallback
		logger.SetHandler(lh)
		return
	}

	// filter messages with the correct log level for swap
	rfh = log.LvlFilterHandler(log.Lvl(swapLogLevel), rfh)

	// dispatch the logs to the default swarm log and also the filtered swap logger
	logger.SetHandler(log.MultiHandler(lh, rfh))
}

// swapRotatingFileHandler returns a RotatingFileHandler this will split the logs into multiple files.
// the files are split based on the limit parameter expressed in bytes
func swapRotatingFileHandler(logdir string) (log.Handler, error) {
	return log.RotatingFileHandler(
		logdir,
		262144,
		log.JSONFormatOrderedEx(false, true),
	)
}

// newSwapLogger returns a new logger for standard swap logs
func newSwapLogger(logPath string, overlayAddr []byte) Logger {
	ctx := []interface{}{"base", hex.EncodeToString(overlayAddr)[:16]}
	return newLogger(logPath, ctx)
}

// newPeerLogger returns a new logger for swap logs with peer info
func newPeerLogger(s *Swap, peerID enode.ID) Logger {
	ctx := []interface{}{"base", hex.EncodeToString(s.params.BaseAddrs.Over())[:16], "peer", peerID.String()[:16]}
	return newLogger(s.params.LogPath, ctx)
}
