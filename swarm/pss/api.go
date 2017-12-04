package pss

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
)

// Wrapper for receiving pss messages when using the pss API
// providing access to sender of message
type APIMsg struct {
	Msg        hexutil.Bytes
	Asymmetric bool
	Key        string
}

// Additional public methods accessible through API for pss
type API struct {
	*Pss
	returntype int
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
func (pssapi *API) Receive(ctx context.Context, topicbytes hexutil.Bytes) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return nil, fmt.Errorf("Subscribe not supported")
	}

	psssub := notifier.CreateSubscription()

	handler := func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		apimsg := &APIMsg{
			Msg:        hexutil.Bytes(msg),
			Asymmetric: asymmetric,
			Key:        keyid,
		}
		if err := notifier.Notify(psssub.ID, apimsg); err != nil {
			log.Warn(fmt.Sprintf("notification on pss sub topic rpc (sub %v) msg %v failed!", psssub.ID, msg))
		}
		return nil
	}

	var topic Topic
	copy(topic[:], topicbytes)
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

func (pssapi *API) GetAddress(topicbytes hexutil.Bytes, asymmetric bool, key string) (hexutil.Bytes, error) {
	var topic Topic
	copy(topic[:], topicbytes)
	var addr *PssAddress
	if asymmetric {
		peer, ok := pssapi.Pss.pubKeyPool[key][topic]
		if !ok {
			return nil, fmt.Errorf("pubkey/topic pair %x/%x doesn't exist", key, topic)
		}
		addr = peer.address
	} else {
		peer, ok := pssapi.Pss.symKeyPool[key][topic]
		if !ok {
			return nil, fmt.Errorf("symkey/topic pair %x/%x doesn't exist", key, topic)
		}
		addr = peer.address

	}
	return hexutil.Bytes(*addr), nil
}

// Retrieves the node's base address in hex form
func (pssapi *API) BaseAddr() hexutil.Bytes {
	return hexutil.Bytes(pssapi.Pss.BaseAddr())
}

// Retrieves the node's public key in hex form
func (pssapi *API) GetPublicKey() (keybytes hexutil.Bytes) {
	key := pssapi.Pss.PublicKey()
	keybytes = crypto.FromECDSAPub(key)
	return hexutil.Bytes(keybytes)
}

// Set Public key to associate with a particular Pss peer
func (pssapi *API) SetPeerPublicKey(pubkey hexutil.Bytes, topicbytes hexutil.Bytes, addrbytes hexutil.Bytes) error {
	var topic Topic
	copy(topic[:], topicbytes)
	addr := make(PssAddress, len(addrbytes))
	copy(addr, addrbytes[:])
	var err = pssapi.Pss.SetPeerPublicKey(crypto.ToECDSAPub(pubkey), topic, &addr)
	if err != nil {
		return fmt.Errorf("Invalid key: %x", pubkey)
	}
	return nil
}

func (pssapi *API) GetSymmetricKey(symkeyid string) (hexutil.Bytes, error) {
	symkey, err := pssapi.Pss.GetSymmetricKey(symkeyid)
	return hexutil.Bytes(symkey), err
}

func (pssapi *API) GetSymmetricAddressHint(topicbytes hexutil.Bytes, symkeyid string) (hexutil.Bytes, error) {
	var topic Topic
	copy(topic[:], topicbytes)
	return hexutil.Bytes(*pssapi.Pss.symKeyPool[symkeyid][topic].address), nil
}

func (pssapi *API) GetAsymmetricAddressHint(topicbytes hexutil.Bytes, pubkeyid string) (hexutil.Bytes, error) {
	var topic Topic
	copy(topic[:], topicbytes)
	addr := pssapi.Pss.pubKeyPool[pubkeyid][topic].address
	return hexutil.Bytes((*addr)[:]), nil
}

func (pssapi *API) StringToTopic(topicstring string) (hexutil.Bytes, error) {
	topic := BytesToTopic([]byte(topicstring))
	return hexutil.Bytes(topic[:]), nil
}

func (pssapi *API) SendAsym(pubkeyhex string, topicbytes hexutil.Bytes, msg hexutil.Bytes) error {
	var topic Topic
	copy(topic[:], topicbytes)
	return pssapi.Pss.SendAsym(pubkeyhex, topic, msg[:])
}

func (pssapi *API) SendSym(symkeyhex string, topicbytes hexutil.Bytes, msg hexutil.Bytes) error {
	var topic Topic
	copy(topic[:], topicbytes)
	return pssapi.Pss.SendSym(symkeyhex, topic, msg[:])
}
