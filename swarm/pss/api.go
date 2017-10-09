package pss

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

// Wrapper for receiving pss messages when using the pss API
// providing access to sender of message
type APIMsg struct {
	Msg        []byte
	Asymmetric bool
	Key        string
}

// Additional public methods accessible through API for pss
type API struct {
	*Pss
}

func NewAPI(ps *Pss) *API {
	return &API{Pss: ps}
}

// Creates a new subscription for the caller. Enables external handling of incoming messages.
//
// A new handler is registered in pss for the supplied topic
//
// All incoming messages to the node matching this topic will be encapsulated in the APIMsg
// struct and sent to the subscriber
func (pssapi *API) Receive(ctx context.Context, topic whisper.TopicType) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}

	psssub := notifier.CreateSubscription()

	handler := func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		apimsg := &APIMsg{
			Msg:        msg,
			Asymmetric: asymmetric,
			Key:        keyid,
		}
		if err := notifier.Notify(psssub.ID, apimsg); err != nil {
			log.Warn(fmt.Sprintf("notification on pss sub topic rpc (sub %v) msg %v failed!", psssub.ID, msg))
		}
		return nil
	}
	deregf := pssapi.Register(&topic, handler)
	go func() {
		defer deregf()
		select {
		case err := <-psssub.Err():
			log.Warn(fmt.Sprintf("caught subscription error in pss sub topic %x: %v", topic, err))
		case <-notifier.Closed():
			log.Warn(fmt.Sprintf("rpc sub notifier closed"))
		}
	}()

	return psssub, nil
}

// Retrieves the node's public key in byte form
func (pssapi *API) GetPublicKey() (keybytes []byte) {
	key := pssapi.Pss.PublicKey()
	keybytes = crypto.FromECDSAPub(key)
	return keybytes
}

// Set Public key to associate with a particular Pss peer
func (pssapi *API) SetPeerPublicKey(pubkey []byte, topic whisper.TopicType, addr PssAddress) error {
	err := pssapi.Pss.SetPeerPublicKey(crypto.ToECDSAPub(pubkey), topic, &addr)
	if err != nil {
		return fmt.Errorf("Invalid key: %x", pubkey)
	}
	return nil
}

func (pssapi *API) GetSymmetricAddressHint(topic whisper.TopicType, asymmetric bool, key string) (PssAddress, error) {
	return *pssapi.Pss.symKeyPool[key][topic].address, nil
}

func (pssapi *API) GetAsymmetricAddressHint(topic whisper.TopicType, asymmetric bool, key string) (PssAddress, error) {
	return *pssapi.Pss.pubKeyPool[key][topic].address, nil
}

func (pssapi *API) StringToTopic(topicstring string) (whisper.TopicType, error) {
	return StringToTopic(topicstring), nil
}
