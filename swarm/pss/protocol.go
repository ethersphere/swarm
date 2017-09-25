package pss

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rlp"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
	"time"
)

// Protocol options to be passed to a new Protocol instance
//
// The parameters specify which encryption schemes to allow
type ProtocolParams struct {
	Asymmetric bool
	Symmetric  bool
}

// PssReadWriter bridges pss send/receive with devp2p protocol send/receive
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
func (prw *PssReadWriter) ReadMsg() (p2p.Msg, error) {
	msg := <-prw.rw
	log.Trace(fmt.Sprintf("pssrw readmsg: %v", msg))
	return msg, nil
}

// Implements p2p.MsgWriter
func (prw *PssReadWriter) WriteMsg(msg p2p.Msg) error {
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
func (prw *PssReadWriter) injectMsg(msg p2p.Msg) error {
	log.Trace(fmt.Sprintf("pssrw injectmsg: %v", msg))
	prw.rw <- msg
	return nil
}

// Convenience object for emulation devp2p over pss
type Protocol struct {
	*Pss
	proto        *p2p.Protocol
	topic        *whisper.TopicType
	spec         *protocols.Spec
	pubKeyRWPool map[string]p2p.MsgReadWriter
	symKeyRWPool map[string]p2p.MsgReadWriter
	Asymmetric   bool
	Symmetric    bool
}

// Activates devp2p emulation over a specific pss topic
//
// One or both encryption schemes must be specified. If
// only one is specified, the protocol will not be valid
// for the other, and will make the message handler
// return errors
func RegisterProtocol(ps *Pss, topic *whisper.TopicType, spec *protocols.Spec, targetprotocol *p2p.Protocol, options *ProtocolParams) (*Protocol, error) {
	if !options.Asymmetric && !options.Symmetric {
		return nil, errors.New(fmt.Sprintf("specify at least one of asymmetric or symmetric messaging mode"))
	}
	pp := &Protocol{
		Pss:          ps,
		proto:        targetprotocol,
		topic:        topic,
		spec:         spec,
		pubKeyRWPool: make(map[string]p2p.MsgReadWriter),
		symKeyRWPool: make(map[string]p2p.MsgReadWriter),
		Asymmetric:   options.Asymmetric,
		Symmetric:    options.Symmetric,
	}
	return pp, nil
}

// Generic handler for incoming messages over devp2p emulation
//
// To be passed to pss.Register()
//
// Will run the protocol on a new incoming peer, provided that
// the encryption key of the message has a match in the internal
// pss keypool
//
// Fails if protocol is not valid for the message encryption scheme,
// if adding a new peer fails, or if the message is not a serialized
// p2p.Msg (which it always will be if it is sent from this object).
func (self *Protocol) Handle(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	var vrw *PssReadWriter
	if self.Asymmetric != asymmetric && self.Symmetric == !asymmetric {
		return errors.New(fmt.Sprintf("invalid protocol encryption"))
	} else if (!self.isActiveSymKey(keyid, *self.topic) && !asymmetric) ||
		(!self.isActiveAsymKey(keyid, *self.topic) && asymmetric) {

		rw, err := self.AddPeer(p, self.proto.Run, *self.topic, asymmetric, keyid)
		if err != nil {
			return err
		}
		vrw = rw.(*PssReadWriter)
	}

	pmsg, err := ToP2pMsg(msg)
	if err != nil {
		return errors.New(fmt.Sprintf("could not decode pssmsg"))
	}
	if asymmetric {
		vrw = self.pubKeyRWPool[keyid].(*PssReadWriter)
	} else {
		vrw = self.symKeyRWPool[keyid].(*PssReadWriter)
	}
	vrw.injectMsg(pmsg)
	return nil
}

// check if (peer) symmetric key is currently registered with this topic
func (self *Protocol) isActiveSymKey(key string, topic whisper.TopicType) bool {
	return self.symKeyRWPool[key] != nil
}

// check if (peer) asymmetric key is currently registered with this topic
func (self *Protocol) isActiveAsymKey(key string, topic whisper.TopicType) bool {
	return self.pubKeyRWPool[key] != nil
}

// Creates a serialized (non-buffered) version of a p2p.Msg, used in the specialized internal p2p.MsgReadwriter implementations
func ToP2pMsg(msg []byte) (p2p.Msg, error) {
	payload := &ProtocolMsg{}
	if err := rlp.DecodeBytes(msg, payload); err != nil {
		return p2p.Msg{}, errors.New(fmt.Sprintf("pss protocol handler unable to decode payload as p2p message: %v", err))
	}

	return p2p.Msg{
		Code:       payload.Code,
		Size:       uint32(len(payload.Payload)),
		ReceivedAt: time.Now(),
		Payload:    bytes.NewBuffer(payload.Payload),
	}, nil
}

// Runs an emulated pss Protocol on the specified peer,
// linked to a specific topic
// `key` and `asymmetric` specifies what encryption key
// to link the peer to.
// The key must exist in the pss store prior to adding the peer.
func (self *Protocol) AddPeer(p *p2p.Peer, run func(*p2p.Peer, p2p.MsgReadWriter) error, topic whisper.TopicType, asymmetric bool, key string) (p2p.MsgReadWriter, error) {
	self.Pss.lock.Lock()
	defer self.Pss.lock.Unlock()
	rw := &PssReadWriter{
		Pss:   self.Pss,
		rw:    make(chan p2p.Msg),
		spec:  self.spec,
		topic: self.topic,
		key:   key,
	}
	if asymmetric {
		rw.sendFunc = self.Pss.SendAsym
	} else {
		rw.sendFunc = self.Pss.SendSym
	}
	if asymmetric {
		if _, ok := self.Pss.pubKeyPool[key]; !ok {
			return nil, errors.New(fmt.Sprintf("asym key does not exist: %s", key))
		}
		self.pubKeyRWPool[key] = rw
	} else {
		if _, ok := self.Pss.symKeyPool[key]; !ok {
			return nil, errors.New(fmt.Sprintf("symkey does not exist: %s", key))
		}
		self.symKeyRWPool[key] = rw
	}
	go func() {
		err := run(p, rw)
		log.Warn(fmt.Sprintf("pss vprotocol quit on addr %v topic %v: %v", topic, err))
	}()
	return rw, nil
}

// Uniform translation of protocol specifiers to topic
func ProtocolTopic(spec *protocols.Spec) whisper.TopicType {
	return whisper.BytesToTopic([]byte(fmt.Sprintf("%s:%d", spec.Name, spec.Version)))
}
