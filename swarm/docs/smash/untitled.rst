**********************************************
SMASH: secured by masked audit secret hash
**********************************************


Motivation
===============================


Caveats and requirements
--------------------------------------

Since we want to use the proof of custody for recurring audits, it is importatnt we optimise on
* blockchain use
* storage overhead
* network overhead


Proof of custody by pregenerated audits
=========================================

Audits based on pregenerated seeds have the generic pattern

* challengee is presented with a seed (a nonce) previously unseen and not possible to guess
* the seed and the stored data is used in an algorithm to provide proof that the challengee has custody of the data
* the proof is in a form of a secret the correctness of which can be checked by a third party without having the file

Calculating audit secret from audit seed
------------------------------------------

Applying a seed to a chunk to result in a secret could simply be done by hashing the entire file with the seed
appended. Importantly, however, if the secret in the challenge is incorrect, storer needs to prove it to the world that they are innocent. The only concern here is how to prove that the auditor is cheating in that the secret calculated from the seed they give does not match the secret. If we used hashing the entire file then giving a proof would only be verified by parties knowing the data, which means that for third parties the data should be part of the proof, which obviously does not scale.
The relevant insight here is that we pick a merkle proof of the file based on the seed and an index, and manipulate only that to result in the audit secret. By doing this we allow proofs with length logarithmic in the file size.

So given a seed :math:`s`, and an index :math:`j` and a chunk :math:`c`, we construct the secret the following way.

1 if :math:`c` is of length less than a predefined maximum chunk length (:math:`l_{max}=2^m`) then keep concatenating to it successively a version of itself salted by the seed.

.. math::
     s^0_0 = s
     s^i_0 = Sha3(c'_{i-1}|s) for i > 0
     s^i_j = Sha3(s^i_{j-1}) for j > 0
     s'_i =  s^i_0|s^i_1|...|s^i_k[0:len(c)] where k-1 * len(s) < len(c) <= k*len(s)
     c'_0 = c
     c'_i = c ^ s'_i
     c' = c'_0|c'_1|...|c'_l[0:2^m] where (l-1)*len(c) < 2^m <= l*len(c)

.. Knowing the repeatability order of the audit masks, we can identify the maximum reasonable piece size as :math:`w_{\mathrm max}=2^{p-r}`.

..   c = c_0 | c_1 | .... | c_{2^r-1}, where
..   c_i = c[w*i:w*(i+1)-1]

2. chop the chunk into hashsized segments. Assuming for convenience that hashsize is not larger than the maximum length and hash is a power of two: :math:`\overline{H} = 2^h` and :math: `h < m`, then

..  math:
   c = c_0|...|c_{n-1} \mathrm{where}
   n = 2^{m-h}

c is a concatenation of :math:`n` segments.
4. Now calculate a salted version of the data. Take the the :math:`j`th segment of the chunk and replace it with a new string that is the salted version of this segment :math:`S(s, c) = c_0|...|c_{j-1}|Sha3(c_j|s)|c_{j+1}|...|c_{n-1}`
5. Then build up the binary merkle tree over the segments of :math:`S(s, c)` using the hash function. Since :math:`n` is a power of 2, the resulting binary tree is regular and balanced and has a depth of :math:`d=m-h`. See figure \ref{fig:mmtree}
6. calculate the merkle root of this regular binary merkle tree for segment `j` to result in the audit secret.

.. math:
    AS(s, c)=MR^{2,n}(S(s, c))

This is formalised as follows.

Note that since the other segments did not change, if one knows the merkle proof the :math:`j`th segment of the original chunk then given the seed, the modified merkle proof can simply be recalculated
in exactly `d`. This essentially means that proof of correctness of the secret is available in logarithic steps.

The strength of this proof is questionable since one iteration does not prove possession of the file.
We can define a parameter :math:`F=2^f` for the force of the proof.

The secret for :math:`AS^f(s, c)` is defined as follows

1. Generate indices :math:`j_0, j_1, ..., j_{F-1}` according to a deterministic pattern depending only on :math:`SH(c)` and `s`.
For simplicity we assume that

..  math:
   s_0 = s
   s_i = Sha3(s_{i-1}|j_{i-1})
   j_i = (s_i % n/F)+i*n/F

2. Define modified segments as follows

.. math:
   S'(s, c, j) = Sha3(c_0|... c_j|)s_j \mathrm{ if } j\in\{j_0, j_1, ..., j_{F-1}}
      c_j  \mathrm{otherwise}
    }
   }

