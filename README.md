# meta-wire

The MΞTA protocol sandbox

## OVERVIEW

The boilerplate code of this POC is based on code the swarm api and cmd implementations.

The object of this phase of the project is to get an Proof-of-concept p2p implementation of **meta-wire** using the same stack and interfaces as eth/swarm, in order to:

- enforce protocol type, structure and version between peers
- issue commands through peer using console (IPC, geth attach)
- interface with storage for retrieval of data
- rudimentary data search engine using swarm

## CONTENTS

Specific META files are:

- All files in /META
- All files in /cmd/META

## INSTALL

1. (might have to go get something, not sure, you'll find out)
2. `go install -v cmd/META`
3. `go install -v cmd/geth`

## RUNNING
 
In terminal 1: 

`META --metaaccount foo --maxpeers 5 --datadir /tmp/meta-0 --verbosity 5`

In terminal 2:

`META --metacccount bar --maxpeers 5 --datadir /tmp/meta-1 --verbosity 5 --port 31667`

(repeat the above for as many peers as you like, incrementing datadir name and port number accordingly)

In last terminal:

`geth attach /tmp/meta-#/META.ipc`, depending on which node # you want to talk to.

## FUNCTIONALITY

The client has all the base functionality of a vanilla geth client (for example port can be set with `--port`) - `geth attach <path-to-ipc>` and see modules init output for details.

Currently it forces you to specify the bogus param `--metaaccount`, all others metioned in `cmd/META/main.go` are optional (but unique params are necessary when using more than one node, see "RUNNING" above)

### TCP LAYER

Listens and dials. Peer connect must be made manually.

Upon connecting, the peer will be added to a pool of peers, contained in `PeerCollection` (see `META/network/peercollection.go`)

### PROTOCOL

Initializes and registers upon connection.

There is only one protocol message struct registered, which sends a notification of the following form:

```
type METAAssetNotification struct {
	Typ uint8 // enum of type of notification, see below
	Bzz storage.Key // swarm address; SHA-3 hash
	Exp []byte // expiry timestamp, binary marshalled time struct
}
```

where Typ is defined as:

```
var METAAssetType = map[uint8]string{
	ERN: "Eletronic Release Notification",
	DSR: "Digital Sales Report",
	MLC: "Music Licensing Company",
}
```
|

Upon manually adding a peer through geth console, the two different protocols will be mapped to two different instances of `p2p/protocols.Peer,` thus occupying two different slots in the `PeerCollection`

### RPC

RPC implements one API item `*ZeroKeyBroadcast.Sendzeronotification(<int>)` for sending a zero swarm hash with an expire time in as a specifiable assettype to all connected peers (see `META/api/api.go`):

### JS CLI

Added module **mw** which has one method:


- `mw.sendzero(<int>)` => RPC `*ZeroKeyBroadcast.Sendzeronotification(<int>)`

## ISSUES

...besides the fact that the META implementation is still at alpha stage at best;

- Current go-ethereum implementation forces modules specifications for the geth client to be hardcoded in `/internal/web3ext/web3ext.go`, forcing adjustments to the ethereum repo itself
- go-package `gopkg.in/urfave/cli.v1` conflicts with existing version in vendor folder in ethereum repo, making it impossible to have code importing this package outside of the repo dir structure.

## VERSION

Current version is **v0.1.0**

META is build on [https://github.com/ethersphere/go-ethereum](https://github.com/ethersphere/go-ethereum) repo, branch *network-testing-framework*

## ROADMAP

*proposed*

### 0.1 - Protocol primitives

1. ~~Implement protocol handshake, so that two separately running nodes can connect.~~
2. ~~Implement handshake and simple demo protocol content: A simple instruction can be sent via **console**, which is sent to a peer.~~
3. ~~Same as above, but receiving peer replies and whose output is echoed to **console**.~~
4. ~~Same as above, but with several listening peers responding~~
5. ~~Same as above, but some peers implement different protocols, or different versions of protocol, and hence should not respond.~~
6. Deploy on test net with simulations and visualizations

### 0.2 - Swarm integration, basic

1. Using LOCALSTORE, set up daemon watching root hash of swarm manifest to be monitored (eq. "guardian silo"), and who is triggered by changes
2. Establish queueing system of changes to be passed on from daemon triggered by changes to broadcasting metawire node
3. Same as 2. but using NETSTORE (and DBSTORE?)

### 0.3 - Access

1. Implement subscription methods in RPC and protocol
2. Set up peer whitelists/blacklists system for subscription
3. ...

### 0.4 - Advanced node communication

[...] Implement pss, protocol over bzz. (This point to be embellished and elaborated)


