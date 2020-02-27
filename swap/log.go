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
	log "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/network"
)

const (
	// UndefinedAction is the default actions filter for swap logs
	UndefinedAction string = "undefined"
	// InitAction used when starting swap
	InitAction string = "init"
	// StopAction used when stopping swap
	StopAction string = "stop"
	// UpdateBalanceAction used when updating balances
	UpdateBalanceAction string = "update_balance"
	// SendChequeAction used for cheque actions
	SendChequeAction string = "send_cheque"
	// HandleChequeAction used for cheque actions
	HandleChequeAction string = "handle_cheque"
	// CashChequeAction used for cheque actions
	CashChequeAction string = "cash_cheque"
	// DeployChequebookAction used when deploying chequebooks
	DeployChequebookAction string = "deploy_chequebook_contract"
)

const swapLogLevel = 3       // swapLogLevel indicates filter level of log messages
const fileSizeLimit = 262144 // max bytes limit for splitting file in parts
const emptyLogPath = ""      // Used when no logPath is specified for a logger

// Logger wraps the ethereum logger with specific information for swap logging
// each log contains a context which will be printed on each message
type Logger struct {
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
	swapLogger = Logger{}
	swapLogger.logger = log.New(ctx...)
	setLoggerHandler(logPath, swapLogger.logger)
	return swapLogger
}

// setLoggerHandler will set the logger handle to write logs to the specified path
// or use the default swarm logger in case this isn't specified or an error occurs
func setLoggerHandler(logpath string, logger log.Logger) {
	lh := log.Root().GetHandler()

	if logpath == emptyLogPath {
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
		fileSizeLimit,
		log.JSONFormatOrderedEx(false, true),
	)
}

// newSwapLogger returns a new logger for standard swap logs
func newSwapLogger(logPath string, baseAddress *network.BzzAddr) Logger {
	ctx := []interface{}{"base", baseAddress.ShortString()}
	return newLogger(logPath, ctx)
}

// newPeerLogger returns a new logger for swap logs with peer info
func newPeerLogger(s *Swap, peerID enode.ID) Logger {
	ctx := []interface{}{"base", s.params.BaseAddrs.ShortString(), "peer", peerID.String()[:16]}
	return newLogger(s.params.LogPath, ctx)
}
