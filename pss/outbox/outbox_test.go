package outbox_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ethersphere/swarm/pss/message"
	"github.com/ethersphere/swarm/pss/outbox"
)

const timeout = 2 * time.Second

func TestOutbox(t *testing.T) {

	outboxCapacity := 2
	failedC := make(chan struct{})
	successC := make(chan struct{})
	continueC := make(chan struct{})

	forwardFail := false

	mockForwardFunction := func(msg *message.Message) error {
		if !forwardFail {
			successC <- struct{}{}
			return nil
		} else {
			failedC <- struct{}{}
			<-continueC
			return errors.New("forced test error forwarding message")
		}
	}

	testOutbox := outbox.NewMock(&outbox.Config{
		NumberSlots: outboxCapacity,
		Forward:     mockForwardFunction,
	})

	testOutbox.Start()
	defer testOutbox.Stop()

	err := testOutbox.Enqueue(testOutboxMessage)
	if err != nil {
		t.Fatalf("unexpected error enqueueing, %v", err)
	}

	//We wait for the forward function to success
	<-successC

	forwardFail = true

	err = testOutbox.Enqueue(testOutboxMessage)
	if err != nil {
		t.Fatalf("unexpected error enqueueing, %v", err)
	}
	//We wait for the forward function to fail
	select {
	case <-failedC:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for failedC")

	}

	// The message will be retried once we send to continueC, so first, we change the forward function
	forwardFail = false
	continueC <- struct{}{}

	//We wait for the retry and success
	select {
	case <-successC:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for successC")
	}
}

var testOutboxMessage = outbox.NewOutboxMessage(&message.Message{
	To:      nil,
	Flags:   message.Flags{},
	Expire:  0,
	Topic:   message.Topic{},
	Payload: nil,
})
