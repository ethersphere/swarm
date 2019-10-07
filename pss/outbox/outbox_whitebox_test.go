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
	"testing"
	"time"

	"github.com/ethersphere/swarm/pss/message"
)

const timeout = 2 * time.Second

// Tests that a slot in the outbox is not freed until a message is successfully forwarded.
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
	// As we haven't signaled processC, the messages are still in the outbox.
	err = testOutbox.Enqueue(testOutboxMessage)
	if err != ErrOutboxFull {
		t.Fatalf("unexpected error type, got %v, wanted %v", err, ErrOutboxFull)
	}
	processC <- struct{}{}

	// There should be a slot in the outbox to enqueue.
	select {
	case <-testOutbox.slots:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for a free slot")
	}
}

var testOutboxMessage = NewOutboxMessage(&message.Message{
	To:      nil,
	Flags:   message.Flags{},
	Expire:  0,
	Topic:   message.Topic{},
	Payload: nil,
})
