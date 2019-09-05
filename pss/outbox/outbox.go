package outbox

import (
	"errors"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/message"
)

type Config struct {
	NumberSlots int
	Forward     forwardFunction
}

type Outbox struct {
	forwardFunc forwardFunction
	queue       []*outboxMsg
	slots       chan int
	process     chan int
	stopC       chan struct{}
}

type forwardFunction func(msg *message.Message) error

var ErrOutboxFull = errors.New("outbox full")

func NewOutbox(config *Config) *Outbox {
	outbox := &Outbox{
		forwardFunc: config.Forward,
		queue:       make([]*outboxMsg, config.NumberSlots),
		slots:       make(chan int, config.NumberSlots),
		process:     make(chan int),
		stopC:       make(chan struct{}),
	}
	// fill up outbox slots
	for i := 0; i < cap(outbox.slots); i++ {
		outbox.slots <- i
	}
	return outbox
}

func (o *Outbox) Start() {
	log.Info("Starting outbox")
	go o.processOutbox()
}

func (o *Outbox) Stop() {
	log.Info("Stopping outbox")
	close(o.stopC)
}

// Enqueue a new element in the outbox if there is any slot available.
// Then send it to process. This method is blocking in the process channel!
func (o *Outbox) Enqueue(outboxMsg *outboxMsg) error {
	// first we try to obtain a slot in the outbox
	select {
	case slot := <-o.slots:
		o.queue[slot] = outboxMsg
		metrics.GetOrRegisterGauge("pss.outbox.len", nil).Update(int64(o.len()))
		// we send this message slot to process
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

//ProcessOutbox starts a routine that tries to forward messages present in the outbox queue
func (o *Outbox) processOutbox() {
	for slot := range o.process {
		go func(slot int) {
			msg := o.queue[slot]
			metrics.GetOrRegisterResettingTimer("pss.handle.outbox", nil).UpdateSince(msg.startedAt)
			if err := o.forwardFunc(msg.msg); err != nil {
				metrics.GetOrRegisterCounter("pss.forward.err", nil).Inc(1)
				log.Debug(err.Error())
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
