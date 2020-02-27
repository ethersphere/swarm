/*
Traditionally Swarm chuns are content addressable. i.e.

Chunk Address = sha3(Chunk Payload)
Lately there is a requirement of chunk structure that adhere to certain modifications. These modification enable
certain features in Swarm such as Feeds. These features allow only the Owner of the feature to construct the chunk.
The chunk structure of feeds follow the following structureTraditionally Swarm chunks are content addressable. i.e.

	Chunk Address = sha3(Chunk Payload)

Lately there is a requirement for a chunk structure that adheres to a certain format such that enable the owner of the
chunks to create certain features in Swarm such as Feeds and Global pinning. These features allow only the Owner of the
data to construct the chunk. Hence these chunks are called Single Owner Chunks (SOC). The format of SOC is as
follows

     Chunk Address = Sha3( pubkey(owner) +   ID   )
                                 (32)    +  (32)       = 64 bytes


    Chunk Data = ID   + Signature + padding + span  + payload
                (32)  +   (65)    +  (23)   + (8)   +  (4096)

          Where, Signature = sign(pubkey(owner) + sha3 (ID + BMTHash(payload)))
                 span = 8 bytes of real chunk
                 padding = 23 bytes

*/

package soc
