******************************
Storage incentives
******************************

..  index:: litigation

..  index::
   challenge

..  index::
   receipt
   litigation
   manifest
   insurance

..  index::
  receipt
  contract
  disconnection
  pair: receipt; storage

Positive collective incentivisation
==================================================

Storage incentives refer to the ability of a system to encourage/enforce preservation of content,
especially if the user explicitly requires that.
IPFS offers an interesting solution. Nodes participating in its network also mine
on their own altchain called filecoin (:cite:`filecoin2014`).
Filecoin's proof of work is defined to include proof that the miner possesses a set of randomly chosen units of storage depending on the parent block.
Using a strong proof of retrievability scheme, IPFS ensures that winning miner had relevant data. As miners compete, they find their chances of winning will be proportional to the percentage of the existing storage units they actually store. This is because the missing ones need to be retrieved from other nodes and thus delaying nodes chance to respond.

We see a whole range of issues with this particular approach:

* it is not clear that network latency cannot be masked by the parallel calculation of the ordinary proof of work component in the algorithm
* if the set of chunks are not selected differently for each node, mining will resemble a DDOS on nodes that actually store the data needed for the round.
* even if the selection of data to prove varies depending on the miner, normal operation incurs huge network traffic
* as the network grows the expected proportion of the data that needs to be retrieved increases. In fact given a practical maximum limit on a node's storage capacity, this proportion reaches a ceiling. If that happens miners will end up effectively competing on bandwidth
* in order to check proof of retrievability responses as part of block validation, existing data needs to be recorded on the blockchain. This leads to excessive use of the blockchain as the network grows and is unlikely to scale.
* competing miners working on the same task mean redundant use of resources
* If content is known to be popular, checking their integrity is spurious. But if choice of storage data to audit for the next block is truely random, there is no distinction between rarely accessed content and popular ones stored by many nodes resulting in wasted resouces.
* Similarly, users originating the content have also no way to indicate directly that some documents are important and not to be lost, while other temporary or derived data they can afford to lose.

Due to excessive use of blockchain and generated network traffic, these issues make the approach suspect, at best hugely wasteful, at worst infeasible on the large scale.

More importantly, however, IPFS provides only a scheme to collectively incentivise the network to store content. This brings in a 'tragedy of the commons' problem in that losing any particular data will have no negative consequence to any one storer node. This lack of individual accountability means the solution is rather limited as a security measure against lost content.

To summarise, we consider positive incentivisation in itself insufficient for ensured archival. In addition to that collective positive incentivisation implemented by competitive proof of retrievability mining is wasteful in terms of network traffic, computational resources as well as blockchain storage. In the subsequent sections we will introduce a very different approach.

Compensation for storage and guarantees for long-term data preservation
========================================================================

While Swarm's core storage component is analogous to traditional DHTs both in terms of network topology and routing used in retrieval, it uses the narrowest interpretation of immutable content addressed archive. Instead of just metadata about the whereabouts of the addressed content, the proximate nodes actually store the data itself.
When a new chunk enters the swarm storage network, it is propagated from node to node via a process called 'syncing'. The goal is for chunks to end up at nodes whose address is closest to the chunk hash. This way chunks can be located later for retrieval using kademlia key-based routing.

..  index::
   retrieve request
   latency

The primary incentive mechanism in swarm is compensation for retrieval where nodes are rewarded for successfully serving a chunk. This reward mechanism has the added benefit of ensuring that the popular content becomes widely distributed (by profit maximising storage nodes serving popular content they get queried for) and as a result retrieval latency is descreased.

The flipside of using this incentive only is that chunks that are rarely retrieved may end up lost. If a chunk is not being accessed for a long time, then as a result of limited storage capacity it will eventually end up garbage collected to make room for new arrivals. In order for the swarm to guarantee long-term availability, the incentive system needs to make sure that additional revenue is generated for chunks that would otherwise be deleted. In other words, unpopular chunks that do not generate sufficient profit from retrievals should compensate the nodes that store them for their opportunities forgone.

Basics of storage incentivisation
------------------------------------------------

A long-term storage incentivisation scheme faces unique challenges. For example, unlike in the case of bandwidth incentives where retrievals are immediately accounted and settled, long-term storage guarantees are promisory in nature and deciding if the promise was kept can only be decided at the end of its validity. Loss of reputation is not an available deterrent against foul play in these instances: since new nodes need to be allowed to provide services right away, cheaters could just resort to new identities to sell (empty) storage promises.

