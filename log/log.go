package log

import (
	l "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
)

// Action Defines swap log actions
type Action string

const (
	// CallDepth is set to 1 in order to influence to reported line number of
	// the log message with 1 skipped stack frame of calling l.Output()
	CallDepth = 1
	//DefaultAction is the default action filter for SwapLogs
	DefaultAction Action = "*"
	//SentChequeAction is a filter for SwapLogs
	SentChequeAction Action = "SentCheque"
)

// SwapLogger wraps the ethereum logger with specific information for swap logging
type SwapLogger struct {
	action      Action
	overlayAddr string
	peerID      string
	logger      l.Logger
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

func wrapCtx(sl SwapLogger, ctx ...interface{}) []interface{} {
	for _, elem := range ctx {
		if elem == "action" && len(ctx)%2 == 0 {
			return ctx
		}
	}
	ctx = addSwapAction(sl, ctx...)
	return ctx
}

// Warn TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Warn(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Warn(msg, ctx...)
}

// Error TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Error(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Error(msg, ctx...)
}

//Crit TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Crit(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Crit(msg, ctx...)
}

//Info TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Info(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Info(msg, ctx...)
}

//Debug TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Debug(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Debug(msg, ctx...)
}

// Trace TODO REVIEW THIS COMMENT is a convenient alias for log.Warn with stats
func (sl SwapLogger) Trace(msg string, ctx ...interface{}) {
	ctx = wrapCtx(sl, ctx...)
	sl.logger.Trace(msg, ctx...)
}

// SetLogAction set the current log action prefix
func (sl *SwapLogger) SetLogAction(action Action) {
	//Adds default action *
	if action == "" {
		sl.action = DefaultAction
		return
	}
	//Todo validate it's a specific action, if not default
	sl.action = action
}

// NewSwapLogger is an alias for log.New
func NewSwapLogger(overlayAddr string) (swapLogger SwapLogger) {
	swapLogger = SwapLogger{
		action:      DefaultAction,
		overlayAddr: overlayAddr,
	}
	ctx := addSwapCtx(swapLogger)
	swapLogger.logger = l.New(ctx...)
	return swapLogger
}

// NewSwapPeerLogger is an alias for log.New
func NewSwapPeerLogger(overlayAddr string, peerID string) (swapLogger SwapLogger) {

	swapLogger = SwapLogger{
		action:      DefaultAction,
		overlayAddr: overlayAddr,
		peerID:      peerID,
	}
	ctx := addSwapCtx(swapLogger)
	swapLogger.logger = l.New(ctx...)
	return swapLogger
}

func addSwapCtx(sl SwapLogger, ctx ...interface{}) []interface{} {
	ctx = append([]interface{}{"base", sl.overlayAddr}, ctx...)
	if sl.peerID != "" {
		ctx = append(ctx, "peer", sl.peerID)
	}
	return ctx
}

func addSwapAction(sl SwapLogger, ctx ...interface{}) []interface{} {
	return append([]interface{}{"action", sl.action}, ctx...)
}

// GetLogger return the underlining logger
func (sl SwapLogger) GetLogger() (logger l.Logger) {
	return sl.logger
}

// GetHandler return the Handler assigned to root
func GetHandler() l.Handler {
	return l.Root().GetHandler()
}
