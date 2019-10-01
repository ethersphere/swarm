package gopubsub

import (
	"fmt"
	"strconv"
	"sync"
)

//PubSubChannel represents a pubsub system where subscriber can .Subscribe() and publishers can .Publish() or .Close().
type PubSubChannel struct {
	subscriptions []*Subscription
	subsMutex     sync.RWMutex
	nextId        int
}

//Subscription is created in PubSubChannel using pubSub.Subscribe(). Subscribers can receive using .ReceiveChannel().
// or .Unsubscribe()
type Subscription struct {
	closed    bool
	removeSub func()
	signal    chan interface{}
	closeOnce sync.Once
	id        string
}

//New creates a new PubSubChannel.
func New() *PubSubChannel {
	return &PubSubChannel{
		subscriptions: make([]*Subscription, 0),
	}
}

//Subscribe creates a subscription to a channel, each subscriber should keep its own Subscription instance.
func (psc *PubSubChannel) Subscribe() *Subscription {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()
	newSubscription := newSubscription(strconv.Itoa(psc.nextId))
	psc.nextId++
	psc.subscriptions = append(psc.subscriptions, &newSubscription)
	newSubscription.removeSub = func() {
		psc.subsMutex.Lock()
		defer psc.subsMutex.Unlock()

		for i, subscription := range psc.subscriptions {
			if subscription.signal == newSubscription.signal {
				fmt.Println("Unsubscribing", "id", subscription.id)
				subscription.closed = true
				psc.subscriptions = append(psc.subscriptions[:i], psc.subscriptions[i+1:]...)
			}
		}
	}
	return &newSubscription
}

//Publish broadcasts a message asynchronously to each subscriber.
//If some of the subscriptions(channels) has been marked as closeable, it does it now.
func (psc *PubSubChannel) Publish(msg interface{}) {
	psc.subsMutex.RLock()
	defer psc.subsMutex.RUnlock()
	for i, sub := range psc.subscriptions {
		if sub.closed {
			fmt.Println("Subscription was closed", "id", sub.id)
			sub.closeChannel()
		} else {
			go func(sub *Subscription, index int) {
				sub.signal <- msg
			}(sub, i)

		}
	}
}

//NumSubscriptions returns how many subscriptions are currently active.
func (psc *PubSubChannel) NumSubscriptions() int {
	psc.subsMutex.RLock()
	defer psc.subsMutex.RUnlock()
	return len(psc.subscriptions)
}

//Close cancels all subscriptions closing the channels associated with them.
//Usually the publisher is in charge of calling Close().
func (psc *PubSubChannel) Close() {
	psc.subsMutex.Lock()
	defer psc.subsMutex.Unlock()
	for _, sub := range psc.subscriptions {
		sub.closed = true
		sub.closeChannel()
	}
}

//Unsubscribe cancels subscription from the subscriber side. Channel is marked as closed but only writer should close it.
func (sub *Subscription) Unsubscribe() {
	sub.closed = true
	sub.removeSub()
}

//ReceiveChannel returns the channel where the subscriber will receive messages.
func (sub *Subscription) ReceiveChannel() <-chan interface{} {
	return sub.signal
}

//IsClosed returns if the subscription is closed via Unsubscribe() or Close() in the pubSub that creates it.
func (sub *Subscription) IsClosed() bool {
	return sub.closed
}

//ID returns a unique id in the PubSubChannel of this subscription. Useful for debugging.
func (sub *Subscription) ID() string {
	return sub.id
}

func (sub *Subscription) closeChannel() {
	sub.closeOnce.Do(func() {
		close(sub.signal)
	})
}

func newSubscription(id string) Subscription {
	return Subscription{
		closed:    false,
		removeSub: nil,
		signal:    make(chan interface{}),
		closeOnce: sync.Once{},
		id:        id,
	}
}
