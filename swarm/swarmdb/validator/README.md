

# SWARMDB Data Flow 

All SwarmDB node types (farmers/buyers, validators) are required to register on child chain ENS in order to participate in SwarmDB services.   A SWARMDB public **swarmdb** table holds this registry:

       nodeID              string - public key of the node, PRIMARY key
       shard               string - node declared shard (must match nodeID)
       address             string - Ethereum address that earns WLK, holds WLK stake
       nodetype            string - swarmdb, aggregator, validator
       ip                  string - IP address for swarmdb nodes
       portHTTP            string - Port 
       registerTimestamp   int - unix timestamp
       targetstorage       int - SWARMDB node "vote" for the storage, measured in wei
       targetbandwidthcost int - same as above, but for bandwidth
       
SWARMDB nodes of Buyers + Farmers expose a number of HTTP interfaces to output buyerlogs, farmerlogs, and receive feedback from validators.  The core node types are:
 - Buyers are responsible for compensating Farmers for their storage and bandwidth with WLK. 
 - Farmers are responsible for storing chunks when accepting buyers' requests and interacting with other peers;  Farmers earn WLK for their storage and bandwidth. 
 - Validators are responsible for tallying/validating storage and bandwidth claims by collecting SWARMDB logs and then running multiple MapReduce jobs on these logs that have as side effects feedback to buyer and farmer SWARMDB nodes.   Validators must hold WLK in a Wolk Manager Contract, which implement Casper's Proof-Of-Stake.  Validators are rewarded for submitting correct blocks and following the protocol correctly.

## SWARMDB Logs
SWARMDB nodes, at the conclusion of each epoch, will generate farmer logs and buyer raw logs for the validators to poll, detailed below.

### Farmer Logs

Farmer nodes will generate **farmerlog** at the conclusion of each epoch with the following format:

     {"farmer":"0x68988336d54ecd93bd9098607c32497bd7df0015","chunkID":"2b3b7615443069fa6886eec0283d23b09b54a906f78df2d1e537db0ebbf148ca","chunkBD":1515542149,"chunkSD":1515542150,"rep":5,"renewable":1}
     ...     
     {"farmer":"0x68988336d54ecd93bd9098607c32497bd7df0015","chunkID":"215b892e8f44980fbbbcc3ee8f92ddad558bcfdcc0f89d6d4852aa62e1a63ebb","chunkBD":1515542149,"chunkSD":1515542157,"rep":4,"renewable":0}
    
In the future, farmerlogs are expected to live in Swarm (and should be retrievable by content hash on ENS registry). But for now, farmerlog will be _retrieved on demand_ by validators via:

     https://<ip:port>/farmerlog/<EpochUnixTimestamp>/<Shard(optional)>

where each farmer node's `<ip:port>` can be looked up from Wolk Chain ENS Registry.
 

### Buyer Logs

