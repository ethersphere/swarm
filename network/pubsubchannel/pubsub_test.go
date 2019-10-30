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
	"os"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/network/pubsubchannel"
)

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
		log.Debug("Finishing subscriptor gofunc", "id", subscription.ID())
	}(subscription)
	return msgBucket, subscription
}

func TestUnsubscribeBeforeReadingMessages(t *testing.T) {
	ps := pubsubchannel.New()
	s := ps.Subscribe()
	defer ps.Close()

	for i := 0; i < 1000; i++ {
		ps.Publish(struct{}{})
	}

	s.Unsubscribe()
	// allow goroutines to finish
	time.Sleep(100 * time.Millisecond)
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
}
