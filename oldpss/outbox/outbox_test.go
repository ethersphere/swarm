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
	"testing"
	"time"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/oldpss/message"
	"github.com/ethersphere/swarm/oldpss/outbox"
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

	// We wait for the forward function to success.
	<-successC

	forwardFail = true

	go func() {
		testOutbox.Enqueue(testOutboxMessage)
		completionC <- struct{}{}
	}()
	expectNotTimeout(t, completionC)

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
	outboxCapacity := 3
	endProcess := make(chan struct{}, outboxCapacity)

	var parallel int32
	var maxParallel int32

	var wg sync.WaitGroup
	var mtx sync.Mutex
	mockForwardFunction := func(msg *message.Message) error {
		mtx.Lock()
		parallel++
		if parallel > maxParallel {
			maxParallel = parallel
		}
		mtx.Unlock()

		<-endProcess
		mtx.Lock()
		parallel--
		mtx.Unlock()

		wg.Done()
		return nil
	}

	testOutbox := outbox.NewMock(&outbox.Config{
		NumberSlots: outboxCapacity,
		Forward:     mockForwardFunction,
	})

	testOutbox.Start()
	defer testOutbox.Stop()

	numMessages := 100

	// Enqueuing numMessages messages in parallel.
	wg.Add(numMessages)
	for i := 0; i < numMessages; i++ {
		go func(num byte) {
			testOutbox.Enqueue(testOutbox.NewOutboxMessage(newTestMessage(num)))
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

	if int(maxParallel) > outboxCapacity {
		t.Errorf("Expected maximum %v worker(s) in parallel but got %v", outboxCapacity, maxParallel)
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
		t.Fatalf("epxected blocking enqueue")
	case <-time.After(blockTimeout):
	}
}

func TestMessageRetriesExpired(t *testing.T) {
	failForwardFunction := func(msg *message.Message) error {
		return errors.New("forward error")
	}

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

	msg := testOutbox.NewOutboxMessage(&message.Message{
		To:      nil,
		Flags:   message.Flags{},
		Expire:  0,
		Topic:   message.Topic{},
		Payload: nil,
	})

	completionC := make(chan struct{})
	go func() {
		testOutbox.Enqueue(msg)
		completionC <- struct{}{}
	}()

	expectNotTimeout(t, completionC)

	numMessages := testOutbox.Len()
	if numMessages != 1 {
		t.Errorf("Expected one message in outbox, instead got %v", numMessages)
	}

	mockClock.Set(now.Add(duration / 2))
	numMessages = testOutbox.Len()
	// Now we wait a bit expecting that the number of messages doesn't change
	iterations := 0
	for numMessages == 1 && iterations < 2 {
		// Wait a bit more to check that the message has not been expired.
		time.Sleep(10 * time.Millisecond)
		iterations++
		numMessages = testOutbox.Len()
	}
	if numMessages != 1 {
		t.Errorf("Expected one message in outbox after half maxRetryTime, instead got %v", numMessages)
	}

	mockClock.Set(now.Add(duration + 1*time.Millisecond))
	numMessages = testOutbox.Len()
	// Now we wait for the process routine to retry and expire message at least 10 iterations
	iterations = 0
	for numMessages != 0 && iterations < 10 {
		// Still not expired, wait a bit more
		log.Debug("Still not there, waiting another iteration", numMessages, iterations)
		time.Sleep(10 * time.Millisecond)
		iterations++
		numMessages = testOutbox.Len()
	}
	if numMessages != 0 {
		t.Errorf("Expected 0 message in outbox after expired message, instead got %v", numMessages)
	}

}
