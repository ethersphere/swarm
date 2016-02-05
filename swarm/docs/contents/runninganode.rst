********************
Running a node
********************

Installation
=======================
Swarm is part of the Ethereum stack and currently only supports running on the ethereum network as a subprotocol of the go client nodes.

Installation instructions for manual build are identical to those for the ethereum go implementation.
See https://github.com/ethereum/go-ethereum/wiki/Developers'-Guide

The source code is found on github: https://github.com/ethereum/go-ethereum/tree/swarm/

Supported Platforms
=========================

Geth runs on all major platforms (linux, MacOSX, Windows, also raspberry pie, android OS, iOS).

..  note::
  This package has not been tested on platforms other than linux and OSX.

Prerequisites
================

building :command:`geth` equires the following packages:

* [Go](https://golang.org)
* [Git](http://git.org)
* [Godep](http://github.com/gotools/godep)


Grab the relevant prerequisites and build from source.

    apt-get install libgmp32-dev golang git

On Mac OSX, using :command:`brew`

    brew install go git
    go get http://github.com/gotools/godep

Now set your environment variables:

  PATH=$GOPATH/bin:$PATH
  GOPATH=`godep path`:$GOPATH

Building from source
========================

Clone the repository (or better, your fork of it) to a directory of your choosing and switch to the working directory of the repo. Now checkout to the ``bzz`` branch, pull the latest version and build the binary:

  clone git@github.com:ethereum/go-ethereum.git
  cd go-ethereum
  git checkout swarm
  git pull
  godep go build -v ./cmd/geth

..  index::
  pair: make; swarm installation
  pair: Makefile; swarm installation

You can now run :command:`./geth` to start your node. See :ref:`Running a node` to learn how to operate a swarm node.

  $ make && sudo make install

in the toplevel directory of the unpacked distribution.

Configuration options
============================

This section lists all the options you can set in the config file:    :file:`<datadir>/bzz/<baseaccount>/config.json`

By default the swarm data directory is under the ethereum's data directory but different depending on the base address. This is important if you run multiple swarm nodes since storage, configuration, connected peers will all be distinct depending on the base address.

Main parameters
-----------------------

.. option:: Path  :file:`<datadir>/bzz/<baseaccount>}`
  swarm data directory

.. option:: Port
  8500
  port to run the http proxy server

.. @defopt PublicKey
..   Public key of your swarm base account
.. @end defopt

.. @defopt BzzKey
..   Swarm node base address (@math{hash(PublicKey)}). This is used to decide storage based on radius and routing by kademlia.
.. @end defopt

Storage parameters
-----------------------------

.. @defopt ChunkDbPath (@file{<datadir>/bzz/<baseaccount>/chunks})
..   leveldb directory for persistent storage of chunks
.. @end defopt

.. @defopt DbCapacity (5000000)
..   chunk storage capacity, number of chunks (5M is roughly 20-25GB)
.. @end defopt

.. @defopt CacheCapacity (5000)
..   Number of recent chunks cached in memory
.. @end defopt

.. @defopt Radius (0)
..   Storage Radius: minimum proximity order (number of identical prefix bits of address key) for chunks to warrant storage. Given a storage radius @math{r} and total number of chunks in the network @math{n}, the node stores @math{n*2^{-r}} chunks minimum. If you allow @math{b} bytes for guaranteed storage and the chunk storage size is @math{c}, your radius should be set to @math{int(log_2(nc/b))}
.. @end defopt

Chunker/bzzhash parameters
-------------------------------


..  index::
   chunker
   bzzhash

.. @defopt Branches (128)
..   Number of branches in bzzhash merkle tree. @math{Branches*ByteSize(Hash)} gives the datasize of chunks.
.. @end defopt

.. @defopt Hash (SHA256)
..   The hash function used by the chunker (base hash algo of bzzhash): SHA3 or SHA256
.. @end defopt

.. @defopt SplitTimeout (120s)
..   Maximum time before splitting a document times out
.. @end defopt

.. @defopt JoinTimeout (120s)
..   Maximum time before joining a document times out. Not used with Lazy Reader.
.. @end defopt

Syncronisation parameters
-------------------------------
..  index::
   syncronisation
   smart sync

.. @defopt KeyBufferSize (1024)
.. In-memory cache for unsynced keys
.. @end defopt

.. @defopt SyncBufferSize (128)
.. In-memory cache for unsynced keys
.. @end defopt

.. @defopt SyncCacheSize (1024)
.. In-memory cache for outgoing deliveries
.. @end defopt

.. @defopt SyncBatchSize (128)
.. Maximum number of unsynced keys sent in one batch
.. @end defopt


@defopt SyncPriorities ([3, 3, 2, 1, 1])
Array of 5 priorities corresponding to 5 delivery types:
delivery, propagation, deletion, history, backlog. Specifying a monotonically decreasing list of priorities is highly recommended.
@end defopt

..  index::
   delivery types

@defopt SyncModes ([true, true, true, true, false])
A boolean array specifying confirmation mode ON corresponding to 5 delivery types:
delivery, propagation, deletion, history, backlog. Specifying true for a type means all deliveries will be preceeded by a confirmation roundtrip: the hash key is sent first in an unsyncedKeysMsg and delivered only if confirmed in a deliveryRequestMsg.
@end defopt

..  index::
   delivery types
   delivery request message
   unsynced keys message


Hive/Kademlia parameters
---------------------------------
..  index::
   Kademlia

.. @defopt CallInterval (1s)
..   Time elapsed before attempting to connect to the most needed peer
.. @end defopt

.. @defopt BucketSize (3)
..   Maximum number of active peers in a kademlia proximity bin. If new peer is added, the worst peer in the bin is dropped.
.. @end defopt

.. @defopt MaxProx (10)
..   Highest Proximity order (i.e., Maximum number of identical prefix bits of address key) considered distinct. Given the total number of nodes in the network @math{N}, MaxProx should be larger than @math{log_2(N/ProxBinSize)}), safely @math{log_2(N)}.
.. @end defopt

.. @defopt ProxBinSize (8)
..   Number of most proximate nodes lumped together in the most proximate kademlia bin
.. @end defopt

.. @defopt KadDbPath (@file{<datadir>/bzz/<baseaccount>/bzz-peers.json})
..   json file path storing the known bzz peers used to bootstrap kademlia table.
.. @end defopt

.. @node SWAP parameters,  , Hive/Kademlia parameters, Configuration options
.. @subsection SWAP parameters
..    SWAP

.. @defopt BuyAt (@math{2*10^{10}} wei)
..   highest accepted price per chunk in wei
.. @end defopt

.. @defopt SellAt (@math{2*10^{10}} wei)
..   offered price per chunk in wei
.. @end defopt

.. @defopt PayAt (100 chunks)
..   Maximum number of chunks served without receiving a cheque. Debt tolerance.
.. @end defopt

.. @defopt DropAt (10000)
..   Maximum number of chunks served without receiving a cheque. Debt tolerance.
.. @end defopt
..    debt tolerance

.. @defopt AutoCashInterval (@math{3*10^{11}}, 5 minutes)
..   Maximum Time before any outstanding cheques are cashed
.. @end defopt

.. @defopt AutoCashThreshold (@math{5*10^{13}})
..   Maximum total amount of uncashed cheques in Wei
.. @end defopt

.. @defopt AutoDepositInterval (@math{3*10^{11}}, 5 minutes)
..   Maximum time before cheque book is replenished if necessary by sending funds from the baseaccount
.. @end defopt

.. @defopt AutoDepositThreshold (@math{5*10^{13}})
..   Minimum balance in Wei required before replenishing the cheque book
.. @end defopt

.. @defopt AutoDepositBuffer (@math{10^{14}})
..   Maximum amount of Wei expected as a safety credit buffer on the cheque book
.. @end defopt

.. @defopt PublicKey (PublicKey(bzzaccount))
..   Public key of your swarm base account use
.. @end defopt

.. @defopt Contract ()
..   Address of the cheque book contract deployed on the Ethereum blockchain. If blank, a new chequebook contract will be deployed.
.. @end defopt

.. @defopt Beneficiary (Address(PublicKey))
..   Ethereum account address serving as beneficiary of incoming cheques
.. @end defopt

@node Getting started,  , Configuration options, Running a node
@section Getting started

Use :command:{geth} with the @code{--bzzaccount} parameter to start the client with Swarm enabled. If you want automatic deposits to your chequebook, then this account should be unlocked @code{--unlock}.

By default, the config file is sought under @file{<datadir>/bzz/<bzzaccount>/config.json}. If this file does not exist at startup, the default config file is created which you can then edit (the directories on the path will be created if necessary). In this case or if @code{config.Contract} is blank (zero address), a new chequebook contract is deployed. Until the contract is confirmed on the blockchain, no outgoing retrieve requests will be allowed.

..  codeblock::
    geth --bzzaccount 0 --unlock

Setting up SWAP
-------------------------


..  index::
   chequebook
   autodeploy (chequebook contract)


SWAP (Swarm accounting protocol) is the  system that allows fair utilisation of bandwidth (see :ref:{Incentivisation}, esp. :ref:{SWAP -- Swarm Accounting Protocol}).
In order for SWAP to be used, a chequebook contract has to have been deployed. If the chequebook contract does not exist when the client is launched or if the contract specified in the config file is invalid, then the client attempts to autodeploy a chequebook:

    [BZZ] SWAP Deploying new chequebook (owner: 0xe10536..  .5e491)

If you already have a valid chequebook on the blockchain you can just enter it in the config file @code{Contract} field.

..  index::
   chequebook contract address (@code{Contract} configuration parameter)
   Contract, chequebook contract address

You can set a separate account as beneficiary to which the cashed cheque payment for your services are to be credited. Set it on the @code{Beneficiary} field in the config file.

..  index::
   maximum accepted chunk price (@code{BuyAt})
   offered chunk price (@code{BuyAt})
   SellAt, offered chunk price
   BuyAt, maximum accepted chunk price
   benefieciary (@code{Beneficiary} configuration parameter)
   Beneficiary, recipient address for service payments

Autodeployment of the chequebook can fail if the baseaccount has no funds and cannot pay for the transaction. Note that this can also happen if your blockchain is not synchronised. In this case you will see the log message:

..  codeblock::
   [BZZ] SWAP unable to deploy new chequebook: unable to send chequebook     creation transaction: Account
    does not exist or account     balance too low..  .retrying in 10s

   [BZZ] SWAP arrangement with <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301>: purchase from peer disabled; selling to peer disabled)

Since no business is possible here, the connection is idle until at least one party has a contract. In fact, this is only enabled for a test phase.
If we are not allowed to purchase chunks, then no outgoing requests are allowed. If we still try to download content that we dont have locally, the request will fail (unless we have credit with other peers).

..  codeblock::
    [BZZ] netStore.startSearch: unable to send retrieveRequest to peer [<addr>]: [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301> we cannot have debt (unable to buy)

Once one of the nodes has funds (say after mining a bit), and also someone on the network is mining, then the autodeployment will eventually succeed:

..  codeblock::
    [CHEQUEBOOK] chequebook deployed at 0x77de9813e52e3a..  .c8835ea7 (owner: 0xe10536ae628f7d6e319435ef9b429dcdc085e491)
    [CHEQUEBOOK] new chequebook initialised from 0x77de9813e52e3a..  .c8835ea7 (owner: 0xe10536ae628f7d6e319435ef9b429dcdc085e491)
    [BZZ] SWAP auto deposit ON for 0xe10536 -> 0x77de98: interval = 5m0s, threshold = 50000000000000, buffer = 100000000000000)
    [BZZ] Swarm: new chequebook set: saving config file, resetting all connections in the hive
    [KΛÐ]: remove node enode://23ae0e6..  .aa4fb@195.228.155.76:30301 from table

Once the node deployed a new chequebook its address is set in the config file and all connections are dropped to be reset with the new conditions. Once we reconnect, purchase in one direction should be enabled. The logs from the point of view of the peer with no valid chequebook:


..  codeblock::
    [CHEQUEBOOK] initialised inbox (0x9585..  .3bceee6c -> 0xa5df94be..  .bbef1e5) expected signer: 041e18592..  ..  ..  702cf5e73cf8d618
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301>    set autocash to every 5m0s, max uncashed limit: 50000000000000
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301>    autodeposit off (not buying)
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301>    remote profile set: pay at: 100, drop at: 10000,    buy at: 20000000000, sell at: 20000000000
    [BZZ] SWAP arrangement with <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301>: purchase from peer disabled;   selling to peer enabled at 20000000000 wei/chunk)


