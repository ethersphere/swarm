package outbox

import (
	"testing"

	"github.com/ethersphere/swarm/pss/message"
)

func TestFullOutbox(t *testing.T) {

	outboxCapacity := 2
	processC := make(chan struct{})
	successForward := func(msg *message.Message) error {
		<-processC
		return nil
	}

	testOutbox := NewMock(&Config{
		NumberSlots: outboxCapacity,
		Forward:     successForward,
	})
	testOutbox.Start()
	defer testOutbox.Stop()

	err := testOutbox.Enqueue(testOutboxMessage)
	if err != nil {
		t.Fatalf("unexpected error enqueueing, %v", err)
	}

	err = testOutbox.Enqueue(testOutboxMessage)
	if err != nil {
		t.Fatalf("unexpected error enqueueing, %v", err)
	}
	//As we haven't signaled processC, the messages are still in the outbox
	err = testOutbox.Enqueue(testOutboxMessage)
	if err != ErrOutboxFull {
		t.Fatalf("unexpected error, got %v, wanted %v", err, ErrOutboxFull)
	}
	processC <- struct{}{}

	//There should be a slot again in the outbox to enqueue
	<-testOutbox.slots
}

var testOutboxMessage = NewOutboxMessage(&message.Message{
	To:      nil,
	Flags:   message.Flags{},
	Expire:  0,
	Topic:   message.Topic{},
	Payload: nil,
})