2. Calculate the audit secret

.. math:
    AS(s, c)=MR^{2,n}(S^f(c, s, j))


Audit
--------------------------
To generate a challenge one needs to know `r` but only that.

1. generate a random number `0<=i<2^r-i`,
2. calculate the seed for `i`. Applying index update if required.
3. reveal the seed (and an index if needed) to the storer

As a response to the challenge, storer will calculate the secret for the key according to the procedure described. Now the storer can check the secret against the known mask and prove they possess the data.

Asnwer to the challenge and (s)mash proofs
--------------------------------------------

Once the challengee know the seed of the challenge, they can set out to find the corresponding secret by following the above procedure. The secret is the answer to the challenge.

The challenge can be thought of as a claim by the auditor that a known mask is the mask corresponding to the secret generated from the chunk and the seed. Anyone that knows and trusts the mask can then verify that a secret is correct by hashing it and comparing it to the said mask.

Note that if the secret had not ever been revealed, the storers must have calculated it, since it cannot be guessed. As the calculation relies on :math:`F` random segments of the chunk, storers are incentivised to keep it in full. It is fair to assume that the swarm had the file at some point. Since the seed had never been revealed or guessable, and also used in the calculation, this also means that storers have the chunk after the seed was revealed.

If the storer kept the data they will always know that their answer is correct.
If a storer chooses not to preserve the data, it is impossible to be 100% sure that they give the right answer to a challange. Given these properties it is valid to say that responding with *an answer to a challenge with a seed can be considered a valid positive proof of current custody*.

In order to have the secret available to check against at the time of audit, they have to be generated when the data is known. If owner want their chunks to be audited without them having the data, the secrets need to be pregenerated.

The auditors do not need to remember the secret only be able to check that it is correct. This is achieved by masking the secret by hashing it and make the mask public. Since unhashing is cryptogaphically impossible, verifying that the secret hashes to the mask is equivalent to checking the secret. If audits are to be repeated, several secrets and their corresponding masks should be pregenerated. Note that in order for any third party to verify that the secret provided is correct, does not require them to remember the particular mask. The pregenerated masks can be organised in a merkle tree then the correctness of a mask can be proven by a merkle proof of the mask assuming the root hash of the merkle tree is known and trusted.

Masked audit secret hash tree
-----------------------------------------

Assume that we have :math:`n=2^r` audit seeds specific to a chunk. Each audit seed allows nodes to launch an independent challenge to the swarm and check that the associated data is preserved.
:math:`r` is the repeatability order of the audit.
Using the audit seeds and the chunk one can construct a *masked audit secret hash tree* (MASH tree) as follows:

