# Postal Service over Swarm

pss provides messaging functionality for swarm nodes without the need for a direct tcp connection between them.

It uses swarm kademlia routing to send and receive messages. Routing is deterministic and will seek the shortest route available on the network based on the information made available to it.

Messages are encapsulated in a devp2p message structure `PssMsg`. These capsules are forwarded from node to node using ordinary tcp devp2p until it reaches it's destination. The destination address is hinted in `PssMsg.To`

The content of a PssMsg can be anything at all, down to a simple, non-descript byte-slices. But convenience methods are made available to implement devp2p protocol functionality on top of it.

For the current state and roadmap of pss development please see https://github.com/ethersphere/swarm/wiki/swarm-dev-progress.

Please report issues on https://github.com/ethersphere/go-ethereum

Feel free to ask questions in https://gitter.im/ethersphere/pss

## STATUS OF THIS DOCUMENT

`pss` is under active development, and the first implementation is yet to be merged to the Ethereum main branch. Expect things to change.

## TOPICS

Pure pss is protocol agnostic. Instead it uses the notion of Topic. This is NOT the "subject" of a message, in terms of an email-messages, for example. Instead this type is used to internally register handlers for messages matching respective Topics.

Topic in this context virtually mean anything; protocols, chatrooms, or social media groups.

## MESSAGES

A pss message has the following layers:

- PssMsg
   Contains a recipient address hint and the Envelope.

- Envelope
   Same as whisperv5.Envelope

- Payload
   Byte-slice of arbitrary data

- ProtocolMsg
   An optional convenience structure for implementation of devp2p protocols. Contains Code, Size and Payload analogous to the p2p.Msg structure, where the payload is a rlp-encoded byteslice. For transport, this struct is serialized and used as the "payload" above.

### EXCHANGE

Message exchange in `pss` requires end-to-end encryption. It implements both asymmetric and symmetric encryption schemes. 

The end recipient of a message is defined as the node that can successfully decrypt that message using stored keys.

Messages are filtered using topics. Only incoming messages that match topic filters will be passed to underlying message handlers.

Address hinting can be used to make message routing more or less directional. See ROUTING below.

### ENCRYPTION

Asymmetric message exchange in theory only requires the peer's public key and a topic to be sent. The public key is derived from the private key passed to the `pss` constructor. There is no PKI functionality in pss. Thus public keys must be looked up externally and must be explicitly stored in the pss instance using the API method `SetPeerPublicKey()`. 

Symmetric message exchange in `pss` is available through a handshake-mechanism on the API level (see HANDSHAKE below). Symmetric keys may also be set explicitly, using the API call `SetPeerSymmetricKey()`. When set explicitly, symmetric key expiry may be set arbitrarily. Adding the symmetric key to the collection of symmetric keys used for decryption is optional.

Symmetric keys are stored using `whisperv5` as backend.

It is assumed that the less resource-demanding symmetric encryption will be preferred over asymmetric encryption for normal message traffic, at least if the volume of messages is significant. 

### DECRYPTION

When processing an incoming message, `pss` detects whether it is encrypted symmetrically or asymmetrically.

When decrypting symmetrically, `pss` iterates through all stored keys, and attempts to decrypt with each key in order. The cache will only store a certain amount of keys, and the iterator will return keys in the order of most recently used key first. (Garbage collection of keys abandoned by cache storage is not yet implemented).

## CONNECTIONS 

A "connection" in pss is a purely virtual construct. There are no mechanisms in place to ensure that the remote peer actually is present and listening. In fact, "adding" a peer involves merely a node's opinion that the peer is there. It may issue messages to that remote peer to a directly connected peer, which in turn passes it on. But if it is not present on the network - or if there is no route to it - the message will never reach its destination through mere forwarding. It may be argued that `pss` to a certain degree adopts the behavior of the `UDP` protocol.

Internally `pss` the most primitive notion of "connection" is adding a peer's public key together with a topic in a keypool. This method takes an address hint as parameter. The topic must be a valid whisper topic. The address hint is used for routing (see ROUTING below), and the value set will be used for all ensuing communcations using this public key / topic combination.

### HANDSHAKE

`pss` implements a handshake algorithm akin to `ssh`, where a special message structure is used to encapsulate the keys to be exchanged. This can be invoked using the API method `Handshake()`. The method takes topic and address hint as arguments.