..  index::
  reputation
  punative measures
  deposit

Instead, we need punitive measures to ensure compliance with storage promises. These will work using a :dfn:`deposit system`. Nodes wanting to sell promisory storage guarantees should have a *stake verified and locked-in* at the time of making their promise. This implies  that nodes must be *registered* in advance with a contract and put up a security deposit.

Following :dfn:`registration`, a node may sell storage promises covering the time period for which their funds are locked. While their registration is active, if they are found to have lost a chunk that was covered by their promise, they stand to loose their deposit.

Requirements
-------------

In this context, :dfn:`owner` refers to the originator of a chunk (the one that uploads a document to the swarm), while :dfn:`storer` refers to a swarm node that actually stores the given chunk.

Let us start from some reasonable usage requirements:

* owners need to express their risk preference when submitting to storage
* storers need to express their risk preference when committing to storage
* there needs to be a reasonable market mechanism to match demand and supply
* there needs to be a litigation system where storers can be charged for not keeping their promise

An Owner's risk preference consists of the time period covered as well as on the :dfn:`degrees of redundancy` or certainty. These preferences should be specified on a per-chunk basis and they should be completely flexible on the protocol level.

Satisfying storers' risk preferences means that they have ways to express their certainty of preserving what they store and factor that in their pricing. Some nodes may not wish to provide storage guarantees that are too long term while others cannot afford to stake too big of a deposit. This differentiates nodes in their competition for service provision.

A *market mechanism* means there is flexible price negotation or discovery or automatic feedback loops that tend to respond to changes in supply and demand.

..  index:: litigation

A :dfn:`litigation` procedure necessitates that there are contractual agreements between parties ultimately linking an owner who pays for securing future avaiability of content and a storer who gets rewarded for preserving it and making it immediately accessible at any point in the future. The incentive structure needs to make sure that litigation is a last resort option.

It is also worth emphasizing that the producer and the consumer of the information may not be the same entity and it is therefore important that failure to make good on the promise to deliver the stored content is penalized even when the unserved consumer was not party to the agreement to store and provide the requested content. Litigation therefore is expected to be available to third parties wishing to retrieve content.

..  index::
   contract
   receipt

The simplest solution to manage storage deals is using direct contracts between owner and storer. This can be implemented with storers returning *signed receipts* of chunks they accept to store and owners paying for the receipts either directly or via escrow. These receipts are used to prove commitment in case of litigation. There are other more indirect variants of litigation which do not rely on owner and storer being in direct contractual agreement, which is the case if the eventual consumer is distinct from the storer and not known to them in advance.

In what follows we will elaborate variations on such storage incentive schemes. Since the basic ingredients are the same, we proceed to describe them in turn, starting with 1) user-side handling of redundancy, 2) registration and deposit, followed by 3) storage receipts and finally 4) the challenge based litigation system.

Owner-side handling of storage redundancy
==============================================================================

