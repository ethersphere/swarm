package notify

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/swarm/pss"
)

const (
	MsgCodeStart = iota
	MsgCodeNotify
	MsgCodeStop
	MsgCodeMax
)

const (
	minimumAddressLength = 1
)

var (
	controlTopic = pss.Topic{0x00, 0x00, 0x00, 0x01}
)

// when code is MsgCodeStart, Payload is address
// when code is MsgCodeNotify, Payload is symkey. If len = 0, keep old key
// when code is MsgCodeStop, Payload is address
type Msg struct {
	Code    byte
	Name    string
	Payload []byte
}

func (self *Msg) Serialize() []byte {
	b := bytes.NewBuffer(nil)
	b.Write([]byte{self.Code})
	ib := make([]byte, 2)
	binary.LittleEndian.PutUint16(ib, uint16(len(self.Name)))
	b.Write(ib)
	binary.LittleEndian.PutUint16(ib, uint16(len(self.Payload)))
	b.Write(ib)
	b.Write([]byte(self.Name))
	b.Write(self.Payload)
	return b.Bytes()
}

func deserializeMsg(msgbytes []byte) (*Msg, error) {
	msg := &Msg{
		Code: msgbytes[0],
	}
	nameLen := binary.LittleEndian.Uint16(msgbytes[1:3])
	dataLen := binary.LittleEndian.Uint16(msgbytes[3:5])
	if int(nameLen+dataLen)+5 != len(msgbytes) {
		return nil, errors.New("Corrupt message")
	}
	msg.Name = string(msgbytes[5 : 5+nameLen])
	msg.Payload = msgbytes[5+nameLen:]
	return msg, nil
}

// a notifier has one sendmux entry for each address space it sends messages to
type sendMux struct {
	address  pss.PssAddress
	symkeyid string
	count    int
}

type notifier struct {
	muxes     []*sendMux
	topic     pss.Topic
	threshold int
	ackFunc   func(string) []byte
}

type Controller struct {
	pss       *pss.Pss
	notifiers map[string]*notifier
}

func NewController(ps *pss.Pss) *Controller {
	return &Controller{
		pss:       ps,
		notifiers: make(map[string]*notifier),
	}
}

func (self *Controller) NewNotifier(name string, threshold int, ackFunc func(string) []byte) error {
	if _, ok := self.notifiers[name]; ok {
		return fmt.Errorf("%s already exists in controller", name)
	}
	self.notifiers[name] = &notifier{
		topic:     pss.BytesToTopic([]byte(name)),
		threshold: threshold,
		ackFunc:   ackFunc,
	}
	return nil
}

func (self *Controller) addToNotifier(name string, address pss.PssAddress) (string, error) {
	notifier, ok := self.notifiers[name]
	if !ok {
		return "", fmt.Errorf("Unknown notifier %s", name)
	}
	for _, m := range notifier.muxes {
		if bytes.Equal(address, m.address) {
			m.count++
			return m.symkeyid, nil
		}
	}
	symkeyid, err := self.pss.GenerateSymmetricKey(notifier.topic, &address, false)
	if err != nil {
		return "", fmt.Errorf("Generate symkey fail: %v", err)
	}
	notifier.muxes = append(notifier.muxes, &sendMux{
		address:  address,
		symkeyid: symkeyid,
		count:    1,
	})
	return symkeyid, nil
}

func (self *Controller) Handler(smsg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
	// control messages should be asym
	if !asymmetric {
		return errors.New("Control messages must be asymmetric")
	}
	pubkey := crypto.ToECDSAPub(common.FromHex(keyid))

	// see if the message is valid
	msg, err := deserializeMsg(smsg)
	if err != nil {
		return fmt.Errorf("Invalid message: %v", err)
	}

	switch msg.Code {
	case MsgCodeStart:

		// if name is not registered for notifications we will not react
		if _, ok := self.notifiers[msg.Name]; !ok {
			return fmt.Errorf("Subscribe attempted on unknown resource %s", msg.Name)
		}

		// parse the address from the message and truncate if longer than our mux threshold
		address := msg.Payload
		if len(msg.Payload) > self.notifiers[msg.Name].threshold {
			address = address[:self.notifiers[msg.Name].threshold]
		}

		// add the address to the notification list
		symkeyid, err := self.addToNotifier(msg.Name, address)
		if err != nil {
			return fmt.Errorf("add address to notifier fail: %v", err)
		}

		// add to address book for send initial notify
		pssaddr := pss.PssAddress(address)
		err = self.pss.SetPeerPublicKey(pubkey, controlTopic, &pssaddr)
		if err != nil {
			return fmt.Errorf("add pss peer for reply fail: %v", err)
		}

		// send initial notify, will contain symkey to use for consequtive messages
		err = self.pss.SendAsym(keyid, controlTopic, self.notifiers[msg.Name].ackFunc(symkeyid))
		if err != nil {
			return fmt.Errorf("send start reply fail: %v", err)
		}
	default:
		return fmt.Errorf("Invalid message code: %d", msg.Code)
	}

	return nil
}