..  index:: autodeposit

Depending on autodeposit settings, the chequebook will be regularly replenished:

..  codeblock::
  [BZZ] SWAP auto deposit ON for 0x6d2c5b -> 0xefbb0c:
   interval = 5m0s, threshold = 50000000000000,
   buffer = 100000000000000)
   deposited 100000000000000 wei to chequebook (0xefbb0c0..  .16dea,  balance: 100000000000000, target: 100000000000000)


The peer with no chequebook (yet) should not be allowed to download and thus retrieve requests will not go out.
The other peer however is able to pay, therefore this other peer can retrieve chunks from the first peer and pay for them. This in turn puts the first peer in positive, which they can then use both to (auto)deploy their own chequebook and to pay for retrieving data as well. If they do not deploy a chequebook for whatever reason, they can use their balance to pay for retrieving data, but only down to 0 balance; after that no more requests are allowed to go out. Again you will see:


..  codeblock::
   [BZZ] netStore.startSearch: unable to send retrieveRequest to peer [aff89da0c6...623e5671c01]: [SWAP]  <enode://23ae0e62...8a4c6bc93b7d2aa4fb@195.228.155.76:30301> we cannot have debt (unable to buy)

If a peer without a chequebook tries to send requests without paying, then the remote peer (who can see that they have no chequebook contract) interprets this as adverserial behaviour resulting in the peer being dropped.

