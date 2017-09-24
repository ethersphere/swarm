package pss

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/protocols"
)

// Generic ping protocol implementation for
// pss devp2p protocol emulation
type PingMsg struct {
	Created time.Time
	Pong    bool // set if message is pong reply
}

type Ping struct {
	Pong bool      // toggle pong reply upon ping receive
	OutC chan bool // trigger ping
	InC  chan bool // optional, report back to calling code
}

func (self *Ping) PingHandler(msg interface{}) error {
	var pingmsg *PingMsg
	var ok bool
	if pingmsg, ok = msg.(*PingMsg); !ok {
		return errors.New("invalid msg")
	}
	log.Debug("ping handler", "msg", pingmsg, "outc", self.OutC)
	if self.InC != nil {
		self.InC <- pingmsg.Pong
	}
	if self.Pong && !pingmsg.Pong {
		self.OutC <- true
	}
	return nil
}

var PingProtocol = &protocols.Spec{
	Name:       "psstest",
	Version:    1,
	MaxMsgSize: 1024,
	Messages: []interface{}{
		PingMsg{},
	},
}

var PingTopic = ProtocolTopic(PingProtocol)

func NewPingProtocol(pingC chan bool, handler func(interface{}) error) *p2p.Protocol {
	return &p2p.Protocol{
		Name:    PingProtocol.Name,
		Version: PingProtocol.Version,
		Length:  uint64(PingProtocol.MaxMsgSize),
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			quitC := make(chan struct{})
			pp := protocols.NewPeer(p, rw, PingProtocol)
			log.Trace(fmt.Sprintf("running pss vprotocol on peer %v", p, "outc", pingC))
			go func() {
				for {
					select {
					case ispong := <-pingC:
						pp.Send(&PingMsg{
							Created: time.Now(),
							Pong:    ispong,
						})
					case <-quitC:
					}
				}
			}()
			err := pp.Run(handler)
			quitC <- struct{}{}
			return err
		},
	}
}
