package network

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pot"
)

type NotificationPeer struct {
	Peer *BzzAddr
	Po   int
}

type KademliaNotification struct {
	Peers  []*NotificationPeer
	Depth  uint8
	Serial uint16
}

func (k *Kademlia) notify(depth uint8, serial uint16, peers ...*NotificationPeer) {
	k.notifyLock.RLock()
	defer k.notifyLock.RUnlock()
	notification := KademliaNotification{
		Depth:  depth,
		Serial: serial,
	}
	for _, p := range peers {
		notification.Peers = append(notification.Peers, p)
	}
	for subId, notifier := range k.subs {
		notifier.Notify(subId, notification)
	}
}

func (k *Kademlia) Receive(ctx context.Context) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}

	sub := notifier.CreateSubscription()
	k.notifyLock.Lock()
	k.subs[sub.ID] = notifier
	k.notifyLock.Unlock()
	return sub, nil
}

func (k *Kademlia) GetConnsBin(addr []byte, closestPo int) ([]*Peer, error) {
	neighbourhoodDepth := k.NeighbourhoodDepth()

	// luminosity is the opposite of darkness. the more bytes are removed from the address, the higher is darkness,
	// but the luminosity is less. here luminosity equals the number of bits given in the destination address.
	luminosityRadius := len(addr) * 8
	padAddr := make([]byte, AddressLength)
	copy(padAddr, addr)

	// proximity order function matching up to neighbourhoodDepth bits (po <= neighbourhoodDepth)
	pof := pot.DefaultPof(neighbourhoodDepth)

	// soft threshold for msg broadcast
	broadcastThreshold, _ := pof(padAddr, k.BaseAddr(), 0)
	if broadcastThreshold > luminosityRadius {
		broadcastThreshold = luminosityRadius
	}

	// if measured from the recipient address as opposed to the base address (see Kademlia.EachConn
	// call below), then peers that fall in the same proximity bin as recipient address will appear
	// [at least] one bit closer, but only if these additional bits are given in the recipient address.
	if broadcastThreshold < luminosityRadius && broadcastThreshold < neighbourhoodDepth {
		broadcastThreshold++
	}

	var peers []*Peer
	if closestPo < broadcastThreshold {
		return peers, nil
	}

	matchPo := -1
	k.EachConn(padAddr, closestPo, func(sp *Peer, po int) bool {
		if matchPo < 0 {
			matchPo = po
		} else if matchPo != po {
			return false
		} else if po < broadcastThreshold {
			return false
		}
		peers = append(peers, sp)
		return true
	})
	if matchPo == -1 {
		matchPo = 0
	}
	log.Debug("matchpo", "po", matchPo)

	return peers, nil
}
