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

// Contains all types and code related to messages and envelopes.
// Currently backed by whisperv6
package pss

import (
	"crypto/ecdsa"

	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

func WBytesToTopic(b []byte) (t Topic) {
	return Topic(whisper.BytesToTopic(b))
}

type Envelope struct {
	Topic  Topic
	Data   []byte
	Expiry uint32
}

// == Envelope ==

// NewSentEnvelope creates and initializes a non-signed, non-encrypted Whisper message
// and then wrap it and encrypt it. It performs what it used to be two function calls:
// msg, e1 := NewSentMessage and env, e := msg.Wrap
func NewSentEnvelope(params *MessageParams) (*Envelope, error) {
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

// OpenSymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *Envelope) OpenSymmetric(key []byte) (*ReceivedMessage, error) {
	whisperEnvelope := toWhisperEnvelope(e)
	whisperMsg, err := whisperEnvelope.OpenSymmetric(key)
	if err != nil {
		return nil, err
	}
	msg := toReceivedMessage(whisperMsg)
	return msg, nil
}

// OpenAsymmetric tries to decrypt an envelope, potentially encrypted with a particular key.
func (e *Envelope) OpenAsymmetric(key *ecdsa.PrivateKey) (*ReceivedMessage, error) {
	whisperEnvelope := toWhisperEnvelope(e)
	whisperMsg, err := whisperEnvelope.OpenAsymmetric(key)
	if err != nil {
		return nil, err
	}

	msg := toReceivedMessage(whisperMsg)
	return msg, nil
}

// == Received message ==

// ReceivedMessage represents a data packet to be received
// and successfully decrypted.
type ReceivedMessage struct {
	Payload   []byte
	Raw       []byte
	Signature []byte
	Salt      []byte
	Padding   []byte

	Src *ecdsa.PublicKey // Message recipient (identity used to decode the message)
	Dst *ecdsa.PublicKey // Message recipient (identity used to decode the message)

}

// ValidateAndParse checks the message validity and extracts the fields in case of success.
func (msg *ReceivedMessage) ValidateAndParse() bool {
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

type MessageParams struct {
	Src     *ecdsa.PrivateKey
	Dst     *ecdsa.PublicKey
	KeySym  []byte
	Topic   Topic
	Payload []byte
	Padding []byte
}

func toReceivedMessage(whisperMsg *whisper.ReceivedMessage) *ReceivedMessage {
	return &ReceivedMessage{
		Payload:   whisperMsg.Payload,
		Raw:       whisperMsg.Raw,
		Signature: whisperMsg.Signature,
		Salt:      whisperMsg.Salt,
		Src:       whisperMsg.Src,
		Dst:       whisperMsg.Dst,
	}
}

// == Conversion functions ==

func toWhisperEnvelope(e *Envelope) *whisper.Envelope {
	// uint32(time.Now().Add(time.Second * time.Duration(defaultWhisperTTL)).Unix()),
	whisperEnvelope := &whisper.Envelope{
		Expiry: e.Expiry,
		TTL:    defaultWhisperTTL,
		Topic:  whisper.TopicType(e.Topic),
		Data:   e.Data,
		Nonce:  0,
	}
	return whisperEnvelope
}

func toPssEnvelope(whisperEnvelope *whisper.Envelope) *Envelope {
	return &Envelope{
		Topic:  Topic(whisperEnvelope.Topic),
		Data:   whisperEnvelope.Data,
		Expiry: whisperEnvelope.Expiry,
	}
}

func toWhisperParams(params *MessageParams) *whisper.MessageParams {
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