1. take the :math:`n` audit seeds and calculate the audit secrets.
2. given the :math:`n` audit secrets, construct :math:`n` masked audit secrets by taking their  hash
3. contruct the MASH tree, the regular binary merkle tree out of the :math:`n` masked audit secrets using a hash function
4. take the root hash :math:`MASH(c)` of the masked audit secret tree and sign it. This has to be part of a challenge and means the owner claims that following this specification, given the seeds :math:`s_0, s_1, ... s_{n-1}` and the chunk will result in the secret :math:`AS^F(c, s_i, j), masked by MASH(c, s_i, j)

.. math:
    S^f(c, s, j) = c_0|...|c_{j-1}|Sha3(c_{j_0}|c_{j_1}|...|s)|c_{j+1}|...|c_{n-1}


5. serialise the masked secrets in the order of indexes, to be included in the store request.


Deriving indexes from the seed
------------------------------------------

Now this however implies that we remember all the masks
The only possible scenario under this simple version is if you responded to a seed once, store the response only while discarding the data. In this case if an auditor challenges the same chunk with the same index, the storer can respond correctly even though they no longer have the file. However, if the indexes are not recycled, storers can be absolutely sure they can get rid of parts of a chunk. Therefore we simply suggest that indexing of the segments of the chunk are derived from a fix slices of bits of the seed (essentially random bits, so indexes will be recycled during successive audits with the same MASH.

Now given the index :math:`j` and seed :math:`s` and :math:`w=2^n` where `0<=n<=p-r` . In fact the index can be deduced from the seed, according to the following conditions,

 :math:`i = s % 2^r`. If `n=r` then `i=j` , otherwise :math:`j=s-i/2^r % 2^n`.  In other words, the last :math:`r` bits map to :math:`i`, and the preceeding :math:`n` bits map to `j`.
With `j` is established, one calculates the seed modulated bits of the file :math:`M(c_j) = Sha3(c_j|s)` using all the indexes.

Challenges and (s)mash proofs
-------------------------------

However, storer can also give a proof of the correctness if they know all the masks in the MASH securing the chunk. Note that the root hash of the MASH tree is signed by the owner and the auditor.

Both the positive and negative response to the challenge contains this secret and a proof.
If the hash of the secret matches the mask in the :math:`i`th position, the refutation consists of the
the MASH proof of the `i`th mask. This is the positive response reassuring the integrity of storage of the chunk. Hence the motto: SMASH = secured by masked audit secret hash proof. We can say the chunk is smash-proof.

If the hash of the revealed secret does not match the mask at the relevant index, then the refutation is
the merkle proof(s) of the relevant segment(s) of the original chunk. This response is called a smash proof, and we can say the challenge has been smashed by the storer.

Given the usual 256bit Keccak SHA3, :math:`\overline[h}=32` used in swarm, mash proof itself is exactly :math:`32r` bytes long. For instance if :math:`r=3`, the proof is a mere 128 bytes.
In order to show how this suffices, let us go through the steps how the proof is validated.

Validating mash proofs
-----------------------------

We assume that the
The length of the mash proof :math:`\overline{\mathrm{smp}}`.

1. :math:`\overline{\mathrm{smp}} % 32 != 0`, reject the proof.
2. take the secret as :math:`s=\mathrm{smp}[0:32]`
3. calculate :math:`r=\over{\overline{\mathrm{smp}},32}`
4. take the seed (known) and calculate the index :math:`i=s % 2^r`
5. take the binary representation if :math:`i`. It is easy to see that the bits give the direction to which the merkleproof on each level. The directional hash function :math:`DH_l(x,y)` is defined as follows:

..  :math:
    DH_l(x,y)=\leftbrace{
      \vbox[align=r]{
        Sha3(x|y) \mathrm{if } \mathrm{bin}(i) \^ \mathrm{bin}(2^{r-l-1}) == bin(0)\\
        Sha3(y|x) \mathrm{otherwise}
        }
    }

6. The storers secret can now be calculated using the following inductive definition

..  math::
  H_0 = DH_0(s,SH(c))  % the mask
  H_i = DH_i(H_{i-1})
  MASH'(c) = H_{r-1}

Note that if the challenge is a peer to peer challenge

7. Now if the MASH :math:`MASH'(c) == MASH(c)` the smash proof is valid and one can conclude with certainty that the file is stored in the swarm.

In the latter case the smash proof is a little longer since it involves giving merkle proofs of segments of the original chunk. Given a seed :math:`s` and the strength of the proof scheme :math:`f`, storer calculated the secret and found that it does not match the audit mask. In this case the Merkle proofs prove the existence and position of the respective segments in the original chunk. This proof is supposed to be very rarely used, since it assumes that auditors are sending frivolous false seeds or publish incorrect masks, which they are decincentives to do.

Repeatability and file-level audits
====================================

The problem of scaling audit repeatability with fixed chunks
--------------------------------------------------------------

The choice of :math:`r` has an impact on the length of merkle proofs which is needed for one type of refutation of the challange as we see shortly. More importantly, though, since someone needs to remember the masks, this scheme has a fix absolute storage overhead which is independent of the size of the pieces we prove the storage of. Since it is not realistic to require more than 5-10% administrative overhead even for very long term
storage period, larger :math:`r` values only scale if the same seeds can guard the integrity of larger data.

