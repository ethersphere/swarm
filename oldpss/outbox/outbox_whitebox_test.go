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

	"github.com/ethersphere/swarm/oldpss/message"
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

	testOutboxMessage := testOutbox.NewOutboxMessage(&message.Message{
		To:      nil,
		Flags:   message.Flags{},
		Expire:  0,
		Topic:   message.Topic{},
		Payload: nil,
	})

	completionC := make(chan struct{})
	go func() {
		testOutbox.Enqueue(testOutboxMessage)
		completionC <- struct{}{}
	}()
	expectNotTimeout(t, completionC)

	go func() {
		testOutbox.Enqueue(testOutboxMessage)
		completionC <- struct{}{}
	}()
	expectNotTimeout(t, completionC)

	go func() {
		testOutbox.Enqueue(testOutboxMessage)
		completionC <- struct{}{}
	}()
	expectTimeout(t, completionC)

	// Now we advance the messages stuck in the forward function. At least 2 of them to leave one space available.
	processC <- struct{}{}
	processC <- struct{}{}

	// There should be a slot in the outbox to enqueue.
	select {
	case <-testOutbox.slots:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for a free slot")
	}
}

const blockTimeout = 100 * time.Millisecond

func expectNotTimeout(t *testing.T, completionC chan struct{}) {
	select {
	case <-completionC:
	case <-time.After(blockTimeout):
		t.Fatalf("timeout waiting for enqueue")
	}
}

func expectTimeout(t *testing.T, completionC chan struct{}) {
	select {
	case <-completionC:
		t.Fatalf("expected blocking enqueue")
	case <-time.After(blockTimeout):
	}
}
