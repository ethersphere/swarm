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
package outbox

import (
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/message"
	"github.com/tilinna/clock"
)

// Config contains the Outbox configuration.
type Config struct {
	NumberSlots  int             // number of slots for messages in Outbox.
	NumWorkers   int             // number of parallel goroutines forwarding messages.
	Forward      forwardFunction // function that executes the actual forwarding.
	MaxRetryTime *time.Duration  // max time a message will be retried in the outbox.
	Clock        clock.Clock     // clock dependency to calculate elapsed time.
}

// Outbox will be in charge of forwarding messages. These will be enqueued and retry until successfully forwarded.
type Outbox struct {
	forwardFunc  forwardFunction
	queue        []*outboxMsg
	slots        chan int
	process      chan int
	numWorkers   int
	stopC        chan struct{}
	maxRetryTime time.Duration
	clock        clock.Clock
}

type forwardFunction func(msg *message.Message) error

// ErrOutboxFull is returned when a caller tries to enqueue a message and all slots are busy.
var ErrOutboxFull = errors.New("outbox full")

const defaultOutboxWorkers = 100
const defaultMaxRetryTime = 10 * time.Minute

// NewOutbox creates a new Outbox. Config must be provided. IF NumWorkers is not providers, default will be used.
func NewOutbox(config *Config) *Outbox {
	if config.NumWorkers == 0 {
		config.NumWorkers = defaultOutboxWorkers
	}
	outbox := &Outbox{
		forwardFunc:  config.Forward,
		queue:        make([]*outboxMsg, config.NumberSlots),
		slots:        make(chan int, config.NumberSlots),
		process:      make(chan int),
		numWorkers:   config.NumWorkers,
		stopC:        make(chan struct{}),
		maxRetryTime: defaultMaxRetryTime,
		clock:        clock.Realtime(),
	}
	if config.MaxRetryTime != nil {
		outbox.maxRetryTime = *config.MaxRetryTime
	}
	if config.Clock != nil {
		outbox.clock = config.Clock
	}
	// fill up outbox slots
	for i := 0; i < cap(outbox.slots); i++ {
		outbox.slots <- i
	}
	return outbox
}

// Start launches the processing of messages in the outbox.
func (o *Outbox) Start() {
	log.Info("Starting outbox")
	go o.processOutbox()
}

// Stop stops the processing of messages in the outbox.
func (o *Outbox) Stop() {
	log.Info("Stopping outbox")
	close(o.stopC)
}

// Enqueue a new element in the outbox if there is any slot available.
// Then send it to process. This method is blocking if there is no workers available.
func (o *Outbox) Enqueue(outboxMsg *outboxMsg) error {
	// first we try to obtain a slot in the outbox.
	select {
	case slot := <-o.slots:
		o.queue[slot] = outboxMsg
		metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		// we send this message slot to process.
		select {
		case <-o.stopC:
		case o.process <- slot:
		}
		return nil
	default:
		metrics.GetOrRegisterCounter("pss.enqueue.outbox.full", nil).Inc(1)
		return ErrOutboxFull
	}
}

// SetForward set the forward function that will be executed on each message.
func (o *Outbox) SetForward(forwardFunc forwardFunction) {
	o.forwardFunc = forwardFunc
}

// NewOutboxMessage creates a new outbox message wrapping a pss message and set the startedTime using the clock
func (o *Outbox) NewOutboxMessage(msg *message.Message) *outboxMsg {
	return &outboxMsg{
		msg:       msg,
		startedAt: o.clock.Now(),
	}
}

// ProcessOutbox starts a routine that tries to forward messages present in the outbox queue.
func (o *Outbox) processOutbox() {
	workerLimitC := make(chan struct{}, o.numWorkers)
	for {
		select {
		case <-o.stopC:
			return
		case slot := <-o.process:
			log.Debug("Processing, taking worker", "workerLimit size", len(workerLimitC), "numWorkers", o.numWorkers)
			workerLimitC <- struct{}{}
			go func(slot int) {
				//Free worker space
				defer func() { <-workerLimitC }()
				msg := o.queue[slot]
				metrics.GetOrRegisterResettingTimer("pss.handle.outbox", nil).UpdateSince(msg.startedAt)
				if err := o.forwardFunc(msg.msg); err != nil {
					metrics.GetOrRegisterCounter("pss.forward.err", nil).Inc(1)
					log.Debug(err.Error())
					limit := msg.startedAt.Add(o.maxRetryTime)
					now := o.clock.Now()
					if now.After(limit) {
						metrics.GetOrRegisterCounter("pss.forward.expired", nil).Inc(1)
						log.Warn("Message expired, won't be requeued", "limit", limit, "now", now)
						o.free(slot)
						metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
						return
					}
					// requeue the message for processing
					o.requeue(slot)
					log.Debug("Message requeued", "slot", slot)
					return
				}
				//message processed, free the outbox slot
				o.free(slot)
				metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
			}(slot)
		}
	}
}

func (o *Outbox) free(slot int) {
	select {
	case <-o.stopC:
	case o.slots <- slot:
	}

}

func (o *Outbox) requeue(slot int) {
	select {
	case <-o.stopC:
	case o.process <- slot:
	}
}
func (o *Outbox) len() int {
	return cap(o.slots) - len(o.slots)
}

func (o *Outbox) CurrentSize() int {
	return cap(o.slots) - len(o.slots)
}
