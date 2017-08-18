package pss

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rlp"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
	"time"
)

// protocol specification of the pss capsule
var pssSpec = &protocols.Spec{
	Name:       "pss",
	Version:    1,
	MaxMsgSize: defaultMaxMsgSize,
	Messages: []interface{}{
		PssMsg{},
	},
}

// Bridges pss send/receive with devp2p protocol send/receive
//
// Implements p2p.MsgReadWriter
type PssReadWriter struct {
	*Pss
	LastActive time.Time
	rw         chan p2p.Msg
	spec       *protocols.Spec
	topic      *whisper.TopicType
	sendFunc   func(string, whisper.TopicType, []byte) error
	key        string
}

// Implements p2p.MsgReader
func (prw PssReadWriter) ReadMsg() (p2p.Msg, error) {
	msg := <-prw.rw
	log.Trace(fmt.Sprintf("pssrw readmsg: %v", msg))
	return msg, nil
}

// Implements p2p.MsgWriter
func (prw PssReadWriter) WriteMsg(msg p2p.Msg) error {
	log.Trace("pssrw writemsg", "msg", msg)
	rlpdata := make([]byte, msg.Size)
	msg.Payload.Read(rlpdata)
	pmsg, err := rlp.EncodeToBytes(ProtocolMsg{
		Code:    msg.Code,
		Size:    msg.Size,
		Payload: rlpdata,
	})
	if err != nil {
		return err
	}
	return prw.sendFunc(prw.key, *prw.topic, pmsg)
}

// Injects a p2p.Msg into the MsgReadWriter, so that it appears on the associated p2p.MsgReader
func (prw PssReadWriter) injectMsg(msg p2p.Msg) error {
	log.Trace(fmt.Sprintf("pssrw injectmsg: %v", msg))
	prw.rw <- msg
	return nil
}

// For devp2p protocol integration only.
//
// Convenience object for passing messages in and out of the p2p layer
type PssProtocol struct {
	*Pss
	proto        *p2p.Protocol
	topic        *whisper.TopicType
	spec         *protocols.Spec
	pubKeyRWPool map[string]p2p.MsgReadWriter
	symKeyRWPool map[string]p2p.MsgReadWriter
	flags        byte
}

// For devp2p protocol integration only.
//
// Maps a Topic to a devp2p protocol.
//
// flags: 0x01 = asymmetric messaging enabled, 0x02 symmetric messaging enabled
func RegisterPssProtocol(ps *Pss, topic *whisper.TopicType, spec *protocols.Spec, targetprotocol *p2p.Protocol, flags byte) (*PssProtocol, error) {
	if flags&0x03 == 0 {
		return nil, fmt.Errorf("specify at least one of asymmetric or symmetric messaging mode")
	}
	pp := &PssProtocol{
		Pss:          ps,
		proto:        targetprotocol,
		topic:        topic,
		spec:         spec,
		pubKeyRWPool: make(map[string]p2p.MsgReadWriter),
		symKeyRWPool: make(map[string]p2p.MsgReadWriter),
		flags:        flags,
	}
	return pp, nil
}

func (self *PssProtocol) isAsymmetric() bool {
	return self.flags&0x01 != 0
}

func (self *PssProtocol) isSymmetric() bool {
	return self.flags&0x02 != 0
}

// For devp2p protocol integration only.
//
// BROKEN! Implementation for pubkey lookup must be implemented
//
// Generic handler for initiating devp2p-like protocol connections
//
// This handler should be passed to Pss.Register with the associated topic.
func (self *PssProtocol) Handle(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	var vrw *PssReadWriter
	if self.isAsymmetric() != asymmetric && self.isSymmetric() == !asymmetric {
		return fmt.Errorf("invalid protocol encryption")
	} else if (!self.isActiveSymKeyProtocol(keyid, *self.topic) && !asymmetric) || (!self.isActiveAsymKeyProtocol(keyid, *self.topic) && asymmetric) {
		rw := &PssReadWriter{
			Pss:   self.Pss,
			rw:    make(chan p2p.Msg),
			spec:  self.spec,
			topic: self.topic,
			key:   keyid,
		}
		if asymmetric {
			rw.sendFunc = self.Pss.SendAsym
		} else {
			rw.sendFunc = self.Pss.SendSym
		}
		self.AddPeer(p, self.proto.Run, *self.topic, rw, asymmetric, keyid)
	}

	pmsg, err := ToP2pMsg(msg)
	if err != nil {
		return fmt.Errorf("could not decode pssmsg")
	}
	if asymmetric {
		vrw = self.pubKeyRWPool[keyid].(*PssReadWriter)
	} else {
		vrw = self.symKeyRWPool[keyid].(*PssReadWriter)
	}
	vrw.injectMsg(pmsg)

	return nil
}

func (self *PssProtocol) isActiveSymKeyProtocol(key string, topic whisper.TopicType) bool {
	return self.symKeyRWPool[key] != nil
}

func (self *PssProtocol) isActiveAsymKeyProtocol(key string, topic whisper.TopicType) bool {
	return self.pubKeyRWPool[key] != nil
}

// Creates a serialized (non-buffered) version of a p2p.Msg, used in the specialized p2p.MsgReadwriter implementations used internally by pss
//
// Should not normally be called outside the pss package hierarchy
func ToP2pMsg(msg []byte) (p2p.Msg, error) {
	payload := &ProtocolMsg{}
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

// For devp2p protocol integration only. Analogous to an outgoing devp2p connection.
//
// Links a remote peer and Topic to a dedicated p2p.MsgReadWriter in the pss peerpool, and runs the specificed protocol using these resources.
//
// The effect is that now we have a "virtual" protocol running on an artificial p2p.Peer, which can be looked up and piped to through Pss using swarm overlay address and topic
//
// The peer's encryption keys must be added separately.
func (self *PssProtocol) AddPeer(p *p2p.Peer, run func(*p2p.Peer, p2p.MsgReadWriter) error, topic whisper.TopicType, rw p2p.MsgReadWriter, asymmetric bool, key string) error {
	self.Pss.lock.Lock()
	defer self.Pss.lock.Unlock()
	if asymmetric {
		if _, ok := self.Pss.pubKeyPool[key]; ok == false {
			return fmt.Errorf("asym key does not exist: %s", key)
		}
		self.pubKeyRWPool[key] = rw
	} else {
		if _, ok := self.Pss.symKeyPool[key]; ok == false {
			return fmt.Errorf("symkey does not exist: %s", key)
		}
		self.symKeyRWPool[key] = rw
	}
	go func() {
		err := run(p, rw)
		log.Warn(fmt.Sprintf("pss vprotocol quit on addr %v topic %v: %v", topic, err))
	}()
	return nil
}
