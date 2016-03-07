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

Compensation for storage and guarantees for long-term data preservation
========================================================================

When a new chunk enters the swarm storage network, it is propagated from node to node via a process called 'syncing'. The goal is for chunks to end up at nodes whose address is closest to the chunk hash. This way chunks can be located later for retrieval using kademlia key-based routing.

..  index::
   retrieve request
   latency

The primary incentive mechanism in swarm is compensation for retrieval where nodes are rewarded for successfully serving a chunk. This reward mechanism has the added benefit of ensuring that the popular content becomes widely distributed (by profit maximising storage nodes serving popular content they get queried for) and as a result retrieval latency is minimised.

The flipside of using this incentive only is that chunks that are rarely retrieved may end up lost. If a chunk is not being accessed for a long time, then as a result of limited storage capacity it will eventually end up garbage collected to make room for new arrivals. In order for the swarm to guarantee long-term availability, the incentive system needs to make sure that additional revenue is generated for chunks that would otherwise be deleted. In other words, unpopular chunks that do not generate sufficient profit from retrievals should compensate the nodes that store them for their opportunities forgone.

Basics of storage incentivisation
------------------------------------------------

A long-term storage incentivisation scheme faces unique challenges. For example, unlike in the case of bandwidth incentives where retrievals are immediately accounted and settled, long-term storage guarantees are promisory in nature and deciding if the promise was kept can only be decided at the end of its validity. Loss of reputation is not an available deterrent against foul play in these instances: since new nodes need to be allowed to provide services right away, cheaters could just resort to new identities to sell (empty) storage promises.

..  index::
  reputation
  punative measures
  deposit

Instead, we need punitive measures to ensure compliance with storage promises. These will work using a :dfn:`deposit system`. Nodes wanting to sell promisory storage guarantees should have a stake verified and locked-in at the time of making their promise. This implies  that nodes must be registered in advance with a contract and put up a security deposit.

Following registration, a node may sell storage promises covering the time period for which their funds are locked. While their registration is active, if they are found to have lost a chunk that was covered by their promise, they stand to loose (part of) their deposit.

Requirements
-------------

In this context, :dfn:`*owner*` refers to the originator of a chunk (the one that uploads a document to the swarm), while :dfn:`storer` refers to a swarm node that actually stores the given chunk.

Let us start from some reasonable usage requirements:

* owners need to express their risk preference when submitting to storage
* storers need to express their risk preference when committing to storage
* there needs to be a reasonable market mechanism to match demand and supply
* there needs to be a litigation system where storers can be charged for not keeping their promise

Owners' risk preference consist in the time period covered as well as a preference for the :dfn:`degrees of redundancy` or certainty. These preferences should be specified on a per-chunk basis and they should be completely flexible on the protocol level.

The total amount of deposit that nodes risk losing in case the chunk is lost could also be variable. Degrees of redundancy could be approximated by the total amount of deposit storers stake: in this approximation two nodes standing to lose 50 each if a chunk is lost provide as much security as five nodes each standing to lose 20. In this kind of network, the security deposit is therefore a variable amount that each node advertises. Variants of this deposit scheme are discussed below.

Satisfying storers' risk preferences means that they have ways to express their certainty of preserving what they store and factor that in their pricing. Some nodes may not wish to provide storage guarantees that are too long term while others cannot afford to stake too big of a deposit. This differentiates nodes in their competition for service provision.

A *market mechanism* means there is flexible price negotation or discovery or automatic feedback loops that tend to respond to changes in supply and demand.

..  index:: litigation

A *litigation* procedure necessitates that there are contractual agreements between parties ultimately linking an owner who pays for securing future avaiability of content and a storer who gets rewarded for preserving it and making it immediately accessible at any point in the future. Litigation is expected to be available to third parties wishing to retrieve content. The incentive structure needs to make sure that litigation is a last resort option.

..  index::
   contract
   receipt

The simplest solution to manage storage deals is using direct contracts between owner and storer. This can be implemented with storers returning **signed receipts of chunks they accept to store and owners paying for the receipts either directly or via escrow. These receipts are used to prove commitment in case of litigation. There are other more indirect variants of litigation which do not rely on owner and storer being in direct contractual agreement.

In what follows we will elaborate variations on such storage incentive schemes. Since the basic ingredients are the same, we proceed to describe them in turn, starting with 1) registration and deposit, followed by 2) storage receipts and finally 3) the challenge based litigation system.

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

