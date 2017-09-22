package pss

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
	"time"
)

const (
	HANDSHAKE_NONE = iota
	HANDSHAKE_OK
	HANDSHAKE_PEND
	HANDSHAKE_SUSPEND
)

const (
	defaultSymKeyRequestExpiry     = 1000 * 8               // max wait ms to receive a response to a handshake symkey request
	defaultSymKeySendLimit         = 32                     // amount of messages a symkey is valid for
	defaultSymKeyFloodThreshold    = 1000 * 1               // ms between handshake requests to avoid flood counter increment
	defaultSymKeyFloodLimit        = 2                      // max number of messages in too quick succession before suspension
	defaultSymKeySuspendDuration   = 1000 * 32              // ms suspend lasts
	defaultSymKeyFloodBanThreshold = 2                      // max number of suspends before permanent ban
	defaultSymKeyMinLength         = 32                     // minimum accepted length of symkey
	defaultSymKeyMaxLength         = defaultSymKeyMinLength // maximum accepted length of symkey
	defaultSymKeyCapacity          = 4                      // max number of symkeys to store/send simultaneously
	defaultSymKeyLength            = defaultSymKeyMaxLength
)

type handshakeMsg struct {
	From      []byte
	Limit     uint16
	Keys      [][]byte
	KeyLength uint8
	Request   uint8
	Topic     whisper.TopicType
}

type handshakeKey struct {
	id        *string
	limit     uint8
	count     uint8
	expiredAt time.Time
}

type handshake struct {
	pubKeyId *string
	topic    *whisper.TopicType
	outKeys  []handshakeKey
	inKeys   []handshakeKey
}

type handshakeGuard struct {
	lastRequest time.Time
	status      uint8
	strikes     uint8
}

type HandshakeParams struct {
	SymKeyRequestExpiry time.Duration
	SymKeySendLimit     uint16
	SymKeyMinLength     uint8
	SymKeyMaxLength     uint8
	SymKeyCapacity      uint16
	SymKeyLength        uint8
}

func NewHandshakeParams() *HandshakeParams {
	return &HandshakeParams{
		SymKeyRequestExpiry: defaultSymKeyRequestExpiry,
		SymKeySendLimit:     defaultSymKeySendLimit,
		SymKeyMinLength:     defaultSymKeyMinLength,
		SymKeyMaxLength:     defaultSymKeyMaxLength,
		SymKeyCapacity:      defaultSymKeyCapacity,
		SymKeyLength:        defaultSymKeyLength,
	}
}

type HandshakeController struct {
	pss                 *Pss
	keyC                map[string]chan []string // adds a channel to report when a handshake succeeds
	symKeyRequestExpiry time.Duration
	symKeySendLimit     uint16
	symKeyMinLength     uint8
	symKeyMaxLength     uint8
	symKeyCapacity      uint16
	symKeyLength        uint8
	handshakes          map[string]map[whisper.TopicType]*handshake
}

func NewHandshakeController(pss *Pss, params *HandshakeParams) *HandshakeController {
	ctrl := &HandshakeController{
		pss:                 pss,
		keyC:                make(map[string]chan []string),
		symKeyRequestExpiry: params.SymKeyRequestExpiry,
		symKeySendLimit:     params.SymKeySendLimit,
		symKeyMinLength:     params.SymKeyMinLength,
		symKeyMaxLength:     params.SymKeyMaxLength,
		symKeyLength:        params.SymKeyLength,
		symKeyCapacity:      params.SymKeyCapacity,
	}
	topic := whisper.BytesToTopic([]byte{})
	pss.Register(&topic, ctrl.handler)
	return ctrl
}

func (self *HandshakeController) validKeys(pubkeyid string, topic *whisper.TopicType, in bool) (validkeys []*string) {
	if _, ok := self.handshakes[pubkeyid]; !ok {
		return []*string{}
	}
	var keys []handshakeKey
	if in {
		keys = self.handshakes[pubkeyid][*topic].inKeys
	} else {
		keys = self.handshakes[pubkeyid][*topic].outKeys
	}

	for _, key := range keys {
		if key.limit <= key.count {
			self.releaseKey(pubkeyid, &key, topic, in)
		} else if !key.expiredAt.IsZero() && time.Now().After(key.expiredAt) {
			self.releaseKey(pubkeyid, &key, topic, in)
		}
		validkeys = append(validkeys, key.id)
	}
	return
}

