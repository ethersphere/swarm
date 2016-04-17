.. _sec:storage:

******************************
Storage Incentives
******************************

..  index::
  litigation

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
especially if the user explicitly requires that in the fashion of 'upload and disappear' discussed in the introduction.

..
    One proposed solution to this is Filecoin (:cite:`filecoin2014`), which can be earned (mined) through replicating other people's content and spent on having one's content replicated.
    From the perspective of the content creator, "upload and disappear" goes as
    follows: they first have to host their own content as an IPFS node and then they
    insert a special transaction into the filecoin blockchain offering a
    mining reward for those who replicate it. Then they wait until someone
    actually does the replication (i.e. inserts their transaction into the
    filecoin blockchain) and then they can disconnect. If nobody replicates,
    their course of action is to submit further transactions, offering more
    reward, until someone finally does.

:dfn:`Filecoin` (:cite:`filecoin2014`), an incentivised p2p storage network using IPFS (:cite:`ipfs2014`)offers an interesting solution. Nodes participating in its network also mine
on the filecoin blockchain. Filecoin can be earned (mined) through replicating other people's content and spent on having one's content replicated.
Filecoin's proof of work is defined to include proof that the miner possesses a set of randomly chosen units of storage depending on the parent block.
Using a strong proof of retrievability scheme, Filecoin ensures that the winning miner had relevant data. As miners compete, they will find that their chances of winning will be proportional to the percentage of the existing storage units they actually store. This is because the missing ones need to be retrieved from other nodes and thus delaying nodes chance to respond.

We see a whole range of issues with this particular approach:

* It is not clear that network latency cannot be masked by the parallel calculation of the ordinary proof of work component in the algorithm.
* If the set of chunks are not selected differently for each node, mining will resemble a DDOS on nodes that actually store the data needed for the round.
* Even if the selection of data to prove varies depending on the miner, normal operation incurs huge network traffic.
* As the network grows, the expected proportion of the data that needs to be retrieved increases. In fact given a practical maximum limit on a node's storage capacity, this proportion reaches a ceiling. If that happens miners will end up effectively competing on bandwidth.
* In order to check proof of retrievability responses as part of block validation, existing data needs to be recorded on the blockchain. This leads to excessive use of the blockchain as the network grows and is unlikely to scale.
* Competing miners working on the same task mean redundant use of resources.
* If content is known to be popular, checking their integrity is spurious. But if choice of storage data to audit for the next block is truely random, there is no distinction between rarely accessed content and popular ones stored by many nodes resulting in wasted resouces.
* Similarly, users originating the content have also no way to indicate directly that some documents are important and not to be lost, while other temporary or derived data they can afford to lose.

Due to excessive use of blockchain and generated network traffic, these issues make the approach suspect: at best hugely wasteful, at worst infeasible on the large scale.

More importantly, however, Filecoin provides only a scheme to collectively incentivise the network to store content. This brings in a 'tragedy of the commons' problem in that losing any particular data will have no negative consequence to any one storer node. This lack of individual accountability means the solution is rather limited as a security measure against lost content.

To summarise, we consider positive incentivisation in itself insufficient for ensured archival. In addition to that collective positive incentivisation implemented by competitive proof of retrievability mining is wasteful in terms of network traffic, computational resources as well as blockchain storage. In the subsequent sections we will introduce a different approach.


Compensation for storage and guarantees for long-term data preservation
===========================================================================

While Swarm's core storage component is analogous to traditional DHTs both in terms of network topology and routing used in retrieval, it uses the narrowest interpretation of immutable content addressed archive. Instead of just metadata about the whereabouts of the addressed content, the proximate nodes actually store the data itself.
When a new chunk enters the swarm storage network, it is propagated from node to node via a process called 'syncing'. The goal is for chunks to end up at nodes whose address is closest to the chunk hash. This way chunks can be located later for retrieval using kademlia key-based routing.

..  index::
   retrieve request
   latency

The primary incentive mechanism in swarm is compensation for retrieval where nodes are rewarded for successfully serving a chunk. This reward mechanism has the added benefit of ensuring that the popular content becomes widely distributed (by profit maximising storage nodes serving popular content they get queried for) and as a result retrieval latency is descreased.

The flipside of using only this incentive on it own is that chunks that are rarely retrieved may end up lost. If a chunk is not being accessed for a long time, then as a result of limited storage capacity it will eventually end up garbage collected to make room for new arrivals. In order for the swarm to guarantee long-term availability, the incentive system needs to make sure that additional revenue is generated for chunks that would otherwise be deleted. In other words, unpopular chunks that do not generate sufficient profit from retrievals should compensate the nodes that store them for their opportunities forgone.

A long-term storage incentivisation scheme faces unique challenges. For example, unlike in the case of bandwidth incentives where retrievals are immediately accounted and settled, long-term storage guarantees are promisory in nature and deciding if the promise was kept can only be decided at the end of its validity. Loss of reputation is not an available deterrent against foul play in these instances: since new nodes need to be allowed to provide services right away, cheaters could just resort to new identities and keep selling (empty) storage promises.

..  index::
  reputation
  punative measures
  deposit

Instead, we need punitive measures to ensure compliance with storage promises. These will work using a :dfn:`deposit system`. Nodes wanting to sell promisory storage guarantees should have a *stake verified and locked-in* at the time of making their promise. This implies  that nodes must be *registered* in advance with a contract and put up a :dfn:`security deposit`.

Following :dfn:`registration`, a node may sell storage promises covering the time period for which their funds are locked. While their registration is active, if they are found to have lost a chunk that was covered by their promise, they stand to loose their deposit.

In this context, :dfn:`owner` refers to the originator of a chunk (the one that uploads a document to the swarm), while :dfn:`storer` refers to a swarm node that actually stores the given chunk.

 Let us start from some reasonable guiding principles:

* owners need to express their risk preference when submitting to storage
* storers need to express their risk preference when committing to storage
* there needs to be a reasonable market mechanism to match demand and supply
* there needs to be guarantees  for the owner that its content is securely stored
* there needs to be a litigation system where storers can be charged for not keeping their promise

