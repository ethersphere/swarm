package network

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
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

func (kn KademliaNotification) String() string {
	return fmt.Sprintf("%d:%d", kn.Serial, kn.Depth)
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

func (k *Kademlia) GetDepth() (int, error) {
	return k.NeighbourhoodDepth(), nil
}

// TODO: return po from rpc
func (k *Kademlia) GetConnsBin(addr []byte, farthestPo int, closestPo int) ([]*Peer, error) {

	var peers []*Peer

	matchPo := -1
	k.EachConn(addr, closestPo, func(sp *Peer, po int) bool {
		if matchPo < 0 {
			matchPo = po
		} else if matchPo != po {
			return false
		} else if po < farthestPo {
			return false
		}
		peers = append(peers, sp)
		log.Warn("found peer", "peer", sp.Peer)
		return true
	})
	if matchPo == -1 {
		matchPo = 0
	}
	log.Debug("matchpo", "po", matchPo)

	return peers, nil
}