func (self *HandshakeController) releaseKey(pubkeyid string, removekey *handshakeKey, topic *whisper.TopicType, in bool) bool {
	if _, ok := self.handshakes[pubkeyid]; !ok {
		return false
	}
	var keys []handshakeKey
	if in {
		keys = self.handshakes[pubkeyid][*topic].inKeys
	} else {
		keys = self.handshakes[pubkeyid][*topic].outKeys
	}

	var match bool
	for i, key := range keys {
		if removekey.id == key.id {
			self.pss.symKeyPool[*key.id][*topic].protected = false
			match = true
			keys[i] = keys[len(keys)-1]
		}
	}
	if !match {
		return false
	}
	keys = keys[:len(keys)-1]
	return true
}

func (self *HandshakeController) handler(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	if !asymmetric {
		return errors.New("symmetric handshake")

	}
	keymsg := &handshakeMsg{}
	err := rlp.DecodeBytes(msg, keymsg)
	if err == nil {
		err := self.handleKeys(keyid, keymsg)
		if err != nil {
			log.Error("handlekeys fail", "error", err)
		}
		return err
	}
	return nil
}

// TODO:
// - flood guard
// - keylength check
func (self *HandshakeController) handleKeys(pubkeyid string, keymsg *handshakeMsg) error {

	// new keys from peer
	if len(keymsg.Keys) > 0 {
		log.Trace("received handshake keys", "pubkeyid", pubkeyid, "from", keymsg.From, "count", len(keymsg.Keys))
		//		if _, ok := self.pubKeySymKeyIndex[pubkeyid]; !ok {
		//			self.pubKeySymKeyIndex[pubkeyid] = make(map[whisper.TopicType][]*string)
		//		}

		var sendsymkeyids []string
		for _, key := range keymsg.Keys {
			sendsymkey := make([]byte, keymsg.KeyLength)
			copy(sendsymkey, key)
			var address PssAddress
			copy(address[:], keymsg.From)
			sendsymkeyid, err := self.pss.SetSymmetricKey(sendsymkey, keymsg.Topic, &address, keymsg.Limit, false)
			if err != nil {
				return err
			}
			//self.pubKeySymKeyIndex[pubkeyid][envelope.Topic] = append(self.pubKeySymKeyIndex[pubkeyid][envelope.Topic], &sendsymkeyid)
			sendsymkeyids = append(sendsymkeyids, sendsymkeyid)
		}
		if len(sendsymkeyids) > 0 {
			self.alertHandshake(pubkeyid, sendsymkeyids)
		}
	}

	// peer request for keys
	if keymsg.Request > 0 {
		log.Trace("sending handshake keys", "pubkeyid", pubkeyid, "from", keymsg.From, "count", keymsg.Request)
		_, err := self.sendKey(pubkeyid, &keymsg.Topic, keymsg.Request, self.symKeySendLimit, keymsg.From)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *HandshakeController) sendKey(pubkeyid string, topic *whisper.TopicType, keycount uint8, msglimit uint16, to PssAddress) ([]string, error) {
	recvkeys := make([][]byte, keycount)
	recvkeyids := make([]string, keycount)

	if _, ok := self.handshakes[pubkeyid]; !ok {
		self.handshakes[pubkeyid] = make(map[whisper.TopicType]*handshake)
	}

	// check if buffer is not full
	inkeys := self.validKeys(pubkeyid, topic, true)
	requestcount := uint8(self.symKeyCapacity - uint16(len(inkeys)))

	// return if there's nothing to be accomplished
	if requestcount == 0 && len(inkeys) == 0 && keycount == 0 {
		return []string{}, nil
	}

	// generate new keys to send
	for i := 0; i < len(recvkeyids); i++ {
		var err error
		recvkeyids[i], err = self.pss.generateSymmetricKey(*topic, &to, msglimit, true)
		if err != nil {
			return []string{}, fmt.Errorf("set receive symkey fail (addr %x pubkey %x topic %x): %v", to, pubkeyid, topic, err)
		}
		recvkeys[i], err = self.pss.GetSymmetricKey(recvkeyids[i])
		if err != nil {
			return []string{}, fmt.Errorf("get generated outgoing symkey fail (addr %x pubkey %x topic %x): %v", to, pubkeyid, topic, err)
		}
		//self.symKeyPubKeyIndex[recvkeyids[i]] = pubkeyid
	}

	// encode and send the message
	recvkeymsg := &handshakeMsg{
		From:      self.pss.BaseAddr(),
		Keys:      recvkeys,
		KeyLength: self.symKeyLength,
		Request:   requestcount,
		Limit:     self.symKeySendLimit,
		Topic:     *topic,
	}
	log.Trace("sending our symkeys", "pubkey", pubkeyid, "symkeys", recvkeyids, "limit", self.symKeySendLimit, "requestcount", requestcount, "keycount", len(recvkeys))
	recvkeybytes, err := rlp.EncodeToBytes(recvkeymsg)
	if err != nil {
		return []string{}, fmt.Errorf("rlp keymsg encode fail: %v", err)
	}
	// if the send fails it means this public key is not registered for this particular address AND topic
	err = self.pss.SendAsym(pubkeyid, *topic, recvkeybytes)
	if err != nil {
		return []string{}, fmt.Errorf("Send symkey failed: %v", err)
	}
	return recvkeyids, nil
}

// used to enable blocked key requests to peer
// if passed without symkey a new keyid array channel is created with pubkeyid as key
// if passed with symkeyids, and the channel on the pubkeyid is active, symkeyids are passed to channel
func (self *HandshakeController) alertHandshake(pubkeyid string, symkeys []string) chan []string {
	if len(symkeys) > 0 {
		if _, ok := self.keyC[pubkeyid]; ok {
			self.keyC[pubkeyid] <- symkeys
			close(self.keyC[pubkeyid])
			delete(self.keyC, pubkeyid)
		}
		return nil
	} else {
		if _, ok := self.keyC[pubkeyid]; !ok {
			self.keyC[pubkeyid] = make(chan []string)
		}
	}
	return self.keyC[pubkeyid]
}

//
//// returns all symkeys that are active for respective public keys after handshake exchange
//func (self *Pss) getSymmetricKeyBuffer(pubkeyid string, topic *whisper.TopicType) (symkeyids []string, remaining []uint16) {
//	if _, ok := self.pubKeySymKeyIndex[pubkeyid]; !ok {
//		return
//	}
//	for _, symkeyid := range self.pubKeySymKeyIndex[pubkeyid][*topic] {
//		capacity, _ := self.GetSymmetricKeyCapacity(*symkeyid)
//		if capacity == 0 {
//			continue
//		}
//		symkeyids = append(symkeyids, *symkeyid)
//		remaining = append(remaining, capacity)
//	}
//	return
//}

// checks if symkey is valid for more messages.
// if not, the symkey will be instantly garbage collected.
//func (self *Pss) GetSymmetricKeyCapacity(symkeyid string) (uint16, error) {
//	if _, ok := self.symKeyPool[symkeyid]; !ok {
//		return 0, errors.New(fmt.Sprintf("Invalid symkeyid %s", symkeyid))
//	}
//	capacity := self.symKeyPool[symkeyid].sendLimit - self.symKeyPool[symkeyid].sendCount
//	if capacity == 0 {
//		delete(self.symKeyPool, symkeyid)
//	}
//	return capacity, nil
//}

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
//func (pssapi *API) Handshake(pubkeyid string, topic whisper.TopicType, to PssAddress, sync bool) ([]string, error) {
//	var err error
//	var hsc chan []string
//	var keys []string
//	_, counts := pssapi.Pss.getSymmetricKeyBuffer(pubkeyid, &topic)
//	requestcount := uint8(pssapi.Pss.symKeyBufferCapacity - len(counts))
//	if requestcount == 0 {
//		return keys, errors.New("Symkey buffer is full")
//	}
//	if sync {
//		hsc = pssapi.Pss.alertHandshake(pubkeyid, []string{})
//	}
//	_, err = pssapi.sendKey(pubkeyid, &topic, 0, pssapi.Pss.symKeySendLimit, to)
//	if err != nil {
//		return []string{}, err
//	}
//	if sync {
//		ctx, _ := context.WithTimeout(context.Background(), pssapi.Pss.symKeyRequestExpiry)
//		select {
//		case keys = <-hsc:
//			log.Trace("sync handshake response receive", "key", keys)
//		case <-ctx.Done():
//			return []string{}, errors.New("timeout")
//		}
//	}
//	return keys, nil
//}

// Get all outgoing symkeys valid for a particular pubkey and topic
//func (pssapi *API) GetSymmetricKeys(pubkeyid string, topic whisper.TopicType) ([]string, error) {
//	keys, _ := pssapi.Pss.getSymmetricKeyBuffer(pubkeyid, &topic)
//	return keys, nil
//}
