# META-wire network testing

## Current test commands

(raw HTTP)

```
POST / HTTP/1.0
Connection: keep-alive
Content-Type: application/x-www-form-urlencoded
Content-Length: 30

{"Id":"abcde","NodeAmount":5}

POST /node/ HTTP/1.0
Connection: keep-alive

POST /node/ HTTP/1.0
Connection: keep-alive

POST /node/ HTTP/1.0
Connection: keep-alive

GET /node/ HTTP/1.0

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":0}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":2,"Other":0}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":2}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 34

{"One":1,"Other":2,"AssetType":2}

```

## Test run output

### HTTP

```
$ telnet localhost 8888
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
POST / HTTP/1.0
Connection: keep-alive
Content-Type: application/x-www-form-urlencoded
Content-Length: 30

{"Id":"abcde","NodeAmount":5}

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 2
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:42 GMT
Date: Sun, 05 Feb 2017 13:03:42 GMT
Connection: keep-alive

{}

POST /node/ HTTP/1.0
Connection: keep-alive

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:42 GMT
Date: Sun, 05 Feb 2017 13:03:42 GMT
Connection: keep-alive

{
  "Id": "0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2"
}

POST /node/ HTTP/1.0
Connection: keep-alive

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:43 GMT
Date: Sun, 05 Feb 2017 13:03:43 GMT
Connection: keep-alive

{
  "Id": "d2d95817f7b35ee4314c03f7f1faa5ca723043e76fcc08dbe6b2289a304bc8c4de72b6330bec0ed2c6757600e9d48b43e6ce58416877ad7382a3695549eedf27"
}

POST /node/ HTTP/1.0
Connection: keep-alive

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:42 GMT
Date: Sun, 05 Feb 2017 13:03:42 GMT
Connection: keep-alive

{
  "Id": "0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641"
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":0}

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 193
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:51 GMT
Date: Sun, 05 Feb 2017 13:03:51 GMT

{
  "Nodes": [
    {
      "id": "0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2",
      "Up": true
    }
  ]
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":2,"Other":0}

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 193
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:03:57 GMT
Date: Sun, 05 Feb 2017 13:03:57 GMT

{
  "Nodes": [
    {
      "id": "0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641",
      "Up": true
    }
  ]
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":2}

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 367
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:04:06 GMT
Date: Sun, 05 Feb 2017 13:04:06 GMT

{
  "Nodes": [
    {
      "id": "0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2",
      "Up": true
    },
    {
      "id": "0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641",
      "Up": true
    }
  ]
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 34

{"One":1,"Other":2,"AssetType":2}
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 2
Content-Type: text/json
Last-Modified: Sun, 05 Feb 2017 13:06:54 GMT
Date: Sun, 05 Feb 2017 13:06:54 GMT

{}

```

### backend log

```
$ ./overlay_meta 
I0205 14:00:54.750586 p2p/simulations/rest_api_server.go:32] Swarm Network Controller HTTP server started on localhost:8888
I0205 14:03:42.337560 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/', Host: '', Path: '/', Referer: '', Accept: ''
I0205 14:03:42.337753 p2p/simulations/journal.go:44] subscribe
I0205 14:03:42.337808 META/network/simulations/overlay_meta.go:94] new network controller on abcde
I0205 14:03:42.338041 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:03:42.338483 p2p/simulations/network.go:252] node 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2 created
I0205 14:03:42.338600 META/network/simulations/overlay_meta.go:144] added node 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2 to network abcde
I0205 14:03:42.338762 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:03:42.339333 p2p/simulations/network.go:252] node 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641 created
I0205 14:03:42.339358 META/network/simulations/overlay_meta.go:144] added node 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641 to network abcde
I0205 14:03:43.247640 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:03:43.247935 p2p/simulations/network.go:252] node d2d95817f7b35ee4314c03f7f1faa5ca723043e76fcc08dbe6b2289a304bc8c4de72b6330bec0ed2c6757600e9d48b43e6ce58416877ad7382a3695549eedf27 created
I0205 14:03:43.247960 META/network/simulations/overlay_meta.go:144] added node d2d95817f7b35ee4314c03f7f1faa5ca723043e76fcc08dbe6b2289a304bc8c4de72b6330bec0ed2c6757600e9d48b43e6ce58416877ad7382a3695549eedf27 to network abcde
I0205 14:03:51.049100 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:03:51.049199 p2p/simulations/network.go:300] starting node 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2: false adapter &{{{0 0} 0 0 0 0} 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2 0xc8201b40e0 0x4cc2c0 map[] [] 0x4c8ad0}
I0205 14:03:51.049309 p2p/simulations/network.go:309] started node 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2: true
I0205 14:03:57.599120 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:03:57.599274 p2p/simulations/network.go:300] starting node 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641: false adapter &{{{0 0} 0 0 0 0} 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641 0xc8201b40e0 0x4cc2c0 map[] [] 0x4c8ad0}
I0205 14:03:57.600676 p2p/simulations/network.go:309] started node 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641: true
I0205 14:04:06.486407 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:04:06.486646 p2p/adapters/inproc.go:168] protocol starting on peer 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641 (connection with 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2)
I0205 14:04:06.486729 p2p/adapters/inproc.go:168] protocol starting on peer 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2 (connection with 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641)
I0205 14:04:06.487006 p2p/protocols/protocol.go:220] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0205 14:04:06.488182 META/network/peercollection.go:33] protopeers now has added peers 0970496a2e447d44d2bbead9c13e1e40cff26daf83bbfe3e77e7da7b33d8ae23f8f6e108eb4316e445d8c3f8a043085c600baabab6de354f15a193c808bf87e2, total 1
I0205 14:04:06.487123 p2p/protocols/protocol.go:220] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0205 14:04:06.488218 META/network/peercollection.go:33] protopeers now has added peers 0a1a1728adc16733fab5201fd510570d8ecc755ca50e3fa63bc90f8fd08a450d2379f10f2824fae4ecbd1d2c4cac96233ee09ca1ab0d787256b82bef74ab9641, total 1
I0205 14:06:53.050568 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 14:06:54.318200 p2p/protocols/protocol.go:276] <= msg #0 (51 bytes)
I0205 14:06:54.318270 p2p/protocols/protocol.go:298] <= &{1 0000000000000000000000000000000000000000000000000000000000000000 [1 0 0 0 14 208 41 96 190 18 246 86 6 0 60]} *network.METAAssetNotification (0)
I0205 14:06:54.318299 p2p/protocols/protocol.go:311] handler 0 for *network.METAAssetNotification
I0205 14:06:54.318309 META/network/protocol.go:67] Peer received asset notification Digital Sales Report

```

