package syncer

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// loopback implements PubSub as a central subscription engine,
// ie a msg sent is received by all handlers registered for the topic
type loopback struct {
	handlers map[string][]func(msg []byte, p *p2p.Peer) error
}

// Register subscribes to a topic with a handler
func (lb *loopback) Register(topic string, handler func(msg []byte, p *p2p.Peer) error) {
	lb.handlers[topic] = append(lb.handlers[topic], handler)
}

// Send publishes a msg with a topic
func (lb *loopback) Send(to []byte, topic string, msg []byte) error {
	p := p2p.NewPeer(enode.ID{}, "", nil)
	for _, handler := range lb.handlers[topic] {
		if err := handler(msg, p); err != nil {
			return err
		}
	}
	return nil
}

//  tests how dispatcher of a pushsyncing node communicate with storers via PubSub
func TestProtocolWithLoopbackPubSub(t *testing.T) {
	chunkCnt := 100
	lb := &loopback{make(map[string][]func(msg []byte, p *p2p.Peer) error)}
	d := newDispatcher(network.RandomAddr().OAddr).withPubSub(lb)
	receiptsC := make(chan storage.Address, 1)
	d.processReceipt = func(a storage.Address) error {
		receiptsC <- a
		return nil
	}
	chunkStore := storage.NewMapChunkStore()
	newStorer(chunkStore).withPubSub(lb)
	timeout := time.NewTimer(100 * time.Millisecond)
	for i := 0; i < chunkCnt; i++ {
		ch := storage.GenerateRandomChunk(int64(rand.Intn(chunk.DefaultSize)))
		d.sendChunk(ch)
		select {
		case <-timeout.C:
			t.Fatalf("timeout")
		case addr := <-receiptsC:
			if !bytes.Equal(addr[:], ch.Address()[:]) {
				t.Fatalf("wrong address synced")
			}
		}
	}
}
