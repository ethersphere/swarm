package pss

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
	"sync"
	"time"
)

const (
	HANDSHAKE_NONE = iota
	HANDSHAKE_OK
	HANDSHAKE_PEND
	HANDSHAKE_SUSPEND
)

var (
	ctrlSingleton *HandshakeController
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
)

type handshakeMsg struct {
	From    []byte
	Limit   uint16
	Keys    [][]byte
	Request uint8
	Topic   whisper.TopicType
}

type handshakeKey struct {
	symKeyId  *string
	pubKeyId  *string
	limit     uint16
	count     uint16
	expiredAt time.Time
}

type handshake struct {
	outKeys []handshakeKey
	inKeys  []handshakeKey
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
	SymKeyCapacity      uint8
}

func NewHandshakeParams() *HandshakeParams {
	return &HandshakeParams{
		SymKeyRequestExpiry: defaultSymKeyRequestExpiry * time.Millisecond,
		SymKeySendLimit:     defaultSymKeySendLimit,
		SymKeyMinLength:     defaultSymKeyMinLength,
		SymKeyMaxLength:     defaultSymKeyMaxLength,
		SymKeyCapacity:      defaultSymKeyCapacity,
	}
}

type HandshakeController struct {
	pss                 *Pss
	keyC                map[string]chan []string // adds a channel to report when a handshake succeeds
	lock                sync.Mutex
	symKeyRequestExpiry time.Duration
	symKeySendLimit     uint16
	symKeyMinLength     uint8
	symKeyMaxLength     uint8
	symKeyCapacity      uint8
	symKeyIndex         map[string]*handshakeKey
	handshakes          map[string]map[whisper.TopicType]*handshake
	deregisterFuncs     map[whisper.TopicType]func()
}

func SetHandshakeController(pss *Pss, params *HandshakeParams) error {
	ctrl := &HandshakeController{
		pss:                 pss,
		keyC:                make(map[string]chan []string),
		symKeyRequestExpiry: params.SymKeyRequestExpiry,
		symKeySendLimit:     params.SymKeySendLimit,
		symKeyMinLength:     params.SymKeyMinLength,
		symKeyMaxLength:     params.SymKeyMaxLength,
		symKeyCapacity:      params.SymKeyCapacity,
		symKeyIndex:         make(map[string]*handshakeKey),
		handshakes:          make(map[string]map[whisper.TopicType]*handshake),
		deregisterFuncs:     make(map[whisper.TopicType]func()),
	}
	api := &HandshakeAPI{
		namespace: "pss",
		ctrl:      ctrl,
	}
	pss.addAPI(rpc.API{
		Namespace: api.namespace,
		Version:   "0.2",
		Service:   api,
		Public:    true,
	})
	ctrlSingleton = ctrl
	return nil
}

func (self *HandshakeController) validKeys(pubkeyid string, topic *whisper.TopicType, in bool) (validkeys []*string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.handshakes[pubkeyid]; !ok {
		return []*string{}
	} else if _, ok := self.handshakes[pubkeyid][*topic]; !ok {
		return []*string{}
	}
	var keystore *[]handshakeKey
	if in {
		keystore = &(self.handshakes[pubkeyid][*topic].inKeys)
	} else {
		keystore = &(self.handshakes[pubkeyid][*topic].outKeys)
	}

	for _, key := range *keystore {
		if key.limit <= key.count {
			self.releaseKey(*key.symKeyId, topic)
		} else if !key.expiredAt.IsZero() && time.Now().After(key.expiredAt) {
			self.releaseKey(*key.symKeyId, topic)
		} else {
			validkeys = append(validkeys, key.symKeyId)
		}
	}
	return
}

func (self *HandshakeController) updateKeys(pubkeyid string, topic *whisper.TopicType, in bool, symkeyids []string, limit uint16) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.handshakes[pubkeyid]; !ok {
		self.handshakes[pubkeyid] = make(map[whisper.TopicType]*handshake)

	}
	if self.handshakes[pubkeyid][*topic] == nil {
		self.handshakes[pubkeyid][*topic] = &handshake{}
	}
	var keystore *[]handshakeKey
	if in {
		keystore = &(self.handshakes[pubkeyid][*topic].inKeys)
	} else {
		keystore = &(self.handshakes[pubkeyid][*topic].outKeys)
	}
	for _, storekey := range *keystore {
		storekey.expiredAt = time.Now()
	}
	for i := 0; i < len(symkeyids); i++ {
		storekey := handshakeKey{
			symKeyId: &symkeyids[i],
			pubKeyId: &pubkeyid,
			limit:    limit,
		}
		*keystore = append(*keystore, storekey)
	}
	for i := 0; i < len(*keystore); i++ {
		self.symKeyIndex[*(*keystore)[i].symKeyId] = &((*keystore)[i])
	}
}