When a registered node receives a request to store a chunk, it can acknowledge accepting it with a signed receipt. It is these signed receipts that are used to enforce penalties for loss of content through the SWEAR contract. Because of the locked collateral backing the receipts, they can be viewed as secured promises for storing and serving a particular chunk up until a particular date. It is these receipts that are sold to nodes initiating requests.
In some schemes the issuer of a receipt can in turn buy further promises from other nodes pontentially leading to a chain of local contracts.

If on litigation it turns out that a chunk (covered by a promise) was lost, the deposit must be at least partly burned. Note that this is necessary because if penalites were paid out as compensation to holders of receipts of lost chunks, it would provide an avenue of early exit for a registered node by "losing" bogus chunks deposited by colluding users. Since users of Swarm are interested in their information being reliably stored, their primary incentive for keeping the receipts is to keep the Swarm motivated, not the potential for compensation.
If deposits are substantial, we can get away with paying out compensation for initiating litigation, however we must have the majority (say 95%) of the deposit burned in order to make sure the easy exit route remains closed.

The SWEAR contract handles all registration and deposit issues. It provides a method to pay the deposit and register the node's public key. An accessor is available for checking that a node is registered.
The corresponding solidity code: https://github.com/ethereum/tree/swarm/swarm/services/swear/swear.sol.

Forwarding chunks
======================

..  index:: retrieve request

In normal swarm operation, chunks are worth storing because of the possibility that they can be profitably "sold" by serving retrieve requests in the future. The probability of retrieve requests for a particular chunk depends on the chunk's popularity and also, crucially, on the proximity to the node's address.

Nodes are expected to forward all chunks to nodes whose address is closer to the chunk. This is the normal syncing protocol. It is compatible with the pay-for-retrieval incentivisation: once a retrieve request reaches a node, the node will either sell the chunk (if it has it) or it will pass on the retrieve request to a closer node. There is no financial loss from syncing chunks to closer nodes because once a retrieve request reaches a closer node, it will not be passed back out, it will only be passed closer. In other words, syncing only serves those retrieve requests that the node would never have profited from anyway and thus in causes no financial harm due to lost revenue.

..  index:: syncing

For insured chunks, a similar logic applies - more so even because there is a positive incentive to sync. If a chunk does not reach its closest nodes in the swarm before someone issues a retrieval request, then the chances of the lookup failing increase and with it the chances of the chunk being reported as lost. The resulting litigation as discussed below poses a burden on all swarm nodes that have ever issued a receipt for the chunk and therefore incentivises nodes to do timely forwarding.

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

The challenge takes the form of a transaction sent to the SWEAR contract in which the challenger presents the receipt(s) of the lost chunk. Any node is allowed to send a challenge for a chunk as long as they have a valid receipt for it (not necessarily issued to them).

This is analogous to a court case in which the issuers of the receipts are the defendants who are guilty until proven innocent. Similarly to a court procedure public litigation on the blockchain should be a last resort when the rules are abused despite the deterrents and positive incentives.

The same transaction also sends a deposit covering the upload of a chunk. The contract verifies if the receipt is valid, ie.,

* receipt was signed with the public key of a registered node
* the expiry date of the receipt has not passed
* sufficient funds are sent alongside to compensate the peer for uploading the chunk in case of a refuted challenge

The last point above is designed to disincentivise frivolous litigation, i.e., bombarding the blockchain with bogus challanges potentially causing a DoS attack.

..  index:: DoS

A challenge is open for a fixed amount of time, the end of which essentially is the deadline to refute the challenge. The challenge is refuted if the chunk is presented (additional ways are discussed below). Refutation of a challenge is easy to validate by the contract since it only involves checking that the hash of the presented chunk matches the receipt. This challenge scheme is the simplest way (i) for the defendants to refute the challenge as well as (ii) to make the actual data available for the nodes that needs it.

In normal operation, litigation should be so rare that it may be necessary to introduce a practice of random :dfn:`*probing*` to test nodes' compliance with distribution rules. In such cases the challenge can carry a flag which when set would indicate that providing the actual chunk, (ii) above, is unnecessary. In order to reduce network traffic, in such cases presenting the chunk can be replaced by providing a proof of custody. Registered nodes could be obligated to publish random challenges regularly. Note that in order not to burden the live chain, this could happen offline and they would only make it to the blockchain if foul play is proved.

