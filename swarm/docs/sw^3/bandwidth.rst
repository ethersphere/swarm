******************
Incentivisation
******************

The objective of an :index:`incentive system` is to encourage cooperative behavior and discourage freeriding: the uncompensated depletion of limited resources. In the context of swarm, storage and bandwidth are the two most important limited resources.
In what follows we present our current thinking for a comprehensive incentive system for swarm implemented through a smart contract. The incentive system leverages the ethereum infrastructure and the underlying value asset, :index:`Ether`.

The incentive strategy outlined here aspires to satisfy the following constraints:

* It is in the node's interest irrespective of whether other nodes follow it or not.
* It makes it expensive to hog other nodes' resources.
* It does not impose unreasonable overhead.
* It plays nice with "naive" nodes.
* It rewards those that play nice, including those following this strategy.

Preliminaries
=================

The ultimate goal of swarm is that end users are served content in a safe and speedy fashion. The underlying unit of accounting must be a chunk since this is the delivery unit that is sourced from a single independent entity. We start from the simplest assumption that delivery of a chunk is a valueable service which is directly chargeable when a chunk is delivered to a node that sent a retrieve request.

From a certain node's perspective, the probability of it being ever requested is proportional to the inverse of its distance from it (the distance, in turn, can be interpreted as the risk of it not being requested). In other words, following the underlying routing protocol by itself incentivises nodes to prefer chunks that are closer to their own address.

In the first iteration, we further assume that nodes have no preference as to which chunks to store other than their access count which is a reasonable predictor of their profitability. As a corollary, this entails that store requests are accepted by nodes irrespective of the chunk they try to store.

Accounting
===============

The idea is that nodes can trade services for services or services for tokens in a flexible way so that in normal operation a zero balance is maintained between any pair of nodes in the swarm.
This is done with the :dfn:`Swarm Accounting Protocol` (:abbr:`SWAP (Swarm Accounting Protocol`)), a scheme of pairwise accounting with negotiable prices.

..  index:: Swarm Accounting Protocol (SWAP)

Triggers for payment and disconnect
-------------------------------------

Each swarm node keeps a tally of offered and received services with each peer. In the simplest form, the service is the delivery of a chunk or more generally an attempt to serve a retrieve request, see later. We use the number of chunks requested and retrieved as a discreet integer unit of accounting. The tally is independently maintained on both ends of each direct connection in the peer-to-peer network for both self and the remote peer. Since disconnects can be arbitrary, it is not necessary to communicate and consent on the exact pairwise balances.

..  index::
    disconnection
    retrieve request

Each chunk delivery on the peer connection is accounted and exhcanged at a rate of one to one. On top of this, there is a possibility to compensate for services with ether (or other blockchain token) at a price agreed on in advance. Receiving payment should be accounted for equivalent service rendered, using the price offered.

In the ideal scenario of compliant use, the balane is kept around zero.
When the mutual balance on a given connection is tilted in favour of one peer, that peer should be compensated in order to bring the balance back to zero. If the balance tilts heavily in the other direction, the peer should be throttled and eventually choked and disconnected. In practice, it is sufficient to implement disconnects of heavily indebted nodes.

In stage one, therefore, we introduce two parameters that represent thresholds that trigger actions when the tally reached them.

..  option:: Payment threshold

  is the limit on self balance which when reached trigger a transfer of some funds to the remote peer's address in the amount of balance unit times unit price offered.

..  option:: Disconnect threshold

  is the limit which when reached triggers disconnect from the peer.

..  index::
   SellAt (offered price)
   offered price (`SellAt`)
   BuyAt highest accepted price)
   highest accepted price (`BuyAt`)
   PayAt, payment threshold)
   payment threshold (`PayAt`)
   DropAt, disconnect threshold)
   disconnect threshold (`DropAt`)


When node A connects with peer B the very first time during one session, the balance will be zero. Since payment is only watched (and safe) if connection is on, B needs to either (i) wait till A's balance reaches a positive target credit level or (ii) allow A to incur debt.
Since putting one node in positive credit is equivalent to the other incurring debt, we simply aim for (ii). In other words, upon connection we let peers get service right away and after the payment threshold is reached, we initiate compensation that brings balance up to zero.

In its simplest form, balances are not persisted between sessions (of the swarm node), but are preserved between subsequent connections to the same remote peer.
Therefore balances can be stored in memory only. Freeriding is already very difficult with this scheme since each peer that a malicious node is exploiting, will provide free service only up to the value of *Disconnect threshold* times unit price. While the node is running no reconnect is allowed unless compensation is paid to bring a balance above *Disconnect threshold*.

Negotiating chunk price
------------------------------
..  index::
  highest accepted chunk price
  offered chunk price
  disconnection

Prices are communicated in the protocol handshake as *highest accepted chunk price* and *offered chunk price*. The handshake involves checking if the highest accepted chunkprice of one peer is less than the offered chunkprice of the other. If this is the case no business is possible and the other peer can only be compensated on a service for service basis. If payment is not possible either way, the peers will try keep a balance until one peer's disconnect limit is reached.
There is also the possibility that when A and B connect, payment is only possible in one direction, from B to A, but A cannot pay B for services. In this case if A reaches past the payment limit, it does nothing. Since this is clearly a risk for B, it may make sense to keep the connection only if B stays predominantly in red (i.e., continually downloads more), otherwise disconnect.

