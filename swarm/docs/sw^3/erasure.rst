
Loss-tolerant Merkle Trees and Erasure Codes
-------------------------------------------------

Recall that the basic data structure in swarm is that of a tree with 32 bytes at the nodes and 128 children per node. Each node represents the root hash of a subtree or, at the last level, the hash of a 4096 byte span (one chunk) of the file. Generically we may think of each chunk as consisting of 128 hashes: 

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
    :alt: a generic node in the tree has 128 children.
    :figclass: align-center

	A generic node in the tree has 128 childern.
    
Recall also that during normal swarm lookups, a swarm client performs a lookup for a hash value and receives a chunk in return. This chunk in turn constitutes another 128 hashes to be looked up in return for another 128 hashes and so on until the chunks received belong to the actual file. Here is a schematic: (:numref:`Figure %s <fig:tree2>`):

..  _fig:tree2:

..  figure:: fig/tree2.pdf
    :align: center
    :alt: the swarm tree
    :figclass: align-center

    The swarm tree broken into chunks.


Erasure coding the Swarm Tree
===================================

We propose using a Cauchy-Reed-Solomon (henceforth CRS) scheme to encode redundancy directly into the swarm tree. The CRS scheme is a systemic erasure code capable of implementing a scheme whereby any :math:`m` out of :math:`n` fix-sized pieces are sufficient to reconstruct the original data blob of size :math:`m` pieces with storage overhead of :math:`n-m` [#]_ .

.. rubric:: Footnotes
.. [#] There are open source libraries to do Reed Solomon or Cauchy-Reed-Solomon coding. See https://www.usenix.org/legacy/event/fast09/tech/full_papers/plank/plank_html/, https://www.backblaze.com/blog/reed-solomon/, http://rscode.sourceforge.net/. 

Once we have the :math:`m` pieces of the original blob, CRS scheme provides a method to inflate it to size :math:`n`  by supplementing :math:`n-m` so called parity pieces. With that done, assuming :math:`p` is the probability of losing one piece, if all :math:`n` pieces are independently stored, the probability of loosing the original content is :math:`p^{n-m+1}` exponential while extra storage is linear. These properties are preserved if we apply the coding to every level of a swarm chunk tree.

Assuming we fix :math:`n=128` the branching factor of the swarm hash (chunker).
The chunker algorithm would proceed the following way when splitting the document:

 1. Set input to the data blob.
 2. Read the input 4096 byte chunks at a time. Count the chunks by incrementing :math:`i`.
  IF fewer than 4096 bytes are left in the file, fill up the last fraction to 4096 
 3. Repeat 1 until there's no more data or :math:`i \equiv 0` mod :math:`m`
 4. If there is no more data add padding of :math:`j` chunks such that :math:`i+j \equiv 0` mod :math:`m`.
 5. use the CRS scheme on the last :math:`m` chunks to produce :math:`128-m` parity chunks resulting in a total of 128 chunks.
 6. Record the hashes of the 128 chunks concatenated to result in the next 4096 byte chunk of the next level.
 7. If there is more data repeat 1. otherwise
 8. If the next level data blob is of size larger than 4096, set the input to this and  repeat from 1.
 9. Otherwise remember the blob as the root chunk


Let us now suppose that we divide our file into 100 equally sized pieces, and then add 28
more parity check pieces using a Reed-Solomon code so that now any 100 of the 128 pieces are
sufficient to reconstruct the file. On the next level up the chunks are composed of the hashes of
their first hundered data chunks and the 28 hashes of the parity chunks. Let’s take the first 100
of these and add an additional 28 parity chunks to those such that any 100 of the resulting 128
chunks are sufficient to reconstruct the origial 100 chunks. And so on on every level. In terms of
availability, every subtree is equally important to every other subtree at this level. The resulting
data structure is not a balanced tree since on every level i the last 28 chunks are parity leaf
chunks while the first 100 are branching nodes encoding a subtree of depth :math:`i-1` redundantly.
Then a typical piece of our tree would look like this: (:numref:`Figure %s <fig:tree-with-erasure>`)

..  _fig:tree-with-erasure:

..  figure:: fig/tree-with-erasure.pdf
    :align: center
    :alt: the swarm tree with erasure coding
    :figclass: align-center

    The swarm tree with extra parity chunks. Chunks :math:`p^{101}` through :math:`p^{128}` are parity data for chunks :math:`h^1_1 - h^1_{128}` through :math:`h^{100}_1  - h^{100}_{128}`.




Two things to note

 * This pattern repeats itself all the way down the tree. Thus hashes :math:`h^1_{101}` through :math:`h^1_{128}` point to parity data for chunks pointed to by :math:`h^1_1` through :math:`h^1_{100}`.
 * Parity chunks :math:`p^i` do not have children and so the tree structure does not have uniform depth.

The special case of the last chunks in each row
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If the size of the file is not a multiple of 4096 bytes, then it cannot be evenly divided into chunks. The current strategy is to simply fill up the last incomplete chunk to fill 4096 bytes. The next step in the encoding process is to add the parity chunks. We choose some erasure coding redundancy parameters (for example 100-of-128) and we process the file 100 chunks at a time, encoding to 128 each time.

If the number of file chunks is not divisible by 100, then we cannot proceed with the last batch in the same way as the others. We propose that we encode the remaining chunks with an erasure code that guarantees at least the same level of security as the others. Note that it is not as simple as choosing the same redundancy. For example a 50-of-100 encoding is much more secure against file loss than a 1-of-2 encoding even though the redundancy is 100% in both cases. Overcompensating, we could say that there should always be the same number of parity chunks (eg. 28) even when there are fewer (than 100) data chunks so that we alwasy end up with m-of-m+28. We repeat this procedure in every row in the tree. 

However it is not possible to use our m-of-n scheme on a single chunk (m=1) because it would amount to n copies of the same chunk. The problem of course is that any number of copies of the same chunk all have the same hash and are therefore indistinguishable in the swarm. Thus when there is only a single chunk left over at some level of the tree, we'd have to apply some transformation to it to obtain a second (but different) copy before we could generate more parity chunks.

In particular this is always the case for the root chunk. To illustrate why this is critically important, consider the following. The root hash points to the root chunk. If this root chunk is lost, then the file is not retrievable from the swarm even if all other data is present. Thus we must find an additional method of securing and accessing the information stored in the root chunk.

Whenever a single chunk is left over (m=1) we propose the following procedure.

 1. If the chunk is smaller than 4096 bytes, we use diferential padding to make n different 4096-byte chunks containing the data.
 2. If the chunk is full size we apply n different reversible permutations to get n different copies. For example we could use cyclic permutations [#]_ .

.. rubric:: Footnotes
.. [#] Alternatively, we could formally differentiate the chunks using the filesize data. In swarm, each 4096 byte chunk is actually stored together with 8 bytes of meta information - currently only storing the size of the subtree encoded by the chunk. It is plausible for a future implementation of swarm to use 1 byte of meta-information in order to differentiate multiple copies of the same data.


Benefits of CRS merkle tree
=============================

This per-level m-of-n Cauchy-Reed-Solomon erasure code introduced into the swarm chunk tree does not only ensure file availability, but also offers further benefits of increased resilience and ways to speed up retrieval.

All chunks are created equal
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
A tree encoded as suggested above has the same redundancy at every node [#]_.

.. rubric:: Footnotes
.. [#] If the filesize is not a specific multiple of 4096 bytes, then the last chunk at every level will actually have a higher redundancy even than the rest.

This means that chunks nearer to the root are no longer more important than chunks near the file. Every node has an m-of-128 redundancy level and no chunk after the root chunk is more important than any other chunk.
A problem that immediately presents itself is the following: if nodes are compensated only for serving chunks, then less popular chunks are less profitable and more likely to be deleted; therefore, if users only download the 100 data chunks and never request the parity chunks, then these are more likely to get deleted and ultimately not be available when they are finally needed.
Another approach would be to use non-systemic coding. A systemic code is one in which the data remains intact and we add extra parity data whereas in a non-systemic code we replace all data with parity data such that (in our example) all 128 pieces are really created equal. While the symmetry of this approach is appealing, this leads to forced decoding and thus to a high CPU usage even in normal operation and it also prevents us from streaming files from the swarm.
Luckily the problem is solved by the automated audit scheme which audits the integrity of all chunks and does not distinguish between data or parity chunks.

Self healing
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Any client downloading a file from the swarm can detect if a chunk has been lost. The client can reconstruct the file from the parity data (or reconstruct the parity data from the file) and resync this data into the swarm. That way, even if a large fraction of the swarm is wiped out simultaneously, this process should allow an organic healing process to occur and it is encouraged that the default client behavior should be to repair any damage detected.

Improving latecy of retrievals
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Alpha is the name the original Kademlia gives to the number of peers in a Kademlia bin that are queried simultaneously during a lookup. The original Kademlia paper sets alpha=3. This is impractical for Swarm because the peers do not report back with new addresses as they would do in pure Kademlia but instead forward all queries to their peers. Swarm is coded in this way to make use of semi-stable longer-term devp2p connections. Setting alpha to anything greater than 1 thus increases the amount of network traffic substantially – setting up an exponential cascade of forwarded lookups (but it would soon collapse back down onto the target of the lookup).
However, setting alpha=1 has its own downsides. For instance, lookups can stall if they are forwarded to a dead node and even if all nodes are live, there could be large delays before a query is complete. The practice of setting alpha=2 in swarm is designed to speed up file retrieval and clients are configured to accept chunks from the first/fastest forwarding connection to be established.
In an erasure coded setting we can in a sense have a best of both worlds. The default behavior should be to set alpha=1 i.e. to query one peer only for each chunk lookup, but crucially, to issue a lookup request not just for the data chunks but for the parity chunks as well. The client then could accept the first m of every 128 chunks queried to get some of the same benefits of faster retrieval that redundant lookups provide without a whole exponential cascade.

