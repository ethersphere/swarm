package mru

import (
	"bytes"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// TopicLength establishes the max length of a topic string
const TopicLength = storage.KeyLength

// Topic represents what a resource talks about
type Topic struct {
	content [TopicLength]byte
}

// NewTopic creates a new topic from a provided name and "related content" byte array,
// merging the two together.
// If relatedContent or name are longer than TopicLength, they will be truncated.
// name can be an empty string
// relatedContent can be nil
func NewTopic(name string, relatedContent []byte) Topic {
	var topic Topic
	if relatedContent != nil {
		contentLength := len(relatedContent)
		if contentLength > TopicLength {
			contentLength = TopicLength
		}
		copy(topic.content[:], relatedContent[:contentLength])
	}
	nameBytes := []byte(name)
	nameLength := len(nameBytes)
	if nameLength > TopicLength {
		nameLength = TopicLength
	}
	bitutil.XORBytes(topic.content[:], topic.content[:], nameBytes[:nameLength])
	return topic
}

// Hex will return the topic encoded as an hex string
func (t *Topic) Hex() string {
	return hexutil.Encode(t.content[:])
}

// FromHex will parse a hex string into this Topic instance
func (t *Topic) FromHex(hex string) error {
	return decodeHexArray(t.content[:], hex, "Topic")
}

// Name will try to extract the resource name out of the topic
func (t *Topic) Name(relatedContent []byte) string {
	nameBytes := t
	if relatedContent != nil {
		contentLength := len(relatedContent)
		if contentLength > TopicLength {
			contentLength = TopicLength
		}
		bitutil.XORBytes(nameBytes.content[:], t.content[:], relatedContent[:contentLength])
	}
	z := bytes.IndexByte(nameBytes.content[:], 0)
	if z < 0 {
		z = TopicLength
	}
	return string(nameBytes.content[:z])

}

// UnmarshalJSON implements the json.Unmarshaller interface
func (t *Topic) UnmarshalJSON(data []byte) error {
	var hex string
	json.Unmarshal(data, &hex)
	return t.FromHex(hex)
}

// MarshalJSON implements the json.Marshaller interface
func (t *Topic) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Hex())
}
