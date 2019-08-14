package network

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
)

type APIPeer struct {
	*Peer
	ID enode.ID
	Po int
}

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
func (k *Kademlia) GetConnsBin(addr []byte, farthestPo int, closestPo int) ([]*APIPeer, error) {

	var peers []*APIPeer

	matchPo := -1
	k.EachConn(addr, closestPo, func(sp *Peer, po int) bool {
		if matchPo < 0 {
			matchPo = po
		} else if matchPo != po {
			return false
		} else if po < farthestPo {
			return false
		}
		peers = append(peers, &APIPeer{
			Peer: sp,
			ID:   sp.BzzPeer.Peer.ID(),
			Po:   po,
		})
		return true
	})

	return peers, nil
}
