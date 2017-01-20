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

## FUNCTIONALITY

The client has all the base functionality of a vanilla geth client (for example port can be set with `--port`) - `geth attach <path-to-ipc>` and see modules motd for details.

Currently it forces you to specify the bogus param `--metaaccount`, all others metioned in `cmd/META/main.go` are optional.

### Node TCP API

Listens, dials and is protocol-ready.

Try `admin.addPeer("enode....")` with two nodes set up (the other different `--port and --datadir`  of course), the sender running `--verbosity 6`, see log output).

### PROTOCOL

Not yet implemented (not even handshake)

### RPC

RPC only implements one API item, defined in `META/api/api.go` - `*Info.Infoo()`, which returns an object with `META/api.Config` settings

### JS CLI

Added module **mw** which currently only has one method, `mw.infoo()`, which returns the object returned by the aforementioned RPC call.

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

1. Implement protocol handshake, so that two separately running nodes can connect.
2. Implement handshake and simple demo protocol content: A simple instruction can be sent via **console**, which is sent to a peer, which then replies and whose output is echoed to **console**.
3. Same as above, but with several listening peers responding
4. Same as above, but some peers implement different protocols, or different versions of protocol, and hence should not respond.
5. Deploy on test net with simulations and visualizations

### 0.2

1. Implement swarm protocol and/or peer alongside META, local storage
2. Implement pss, protocol over bzz. (This point to be embellished and elaborated)
3. Same as 1. amd 2. above but using testnet

