// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
package localstore

import (
	"fmt"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/shed"
	"strconv"
)

const (
	PinVersion = "1.0"
	DONT_PIN   =  0
	MAX_PINS   =  256
)


type PinApi struct{
	db    *DB
}

func NewPinApi(lstore *DB) *PinApi {
	return &PinApi {
		db: lstore,
	}
}


func (p *PinApi) ShowDatabase() string {

	schemaName, err := p.db.schemaName.Get()
	if err != nil {
		schemaName = " - "
	}
	log.Info("Database schema name", "schemaName", schemaName)


	gcSize, err := p.db.gcSize.Get()
	gc_size := ""
	if err != nil {
		gc_size = " - "
	} else {
		gc_size = strconv.FormatUint(gcSize, 10)
	}
	log.Info("Database GC size", "gc_size", gc_size)


	for i := 0; i < 256; i++ {
		val, err := p.db.binIDs.Get(uint64(i))
		if err == nil && val != 0 {
			log.Info("Database binIds", strconv.Itoa(i),strconv.Itoa(int(val)) )
		}
	}

	// Get schema
	 s, err := p.db.shed.GetSchema()
	 if err == nil {
	 	for k, v :=  range s.Fields {
			log.Info("Schema field ", k, v.Type)
		}
	 	for k, v :=  range s.Indexes {
			log.Info("Schema Index", strconv.Itoa(int(k)), v.Name)
	 	}
	 }

	 // Print the retrievalDataIndex
	_ = p.db.retrievalDataIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("retrievalDataIndex",
			fmt.Sprintf("2|pinc=%d|%0x064x", item.PinCounter, item.Address),
			fmt.Sprintf("storeTS=%d|binId=%d|datalen=%d", item.StoreTimestamp, item.BinID, len(item.Data)))

		return false, nil
	}, nil)

	_ = p.db.retrievalAccessIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("retrievalAccessIndex",
			fmt.Sprintf("3|%0x064x", item.Address),
			fmt.Sprintf("accessTS=", item.AccessTimestamp))

		return false, nil
	}, nil)

	_ = p.db.pullIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("pullIndex",
			fmt.Sprintf("4|po=%d|binID=%d", p.db.po(item.Address), item.BinID),
			fmt.Sprintf("%0x064x=", item.Address))

		return false, nil
	}, nil)

	_ = p.db.pushIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("pushIndex",
			fmt.Sprintf("5|storeTS=%d|%0x064x", item.StoreTimestamp, item.Address),
			fmt.Sprintf("tags=", ""))

		return false, nil
	}, nil)

	_ = p.db.gcIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("gcIndex",
			fmt.Sprintf("6|accessTS=%d|binID=%d|%0x064x", item.AccessTimestamp, item.BinID, item.Address),
			fmt.Sprintf("value=", ""))

		return false, nil
	}, nil)

	_ = p.db.pinIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		log.Info("pinIndex",
			fmt.Sprintf("6|%d", item.AccessTimestamp),
			fmt.Sprintf("treeSize=%d|storeTS=%d", item.TreeSize, item.StoreTimestamp))

		return false, nil
	}, nil)







	return "Check the swarm log file for the output"


}

func (p *PinApi) IsHashPinned (addr []byte) bool{

	var foundIt bool

	_ = p.db.pinIndex.Iterate(func(item shed.Item) (stop bool, err error) {

		if len(addr) != len(item.Address) {
			foundIt = false
			return true, nil
		}

		for i := range addr {
			if addr[i] != item.Address[i] {
				foundIt = false
				return true, nil
			}
		}

		return false, nil
	}, nil)

	return foundIt

}




func (p *PinApi) PinHash() string {

	// see if hash is valid and present in local DB

	// call loopAndPinHash (hash)


	return "Pin called"
}

//
//
//func (p *PinApi)  UnPinHash(hash Address) string {
//
//	// See if the root hash is pinned
//
//	// call
//
//
//	return "UnPin called"
//}

//func (p *PinApi)  ListPinnedHashes() PinInfo {
//
//
//	return pinInfo
//}
//
//
//
//func loopAndPinHash(Address hashToPin) {
//
//	// for all the chunks in the hash
//
//	//   - Send to the Pin Queue
//
//	//   - When all chunks are pinned without error, return true otherwise false
//
//	//	 All pin ref. increment should be atomic
//
//	//
//
//}
//
//func loopAndUnpinHash(Address hashToUnpin) {
//
//	// for all the chunks in hash
//
//	//  - Send to the unpin Queue
//
//	//  - When all chunks are unpinned without error, return true otherwise false
//
//	//   All unpin ref. decrement should be atomic
//}
//
//
//func pinChunk(hunk chunkToPin) {
//
//	// < This should be spawned as a go-routine, 8 go-routine >
//
//	// read chunk from the pin Queue
//
//	// Increment the pinning reference counter
//
//}
//
//
//func unpinChunk(Chunk chunkToUnpin) {
//
//	// < This should be spawned as a go-routine, 8 go-routine >
//
//	// read the chunk address from the unpin Queue
//
//	// decrement the pinning reference counter
//
//}
