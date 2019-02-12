// Copyright 2018 The go-ethereum Authors
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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/swarm/log"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

// Links a peer ECDSA public key to a topic
//
// This is required for asymmetric message exchange
// on the given topic
//
// The value in `address` will be used as a routing hint for the
// public key / topic association
func (p *Pss) SetPeerPublicKey(pubkey *ecdsa.PublicKey, topic Topic, address PssAddress) error {
	if err := validateAddress(address); err != nil {
		return err
	}
	pubkeybytes := crypto.FromECDSAPub(pubkey)
	if len(pubkeybytes) == 0 {
		return fmt.Errorf("invalid public key: %v", pubkey)
	}
	pubkeyid := common.ToHex(pubkeybytes)
	psp := &pssPeer{
		address: address,
	}
	p.pubKeyPoolMu.Lock()
	if _, ok := p.pubKeyPool[pubkeyid]; !ok {
		p.pubKeyPool[pubkeyid] = make(map[Topic]*pssPeer)
	}
	p.pubKeyPool[pubkeyid][topic] = psp
	p.pubKeyPoolMu.Unlock()
	log.Trace("added pubkey", "pubkeyid", pubkeyid, "topic", topic, "address", address)
	return nil
}

// Automatically generate a new symkey for a topic and address hint
func (p *Pss) GenerateSymmetricKey(topic Topic, address PssAddress, addToCache bool) (string, error) {
	keyid, err := p.w.GenerateSymKey()
	if err != nil {
		return "", err
	}
	p.addSymmetricKeyToPool(keyid, topic, address, addToCache, false)
	return keyid, nil
}

// Links a peer symmetric key (arbitrary byte sequence) to a topic
//
// This is required for symmetrically encrypted message exchange
// on the given topic
//
// The key is stored in the whisper backend.
//
// If addtocache is set to true, the key will be added to the cache of keys
// used to attempt symmetric decryption of incoming messages.
//
// Returns a string id that can be used to retrieve the key bytes
// from the whisper backend (see pss.GetSymmetricKey())
func (p *Pss) SetSymmetricKey(key []byte, topic Topic, address PssAddress, addtocache bool) (string, error) {
	if err := validateAddress(address); err != nil {
		return "", err
	}
	return p.setSymmetricKey(key, topic, address, addtocache, true)
}

func (p *Pss) setSymmetricKey(key []byte, topic Topic, address PssAddress, addtocache bool, protected bool) (string, error) {
	keyid, err := p.w.AddSymKeyDirect(key)
	if err != nil {
		return "", err
	}
	p.addSymmetricKeyToPool(keyid, topic, address, addtocache, protected)
	return keyid, nil
}

// adds a symmetric key to the pss key pool, and optionally adds the key
// to the collection of keys used to attempt symmetric decryption of
// incoming messages
func (p *Pss) addSymmetricKeyToPool(keyid string, topic Topic, address PssAddress, addtocache bool, protected bool) {
	psp := &pssPeer{
		address:   address,
		protected: protected,
	}
	p.symKeyPoolMu.Lock()
	if _, ok := p.symKeyPool[keyid]; !ok {
		p.symKeyPool[keyid] = make(map[Topic]*pssPeer)
	}
	p.symKeyPool[keyid][topic] = psp
	p.symKeyPoolMu.Unlock()
	if addtocache {
		p.symKeyDecryptCacheCursor++
		p.symKeyDecryptCache[p.symKeyDecryptCacheCursor%cap(p.symKeyDecryptCache)] = &keyid
	}
	key, _ := p.GetSymmetricKey(keyid)
	log.Trace("added symkey", "symkeyid", keyid, "symkey", common.ToHex(key), "topic", topic, "address", address, "cache", addtocache)
}

// Returns a symmetric key byte seqyence stored in the whisper backend
// by its unique id
//
// Passes on the error value from the whisper backend
func (p *Pss) GetSymmetricKey(symkeyid string) ([]byte, error) {
	symkey, err := p.w.GetSymKey(symkeyid)
	if err != nil {
		return nil, err
	}
	return symkey, nil
}