Following on in this example, we start mining and then restart the node. The second chequebook autodeploys, the peers sync their chains and reconnect and then if all goes smoothly the logs will show something like:

..  codeblock::
    initialised inbox (0x95850c6..  .bceee6c -> 0xa5df94b..  .bef1e5) expected signer: 041e185925bb..  ..  ..  702cf5e73cf8d618
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301> set autocash to every 5m0s, max uncashed limit: 50000000000000
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301> set autodeposit to every 5m0s, pay at: 50000000000000, buffer: 100000000000000
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301> remote profile set: pay at: 100, drop at: 10000, buy at: 20000000000, sell at: 20000000000
    [SWAP] <enode://23ae0e62..  ..  ..  8a4c6bc93b7d2aa4fb@195.228.155.76:30301> remote profile set: pay at: 100, drop at: 10000, buy at: 20000000000, sell at: 20000000000
    [BZZ] SWAP arrangement with <node://23ae0e62...8a4c6bc93b7d2aa4fb@195.228.155.76:30301>: purchase from peer enabled at 20000000000 wei/chunk; selling to peer enabled at 20000000000 wei/chunk)

As part of normal operation, after a peer reaches a balance of @code{PayAt} (number of chunks), a cheque payment is sent via the protocol. Logs on the receiving end:

