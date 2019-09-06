package message

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethersphere/swarm/storage"
)

// TopicLength sets the length of the message topic
const TopicLength = 4

// Topic is the PSS encapsulation of the Whisper topic type
type Topic [TopicLength]byte

// NewTopic hashes an arbitrary length byte slice and truncates it to the length of a topic, using only the first bytes of the digest
func NewTopic(b []byte) Topic {
	topicHashFunc := storage.MakeHashFunc("SHA256")()
	topicHashFunc.Write(b)
	return toTopic(topicHashFunc.Sum(nil))
}

// toTopic converts from the byte array representation of a topic
// into the Topic type.
func toTopic(b []byte) (t Topic) {
	sz := TopicLength
	if x := len(b); x < TopicLength {
		sz = x
	}
	for i := 0; i < sz; i++ {
		t[i] = b[i]
	}
	return t
}

func (t *Topic) String() string {
	return hexutil.Encode(t[:])
}

// MarshalJSON implements the json.Marshaler interface
func (t Topic) MarshalJSON() (b []byte, err error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements the json.Marshaler interface
func (t *Topic) UnmarshalJSON(input []byte) error {
	topicbytes, err := hexutil.Decode(string(input[1 : len(input)-1]))
	if err != nil {
		return err
	}
	copy(t[:], topicbytes)
	return nil
}