The outcome of a challenge
-------------------------------------

Successful refutation of the challange is done by anyone sending the chunk as data within a transaction to the blockchain. Upon verifying the format of the refutation, the contract checks its validity by checking the hash of the chunk payload against the hash that is litigated. If the refutation is valid, the cost of uploading the chunk is compensated from the deposit of the challenge, with the remainder refunded.

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

What is a node's incentive to forward the request? Note that denying the chunk from peers that are not in their proximate bin have no benefit in retrieval (since requests served by the peer in question would never reach the node). If nonetheless they still do not forward, searches end up not finding the chunk, and they will be challenged. Having the chunk, they can always refute the challenge and the litigation costs may not be higher than what they gained from not purchasing receipts from a closer node. However, the litigation reveals that they cheated on syncing not offering the chunk in question. Learning this will prompt peers to stop doing business with the node. Alternatively, this could even be enforced on the protocol level requiring proof of forwarding on top of presenting the chunk, to avoid suspension.



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


Chaining challenges
--------------------

The other model is based on the observation that establishing the link between owner and storer can be delayed to take place at the time of litigation. Instead of waiting for receipts issued by storers, the owner direcly contracts their (registered) connected peer(s) that immediately issues a receipt for storing a chunk.

When registered nodes connect, they are expected to have negotiated a price and from then on are obligated to give receipts for chunks that are sent their way according to the rules. This enables nodes to guarantee successful forwarding and therefore they can immediately issue receipts to the peer they receive the request from. Put in a different way, registered nodes enter into contract implicitly by connecting to the network and syncing.

..  index::
    sycing
    litigation
    forwarding
    receipt

The receipt(s) that the owner gets from their connected peer can be used in a challenge.
When it comes to litigation, we play a blame game; challenged nodes defend the,selves not necessarily by presenting the chunk, but by presenting a receipt for said chunk issued by a registered node closer to the chunk address. Thus litigation will involve a chain of challenges with receipts pointing from owner via forwarding nodes all the way to the storer who must then present the chunk or be punished.

The litigation is thus a recursive process where one way for a node to refute a challenge is to shift responsibility and implicate another node to be the culprit.
The idea is that contracts are local between connected peers and blame is shifted along the same route as what the chunk travels.

The challenge is constituted in submitting a receipt for the chunk signed by a node. (Once again everybody having a receipt is able to litigate).
Litigation starts with a node submitting a receipt for the chunk that is not found.
This will likely be the receipt(s) that the owner received directly from the peer(s) it first sent the request to (a node that was directly connected to it at the time the request was initiated). The node implicated can refute the challenge by sending either the chunk itself to the blockchain as explained above or sending a receipt for the chunk signed by another node. This receipt needs to be issued by a node closer to the target. Additionally we stipulate that the redundancy requirement expressed by total deposit staked should also be preserved. In other words, if a node is accused with a receipt with deposit value of X, it needs to provide valid receipts from closer nodes with deposit totalling X or more. These validations are easy to carry out, so verification.

If a node is unable to produce either the chunk or the receipts, it is considered a proof that the node had the chunk, should have kept it but deleted it. If all nodes delete the chunk but preserve their receipt, this process will end up blaming the single closest node for the loss.

Compared to the scheme where owners collected direct receipts from storers with the help of forwarding, this is a regression in the sense that it is unable to factor in required redundancy (or storers certainty). No matter how many inital receipts one buys, challenges may all end up at the same single node.
This system also cannot deal with varying deposits and prices and seems to be feasable only in the context of fixed equal deposits for all registered nodes as well as a system-wide fixed price for a chunk per epoch. Not ideal.

..  index:: double signing

Ultimately the problem is that multiple separate receipts can be forwarded to the same node. This node sells multiple receipts (to different parties) all covering the same chunk, thereby reducing the total deposit securing the chunk from multiple nodes' worth, to just one. We propose to fix this flaw by explicitly forbidding nodes to issue multiple receipts for the same chunk (:dfn:`double signing`). To enforce this, we need to use a more complex system of deposits.

Reserved deposit and punishment for double signing
---------------------------------------------------

..  index::
    deposit
    chating
    double signing

