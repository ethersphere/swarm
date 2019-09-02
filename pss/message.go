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

	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

func toTopic(b []byte) (t Topic) {
	return Topic(whisper.BytesToTopic(b))
}

type envelope struct {
	Topic  Topic
	Data   []byte
	Expiry uint32
}

// == envelope ==

// newSentEnvelope creates and initializes a non-signed, non-encrypted Whisper message
// and then wrap it and encrypt it. It performs what it used to be two function calls:
// msg, e1 := NewSentMessage and env, e := msg.Wrap
func newSentEnvelope(params *messageParams) (*envelope, error) {
	whisperParams := toWhisperParams(params)

	message, e := whisper.NewSentMessage(whisperParams)
	if e != nil {
		return nil, e
	}

	whisperEnvelope, e := message.Wrap(whisperParams)
	if e != nil {
		return nil, e
	}

	return toPssEnvelope(whisperEnvelope), nil
}

// openSymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *envelope) openSymmetric(key []byte) (*receivedMessage, error) {
	whisperEnvelope := toWhisperEnvelope(e)
	whisperMsg, err := whisperEnvelope.OpenSymmetric(key)
	if err != nil {
		return nil, err
	}
	msg := toReceivedMessage(whisperMsg)
	return msg, nil
}

// openAsymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *envelope) openAsymmetric(key *ecdsa.PrivateKey) (*receivedMessage, error) {
	whisperEnvelope := toWhisperEnvelope(e)
	whisperMsg, err := whisperEnvelope.OpenAsymmetric(key)
	if err != nil {
		return nil, err
	}

	msg := toReceivedMessage(whisperMsg)
	return msg, nil
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
}

// validateAndParse checks the message validity and extracts the fields in case of success.
func (msg *receivedMessage) validateAndParse() bool {
	whisperRecvMsg := &whisper.ReceivedMessage{
		Raw: msg.Raw,
	}

	success := whisperRecvMsg.ValidateAndParse()
	if success {
		msg.Signature = whisperRecvMsg.Signature
		msg.Src = whisperRecvMsg.Src
		msg.Payload = whisperRecvMsg.Payload
		msg.Padding = whisperRecvMsg.Padding
	}
	return success
}

type messageParams struct {
	Src     *ecdsa.PrivateKey
	Dst     *ecdsa.PublicKey
	KeySym  []byte
	Topic   Topic
	Payload []byte
	Padding []byte
}

// == Conversion functions to/from whisper ==

func toReceivedMessage(whisperMsg *whisper.ReceivedMessage) *receivedMessage {
	return &receivedMessage{
		Payload:   whisperMsg.Payload,
		Raw:       whisperMsg.Raw,
		Signature: whisperMsg.Signature,
		Salt:      whisperMsg.Salt,
		Src:       whisperMsg.Src,
		Dst:       whisperMsg.Dst,
	}
}

func toWhisperEnvelope(e *envelope) *whisper.Envelope {
	whisperEnvelope := &whisper.Envelope{
		Expiry: e.Expiry,
		TTL:    defaultWhisperTTL,
		Topic:  whisper.TopicType(e.Topic),
		Data:   e.Data,
		Nonce:  0,
	}
	return whisperEnvelope
}

func toPssEnvelope(whisperEnvelope *whisper.Envelope) *envelope {
	return &envelope{
		Topic:  Topic(whisperEnvelope.Topic),
		Data:   whisperEnvelope.Data,
		Expiry: whisperEnvelope.Expiry,
	}
}

func toWhisperParams(params *messageParams) *whisper.MessageParams {
	whisperParams := &whisper.MessageParams{
		TTL:      defaultWhisperTTL,
		Src:      params.Src,
		Dst:      params.Dst,
		KeySym:   params.KeySym,
		Topic:    whisper.TopicType(params.Topic),
		WorkTime: defaultWhisperWorkTime,
		PoW:      defaultWhisperPoW,
		Payload:  params.Payload,
		Padding:  params.Padding,
	}
	return whisperParams
}