..  codeblock::
    [CHEQUEBOOK] verify cheque: contract: 0x95850..  .eee6c, beneficiary: 0xe10536ae628..  .cdc085e491, amount: 868020000000000,signature: a7d52dc744b8..  ..  ..  f1fe2001 - sum: 866020000000000
    [CHEQUEBOOK] received cheque of 2000000000000 wei in inbox (0x95850..  .eee6c, uncashed: 42000000000000)


..  index:: autocash, cheque

The cheque is verified. If uncashed cheques have an outstanding balance of more than @code{AutoCashThreshold}, the last cheque (with a cumulative amount) is cashed. This is done by sending a transaction containing the cheque to the remote peer's cheuebook contract. Therefore in order to cash a payment, your sender account (baseaddress) needs to have funds and the network should be mining.

..  codeblock::
   [CHEQUEBOOK] cashing cheque (total: 104000000000000) on chequebook (0x95850c6..  .eee6c) sending to 0xa5df94be..  .e5aaz

For further fine tuning of SWAP, see :ref:{SWAP parameters}.

..  index::
   AutoDepositBuffer, credit buffer
   AutoCashThreshold, autocash threshold
   AutoDepositThreshold: autodeposit threshold
   AutoCashInterval, autocash interval
   AutoCashBuffer, autocash target credit buffer


