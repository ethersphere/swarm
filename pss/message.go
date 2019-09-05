// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

// Contains all types and code related to messages and envelopes.
// Currently backed by whisperv6
package pss

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethersphere/swarm/log"
	"strconv"

	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

func toTopic(b []byte) (t Topic) {
	return Topic(whisper.BytesToTopic(b))
}

type rawMessage []byte

const (
	flagsLength             = 1
	payloadSizeFieldMaxSize = 4
	signatureLength         = 65      // in bytes
	padSizeLimit            = 256     // just an arbitrary number, could be changed without breaking the protocol
	SizeMask                = byte(3) // mask used to extract the size of payload size field from the flags
	signatureFlag           = byte(4)
	aesKeyLength            = 32 // in bytes
)

// == envelope ==

type envelope struct {
	Topic Topic
	Data  []byte
}

// newSentEnvelope creates and initializes a non-signed, non-encrypted Whisper message
// and then wrap it and encrypt it. It performs what it used to be two function calls:
// msg, e1 := NewSentMessage and env, e := msg.Wrap
func newSentEnvelope(params *messageParams, crypto CryptoBackend) (env *envelope, e error) {
	//whisperParams := toWhisperParams(params)

	var rawBytes rawMessage
	rawBytes = make([]byte, 1,
		flagsLength+payloadSizeFieldMaxSize+len(params.Payload)+len(params.Padding)+signatureLength+padSizeLimit)
	rawBytes[0] = 0 // set all the flags to zero
	// Bytes is flags+PayloadSize+Payload+padding+signature
	rawBytes = addPayloadSizeField(rawBytes, params.Payload)
	rawBytes = append(rawBytes, params.Payload...)
	rawBytes, e = appendPadding(rawBytes, params, crypto)
	if e != nil {
		return nil, e
	}

	if params.Src != nil {
		if rawBytes, e = sign(rawBytes, params.Src, crypto); e != nil {
			return nil, e
		}
	}
	if params.Dst != nil {
		rawBytes, e = crypto.EncryptAsymmetric(rawBytes, params.Dst)
	} else if params.KeySym != nil {
		rawBytes, e = crypto.EncryptSymmetric(rawBytes, params.KeySym)
	} else {
		e = errors.New("unable to encrypt the message: neither symmetric nor assymmetric key provided")
	}
	if e != nil {
		return nil, e
	}

	// the envelope, once the Nonce and PoW is removed only calculates Expiry
	env = newEnvelope(params.Topic, rawBytes)
	// Removed Seal that was only adding a PoW!!!
	return env, nil
}

// NewEnvelope wraps a Whisper message with expiration and destination data
// included into an envelope for network forwarding.
func newEnvelope(topic Topic, rawBytes rawMessage) *envelope {
	return &envelope{
		Topic: topic,
		Data:  rawBytes,
	}
}

func addPayloadSizeField(rawBytes rawMessage, payload []byte) rawMessage {
	fieldSize := getSizeOfPayloadSizeField(payload)
	field := make([]byte, 4)
	binary.LittleEndian.PutUint32(field, uint32(len(payload)))
	field = field[:fieldSize]
	rawBytes = append(rawBytes, field...)
	rawBytes[0] |= byte(fieldSize)
	return rawBytes
}

// openSymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *envelope) openSymmetric(key []byte, crypto CryptoBackend) (msg *receivedMessage, err error) {
	msg = &receivedMessage{crypto: crypto}
	decrypted, salt, err := crypto.DecryptSymmetric(e.Data, key)
	if err != nil {
		msg = nil
	} else {
		msg.Raw = decrypted
		msg.Salt = salt
	}
	return msg, err
}

// openAsymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *envelope) openAsymmetric(key *ecdsa.PrivateKey, crypto CryptoBackend) (*receivedMessage, error) {
	decrypted, err := crypto.DecryptAsymmetric(e.Data, key)
	switch err {
	case nil:
		message := &receivedMessage{Raw: decrypted, crypto: crypto}
		return message, nil
	case ecies.ErrInvalidPublicKey: // addressed to somebody else
		return nil, err
	default:
		return nil, fmt.Errorf("unable to open envelope, decrypt failed: %v", err)
	}
}

// == received message ==

// receivedMessage represents a data packet to be received
// and successfully decrypted.
type receivedMessage struct {
	Payload   []byte
	Raw       []byte
	Signature []byte
	Salt      []byte
	Padding   []byte

	Src *ecdsa.PublicKey
	Dst *ecdsa.PublicKey

	crypto CryptoBackend
}

