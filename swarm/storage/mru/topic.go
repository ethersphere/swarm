package mru

import (
	"bytes"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const topicLength = storage.KeyLength

type Topic struct {
	content [topicLength]byte
}

func NewTopic(name string, relatedContent storage.Address) Topic {
	var topic Topic
	copy(topic.content[:], relatedContent)
	nameLength := len(name)
	if nameLength > topicLength {
		nameLength = topicLength
	}
	bitutil.XORBytes(topic.content[:], topic.content[:], []byte(name)[:nameLength])
	return topic
}

func (t *Topic) Hex() string {
	return hexutil.Encode(t.content[:])
}

func (t *Topic) FromHex(hex string) error {
	return decodeHexArray(t.content[:], hex, "Topic")
}

func (t *Topic) Name(relatedContent storage.Address) string {
	nameBytes := t
	bitutil.XORBytes(nameBytes.content[:], t.content[:], relatedContent)
	z := bytes.IndexByte(nameBytes.content[:], 0)
	if z < 0 {
		z = topicLength
	}
	return string(nameBytes.content[:z])

}

func (t *Topic) UnmarshalJSON(data []byte) error {
	var hex string
	json.Unmarshal(data, &hex)
	return t.FromHex(hex)
}

func (t *Topic) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Hex())
}