In particular, take the example of a standard swarm chunk, that  is 4096 bytes, :math:`m=12`.
Assuming standard keccak 256bit sha3 we have :math:`h=5, d=7`.
This allows for merkle proofs with length of :math:`2*(d-1)2^+2^f`.   for 128 independent audits at a 100% storage overhead. Instead for a chunk :math:`r=0,1,2,3,4` seem realistic choices for :math:`r=0.8,1.6,3.2,6.3,12.5%` storage overhead.

Ultimately repeatability order should reflect the TTL (storage period) of the request, therefore repeatability and fix chunk size cannot scale unless we compensate the overhead by reusing them over several chunks.
This problem does not occur with Storj since the shards can be sufficiently big, however with swarm the base unit of contracting is the chunk.
The insight here is that we can reuse the same seed over several chunks if and only if we query the integrity of those chunks at the same time.

Discussing the sw^3 approach to chunk insurance, we mentioned among the problems that users will probably want to check the integrity of their assets on semantic units like document or document collection. Solution should be in place to make sure litigation and auditing is easily managed for these units.

Incidentally, smash auditing solves both problems at one go. This is the topic of the this section.


Document-level audit
--------------------------------

Given a seed, we define the document-level secret as follows:

1. take the chunk tree of a document as defined by the geswarm hash chunker. See figure \ref{fig:swarmhash}.
2. define a structurally parallel chunk tree but when calculating the :math:`i`th segment of a non-leaf chunk, the smash secret is calculated on the chunk.
3. the hash at the root of the chunk-tree is the swarm audit secret for the file.


..  image:: fig/bzzhash.pdf
   :height: 300px
   :width: 300 px
   :scale: 50 %
   :alt: swarm-hash
   :align: centre


In practice given a file the owner wants to store, the secrets can be efficiently generated at the time the file is chunked. As the chunks are uploaded, and guardian addresses and their receipts are stored in a structure parallel to the chunktree.

Without loss of generality let us assume that `r=128`, so the masks fit into one chunk. for a 20-chunk file
(80KB), this will allow 128 independent audits for extra 5% storage overhead.
This pattern can be extended to document collections covering entire sites and therefore scale very well.
For a TTL requiring repeatability order :math:`r` (for :math:`2^r` independent audits without ever seeing the files again), and given a :math:`o` as the maximum storage overhead ratio. the minimum data size is :math:`2^{r-7}*o*2^12 = o2^{r+5}`.

This audit will not reveal the secret to the individual storers of chunks, therefore it can never be used to prove to third parties that a challenge is invalid. For the same reason it is not used for public litigation. However combined with smash proof for inividual chunks it can be used to solve the repeatability issue.

Here is the multi-stage swindle process.

The auditor may periodically audit the document by sending off audit requests of the simple type which are similar to retrieval requests instead of joining the chunks, recalculate the secret. If everybody responds, and the secret matches

The pregenerated
Now this however implies that we remember all the masks
The only possible scenario under this simple version is if you responded to a seed once, store the response only while discarding the data. In this case if an auditor challenges the same chunk with the s:me index, the storer can respond correctly even though they no longer have the file. However, if the indexes are not recycled, storers can be absolutely sure they can get rid of parts of a chunk. Therefore we simply suggest that indexing of the segments of the chunk are derived from a fix slices of bits of the seed (essentially random bits, so indexes will be recycled during successive audits with the same MASH.



Store requests and storage receipts
--------------------------------------
Let us recap how the network communications change given the possibilities of
How this is generated and made available is gonna be discussed below.
:math:`n = 2^r`
the storage audit metadata is a tuple containing the TTL, the signed audit root (SAR), and the root swarm hash of the chunk
the chunk itselfresulting 4096 byte-long data blob
The audit secret for index :math:`i` is constructed as follows:
:math:`i`t
* The audit challange index is a random integer between :math:`0` and :math:`n-1`.
* the owner calculates the base audit seed for the chunk,
* from the chunk's base audit seed, :math:`n` audit seeds are generated
* from the audit seeds and the chunk itself, owner generates :math:`n` audit secrets for the chunk
 a receipt* from the audit seeds and the chunk itself, owner generates :math:`n` audit secrets for the chunk


