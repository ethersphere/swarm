
Glossary
======================

.. glossary::

  owner
      node that produces/originates content by sending a store request

  storer
      node that accepted a store request and stores the content

  guardian
      the first node to accept a store request of a chunk/

  custodian
      node that has no online peer that is closer to a chunk address

  auditor
      node that initiates an audit by sending an audit request

  insurer
      node that is commissioned by an owner to launch audit requests on their behalf

  swear
      the component of the system that handles membership registration, terms of membership and security deposit

  swindle
      the component of the system that handles audits and escalates to litigation

  swap
      information exchange between connected peers relating to contracting, request forwarding, accounting and payment

  SWAP
      Swarm Accounting Protocol, Secured With Automated PayIns. And the name of the suite of smart contracts on the blockchain handling delayed payments, payment channels, escrowed obligations, manage debt etc.

  SWEAR
    Storage With Enforced Archiving Rules or Swarm Enforcement And Registration
    the smart contract on the ethereum blockchain which coordinates registration, handles deposits and verifies challenges and their refutations

  sworn node, registered node, swarm member
    a node which registered via the SWEAR contract and is able to issue storage receipts until the expire of its membership

  suspension
    punative measure that terminates a node's registered status and
    burns all available deposit locked in the SWAR contract after
    paying out all compensation

  registration
    nodes can register their public key in the SWEAR contract
    by sending a transaction with deposit and parameters to the SWEAR contract
    they will have an entry

  audit
    special form of litigation where possession of a chunk is proved by proof of custody. The litigation does not stop but forces node to iteratively prove they synced according to the rules.

  SWINDLE
    swarm insurance driven litigation engine
    the module in the client code that drives the iterative litigation procedure, initiates litigation in case loss of a chunk is detected and respond with refutation if the node itself is challenged.

  proof of custody
    A cryptographic construct that can prove the possession of data without revealing it. Various schemes offer different properties in terms of compactness,  repeatability, outsourceability.

  audit
    An integrity audit is a protocol to request and provide proof of custody of a chunk, document or collection.

  Erasure codes
    An error-correcting scheme to redundantly recode and distribute data so that it is recovered in full integrity even if parts of it are not available.

  bzz protocol
    The devp2p network communication protocol Swarm uses to exchange information with connected peers.

  chequebook contract
    A smart contract on the blockchain that handles future obligations, by issuing signed
    cheques redeemable to the beneficiary.

  syncing
    The protocol that makes sure chunks are distributed properly finding their custodians as the node closest to the chunk's address.

  deposit
    The amount of ether that registered nodes need to lock up to serve as collateral in case they are proven to break the rules (lose a litigation).

  registration
    Swarm nodes need to register their public key and collateralise their service if they are to issue receipts for storage insurance.

  litigation
    A process of challenge response mediated by the blockchain which is initiated if a node is found suspect not to keep to their obligation (to store a chunk). The idea is that both challenge and its refutation is validated by a smart contract which can execute the terms agreed in the breached contract or any condition of service delivery.

  chunk
    A fixed-sized datablob, the underlying unit of storage in swarm. Documents input to the API are split into chunks and recoded as a Merkle tree, each node corresponding to a chunk.

  content addressing
    A scheme whereby certain information about a content is index by the content itself (with the hash of the content).

  receipt
    signed acknowledgement of receiving (or already having) a chunk.

  SMASH proof
    Secured with Masked Audit Secret Hash: a family of proof of custody schemes

  CRASH proof
    Collective Recursive Audit Secret Hash: a proof of custody scheme for collections