Under the modified rules we allow for receipts to be backed by only a specified fraction of the total deposit. Then, if a node is found to have lost a chunk, only that part of the deposit is forfeited - with the caveat that from that moment on, the node may no longer be seen as trustworthy and it may not be able to sell receipts until it restores its deposit to the original total.
If the deposit is not restored, the node can still be litigated against based on outstanding receipits and will continue to lose stake if found guilty. Additional limits of tolerance can be introduced: for instance if the cumulative deposit lost on chunks reaches X percent of the total deposit, the node loses their entire deposit. Another way to put it may be that you got a set number of lives before the game is over.

On top of the amount of collateral dedicated to penalties for lost chunks, the deposit has an extra reserve amount.
This is designed to prevent double signing (or any other form of premeditated cheating). If a node signs a receipt for a chunk and is required to store it (i.e. it does not have a receipt from a node closer to the chunk hash whom it can point to in case of litigation); then the node must not sell another receipt covering the same chunk. Doing so is considered double signing, and if a node is found to be double signing (preventable misbehaviour), it loses its entire deposit and has to reregister before further operation.

With this scheme in place, we can once again ensure a minimum redundancy by purchasing multiple receipts. This works as follows: The owner purchases storage receipts form multiple connected peers. Each of these peers attempts to pass the chunk forward (obtain secondary receipts from nodes with addresses closer to the chunk hash). Along the way, no node can accept the chunk from two different nodes without first securing a forward receipt otherwise they would violate the rule against double signing. When the chunk reaches its home among the peers closest to it, there will be no closer nodes to pass the chunk on to and multiple nodes are left with the responsibility of storing the chunk. Thus we have reestablished that multiple receipts entail more redundancy.

Furthermore we were to explicitly allow receipts to stake an arbitrary amount of security deposit on the line. A receipt with a higher deposit value will also be more expensive. The effect of this scheme is that the degree of redundancy bought by the originator can never decrease. To see why, consider a node that has signed a receipt for a chunk with a deposit at stake of 100. This node may in turn purchase a receipt worth 100 from a closer node, or, it may purchase two receipts worth 50 each from two different closer nodes. These two receipts of 50 can never be recombined into one worth 100 because this would require double signing. In principle redundancy level can increase but in practice this is unlikely to ever happen. Litigation rules require that when a node is challenged, the challenge can be refuted by showing receipts secured with total deposits adding up to no less than the receipt the node is challenged with. In other words, the redundancy level has to be matched. Nodes have no motivation to purchase receipts secured by a higher value.
The fact that receipts can be issued of any amount below the total deposit of the node also makes it possible to match the exact degree of redundancy.

The fact that chunks stored by one and the same node are insured with variable stake
and that losing a chunk does not lead to suspension, changes the strategy nodes use when  deciding which chunks to delete (i.e., it may be worth deleting chunks that are insured against a tiny fraction of our deposit). This may be somewhat problematic.

Under this scheme however, there is another bottleneck too. As an owner, a node wants to collect more receipts than the number of their direct connections. This means that they need to wait until their connection forwards the chunk, i.e., found a new peer that gave a receipt.
This reintroduces non-immediacy in chunk submissions, moreover it necessitates a new process that records how much of the redundancy is already covered and continuously attempts to cover the rest.

To summarize so far, the latest scheme has the caveats:

* delay in securing receipts due to double signing restriction
* problematic semantics of request with variable partial stake

Maximising degree of redundancy
----------------------------------

In the special case when the redundancy requirement is within the total collateral of the  proximate bin of nodes, the last scheme can be improved.

Under this scheme forwarding the original request is explicitly delineated from distributing chunks according to the redundancy criteria. Similarly to the double signing pattern, owners need to accept that their chunk is insured only by one node's deposit. In the first phase of forwarding nodes pass on the original request (with its redundancy requirements) to nodes closer to the chunk address all the way to the current closest online node. This node acts as a coordinator to distribute chunks among nodes in its proximate bin. Forwarding to the closest node is enforced by the risk of losing the entire deposit: when challenged, a node can present the chunk or present a receipt from a node closer to the chunk, otherwise they stand to lose their deposit and registered status.