Owners' risk preference consist in the time period covered as well as a preference for the :dfn:`degrees of redundancy` or certainty. These preferences should be specified on a per-chunk basis and they should be completely flexible on the protocol level.

Satisfying storers' risk preferences means that they have ways to express their certainty of preserving what they store and factor that in their pricing. Some nodes may not wish to provide storage guarantees that are too long term while others cannot afford to stake too big of a deposit. This differentiates nodes in their competition for service provision.

A *market mechanism* means there is flexible :dfn:`price negotation` or discovery or automatic feedback loops that tend to respond to changes in supply and demand.

..  index:: litigation

A :dfn:`litigation` procedure necessitates that there are contractual agreements between parties ultimately linking an owner who pays for securing future availability of content and a storer who gets rewarded for preserving it and making it immediately accessible at any point in the future. The incentive structure needs to make sure that litigation is a last resort option.

It is also worth emphasizing that the producer and the consumer of the information may not be the same entity and it is therefore important that failure to make good on the promise to deliver the stored content is penalized even when the unserved consumer was not party to the agreement to store and provide the requested content. Litigation therefore is expected to be available to third parties wishing to retrieve content.

..  index::
   contract
   receipt

The simplest solution to manage storage deals is using direct contracts between owner and storer. This can be implemented with storers returning :dfn:`signed receipts` of chunks they accept to store and owners paying for the receipts either directly or via escrow.
In the latter case, storer only gets awarded the locked funds if they provide proof that the chunk they stored is valid. Such delayed payment solutions would enable operation entirely without litigation. The receipts collected can used to prove commitment in case of litigation. There are other more indirect variants of litigation which do not rely on owner and storer being in direct contractual agreement, which is the case if the eventual consumer is distinct from the storer and not known to them in advance.

In what follows we will elaborate on a class of incentive schemes we call :dfn:`swap, swear and swindle` due to the basic components:

:dfn:`swap`
  Nodes are in semipermanent long term contact with their registered peers. Along these connections the peers are swapping various pieces of information relating to syncing, receipting, price negotiation, auditing and offchain payments.

:dfn:`swear`
  Nodes registered on the swarm network are accountable and stand to lose their deposit if they are found to violate the rules of the swarm in an on-chain litigation process.

:dfn:`swindle`
   A scheme to pool resources to enforce adherence to the rules, by regular auditing, policing, and eventually conscientious litigation.

..  swindle

As we go along, these names will reveal their secondary meanings.

Security begins at home and so the first step in securing data begins with the owner; this is the topic of the following section. Then in section :ref:`sec:swear` we describe how the owner hands over custody of their data to registered nodes in the swarm subject to an insurance contract. Finally, in section :ref:`sec:swindle:`, we turn to how such insurance is enforced by the ethereum smart contract based litigation system (SWINDLE).

Owner-side handling of storage redundancy
==============================================================================

First we show how to delegate setting arbitrary :dfn:`levels of storage security` to the owner. The idea is that :dfn:`redundancy` is encoded in the document structure before its chunks are uploaded.
This is important since this entails that the degree of redundancy does not need to be among the parameters handled by store requests, pricing or litigation.

