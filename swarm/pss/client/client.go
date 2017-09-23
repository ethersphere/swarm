package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/pss"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

const (
	inboxCapacity         = 3000
	outboxCapacity        = 100
	addrLen               = common.HashLength
	handshakeRetryTimeout = 5000
	handshakeRetryCount   = 3
)

var ()

// After a successful connection with Client.Start, BaseAddr contains the swarm overlay address of the pss node
type Client struct {
	BaseAddr []byte

	// peers
	peerPool map[whisper.TopicType]map[string]*pssRPCRW
	protos   map[whisper.TopicType]*p2p.Protocol

	// rpc connections
	rpc *rpc.Client
	sub *rpc.ClientSubscription

	// channels
	topicsC chan []byte
	msgC    chan pss.APIMsg
	quitC   chan struct{}

	lock sync.Mutex
}

// implements p2p.MsgReadWriter
type pssRPCRW struct {
	*Client
	hextopic string
	msgC     chan []byte
	addr     pss.PssAddress
	pubKeyId string
	symKeyId *string
	lastSeen time.Time
}

func (self *Client) newpssRPCRW(pubkey *ecdsa.PublicKey, addr pss.PssAddress, topic *whisper.TopicType) *pssRPCRW {
	hextopic := fmt.Sprintf("%x", *topic)
	pubkeybytes := crypto.FromECDSAPub(pubkey)
	err := self.rpc.Call(nil, "pss_setPeerPublicKey", pubkeybytes, hextopic, addr)
	if err != nil {
		return nil
	}
	return &pssRPCRW{
		Client:   self,
		hextopic: hextopic,
		msgC:     make(chan []byte),
		addr:     addr,
		pubKeyId: common.ToHex(pubkeybytes),
	}
}

func (rw *pssRPCRW) ReadMsg() (p2p.Msg, error) {
	msg := <-rw.msgC
	log.Trace("pssrpcrw read", "msg", msg)
	pmsg, err := pss.ToP2pMsg(msg)
	if err != nil {
		return p2p.Msg{}, err
	}

	return pmsg, nil
}

// if current symkey (pointed to by rw.symKeyId) is expired,
// pointer is changed to next in buffer
// then new is requested through handshake
// if buffer is empty, handshake request blocks until return
// after which pointer is changed to first new key in buffer
// will fail if:
// - any api calls fail
// - handshake retries are exhausted without reply,
// - send fails
func (rw *pssRPCRW) WriteMsg(msg p2p.Msg) error {
	log.Trace("got writemsg pssclient", "msg", msg)
	rlpdata := make([]byte, msg.Size)
	msg.Payload.Read(rlpdata)
	pmsg, err := rlp.EncodeToBytes(pss.ProtocolMsg{
		Code:    msg.Code,
		Size:    msg.Size,
		Payload: rlpdata,
	})
	if err != nil {
		return err
	}

	// If we have a pointer, check if it is expired
	var symkeycap uint16
	if rw.symKeyId != nil {
		err = rw.Client.rpc.Call(&symkeycap, "pss_getHandshakeKeyCapacity", *rw.symKeyId)
		if err != nil {
			return err
		}
	}

	if symkeycap == 0 {
		// The key has expired. Check if we have more in the buffer
		var symkeyids []string
		err = rw.Client.rpc.Call(&symkeyids, "pss_getSymmetricKeys", rw.pubKeyId, rw.hextopic)
		if err != nil {
			return err
		}
		// set the rw's point to the next key in the buffer
		var retries int
		var sync bool
		if len(symkeyids) > 0 {
			rw.symKeyId = &symkeyids[0]
		} else {
			retries = handshakeRetryCount
			sync = true
		}
		// initiate handshake
		keyid, err := rw.handshake(retries, sync)
		if err != nil {
			return err
		}
		if len(symkeyids) == 0 {
			rw.symKeyId = &keyid
		}
	}
	return rw.Client.rpc.Call(nil, "pss_sendSym", *rw.symKeyId, rw.hextopic, pmsg)
}

// retry and synchronicity wrapper for handshake api call
// returns first new symkeyid upon successful execution
func (rw *pssRPCRW) handshake(retries int, sync bool) (string, error) {

	var symkeyids []string
	var i int
	// request new keys
	// if the key buffer was depleted, make this as a blocking call and try several times before giving up
	for i = 0; i < 1+retries; i++ {
		log.Debug("handshake attempt pssrpcrw", "pubkeyid", rw.pubKeyId, "topic", rw.hextopic, "sync", sync)
		err := rw.Client.rpc.Call(&symkeyids, "pss_handshake", rw.pubKeyId, rw.hextopic, rw.addr, sync)
		if err == nil {
			var keyid string
			if sync {
				keyid = symkeyids[0]
			}
			return keyid, nil
		}
		if i-1+retries > 1 {
			time.Sleep(time.Millisecond * handshakeRetryTimeout)
		}
	}

	return "", errors.New(fmt.Sprintf("handshake failed after %d attempts", i))
}