Assume the litigation reached the closest node. The closest node has another rule to follow, they can only refute the challenge by presenting several receipts from registered nodes in their proximate bin. Total deposit of the nodes must be higher than the total cost per epoch offered by the owner.
If the chain of receipts reaches the closest node and it sends in the batch of receipts to the SWEAR contract during litigation, the set of peers are considered jointly responsible. If the chunk is lost, each one of them lose their deposit.

If there is system-wide maximum limit on the degree of redundancy owners can require, then registered nodes can make sure their known neighbourhood can always satisfy that, simply by keeping their proximate bin large enough. Conversely, if there is maximum deposit that a node is allowed to stake for one chunk, owners offering multiples of this can safely assume that their chunk will be stored in several replicas.

This scheme can remedy both the variable deposit issue as well as the immediacy problem.

After the initial distribution, it is up to individual nodes to trickle down their receipts, always fully filling in their capacity according to the previous scheme, constrained by double signing and maximum deposit. Due to syncing, this pattern makes sure that storers occupy the neighbourhood of the chunk address irrespective of how the network grows.

The insight here is that once the owner has the receipt to initiate litigation and the chunk reached its proximate nodes delays (due to the double signing constraint) when adapting to network growth are perfectly acceptable.

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

We propose the following deposit structure.

..  math::
    d = spl + r

where

* :math:`d` is the nodes deposit
* :math:`p` is the price per chunk per epoch that the node is asking for the maximum stake
* :math:`s` is a system-wide security constant dictating the ratio between price and deposit staked in case of loss
* :math:`l` is the number of lives, i.e., the number of chunks a node can provably lose and not replenish their deposit before they are suspended
* :math:`r` reserve deposit is a system wide fix amount that represents the minimum stake swarm has as collateral for cheating.

The number of lives can be maximised in :math:`2`.
We recommend an extra constant for the overall maximum lives, i.e., the total number of losses a node is allowed before it is suspended. This is not constrained by the deposit since we allow a node to replenish their deposit after part of it is burned as a result of punishment for chunk loss.
Price per chunk per epoch is truely freely variable and dictated by the free market.
The reserve deposit is a much higher amount. We recommend that it is at least :math:`2m` where :math:`m` is the maximum total stake for a chunk. Nodes need to be prepared to cover :math:`m` in their proximate bin. It is not realistic to have more than :math:`10-20` peers in there, so :math:`m` is effectively maximised in :math:`20k` where :math:`k` is the maximum stake per chunk for a single node.
With these constraints the maximum deposit a node can put up is: :math:`2*20ps + 2ps=42ps`,
choosing :math:`s=10`, this gives :math:`420p`.


Pricing storage in units of chunk retrieval
---------------------------------------------

With the scheme laid out in the previous section we established an implicit insurance system where

* all costs and obligation can be settled between connected peers
* signed promises (commitments that can be used to initiate litigation) are available as immediate responses to store requests (syncing)

As a consequence, all payment or accounting for storage promises can be done exactly the same way as with bandwidth. Subsuming settlement of storage expenses under SWAP is a major advantage and simplifies our system.

However, unlike in the case of retrieval, storage receipts represent an insurance of sorts and therefore their pricing is important. There is no sense in which chunk storage can be traded service for service one for one.

However, their price can always be expressed in terms of chunk retrievals, so SWAP can simply handle their accounting in a trivial way.

Owner-side handling of storage redundancy
==============================================================================

In the previous sections we established that replication-based redundancy can only
work under serious restrictions. A variable deposit scheme leads to difficult accounting anomalies and at best workable in the initial round of syncing when after uploading the chunk reaches the closest node. If the network grows the same problems emerge with splitting the offered receipt price among peers. On top of this replication is limited by the size of proximate bins. Keeping the most proximate bin above a minimal size to be able to satisfy all storage requests also puts extra processing burden on nodes.

Luckily, there is an entirely different approach which makes it possible to delegate arbitrary security to the owner. The idea is that redundancy is encoded in the document structure before its chunks are uploaded. For instance the simplest method of guarateeing redundancy of a file is to chunk the file into chunks that are one byte shorter than the normal chunksize and add a nonce byte to each chunk. This guarantees that each chunk is different and as a consequence all chunks of the modified file is different. When joining the last byte of each chunk is ignored so all variants map to the same original.
Assuming all chunks of the original file are different this yields a potential  :math:`256^x` equivalent replicas the owner can upload.

