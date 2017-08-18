package pss

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

const (
	defaultSymKeyCacheCapacity = 512
	defaultDigestCacheTTL      = time.Second
)

// Pss configuration parameters
type PssParams struct {
	CacheTTL            time.Duration
	privateKey          *ecdsa.PrivateKey
	SymKeyCacheCapacity int
}

// Sane defaults for Pss
func NewPssParams(privatekey *ecdsa.PrivateKey) *PssParams {
	return &PssParams{
		CacheTTL:            defaultDigestCacheTTL,
		privateKey:          privatekey,
		SymKeyCacheCapacity: defaultSymKeyCacheCapacity,
	}
}

// variable length address
type PssAddress []byte

// abstraction to enable access to p2p.protocols.Peer.Send
type senderPeer interface {
	ID() discover.NodeID
	Address() []byte
	Send(interface{}) error
}

// used to encapsulate symkey in asymmetric key exchange
// if nonce is not nil. the message is symmetrically encrypted
// an encrypted keymsg is a response to an exchange request,
// where the key in the request is used to encrypt the response
type pssKeyMsg struct {
	From  []byte
	Key   []byte
	Nonce []byte
}

type pssPeer struct {
	rw      p2p.MsgReadWriter
	address PssAddress
	expires time.Time
}

type pssCacheEntry struct {
	expiresAt    time.Time
	receivedFrom []byte
}

type pssDigest [digestLength]byte

// Encapsulates messages transported over pss.
type PssMsg struct {
	To      []byte
	Payload *whisper.Envelope
}

// serializes the message for use in cache
func (msg *PssMsg) serialize() []byte {
	rlpdata, _ := rlp.EncodeToBytes(msg)
	return rlpdata
}

// String representation of PssMsg
func (self *PssMsg) String() string {
	return fmt.Sprintf("PssMsg: Recipient: %x", common.ByteLabel(self.To))
}

// Convenience wrapper for devp2p protocol messages for transport over pss
type ProtocolMsg struct {
	Code       uint64
	Size       uint32
	Payload    []byte
	ReceivedAt time.Time
}

// Creates a ProtocolMsg
func NewProtocolMsg(code uint64, msg interface{}) ([]byte, error) {

	rlpdata, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return nil, err
	}

	// TODO verify that nested structs cannot be used in rlp
	smsg := &ProtocolMsg{
		Code:    code,
		Size:    uint32(len(rlpdata)),
		Payload: rlpdata,
	}

	return rlp.EncodeToBytes(smsg)
}

// Signature for a message handler function for a PssMsg
//
// Implementations of this type are passed to Pss.Register together with a topic,
type Handler func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error
