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
package outbox_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethersphere/swarm/pss/message"
	"github.com/ethersphere/swarm/pss/outbox"
	"github.com/tilinna/clock"
)

const timeout = 2 * time.Second

// Tests successful and failed forwarding. Failure to forward should requeue the failed message.
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

	// We wait for the forward function to success.
	<-successC

	forwardFail = true

	err = testOutbox.Enqueue(testOutboxMessage)
	if err != nil {
		t.Fatalf("unexpected error enqueueing, %v", err)
	}
	// We wait for the forward function to fail
	select {
	case <-failedC:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for failedC")

	}

	// The message will be retried once we send to continueC, so first, we change the forward function.
	forwardFail = false
	continueC <- struct{}{}

	// We wait for the retry and success.
	select {
	case <-successC:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for successC")
	}
}

// TestOutboxWorkers checks that the number of goroutines for processing have a maximum and that there is no
// deadlock operating.
func TestOutboxWorkers(t *testing.T) {
	outboxCapacity := 100
	workers := 3
	endProcess := make(chan struct{}, outboxCapacity)

	var parallel int32
	var maxParallel int32

	var wg sync.WaitGroup
	mockForwardFunction := func(msg *message.Message) error {
		atomic.AddInt32(&parallel, 1)
		if parallel > maxParallel {
			atomic.StoreInt32(&maxParallel, parallel)
			maxParallel = parallel
		}
		<-endProcess
		atomic.AddInt32(&parallel, -1)
		wg.Done()
		return nil
	}

	testOutbox := outbox.NewMock(&outbox.Config{
		NumberSlots: outboxCapacity,
		Forward:     mockForwardFunction,
		NumWorkers:  workers,
	})

	testOutbox.Start()
	defer testOutbox.Stop()

	numMessages := 100

	// Enqueuing numMessages messages in parallel.
	wg.Add(numMessages)
	for i := 0; i < numMessages; i++ {
		go func(num byte) {
			testOutbox.Enqueue(outbox.NewOutboxMessage(newTestMessage(num)))
		}(byte(i))
	}

	// We need this sleep to allow workers to be launched and accumulated before starting signaling the channel.
	// If not, there never will be workers in parallel and the test will be useless.
	time.Sleep(10 * time.Millisecond)
	// Signaling 100 messages.
	for i := 0; i < numMessages; i++ {
		endProcess <- struct{}{}
	}

	wg.Wait()

	if int(maxParallel) > workers {
		t.Errorf("Expected maximum %v worker(s) in parallel but got %v", workers, maxParallel)
	}
}

func newTestMessage(num byte) *message.Message {
	return &message.Message{
		To:      nil,
		Flags:   message.Flags{},
		Expire:  0,
		Topic:   message.Topic{},
		Payload: []byte{num},
	}
}

func TestMessageRetriesExpired(t *testing.T) {
	failForwardFunction := func(msg *message.Message) error {
		return errors.New("forward error")
	}
	msg := outbox.NewOutboxMessage(&message.Message{
		To:      nil,
		Flags:   message.Flags{},
		Expire:  0,
		Topic:   message.Topic{},
		Payload: nil,
	})
	// We are going to simulate that 5 minutes has passed with a mock Clock
	duration := 5 * time.Minute
	now := time.Now()
	mockClock := clock.NewMock(now)
	testOutbox := outbox.NewMock(&outbox.Config{
		NumberSlots:  1,
		Forward:      failForwardFunction,
		MaxRetryTime: &duration,
		Clock:        mockClock,
	})

	testOutbox.Start()
	defer testOutbox.Stop()

	err := testOutbox.Enqueue(msg)
	if err != nil {
		t.Errorf("Expected no error enqueueing, instead got %v", err)
	}

	numMessages := testOutbox.CurrentSize()
	if numMessages != 1 {
		t.Errorf("Expected one message in outbox, instead got %v", numMessages)
	}

	mockClock.Set(now.Add(duration / 2))
	numMessages = testOutbox.CurrentSize()
	if numMessages != 1 {
		t.Errorf("Expected one message in outbox after half maxRetryTime, instead got %v", numMessages)
	}

	mockClock.Set(now.Add(duration + 1*time.Nanosecond))
	// Just sleep a moment to allow the process routine to retry and expire message
	time.Sleep(10 * time.Millisecond)
	numMessages = testOutbox.CurrentSize()
	if numMessages != 0 {
		t.Errorf("Expected 0 message in outbox after expired message, instead got %v", numMessages)
	}
}

var testOutboxMessage = outbox.NewOutboxMessage(&message.Message{
	To:      nil,
	Flags:   message.Flags{},
	Expire:  0,
	Topic:   message.Topic{},
	Payload: nil,
})
