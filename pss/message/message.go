package message

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// Message encapsulates messages transported over pss.
type Message struct {
	To      []byte
	Flags   Flags
	Expire  uint32
	Topic   Topic
	Payload []byte
}

const digestLength = 32 // byte length of digest used for pss cache (currently same as swarm chunk hash)

// Digest holds the digest of a message used for caching
type Digest [digestLength]byte

// New creates a new PSS message
func New(flags Flags) *Message {
	return &Message{
		Flags: flags,
	}
}

// Digest computes a message digest for use as a cache key
func (msg *Message) Digest() Digest {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(msg.To)
	hasher.Write(msg.Topic[:])
	hasher.Write(msg.Payload)
	key := hasher.Sum(nil)
	d := Digest{}
	copy(d[:], key[:digestLength])
	return d
}

// String representation of a PSS message
func (msg *Message) String() string {
	return fmt.Sprintf("PssMsg: Recipient: %s, Topic: %v", common.ToHex(msg.To), msg.Topic.String())
}
