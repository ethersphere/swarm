package pss

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

// PssAPI is the RPC API module for Pss
type PssAPI struct {
	pss.PssAdapter
}

// NewPssAPI constructs a PssAPI instance
func NewPssAPI(ps pss.PssAdapter) *PssAPI {
	return &PssAPI{PssAdapter: ps}
}

// NewMsg API endpoint creates an RPC subscription
func (pssapi *PssAPI) NewMsg(ctx context.Context, topic pss.PssTopic) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}

	psssub := notifier.CreateSubscription()
	handler := func(msg []byte, p *p2p.Peer, from []byte) error {
		apimsg := &pss.PssAPIMsg{
			Msg:  msg,
			Addr: from,
		}
		if err := notifier.Notify(psssub.ID, apimsg); err != nil {
			log.Warn(fmt.Sprintf("notification on pss sub topic %v rpc (sub %v) msg %v failed!", topic, psssub.ID, msg))
		}
		return nil
	}
	deregf := pssapi.PssAdapter.Register(&topic, handler)

	go func() {
		defer deregf()
		//defer psssub.Unsubscribe()
		select {
		case err := <-psssub.Err():
			log.Warn(fmt.Sprintf("caught subscription error in pss sub topic: %v", topic, err))
		case <-notifier.Closed():
			log.Warn(fmt.Sprintf("rpc sub notifier closed"))
		}
	}()

	return psssub, nil
}

// SendRaw sends the message (serialized into byte slice) to a peer with topic
func (pssapi *PssAPI) SendRaw(topic pss.PssTopic, msg pss.PssAPIMsg) error {
	err := pssapi.PssAdapter.Send(msg.Addr, topic, msg.Msg)
	if err != nil {
		return fmt.Errorf("send error: %v", err)
	}
	return fmt.Errorf("ok sent")
}

// BaseAddr gets our own overlayaddress
func (pssapi *PssAPI) BaseAddr() ([]byte, error) {
	log.Warn("inside baseaddr")
	return pssapi.PssAdapter.BaseAddr(), nil
}
