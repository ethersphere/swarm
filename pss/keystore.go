// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/crypto"
	"github.com/ethersphere/swarm/pss/message"
)

type KeyStore struct {
	Crypto                   crypto.Crypto // key and encryption crypto
	mx                       sync.RWMutex
	pubKeyPool               map[string]map[message.Topic]*peer // mapping of hex public keys to peer address by topic.
	symKeyPool               map[string]map[message.Topic]*peer // mapping of symkeyids to peer address by topic.
	symKeyDecryptCache       []*string                          // fast lookup of symkeys recently used for decryption; last used is on top of stack
	symKeyDecryptCacheCursor int                                // modular cursor pointing to last used, wraps on symKeyDecryptCache array
}

func loadKeyStore() *KeyStore {
	return &KeyStore{
		Crypto:             crypto.New(),
		pubKeyPool:         make(map[string]map[message.Topic]*peer),
		symKeyPool:         make(map[string]map[message.Topic]*peer),
		symKeyDecryptCache: make([]*string, defaultSymKeyCacheCapacity),
	}
}

func (ks *KeyStore) isSymKeyStored(key string) bool {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	var ok bool
	_, ok = ks.symKeyPool[key]
	return ok
}

func (ks *KeyStore) isPubKeyStored(key string) bool {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	var ok bool
	_, ok = ks.pubKeyPool[key]
	return ok
}

func (ks *KeyStore) getPeerSym(symkeyid string, topic message.Topic) (*peer, bool) {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	psp, ok := ks.symKeyPool[symkeyid][topic]
	return psp, ok
}

func (ks *KeyStore) getPeerPub(pubkeyid string, topic message.Topic) (*peer, bool) {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	psp, ok := ks.pubKeyPool[pubkeyid][topic]
	return psp, ok
}

// Links a peer ECDSA public key to a topic.
// This is required for asymmetric message exchange on the given topic.
// The value in `address` will be used as a routing hint for the public key / topic association.
func (ks *KeyStore) SetPeerPublicKey(pubkey *ecdsa.PublicKey, topic message.Topic, address PssAddress) error {
	if err := validateAddress(address); err != nil {
		return err
	}
	pubkeybytes := ks.Crypto.SerializePublicKey(pubkey)
	if len(pubkeybytes) == 0 {
		return fmt.Errorf("invalid public key: %v", pubkey)
	}
	pubkeyid := common.ToHex(pubkeybytes)
	psp := &peer{
		address: address,
	}
	ks.mx.Lock()
	if _, ok := ks.pubKeyPool[pubkeyid]; !ok {
		ks.pubKeyPool[pubkeyid] = make(map[message.Topic]*peer)
	}
	ks.pubKeyPool[pubkeyid][topic] = psp
	ks.mx.Unlock()
	log.Trace("added pubkey", "pubkeyid", pubkeyid, "topic", topic, "address", address)
	return nil
}

// adds a symmetric key to the pss key pool, and optionally adds the key to the
// collection of keys used to attempt symmetric decryption of incoming messages
func (ks *KeyStore) addSymmetricKeyToPool(keyid string, topic message.Topic, address PssAddress, addtocache bool, protected bool) {
	psp := &peer{
		address:   address,
		protected: protected,
	}
	ks.mx.Lock()
	if _, ok := ks.symKeyPool[keyid]; !ok {
		ks.symKeyPool[keyid] = make(map[message.Topic]*peer)
	}
	ks.symKeyPool[keyid][topic] = psp
	ks.mx.Unlock()
	if addtocache {
		ks.symKeyDecryptCacheCursor++
		ks.symKeyDecryptCache[ks.symKeyDecryptCacheCursor%cap(ks.symKeyDecryptCache)] = &keyid
	}
}

// Returns all recorded topic and address combination for a specific public key
func (ks *KeyStore) GetPublickeyPeers(keyid string) (topic []message.Topic, address []PssAddress, err error) {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	for t, peer := range ks.pubKeyPool[keyid] {
		topic = append(topic, t)
		address = append(address, peer.address)
	}
	return topic, address, nil
}

func (ks *KeyStore) getPeerAddress(keyid string, topic message.Topic) (PssAddress, error) {
	ks.mx.RLock()
	defer ks.mx.RUnlock()
	if peers, ok := ks.pubKeyPool[keyid]; ok {
		if t, ok := peers[topic]; ok {
			return t.address, nil
		}
	}
	return nil, fmt.Errorf("peer with pubkey %s, topic %x not found", keyid, topic)
}

