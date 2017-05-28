package pss

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/p2p"
	//"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	TopicLength                 = 32
	DefaultTTL                  = 6000
	defaultDigestCacheTTL       = time.Second
)

type PssAdapter interface {
	Send (to []byte, topic PssTopic, msg []byte) error
	Register(topic *PssTopic, handler PssHandler) func()
	BaseAddr() []byte
	// implements node.Service
	Start(srv *p2p.Server) error
	Stop() error
	Protocols() []p2p.Protocol
	APIs() []rpc.API
	Process(pssmsg *PssMsg) error
	/*IsActive(pot.Address, PssTopic) bool
	AddPeer (*p2p.Peer, pot.Address, func(*p2p.Peer, p2p.MsgReadWriter) error, PssTopic, p2p.MsgReadWriter) error */
}

// Defines params for Pss
type PssParams struct {
	Cachettl time.Duration
}

// Initializes default params for Pss
func NewPssParams() *PssParams {
	return &PssParams{
		Cachettl: defaultDigestCacheTTL,
	}
}

// Encapsulates the message transported over pss.
type PssMsg struct {
	To      []byte
	Payload *PssEnvelope
}

func (msg *PssMsg) Serialize() []byte {
	rlpdata, _ := rlp.EncodeToBytes(msg)
	/*buf := bytes.NewBuffer(nil)
	buf.Write(self.PssEnvelope.Topic[:])
	buf.Write(self.PssEnvelope.Payload)
	buf.Write(self.PssEnvelope.From)
	return buf.Bytes()*/
	return rlpdata
}


// String representation of PssMsg
func (self *PssMsg) String() string {
	return fmt.Sprintf("PssMsg: Recipient: %x", common.ByteLabel(self.To))
}

// Topic defines the context of a message being transported over pss
// It is used by pss to determine what action is to be taken on an incoming message
// Typically, one can map protocol handlers for the message payloads by mapping topic to them; see *Pss.Register()
type PssTopic [TopicLength]byte

func (self *PssTopic) String() string {
	return fmt.Sprintf("%x", self)
}

// Pre-Whisper placeholder, payload of PssMsg
type PssEnvelope struct {
	Topic   PssTopic
	TTL     uint16
	Payload []byte
	From    []byte
}

// creates Pss envelope from sender address, topic and raw payload
func NewPssEnvelope(addr []byte, topic PssTopic, payload []byte) *PssEnvelope {
	return &PssEnvelope{
		From:    addr,
		Topic:   topic,
		TTL:     DefaultTTL,
		Payload: payload,
	}
}

// encapsulates a protocol msg as PssEnvelope data
type PssProtocolMsg struct {
	Code       uint64
	Size       uint32
	Payload    []byte
	ReceivedAt time.Time
}

// PssAPIMsg is the type for messages, it extends the rlp encoded protocol Msg
// with the Sender's overlay address
type PssAPIMsg struct {
	Msg  []byte
	Addr []byte
}

func NewProtocolMsg(code uint64, msg interface{}) ([]byte, error) {

	rlpdata, err := rlp.EncodeToBytes(msg)
	if err != nil {
		return nil, err
	}

	// previous attempts corrupted nested structs in the payload iself upon deserializing
	// therefore we use two separate []byte fields instead of peerAddr
	// TODO verify that nested structs cannot be used in rlp
	smsg := &PssProtocolMsg{
		Code:    code,
		Size:    uint32(len(rlpdata)),
		Payload: rlpdata,
	}

	return rlp.EncodeToBytes(smsg)
}

// Message handler func for a topic
type PssHandler func(msg []byte, p *p2p.Peer, from []byte) error

// constructs a new PssTopic from a given name and version.
//
// Analogous to the name and version members of p2p.Protocol
func NewTopic(s string, v int) (topic PssTopic) {
	h := sha3.NewKeccak256()
	h.Write([]byte(s))
	buf := make([]byte, TopicLength/8)
	binary.PutUvarint(buf, uint64(v))
	h.Write(buf)
	copy(topic[:], h.Sum(buf)[:])
	return topic
}

func ToP2pMsg(msg []byte) (p2p.Msg, error) {
	payload := &PssProtocolMsg{}
	if err := rlp.DecodeBytes(msg, payload); err != nil {
		return p2p.Msg{}, fmt.Errorf("pss protocol handler unable to decode payload as p2p message: %v", err)
	}

	return p2p.Msg{
		Code:       payload.Code,
		Size:       uint32(len(payload.Payload)),
		ReceivedAt: time.Now(),
		Payload:    bytes.NewBuffer(payload.Payload),
	}, nil
}
