package client

import (
	"context"
	"flag"
	"fmt"
	//"net"
	//"net/http"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

const (
	pssServiceName = "pss"
	bzzServiceName = "bzz"
)

type protoCtrl struct {
	C        chan struct{}
	protocol *pss.PssProtocol
	run      func(*p2p.Peer, p2p.MsgReadWriter) error
}

var (
	debugdebugflag = flag.Bool("vv", false, "veryverbose")
	debugflag      = flag.Bool("v", false, "verbose")
	w              *whisper.Whisper
	wapi           *whisper.PublicWhisperAPI
	// custom logging
	psslogmain   log.Logger
	pssprotocols map[string]*protoCtrl
)

var services = newServices()

func init() {
	flag.Parse()
	rand.Seed(time.Now().Unix())

	adapters.RegisterServices(services)

	loglevel := log.LvlInfo
	if *debugflag {
		loglevel = log.LvlDebug
	} else if *debugdebugflag {
		loglevel = log.LvlTrace
	}

	psslogmain = log.New("psslog", "*")
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)

	w = whisper.New()
	wapi = whisper.NewPublicWhisperAPI(w)

	pssprotocols = make(map[string]*protoCtrl)
}

func TestHandshake(t *testing.T) {
	topic := ProtocolTopic(pss.PingProtocol)
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	lpsc, err := NewClientWithRPC(clients[0])
	if err != nil {
		t.Fatal(err)
	}
	rpsc, err := NewClientWithRPC(clients[1])
	if err != nil {
		t.Fatal(err)
	}
	lpssping := &pss.Ping{
		OutC: make(chan struct{}),
		InC:  make(chan struct{}),
	}
	rpssping := &pss.Ping{
		OutC: make(chan struct{}),
		InC:  make(chan struct{}),
	}
	lproto := pss.NewPingProtocol(lpssping.OutC, lpssping.PingHandler)
	rproto := pss.NewPingProtocol(rpssping.OutC, rpssping.PingHandler)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	err = lpsc.RunProtocol(ctx, lproto)
	if err != nil {
		t.Fatal(err)
	}
	err = rpsc.RunProtocol(ctx, rproto)
	if err != nil {
		t.Fatal(err)
	}
	loaddr := make([]byte, 32)
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	roaddr := make([]byte, 32)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}

	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, roaddr)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, loaddr)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	err = lpsc.AddPssPeer(crypto.ToECDSAPub(rpubkey), roaddr, pss.PingProtocol)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	lpssping.OutC <- struct{}{}

	<-rpssping.InC
	log.Warn("ok")
}

//func TestRunProtocol(t *testing.T) {
//	quitC := make(chan struct{})
//	ps := pss.NewTestPss(nil)
//	ping := &pss.Ping{
//		C: make(chan struct{}),
//	}
//	proto := newProtocol(ping)
//	_, err := baseTester(t, proto, ps, nil, quitC)
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	quitC <- struct{}{}
//}
//
//func TestIncoming(t *testing.T) {
//	t.Skip("pssclient is broken, needs whisper integration")
//	quitC := make(chan struct{})
//	ps := pss.NewTestPss(nil)
//	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
//	var addr []byte
//	ping := &pss.Ping{
//		C: make(chan struct{}),
//	}
//	proto := newProtocol(ping)
//	client, err := baseTester(t, proto, ps, ctx, quitC)
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//
//	client.rpc.Call(&addr, "psstest_baseAddr")
//
//	code, _ := pss.PingProtocol.GetCode(&pss.PingMsg{})
//	rlpbundle, err := pss.NewProtocolMsg(code, &pss.PingMsg{
//		Created: time.Now(),
//	})
//	if err != nil {
//		t.Fatalf("couldn't make pssmsg: %v", err)
//	}
//
//	_ = rlpbundle
//	pssenv := &whisper.Envelope{}
//	pssmsg := pss.PssMsg{
//		To:      addr,
//		Payload: pssenv,
//	}
//
//	ps.Process(&pssmsg)
//
//	<-ping.C
//
//	quitC <- struct{}{}
//}
//
//func TestOutgoing(t *testing.T) {
//	t.Skip("pssclient is broken, needs whisper integration")
//	quitC := make(chan struct{})
//	ps := pss.NewTestPss(nil)
//	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*250)
//	var addr []byte
//	var potaddr pot.Address
//
//	ping := &pss.Ping{
//		C: make(chan struct{}),
//	}
//	proto := newProtocol(ping)
//	client, err := baseTester(t, proto, ps, ctx, quitC)
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//
//	client.rpc.Call(&addr, "psstest_baseAddr")
//	copy(potaddr[:], addr)
//
//	msg := &pss.PingMsg{
//		Created: time.Now(),
//	}
//	topic := whisper.BytesToTopic([]byte(fmt.Sprintf("%s:%d", pss.PingProtocol.Name, pss.PingProtocol.Version)))
//	client.AddPssPeer(potaddr, pss.PingProtocol)
//	nid, _ := discover.HexID("0x00")
//	p := p2p.NewPeer(nid, fmt.Sprintf("%v", potaddr), []p2p.Cap{})
//	pp := protocols.NewPeer(p, client.peerPool[topic][potaddr], pss.PingProtocol)
//	pp.Send(msg)
//	<-ping.C
//	quitC <- struct{}{}
//}
//
//func baseTester(t *testing.T, proto *p2p.Protocol, ps *pss.Pss, ctx context.Context, quitC chan struct{}) (*Client, error) {
//	var err error
//
//	client := newTestclient(t, quitC)
//
//	err = client.RunProtocol(context.Background(), proto)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return client, nil
//}
//
//func newProtocol(ping *pss.Ping) *p2p.Protocol {
//
//	return &p2p.Protocol{
//		Name:    pss.PingProtocol.Name,
//		Version: pss.PingProtocol.Version,
//		Length:  1,
//		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
//			pp := protocols.NewPeer(p, rw, pss.PingProtocol)
//			pp.Run(ping.PingHandler)
//			return nil
//		},
//	}
//}
//
//
//func newTestclient(t *testing.T, quitC chan struct{}) *Client {
//
//	ps := pss.NewTestPss(nil)
//	srv := rpc.NewServer()
//	srv.RegisterName("pss", pss.NewAPI(ps))
//	srv.RegisterName("psstest", pss.NewAPITest(ps))
//	ws := srv.WebsocketHandler([]string{"*"})
//	uri := fmt.Sprintf("%s:%d", "localhost", 8546)
//
//	sock, err := net.Listen("tcp", uri)
//	if err != nil {
//		t.Fatalf("Tcp (recv) on %s failed: %v", uri, err)
//	}
//
//	go func() {
//		http.Serve(sock, ws)
//	}()
//
//	go func() {
//		<-quitC
//		sock.Close()
//	}()
//
//	pssclient, err := NewClient("ws://localhost:8546")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//
//	return pssclient
//}

