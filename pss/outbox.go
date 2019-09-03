package pss

import (
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/metrics"

	"github.com/ethersphere/swarm/log"
)

type outbox struct {
	queue   []*outboxMsg
	slots   chan int
	process chan int
	quitC   chan struct{}
	forward func(msg *PssMsg) error
}

func newOutbox(capacity int, quitC chan struct{}, forward func(msg *PssMsg) error) outbox {
	outbox := outbox{
		queue:   make([]*outboxMsg, capacity),
		slots:   make(chan int, capacity),
		process: make(chan int),
		quitC:   quitC,
		forward: forward,
	}
	// fill up outbox slots
	for i := 0; i < cap(outbox.slots); i++ {
		outbox.slots <- i
	}
	return outbox
}

// enqueue a new element in the outbox if there is any slot available.
// Then send it to process. This method is blocking in the process channel!
func (o *outbox) enqueue(outboxmsg *outboxMsg) error {
	// first we try to obtain a slot in the outbox
	select {
	case slot := <-o.slots:
		o.queue[slot] = outboxmsg
		metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		// we send this message slot to process
		select {
		case o.process <- slot:
		case <-o.quitC:
		}
		return nil
	default:
		metrics.GetOrRegisterCounter("pss.enqueue.outbox.full", nil).Inc(1)
		return errors.New("outbox full")
	}
}

func (o *outbox) processOutbox() {
	for slot := range o.process {
		go func(slot int) {
			msg := o.msg(slot)
			metrics.GetOrRegisterResettingTimer("pss.handle.outbox", nil).UpdateSince(msg.startedAt)
			if err := o.forward(msg.msg); err != nil {
				metrics.GetOrRegisterCounter("pss.forward.err", nil).Inc(1)
				// if we failed to forward, re-insert message in the queue
				log.Debug(err.Error())
				// reenqueue the message for processing
				o.reenqueue(slot)
				log.Debug("Message re-enqued", "slot", slot)
				return
			}
			// free the outbox slot
			o.free(slot)
			metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		}(slot)
	}
}

func (o outbox) msg(slot int) *outboxMsg {
	return o.queue[slot]
}

func (o outbox) free(slot int) {
	o.slots <- slot
}

func (o outbox) reenqueue(slot int) {
	select {
	case o.process <- slot:
	case <-o.quitC:
	}

}
func (o outbox) len() int {
	return cap(o.slots) - len(o.slots)
}

type outboxMsg struct {
	msg       *PssMsg
	startedAt time.Time
}

func newOutboxMsg(msg *PssMsg) *outboxMsg {
	return &outboxMsg{
		msg:       msg,
		startedAt: time.Now(),
	}
}
