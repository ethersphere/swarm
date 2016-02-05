**********************************
Parameter settings
**********************************


SWEAR & SWINDLE scheme parameters
======================================

The following parameters of the scheme are fixed constants and not configurable by the client.

..  option:: epoch-length

  the base unit defining the granularity of storage period; the shortest valid period for storage guarantees. Expressed as integer representing number of blocks.

..  option:: maximum-simultaneous-losses

   number of chunks that a peer can be simultanously proven guilty of losing and have associated stake forfeited. Unless the missing deposit is replenished, one more chunk lost leads to suspension


..  option:: maximum-total-losses

  number of chunks a registered node is allowed to lose during its entire registration.
  This may not be needed or could be tracked on chain as basis for reputation.

..  option:: maximum-total-stake-for-single-chunk

  the maximum deposit all storers stake in total for a single chunk. This translates to a minimum proximate bin size.

..  option:: security-constant

   price/deposit ratio. serves to quantify what percentage of the deposit dedicated to one chunk a node is allowed to ask per epoch.

..  option:: reserve-deposit

  amount of collateral on top of the deposit dedicated to cover lost chunks.
  This needs to be chosen so that the total deposit a single node risks is sufficient deterrent against malicious action.

..  option:: deposit-restore-period

  The maximum amount of time a node is given to restore their deposit after partial forfeiture before they are suspended and lose their deposit in full. Integer for number of blocks.

..  option:: challenge-open-period

  The amount of time an accused node is given to refute a challenge submitted against them. After this period ends, a node with no successful refutation will be regarded as guilty and their punishment is enforced.

..  option:: chunk-upload-compensation

  The percentage of forfeited deposit that goes to the node(s) that initiated the challenge. This only applies to challenges that start an iterative litigation (chain of challenges),  intermediate nodes need no extra incentive beyond self-defence


..  option:: owner-compensation-percentage

  The percentage of forfeited deposit that goes to the victim (i.e., dedicated to compensate owners as recorded in the request)


SWAP options
=====================

Pricing options
----------------------

..  option:: BuyAt (@math{2*10^{10}} wei)

   highest accepted price per chunk in wei


..  option::  SellAt (@math{2*10^{10}} wei)

   offered price per chunk in wei


..  option::  PayAt (100 chunks)

   Maximum number of chunks served without receiving a cheque. Debt tolerance.


..  option::  DropAt (10000)

   Maximum number of chunks served without receiving a cheque. Debt tolerance.

Debt tolerance settings
------------------------

..  option::  AutoCashInterval (@math{3*10^{11}}, 5 minutes)

   Maximum Time before any outstanding cheques are cashed


..  option::  AutoCashThreshold (@math{5*10^{13}})

   Maximum total amount of uncashed cheques in Wei


..  option::  AutoDepositInterval (@math{3*10^{11}}, 5 minutes)

   Maximum time before cheque book is replenished if necessary by sending funds from the baseaccount


..  option::  AutoDepositThreshold (@math{5*10^{13}})

  Minimum balance in Wei required before replenishing the cheque book


..  option::  AutoDepositBuffer (@math{10^{14}})

  Maximum amount of Wei expected as a safety credit buffer on the cheque book


..  option::  PublicKey (PublicKey(bzzaccount))

  Public key of your swarm base account use


..  option::  Contract ()

  Address of the cheque book contract deployed on the Ethereum blockchain. If blank, a new chequebook contract will be deployed.


..  option::  Beneficiary (Address(PublicKey))

  Ethereum account address serving as beneficiary of incoming cheques


Data structures
=======================

[WIP]

Challenge attributes
-----------------------

..  option:: proof-of-custody-seed

  indicates that no chunk upload is necessary and the challenge is purely probing syncing behaviour compliant with the rules

..  option:: receipt

  the signed receipt of a chunk with information about the original owner, the accused (signer), sync token, session index, blockheight at the time of receiving.



Glossary
===============

[WIP]

SWEAR
  Storage With Enforced Archiving Rules
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

probing
  special form of litigation where possession of a chunk is proved by proof of custody. The litigation does not stop but forces node to iteratively prove they synced according to the rules

SWINDLE
  swarm insurance driven litigation engine
  the module in the client code that drives the iterative litigation procedure, initiates litigation in case loss of a chunk is detected and respond with refutation if the node itself is challenged

