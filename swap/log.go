package swap

import (
	"encoding/hex"

	log "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

const (
	// UndefinedAction is the default actions filter for swap logs
	UndefinedAction        string = "undefined"
	InitAction             string = "init"
	StopAction             string = "stop"
	UpdateBalanceAction    string = "update_balance"
	SendChequeAction       string = "send_cheque"
	HandleChequeAction     string = "handle_cheque"
	CashChequeAction       string = "cash_cheque"
	DeployChequebookAction string = "deploy_chequebook_contract"
)

//var swapLog log.Logger // logger for Swap related messages and audit trail
const swapLogLevel = 3 // swapLogLevel indicates filter level of log messages

// Logger wraps the ethereum logger with specific information for swap logging
// this struct contains an action string that is used for grouping similar logs together
// each log contains a context which will be printed on each message
type Logger struct {
	action string
	logger log.Logger
}

func wrapCtx(sl Logger, action string, ctx ...interface{}) []interface{} {
	return append([]interface{}{"swap_action", action}, ctx...)
}

// Warn is a convenient alias for log.Warn with a defined action context
func (sl Logger) Warn(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Warn(msg, ctx...)
}

// Error is a convenient alias for log.Error with a defined action context
func (sl Logger) Error(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Error(msg, ctx...)
}

//Crit is a convenient alias for log.Crit with a defined action context
func (sl Logger) Crit(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Crit(msg, ctx...)
}

//Info is a convenient alias for log.Info with a defined action context
func (sl Logger) Info(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Info(msg, ctx...)
}

//Debug is a convenient alias for log.Debug with a defined action context
func (sl Logger) Debug(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Debug(msg, ctx...)
}

// Trace is a convenient alias for log.Trace with a defined action context
func (sl Logger) Trace(action string, msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, action, ctx...)
	sl.logger.Trace(msg, ctx...)
}

// newLogger return a new SwapLogger Instance with ctx loaded for swap
func newLogger(logPath string, ctx []interface{}) (swapLogger Logger) {
	swapLogger = Logger{
		action: UndefinedAction,
	} //TODO:REMOVE ACTION FROM LOGGER
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