func setupNetwork(numnodes int) (clients []*rpc.Client, err error) {
	nodes := make([]*simulations.Node, numnodes)
	clients = make([]*rpc.Client, numnodes)
	if numnodes < 2 {
		return nil, fmt.Errorf("Minimum two nodes in network")
	}
	adapter := adapters.NewSimAdapter(services)
	net := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "bzz",
	})
	for i := 0; i < numnodes; i++ {
		nodes[i], err = net.NewNodeWithConfig(&adapters.NodeConfig{
			Services: []string{"bzz", "pss"},
		})
		if err != nil {
			return nil, fmt.Errorf("error creating node 1: %v", err)
		}
		err = net.Start(nodes[i].ID())
		if err != nil {
			return nil, fmt.Errorf("error starting node 1: %v", err)
		}
		if i > 0 {
			err = net.Connect(nodes[i].ID(), nodes[i-1].ID())
			if err != nil {
				return nil, fmt.Errorf("error connecting nodes: %v", err)
			}
		}
		clients[i], err = nodes[i].Client()
		if err != nil {
			return nil, fmt.Errorf("create node 1 rpc client fail: %v", err)
		}
	}
	if numnodes > 2 {
		err = net.Connect(nodes[0].ID(), nodes[len(nodes)-1].ID())
		if err != nil {
			return nil, fmt.Errorf("error connecting first and last nodes")
		}
	}
	return clients, nil
}

func newServices() adapters.Services {
	stateStore := adapters.NewSimStateStore()
	kademlias := make(map[discover.NodeID]*network.Kademlia)
	kademlia := func(id discover.NodeID) *network.Kademlia {
		if k, ok := kademlias[id]; ok {
			return k
		}
		addr := network.NewAddrFromNodeID(id)
		params := network.NewKadParams()
		params.MinProxBinSize = 2
		params.MaxBinSize = 3
		params.MinBinSize = 1
		params.MaxRetries = 1000
		params.RetryExponent = 2
		params.RetryInterval = 1000000
		kademlias[id] = network.NewKademlia(addr.Over(), params)
		return kademlias[id]
	}
	return adapters.Services{
		"pss": func(ctx *adapters.ServiceContext) (node.Service, error) {
			cachedir, err := ioutil.TempDir("", "pss-cache")
			if err != nil {
				return nil, fmt.Errorf("create pss cache tmpdir failed", "error", err)
			}
			dpa, err := storage.NewLocalDPA(cachedir)
			if err != nil {
				return nil, fmt.Errorf("local dpa creation failed", "error", err)
			}

			keys, err := wapi.NewKeyPair()
			privkey, err := w.GetPrivateKey(keys)
			pssp := pss.NewPssParams(privkey)
			pskad := kademlia(ctx.Config.ID)
			ps := pss.NewPss(pskad, dpa, pssp)

			ping := &pss.Ping{
				OutC: make(chan struct{}),
				InC:  make(chan struct{}),
			}
			p2pp := pss.NewPingProtocol(ping.OutC, ping.PingHandler)
			pp, err := pss.RegisterPssProtocol(ps, &pss.PingTopic, pss.PingProtocol, p2pp, &pss.PssProtocolOptions{Asymmetric: true})
			if err != nil {
				return nil, err
			}
			ps.Register(&pss.PingTopic, pp.Handle)
			if err != nil {
				log.Error("Couldnt register pss protocol", "err", err)
				os.Exit(1)
			}
			pssprotocols[ctx.Config.ID.String()] = &protoCtrl{
				C:        ping.OutC,
				protocol: pp,
				run:      p2pp.Run,
			}
			return ps, nil
		},
		"bzz": func(ctx *adapters.ServiceContext) (node.Service, error) {
			addr := network.NewAddrFromNodeID(ctx.Config.ID)
			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			return network.NewBzz(config, kademlia(ctx.Config.ID), stateStore), nil
		},
	}
}
