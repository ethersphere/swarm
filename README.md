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
4. In **$GODIR** make sure that a symlink path `github.com/ethereum/go-ethereum`  points to the root of repo (because of import paths)

## RUNNING
 
In terminal 1: 

`META --metaaccount foo --maxpeers 5 --datadir /tmp/metafoo --verbosity 5`

In terminal 2:

`META --metacccount bar --maxpeers 5 --datadir /tmp/metabar --verbosity 5 --port 31667`

In terminal 3:

`geth attach /tmp/metafoo/META.ipc` or  `geth attach /tmp/metabar/META.ipc`, depending on which node you want to talk to.

## FUNCTIONALITY

The client has all the base functionality of a vanilla geth client (for example port can be set with `--port`) - `geth attach <path-to-ipc>` and see modules motd for details.

Currently it forces you to specify the bogus param `--metaaccount`, all others metioned in `cmd/META/main.go` are optional.

### Node TCP API

Listens and dials. Peer connect must be made manually.

### PROTOCOL

Initializes and registers upon connection.

One test protocol structure `Hellofirstnodemsg` is implemented:

```
type Hellofirstnodemsg struct {
	Pmsg string
	Sub protocols.Peer
}
```

### RPC

RPC implements two API items (see `META/api/api.go`:

- `*Info.Infoo()`: merely returns an object with `META/api.Config` settings
- `*ParrotNode.Hellofirstnode(<string>)`: Sends **<string>** packed into `Hellofirstnodemsg` protocol structure to first connected Peer, returns success/fail. 

### JS CLI

Added module **mw** which has two methods:

- `mw.infoo()` mapped to RPC `*Info.Infoo()` 
- `mv.hello(<string>)` mapped to RPC `*ParrotNode.Hellofirstnode()`

## ISSUES

...besides the fact that the META implementation is still at alpha stage at best;

- Current go-ethereum implementation forces modules specifications for the geth client to be hardcoded in `/internal/web3ext/web3ext.go`, forcing adjustments to the ethereum repo itself
- go-package `gopkg.in/urfave/cli.v1` conflicts with existing version in vendor folder in ethereum repo, making it impossible to have code importing this package outside of the repo dir structure.

## VERSION

Current version is **v0.1.0**

META is build on [https://github.com/ethersphere/go-ethereum](https://github.com/ethersphere/go-ethereum) repo, branch *network-testing-framework*

## ROADMAP

*proposed*

### 0.1

1. ~~Implement protocol handshake, so that two separately running nodes can connect.~~
2. ~~Implement handshake and simple demo protocol content: A simple instruction can be sent via **console**, which is sent to a peer.~~
3. ~~Same as above, but receiving peer replies and whose output is echoed to **console**.~~
4. Same as above, but with several listening peers responding
5. Same as above, but some peers implement different protocols, or different versions of protocol, and hence should not respond.
6. Deploy on test net with simulations and visualizations

### 0.2

1. Implement swarm protocol and/or peer alongside META, local storage
2. Implement pss, protocol over bzz. (This point to be embellished and elaborated)
3. Same as 1. amd 2. above but using testnet

