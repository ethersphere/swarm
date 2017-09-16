package pss

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

type PingMsg struct {
	Created time.Time
	Pong    bool
}

type Ping struct {
	Pong bool
	OutC chan bool
	InC  chan bool
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

// Sample protocol used for tests
var PingProtocol = &protocols.Spec{
	Name:       "psstest",
	Version:    1,
	MaxMsgSize: 10 * 1024 * 1024,
	Messages: []interface{}{
		PingMsg{},
	},
}

var PingTopic = whisper.BytesToTopic([]byte(fmt.Sprintf("%s:%d", PingProtocol.Name, PingProtocol.Version)))

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

func NewTestPss(privkey *ecdsa.PrivateKey, ppextra *PssParams) *Pss {

	var nid discover.NodeID
	copy(nid[:], crypto.FromECDSAPub(&privkey.PublicKey))
	addr := network.NewAddrFromNodeID(nid)

	// set up storage
	cachedir, err := ioutil.TempDir("", "pss-cache")
	if err != nil {
		log.Error("create pss cache tmpdir failed", "error", err)
		os.Exit(1)
	}
	dpa, err := storage.NewLocalDPA(cachedir)
	if err != nil {
		log.Error("local dpa creation failed", "error", err)
		os.Exit(1)
	}

	// set up routing
	kp := network.NewKadParams()
	kp.MinProxBinSize = 3

	// create pss
	pp := NewPssParams(privkey)
	if ppextra != nil {
		pp.SymKeyCacheCapacity = ppextra.SymKeyCacheCapacity
	}

	overlay := network.NewKademlia(addr.Over(), kp)
	ps := NewPss(overlay, dpa, pp)

	return ps
}
