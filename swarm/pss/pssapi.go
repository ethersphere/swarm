package pss

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

// Convenience wrapper for sending and receiving pss messages when using the pss API
type APIMsg struct {
	Msg        []byte
	Asymmetric bool
	Key        string
}

// for debugging, show nice hex version
func (self *APIMsg) String() string {
	return fmt.Sprintf("APIMsg: asym: %v, key: %s", self.Asymmetric, self.Key)
}

// Pss API services
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
// All incoming messages to the node matching this topic will be encapsulated in the APIMsg struct and sent to the subscriber
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

// Sends the message wrapped in APIMsg through pss using symmetric encryption
//
// The method will pass on the error received from pss. It will fail if no public key for the Pss peer has been added
// The addresslength parameter decides how many bytes of the address to reveal it transit. -1 equals showing all. 0 means show none.
//func (pssapi *API) SendAsym(key string, topic whisper.TopicType, msg []byte) error {
//	return pssapi.Pss.SendAsym(key, topic, msg)
//}

// BaseAddr returns the local swarm overlay address of the Pss node
//
// Note that the overlay address is NOT inferable. To really know the node's overlay address it must reveal it itself.
func (pssapi *API) BaseAddr() ([]byte, error) {
	return pssapi.Pss.BaseAddr(), nil
}

// Retrieves the node's public key in byte form
func (pssapi *API) GetPublicKey() []byte {
	key := pssapi.Pss.PublicKey()
	return crypto.FromECDSAPub(&key)
}

// Set Public key to associate with a particular Pss peer
func (pssapi *API) SetPeerPublicKey(pubkey []byte, topic whisper.TopicType, addr PssAddress) error {
	pssapi.Pss.SetPeerPublicKey(crypto.ToECDSAPub(pubkey), topic, addr)
	return nil
}

// Get address hint for topic and key combination
func (pssapi *API) GetAddress(topic whisper.TopicType, asymmetric bool, key string) (PssAddress, error) {
	if asymmetric {
		return pssapi.Pss.pubKeyPool[key][topic].address, nil
	} else {
		return pssapi.Pss.symKeyPool[key].address, nil
	}
}

// Initiate a handshake with a peer for a specific topic.
//
// Will request new symkeys from the peer to fill up the buffer allowance of outgoing symkeys,
// If the peer's buffer also isn't full, it will respond with keys and request keys from the node, which will be automatically sent.
//
// The symkeys will be stored with the address hint specified in the "to" parameter
//
// If the "sync" parameter is set, the call will block until they are received or the request times out.
//
// will fail if buffer is already full, if handshake request cannot be sent, or by timeout if "sync" is set
func (pssapi *API) Handshake(pubkeyid string, topic whisper.TopicType, to PssAddress, sync bool) ([]string, error) {
	var err error
	var hsc chan []string
	var keys []string
	_, counts := pssapi.Pss.getSymmetricKeyBuffer(pubkeyid, &topic)
	requestcount := uint8(pssapi.Pss.symKeyBufferCapacity - len(counts))
	if requestcount == 0 {
		return keys, errors.New("Symkey buffer is full")
	}
	if sync {
		hsc = pssapi.Pss.alertHandshake(pubkeyid, []string{})
	}
	_, err = pssapi.sendKey(pubkeyid, &topic, 0, pssapi.Pss.symKeySendLimit, to)
	if err != nil {
		return []string{}, err
	}
	if sync {
		ctx, _ := context.WithTimeout(context.Background(), pssapi.Pss.symKeyRequestExpiry)
		select {
		case keys = <-hsc:
			log.Trace("sync handshake response receive", "key", keys)
		case <-ctx.Done():
			return []string{}, errors.New("timeout")
		}
	}
	return keys, nil
}

// Get all outgoing symkeys valid for a particular pubkey and topic
func (pssapi *API) GetSymmetricKeys(pubkeyid string, topic whisper.TopicType) ([]string, error) {
	keys, _ := pssapi.Pss.getSymmetricKeyBuffer(pubkeyid, &topic)
	return keys, nil
}

func (pssapi *API) GetPublicKeyFromSymmetricKey(symkeyid string) (string, error) {
	return pssapi.Pss.symKeyPubKeyIndex[symkeyid], nil
}

// PssAPITest are temporary API calls for development use only
type APITest struct {
	*Pss
}

func NewAPITest(ps *Pss) *APITest {
	return &APITest{Pss: ps}
}

// force expiry of a symkey
func (self *APITest) DepleteSymKey(symkeyid string) error {
	if _, ok := self.Pss.symKeyPool[symkeyid]; !ok {
		return errors.New(fmt.Sprintf("invalid symkey %s", symkeyid))
	}
	self.Pss.symKeyPool[symkeyid].sendCount = self.Pss.symKeyPool[symkeyid].sendLimit
	return nil
}

// get all valid in- and outgoing symkeys for a pubkey and topic
func (self *APITest) DumpSymKeys(pubkeyid string, topic whisper.TopicType) (ids []string, err error) {
	for id, psp := range self.Pss.symKeyPool {
		if psp.sendLimit > psp.sendCount {
			ids = append(ids, id)
		}
	}
	return
}

// manually set in- and outgoing pair of symkeys
// mimics state after symkey exchange handshake
func (self *APITest) SetSymKeys(pubkeyid string, recvkey []byte, sendkey []byte, sendlimit uint16, topic whisper.TopicType, addr PssAddress) (keyids [2]string, err error) {
	keyids[0], err = self.w.AddSymKeyDirect(recvkey)
	if err != nil {
		return [2]string{}, err
	}
	keyids[1], err = self.w.AddSymKeyDirect(sendkey)
	if err != nil {
		return [2]string{}, err
	}
	log.Debug("manual sumkey add", "recv", keyids[0], "send", keyids[1])
	self.Pss.symKeyPool[keyids[0]] = &pssPeer{}
	self.Pss.symKeyPool[keyids[0]].address = addr
	self.Pss.symKeyPool[keyids[1]] = &pssPeer{}
	self.Pss.symKeyPool[keyids[1]].address = addr
	self.Pss.symKeyPool[keyids[1]].sendLimit = sendlimit
	if _, ok := self.Pss.pubKeySymKeyIndex[pubkeyid]; !ok {
		self.Pss.pubKeySymKeyIndex[pubkeyid] = make(map[whisper.TopicType][]*string)
	}
	self.Pss.pubKeySymKeyIndex[pubkeyid][topic] = append(self.Pss.pubKeySymKeyIndex[pubkeyid][topic], &keyids[1])
	self.Pss.symKeyCacheCursor++
	self.Pss.symKeyCache[self.Pss.symKeyCacheCursor%len(self.Pss.symKeyCache)] = &keyids[0]
	return keyids, nil
}
