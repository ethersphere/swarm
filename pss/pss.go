// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/capability"
	"github.com/ethersphere/swarm/oldpss/outbox"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/pss/internal/ticker"
	"github.com/ethersphere/swarm/pss/internal/ttlset"
	"github.com/ethersphere/swarm/pss/message"
	trojan "github.com/ethersphere/swarm/pss/trojan"
)

const (
	defaultMsgTTL              = time.Second * 120
	defaultDigestCacheTTL      = time.Second * 30
	defaultSymKeyCacheCapacity = 512
	CapabilityID               = capability.CapabilityID(1)
)

var (
	addressLength = len(pot.Address{})
)

// abstraction to enable access to p2p.protocols.Peer.Send
type senderPeer interface {
	Info() *p2p.PeerInfo
	ID() enode.ID
	Address() []byte
	Send(context.Context, interface{}) error
}

// per-key peer related information
// member `protected` prevents garbage collection of the instance
type peer struct {
	lastSeen  time.Time
	address   PssAddress
	protected bool
}

// Pss configuration parameters
type Params struct {
	MsgTTL              time.Duration
	CacheTTL            time.Duration
	privateKey          *ecdsa.PrivateKey
	SymKeyCacheCapacity int
	AllowRaw            bool // If true, enables sending and receiving messages without builtin pss encryption
	AllowForward        bool
}

// Sane defaults for Pss
func NewParams() *Params {
	return &Params{
		MsgTTL:              defaultMsgTTL,
		CacheTTL:            defaultDigestCacheTTL,
		SymKeyCacheCapacity: defaultSymKeyCacheCapacity,
	}
}

func (params *Params) WithPrivateKey(privatekey *ecdsa.PrivateKey) *Params {
	params.privateKey = privatekey
	return params
}

// Pss is the top-level struct, which takes care of message sending, receiving, decryption and encryption, message handler dispatchers
// and message forwarding. Implements node.Service
type Pss struct {
	*network.Kademlia // we can get the Kademlia address from this
	*KeyStore
	kademliaLB   *network.KademliaLoadBalancer
	forwardCache *ttlset.TTLSet
	gcTicker     *ticker.Ticker

	privateKey *ecdsa.PrivateKey // pss can have it's own independent key
	auxAPIs    []rpc.API         // builtins (handshake, test) can add APIs

	// sending and forwarding
	peers   map[string]*protocols.Peer // keep track of all peers sitting on the pssmsg routing layer
	peersMu sync.RWMutex

	msgTTL    time.Duration
	capstring string
	outbox    *outbox.Outbox

	// message handling
	handlers           map[message.Topic]map[*handler]bool // topic and version based pss payload handlers. See pss.Handle()
	handlersMu         sync.RWMutex
	topicHandlerCaps   map[message.Topic]*handlerCaps // caches capabilities of each topic's handlers
	topicHandlerCapsMu sync.RWMutex

	// process
	quitC chan struct{}
}

// Send a message without encryption
// Generate a trojan chunk envelope and is stored in localstore for desired targets to mine this chunk and retrieve message
func Send(ctx context.Context, localStore chunk.Store, targets [][]byte, topic string, msg []byte) (chunk.Chunk, error) {
	metrics.GetOrRegisterCounter("trojanchunk/send", nil).Inc(1)
	//construct Trojan Chunk
	t := trojan.NewTopic(topic)
	m, err := trojan.NewMessage(t, msg)
	if err != nil {
		return nil, err
	}
	var tc chunk.Chunk
	tc, err = trojan.Wrap(targets, m)
	if err != nil {
		return nil, err
	}

	//SAVE trojanChunk to localstore, if already present do not throw error
	//TODO: for second phase, use tags --> listen for response of recipient, recipient offline
	_, err = localStore.Put(ctx, chunk.ModePutUpload, tc)
	if err != nil {
		return nil, err
	}

	//TODO: verify correctness of tc?, that it will hit it's targets, should this be in the trojan package?

	return tc, nil
}

func validateAddress(addr PssAddress) error {
	if len(addr) > addressLength {
		return errors.New("address too long")
	}
	return nil
}