Note also that if we replicate a chunk only at its neighbourhood, but that particular chunk is crucial in the reconstruction of the content, the swarm is much more vulnerable to chunk loss or latency due to attacks. This is because if the storers of the replicas are close, inflitrating in the storers' neighbourhood can be done with as many nodes as chunk type (as opposed to as many as chunk replicas). If there is a measure to protect against sybil attacks this brings down the cost by a factor of n where n is the number of replicas.

Luckily there are a lot more economical ways to encode a file redundantly.

Importantly however, we do not want to replace local replication completely. Although the cloud industry is trying very hard to get away from the explicit x-fold redundancy model because it is very wasteful and incurs high costs – erasure coding can guarantee the same level of security using only a fraction of the storage space. However, in a data center redundancy is interpreted in the context of hard drives whose failure rates are low, independent and predictable and their connectivity is almost guaranteed at highest possible speed due to proximity. In a peer-to-peer network scenario however, nodes could disappear much more frequently than hard drives fail. In the beginning,  we may expect larger than n replicas of chunks, but as the swarm grows and storage space is filling up, redundancy will drop automatically. Erasure coding is then the best way to ensure file availability. Incidentally, redundant coding offers further benefits of increased resilience and ways to speed up retrieval.

In what follows we spell out our proposal to introduce a per-level m-of-n Cauchy-Reed-Solomon erasure code into the swarm trie.

