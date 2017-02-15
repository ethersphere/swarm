package network

import (
	"github.com/ethereum/go-ethereum/swarm/storage"
	//"github.com/ethereum/go-ethereum/p2p/adapters"
	
)

/***
 * \todo test expenditure if struct will take more memory and/or processing than map
 */
 
 
var METATmpSwarmRegistryKeys []storage.Key
var METATmpSwarmRegistryLookup map[string][2]string
//var METATmpSwarmRegistryLookup map[adapters.NodeId]string
//var METATmpSwarmRegistryLookupReverse map[string]adapters.NodeId
/*var METATmpSwarmRegistryLookup map[adapters.NodeId]*storage.Key
var METATmpSwarmRegistryLookupReverse map[*storage.Key]adapters.NodeId*/

func init() {
	METATmpSwarmRegistryLookup = make(map[string][2]string)
	//METATmpSwarmRegistryLookup = make(map[adapters.NodeId]string)
	//METATmpSwarmRegistryLookupReverse = make(map[string]adapters.NodeId)
	/*METATmpSwarmRegistryLookup = make(map[adapters.NodeId]*storage.Key)
	METATmpSwarmRegistryLookupReverse = make(map[*storage.Key]adapters.NodeId)*/
}
