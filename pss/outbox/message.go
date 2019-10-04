package outbox

import (
	"time"

	"github.com/ethersphere/swarm/pss/message"
)

type outboxMsg struct {
	msg       *message.Message
	startedAt time.Time
}

//NewOutboxMessage creates a new outbox message wrapping a pss message
func NewOutboxMessage(msg *message.Message) *outboxMsg {
	return &outboxMsg{
		msg:       msg,
		startedAt: time.Now(),
	}
}
