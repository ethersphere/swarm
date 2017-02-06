# META-wire network testing

Use `go test -v .` (uses overlay_meta_test.go) for testing.

Needs manual start of `go run -v overlay_meta.go` before test run (of course it shouldn't need that => TODO)

## Test run should output something like

### HTTP

```
$ go test -v .
=== RUN   TestMETASession
--- PASS: TestMETASession (0.01s)
	overlay_meta_test.go:74: ***** SENT 'POST /'
		***** GOT:
		
		{}
		
	overlay_meta_test.go:74: ***** SENT 'POST /testnet/node/'
		***** GOT:
		
		{
		  "Id": "dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633"
		}
		
	overlay_meta_test.go:74: ***** SENT 'POST /testnet/node/'
		***** GOT:
		
		{
		  "Id": "b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c"
		}
		
	overlay_meta_test.go:74: ***** SENT 'POST /testnet/node/'
		***** GOT:
		
		{
		  "Id": "97f135279c3fff8408d7139fda0ed16b662f3786f7519412f7f0a4399cc04ebe86e294e1fb0008b94eedb45daf82109dbf072250fe47e681526aede45bdd9d6d"
		}
		
	overlay_meta_test.go:74: ***** SENT 'GET /testnet/node/'
		***** GOT:
		
		{
		  "Nodes": [
		    {
		      "id": "dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633",
		      "Up": false
		    },
		    {
		      "id": "b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c",
		      "Up": false
		    },
		    {
		      "id": "97f135279c3fff8408d7139fda0ed16b662f3786f7519412f7f0a4399cc04ebe86e294e1fb0008b94eedb45daf82109dbf072250fe47e681526aede45bdd9d6d",
		      "Up": false
		    }
		  ]
		}
		
	overlay_meta_test.go:74: ***** SENT 'PUT /testnet/node/'
		***** GOT:
		
		{
		  "Nodes": [
		    {
		      "id": "dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633",
		      "Up": true
		    }
		  ]
		}
		
	overlay_meta_test.go:74: ***** SENT 'PUT /testnet/node/'
		***** GOT:
		
		{
		  "Nodes": [
		    {
		      "id": "b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c",
		      "Up": true
		    }
		  ]
		}
		
	overlay_meta_test.go:74: ***** SENT 'PUT /testnet/node/'
		***** GOT:
		
		{
		  "Nodes": [
		    {
		      "id": "dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633",
		      "Up": true
		    },
		    {
		      "id": "b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c",
		      "Up": true
		    }
		  ]
		}
		
	overlay_meta_test.go:74: ***** SENT 'PUT /testnet/node/'
		***** GOT:
		
		{}
		
	overlay_meta_test.go:74: ***** SENT 'POST /testnet/debug/'
		***** GOT:
		
		{
		  "Results": [
		    "\u0026{2017-02-06 23:58:21.436257963 +0100 CET \u003cAction: up, Type: node, Data: Node dd43\u003e\n}",
		    "\u0026{2017-02-06 23:58:21.436715343 +0100 CET \u003cAction: up, Type: node, Data: Node b5dd\u003e\n}",
		    "\u0026{2017-02-06 23:58:21.437251778 +0100 CET \u003cAction: up, Type: conn, Data: Conn dd43-\u003eb5dd\u003e\n}",
		    "\u0026{2017-02-06 23:58:21.437854418 +0100 CET \u003cAction: 0, Type: msg, From: Msg(0) dd43-\u003eb5dd\u003e\n}",
		    "\u0026{2017-02-06 23:58:21.437935524 +0100 CET \u003cAction: 0, Type: msg, From: Msg(0) dd43-\u003eb5dd\u003e\n}"
		  ]
		}
		
PASS
ok  	github.com/ethereum/go-ethereum/META/network/simulations	0.017s

```

### backend log

```
$ go run -v overlay_meta.go
[...]
I0206 23:58:19.384392 p2p/simulations/rest_api_server.go:32] Swarm Network Controller HTTP server started on localhost:8888
I0206 23:58:21.432519 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/', Host: '', Path: '/', Referer: '', Accept: ''
I0206 23:58:21.432803 p2p/simulations/journal.go:44] subscribe
I0206 23:58:21.432893 META/network/simulations/overlay_meta.go:95] new network controller on testnet
I0206 23:58:21.433480 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.433783 p2p/simulations/network.go:285] node dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633 created
I0206 23:58:21.433865 META/network/simulations/overlay_meta.go:131] added node dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633 to network testnet
I0206 23:58:21.434151 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.434404 p2p/simulations/network.go:285] node b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c created
I0206 23:58:21.434423 META/network/simulations/overlay_meta.go:131] added node b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c to network testnet
I0206 23:58:21.434744 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.435010 p2p/simulations/network.go:285] node 97f135279c3fff8408d7139fda0ed16b662f3786f7519412f7f0a4399cc04ebe86e294e1fb0008b94eedb45daf82109dbf072250fe47e681526aede45bdd9d6d created
I0206 23:58:21.435026 META/network/simulations/overlay_meta.go:131] added node 97f135279c3fff8408d7139fda0ed16b662f3786f7519412f7f0a4399cc04ebe86e294e1fb0008b94eedb45daf82109dbf072250fe47e681526aede45bdd9d6d to network testnet
I0206 23:58:21.435360 p2p/simulations/rest_api_server.go:37] HTTP GET request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.436081 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.436140 p2p/simulations/network.go:333] starting node dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633: false adapter &{{{0 0} 0 0 0 0} dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633 0xc8201a40e0 0x4cc930 map[] [] 0x4c9140}
I0206 23:58:21.436238 p2p/simulations/network.go:342] started node dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633: true
I0206 23:58:21.436626 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.436661 p2p/simulations/network.go:333] starting node b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c: false adapter &{{{0 0} 0 0 0 0} b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c 0xc8201a40e0 0x4cc930 map[] [] 0x4c9140}
I0206 23:58:21.436703 p2p/simulations/network.go:342] started node b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c: true
I0206 23:58:21.437110 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.437171 p2p/adapters/inproc.go:168] protocol starting on peer b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c (connection with dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633)
I0206 23:58:21.437215 p2p/adapters/inproc.go:168] protocol starting on peer dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633 (connection with b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c)
I0206 23:58:21.437344 p2p/protocols/protocol.go:219] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0206 23:58:21.437426 META/network/peercollection.go:33] protopeers now has added peers dd43884c34c0c59d4d3ea96e4f5f5f66e869b4444b60273834a77bf99a205565c6660df184ecd47064b6871fc12e4fdb3916607126678f7d8bdd18bfe544e633, total 1
I0206 23:58:21.437389 p2p/protocols/protocol.go:219] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0206 23:58:21.437578 META/network/peercollection.go:33] protopeers now has added peers b5dddc1a74705384a19b6bda5b004100577c195bb13eb7bc15cf14a9c4c2ed595801dccb12e2ad69b91a3c4bc7203e7d07d71702c91fa7a759a89c0a8422f03c, total 1
I0206 23:58:21.437685 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/testnet/node/', Host: '', Path: '/testnet/node/', Referer: '', Accept: ''
I0206 23:58:21.437759 p2p/protocols/protocol.go:275] <= msg #0 (51 bytes)
I0206 23:58:21.437804 p2p/protocols/protocol.go:297] <= &{0 0000000000000000000000000000000000000000000000000000000000000000 [1 0 0 0 14 208 43 60 221 26 22 240 134 0 60]} *network.METAAssetNotification (0)
I0206 23:58:21.437827 p2p/protocols/protocol.go:310] handler 0 for *network.METAAssetNotification
I0206 23:58:21.437836 META/network/protocol.go:67] Peer received asset notification Eletronic Release Notification
I0206 23:58:21.437865 p2p/protocols/protocol.go:275] <= msg #0 (51 bytes)
I0206 23:58:21.437891 p2p/protocols/protocol.go:297] <= &{0 0000000000000000000000000000000000000000000000000000000000000000 [1 0 0 0 14 208 43 60 221 26 22 240 134 0 60]} *network.METAAssetNotification (0)
I0206 23:58:21.437917 p2p/protocols/protocol.go:310] handler 0 for *network.METAAssetNotification
I0206 23:58:21.437926 META/network/protocol.go:67] Peer received asset notification Eletronic Release Notification
I0206 23:58:21.438516 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/testnet/debug/', Host: '', Path: '/testnet/debug/', Referer: '', Accept: ''
I0206 23:58:21.438731 p2p/simulations/journal.go:195] cursor reset from 5 to 4/5 (5)


```

