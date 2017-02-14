package simulations

import (
	"unsafe"
	"bytes"
	"testing"
	"net/http"
	"encoding/hex"
	"encoding/json"
	"encoding/binary"
	"io/ioutil"
	
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/p2p/adapters"
	
	//"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

func init() {
	glog.SetV(6)
	glog.SetToStderr(true)
}

type sessionrequestpayload interface {
}

type sessionrequest struct {
	method string
	url string
	payload sessionrequestpayload
}


func TestMETATmpName(t *testing.T) {
	hostport := "http://127.0.0.1:8888"
	c := http.Client{}
	strtohash := "fingered"
	strlengthtohash := make([]byte, 8)
	binary.LittleEndian.PutUint64(strlengthtohash, uint64(len(strtohash)))
	
	bytestohash := [][]byte{
		strlengthtohash,
		bytes.NewBufferString(strtohash).Bytes(),
	}
	concatbytestohash := bytes.Join(bytestohash, nil)
	
	t.Logf("concatbytes %v", concatbytestohash)
	
	networkname := "meta"
	hashit := storage.MakeHashFunc("SHA3")()
	
	
	tmpnameupdate := &METANameIF {
		Squealernode: 1,
		Victimnode: hex.EncodeToString((*((*[discover.NodeIDBits / 8]byte)(unsafe.Pointer(adapters.RandomNodeId()))))[:]),
		Name: "fingered",
		Swarmhash: hashit.Sum(concatbytestohash),
	}
	
	reqs := []sessionrequest{
		sessionrequest{method: "POST", url: "/", payload: &struct{Id string}{Id: networkname},},
		sessionrequest{method: "POST", url: "/" + networkname + "/node/", payload: nil,},
		sessionrequest{method: "POST", url: "/" + networkname + "/node/", payload: nil,},
		sessionrequest{method: "PUT", url: "/" + networkname + "/node/", payload: &struct{One uint}{One: 1},},
		sessionrequest{method: "PUT", url: "/" + networkname + "/node/", payload: &struct{One uint}{One: 2},},
		sessionrequest{method: "PUT", url: "/" + networkname + "/node/", payload: &struct{One uint
Other uint}{One: 1, Other: 2,},},
		sessionrequest{method: "POST", url: "/" + networkname + "/node/tmpname/", payload: tmpnameupdate,},
	}
	
	playReqs(t, reqs, hostport, c)
}

func TestMETASession(t *testing.T) {
	
	// need to start up manually, not so good
	// dont have access to input structs, not so good either...
	
	hostport := "http://127.0.0.1:8888"
	c := http.Client{}
	
	
	reqs := []sessionrequest{
		sessionrequest{method: "POST", url: "/", payload: &struct{Id string}{Id: "testnet"},},
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		sessionrequest{method: "POST", url: "/testnet/node/", payload: nil,},	
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 1},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 2},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 3},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 4},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 5},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 6},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 7},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 8},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 9},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 10},},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 1, Other: 2,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 2, Other: 3,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 3, Other: 4,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 4, Other: 5,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 5, Other: 6,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 6, Other: 7,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 7, Other: 8,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 8, Other: 9,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 9, Other: 10,},},
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint}{One: 10, Other: 1,},},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
Other uint
AssetType uint}{One: 1, Other: 2, AssetType: 1}},
		
		sessionrequest{method: "POST", url: "/testnet/debug/", payload: nil,},
	}
	
	playReqs(t, reqs, hostport, c)
	
	/*
	for _, req := range reqs {
		var hresp *http.Response

		p, _ := json.Marshal(req.payload)
		
		hreq,err := http.NewRequest(req.method, hostport + req.url, bytes.NewReader(p))
		if req.method != "GET" {
			hreq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
		hresp, err = c.Do(hreq)		
		if err != nil {
			t.Fatalf("Couldn't %v: %v", req.method, err)
		}
		if hresp.StatusCode != 200 {
			t.Fatalf("'%s %s' failed: %s", req.method, req.url, hresp.Status)
		} 
		
		rbody, err := ioutil.ReadAll(hresp.Body)
		
		t.Logf("***** SENT '%s %s'\n***** GOT:\n\n%s\n\n", req.method, req.url, rbody)
		
		hresp.Body.Close()
	}*/
}

func playReqs(t *testing.T, reqs []sessionrequest, hostport string, c http.Client) {
	for _, req := range reqs {
		var hresp *http.Response

		p, _ := json.Marshal(req.payload)
		
		hreq,err := http.NewRequest(req.method, hostport + req.url, bytes.NewReader(p))
		if req.method != "GET" {
			hreq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
		hresp, err = c.Do(hreq)		
		if err != nil {
			t.Fatalf("Couldn't %v: %v", req.method, err)
		}
		if hresp.StatusCode != 200 {
			t.Fatalf("'%s %s' failed: %s", req.method, req.url, hresp.Status)
		} 
		
		rbody, err := ioutil.ReadAll(hresp.Body)
		
		t.Logf("***** SENT '%s %s'\n***** GOT:\n\n%s\n\n", req.method, req.url, rbody)
		
		hresp.Body.Close()
	}
}
