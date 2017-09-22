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
	pssapi.Pss.SetPeerPublicKey(crypto.ToECDSAPub(pubkey), topic, &addr)
	return nil
}

// Get address hint for topic and key combination
func (pssapi *API) GetAddress(topic whisper.TopicType, asymmetric bool, key string) (PssAddress, error) {
	if asymmetric {
		return *pssapi.Pss.pubKeyPool[key][topic].address, nil
	} else {
		return *pssapi.Pss.symKeyPool[key][topic].address, nil
	}
}

// PssAPITest are temporary API calls for development use only
type APITest struct {
	*Pss
}

func NewAPITest(ps *Pss) *APITest {
	return &APITest{Pss: ps}
}

func (apitest *APITest) SetSymKeys(pubkeyid string, recvsymkey []byte, sendsymkey []byte, limit uint16, topic whisper.TopicType, to []byte) ([2]string, error) {
	addr := make(PssAddress, 32)
	copy(addr[:], to)
	recvsymkeyid, err := apitest.SetSymmetricKey(recvsymkey, topic, &addr, limit, true)
	if err != nil {
		return [2]string{}, err
	}
	sendsymkeyid, err := apitest.SetSymmetricKey(sendsymkey, topic, &addr, limit, false)
	if err != nil {
		return [2]string{}, err
	}
	return [2]string{recvsymkeyid, sendsymkeyid}, nil
}
