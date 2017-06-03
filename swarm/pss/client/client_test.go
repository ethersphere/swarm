package client

import (
	"context"
	"fmt"
	"os"
	"net"
	"net/http"
	"testing"
	"time"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/protocols"
	"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rpc"
	//"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/pss"
	pssapi "github.com/ethereum/go-ethereum/swarm/pss/api"
)

func init() {
	h := log.CallerFileHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(true)))
	log.Root().SetHandler(h)
}

func TestRunProtocol(t *testing.T) {
	quitC := make(chan struct{})
	ps := newTestPss(nil)
	ping := &pss.PssPing{
		QuitC: make(chan struct{}),
	}
	proto := newProtocol(ping)	
	_, err := baseTester(t, proto, ps, nil, nil, quitC)
	if err != nil {
		t.Fatalf(err.Error())
	}
	quitC <- struct{}{}
}

func TestIncoming(t *testing.T) {
	quitC := make(chan struct{})
	ps := newTestPss(nil)
	ctx, cancel := context.WithCancel(context.Background())
	var addr []byte
	ping := &pss.PssPing{
		QuitC: make(chan struct{}),
	}
	proto := newProtocol(ping)	
	client, err := baseTester(t, proto, ps, ctx, cancel, quitC)
	if err != nil {
		t.Fatalf(err.Error())
	}
	
	client.ws.Call(&addr, "pss_baseAddr")

	code, _ := pss.PssPingProtocol.GetCode(&pss.PssPingMsg{})
	rlpbundle, err := pss.NewProtocolMsg(code, &pss.PssPingMsg{
		Created: time.Now(),
	})
	if err != nil {
		t.Fatalf("couldn't make pssmsg")
	}

	pssenv := pss.PssEnvelope{
		From: addr,
		Topic:       pss.NewTopic(proto.Name, int(proto.Version)),
		TTL:         pss.DefaultTTL,
		Payload:     rlpbundle,
	}
	pssmsg := pss.PssMsg{
		To: addr,
		Payload: &pssenv,
	}
	
	ps.Process(&pssmsg)
	
	go func() {
		<-ping.QuitC
		client.cancel()
	}()
	
	<-client.ctx.Done()
	quitC <- struct{}{}
}

func TestOutgoing(t *testing.T) {
	quitC := make(chan struct{})
	ps := newTestPss(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 250)
	var addr []byte
	var potaddr pot.Address
	
	ping := &pss.PssPing{
		QuitC: make(chan struct{}),
	}
	proto := newProtocol(ping)	
	client, err := baseTester(t, proto, ps, ctx, cancel, quitC)
	if err != nil {
		t.Fatalf(err.Error())
	}
	
	client.ws.Call(&addr, "pss_baseAddr")
	copy(potaddr[:], addr)
					
	msg := &pss.PssPingMsg{
		Created: time.Now(),
	}
	
	topic := pss.NewTopic(pss.PssPingProtocol.Name, int(pss.PssPingProtocol.Version))
	client.AddPssPeer(potaddr, pss.PssPingProtocol)
	nid, _ := discover.HexID("0x00")
	p := p2p.NewPeer(nid, fmt.Sprintf("%v", potaddr), []p2p.Cap{})
	pp := protocols.NewPeer(p, client.peerPool[topic][potaddr], pss.PssPingProtocol)
	pp.Send(msg)
	<-client.ctx.Done()
	quitC <- struct{}{}
}

func baseTester(t *testing.T, proto *p2p.Protocol, ps pss.PssAdapter, ctx context.Context, cancel func(), quitC chan struct{}) (*PssClient, error) {
	var err error
	
	client := newClient(t, ctx, cancel, quitC)
	
	err = client.Start()
	if err != nil {
		return nil, err
	}
	
	//err = client.RunProtocol(proto, pssPingProtocol)
	err = client.RunProtocol(proto)
	
	if err != nil {
		return nil, err
	}
	
	return client, nil
}

func newProtocol(ping *pss.PssPing) *p2p.Protocol {
	
	return &p2p.Protocol{
		Name: pss.PssPingProtocol.Name,
		Version: pss.PssPingProtocol.Version,
		Length: 1,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			pp := protocols.NewPeer(p, rw, pss.PssPingProtocol)
			pp.Run(ping.PssPingHandler)
			return nil
		},
	}
}

func newClient(t *testing.T, ctx context.Context, cancel func(), quitC chan struct{}) *PssClient {
	
	conf := &PssClientConfig{
	}
	
	pssclient := NewPssClient(ctx, cancel, conf)
	
	ps := newTestPss([]byte{0})
	srv := rpc.NewServer()
	srv.RegisterName("pss", pssapi.NewPssAPI(ps))
	ws := srv.WebsocketHandler([]string{"*"})
	uri := fmt.Sprintf("%s:%d", "localhost", 8546)
	
	sock, err := net.Listen("tcp", uri)
	if err != nil {
		t.Fatalf("Tcp (recv) on %s failed: %v", uri, err)
	}
	
	go func() {
		http.Serve(sock, ws)
	}()

	go func() {
		<-quitC
		sock.Close()
	}()	
	return pssclient
} 


func newTestPss(addr []byte) pss.PssAdapter {	
	return &testPss{
		addr: addr,
	}
}

type testPss struct {
	addr []byte
}

func (self *testPss) Send (to []byte, topic pss.PssTopic, msg []byte) error {
	return nil
}

func (self *testPss) Register(topic *pss.PssTopic, handler pss.PssHandler) func() {
	return func() {return}
}

func (self *testPss) BaseAddr() []byte {
	return self.addr
}

func (self *testPss) Start(srv *p2p.Server) error {
	return nil
}

func (self *testPss) Stop() error {
	return nil
}

func (self *testPss) Protocols() []p2p.Protocol {
	return nil
}

func (self *testPss) APIs() []rpc.API {
	return nil
}

func (self *testPss) Process(pssmsg *pss.PssMsg) error {
	return nil
}
