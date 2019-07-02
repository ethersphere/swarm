package network

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/log"
)

type CapabilitiesAPI struct {
	*Capabilities
	notifiers map[rpc.ID]*rpc.Notifier
}

func NewCapabilitiesAPI(c *Capabilities) *CapabilitiesAPI {
	return &CapabilitiesAPI{
		Capabilities: c,
		notifiers:    make(map[rpc.ID]*rpc.Notifier),
	}
}

func (a *CapabilitiesAPI) SetCapability(id uint8, flags []byte) error {
	return a.Capabilities.set(id, flags)
}

func (a *CapabilitiesAPI) RemoveCapability(id uint8, flags []byte) error {
	return a.Capabilities.unset(id, flags)
}

func (a *CapabilitiesAPI) RegisterCapabilityModule(id uint8, length uint8) error {
	return a.Capabilities.registerModule(id, length)
}

func (a CapabilitiesAPI) notify(c capability) {
	for id, notifier := range a.notifiers {
		notifier.Notify(id, c)
	}
}

func (a CapabilitiesAPI) SubscribeChanges(ctx context.Context) (*rpc.Subscription, error) {
	notifier, ok := rpc.NotifierFromContext(ctx)
	if !ok {
		return nil, errors.New("notifications not supported")
	}
	sub := notifier.CreateSubscription()
	a.notifiers[sub.ID] = notifier
	go func(sub *rpc.Subscription, notifier *rpc.Notifier) {
		select {
		case err := <-sub.Err():
			log.Warn("rpc capabilities subscription end", "err", err)
		case <-notifier.Closed():
			log.Warn("rpc capabilities notifier closed")
		}
	}(sub, notifier)
	return sub, nil
}
