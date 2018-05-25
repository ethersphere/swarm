package mru

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/pss/notify"
)

type Notification struct {
	Name    string
	Period  uint32
	Version uint32
}

type API struct {
	subscribers   map[string]*rpc.Notifier
	subscriptions map[string]rpc.ID
	//notifier      *notify.Controller
	resourceHandler *ResourceHandler
}

func NewAPI(rh *ResourceHandler) *API {
	api := &API{
		subscribers:     make(map[string]*rpc.Notifier),
		subscriptions:   make(map[string]rpc.ID),
		resourceHandler: rh,
	}
	return api
}

func (self *API) RequestNotification(ctx context.Context, name string, pubkeyHex hexutil.Bytes, address hexutil.Bytes) (*rpc.Subscription, error) {
	pubkey := crypto.ToECDSAPub(pubkeyHex)
	self.resourceHandler.notifier.Request(name, pubkey, pss.PssAddress(address), self.handle)
	rpcnotifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}
	sub := rpcnotifier.CreateSubscription()
	self.subscribers[name] = rpcnotifier
	self.subscriptions[name] = sub.ID
	return sub, nil
}

// EnableNotifications turns on pss notifications for the specified resource
// The resource has to be loaded
func (self *API) EnableNotifications(name string) error {
	nameHash := ens.EnsNode(name)
	if _, ok := self.resourceHandler.resources[nameHash.Hex()]; !ok {
		return fmt.Errorf("Unknown resource '%s'", name)
	}
	self.resourceHandler.notifier.NewNotifier(name, notify.DefaultAddressLength, self.resourceHandler.CreateNotification)
	return nil
}

func (self *API) handle(name string, data []byte) error {
	period := binary.LittleEndian.Uint32(data[:4])
	version := binary.LittleEndian.Uint32(data[4:])
	return self.subscribers[name].Notify(self.subscriptions[name], &Notification{
		Name:    name,
		Period:  period,
		Version: version,
	})
}