All in all, it is not necessary for both ends to agree on the same price (or even agree on any price) in order to successfully cooperate. Potentially different pricing of retrievals is meant to reflect varying bandwidth costs. Setting highest accepted chunk price as 0 can also be used to communicate that one is unable or unwilling to pay with tokens.

Modes of payment
--------------------

Since transfer of ether is constrained by blocktime, actual transactions sent via the blockchain can effectively rate-limit a peer, moreover various delays in transaction inclusion might interfere with the timing requirements of accounting compensation.

Things can be improved if peers send some provable commitment to compensation directly in the :index:`bzz protocol`. On the one hand, as long as these commitments need blockchain transactions to verify, the risk for receiver is similar: by the time failing transactions are recognised by the creditor node, the indebted node is already more in debt. Whether the balance is restored after this can only be verified by checking the canonical chain after sending the transactions. On the other hand, provable commitments have two advantages: (i) they keep the accounting real time and (ii) allow for a differential treatment of inadvertant non-payment versus deliberate cheating.

..  index::
   cheating

One particular implementation could use ethereum transactions directly within the bzz protocol. Unfortunately, barring basic verification, no guarantees can be gained from the raw transaction. Also, sending them to the network is not a viable way to cash the payment they represent: If the same address is used to send transaction to multiple peers that act independently, there is no guarantee that the transactions end up in the same block or follow exactly the nonce order. Therefore, besides insufficient balance, they may fail on incorrect nonce.

Smart contracts, however, make it easy to implement more secure payment process.
Instead of a simple account, the sender address holds a *chequebook contract*. This chequebook contract is similar to a wallet holding an ether balance for the owner and allows signed cheques to be cashed by the recipient (or anyone), who simply send a transaction with the cheque as data to the contract's *cash* method.

..  index::
    pair: chequebook; contract

.. function:: cheque
      sign(<contract address, beneficiary, amount>)


* the contract keeps track of the cumulative total amount sent during the time of the connection.
* sender makes sure each new cheque sent increments the cumultive total amount sent.
* after connection is established, the cumulative amount for a remote peer is set based on the tally on the blockchain
* the cumulative amount for self (local peer) must be persisted since valid transactions may be in transit

the cheque is valid if:

* the contract address matches the address on the cheque,
* the cheque is signed by the owner (NodeId = public key sent in handshake)
* the signed data is a valid encoding of <contract address,beneficiary,amount>
* the cumulative total amount is greater than in the previous cheque sent.

Receiver may only keep the last cheque received from each peer and periodically cash it by sending it to the chequebook contract: a scheme that allows trusted peers to save on transaction costs.

Peers watch their receiving address and account all payments from the peer's chequebook and when they are considered confirmed, the tally is adjusted.
The long term use of a chequebook provides a credit history, use without failure (bounced cheques) constitues proof of compliance. Using the cumulative volume on the chequebook to quantify reliability renders chequebooks a proper *reputation system*.

..  index::
  reputation system


Charging for Retrieval
=========================


When a retrieve request is received the peer responds with delivery if the preimage chunk is found or a peers message if further search is initiated [#]_.

.. rubric:: Footnotes
.. [#] Each of these provides a valuable service to the initiator and therefore is charged on them. Due to their size in bytes, a peers message is roughly two orders of magnitude cheaper than delivery of the chunk payload. This should be reflected in their respective accounting weight but this would complicate things unduely. As long as each retrieval request triggers a chargeable response, accounting is sufficient to prevent denial of service attacks: when a node is spammed with retrieve requests (querying either existing or non-existing content) it is charged for each response so network integrity is protected by the fact that the attacker can only ever freeride for upto a value of *Disconnect limit*.

..  index:: :abbr:`DoS (denial of service attack)`

As a simplification, we assume that requesters credit their peers only upon first successful delivery, while nodes receiving the request charge for their forwarding effort right away. This keeps a perfect balance if each retrieve request results in successful retrieval or the ratio of failed requests is similar for the two peers (and have small variance accomodated by the disconnect threshold). In cases that this balance is genuinely skewed, one node must be requesting non-existing chunks or the other peer has inadequate connections or bandwidth resulting in its inability to deliver the requested existing chunks. Both situations warrant disconnection.

By default nodes will store all chunks forwarded as the response to a retrieve requests.
These lookup results are worth storing because repeated requests for the same chunk can be served from the node's local storage without the need to "purchase" the chunk again from others. This strategy implicitly takes care of auto-scaling the network. Chunks originating from retrieval traffic will fill up the local storage adjusting redundancy to use maximum dedicated disk/memory capacity of all nodes. A preference to store frequently retrieved chunks results in higher redundancy aligning with more current usage. All else being equal, the more redundant a chunk, the fewer forwarding hops are expected for their retrieval, thereby reducing expected latency as well as network traffic for popular content.


Whereas retrieval compensation may prove sufficient for keeping the network in a relatively healthy state in terms of latency, from a resilience point of view, extra incentives are  needed. We turn to this problem now.