// validateAndParse checks the message validity and extracts the fields in case of success.
func (msg *receivedMessage) validateAndParse() bool {
	end := len(msg.Raw)
	if end < 1 {
		return false
	}
	if isMessageSigned(msg.Raw[0]) {
		end -= signatureLength
		if end <= 1 {
			return false
		}
		msg.Signature = msg.Raw[end : end+signatureLength]
		msg.Src = msg.sigToPubKey()
		if msg.Src == nil {
			return false
		}
	}

	beg := 1
	payloadSize := 0
	sizeOfPayloadSizeField := int(msg.Raw[0] & SizeMask) // number of bytes indicating the size of payload
	if sizeOfPayloadSizeField != 0 {
		payloadSize = int(bytesToUintLittleEndian(msg.Raw[beg : beg+sizeOfPayloadSizeField]))
		if payloadSize+1 > end {
			return false
		}
		beg += sizeOfPayloadSizeField
		msg.Payload = msg.Raw[beg : beg+payloadSize]
	}

	beg += payloadSize
	msg.Padding = msg.Raw[beg:end]
	return true
}

// sigToPubKey returns the public key associated to the message's signature.
func (msg *receivedMessage) sigToPubKey() *ecdsa.PublicKey {
	defer func() { recover() }() // in case of invalid signature

	pub, err := msg.crypto.SigToPub(msg.hash(), msg.Signature)
	if err != nil {
		log.Error("failed to recover public key from signature", "err", err)
		return nil
	}
	return pub
}

// hash calculates the SHA3 checksum of the message flags, payload size field, payload and padding.
func (msg *receivedMessage) hash() []byte {
	if isMessageSigned(msg.Raw[0]) {
		sz := len(msg.Raw) - signatureLength
		return msg.crypto.Keccak256(msg.Raw[:sz])
	}
	return msg.crypto.Keccak256(msg.Raw)
}

// bytesToUintLittleEndian converts the slice to 64-bit unsigned integer.
func bytesToUintLittleEndian(b []byte) (res uint64) {
	mul := uint64(1)
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) * mul
		mul *= 256
	}
	return res
}

// === end messageReceived ===

type messageParams struct {
	Src     *ecdsa.PrivateKey
	Dst     *ecdsa.PublicKey
	KeySym  []byte
	Topic   Topic
	Payload []byte
	Padding []byte
}

// Code copied from whisper message methods (here as functions receiving a rawMessage param)

// appendPadding appends the padding specified in params.
// If no padding is provided in params, then random padding is generated.
func appendPadding(rawBytes rawMessage, params *messageParams, crypto CryptoBackend) (rawMessage, error) {
	if len(params.Padding) != 0 {
		// padding data was provided, just use it as is
		rawBytes = append(rawBytes, params.Padding...)
		return rawBytes, nil
	}

	rawSize := flagsLength + getSizeOfPayloadSizeField(params.Payload) + len(params.Payload)
	if params.Src != nil {
		rawSize += signatureLength
	}
	odd := rawSize % padSizeLimit
	paddingSize := padSizeLimit - odd
	pad := make([]byte, paddingSize)
	_, err := crand.Read(pad)
	if err != nil {
		return nil, err
	}

	if len(pad) != paddingSize || isAllZeroes(pad) {
		return nil, errors.New("failed to generate random padding of size " + strconv.Itoa(paddingSize))
	}
	rawBytes = append(rawBytes, pad...)
	return rawBytes, nil
}

func isAllZeroes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// getSizeOfPayloadSizeField returns the number of bytes necessary to encode the size of payload
func getSizeOfPayloadSizeField(payload []byte) int {
	s := 1
	for i := len(payload); i >= 256; i /= 256 {
		s++
	}
	return s
}

// sign calculates and sets the cryptographic signature for the message,
// also setting the sign flag.
func sign(rawBytes rawMessage, key *ecdsa.PrivateKey, crypto CryptoBackend) (rawMessage, error) {
	if isMessageSigned(rawBytes[0]) {
		// this should not happen, but no reason to panic
		log.Error("failed to sign the message: already signed")
		return rawBytes, nil
	}
	rawBytes[0] |= signatureFlag // it is important to set this flag before signing

	signature, err := crypto.Sign(rawBytes, key)

	if err != nil {
		rawBytes[0] &= (0xFF ^ signatureFlag) // clear the flag
		return nil, err
	}
	rawBytes = append(rawBytes, signature...)
	return rawBytes, nil
}

func isMessageSigned(flags byte) bool {
	return (flags & signatureFlag) != 0
}