// Returns all recorded topic and address combination for a specific public key
func (p *Pss) GetPublickeyPeers(keyid string) (topic []Topic, address []PssAddress, err error) {
	p.pubKeyPoolMu.RLock()
	defer p.pubKeyPoolMu.RUnlock()
	for t, peer := range p.pubKeyPool[keyid] {
		topic = append(topic, t)
		address = append(address, peer.address)
	}

	return topic, address, nil
}

func (p *Pss) getPeerAddress(keyid string, topic Topic) (PssAddress, error) {
	p.pubKeyPoolMu.RLock()
	defer p.pubKeyPoolMu.RUnlock()
	if peers, ok := p.pubKeyPool[keyid]; ok {
		if t, ok := peers[topic]; ok {
			return t.address, nil
		}
	}
	return nil, fmt.Errorf("peer with pubkey %s, topic %x not found", keyid, topic)
}

// Attempt to decrypt, validate and unpack a
// symmetrically encrypted message
// If successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the whisper backend id
// of the symmetric key used to decrypt the message.
// It fails if decryption of the message fails or if the message is corrupted
func (p *Pss) processSym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, PssAddress, error) {
	metrics.GetOrRegisterCounter("pss.process.sym", nil).Inc(1)

	for i := p.symKeyDecryptCacheCursor; i > p.symKeyDecryptCacheCursor-cap(p.symKeyDecryptCache) && i > 0; i-- {
		symkeyid := p.symKeyDecryptCache[i%cap(p.symKeyDecryptCache)]
		symkey, err := p.w.GetSymKey(*symkeyid)
		if err != nil {
			continue
		}
		recvmsg, err := envelope.OpenSymmetric(symkey)
		if err != nil {
			continue
		}
		if !recvmsg.ValidateAndParse() {
			return nil, "", nil, fmt.Errorf("symmetrically encrypted message has invalid signature or is corrupt")
		}
		p.symKeyPoolMu.Lock()
		from := p.symKeyPool[*symkeyid][Topic(envelope.Topic)].address
		p.symKeyPoolMu.Unlock()
		p.symKeyDecryptCacheCursor++
		p.symKeyDecryptCache[p.symKeyDecryptCacheCursor%cap(p.symKeyDecryptCache)] = symkeyid
		return recvmsg, *symkeyid, from, nil
	}
	return nil, "", nil, fmt.Errorf("could not decrypt message")
}

// Attempt to decrypt, validate and unpack an
// asymmetrically encrypted message
// If successful, returns the unpacked whisper ReceivedMessage struct
// encapsulating the decrypted message, and the byte representation of
// the public key used to decrypt the message.
// It fails if decryption of message fails, or if the message is corrupted
func (p *Pss) processAsym(envelope *whisper.Envelope) (*whisper.ReceivedMessage, string, PssAddress, error) {
	metrics.GetOrRegisterCounter("pss.process.asym", nil).Inc(1)

	recvmsg, err := envelope.OpenAsymmetric(p.privateKey)
	if err != nil {
		return nil, "", nil, fmt.Errorf("could not decrypt message: %s", err)
	}
	// check signature (if signed), strip padding
	if !recvmsg.ValidateAndParse() {
		return nil, "", nil, fmt.Errorf("invalid message")
	}
	pubkeyid := common.ToHex(crypto.FromECDSAPub(recvmsg.Src))
	var from PssAddress
	p.pubKeyPoolMu.Lock()
	if p.pubKeyPool[pubkeyid][Topic(envelope.Topic)] != nil {
		from = p.pubKeyPool[pubkeyid][Topic(envelope.Topic)].address
	}
	p.pubKeyPoolMu.Unlock()
	return recvmsg, pubkeyid, from, nil
}

// Symkey garbage collection
// a key is removed if:
// - it is not marked as protected
// - it is not in the incoming decryption cache
func (p *Pss) cleanKeys() (count int) {
	for keyid, peertopics := range p.symKeyPool {
		var expiredtopics []Topic
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
			p.symKeyPoolMu.Lock()
			delete(p.symKeyPool[keyid], topic)
			log.Trace("symkey cleanup deletion", "symkeyid", keyid, "topic", topic, "val", p.symKeyPool[keyid])
			p.symKeyPoolMu.Unlock()
			count++
		}
	}
	return
}