A simplistic method of guarateeing redundancy of a file is to split the file into chunks that are one byte shorter than the normal chunksize and add a nonce byte to each chunk. This guarantees that each chunk is different and as a consequence all chunks of the modified file are different. When joining the last byte of each chunk is ignored so all variants map to the same original.
This yields a potential :math:`256` equivalent replicas of each chunk for the owner to upload (and up to :math:`256^x` different root hashes) [#]_ .

..  rubric:: Footnotes
.. [#] We also explored the possibility that degree of redundancy is subsumed under local replication (section :ref:`sec:localreplication`). Local replicas are instances of a chunk stored by nodes in a close neighbourhood. If that particular chunk is crucial in the reconstruction of the content, the swarm is much more vulnerable to chunk loss or latency due to attacks. This is because if the storers of the replicas are close, inflitrating in the storers' neighbourhood can be done with as many nodes as chunk type (as opposed to as many as chunk replicas). If there is cost to sybil attacks this brings down the cost by a factor of n where n is the number of replicas. We concluded that local replication is important for resilience in case of intermittent node dropouts, however, inferior to other solutions at implementing levels of security.

Luckily there are a lot more economical ways to encode data redundantly. In what follows we spell out our proposal to introduce a scheme for a *loss tolerant swarm hash*.


Loss-tolerant Merkle Trees and Erasure Codes
-------------------------------------------------

Recall that the basic data structure in swarm is a :dfn:`merkle tree`. Assuming :math:`h` the size of the hash output of the hash function used in bytes, :math:`b` is the branching factor. Each node represents the root hash of a subtree or, at the last level, the hash of a :math:`b*h` long span (one chunk) of the file. Generically we may think of each chunk as consisting of :math:`b` hashes:

..  _fig:chunk:

..  figure:: fig/chunk.pdf
    :align: center
    :alt: A chunk consisting of 128 hashes
    :figclass: align-center

    A chunk consists of 4096 bytes of the file or a sequence of 128 subtree hashes.

while in the tree structure, the 32 bytes stored at the node represent the hash of the 128 children.

..  _fig:treebasic:

..  figure:: fig/treebasic.pdf

    :align: center
    :alt: a generic node in the tree has 128 children
    :figclass: align-center

  A generic node in the tree has 128 childern.

Recall also that during normal swarm lookups, a swarm client performs a lookup for a hash value and receives a chunk in return. This chunk in turn constitutes another 128 hashes to be looked up in return for another 128 hashes and so on until the chunks received belong to the actual document. Here is a schematic: (:numref:`Figure %s <fig:tree2>`):

..  _fig:tree2:

..  figure:: fig/tree2.pdf

    :align: center
    :alt: the swarm tree
    :figclass: align-center

    The swarm tree is the data structure encoding how a document is split into chunks.

..  index::


We propose using a Cauchy-Reed-Solomon (henceforth :abbr:`CRS (Cauchy-Reed-Solomon)`) scheme to encode redundancy directly into the swarm tree. The :dfn:`CRS scheme` [#]_  is a :dfn:`systemic erasure code` which applied to a data blob of :math:`m` fixed-size pieces, produces :math:`k` extra pieces (so called :dfn:`parity pieces`) of the same size in such a way that any :math:`m` out of :math:`n=m+k` fix-sized pieces are to reconstruct the original blob with storage overhead of :math:`\frac{k}{m}`.

.. rubric:: Footnotes
.. [#] There are open source libraries that implement Reed Solomon or Cauchy-Reed-Solomon coding. See https://www.usenix.org/legacy/event/fast09/tech/full_papers/plank/plank_html/, https://www.backblaze.com/blog/reed-solomon/, http://rscode.sourceforge.net/.

Assuming :math:`p` is the probability of losing one piece, if all :math:`n` pieces are independently stored, the probability of loosing the original content is :math:`p^{n-m+1}` which is exponential while extra storage is linear. These properties are preserved if we apply the coding to every level of a swarm chunk tree.

The :dfn:`chunker` algorithm using :math:`mc` CRS coding would proceed the following way when splitting the document:

 1. Set input to the data blob.
 2. Read the input one chunk (say fixed 4096 bytes) at a time. Count the chunks by incrementing :math:`i`. The last chunk read may be shorter.
 3. Repeat 2 until there's no more data or :math:`i \equiv 0` mod :math:`m`
 5. use the CRS scheme on the last :math:`i \mod\ m` chunks to produce :math:`k` parity chunks resulting in a total of :math:`n \leq m+k` chunks.
 6. Calculate the hashes of all the these chunks and concatenate then to result in the next chunk (of size :math:`i\mod m` of the next level. Record this chunk as the next
 7. If there is more data repeat 2. otherwise
 8. If the next level data blob has more than one chunk, set the input to this and  repeat from 2.
 9. Otherwise remember the blob as the root chunk.

Assuming we fix the branching factor of the swarm hash (chunker) as :math:`n=128` and :math:`h=32` as the size of the :dfn:`SHA3 Keccak hash`. This gives us a chunk size of :math:`4096` bytes.

Let us now suppose that we start splitting out inpuy document data into chunks, and after each :math:`m` chunks then add :math:`k=n-m` parity check pieces using a Reed-Solomon code so that now any :math:`m\text{-out-of-}n` chunks are
sufficient to reconstruct the document. On the next level up the chunks are composed of the hashes of the :math:`m`  data chunks and the :math:`k` hashes of the parity chunks. Let’s take the first :math:`m`
of these and add an additional :math:`k` parity chunks to those such that any :math:`m` of the resulting :math:`n`
chunks are sufficient to reconstruct the origial :math:`m` chunks. And so on and on every level. In terms of
availability, every subtree is equally important to every other subtree at this level. The resulting
data structure is not a balanced tree since on every level :math:`i` the last :math:`k` chunks are parity leaf
chunks while the first :math:`m` are branching nodes encoding a subtree of depth :math:`i-1` redundantly.
A typical piece of our tree would look like this: (:numref:`Figure %s <fig:tree-with-erasure>`)

..  _fig:tree-with-erasure:

..  figure:: fig/tree-with-erasure.pdf

    :align: center
    :alt: the swarm tree with erasure coding
    :figclass: align-center

    The swarm tree with extra parity chunks using 100 out of 128 CRS code. Chunks :math:`p^{101}` through :math:`p^{128}` are parity data for chunks :math:`h^1_1 - h^1_{128}` through :math:`h^{100}_1  - h^{100}_{128}`.


Two things to note

 * This pattern repeats itself all the way down the tree. Thus hashes :math:`h^1_{101}` through :math:`h^1_{128}` point to parity data for chunks pointed to by :math:`h^1_1` through :math:`h^1_{100}`.
 * Parity chunks :math:`p^i` do not have children and so the tree structure does not have uniform depth.

The special case of the last chunks in each row
--------------------------------------------------


If the number of file chunks is not divisible by :math:`m`, then we cannot proceed with the last batch in the same way as the others. We propose that we encode the remaining chunks with an erasure code that guarantees at least the same level of security as the others. Note that it is not as simple as choosing the same redundancy. For example a 50-of-100 encoding is much more secure against file loss than a 1-of-2 encoding even though the redundancy is 100% in both cases. Overcompensating, we could say that there should always be the same number of parity chunks even when there are fewer than :math:`m` data chunks so that we always end up with :math:`m\text{-out-of-}n`. We repeat this procedure in every row in the tree.

This leaves us with only one corner case: it is not possible to use our :math:`m\text{-out-of-}n` scheme on a single chunk (:math:`m=1`) because it would amount to :math:`k+1` copies of the same chunk. The problem of course is that any number of copies of the same chunk all have the same hash and are therefore indistinguishable in the swarm. Thus when there is only a single chunk left over at some level of the tree, we'd have to apply some transformation to it to obtain a second (but different) copy before we could generate more parity chunks.

In particular this is always the case for the root chunk. To illustrate why this is critically important, consider the following. The root hash points to the root chunk. If this root chunk is lost, then the file is not retrievable from the swarm even if all other data is present. Thus we must find an additional method of securing and accessing the information stored in the root chunk.

Whenever a single chunk is left over (:math:`m=1`) we propose to append an extra padding byte to the chunk not counting towards its size. In swarm, each 4096 byte chunk is actually stored together with 8 bytes of meta information - currently only storing the size of the subtree encoded by the chunk. Since the subtree size determines exactly what span of the chunk is substantive data, the padding differential byte is easily ignored when the document is assembled [#]_ .

.. rubric:: Footnotes
.. [#] Note that the typical values for :math:`k` will be in the single digits so a single byte will allways suffice. Note that in the special cornercase when the singleton leftover chunk is a full chunk, we end up having an oversized chunk.


Benefits of CRS merkle tree
------------------------------------

This per-level :math:`m\text{-out-of-}n` Cauchy-Reed-Solomon erasure code once introduced into the swarm chunk tree does not only ensure file availability, but also offers further benefits of increased resilience and ways to speed up retrieval.

All chunks are created equal
  A tree encoded as suggested above has the same redundancy at every node [#]_. This means that chunks nearer to the root are no longer more important than chunks closer to the leaf nodes. Every node has an m-of-128 redundancy level and no chunk after the root chunk is more important than any other chunk [#]_ . Luckily the problem is solved by the automated audit scheme which audits the integrity of all chunks and does not distinguish between data or parity chunks.

.. rubric:: Footnotes
.. [#] If the filesize is not a multiple of 4096 bytes, then the last chunk at every level will actually have a higher redundancy even than the rest.
.. [#] If nodes are compensated only for serving chunks, then less popular chunks are less profitable and more likely to be deleted; therefore, if users only download the data chunks and never request the parity chunks, then these are more likely to get deleted and ultimately not be available when they are finally needed. Another approach would be to use non-systemic coding. A systemic code is one in which the data remains intact and we add extra parity data whereas in a non-systemic code we replace all data with parity data such that (in our example) all 128 pieces are really created equal. While the symmetry of the non-systemic approach is appealing, it leads to forced decoding and thus to a high CPU usage even in normal operation. Moreover it breaks random access property of the chunk tree making it impossible to stream media files from the swarm.

Self healing
  Any client downloading a file from the swarm can detect if a chunk has been lost. The client can reconstruct the file from the parity data (or reconstruct the parity data from the file) and resync this data into the swarm. That way, even if a large fraction of the swarm is wiped out simultaneously, this process should allow an organic healing process to occur and it is encouraged that the default client behavior should be to repair any damage detected.

Improving latecy of retrievals
  Alpha is the name the original :dfn:`Kademlia` (:cite:`kademlia`) gives to the number of peers in a Kademlia bin that are queried simultaneously during a lookup. The original Kademlia paper sets alpha=3. This is impractical for Swarm because the peers do not report back with new addresses as they would do in pure Kademlia but instead forward all queries to their peers. Swarm is coded in this way to make use of semi-stable longer-term devp2p connections. Setting alpha to anything greater than 1 thus increases the amount of network traffic substantially – setting up an exponential cascade of forwarded lookups (but it would soon collapse back down onto the target of the lookup).

  However, setting alpha=1 has its own downsides. For instance, lookups can stall if they are forwarded to a dead node and even if all nodes are live, there could be large delays before a query is complete.

  In an erasure coded setting we can in a sense have a best of both worlds. Issueing a lookup request not just for the data chunks but for the parity chunks, the client could accept the first :math:`m` of every 128 chunks queried to get some of the same benefits of faster retrieval that redundant lookups provide without a whole exponential cascade.

..  sec:swear:

Registered nodes and Ensured archival (SWEAR)
===================================================


Once the owner has prepared their data they upload the chunks to the swarm where they are replicated and stored. To decrease the risk that the data will be lost, the owner may purchase storage promises from other nodes as a form of insurance.
Before a node can sell these promises of long-term storage however, it must first register via a contract on the blockchain we call the :dfn:`SWEAR` (Secure Ways of Ensuring ARchival or SWarm Enforcement and Registration) contract.
The :abbr:`SWEAR (Secure Ways of Ensuring ARchival)` contract allows nodes to register their public key to become accountable participants in the swarm by putting up a deposit. Registration is done by sending the deposit to the SWEAR contract, which serves as colleteral if terms that registered nodes 'swear' to keep are violated (i.e., nodes do not keep their promise to store).
:dfn:`Registration` is valid only for a set period, at the end of which a swarm node is entitled to their deposit.
Users of Swarm should be able to count on the loss of deposit as a disincentive against foul play as long as enrolled status is granted. As a result the deposit must not be refunded before the registration expires.

..  index::
   registration
   receipt

Registration in swarm is not compulsory, it is only necessary if the node wishes to sell promises of storage. Nodes that only charge for retrieval can operate unregistered. The incentive to register and sign promises is that they can be sold for profit. When a peer connection is established, the contract state is queried to check if the remote peer is a registered node. Only registered nodes are allowed to issue valid receipts and charge for storage.

When a registered node receives a request to store a chunk, it can acknowledge accepting it with a signed receipt. It is these signed receipts that are used to enforce penalties for loss of content through the :abbr:`SWEAR (SWarm Enforcement and Registration)` contract. Because of the locked collateral backing them, the receipts  can be viewed as secured promises for storing and serving a particular chunk up until a particular date. It is these receipts that are sold to nodes initiating requests.
In some schemes the issuer of a receipt can in turn buy further promises from other nodes pontentially leading to a chain of local contracts.

If on litigation it turns out that a chunk (covered by a promise) was lost, the deposit must be at least partly burned. Note that this is necessary because if penalites were paid out as compensation to holders of receipts of lost chunks, it would provide an avenue of early exit for a registered node by "losing" bogus chunks deposited by colluding users. Since users of Swarm are interested in their information being reliably stored, their primary incentive for keeping the receipts is to keep the Swarm motivated, not the potential for compensation.
If deposits are substantial, we can get away with paying out compensation for initiating litigation, however we must have the majority (say 95%) of the deposit burned in order to make sure the easy exit route remains closed.

..  sec:swindle:

Litigation on loss of content (SWINDLE)
========================================

If a node fails to observe the rules of the swarm they 'swear' to keep, the punative measures need to be enforced which is preceded by a litigation procedure. The implementation of this process is called :abbr:`SWINDLE (SWarm INsurance Driven Litigation Engine)`.


Submitting a challenge
------------------------------


..  index::
  challenge
  refutation

Nodes provide signed receipts for stored chunks which they are allowed to charge arbitrary amounts for. The pricing and deposit model is discussed in detail in section :ref:`sec:accounting`. If a promise is not kept and a chunk is not found in the swarm anyone can report the loss by putting up a :dfn:`challenge`. The response to a challenge is a :dfn:`refutation`. Validity of the challenge as well as its refutation need to be easily verifyable by the contract.
For now, we can just assume that the litigation is started by the challenge after a user attempts to retrieve insured content and fails to find a chunk. Litigation will be discussed below in the wider context of regular integrity audits of content in the swarm.

The challenge takes the form of a transaction sent to the :dfn:`SWINDLE` (SWarm INsurance Driven Litigation Engine) relevant swarm contract in which the challenger presents the receipt(s) of the lost chunk. Any node is allowed to send a challenge for a chunk as long as they have a valid receipt for it (not necessarily issued to them).

This is analogous to a court case in which the issuers of the receipts are the defendants who are guilty until proven innocent. Similarly to a court procedure public litigation on the blockchain should be a last resort when the rules are abused despite the deterrents and positive incentives.

The same transaction also sends a deposit covering the upload of a chunk. The contract verifies if the receipt is valid, ie.,

* receipt was signed with the public key of a registered node
* the expiry date of the receipt has not passed
* sufficient funds are sent alongside to compensate the peer for uploading the chunk in case of a refuted challenge

The last point above is designed to disincentivise frivolous litigation, i.e., bombarding the blockchain with bogus challenges potentially causing a :dfn:`DoS attack`.

..  index:: DoS

A challenge is open for a fixed amount of time, the end of which essentially is the deadline to refute the challenge. The challenge is refuted if the chunk is presented (additional ways are discussed below). Refutation of a challenge is easy to validate by the contract since it only involves checking that the hash of the presented chunk matches the receipt. This challenge scheme is the simplest way (i) for the defendants to refute the challenge as well as (ii) to make the actual data available for the nodes that needs it.

In normal operation, litigation should be so rare that it may be necessary to introduce a practice of regular :dfn:`audits` to test nodes' compliance with distribution rules. In such cases the challenge can carry a flag which when set would indicate that providing the actual chunk, (ii) above, is unnecessary. In order to reduce network traffic, in such cases presenting the chunk can be replaced by providing a :dfn:`proof of custody`. Note that in order not to burden the live chain, audits could happen off chain and they would only make it to the blockchain if foul play is detected. Conversely, if such auditing is a regular automated process, then litigation will typically be initiated as part of escalating a failed audit.
:cite:`ethersphere2016smash` describes such an audit protocol using the smash proof of custody construct.

The outcome of a challenge
-------------------------------------

Successful refutation of the challenge is done by anyone sending the chunk or a proof of custody thereof as data within a transaction to the blockchain. Upon verifying the format of the refutation, the contract checks its validity by checking the hash of the chunk payload against the hash that is litigated or validating the proof of custody. If the refutation is valid, the cost of uploading the chunk is compensated from the deposit of the challenge, with the remainder refunded.

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

If however the deadline passes without successful refutation of the challenge, then the charge is regarded as proven and the case enters into enforcement stage. Nodes that are proven guilty of losing a chunk lose their deposit. Enforcement is guaranteed by the fact that deposits are locked up in the SWEAR contract.

..  index::
  suspension
  cheating

Punishment can entail :dfn:`suspension`, meaning a node found guilty is no longer considered a registered swarm node. Such a node is only able to resume selling storage receipts once they create a new identity and put up a deposit once again. Note that the stored chunks are in the proximity of the address, so having to create a new identity will imply extra bandwidth to replenish storage.This is extra pain inflicted on offending nodes.

If refutation of litigation is found to be common enough, sending transactions is not desirable since it is bloating the blockchain.
The audit challenges using the smash proof of custody described in :cite:`ethersphere2016smash` enable us to improve on this and make litigation a two step process. Upon finding a missing chunk, the litigation is started by the challenger sending an audit request [#]_ .

..  rubric:: Footnotes
.. [#] See :cite:`ethersphere2016smash` for the explanation of particular audit types. In fact any audit challenge when fail should be escalated to the blockchain. The smash smart contract provides an interface to check validity of audit requests (as challenges) and verify the various response types (as refutations).

Playing nice is further incentivized if a challenge is allowed to extend the risk of loss to all nodes that have given a promise to store the lost chunk. This means that when one storer is challenged, all nodes that have outstanding receipts covering the (allegedly) lost chunk stand to lose their deposit. Holders of receipts by other swarm nodes can punish them as well for losing the chunk, which, in turn, incentivizes whoever may hold the chunk to present it (and thus refute the challenge) even if they are not the named defendant first accused.

Owners express their preference for storage period.
As for storage period, the base unit used will be a :dfn:`swarm epoch`. The swarm epoch is the minimum interval a swarm node can register for.

Nodes can choose to gamble of course by selling storage receipts without storing the chunk, in the hope of being able to retrieve the chunk from the swarm as needed. However, since storers have no real way to trust other nodes to fall back on, the nodes that issue receipts have a strong incentive to actually store the chunk themselves. Collecting receipts from several nodes therefore means that several replicas are likely to be kept in the swarm. Slogan: more receipts means more redundancy.

A priori this only works, however, in the simplest system in which the owner needs to receive and keep all the receipts signed by the storers.

Publicly accessible receipts and consumer driven litigation
------------------------------------------------------------

End-users that store important information in the swarm have an obvious interest in keeping the receipt available for litigation. The storage space required for storing a receipt is a sizable fraction of that used for storing the information itself, so end users can reduce their storage requirement further by storing the receipts in Swarm as well. Doing this recursively would result in end users only having to store a single receipt, the root receipt, yet being able to penalize quite a few Swarm nodes, in case only a small part of their stored information is lost.

A typical usecase is when content producers would like to make sure their content is available. This is supported by implementing the process of collecting receipts and putting them together in a format which allows for the easy pairing of chunks and receipts for an entire document. Storing this document-level receipt collection in the swarm has a non-trivial added benefit. If such a pairing is public and accessible, then consumers/downloaders (not only creators/uploaders) of content are able to litigate in case a chunk is missing. On the other hand, if the likely outcome of this process is punishment for the false promise (burning the deposit), motivation to litigate for any particular bit lost is slim.

This pattern can be further extended to apply to a document collection (dapp/website level). Here all document-level root receipts (of the sort just discussed) can simply be included as metadata in the manifest entry for the document alongside its root hash. Therefore a manifest file itself can store its own warranty.
The question arises what happens if the hash of this entire collection is not found, if this is a possibility then all the effort in insuring the chunks was futile [#]_ .

.. rubric:: Footnotes
.. [#] One proposal is to introduce a special content addressed storage, whereby litigation information (notably the receipt from the guardian) is stored at an address derivable from the swarm hash. The address would be derived from the hash by flipping its first bit which would guarantee that the receipt is stored at an opposite end of the swarm. This would make litigation on the chunk level independent of document-level structures and would allow any third party to initiate audits and litigation against a loss chunk knowing only the hash. It is unclear whether this would work though: if a chunk is not found due to it not having been retrieved for some time, chances are high that the receipt has also not been accessed and has been deleted too.

Receipt forwarding or chained challenges
===========================================

In this section we zoom in on the swapping and elaborate how owners initiate storage requests, how chunks find their storers and how information is passed around between peers so that it creates an incentive compatible resilient system with last resort litigation.

Forwarding chunks
----------------------

..  index:: retrieve request

In normal swarm operation, chunks are worth storing because of the possibility that they can be profitably sold by serving retrieve requests in the future. The probability of retrieve requests for a particular chunk depends on the chunk's popularity and also, crucially, on the proximity to the node's address.

Nodes are expected to forward all chunks to nodes whose address is closer to the chunk. This :dfn:`forwarding` is the normal syncing protocol. It is compatible with the pay-for-retrieval incentivisation: once a retrieve request reaches a node, the node will either sell the chunk (if it has it) or it will pass on the retrieve request to a closer node. There is no financial loss from syncing chunks to closer nodes because once a retrieve request reaches a closer node, it will not be passed back out, it will only be passed closer. In other words, syncing only serves those retrieve requests that the node would never have profited from anyway and thus it causes no financial harm due to lost revenue.

..  index:: syncing

For insured chunks, a similar logic applies - even more sos because there is a positive incentive to sync. If a chunk does not reach its closest nodes in the swarm before someone issues a retrieval request, then the chances of the lookup failing increase and with it the chances of the chunk being reported as lost. The resulting litigation as discussed below poses a burden on all swarm nodes that have ever issued a receipt for the chunk and therefore incentivises nodes to do timely forwarding. The audit process described in :cite:`tronetal2016smash` provides additional guarantees that chunks are forwarded all the way to the proximate nodes.

Swarm assumes that nodes are content agnostic, i.e., whether a node accepts a new chunk for storage should depend only on their storage capacity [#]_ . Registered nodes have the option to indicate that they are full capacity. This effectively means they are temporarily not able to issue receipts so in the eyes of connected peers they will act as unregistered. As a result, when syncing to registered nodes, we do not take no for an answer: all chunks legitimately sent to a registered node can be considered receipted. If the node already has the chunk (received it earlier from another peer), the receipt is not paid for.

The purpose of the receipt is to prove that a node closer to the target chunk than the node itself received the chunk and will either store it or forward it.
This is exactly what synchronisation does, therefore, proving (in)correct synchronisation is
a potential substitute for receipt based litigation.
If we further stipulate that registered nodes need to sign sync state and able to prove a particular chunk was part of the synced batch, we can get away without storing individual receipts altogether and implement the persistence of receipts as part of the chunkstore mechanism on the one hand and the passing of receipts as part of the syncing mechanism on the other.

An advantage of using sync state as receipt is that when litigation takes place, one can point fingers to a node which already had the chunk at the time of syncing as long as it is registered.



.. rubric:: Footnotes
.. [#] We will use a double masking techique as a basic measure to ensure plausible deniability.

Collecting storer receipts and direct contracts
-------------------------------------------------

There are a few schemes we may employ. In the first, a storage request is forwarded from node to node until it reaches a registered node close to the chunk address. This storer node then issues a receipt which is passed back along the same route to the chunk owner.
The owner then can keep these receipts for later litigation.


Explicit direct contracts signed by storers held by owners has a lot of advantages. On top of its transparency and simplicity, this scheme enables owners to make sure that any degree of redundancy (certainty) promise is secured by deposits of distinct nodes via their signed promises. In particular it allows owners to insure their chunks against a total collateral  higher than any individual node's deposit. Also insuring a chunk against different deposits for varying periods is easy. Unfortunately, this rather transparent system has caveats.

First of all, forwarding back receipts creates a lot of network traffic. The only purpose of receipts is to be able to use them in litigation, which is very rare, rendering virtually all this traffic spurious.

Secondly, since availability of a storer node cannot always be guaranteed, getting receipts back from storers may incur indefinite delays. The owner (who submits the request) needs a receipt that can be used for litigation later. If this receipt needs to come from the storer, then the process requires an entire roundtrip.

Furthermore, deciding on storers at the time the promise is made has a major drawback.
If the storage period is long enough the network may grow and new registered nodes come online in the proximity of the chunk. It can happen that routing at retrieval will bypass this storer. Though syncing makes sure that even in these cases the chunk is passed along and reaches closest nodes, their accountability regarding this old chunk cannot be guaranteed without further complications.

To summarize, explicit transparent contracts between owner and storer necessitate forwarding back receipts which has the following caveats:

* spurious network traffic
* delayed response
* potential non-accountability after network growth


.. What is a node's incentive to forward the request? Note that denying the chunk from peers that are not in their proximate bin have no benefit in retrieval (since requests served by the peer in question would never reach the node). If nonetheless they still do not forward, searches end up not finding the chunk, and they will be challenged. Having the chunk, they can always refute the challenge and the litigation costs may not be higher than what they gained from not purchasing receipts from a closer node. However, the litigation reveals that they cheated on syncing not offering the chunk in question. Learning this will prompt peers to stop doing business with the node. Alternatively, this could even be enforced on the protocol level requiring proof of forwarding on top of presenting the chunk, to avoid suspension.

Chaining challenges
--------------------

The other model is based on the observation that establishing the link between owner and storer can be delayed, allowing it to take place at the time of litigation. Instead of waiting for receipts issued by storers, the owner direcly contracts their (registered) connected peer(s) and they immediately issue a receipt for storing a chunk.

When registered nodes connect, they are expected to have negotiated a price and from then on are obligated to give receipts for chunks that are sent their way according to the rules. This enables nodes to guarantee successful forwarding and therefore they can immediately issue receipts to the peer they receive the request from. Put in a different way, registered nodes enter into contract implicitly by connecting to the network and syncing.

..  index::
    sycing
    litigation
    forwarding
    receipt

When issuing a receipt in response to a store request a node act as the entrypoint for a chunk. In this case the node acts as the :dfn:`guardian` of the chunk in question. The receipt(s) that the owner gets from their connected peer can be used in a challenge. Since the transaction immediately settles, the owner can 'upload and disappear'. The guardian in turn obtains a receipt from the node they are forwarding to and so on creating a chain of contracts all the way to the node proxinate to the target chunk.

When it comes to litigation, the nodes play a blame game; challenged nodes defend themselves not necessarily by presenting the chunk (or proof of custody), but by presenting a receipt for said chunk issued by a registered node closer to the chunk address. Thus litigation will involve a chain of challenges with receipts pointing from owner via forwarding nodes all the way to the storer who must then present the chunk or be punished.

The litigation is thus a recursive process where one way for a node to refute a challenge is to shift responsibility and implicate another node to be the culprit.
The idea is that contracts are local between connected peers and blame is shifted along the same route as what the chunk travels during sycing (restricted to registered nodes).

The challenge is constituted in submitting a receipt for the chunk signed by a node. (Once again everybody having a receipt is able to litigate) [#]_ .
Litigation starts with a node submitting a receipt for the chunk that is not found.
This will likely be the receipt(s) that the owner received directly from the guardian. The node implicated can refute the challenge by sending either the direct refutation (audit response or the chunk itself depending on the size and stage) to the blockchain as explained above or sending a receipt for the chunk signed by another node. This receipt needs to be issued by a node closer to the target. In other words, if a node is accused with a receipt with deposit value of X, it needs to provide valid receipts from closer nodes with deposit totalling X or more. These validations are easy to carry out, so verification of chained challenges is perfectly feasible to add to the smart contract.

.. rubric:: Footnotes
.. [#] There is no measure to prevent double receipting, i.e., the same node can sell storage insurance about the same chunks to different parties.

If a node is unable to produce either the refutation or the receipts, it is considered a proof that the node had the chunk, should have kept it but deleted it. This process will end up blaming a single node for the loss. We call these landing nodes :dfn:`custodians`. If syncronisation was correctly followed and all the nodes forwarding kept their receipt, then eventually the blame will point to the node that was closest to the chunk to be stored at the time the request was received.
if an audit request for a a chunk is not responded to, the audit request is delegated to the guardian, and travels the same trajectory as that the original store request  (see :numref:`Figure %s <fig:normaloperations>`). Analogously, if
a chunk is not found and the case is escalated to litigation on the blockchain, then finger pointing will also follow the same path (see :numref:`Figure %s <fig:failure-and-audit>`) [#]_ .

.. rubric:: Footnotes
.. [#] In the latter case the transaction is more metaphorical, finger pointing is mediated by state changes in the blockchain: when a node gets notified of a challenge (via a log event) they are sending in their receipts as a refutation and as a result the new closer node gets challenged.


..  _fig:normaloperations:

..  figure:: fig/normaloperations.pdf
    :align: center
    :alt: chain of local peer to peer interactions
    :figclass: align-center

    The arrows represent local transactions between connected peers. In normal operation these transactions involve the farther nodes (1) sending store request (2) receiving delivery request (3) sending payment (4) receiving a receipt.

..  _fig:failure-and-audit:

..  figure:: fig/failure-and-audit.pdf
    :align: center
    :alt: chain of local peer to peer interactions
    :figclass: align-center

    The arrows represent local transactions between connected peers. Following a failed lookup (left), the guardian is sent an audit/request and the edges correspond to audit requests forwarded to the peer that the node originally got the receipt from (right). Analogously when a case is escalated to litigation on the blockchain, the chain of challenges follow the same trajectory.


When the network grows, it can happen that a custodian finds a new registered node closer to its chunk. This means they need to forward the original store request, the moment they obtain a receipt they can use it in finger pointing, they cease to be custodians and the ball is in the new custodian's court.

.. _sec:localreplication

Multiple receipts and local replication
----------------------------------------------

As discussed above owners can manage the desired security level by using erasure coding with arbitrary degree of redundancy. However, it still makes sense to require that more than one node actually store the chunk. Although the cloud industry is trying to get away from the explicit x-fold redundancy model because it is very wasteful and incurs high costs – erasure coding can guarantee the same level of security using only a fraction of the storage space. However, in a data center, redundancy is interpreted in the context of hard drives whose failure rates are low, independent and predictable and their connectivity is almost guaranteed at highest possible speed due to proximity. In a peer-to-peer network scenario, nodes could disappear much more frequently than hard drives fail. In order to guarantee robust operation, we need to require several local replicas of each chunk (commonly 3, see :cite:`wilkinson2014metadisk`). Since the syncing protocol already provides replication across the proximate bin, regular resyncing of the chunk may be sufficient to ensure availability in case the custodian drops off. If this proves too weak in practice we may require the custodian to get receipts from two proximate peers who act as cocustodians. The added benefit of this extra complexity is unclear.


.. _sec:accounting:

Pricing, deposit, accounting
=============================

In this section we explore the pricing, accounting and settlement of storage services.
We conclude that the fully featured version of the SWAP protocol is ideal to manage both
unregistered use as well as registered use, delayed as well as immediate payments.

Pricing
----------------

We posited in the introduction that registered nodes should be allowed to compete in quality of service and factor their certainty of storage in their prices. Market pricing of storage is all the more important once we realise that unlike gas, system-wide fixed storage price is neither easy nor necessary.

:dfn:`Gas` is the accounting unit of computation on the ethereum blockchain, it is paid in as ether sent with the transaction and paid out in ether to the miner as part of the protocol.
The actual price of gas for a block is fixed system-wide yet it is dictated by the market. It needs to be fixed since accounting for computation needs to be identical across all nodes of the network. It still can be dictated by the market since the miners the providers of the service gas is supposed to pay for, have a way to 'vote' on it. Miners of a block can change the gas price (based on how full the block is) [#]_ . Also such a mechanism of voting by service providers is not available. Note that in principle there is some information on the blockchain which could be used to inform prices: the number of (successful) litigations. If there is an increase in the percentage of litigations (number of proven charges normalised by the number of registered nodes), that is indication that system capacity is lower than the demand, therefore prices need to rise.
The other direction, however, when prices need to decrease has no such indicator: due to the floor effect of no litigation (quite expected normal operation), information on the blockchain is inconsequential as to whether the storage is overpriced.

.. rubric:: Footnotes
.. [#] To mitigate against extreme price volatilty, one can regulate the price by introducing restrictions on rate of change (absolute upper limit of percentage of change allowed from block to block).

Hence we conclude that fixed pricing of storage is not viable without central authority or trusted third parties. Instead we assume that storage price is negotiated between peers and accepting the protocol handshake and establishing the swarm connection implicitly constitutes an arrangement.


Deposit
---------------------

Another important decision is whether maximum deposits staked for a single chunk should vary independently of price. It is hard to conceptualise what this would mean in the first place. Assume that nodes' deposit varies and affects the probability that they are chosen as storers: a peer is chosen whose deposit is higher out of two advertising the same price. In this case, the nodes have an incentive to up the ante, and start a bidding war. In case of normal operation, this bidding would not be measuring confidence in quality of service but would simply reflect wealth.
We conclude that prices should be variable and entirely up to the node, but higher confidence or certainty should also be reflected directly in the amount of deposit they stake: deposit staked per chunk should be a constant multiple of the price.

Assuming  :math:`s` is a system-wide security constant dictating the ratio between price and deposit staked in case of loss, for an advertised price of :math:`p`, the minimum deposit [#]_ is :math:`d=s\cdot p`. Price per chunk per epoch is freely configurable and dictated by supply and demand in the free market. Nodes are free to follow any price oracle or form cartels agreeing on price.

.. rubric:: Footnotes
.. [#] Although it never matters if the deposit is above the minimum, but it can happen that a peer wants to lower their price without liquidating their funds in anticipation of an opportunity to raise prices in the future.


Accounting and settlement
------------------------------

In the context of contractual agreements, forwarding of a chunk is equivalent to subcontracting for service provision that has a price. Since receipts are promises about the future, it is not in the interest of the buyer to pay before the promise is proved to have been kept. However, delayed payments without locked funds leave storers vulnerable to non-payment.

In order to lock funds nodes could use an escrow contract on the blockchain, however, burdening the blockchain with pairwise accounting is unnecessary. With a :dfn:`two-way payment channel`, the parties can safely lock parts of their balance as well as do accounting off chain.

.. index:: payment channel

Advance payments (i.e., payment settled at the time of contracting, not after the storage period ends) on the other hand, leave the buyers vulnerable to cheating.
Without limiting the total value of receipts that nodes can sell, a malicious node can collect more than their deposit and disappear. Though forfeiting their deposit, they walk away with a profit even though they broke their promise. Given a network size and a relatively steady demand for insured storage (in chunk epoch), the deposit could be set sufficiently high so this attack is no longer economical [#]_ .

.. rubric:: Footnotes
.. [#] This could be further improved by enforcing a fixed maximum total value of receipts one node can issue. Without central registry, we need to rely on the receipts. We stipulate that receipts issued by storers contain their cumulative volume of receipted promises (counted in chunk-epoch). They would also report that number to the blockchain every epoch and keep it under a threshold. The node is incentivised to underreport this number but that can be detected and punished (any node who received a higher number, sends their receipt to the blockchain). Likewise, it can also be detected if the node issued two subsequent receipts with non-increasing ranges, hence the current volume can be considered trusted. In the special case that each chunk is insured for the same length period, the current value of insured storage (counted in chunk-epochs) can be calculated since volume = cumulative volume - cumulative expired volume. Thanks to Nick Johnson for proposing this idea.

Another idea is to allow payment by installments, which would similarly keep the total income under a threshold. However, this means that the validity of a receipt can no longer be established, since non-payment of any of the obligations would void the contract.

We can combine the best of both worlds. On the one hand we can lock the total price of storing a chunk for the entire storage period, and tie the release of funds to an escrow condition, eliminates the non-payment attack.
As long as funds are locked and the escrow condition is acceptable for the storer, the settlement is immediate and they can safely issue a receipt for the entire storage period.
Since payment is delayed it is no longer possible to collect funds before the work is complete, which eliminates a :dfn:`collect-and-run attack` entirely.
Release of locked funds in installments can be tied to audits via the escrow release conditions, i.e., the installment is released on the condition that the node provides a proof of custody.

The enhanced version of the SWAP protocol uses a fully-fledged state-channel/payment channel beside the chequebook and is a perfect candidate for implementing these features.
The blockchain implementation and configuration of the payment channel, registration and litigation is discussed in a separate paper.

*************************
Conclusion
*************************

This paper explored ways of incentivising smooth operation in a peer to peer document storage and content delivery system and honed in on a particular proposal for swarm, an ethereum base layer service.
Our approach uses SWAP, the Swarm Accounting Protocol to do pairwise accounting of micropayments relevant in charging for bandwidth. The channel allows swapping service for service  in chunk retrieval and allows joining the network without funds. A chequebook contract is used to issue cheques as instruments of delayed payments, which can be cashed by the counterparty at any point to redeem promised funds as long as the sender is solvent.
Data preservation in long term storage is incentivised on an individual level both by compensation as well as penalty in case of chunk loss. The loss of insured tokens is a major offence punishable by suspension of account and forfeiture of application-global deposit.
Various ways of escrow conditions on the release of funds are able to capture quite a few usecases including pay in installments depending on successful audit. Swarm secures storage with proof of custody audits, valid audits can be tied to escrow conditions of delayed payments. Combine it with locked funds, immediate settlement and receipting, ad it offers the upload and disappear mode of operandi.
We presented a technique of erasure coding applicable to the swarm hash tree, which makes it possible for clients to manage the level of storage security and redundancy within their own remit.



.. bibliography:: ../refs.bib
   :cited:
   :style: alpha
