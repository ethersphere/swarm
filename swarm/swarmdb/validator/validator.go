// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

/*
 // Validator map reduce operations
 // GENERATE test data/{farmerlog,buyerlog,swaplog}.txt
 ./validator -mapred=generatelog

 // SMASH
 cat data/farmerlog.txt data/buyerlog.txt | ./validator -mapred=smashmap | sort | ./validator  -mapred=smashred > data/storage-input.txt

 // STORAGE
 cat data/storage-input.txt | ./validator -mapred=storagemap | sort | ./validator -mapred=storagered > data/collation-storage.txt

 // SWAP
 cat data/swaplog.txt | ./validator -mapred=swapmap | sort | ./validator  -mapred=swapred > data/bandwidth-input.txt

 // BANDWIDTH
 cat data/bandwidth-input.txt | ./validator -mapred=bandwidthmap | sort | ./validator -mapred=bandwidthred > data/collation-bandwidth.txt

 // COLLATION
 cat data/collation-*.txt | ./validator -mapred=collationmap | sort | ./validator -mapred=collationred > data/submittx-input.txt
*/
package main

import (
	"flag"
	"fmt"
	"swarmdb"
)

var mapred string

func init() {
	flag.StringVar(&mapred, "mapred", "swapmap", "map-reduce task (smashmap, smashred, swapmap, swapred, storagemap, storagered, bandwidthmap, bandwidthred, collationmap, collationred)")
	flag.Parse()
}

func main() {
	var v swarmdb.Validator
	v.AddNode("farmer1", "0.0.0.0", 2001)
	v.BandwidthCost = 271828
	v.StorageCost = 314159

	switch mapred {
	case "generatelog":
		v.GenerateLogs()
	case "smashmap":
		v.SmashMap()
	case "smashred":
		v.SmashRed()
	case "swapmap":
		v.SwapMap()
	case "swapred":
		v.SwapRed()
	case "storagemap":
		v.StorageMap()
	case "storagered":
		v.StorageRed()
	case "bandwidthmap":
		v.BandwidthMap()
	case "bandwidthred":
		v.BandwidthRed()
	case "collationmap":
		v.CollationMap()
	case "collationred":
		v.CollationRed()
	default:
		fmt.Printf("unknown mapred: %s\n", mapred)
	}
}