func (self *HandshakeController) releaseKey(symkeyid string, topic *whisper.TopicType) bool {
	if self.symKeyIndex[symkeyid] == nil {
		return false
	}
	self.pss.symKeyPool[symkeyid][*topic].protected = false
	self.symKeyIndex[symkeyid].expiredAt = time.Now()
	return true
}

func (self *HandshakeController) cleanHandshake(pubkeyid string, topic *whisper.TopicType, in bool, out bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
}

func (self *HandshakeController) clean() {
	peerpubkeys := self.handshakes
	now := time.Now()
	for pubkeyid, peertopics := range peerpubkeys {
		for topic, handshake := range peertopics {
			var keepcount int
			var deletes []string
			log.Debug("handshake clean", "pubkey", pubkeyid, "topic", topic)
			self.lock.Lock()
			for i, key := range handshake.inKeys {
				if key.expiredAt.Before(now) || key.limit <= key.count {
					deletes = append(deletes, *key.symKeyId)
					handshake.inKeys[keepcount] = handshake.inKeys[i]
					keepcount++
				}
			}
			handshake.inKeys = handshake.inKeys[:keepcount]
			keepcount = 0
			for i, key := range handshake.outKeys {
				if key.expiredAt.Before(now) || key.limit <= key.count {
					deletes = append(deletes, *key.symKeyId)
					handshake.outKeys[keepcount] = handshake.outKeys[i]
					keepcount++
				}
			}
			handshake.outKeys = handshake.outKeys[:keepcount]
			for _, keyid := range deletes {
				delete(self.symKeyIndex, keyid)
			}
			self.lock.Unlock()
		}
	}
	//	if _, ok := self.handshakes[pubkeyid]; !ok {
	//		return false
	//	}
	//	var keys *[]handshakeKey
	//	if in {
	//		keys = &self.handshakes[pubkeyid][*topic].inKeys
	//	} else {
	//		keys = &self.handshakes[pubkeyid][*topic].outKeys
	//	}
	//	var match bool
	//	for i, key := range *keys {
	//		if *symkeyid == *key.symKeyId {
	//			self.pss.symKeyPool[*key.symKeyId][*topic].protected = false
	//			match = true
	//			(*keys)[i] = (*keys)[len(*keys)-1]
	//			delete(self.symKeyIndex, *key.symKeyId)
	//		}
	//	}
	//	if !match {
	//		return false
	//	}
	//	(*keys) = (*keys)[:len(*keys)-1]
	//	return true
}

