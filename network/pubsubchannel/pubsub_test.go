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
package pubsubchannel_test

import (
	"fmt"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/pubsubchannel"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

func TestPubSeveralSub(t *testing.T) {
	pubSub := pubsubchannel.New()
	var group sync.WaitGroup
	bucketSubs1, _ := testSubscriptor(pubSub, 2, &group)
	bucketSubs2, _ := testSubscriptor(pubSub, 2, &group)

	log.Debug("Adding message 0")
	pubSub.Publish(struct{}{})
	log.Debug("Adding message 1")
	pubSub.Publish(struct{}{})
	group.Wait()
	pubSub.Close()
	if len(bucketSubs1) != 2 {
		t.Errorf("Subscriptor 1 should have received 2 message, instead %v", len(bucketSubs1))
	}

	if len(bucketSubs2) != 2 {
		t.Errorf("Subscriptor 1 should have received 2 message, instead %v", len(bucketSubs2))
	}

}

func TestPubUnsubscribe(t *testing.T) {
	pubSub := pubsubchannel.New()
	var group sync.WaitGroup
	_, subscription := testSubscriptor(pubSub, 0, &group)
	msgBucket2, _ := testSubscriptor(pubSub, 1, &group)
	pubSub.Publish(struct{}{})
	group.Wait()
	if len(msgBucket2) != 1 {
		t.Errorf("Subscriptor 2 should have received 1 message regardless of sub 1 unsubscribing, instead %v", len(msgBucket2))
	}

	if pubSub.NumSubscriptions() == 2 || !subscription.IsClosed() {
		t.Errorf("Subscription should have been closed")
	}
}

func testSubscriptor(pubsub *pubsubchannel.PubSubChannel, expectedMessages int, group *sync.WaitGroup) (map[int]interface{}, *pubsubchannel.Subscription) {
	msgBucket := make(map[int]interface{})
	subscription := pubsub.Subscribe()
	group.Add(1)
	go func(subscription *pubsubchannel.Subscription) {
		defer group.Done()
		if expectedMessages == 0 {
			subscription.Unsubscribe()
			return
		}
		var i int
		for msg := range subscription.ReceiveChannel() {
			log.Debug("Received message", "id", subscription.ID(), "msg", msg)
			msgBucket[i] = msg
			i++
			if i >= expectedMessages {
				return
			}
		}
		log.Debug("Finishing subscriber gofunc", "id", subscription.ID())
	}(subscription)
	return msgBucket, subscription
}

// TestUnsubscribeBeforeReadingMessages tests that there is no goroutine leak when a subscription is finished
// before reading pending messages from the channel.
func TestUnsubscribeBeforeReadingMessages(t *testing.T) {
	ps := pubsubchannel.New()
	s := ps.Subscribe()
	defer ps.Close()

	for i := 0; i < 1000; i++ {
		ps.Publish(struct{}{})
	}

	s.Unsubscribe()
	// allow goroutines to finish, no pending messages
	var pendingMessages int64
	for i := 0; i < 500 && pendingMessages > 0; i++ {
		time.Sleep(10 * time.Millisecond)
		pendingMessages = s.Pending()
		if pendingMessages <= 0 {
			break
		}
	}

	if pendingMessages > 0 {
		t.Errorf("%v new goroutines were active after unsubscribe, want none", pendingMessages)
		pprof.Lookup("goroutine").WriteTo(newTestingErrorWriter(t), 1)
	}
}

type testingErrorWriter struct {
	t *testing.T
}

func newTestingErrorWriter(t *testing.T) testingErrorWriter {
	return testingErrorWriter{t: t}
}

func (w testingErrorWriter) Write(b []byte) (int, error) {
	w.t.Error(string(b))
	return len(b), nil
}

// TestMessageAfterUnsubscribe checks that if some pending message are still readable from the channel, after
// Unsubscribe(), the publishing goroutines will be exited and no message is received in the channel (even though the
// channel is still not closed). However, we need to wait a bit before extracting messages from the channel to allow
// the blocked publishers exit. In a real case, the moment a new message is published the channel will be closed.
func TestMessagesAfterUnsubscribe(t *testing.T) {
	ps := pubsubchannel.New()
	defer ps.Close()

	s := ps.Subscribe()

	for i := 0; i < 1000; i++ {
		ps.Publish(fmt.Sprintf("Message %v", i))
	}
	c := s.ReceiveChannel()

	s.Unsubscribe()

	ps.Publish("End message")
	var n int
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case _, ok := <-c:
			if !ok {
				break loop
			}
			n++
		case <-timeout:
			t.Log("timeout")
			break loop
		}
	}

	t.Log("got", n, "messages")
	if n > 0 {
		t.Errorf("Expected no message received after unsubscribing but got %v", n)
	}

}