The handshake creates two separate symmetric keys, one for outgoing and one for incoming traffic. Symmetric keys have expiry times, and an expired symmetric key invalidates the handshake. Handshakes may be initiated at any time for any peer, regardless of expiry.

Internally the handshake is performed as follows:

* **Public key** of peer is added with originator's API call `SetPeerPublicKey()`
* originator's API call `Handshake()` is executed with the peer's **public key**.
* `pss` generates and stores a random symmetric key using whisper backend
* the unencrypted symmetric key is wrapped in a `pssKeyMsg` struct
* `pssKeyMsg` struct is sent asymmetrically to peer.
* peer decodes the pss message and detects a `pssKeyMsg`
* peer adds the symmetric key to its keypool, with the address byteslice provided in the `pssKeyMsg`
* peer generates and stores a random symmetric key using whisper backend.
* peer pairs the generated and received symmetric keys in an internal keypair index. (from now on peer considers this handshake as completed)
* peer encrypts the generated symmetric key
* the encrypted symmetric key is wrapped in a `pssKeyMsg` struct
* `pssKeyMsg` struct is sent asymmetrically to originator.
* originator decodes the pss message and detects a `pssKeyMsg`
* originator attempts to decrypt the received symmetric key by iterating through its stored symmetric keys
* originator adds the decrypted symmetric key to its keypool, with the address byteslice provided in the `pssKeyMsg`
* originator pairs the received symmetric key with the symmetric key used to encrypt it in an internal keypair index
* handshake is complete

## ROUTING 

(please refer to swarm kademlia routing for an explanation of the routing algorithm used for pss)

`pss` uses *address hinting* for routing. The address hint is an arbitrary-length MSB byte slice of the peer's swarm overlay address. It can be the whole address, part of the address, or even an empty byte slice. The slice will be matched to the MSB slice of the same length of all devp2p peers in the routing stage.

If an empty byte slice is passed, all devp2p peers will match the address hint, and the message will be forwarded to everyone. This is equivalent to `whisper` routing, and makes it theoretically impossible to use traffic analysis based on who messages are forwarded to.

A node will also forward to everyone if the address hint provided is in its proximity bin, both to provide saturation to increase chances of delivery, and also for recipient obfuscation to thwart traffic analysis attacks. The recipient node(s) will always forward to all its peers.

## CACHING

pss implements a simple caching mechanism, using the swarm DPA for storage of the messages and generation of the digest keys used in the cache table. The caching is intended to alleviate the following:

- save messages so that they can be delivered later if the recipient was not online at the time of sending.

- drop an identical message to the same recipient if received within a given time interval

- prevent backwards routing of messages

the latter may occur if only one entry is in the receiving node's kademlia, or if the proximity of the current node recipient hinted by the address is so close that the message will be forwarded to everyone. In these cases the forwarder will be provided as the "nearest node" to the final recipient. The cache keeps the address of who the message was forwarded from, and if the cache lookup matches, the message will be dropped.

## USING PSS AS DEVP2P

When implementing the devp2p protocol stack, the "adding" of a remote peer is a prerequisite for the side actually initiating the protocol communication. Adding a peer in effect "runs" the protocol on that peer, and adds an internal mapping between a topic and that peer. It also enables sending and receiving messages using the main io-construct in devp2p - the p2p.MsgReadWriter.

Under the hood, pss implements its own MsgReadWriter, which bridges MsgReadWriter.WriteMsg with Pss.SendRaw, and deftly adds an InjectMsg method which pipes incoming messages to appear on the MsgReadWriter.ReadMsg channel.

An incoming connection is nothing more than an actual PssMsg appearing with a certain Topic. If a Handler har been registered to that Topic, the message will be passed to it. This constitutes a "new" connection if:

- The pss node never called AddPeer with this combination of remote peer address and topic, and

- The pss node never received a PssMsg from this remote peer with this specific Topic before.

If it is a "new" connection, the protocol will be "run" on the remote peer, as if the peer was added via the API. 

### TOPICS IN DEVP2P

When implementing devp2p protocols, topics are derived from protocols' name and version. The pss package provides the PssProtocol convenience structure, and a generic Handler that can be passed to Pss.Register. This makes it possible to use the same message handler code for pss that is used for directly connected peers in devp2p.