SWARMDB nodes will generate **buyerlog** logs  with the following format:
    

    {"buyer":"0xd80a4004350027f618107fe3240937d54e46c21b","chunkID":"2b3b7615443069fa6886eec0283d23b09b54a906f78df2d1e537db0ebbf148ca","chunkBD":1515542149,"chunkSD":1515542149,"rep":5,"renewable":1,"sig":"","smash":""}
     ...
    {"buyer":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","chunkID":"215b892e8f44980fbbbcc3ee8f92ddad558bcfdcc0f89d6d4852aa62e1a63ebb","chunkBD":1515542149,"chunkSD":1515542149,"rep":5,"renewable":1,"sig":"","smash":""}

Similarly,  multiple buyerlogs from a single buyer node can be _retrieved on demand_ via:
  
     https://<ip:port>/buyerlog/<EpochUnixTimestamp>/<Shard(optional)>

where each buyer node's `<ip:port>` can be looked up from Wolk Chain ENS Registry.

# Validator

Given the above SWARMDB HTTP interfaces, a validator (potentially, working on a specific  **Shard**) for a given epoch identified by **EpochUnixTimestamp** and a list of active SWARMDB nodes from the **swarmdb** table will poll ALL SWARMDB nodes for their logs using:

     https://<ip:port>/farmerlog/<EpochUnixTimestamp>/<Shard(optional)>
     https://<ip:port>/buyerlog/<EpochUnixTimestamp>/<Shard(optional)>

And process these logs in two stages:
![WLK Chain Data Flow](https://raw.githubusercontent.com/wolktoken/swarm.wolk.com/master/src/github.com/ethereum/go-ethereum/swarmdb/validator/Data%20Flow.jpg)



     
## Stage 1J:  Chunk Join

Validators receive this as input and run MR1, mapreduce Hadoop job to summarize what has happened to chunks in their shard across both farmerlogs and buyerlogs:

           $ cat buyerlog-input*.txt farmerlog-input*.txt |  smash-map | sort |  smash-reduce
           {"chunkID":"1d4b6e4aa86d48c464c9adf83940d4e00df8affc","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"rep":5,"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"sig":"a1"}
           {"chunkID":"4b09668b93c718092a408c4222867968fcd3ad98","buyers":["0xd80a4004350027f618107fe3240937d54e46c21b"],"rep":5,"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"sig":"a1"}
           {"chunkID":"aeec6f5aca72f3a005af1b3420ab8c8c7009bac8","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","0xd80a4004350027f618107fe3240937d54e46c21b"],"rep":5,"renew":0,"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","0xfd990c3c42446f6705dd66376bf5820cf2c09527"],"sig":"a1"}
           {"chunkID":"d368b1c09e7ddfb6aff24e8e6f181ffeea905d31","buyers":["0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8"],"rep":5,"renew":0,"farmers":["0xf6b55acbbc49f4524aa48d19281a9a77c54de10f"],"sig":"a1"}

The aggregator actually verifies via Tron's SMASH proof the correct storage of the chunk using the buyerlog provided smash proof by reaching out to the farmer node

     https://<ip:port>/<smash>/<CHUNKID>



### Validator Response to Invalid SMASH proof

If the validator receives invalid SMASH proofs from a farmer, the validator will advise the farmer:

     https://ip:port/rejectedsmash/<CHUNKID>/<SUBMITTED_SMASH_PROOF>/<Hash:Validator_sig>

If this is due to innocent reasons (corrupt disk, network issues, etc.), the farmer could simply delete the chunk.

### Farmer Response to Validator Claim of Invalid SMASH proof

If the farmer is faced with a validator maliciously rejecting a sound SMASH proof, the farmer can submit its counter-proof to a public swarmdb table holding active litigation in *recent epochs: 

litigation table has following columns:
	
    chunkID		string - hash
	id		string - nodeID of the SWARMDB node
    validatorID	string - public key of the validator node
    validatorSig	string - Validator's signiture for signing a rejection
    farmerSig	string - farmer's signiture for signing such litigation
    Crash		string - Smash-proof in dispute 
    epochtimestamp	  int  - epoch of litigation
    shard		string - the shard to which the validator belongs to
       
During Casper block proposal, validators will also vote on settling active litigations from recent epochs. If a farmer's litigations are found valid, such farmer shall be compensated and proper adjustments shall be made in proposed epoch. Honest validators will be rewarded with finder fees while malicious validator's deposit will be slashed. 

### Other Aspects of SMASH/SWINDLE System

* [TODO: Michael] How are the SMASH/SWINDLE method covered?

## Stage 1S: Storage Log Summary

The output of Stage 1J joins based on chunkID, but does not select the precise farmers to be awarded. 

     $ cat storage-input*.txt | storage-map | sort | storage-reduce
     {"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","s":123}
     {"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","s":-123456}
 
 The `rep` parameter is used with a RANDAO type method to select the farmers pseudo-deterministically, based on the chunkID and farmerID. 
 * [TODO: Sourabh] How exactly does this work?

If there are too few farmers for any chunk (based on `minrep`), the validator should rerequest the chunk and prove that it has done so by providing a `chunksig`, signing the chunks content with its public key.

### Validator Response to Missing buyer request

If the validator receives claims from farmers that do not have a buyer associated with it, the farmer claim is not included in the validator's output, however, the validator will let the farmer know:

     https://<ip:port>/failedchunkclaim/<CHUNKID>/<Hash:Validator_sig>

The farmer may then mark the chunk for deletion, or delete all the buyers that dominate the failed chunk claims.

### Validator Response to Over-Replication

If the validator notices that a chunk is being overreplicated (by a factor of 2 over rep), the validator will ping the following end point of the farmer: 

     https://<ip:port>/overreplication/<CHUNKID>/<Rep>/<NumOfFarmers>/<Hash:Validator_sig>

The farmer can decide to remove the chunk based on the supplied number of farmers.

### Validator Response to Under-Replication

If the validator receives buyerlog that do not have sufficient replication (of at least 4 farmers), the validator will ping the following end point: 

     https://<ip:port>/underreplication/<CHUNKID>/<NumOfFarmers>/<Hash:Validator_sig>

This ping could be sent to the buyers's designated guardian or insurer instead in a later version.

### Validator Signatures

Validators must sign all feedback requests sent into SWARMDB nodes.   SWARMDB nodes are expected to check the **swarmdb** table to validate validator signatures.

## Stage 1B: Bandwidth Log Summary

The essence of the SWAP protocol is done *without* a checkbook using a record of peer-to-peer bandwidth logs retrieved on demand by the validator from the following http interface of SWARMDB:

     https://<ip:port>/bandwidthlog
     $ cat bandwidthlog-input1.txt
     {"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","remote":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","s":123456,"receipt":"r1"}

     $ cat bandwidthlog-input2.txt
     {"id":"0xdf990c3c42446f6705dd66376bf5820cf2c09527","remote":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","s":-123456,"receipt":"r2"}

In the above case, `f6..` has sent 123456 chunks to `df..`, and both `f6` and `df` have a record of the score being settled at the `PayAt` level. 

When a claim is matchable (the validator can find matching claims from both parties), the aggregate amount must be transferred.

     $ cat bandwidthlog-input*.txt | bandwidth-map | sort | bandwidth-reduce
     {"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","b":123456}
     {"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","b":-123456}

Take 3 nodes A,B,C where where A sends a  receives a chunk from A and C receives a chunk from B and each shuts down the channel:

     (Farmer) A -1-> B -1-> C (Buyer)

the bandwidthlogs of each would look like:

     $ cat bandwidthlog-inputA.txt
     {"id":"A","remote":"B","s":1,"receipt":"r1"}   

     $ cat bandwidthlog-inputB.txt
     {"id":"B","remote":"A","s":-1,"receipt":"r2"}
     {"id":"B","remote":"C","s":1,"receipt":"r3"}

     $ cat bandwidthlog-inputC.txt
     {"id":"C","remote":"B","s":-1,"receipt":"r4"}

the net for B would be **0**:

     $ cat bandwidthlog-input?.txt | validator-1b-map | sort | validator-1b-reduce
     {"id":"A","b":1}
     {"id":"C","b":-1}


### Validator Response to Mismatched Claims 

If there is an unmatched claim detected by the aggregator, the aggregator should report it back to the farmer at the following report:

     https://<ip:port>/failedbandwidthlog/<Peer>/<Receipt>/<Hash:Sig>

The SWARMDB node may mark this peer as being untrustworthy and drop them for trading if the total amount of bandwidth exceeds an unacceptable threshold.

## Stage 2:  Proposing Blocks - Storage and Bandwidth Costs

After completing Stage 1J followed by 1S (resulting in "s") and Stage 1B (resulting in "b"), the outputs of both storage (s) and bandwidth (b) can be tallied in SWARMDB using the previous epochs `storagecost` and `bandwidthcost`, running a MapReduce operation to derive intermediate outputs of `s`, `b`, and `t`:

    $ cat *-input*.txt | collation-map | sort | collation-reduce 
    {"validatorID":"validator_publickey1","epochtimestamp":12431234,"id":"0xcb2fa2c491451cac943bb5a0261eb101cc36a4f8","s":3,"b":1,"t":400,"sig":"s1"}
    {"validatorID":"validator_publickey1","epochtimestamp":12431234,"id":"0xd80a4004350027f618107fe3240937d54e46c21b","s":1,"b":-1,"t":0,"sig":"s2"}
    {"validatorID":"validator_publickey1","epochtimestamp":12431234,"id":"0xf6b55acbbc49f4524aa48d19281a9a77c54de10f","s":-2,"b":99,"t":9700,"sig":"s3"}
    {"validatorID":"validator_publickey1","epochtimestamp":12431234,"id":"0xfd990c3c42446f6705dd66376bf5820cf2c09527","s":-2,"b":-99,"t":-101000,"sig":"s4"}

following a schema of:

    validatorID       string - public key of the validator node
    epochtimestamp    int - epoch of coverage
    shard             string - the shard to which the validator belongs to
    id                string - nodeID of the SWARMDB node 
    s                 int - number of replicates to be paid by buyer 
    b                 int - number of replicates to be paid by buyer
    t                 int - s*storagecost + b*bandwidthcost, in wei 

This output is stored in SWARM submitted with a simple SWARM hash to the Wolk Manager Contract and represents a validators proposal for the epoch.   The Wolk Manager Contract implements Casper's Proof-of-stake with deposits / withdrawals / votes / slashing and results in validators also earning tokens for bets that payoff.

# Sharding 

* [TODO: Sourabh] How does a validator choose to change its sharding level *exactly*?

# SWARMDB Voting System 

*  [TODO: Sourabh] How are the local SWARMDB node inputs of `targetstoragecost` and `targetbandwidthcost` { read by validators, processed by validators, submitted in block proposals} *exactly*?


