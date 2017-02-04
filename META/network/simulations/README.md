# META-wire network testing

## Current test commands

(raw HTTP)

```POST / HTTP/1.0
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

GET /node/ HTTP/1.0

GET / HTTP/1.0
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

POST /node/ HTTP/1.0
Connection: keep-alive

POST /node/ HTTP/1.0
Connection: keep-alive

POST /node/ HTTP/1.0
Connection: keep-alive
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 2
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:26 GMT
Date: Sat, 04 Feb 2017 23:39:26 GMT
Connection: keep-alive

{}HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:26 GMT
Date: Sat, 04 Feb 2017 23:39:26 GMT
Connection: keep-alive

{
  "Id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c"
}HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:26 GMT
Date: Sat, 04 Feb 2017 23:39:26 GMT
Connection: keep-alive

{
  "Id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d"
}
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 142
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:28 GMT
Date: Sat, 04 Feb 2017 23:39:28 GMT
Connection: keep-alive

{
  "Id": "b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce"
}
GET /node/ HTTP/1.0

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 590
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:36 GMT
Date: Sat, 04 Feb 2017 23:39:36 GMT

{
  "One": 0,
  "Other": 0,
  "Nodes": [
    {
      "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
      "Up": false
    },
    {
      "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
      "Up": false
    },
    {
      "id": "b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce",
      "Up": false
    }
  ],
  "MessageType": 0
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":0}
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 239
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:44 GMT
Date: Sat, 04 Feb 2017 23:39:44 GMT

{
  "One": 0,
  "Other": 0,
  "Nodes": [
    {
      "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
      "Up": true
    }
  ],
  "MessageType": 0
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":2,"Other":0}
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 239
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:52 GMT
Date: Sat, 04 Feb 2017 23:39:52 GMT

{
  "One": 0,
  "Other": 0,
  "Nodes": [
    {
      "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
      "Up": true
    }
  ],
  "MessageType": 0
}

PUT /node/ HTTP/1.0
Content-Type: application/x-www-form-urlencoded
Content-Length: 20

{"One":1,"Other":2}
HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 413
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:39:56 GMT
Date: Sat, 04 Feb 2017 23:39:56 GMT

{
  "One": 0,
  "Other": 0,
  "Nodes": [
    {
      "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
      "Up": true
    },
    {
      "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
      "Up": true
    }
  ],
  "MessageType": 0
}

GET /node/ HTTP/1.0

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 588
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:40:04 GMT
Date: Sat, 04 Feb 2017 23:40:04 GMT

{
  "One": 0,
  "Other": 0,
  "Nodes": [
    {
      "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
      "Up": true
    },
    {
      "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
      "Up": true
    },
    {
      "id": "b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce",
      "Up": false
    }
  ],
  "MessageType": 0
}

GET / HTTP/1.0

HTTP/1.0 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Content-Length: 2032
Content-Type: text/json
Last-Modified: Sat, 04 Feb 2017 23:41:47 GMT
Date: Sat, 04 Feb 2017 23:41:47 GMT

{
  "Current": {
    "nodes": [
      {
        "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
        "Up": true
      },
      {
        "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
        "Up": true
      },
      {
        "id": "b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce",
        "Up": false
      }
    ],
    "conns": [
      {
        "one": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
        "other": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
        "up": true,
        "reverse": false
      }
    ],
    "Id": "abcde"
  },
  "Available": [
    {
      "nodes": [
        {
          "id": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
          "Up": true
        },
        {
          "id": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
          "Up": true
        },
        {
          "id": "b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce",
          "Up": false
        }
      ],
      "conns": [
        {
          "one": "107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c",
          "other": "1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d",
          "up": true,
          "reverse": false
        }
      ],
      "Id": "abcde"
    }
  ]
}
```

### backend log

```
lash@cantando ~/programming/projects/go/src/github.com/ethereum/go-ethereum/META/network/simulations $ ./overlay_meta 
I0205 00:39:14.754043 p2p/simulations/rest_api_server.go:32] Swarm Network Controller HTTP server started on localhost:8888
I0205 00:39:26.629383 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/', Host: '', Path: '/', Referer: '', Accept: ''
I0205 00:39:26.629667 p2p/simulations/journal.go:44] subscribe
I0205 00:39:26.629775 META/network/simulations/overlay_meta.go:98] new network controller on abcde
I0205 00:39:26.630202 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:26.631208 p2p/simulations/network.go:252] node 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c created
I0205 00:39:26.631496 META/network/simulations/overlay_meta.go:149] added node 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c to network abcde
I0205 00:39:26.631935 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:26.633206 p2p/simulations/network.go:252] node 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d created
I0205 00:39:26.633268 META/network/simulations/overlay_meta.go:149] added node 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d to network abcde
I0205 00:39:28.316419 p2p/simulations/rest_api_server.go:37] HTTP POST request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:28.317340 p2p/simulations/network.go:252] node b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce created
I0205 00:39:28.317389 META/network/simulations/overlay_meta.go:149] added node b7cbcbf614ddd8ecc08b33c92d02e1566eb822d6a4bdc4ac7ed0c7f17b5db30a7745dba509e157dc6a60e108f871c8064dbbde8249ad84d32642e2a554705dce to network abcde
I0205 00:39:36.028265 p2p/simulations/rest_api_server.go:37] HTTP GET request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:44.879905 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:44.880125 p2p/simulations/network.go:300] starting node 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c: false adapter &{{{0 0} 0 0 0 0} 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c 0xc8201257a0 0x4b1510 map[] [] 0x67c220}
I0205 00:39:44.880480 p2p/simulations/network.go:309] started node 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c: true
I0205 00:39:52.566399 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:52.566646 p2p/simulations/network.go:300] starting node 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d: false adapter &{{{0 0} 0 0 0 0} 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d 0xc8201257a0 0x4b1510 map[] [] 0x67c220}
I0205 00:39:52.568356 p2p/simulations/network.go:309] started node 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d: true
I0205 00:39:56.852092 p2p/simulations/rest_api_server.go:37] HTTP PUT request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:39:56.852395 p2p/adapters/inproc.go:168] protocol starting on peer 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d (connection with 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c)
I0205 00:39:56.852494 p2p/adapters/inproc.go:168] protocol starting on peer 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c (connection with 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d)
I0205 00:39:56.852934 p2p/protocols/protocol.go:216] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0205 00:39:56.853133 META/network/peercollection.go:33] protopeers now has added peers 107e06849ed26d0e8bbb0e8f74bdfbef03bb3cf6aadb945dbcd110f40f29adad9d5ab0434d413b84208b1ee89226446dffcf83db46dc69fe0dabf82f2f44750c, total 1
I0205 00:39:56.853205 p2p/protocols/protocol.go:216] registered handle for &{0 0000000000000000000000000000000000000000000000000000000000000000 []} *network.METAAssetNotification
I0205 00:39:56.853405 META/network/peercollection.go:33] protopeers now has added peers 1120437f6b1b5d1220d94a151cc74fc017ba281105b9616f3fce909c4f4551bba6139a53d6dcdbd73adbd7b3439ad722f33cfb6c6485f3cbe5a5e8be5c82e26d, total 2
I0205 00:40:04.595836 p2p/simulations/rest_api_server.go:37] HTTP GET request URL: '/node/', Host: '', Path: '/node/', Referer: '', Accept: ''
I0205 00:41:47.091392 p2p/simulations/rest_api_server.go:37] HTTP GET request URL: '/', Host: '', Path: '/', Referer: '', Accept: ''
```

