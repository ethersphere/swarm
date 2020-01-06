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
	"sync/atomic"

	"github.com/ethersphere/swarm/log"
)

// PubSubChannel represents a pubsub system where subscriber can .Subscribe() and publishers can .Publish() or .Close().
// When it publishes a message, it notifies all subscribers semi-asynchronously, meaning that each subscription will have
// an inbox of size inboxSize, but then a different goroutine will send those messages to the subscribers.
type PubSubChannel struct {
	subscriptions []*Subscription
	subsMutex     sync.RWMutex
	nextId        int
	quitC         chan struct{}
	inboxSize     int // size of the inbox channels in subscriptions. Depends on the number of pseudo-simultaneous messages expected to be published.
}

// Subscription is created in PubSubChannel using pubSub.Subscribe(). Subscribers can receive using .ReceiveChannel().
// or .Unsubscribe()
type Subscription struct {
	closed    bool
	pubSubC   *PubSubChannel
	inbox     chan interface{}
	signal    chan interface{}
	closeOnce sync.Once
	id        string
	lock      sync.RWMutex
	quitC     chan struct{} // close channel for publisher goroutines
	msgCount  int
	pending   *int64
}

// New creates a new PubSubChannel.
func New(inboxSize int) *PubSubChannel {
	return &PubSubChannel{
		subscriptions: make([]*Subscription, 0),
		quitC:         make(chan struct{}),
		inboxSize:     inboxSize,
	}
}

// Subscribe creates a subscription to a channel, each subscriber should keep its own Subscription instance.
func (psc *PubSubChannel) Subscribe() *Subscription {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()
	newSubscription := newSubscription(strconv.Itoa(psc.nextId), psc, psc.inboxSize)
	psc.nextId++
	psc.subscriptions = append(psc.subscriptions, newSubscription)

	return newSubscription
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

// Publish broadcasts a message synchronously to each subscriber inbox.
func (psc *PubSubChannel) Publish(msg interface{}) {
	psc.subsMutex.RLock()
	defer psc.subsMutex.RUnlock()
	for _, sub := range psc.subscriptions {
		psc.publishToSub(sub, msg)
	}
}

// publishToSub will block on the subscription inbox if there are more than inboxSize messages accumulated
func (psc *PubSubChannel) publishToSub(sub *Subscription, msg interface{}) {
	atomic.AddInt64(sub.pending, 1)
	defer atomic.AddInt64(sub.pending, -1)
	select {
	case <-psc.quitC:
	case <-sub.quitC:
	case sub.inbox <- msg:
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
		close(sub.quitC)
		sub.lock.Unlock()
	}
	close(psc.quitC)
}

// Unsubscribe cancels subscription from the subscriber side. Channel is marked as closed but only writer should close it.
func (sub *Subscription) Unsubscribe() {
	close(sub.quitC)
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

func (sub *Subscription) MessageCount() int {
	return sub.msgCount
}

func (sub *Subscription) Pending() int64 {
	return *sub.pending
}

func newSubscription(id string, psc *PubSubChannel, inboxSize int) *Subscription {
	var pending int64
	subscription := &Subscription{
		closed:    false,
		pubSubC:   psc,
		inbox:     make(chan interface{}, inboxSize),
		signal:    make(chan interface{}),
		closeOnce: sync.Once{},
		id:        id,
		quitC:     make(chan struct{}),
		msgCount:  0,
		pending:   &pending,
	}
	// publishing goroutine. It closes the signal channel whenever it receives the quitC signal
	go func(sub *Subscription) {
		for {
			select {
			case <-sub.quitC:
				close(sub.signal)
				return
			case msg := <-sub.inbox:
				log.Debug("Retrieved inbox message", "msg", msg)
				select {
				case <-psc.quitC:
					return
				case <-sub.quitC:
					close(sub.signal)
					return
				case sub.signal <- msg:
					sub.msgCount++
				}
			case <-psc.quitC:
				return
			}
		}
	}(subscription)
	return subscription
}
