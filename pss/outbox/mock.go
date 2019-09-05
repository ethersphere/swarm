package outbox

import (
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/message"
)

const (
	defaultOutboxCapacity = 100000
)

var mockForwardFunction = func(msg *message.Message) error {
	log.Debug("Forwarded message", "msg", msg)
	return nil
}

func NewMock(config *Config) (outboxMock *Outbox) {
	if config == nil {
		config = &Config{
			NumberSlots: defaultOutboxCapacity,
			Forward:     mockForwardFunction,
		}
	} else {
		if config.Forward == nil {
			config.Forward = mockForwardFunction
		}
		if config.NumberSlots == 0 {
			config.NumberSlots = defaultOutboxCapacity
		}
	}
	return NewOutbox(config)
}