func NewClient(rpcurl string) (*Client, error) {
	rpcclient, err := rpc.Dial(rpcurl)
	if err != nil {
		return nil, err
	}

	client, err := NewClientWithRPC(rpcclient)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Constructor for test implementations
// The 'rpcclient' parameter allows passing a in-memory rpc client to act as the remote websocket RPC.
func NewClientWithRPC(rpcclient *rpc.Client) (*Client, error) {
	client := newClient()
	client.rpc = rpcclient
	err := client.rpc.Call(&client.BaseAddr, "pss_baseAddr")
	if err != nil {
		return nil, fmt.Errorf("cannot get pss node baseaddress: %v", err)
	}
	return client, nil
}

func newClient() (client *Client) {
	client = &Client{
		msgC:     make(chan pss.APIMsg),
		quitC:    make(chan struct{}),
		peerPool: make(map[whisper.TopicType]map[string]*pssRPCRW),
		protos:   make(map[whisper.TopicType]*p2p.Protocol),
	}
	return
}

// Mounts a new devp2p protcool on the pss connection
//
// the protocol is aliased as a "pss topic"
// uses normal devp2p Send and incoming message handler routines from the p2p/protocols package
//
// when an incoming message is received from a peer that is not yet known to the client, this peer object is instantiated, and the protocol is run on it.
func (self *Client) RunProtocol(ctx context.Context, proto *p2p.Protocol) error {
	topic := whisper.BytesToTopic([]byte(fmt.Sprintf("%s:%d", proto.Name, proto.Version)))
	hextopic := fmt.Sprintf("%x", topic)
	msgC := make(chan pss.APIMsg)
	self.peerPool[topic] = make(map[string]*pssRPCRW)
	sub, err := self.rpc.Subscribe(ctx, "pss", msgC, "receive", hextopic)
	if err != nil {
		return fmt.Errorf("pss event subscription failed: %v", err)
	}
	self.sub = sub

	// dispatch incoming messages
	go func() {
		for {
			select {
			case msg := <-msgC:
				// we only allow sym msgs here
				if msg.Asymmetric {
					continue
				}
				// we get passed the symkeyid
				// need the symkey itself to resolve to peer's pubkey
				var pubkeyid string
				err = self.rpc.Call(&pubkeyid, "pss_getHandshakePublicKey", msg.Key)
				if err != nil || pubkeyid == "" {
					log.Trace("proto err or no pubkey", "err", err, "symkeyid", msg.Key)
					continue
				}
				// if we don't have the peer on this protocol already, create it
				// this is more or less the same as AddPssPeer, less the handshake initiation
				if self.peerPool[topic][pubkeyid] == nil {
					var addr pss.PssAddress
					err := self.rpc.Call(&addr, "pss_getAddress", hextopic, false, msg.Key)
					if err != nil {
						log.Trace("no addr")
						continue
					}
					rw := self.newpssRPCRW(crypto.ToECDSAPub(common.FromHex(pubkeyid)), addr, &topic)
					self.peerPool[topic][pubkeyid] = rw
					nid, _ := discover.HexID("0x00")
					p := p2p.NewPeer(nid, fmt.Sprintf("%v", addr), []p2p.Cap{})
					go proto.Run(p, self.peerPool[topic][pubkeyid])
				}
				go func() {
					self.peerPool[topic][pubkeyid].msgC <- msg.Msg
				}()
			case <-self.quitC:
				return
			}
		}
	}()

	self.protos[topic] = proto
	return nil
}

// Always call this to ensure that we exit cleanly
func (self *Client) Stop() error {
	return nil
}

// Preemptively add a remote pss peer
func (self *Client) AddPssPeer(key *ecdsa.PublicKey, addr []byte, spec *protocols.Spec) error {
	pubkeyid := common.ToHex(crypto.FromECDSAPub(key))
	topic := ProtocolTopic(spec)
	if self.peerPool[topic] == nil {
		return errors.New("addpeer on unset topic")
	}
	if self.peerPool[topic][pubkeyid] == nil {
		rw := self.newpssRPCRW(key, addr, &topic)
		symkeyid, err := rw.handshake(handshakeRetryCount, true)
		rw.symKeyId = &symkeyid
		if err != nil {
			return err
		}
		self.peerPool[topic][pubkeyid] = rw
		nid, _ := discover.HexID("0x00")
		p := p2p.NewPeer(nid, fmt.Sprintf("%v", addr), []p2p.Cap{})
		go self.protos[topic].Run(p, self.peerPool[topic][pubkeyid])
	}
	return nil
}

// Remove a remote pss peer
//
// Note this doesn't actually currently drop the peer, but only remmoves the reference from the client's peer lookup table
func (self *Client) RemovePssPeer(pubkeyid string, spec *protocols.Spec) {
	topic := ProtocolTopic(spec)
	delete(self.peerPool[topic], pubkeyid)
}

// Uniform translation of protocol specifiers to topic
func ProtocolTopic(spec *protocols.Spec) whisper.TopicType {
	return whisper.BytesToTopic([]byte(fmt.Sprintf("%s:%d", spec.Name, spec.Version)))
}