First we show how to delegate arbitrary security to the owner. This is important since this entails that the degree of redundancy does not need to be among the parameters handled by store requests, pricing or litigation. The idea is that redundancy is encoded in the document structure before its chunks are uploaded. For instance the simplest method of guarateeing redundancy of a file is to split the file into chunks that are one byte shorter than the normal chunksize and add a nonce byte to each chunk. This guarantees that each chunk is different and as a consequence all chunks of the modified file are different. When joining the last byte of each chunk is ignored so all variants map to the same original.
Assuming all chunks of the original file are different this yields a potential  :math:`256^x` equivalent replicas the owner can upload [#]_ .

..  rubric:: Footnotes
.. [#] We also explored the possibility that degree of redundancy is subsumed under local replication. Local replicas are instances of a chunk stored by nodes in a close neighbourhood. If that particular chunk is crucial in the reconstruction of the content, the swarm is much more vulnerable to chunk loss or latency due to attacks. This is because if the storers of the replicas are close, inflitrating in the storers' neighbourhood can be done with as many nodes as chunk type (as opposed to as many as chunk replicas). If there is cost to sybil attacks this brings down the cost by a factor of n where n is the number of replicas. We concluded that local replication is important for resilience in case of intermittend node dropouts, however, inferior to other solutions to express security level as expressed by the owner.

Luckily there are a lot more economical ways to encode a file redundantly. In particular the erasure coded, loss tolerant merkle tree, discussed in the previous section, allows the user to choose their own level of guaranteed data redundancy and security.

From here on we assume that the user applied CRS encoding when splitting their content and therefore expressed their desired degree of redundancy in the CRS parameters, the price of which they pay in terms of the increased number of chunks they need to pay storage for without adding complexity to the storage distribution and pricing.


Registration and deposit (SWEAR)
=================================

..   index:: :abbr:`SWEAR Secure Ways of Ensuring ARchival or SWarm Enforcement and Registration`


In order to sell promises of long-term storage, nodes must first register via a contract on the blockchain we call the *SWEAR* contract (Secure Ways of Ensuring ARchival or SWarm Enforcement and Registration).
The SWEAR contract allows nodes to register their public key to become accountable participants in the swarm by putting up a deposit. Registration is done by sending the deposit to the SWEAR contract, which serves as colleteral if terms that registered nodes 'swear' to keep are violated (i.e., nodes do not keep their promise to store).
Registration is valid only for a set period, at the end of which a swarm node is entitled to their deposit.
Users of Swarm should be able to count on the loss of deposit as a disincentive against foul play as long as enrolled status is granted. As a result the deposit must not be refunded before the registration expires.

..  index:: registration
   receipt

Registration in swarm is not compulsory, it is only necessary if the node wishes to sell promises of storage. Nodes that only charge for retrieval can operate unregistered. The incentive to register and sign promises is that they can be sold for profit. When a peer connection is established, the contract state is queried to check if the remote peer is a registered node. Only registered nodes are allowed to issue valid receipts and charge for storage.

When a registered node receives a request to store a chunk, it can acknowledge accepting it with a signed receipt. It is these signed receipts that are used to enforce penalties for loss of content through the SWEAR contract. Because of the locked collateral backing them, the receipts  can be viewed as secured promises for storing and serving a particular chunk up until a particular date. It is these receipts that are sold to nodes initiating requests.
In some schemes the issuer of a receipt can in turn buy further promises from other nodes pontentially leading to a chain of local contracts.

If on litigation it turns out that a chunk (covered by a promise) was lost, the deposit must be at least partly burned. Note that this is necessary because if penalites were paid out as compensation to holders of receipts of lost chunks, it would provide an avenue of early exit for a registered node by "losing" bogus chunks deposited by colluding users. Since users of Swarm are interested in their information being reliably stored, their primary incentive for keeping the receipts is to keep the Swarm motivated, not the potential for compensation.
If deposits are substantial, we can get away with paying out compensation for initiating litigation, however we must have the majority (say 95%) of the deposit burned in order to make sure the easy exit route remains closed.

The SWEAR contract handles all registration and deposit issues. It provides a method to pay the deposit and register the node's public key. An accessor is available for checking that a node is registered.

.. The corresponding solidity code: https://github.com/ethereum/tree/swarm/swarm/services/swear/swear.sol.

Forwarding chunks
======================

..  index:: retrieve request

In normal swarm operation, chunks are worth storing because of the possibility that they can be profitably "sold" by serving retrieve requests in the future. The probability of retrieve requests for a particular chunk depends on the chunk's popularity and also, crucially, on the proximity to the node's address.

Nodes are expected to forward all chunks to nodes whose address is closer to the chunk. This is the normal syncing protocol. It is compatible with the pay-for-retrieval incentivisation: once a retrieve request reaches a node, the node will either sell the chunk (if it has it) or it will pass on the retrieve request to a closer node. There is no financial loss from syncing chunks to closer nodes because once a retrieve request reaches a closer node, it will not be passed back out, it will only be passed closer. In other words, syncing only serves those retrieve requests that the node would never have profited from anyway and thus it causes no financial harm due to lost revenue.

..  index:: syncing

For insured chunks, a similar logic applies - more so even because there is a positive incentive to sync. If a chunk does not reach its closest nodes in the swarm before someone issues a retrieval request, then the chances of the lookup failing increase and with it the chances of the chunk being reported as lost. The resulting litigation as discussed below poses a burden on all swarm nodes that have ever issued a receipt for the chunk and therefore incentivises nodes to do timely forwarding. The audit process described in :cite:`tronetal2016smash` provides additional guarantees that chunks are forwarded all the way to the proximate nodes.

Swarm assumes that nodes are content agnostic, i.e., whether a node accepts a new chunk for storage should depend only on their storage capacity. Registered nodes have the option to indicate that they are at full capacity. This effectively means they are temporarily not able to issue receipts so in the eyes of connected peers they will act as unregistered. As a result, when syncing to registered nodes, we do not take no for an answer: all chunks sent to a registered node can be considered receipted. If the node already has the chunk (received it earlier from another peer), the receipt is not paid for.
As we show later the protocol for issuing of receipts can be made part of the syncing protocol.

Litigation on loss of content (SWINDLE)
========================================

..  index:: :abbr:`SWINDLE = SWarm INsurance Driven Litigation Engine`

If a node fails to observe the rules of the swarm they 'swear' to keep, the punative measures need to be enforced which is preceded by a litigation procedure. The implementation of this process is called SWINDLE = SWarm INsurance Driven Litigation Engine.

Submitting a challenge
------------------------------


..  index::
  challenge
  refutation

Nodes provide signed receipts for stored chunks which they are allowed to charge arbitrary amounts for. The pricing and deposit model is discussed in detail below. If a promise is not kept and a chunk is not found in the swarm anyone can report the loss by putting up a :dfn:`*challenge*`. The response to a challenge is a :dfn:`*refutation*`. Validity of the challenge as well as its refutation need to be easily verifyable by the contract.
For now, we can just assume that the litigation is started by the challenge after a user attempts to retrieve insured content and fails to find a chunk. Litigation will be discussed below in the wider context of regular integrity audits of content in the swarm.

The challenge takes the form of a transaction sent to the SWEAR contract in which the challenger presents the receipt(s) of the lost chunk. Any node is allowed to send a challenge for a chunk as long as they have a valid receipt for it (not necessarily issued to them).

This is analogous to a court case in which the issuers of the receipts are the defendants who are guilty until proven innocent. Similarly to a court procedure public litigation on the blockchain should be a last resort when the rules are abused despite the deterrents and positive incentives.

The same transaction also sends a deposit covering the upload of a chunk. The contract verifies if the receipt is valid, ie.,

* receipt was signed with the public key of a registered node
* the expiry date of the receipt has not passed
* sufficient funds are sent alongside to compensate the peer for uploading the chunk in case of a refuted challenge

The last point above is designed to disincentivise frivolous litigation, i.e., bombarding the blockchain with bogus challanges potentially causing a DoS attack.

..  index:: DoS

A challenge is open for a fixed amount of time, the end of which essentially is the deadline to refute the challenge. The challenge is refuted if the chunk is presented (additional ways are discussed below). Refutation of a challenge is easy to validate by the contract since it only involves checking that the hash of the presented chunk matches the receipt. This challenge scheme is the simplest way (i) for the defendants to refute the challenge as well as (ii) to make the actual data available for the nodes that need it.

In normal operation, litigation should be so rare that it may be necessary to introduce a practice of random :dfn:`*auditing*` to test nodes' compliance with distribution rules. In such cases the challenge can carry a flag which when set would indicate that providing the actual chunk, (ii) above, is unnecessary. In order to reduce network traffic, in such cases presenting the chunk can be replaced by providing a *proof of custody*. Registered nodes could be obligated to publish random challenges regularly. Note that in order not to burden the live chain, this could happen off-chain and they would only make it to the blockchain if foul play is proved.

Conversely, if such auditing is a regular automated process, then litigation will typically be initiated as part of escalating a failed audit.
:cite:`ethersphere2016smash` describes such just such an audit protocol using the SMASH proof-of-custody construct.

The outcome of a challenge
-------------------------------------

Successful refutation of the challenge is done by anyone sending the chunk as data within a transaction to the blockchain. Upon verifying the format of the refutation, the contract checks its validity by checking the hash of the chunk payload against the hash that is litigated. If the refutation is valid, the cost of uploading the chunk is compensated from the deposit of the challenge, with the remainder refunded.

..  index::
    DoS

In order to prevent DoS attacks, the deposit for compensating the swarm node for uploading the chunk into the blockchain should actually be substantially higher than (e.g., a small integer multiple of) the corresponding gas price used to upload the demanded chunk.

The contract also comes with an accessor for checking that a given node is challenged (potentially liable for penalty), so the accused nodes can get notified to present the chunk in a timely fashion.

If a challenge is refuted within the period the challenge is open, no deposit of any node is touched.
After successful refutation the challenge is cleared from the blockchain state.

..  index::
   deposit
   refutation
   challenge

If however the deadline passes without successful refutation of the challenge, then the charge is regarded as proven and the case enters into enforcement stage. Nodes that are proven guilty of losing a chunk lose their deposit (in part or full depending on the variant). Enforcement is guaranteed by the fact that deposits are locked up in the SWEAR contract.

..  index::
  suspension
  cheating

Punishment can entail :dfn:`*suspension*`, meaning a node found guilty is no longer considered a registered swarm node. Such a node is only able to resume selling storage receipts once they create a new identity and put up a deposit once again. This is extra pain inflicted on nodes for cheating.
Below we propose a system where nodes lose only part of their deposit for each chunk lost and only in case of deliberate cheating do they lose their entire deposit and get suspended.

If refutation of litigation is found to be common enough, sending transactions is not desirable since it is bloating the blockchain.
The audit challenges using the SMASH proof-of-custody described in :cite:`ethersphere2016smash` enable us to improve on this and make litigation a two step process. Upon finding a missing chunk, the litigation is started by the challenger sending an audit request [#]_ .

..  rubric:: Footnotes
.. [#] See :cite:`ethersphere2016smash` for the explanation of particular audit types. In fact any audit challenge should be escalated to the blockchain upon failure. The smash smart contract provides an interface to check validity of audit requests (as challenges) and verify the various response types (as refutations).

Multiple receipts - multiple defendants
----------------------------------------

Playing nice is further incentivized if a challenge is allowed to extend the risk of loss to all nodes that have given a promise to store the lost chunk. This means that when one storer is challenged, all nodes that have outstanding receipts covering the (allegedly) lost chunk stand to lose their deposit.

The SWEAR contract comes with an accessor for checking that a given chunk has been reported lost, so that holders of receipts by other swarm nodes can punish them as well for losing the chunk, which, in turn, incentivizes whoever may hold the chunk to present it (and thus refute the challenge) even if they are not the named defendant first accused.

Redundancy and multiple receipts
------------------------------------

Owners express their preference for storage period and for degree of redundancy/certainty.
As for storage period, the base unit used will be a :dfn:`*swarm epoch*`. The swarm epoch is the minimum interval a swarm node can register for.

To quantify redundancy level, a node specifies a total (minimum) amount of deposit that is to be at stake.  Customers (chunk owners or users) express this risk preference by collecting more than one receipt.

Nodes can choose to gamble of course by selling storage receipts without storing the chunk, in the hope of being able to retrieve the chunk from the swarm as needed. However, since storers have no real way to trust other nodes to fall back on, the nodes that issue receipts have a strong incentive to actually store the chunk themselves. Collecting receipts from several nodes therefore means that several replicas are likely to be kept in the swarm. Slogan: more receipts means more redundancy.

A priori this only works, however, in the simplest system in which the owner needs to receive and keep all the receipts signed by the storers. We shall return to this point later.

Receipt forwarding or chained challenges
===========================================

Collecting storer receipts and direct contracts
-------------------------------------------------

There are a few schemes we may employ. In the first, a storage request is forwarded from node to node until it reaches a registered node close to the chunk address. This storer node then issues a receipt which is passed back along the same route to the chunk owner.
The owner then can keep these receipts for later litigation.


Explicit direct contracts signed by storers held by owners has a lot of advantages. On top of its transparency and simplicity, this scheme enables owners to make sure that any degree of redundancy (certainty) promise is secured by deposits of distinct nodes via their signed promises. In particular it allows owners to insure their chunks against a total collateral  higher than any individual node's deposit. Also insuring a chunk against different deposits for varying periods is easy.

Unfortunately, this rather transparent system has caveats.

First of all, forwarding back receipts creates a lot of network traffic. The only purpose of receipts is to be able to use them in litigation, which is very rare, rendering virtually all this traffic spurious.

Secondly, since availability of a storer node cannot always be guaranteed, getting receipts back from storers may incur indefinite delays. The owner (who submits the request) needs a receipt that can be used for litigation later. If this receipt needs to come from the storer, then the process requires an entire roundtrip. If the owner requests additional security in the form of multiple receipts, receipts from all storers need to be passed back to the owner and stored. This means additional cost and overhead.

Furthermore, deciding on storers at the time the promise is made has a major drawback.
If the storage period is long enough the network may grow and new registered nodes come online in the proximity of the chunk. It can happen that routing at retrieval will bypass this storer. Though syncing makes sure that even in these cases the chunk is passed along and reaches closest nodes, their accountability regarding this old chunk cannot be guaranteed without further complications.

To summarize, explicit transparent contracts between owner and storer necessitate forwarding back receipts which has the following caveats:

* spurious network traffic
* delayed response
* potential non-accountability after network growth


.. What is a node's incentive to forward the request? Note that denying the chunk from peers that are not in their proximate bin have no benefit in retrieval (since requests served by the peer in question would never reach the node). If nonetheless they still do not forward, searches end up not finding the chunk, and they will be challenged. Having the chunk, they can always refute the challenge and the litigation costs may not be higher than what they gained from not purchasing receipts from a closer node. However, the litigation reveals that they cheated on syncing not offering the chunk in question. Learning this will prompt peers to stop doing business with the node. Alternatively, this could even be enforced on the protocol level requiring proof of forwarding on top of presenting the chunk, to avoid suspension.

Chaining challenges
--------------------

The other model is based on the observation that establishing the link between owner and storer can be delayed to take place at the time of litigation. Instead of waiting for receipts issued by storers, the owner direcly contracts their (registered) connected peer(s) and they immediately issue a receipt for storing a chunk.

When registered nodes connect, they are expected to have negotiated a price and from then on are obligated to give receipts for chunks that are sent their way according to the rules. This enables nodes to guarantee successful forwarding and therefore they can immediately issue receipts to the peer they receive the request from. Put in a different way, registered nodes enter into contract implicitly by connecting to the network and syncing.

..  index::
    sycing
    litigation
    forwarding
    receipt

When issuing a receipt in response to a store request you act as the entrypoint for a chunk. In this case the node is a *:dfn:`guardian`*, they act as the guardians of the chunk in question.


The receipt(s) that the owner gets from their connected peer can be used in a challenge.
When it comes to litigation, we play a blame game; challenged nodes defend themselves not necessarily by presenting the chunk (or proof of custody), but by presenting a receipt for said chunk issued by a registered node closer to the chunk address. Thus litigation will involve a chain of challenges with receipts pointing from owner via forwarding nodes all the way to the storer who must then present the chunk or be punished.

The litigation is thus a recursive process where one way for a node to refute a challenge is to shift responsibility and implicate another node to be the culprit.
The idea is that contracts are local between connected peers and blame is shifted along the same route that the chunk was forwarded on.

The challenge is constituted in submitting a receipt for the chunk signed by a node. (Once again everybody having a receipt is able to litigate).
Litigation starts with a node submitting a receipt for the chunk that is not found.
This will likely be the receipt(s) that the owner received directly from the guardian. The node implicated can refute the challenge by sending either the direct refutation (audit response or the chunk itself depending on the size and stage) to the blockchain as explained above or sending a receipt for the chunk signed by another node. This receipt needs to be issued by a node closer to the target. Additionally we stipulate that the redundancy requirement expressed by total deposit staked should also be preserved. In other words, if a node is accused with a receipt with deposit value of X, it needs to provide valid receipts from closer nodes with deposit totalling X or more. These validations are easy to carry out, so verification of chained challenges is perfectly feasible to add to the smart contract.

If a node is unable to produce either the refutation or the receipts, it is considered a proof that the node had the chunk, should have kept it but deleted it. This process will end up blaming a single node for the loss. If syncronisation was correctly followed and all the nodes forwarding kept their receipt, the node that is accused eventually is the node that was closest to the chunk to be stored when they received the request. We call this landing node the :dfn:`custodian` of the chunk.

Local replication
----------------------------------

As discussed above owners can manage the desired security level by using erasure coding with arbitrary degree of redundancy. Yet we do not want to replace local replication completely. Although the cloud industry is trying very hard to get away from the explicit x-fold redundancy model because it is very wasteful and incurs high costs – erasure coding can guarantee the same level of security using only a fraction of the storage space. However, in a data center redundancy is interpreted in the context of hard drives whose failure rates are low, independent and predictable and their connectivity is almost guaranteed at highest possible speed due to proximity. In a peer-to-peer network scenario, nodes could disappear much more frequently than hard drives fail. In the beginning,  we may expect larger than n replicas of chunks, but as the swarm grows and storage space is filling up, redundancy will drop automatically.
In order to guarantee robust operation, we need to require several replicas of each chunk. We will assume the magic number 3 (see :cite:`wilkinson2014metadisk`), i.e., make sure there are 3 distinct replicas of each chunk always preserved.

The simplest way we find is to delegate this task to the custodian. Upon receiving receipts, the custodian needs to collect receipts from the two closest registered peers. We simply introduce this stronger criteria on the audits: responses are expected to come from custodians.
When a node receives a store request, they are either  intermediate nodes (i.e., they have a registered peer that is closer to the chunk than they are) or custodians. If they are intermediate they are supposed to forward the request and have a receipt so they can point fingers to their neighbour when it comes to litigation. If they are custodians themselves, they need to get receipts from two extra registered nodes in their proximate bin.

As per the syncing protocol then the custodian indicates to a node that they are chosen as fellow custodians. If they respond with the receipt,  the custodian keeps it for the litigation. If they refuse to sign, they need to provide evidence that they know 1 or 2 registered nodes that are closer to the chunk than they are, not counting the custodian themselves. This should be a peer suggestion to the custodian and expected to happen only if the node is newly online. To prevent frivolous refusals, the co-custodian is expected to provide a signed and timestamped peer message they sent to that peer when it comes to litigation. If the co-custodian fails to obtain the receipt from their peer, it is considered a protocol breach and the co-custodian needs to disconnect which will make them real co-custodians so they need to respond with a receipt.

If the peer chosen as co-custodian does not give a receipt but fails to respond with a peer suggestion, it is considered a protocol breach and the peer is disconnected.

In order to be safe the custodian needs to have the 2 receipts, therefore it is important that each node maintains a proximate bin that contains at least 5 registered nodes. This number is also important in situations when the network grows.

Growing and shrinking network
-----------------------------------

For rubust resilient operation, it is crucial that retrieval is sound even when the network shrinks or grows.

When the network grows, it can happen that a custodian finds a node closer to its chunk. This means they need to forward the original store request, the moment they obtain a receipt they can use in finger pointing, they cease to be custodians and the ball is in the new custodians' court. When a node ceases to be custodian, they send their receipt to the co-custodians who are freed from their duty also.



Pricing, deposit, accounting
=============================

We posited in the introduction that registered nodes should be allowed to compete on quality of service and factor their certainty of storage in their prices. Market pricing of storage is all the more important once we realise that unlike gas, system-wide fixed storage price is neither easy nor necessary.

Gas is the accounting unit of computation on the ethereum blockchain, it is paid in as ether sent with the transaction and paid out in ether to the miner as part of the protocol.
The actual price of gas for a block is fixed system-wide yet it is dictated by market. It needs to be fixed since accounting for computation needs to be identical across all nodes of the network. It still can be dictated by the market since the miners the providers of the service gas is supposed to pay for, have a way to 'vote' on it. Miners of a block can change the gas price (based on how full the block is). To mitigate against extreme price volatilty, one can regulate the price by introducing restrictions on rate of change (absolute upper limit of percentage of change allowed from block to block).

Storage price is accounted for between p2p arrangements and therefore need not be fixed system-wide. Also such a mechanism of voting by service providers is not available. Note that in principle there is some information on the blockchain which could be used to inform prices: the number of (successful) litigations. If there is an increase in the percentage of litigations (number of proven charges normalised by the number of registered nodes), that is indication that system capacity is lower than the demand, therefore prices need to rise.
The other direction however when prices need to decrease has no such indicator: due to the floor effect of no litigation (quite expected normal operation), information on the blockchain is inconsequential as to whether the storage is overpriced.
Hence we conclude, fixed pricing of storage, is not viable without central authority or trusted third parties.

Another important decision is whether maximum deposits staked for a single chunk should vary independently of price. It is hard to conceptualise what this would mean in the first place. Assume that nodes' deposit varies and affects the probability that they are chosen as storers: a peer is chosen whose deposit is higher out of two advertising the same price. In this case, the nodes have an incentive to up the ante, and start a bidding war. In case of normal operation, this bidding would not be measuring confidence in quality of service but would simply reflect wealth.
We conclude that prices should be variable and entirely up to the node, but higher confidence or certainty should also be reflected directly in the amount of deposit they stake: deposit staked per chunk should be a constant multiple of the price.

Assuming  :math:`s` is a system-wide security constant dictating the ratio between price and deposit staked in case of loss, for an advertised price of :math:`p`, the minimum deposit is :math:`d=s\cdot p`. Although it never matters if the deposit is above the minimum, but it can happen peer want to lower their price without liquidating their funds in anticipation of an opportunity to raise prices in the future.

Price per chunk per epoch is truely freely variable and dictated by the free market.

Pricing storage in units of chunk retrieval
---------------------------------------------

With the scheme laid out in the previous section we established an implicit insurance system where

* all costs and obligation can be settled between connected peers
* signed promises (commitments that can be used to initiate litigation) are available as immediate responses to store requests (syncing)

As a consequence, all payment or accounting for storage promises can be done exactly the same way as with bandwidth. Handling settlement of storage expenses within SWAP is a major advantage and simplifies our system.

However, unlike in the case of retrieval, storage receipts represent an insurance of sorts and therefore their pricing is important. There is no sense in which chunk storage can be traded service for service one for one.

However, their price can always be expressed in terms of chunk retrievals, so SWAP can simply handle their accounting in a trivial way.


Optimising storage of receipts
=====================================

Implementation of chained receipts: Storage receipts and sync state
--------------------------------------------------------------------

[This needs more work]

The purpose of the receipt is to prove that a node closer to the target chunk than the node itself received the chunk and will either store it or forward it.
This is exactly what synchronisation does, therefore, proving (in)correct synchronisation is
a potential substitute for receipt based litigation.

If we stipulate that registered nodes need to sign sync state and able to prove a particular chunk was part of the synced batch, we can get away without storing individual receipts altogether and implement the persistence of receipts as part of the chunkstore mechanism on the one hand and the passing of receipts as part of the syncing mechanism on the other.

An advantage of using sync tokens as receipts is that when litigation takes place, one can point fingers to a node which already had the chunk at the time of syncing.
.. Another one is that receipts are not increasing network traffic.

Trading trust for storage
----------------------------

[This needs more work]

One bottleneck of the indirect litigation scheme is that nodes need to store receipts of old chunk they do not store just to point fingers to nodes they synced with in case of litigation. This is only an issue with nodes not in the proximate bin.

We can further explore the possibility that peers that a node has had a long syncing history with and had lots of chances to probe are trusted so instead of keeping receipts to implicate them, one can just present a sync token (not specific to a chunk, just a time period) that serves to (i) notify that peer to continue the litigation and (ii) indicate to the swarm that the two nodes take joint responsibility if the chunk is lost on them. For nodes that are supposed to store the chunk, this scheme would provide explicit framework to collude and cheat on the degree of redundancy, but for forwarding nodes this solves two issues. First it multiplies the overall stake on the line and second seriously reduces the storage requirements. Because of the join responsibility, a node no longer needs to keep receipts of old non-stored chunks if they can show they have a pact of joint responsibility with a node that is closer.

Publicly accessible receipts and consumer driven litigation
------------------------------------------------------------

End-users that store important information in the swarm have an obvious interest in keeping as many receipts of it as possible available for "litigation". The storage space required for storing a receipt is a sizable fraction of that used for storing the information itself, so end users can reduce their storage requirement further by storing the receipts in Swarm as well. Doing this recursively would result in end users only having to store a single receipt, the root receipt, yet being able to penalize quite a few Swarm nodes, in case only a small part of their stored information is lost.

A typical usecase is when content producers would like to make sure their content is available. This is supported by implementing the process of collecting receipts and putting them together in a format which allows for the easy pairing of chunks and receipts for an entire document. Storing this document-level receipt collection in the swarm has a non-trivial added benefit. If such a pairing is public and accessible, then consumers/downloaders (not only creators/uploaders) of content are able to litigate in case a chunk is missing. On the other hand, if the likely outcome of this process is punishment for the false promise (burning the deposit), motivation to litigate for any particular bit lost is slim.

This pattern can be further extended to apply to a document collection (Dapp/website level). Here all document-level root receipts (of the sort just discussed) can simply be included as metadata in the manifest entry for the document alongside its root hash. Therefore a manifest file itself can store its own warranty.


.. bibliography:: sw^3.bib
   :cited:
   :style: plain
