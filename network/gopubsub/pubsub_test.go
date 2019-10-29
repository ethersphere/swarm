package gopubsub_test

import (
    "fmt"
    "github.com/ethersphere/swarm/network/gopubsub"
    "sync"
    "testing"
)

func TestPubSeveralSub(t *testing.T) {
    pubSub := gopubsub.New()
    var group sync.WaitGroup
    bucketSubs1, _ := testSubscriptor(pubSub, 2, &group)
    bucketSubs2, _ := testSubscriptor(pubSub, 2, &group)

    fmt.Println("Adding message 0")
    pubSub.Publish(struct {}{})
    fmt.Println("Adding message 1")
    pubSub.Publish(struct {}{})
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
    pubSub := gopubsub.New()
    var group sync.WaitGroup
    _, subscription := testSubscriptor(pubSub, 0, &group)
    msgBucket2, _ := testSubscriptor(pubSub, 1, &group)
    pubSub.Publish(struct {}{})
    group.Wait()
    if len(msgBucket2) != 1 {
        t.Errorf("Subscriptor 2 should have received 1 message regardless of sub 1 unsubscribing, instead %v", len(msgBucket2))
    }

    if pubSub.NumSubscriptions() == 2 || !subscription.IsClosed() {
        t.Errorf("Subscription should have been closed")
    }
}

func testSubscriptor( pubsub *gopubsub.PubSubChannel, expectedMessages int, group *sync.WaitGroup) (map[int]interface{}, *gopubsub.Subscription) {
    msgBucket := make(map[int]interface{})
    subscription := pubsub.Subscribe()
    group.Add(1)
    go func(subscription *gopubsub.Subscription) {
        defer group.Done()
        if expectedMessages == 0 {
            subscription.Unsubscribe()
            return
        }
        var i int
        for msg := range subscription.ReceiveChannel() {
            fmt.Println("Received message", "id", subscription.ID(), "msg", msg)
            msgBucket[i] = msg
            i++
            if i >= expectedMessages {
                return
            }
        }
        fmt.Println("Finishing subscriptor gofunc", "id", subscription.ID())
    }(subscription)
    return msgBucket, subscription
}
