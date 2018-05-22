package notify

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/state"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

var (
	loglevel = flag.Int("l", 3, "loglevel")
	psses    []*pss.Pss
	w        *whisper.Whisper
	wapi     *whisper.PublicWhisperAPI
	msgSeq   int
)

func init() {
	flag.Parse()
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(log.Lvl(*loglevel), hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)

	w = whisper.New(&whisper.DefaultConfig)
	wapi = whisper.NewPublicWhisperAPI(w)
}

func TestStart(t *testing.T) {
	adapter := adapters.NewSimAdapter(newServices(false))
	net := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "bzz",
	})
	l_nodeconf := adapters.RandomNodeConfig()
	l_nodeconf.Services = []string{"bzz", "pss"}
	l_node, err := net.NewNodeWithConfig(l_nodeconf)
	if err != nil {
		t.Fatal(err)
	}
	err = net.Start(l_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	r_nodeconf := adapters.RandomNodeConfig()
	r_nodeconf.Services = []string{"bzz", "pss"}
	r_node, err := net.NewNodeWithConfig(r_nodeconf)
	if err != nil {
		t.Fatal(err)
	}
	err = net.Start(r_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	err = net.Connect(r_node.ID(), l_node.ID())
	if err != nil {
		t.Fatal(err)
	}

	l_rpc, err := l_node.Client()
	if err != nil {
		t.Fatal(err)
	}

	r_rpc, err := r_node.Client()
	if err != nil {
		t.Fatal(err)
	}

	var l_addr string
	err = l_rpc.Call(&l_addr, "pss_baseAddr")
	if err != nil {
		t.Fatal(err)
	}

	var r_addr string
	err = r_rpc.Call(&r_addr, "pss_baseAddr")
	if err != nil {
		t.Fatal(err)
	}

	var l_pub string
	err = l_rpc.Call(&l_pub, "pss_getPublicKey")
	if err != nil {
		t.Fatal(err)
	}

	err = r_rpc.Call(nil, "pss_setPeerPublicKey", l_pub, controlTopic, l_addr)
	if err != nil {
		t.Fatal(err)
	}

	rsrcName := "foo.eth"
	rsrcTopic := pss.BytesToTopic([]byte(rsrcName))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	rmsgC := make(chan *pss.APIMsg)
	r_sub, err := r_rpc.Subscribe(ctx, "pss", rmsgC, "receive", controlTopic)
	if err != nil {
		t.Fatal(err)
	}
	defer r_sub.Unsubscribe()
	r_sub_update, err := r_rpc.Subscribe(ctx, "pss", rmsgC, "receive", rsrcTopic)
	if err != nil {
		t.Fatal(err)
	}
	defer r_sub_update.Unsubscribe()

	updateMsg := []byte("xyzzy")
	ctrl := NewController(psses[0])
	ctrl.NewNotifier("foo.eth", 2, func(name string) ([]byte, error) {
		msgSeq++
		return updateMsg, nil
	})

	msg := NewMsg(MsgCodeStart, rsrcName, common.FromHex(r_addr))
	smsg, err := rlp.EncodeToBytes(msg)
	if err != nil {
		t.Fatal(err)
	}
	err = r_rpc.Call(nil, "pss_sendAsym", l_pub, controlTopic, common.ToHex(smsg))
	if err != nil {
		t.Fatal(err)
	}

	var inMsg *pss.APIMsg
	select {
	case inMsg = <-rmsgC:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	dMsg := &Msg{}
	err = rlp.DecodeBytes(inMsg.Msg, dMsg)
	if err != nil {
		t.Fatal(err)
	} else if dMsg.GetName() != rsrcName {
		t.Fatalf("expected name %s, got %s", rsrcName, dMsg.GetName())
	} else if !bytes.Equal(dMsg.Payload[:len(updateMsg)], updateMsg) {
		t.Fatalf("expected payload first %d bytes '%x', got '%x'", len(updateMsg), updateMsg, dMsg.Payload[:len(updateMsg)])
	} else if len(updateMsg)+symKeyLength != len(dMsg.Payload) {
		t.Fatalf("expected payload length %d, have %d", len(updateMsg)+symKeyLength, len(dMsg.Payload))
	}

	l_pssAddr := pss.PssAddress(common.FromHex(l_addr))
	psses[1].SetSymmetricKey(dMsg.Payload[len(updateMsg):], rsrcTopic, &l_pssAddr, true)

	nextUpdateMsg := []byte("plugh")
	ctrl.Notify(rsrcName, nextUpdateMsg)
	select {
	case inMsg = <-rmsgC:
	case <-ctx.Done():
		log.Error("timed out waiting for msg", "topic", fmt.Sprintf("%x", rsrcTopic))
		t.Fatal(ctx.Err())
	}
	dMsg = &Msg{}
	err = rlp.DecodeBytes(inMsg.Msg, dMsg)
	if err != nil {
		t.Fatal(err)
	} else if dMsg.GetName() != rsrcName {
		t.Fatalf("expected name %s, got %s", rsrcName, dMsg.GetName())
	} else if !bytes.Equal(dMsg.Payload, nextUpdateMsg) {
		t.Fatalf("expected payload '%x', got '%x'", nextUpdateMsg, dMsg.Payload)
	}

}

func newServices(allowRaw bool) adapters.Services {
	stateStore := state.NewInmemoryStore()
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
			ctxlocal, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			keys, err := wapi.NewKeyPair(ctxlocal)
			privkey, err := w.GetPrivateKey(keys)
			pssp := pss.NewPssParams().WithPrivateKey(privkey)
			pssp.MsgTTL = time.Second * 30
			pssp.AllowRaw = allowRaw
			pskad := kademlia(ctx.Config.ID)
			ps, err := pss.NewPss(pskad, pssp)
			if err != nil {
				return nil, err
			}
			psses = append(psses, ps)
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
			return network.NewBzz(config, kademlia(ctx.Config.ID), stateStore, nil, nil), nil
		},
	}
}
