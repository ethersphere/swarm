*******************
Specifications
*******************

Bzz wire protocol
========================

Handshake
----------

 [0x01, Version: B_8, ID: B, Addr: [NodeID: B_64, IP: B_4 or B_6, Port: P], NetworkID; B_8, Caps: [[cap1: B_3, capVersion1: P], [cap2: B_3, capVersion2: P], ...]]

* Version: 8 byte integer version of the protocol
* ID: arbitrary byte sequence client identifier human readable
* Addr: the address advertised by the node, format identical to DEVp2p wire protocol
* NetworkID: 8 byte integer network identifier

  .. * Caps: swarm-specific capabilities, format identical to devp2p

Store request
----------------

  s[+0x02, key: B_32, metadata: [], data: B_4k]: the data chunk to be stored, preceded by its key.


Timeout in milliseconds. Note that zero timeout retrieval requests do not request forwarding, but prompt for a peers message response. therefore they serve also
as messages to retrieve peers.

MaxSize specifies the maximum size that the peer will accept. This is useful in
particular if we allow storage and delivery of multichunk payload representing
the entire or partial subtree unfolding from the requested root key.
So when only interested in limited part of a stream (infinite trees) or only
testing chunk availability etc etc, we can indicate it by limiting the size here.

Request ID can be newly generated or kept from the request originator.
If request ID Is missing or zero, the request is handled as a lookup only
prompting a peers response but not launching a search. Lookup requests are meant
to be used to bootstrap kademlia tables.

In the special case that the key is the zero value as well, the remote peer's
address is assumed (the message is to be handled as a self lookup request).
The response is a PeersMsg with the peers in the kademlia proximity bin
corresponding to the address.


  [0x03, key: B_32, Id: B_8, MaxSize: B_8, MaxPeers: B_8, Timeout: B_8, metadata: B]: key of the data chunk to be retrieved, timeout in milliseconds. Note that zero timeout retrievals serve also as messages to retrieve peers.

Peers Msg
------------------

peers Msg is one response to retrieval; it is always encouraged after a retrieval
request to respond with a list of peers in the same kademlia proximity bin.
The encoding of a peer is identical to that in the devp2p base protocol peers
messages: [IP, Port, NodeID]
note that a node's DPA address is not the NodeID but the hash of the NodeID.

Timeout serves to indicate whether the responder is forwarding the query within
the timeout or not.

The Key is the target (if response to a retrieval request) or missing (zero value)
peers address (hash of NodeID) if retrieval request was a self lookup.

Peers message is requested by retrieval requests with a missing or zero value request ID

[0x04, Key: B_32, peers: [[IP, Port, NodeID], [IP, Port, NodeID], .... ], Timeout: B_8, Id: B_8 ]


Bzz url scheme
========================

Manifest
===============

SWAP - swarm accounting protocol
=========================================