// Attempt to decrypt, validate and unpack a symmetrically encrypted message.
// If successful, returns the payload of the message and the id
// of the symmetric key used to decrypt the message.
// It fails if decryption of the message fails or if the message is corrupted/not valid.
func (ks *KeyStore) processSym(pssMsg *message.Message) ([]byte, string, PssAddress, error) {
	metrics.GetOrRegisterCounter("pss.process.sym", nil).Inc(1)

	for i := ks.symKeyDecryptCacheCursor; i > ks.symKeyDecryptCacheCursor-cap(ks.symKeyDecryptCache) && i > 0; i-- {
		symkeyid := ks.symKeyDecryptCache[i%cap(ks.symKeyDecryptCache)]
		symkey, err := ks.Crypto.GetSymKey(*symkeyid)
		if err != nil {
			continue
		}
		unwrapParams := &crypto.UnwrapParams{
			SymmetricKey: symkey,
		}
		recvmsg, err := ks.Crypto.UnWrap(pssMsg.Payload, unwrapParams)
		if err != nil {
			continue
		}
		payload, validateError := recvmsg.GetPayload()
		if validateError != nil {
			return nil, "", nil, validateError
		}

		var from PssAddress
		ks.mx.RLock()
		if ks.symKeyPool[*symkeyid][pssMsg.Topic] != nil {
			from = ks.symKeyPool[*symkeyid][pssMsg.Topic].address
		}
		ks.mx.RUnlock()
		ks.symKeyDecryptCacheCursor++
		ks.symKeyDecryptCache[ks.symKeyDecryptCacheCursor%cap(ks.symKeyDecryptCache)] = symkeyid
		return payload, *symkeyid, from, nil
	}
	return nil, "", nil, errors.New("could not decrypt message")
}

// Attempt to decrypt, validate and unpack an asymmetrically encrypted message.
// If successful, returns the payload of the message and the hex representation of
// the public key used to decrypt the message.
// It fails if decryption of message fails, or if the message is corrupted.
func (p *Pss) processAsym(pssMsg *message.Message) ([]byte, string, PssAddress, error) {
	metrics.GetOrRegisterCounter("pss.process.asym", nil).Inc(1)

	unwrapParams := &crypto.UnwrapParams{
		Receiver: p.privateKey,
	}
	recvmsg, err := p.Crypto.UnWrap(pssMsg.Payload, unwrapParams)
	if err != nil {
		return nil, "", nil, fmt.Errorf("could not decrypt message: %s", err)
	}

	payload, validateError := recvmsg.GetPayload()
	if validateError != nil {
		return nil, "", nil, validateError
	}

	pubkeyid := common.ToHex(p.Crypto.SerializePublicKey(recvmsg.GetSender()))
	var from PssAddress
	p.mx.RLock()
	if p.pubKeyPool[pubkeyid][pssMsg.Topic] != nil {
		from = p.pubKeyPool[pubkeyid][pssMsg.Topic].address
	}
	p.mx.RUnlock()
	return payload, pubkeyid, from, nil
}

// Symkey garbage collection
// a key is removed if:
// - it is not marked as protected
// - it is not in the incoming decryption cache
func (p *Pss) cleanKeys() (count int) {
	p.mx.Lock()
	defer p.mx.Unlock()
	for keyid, peertopics := range p.symKeyPool {
		var expiredtopics []message.Topic
		for topic, psp := range peertopics {
			if psp.protected {
				continue
			}

			var match bool
			for i := p.symKeyDecryptCacheCursor; i > p.symKeyDecryptCacheCursor-cap(p.symKeyDecryptCache) && i > 0; i-- {
				cacheid := p.symKeyDecryptCache[i%cap(p.symKeyDecryptCache)]
				if *cacheid == keyid {
					match = true
				}
			}
			if !match {
				expiredtopics = append(expiredtopics, topic)
			}
		}
		for _, topic := range expiredtopics {
			delete(p.symKeyPool[keyid], topic)
			log.Trace("symkey cleanup deletion", "symkeyid", keyid, "topic", topic, "val", p.symKeyPool[keyid])
			count++
		}
	}
	return count
}

// Automatically generate a new symkey for a topic and address hint
func (ks *KeyStore) GenerateSymmetricKey(topic message.Topic, address PssAddress, addToCache bool) (string, error) {
	keyid, err := ks.Crypto.GenerateSymKey()
	if err == nil {
		ks.addSymmetricKeyToPool(keyid, topic, address, addToCache, false)
	}
	return keyid, err
}

// Returns a symmetric key byte sequence stored in the crypto backend by its unique id.
// Passes on the error value from the crypto backend.
func (ks *KeyStore) GetSymmetricKey(symkeyid string) ([]byte, error) {
	return ks.Crypto.GetSymKey(symkeyid)
}

// Links a peer symmetric key (arbitrary byte sequence) to a topic.
//
// This is required for symmetrically encrypted message exchange on the given topic.
//
// The key is stored in the crypto backend.
//
// If addtocache is set to true, the key will be added to the cache of keys
// used to attempt symmetric decryption of incoming messages.
//
// Returns a string id that can be used to retrieve the key bytes
// from the crypto backend (see pss.GetSymmetricKey())
func (ks *KeyStore) SetSymmetricKey(key []byte, topic message.Topic, address PssAddress, addtocache bool) (string, error) {
	if err := validateAddress(address); err != nil {
		return "", err
	}
	return ks.setSymmetricKey(key, topic, address, addtocache, true)
}

func (ks *KeyStore) setSymmetricKey(key []byte, topic message.Topic, address PssAddress, addtocache bool, protected bool) (string, error) {
	keyid, err := ks.Crypto.AddSymKey(key)
	if err == nil {
		ks.addSymmetricKeyToPool(keyid, topic, address, addtocache, protected)
	}
	return keyid, err
}