func (self *HandshakeController) handler(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	if !asymmetric {
		if self.symKeyIndex[keyid] != nil {
			self.symKeyIndex[keyid].count++
			log.Debug("tick", "symkeyid", keyid, "count", self.symKeyIndex[keyid].count)
		}
		return nil
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
		log.Debug("received handshake keys", "pubkeyid", pubkeyid, "from", keymsg.From, "count", len(keymsg.Keys))
		var sendsymkeyids []string
		for _, key := range keymsg.Keys {
			sendsymkey := make([]byte, len(key))
			copy(sendsymkey, key)
			var address PssAddress
			copy(address[:], keymsg.From)
			sendsymkeyid, err := self.pss.SetSymmetricKey(sendsymkey, keymsg.Topic, &address, keymsg.Limit, false)
			if err != nil {
				return err
			}
			sendsymkeyids = append(sendsymkeyids, sendsymkeyid)
		}
		if len(sendsymkeyids) > 0 {
			self.updateKeys(pubkeyid, &keymsg.Topic, false, sendsymkeyids, keymsg.Limit)

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

	self.lock.Lock()
	if _, ok := self.handshakes[pubkeyid]; !ok {
		self.handshakes[pubkeyid] = make(map[whisper.TopicType]*handshake)
	}
	self.lock.Unlock()

	// check if buffer is not full
	outkeys := self.validKeys(pubkeyid, topic, false)
	requestcount := uint8(self.symKeyCapacity - uint8(len(outkeys)))

	// return if there's nothing to be accomplished
	if requestcount == 0 && len(outkeys) == 0 && keycount == 0 {
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
	}
	self.updateKeys(pubkeyid, topic, true, recvkeyids, self.symKeySendLimit)

	// encode and send the message
	recvkeymsg := &handshakeMsg{
		From:    self.pss.BaseAddr(),
		Keys:    recvkeys,
		Request: requestcount,
		Limit:   self.symKeySendLimit,
		Topic:   *topic,
	}
	log.Debug("sending our symkeys", "pubkey", pubkeyid, "symkeys", recvkeyids, "limit", self.symKeySendLimit, "requestcount", requestcount, "keycount", len(recvkeys))
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

type HandshakeAPI struct {
	namespace string
	ctrl      *HandshakeController
}

func (self *HandshakeAPI) Handshake(pubkeyid string, topic whisper.TopicType, to PssAddress, sync bool, flush bool) (keys []string, err error) {
	var hsc chan []string
	var keycount uint8
	if flush {
		keycount = self.ctrl.symKeyCapacity
	}
	validkeys := self.ctrl.validKeys(pubkeyid, &topic, false)
	requestcount := uint8(self.ctrl.symKeyCapacity - uint8(len(validkeys)))
	if requestcount == 0 {
		return keys, errors.New("Symkey buffer is full")
	}
	if sync {
		hsc = self.ctrl.alertHandshake(pubkeyid, []string{})
	}
	_, err = self.ctrl.sendKey(pubkeyid, &topic, keycount, self.ctrl.symKeySendLimit, to)
	if err != nil {
		return keys, err
	}
	if sync {
		ctx, _ := context.WithTimeout(context.Background(), self.ctrl.symKeyRequestExpiry)
		select {
		case keys = <-hsc:
			log.Trace("sync handshake response receive", "key", keys)
		case <-ctx.Done():
			return []string{}, errors.New("timeout")
		}
	}
	return keys, nil
}

func (self *HandshakeAPI) AddHandshake(topic *whisper.TopicType) error {
	self.ctrl.deregisterFuncs[*topic] = self.ctrl.pss.Register(topic, self.ctrl.handler)
	return nil
}

func (self *HandshakeAPI) RemoveHandshake(topic *whisper.TopicType) error {
	if _, ok := self.ctrl.deregisterFuncs[*topic]; ok {
		self.ctrl.deregisterFuncs[*topic]()
	}
	return nil
}

func (self *HandshakeAPI) GetHandshakeKeys(pubkeyid string, topic whisper.TopicType, in bool, out bool) (keys []string, err error) {
	if in {
		for _, inkey := range self.ctrl.validKeys(pubkeyid, &topic, true) {
			keys = append(keys, *inkey)
		}
	}
	if out {
		for _, outkey := range self.ctrl.validKeys(pubkeyid, &topic, false) {
			keys = append(keys, *outkey)
		}
	}
	return keys, nil
}

func (self *HandshakeAPI) GetHandshakeKeyCapacity(symkeyid string) (uint16, error) {
	storekey := self.ctrl.symKeyIndex[symkeyid]
	if storekey == nil {
		return 0, errors.New(fmt.Sprintf("invalid symkey id %s", symkeyid))
	}
	return storekey.limit - storekey.count, nil
}

func (self *HandshakeAPI) GetHandshakePublicKey(symkeyid string) (string, error) {
	storekey := self.ctrl.symKeyIndex[symkeyid]
	if storekey == nil {
		return "", errors.New(fmt.Sprintf("invalid symkey id %s", symkeyid))
	}
	return *storekey.pubKeyId, nil
}

func (self *HandshakeAPI) ReleaseHandshakeKey(pubkeyid string, topic whisper.TopicType, symkeyid string) (removed bool, err error) {
	removed = self.ctrl.releaseKey(pubkeyid, &topic)
	if !removed {
		removed = self.ctrl.releaseKey(pubkeyid, &topic)
	}
	return
}

func (self *HandshakeAPI) SendSym(symkeyid string, topic whisper.TopicType, msg []byte) (err error) {
	err = self.ctrl.pss.SendSym(symkeyid, topic, msg)
	if self.ctrl.symKeyIndex[symkeyid] != nil {
		self.ctrl.symKeyIndex[symkeyid].count++
	}
	return
}
