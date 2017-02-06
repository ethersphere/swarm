package main

import (	
	"bytes"
	"testing"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

type sessionrequestpayload interface {
}

type sessionrequest struct {
	method string
	url string
	payload sessionrequestpayload
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
		
		sessionrequest{method: "GET", url: "/testnet/node/", payload: nil,},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 1},},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint}{One: 2},},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
			Other uint}{One: 1, Other: 2},},
		
		sessionrequest{method: "PUT", url: "/testnet/node/", payload: &struct{One uint
			Other uint
			AssetType uint8}{One: 1, Other: 2, AssetType: 1},},
		
		sessionrequest{method: "POST", url: "/testnet/debug/", payload: nil,},

	}
	
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
