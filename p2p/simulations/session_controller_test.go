package simulations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	"github.com/ethereum/go-ethereum/rpc"
)

/***
 * \todo rewrite this with a scripting engine to do http protocol xchanges more easily
 */
const (
	domain = "http://localhost"
	port   = "8888"
)

var quitc chan bool
var controller *ResourceController
var netctrl *Network 

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlError, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))

  conf := &NetworkConfig {
    Id: "0",
		Backend: false,
	  DefaultMockerConfig: DefaultMockerConfig(),
  }
  netctrl = NewNetwork(conf)
  netctrl.SetNaf(naf)
  netctrl.SetInit(func(ids []*adapters.NodeId) {})
	controller, quitc = RunDefaultNet(netctrl)
	StartRestApiServer(port, controller)
}

func naf(conf *NodeConfig) adapters.NodeAdapter {
  id := conf.Id
  node := &testnode{}
	return adapters.NewSimNode(id, node, netctrl)
}

func url(port, path string) string {
	return fmt.Sprintf("%v:%v/%v", domain, port, path)
}

func TestDelete(t *testing.T) {
	req, err := http.NewRequest("DELETE", url(port, ""), nil)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	var resp *http.Response
	go func() {
		r, err := (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		resp = r
	}()
	timeout := time.NewTimer(1000 * time.Millisecond)
	select {
	case <-quitc:
	case <-timeout.C:
		t.Fatalf("timed out: controller did not quit, response: %v", resp)
	}
}

func TestCreate(t *testing.T) {
	s, err := json.Marshal(&struct{ Id string }{Id: "testnetwork"})
	req, err := http.NewRequest("POST", domain+":"+port, bytes.NewReader(s))
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	req, err = http.NewRequest("POST", domain+":"+port+"/testnetwork/debug/", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
}

func TestNodes(t *testing.T) {
	networkname := "testnetworkfornodes"

	s, err := json.Marshal(&struct{ Id string }{Id: networkname})
	req, err := http.NewRequest("POST", domain+":"+port, bytes.NewReader(s))
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	_, err = (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	for i := 0; i < 3; i++ {
		req, err = http.NewRequest("POST", domain+":"+port+"/"+networkname+"/node/", nil)
		if err != nil {
			t.Fatalf("unexpected error creating request: %v", err)
		}
		_, err = (&http.Client{}).Do(req)
		if err != nil {
			t.Fatalf("unexpected error on http.Client request: %v", err)
		}
	}
}

func testResponse(t *testing.T, method, addr string, r io.ReadSeeker) []byte {

	req, err := http.NewRequest(method, addr, r)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("unexpected error on http.Client request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body: %v", err)
	}
	return body

}

func TestUpdate(t *testing.T) {
	t.Skip("...")
	mc := NewNetworkController(netctrl)
	controller.SetResource(netctrl.Config().Id, mc)
	exp := `{
  "add": [
    {
      "data": {
        "id": "aa7c",
        "up": true
      },
      "group": "nodes"
    },
    {
      "data": {
        "id": "f5ae",
        "up": true
      },
      "group": "nodes"
    }
  ],
  "remove": [],
  "message": []
}`
	s, _ := json.Marshal(&SimConfig{})
	resp := testResponse(t, "GET", url(port, "0"), bytes.NewReader(s))
	if string(resp) != exp {
		t.Fatalf("incorrect response body. got\n'%v', expected\n'%v'", string(resp), exp)
	}
}

func createConfigFromId(id *adapters.NodeId) *NodeConfig {
	key, err := crypto.GenerateKey()
	if err != nil {
		panic("unable to generate key")
	}
	return &NodeConfig{
		Id:         id,
		PrivateKey: key,
	}
}

func mockNewNodes(eventer *event.TypeMux, ids []*adapters.NodeId) {
	log.Trace("mock starting")
	for _, id := range ids {
		log.Trace(fmt.Sprintf("mock adding node %v", id))
		conf := createConfigFromId(id)
		node := &Node{NodeConfig: *conf, Up: true}
		eventer.Post(node.EmitEvent(LiveEvent))
	}
}

type testnode struct {
	run func(*p2p.Peer, p2p.MsgReadWriter) error
}

func (n *testnode) Protocols() []p2p.Protocol {
	return []p2p.Protocol{{Run: n.run}}
}

func (n *testnode) APIs() []rpc.API {
  return nil
}

func (n *testnode) Start(server p2p.Server) error {
  return nil
}

func (n *testnode) Stop() error {
  return nil
}

func (n *testnode) Info() string {
  return "" 
}
