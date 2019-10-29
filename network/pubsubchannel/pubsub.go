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
package pubsubchannel

import (
	"strconv"
	"sync"
	"time"

	"github.com/ethersphere/swarm/log"
)

var subscriptionTimeout = 100 * time.Millisecond

//PubSubChannel represents a pubsub system where subscriber can .Subscribe() and publishers can .Publish() or .Close().
type PubSubChannel struct {
	subscriptions []*Subscription
	subsMutex     sync.RWMutex
	nextId        int
}

// Subscription is created in PubSubChannel using pubSub.Subscribe(). Subscribers can receive using .ReceiveChannel().
// or .Unsubscribe()
type Subscription struct {
	closed  bool
	pubSubC *PubSubChannel
	//removeSub func()
	signal    chan interface{}
	closeOnce sync.Once
	id        string
	lock      sync.RWMutex
}

// New creates a new PubSubChannel.
func New() *PubSubChannel {
	return &PubSubChannel{
		subscriptions: make([]*Subscription, 0),
	}
}

// Subscribe creates a subscription to a channel, each subscriber should keep its own Subscription instance.
func (psc *PubSubChannel) Subscribe() *Subscription {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()
	newSubscription := newSubscription(strconv.Itoa(psc.nextId), psc)
	psc.nextId++
	psc.subscriptions = append(psc.subscriptions, &newSubscription)

	return &newSubscription
}

func (psc *PubSubChannel) removeSub(s *Subscription) {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()

	for i, subscription := range psc.subscriptions {
		if subscription.signal == s.signal {
			log.Debug("Unsubscribing", "id", subscription.id)
			subscription.lock.Lock()
			subscription.closed = true
			subscription.lock.Unlock()
			psc.subscriptions = append(psc.subscriptions[:i], psc.subscriptions[i+1:]...)
		}
	}
}

// Publish broadcasts a message asynchronously to each subscriber.
// If some of the subscriptions(channels) has been marked as closeable, it does it now.
func (psc *PubSubChannel) Publish(msg interface{}) {
	psc.subsMutex.RLock()
	defer psc.subsMutex.RUnlock()
	for _, sub := range psc.subscriptions {
		sub.lock.RLock()
		if sub.closed {
			log.Debug("Subscription was closed", "id", sub.id)
			sub.closeChannel()
		} else {
			go func(sub *Subscription) {
				select {
				case sub.signal <- msg:
				case <-time.After(subscriptionTimeout):
					log.Warn("Subscription unattended after timeout", "subId", sub.ID(), "timeout", subscriptionTimeout)
				}
			}(sub)
		}
		sub.lock.RUnlock()
	}
}

// NumSubscriptions returns how many subscriptions are currently active.
func (psc *PubSubChannel) NumSubscriptions() int {
	psc.subsMutex.RLock()
	defer psc.subsMutex.RUnlock()
	return len(psc.subscriptions)
}

// Close cancels all subscriptions closing the channels associated with them.
// Usually the publisher is in charge of calling Close().
func (psc *PubSubChannel) Close() {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()
	for _, sub := range psc.subscriptions {
		sub.lock.Lock()
		sub.closed = true
		sub.closeChannel()
		sub.lock.Unlock()
	}
}

// Unsubscribe cancels subscription from the subscriber side. Channel is marked as closed but only writer should close it.
func (sub *Subscription) Unsubscribe() {
	sub.pubSubC.removeSub(sub)
}

// ReceiveChannel returns the channel where the subscriber will receive messages.
func (sub *Subscription) ReceiveChannel() <-chan interface{} {
	return sub.signal
}

// IsClosed returns if the subscription is closed via Unsubscribe() or Close() in the pubSub that creates it.
func (sub *Subscription) IsClosed() bool {
	sub.lock.RLock()
	defer sub.lock.RUnlock()
	return sub.closed
}

// ID returns a unique id in the PubSubChannel of this subscription. Useful for debugging.
func (sub *Subscription) ID() string {
	return sub.id
}

func (sub *Subscription) closeChannel() {
	sub.closeOnce.Do(func() {
		close(sub.signal)
	})
}

func newSubscription(id string, psc *PubSubChannel) Subscription {
	return Subscription{
		closed:    false,
		pubSubC:   psc,
		signal:    make(chan interface{}, 20),
		closeOnce: sync.Once{},
		id:        id,
	}
}
