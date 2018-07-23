package mru

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const topicLength = storage.KeyLength

type Topic [topicLength]byte

func NewTopic(name string, relatedContent storage.Address) Topic {
	var topic Topic
	copy(topic[:], relatedContent)
	nameLength := len(name)
	if nameLength > topicLength {
		nameLength = topicLength
	}
	bitutil.XORBytes(topic[:], topic[:], []byte(name)[:nameLength])
	return topic
}

func (t *Topic) Hex() string {
	return hexutil.Encode(t[:])
}

func (t *Topic) FromHex(hex string) error {
	return decodeHexArray(t[:], hex, "Topic")
}

func (t *Topic) Name(relatedContent storage.Address) string {
	nameBytes := t
	bitutil.XORBytes(nameBytes[:], t[:], relatedContent)
	z := bytes.IndexByte(nameBytes[:], 0)
	if z < 0 {
		z = topicLength
	}
	return string(nameBytes[:z])

}