The Cauchy-Reed-Solomon (henceforth CRS) scheme is a systemic erasure codes capable of implementing a scheme whereby any :math:`m` out of :math:`n` fix-sized pieces are able to reconstruct the original data blob of size :math:`m` pieces with storage overhead of :math:`n-m`.[#]_ Once we got the :math:`m` pieces of the original blob, CRS scheme provides a method to inflate it to size :math:`n` by supplementing :math:`n-m` so called parity pieces. With that done, assuming `p` is the probability of losing one piece, if all :math:`n` pieces are independently stored, the probability of loosing the original content is :math:`p^{n-m+1}` exponential while extra storage is linear. These properties are preserved if we apply the coding to every level of a swarm trie.

.. rubric:: There are open source libraries to do Reed Solomon or Cauchy-Reed Solomon encoding. See https://www.usenix.org/legacy/event/fast09/tech/full_papers/plank/plank_html/, https://www.backblaze.com/blog/reed-solomon/, http://rscode.sourceforge.net/

The chunker algorithm would proceed the following way when splitting the document:

0. Set input to the data blob.
1. Read the input 4096 byte chunks at a time. Count the chunks by incrementing :math:`i`
  IF fewer than 4096 bytes are left in the file, fill up the last fraction to 4096
2. Repeat 1 until there's no more data or :math:`i%m=0`
3. If there is no more data add padding of :math:`j` chunks such that :math:`i+j%m=0`.
3. use the CRS scheme on the last :math:`m` chunks to produce :math:`128-m` parity chunks resulting in a total of 128 chunks.
4. Record the hashes of the 128 chunks cocatenated to result in the next 4096 byte chunk of the next level.
5. If there is more data repeat 1. otherwise
6. If the next level data blob is of size larger than 4096, set the input to this and  repeat from 1.
7. Otherwise remember the blob as  the root chunk

The swarm trie also includes a file-size integer I believe. I do not think this is necessary; it should only be necessary to supply a filesize int along with the root hash. This then allows everyone to calculate what the Cauchy-Reed-Solomon redundancy coding is at every node and also which nodes are original file data and which are parity data.

Benefits of CRS merkle trie
====================================

All chunks are created equal
------------------------------
A trie encoded as suggested above has the same (*) redundancy at every node. This means that chunks nearer to the root are no longer more important than chunks near the file. Every node as an m-of-128 redundancy level and no chunk after the root chunk is more important than any other chunk.

(*) If the filesize is not a specific multiple of 4096 bytes, then the last chunk at every level will actually have a higher redundancy even than the rest.

Self healing
---------------------------

Any(!) client downloading a file from the swarm can detect if a chunk has been lost. The client can reconstruct the file from the parity data (or reconstruct the parity data from the file) and re-sync this data into the swarm. That way, even if a large fraction of the swarm is wiped out simultaneously, this process should allow an organic healing process to occur and it is encouraged that the default client behavior should be to repair any damage detected.

Improving latecy of retrievals
---------------------------------------------

Alpha is the name Kademlia gives to the number of peers in a Kademlia bin that are queried simultaneously during a lookup. The original Kademlia paper sets alpha=3. This is impractical for Swarm because the peers do not report back with new addresses as they would do in pure Kademlia but instead forward all queries to their peers. Swarm is coded in this way to make use of semi-stable longer-term ethp2p connections. Setting alpha to anything greater than 1 thus increases the amount of network traffic substantially – setting up an exponential cascade of forwarded lookups.
[ Lookups would cause an exponentially growing cascade at first but it would soon collapse back down onto the target of the lookup. ]
However, setting alpha=1 has its own downsides. For starters, lookups can stall if they are forwarded to a dead node and even if all nodes are live, there could be large delays before a query is complete. The practice of setting alpha=2 in swarm is designed to speed up file retrieval and clients are configured to accept chunks from the first/fastest forwarding connection to be established.
In an erasure coded setting we can in a sense have a best of both worlds. The default behavior should be the set alpha=1 i.e. to query one peer only for each chunk lookup, but crucially, to issue a lookup request not just for the data chunks but for the parity chunks as well. The client then could accept the first m of every 128 chunks queried to get some of the same benefits of faster retrieval that redundant lookups provide without a whole exponential cascade.

Improving resilience in case of non-saturated Kademlia table
-----------------------------------------------------------------

Earlier version

Not all chunks (in the Merkle Trie) are created equal
------------------------------------------------------

When we encode a file in Swarmz, the chunks that represent nodes near the root of the tree are in some sense more important than the nodes nearer to the file layer. More specifically, if the root chunk is lost, then the entire file is lost; if one of the following chunks on the next level is lost, then 1/128 of the file is lost and so on.
In many cases this distinction may seem unimportant. For example, if the file uploaded is compressed then every chunk is as important to me as any other because with even one chunk missing one is unable to uncompress the file. Ultimately we want every chunk to be equally important to any other chunk.


Loss-tolerant Merkle Trees
----------------------------------------------------------

Recall that each node (except possibly the last one on each level) has 128 children each of which represent the root hash of a subtree or, at the last level, represent a 4096 byte span of the file. Let us now suppose that we divide our file into 100 equally sized pieces, and then add 28 more parity check pieces using a Reed-Solomon code so that now any 100 of the 128 pieces are sufficient to reconstruct the file. On the next level up the chunks are composed of the hashes of their first hunder data chunks and the 28 hashes of the parity chunks. Let's take the first 100 of these and add 28 parity chunks such that each 100 of the resultig chunks can reconstruct the origial 100 chunks. And so on every level.
In terms of availability, every subtree is equally important to every other subtree at this level. The resulting data structure is not balanced tree since on every level :math:`i` the last 28 chunks are parity leaf chunks while the first 100 are branching nodes encoding a subtree of depth :math:`i-1` redundanly.

In practice of course, data chunks are still prefered over the parity chunks in order to avoid CPU overhead in reconstruction. This data structure has preserved its merkle properties and can be used for partial integrity check.


The effect of the encoding on Insurance Pricing
--------------------------------------------------

Using erasure codes in this way allows us to do away with our previous redundancy pricing model. We can posit that any chunk must be stored at the threeclosest nodes. Multiple copies exist not only for redundancy purposes but also to make sure that retrieval succeds. Beyond that, if a user requires a higher level of insurance that their file will remain available, the user may set the parameters of the erasure code (that was the 100 in the example above). In this way, a lower number means more security but also means a larger file size and more chunks and thus a higher insurance cost.


The effect of the encoding on retrieval pricing and storage incentives
-------------------------------------------------------------------------

A problem that immediately presents itself is the following: if nodes are compensated only for serving chunks, then less popular chunks are less profitable and more likely to be deleted; therefore, if users only download the 100 data chunks and never request the parity chunks, then these are more likely to get deleted and ultimately not be available when they are finally needed.

Another approach would be to use non-systemic coding. Recall that a systemic code is one in which the data remains in tact and we add extra parity data whereas in a non-systemic code we replace all data with parity date such that (in our example) all 128 pieces are really created equal. While the symmetry of this approach is appealing, this leads to forced decoding and thus to a high CPU usage and it also prevents us from streaming files from the swarm. Since we anticipate that streaming will be a common usage, we cannot choose any non-systemic code and would have to find/choose codes that are streamable.


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

