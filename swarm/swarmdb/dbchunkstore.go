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

package swarmdb

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/swarmdb/ash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
)

const (
	hashChunkSize = 4000
	epochSeconds  = 600
)

type DBChunkstore struct {
	lstore   *storage.LocalStore
	km       *KeyManager
	netstats *Netstats
	farmer   common.Address
	filepath string
}

type DBChunk struct {
	Val []byte
	Enc byte
}

type ChunkAsh struct {
	chunkID []byte //internal
	epoch   []byte //internal
	Seed    []byte
	Root    []byte
	Renewal byte
}

type ChunkLog struct {
	Farmer           string `json:"farmer"`
	ChunkID          string `json:"chunkID"`
	ChunkHash        []byte `json:"-"`
	ChunkBD          int    `json:"chunkBD"`
	ChunkSD          int    `json:"chunkSD"`
	ReplicationLevel int    `json:"rep"`
	Renewable        int    `json:"renewable"`
}

func (u *ChunkAsh) MarshalJSON() ([]byte, error) {
	mash := ash.Computehash(u.Root)
	epochstr := common.ToHex(u.epoch)

	return json.Marshal(
		&struct {
			Epoch   string `json: "epoch"`
			ChunkID string `json: "chunkID"`
			Seed    string `json: "seed"`
			Mash    string `json: "mash"`
			Renewal byte   `json: "renew"`
		}{
			Epoch:   epochstr,
			ChunkID: hex.EncodeToString(u.chunkID),
			Seed:    hex.EncodeToString(u.Seed),
			Mash:    hex.EncodeToString(mash),
			Renewal: u.Renewal,
		})
}

func NewDBChunkStore(config *SWARMDBConfig, swarmlstore *storage.LocalStore, netstats *Netstats) (self *DBChunkstore, err error) {
	path := config.ChunkDBPath

	//ldb, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return self, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:NewDBChunkstore] Unable to Open DB at path [%s] | Error: %s", path, err), ErrorCode: 499, ErrorMessage: fmt.Sprintf("[dbchunkstore:NewDBChunkstore] Unable to Open DB [%s] | %s", path, err)}
	}

	km, errKM := NewKeyManager(config)
	if errKM != nil {
		return nil, sdbc.GenerateSWARMDBError(errKM, fmt.Sprintf("[dbchunkstore:NewDBChunkStore] NewKeyManager %s", errKM.Error()))
	}

	userWallet := config.Address
	walletAddr := common.HexToAddress(userWallet)

	self = &DBChunkstore{
		lstore:   swarmlstore,
		km:       &km,
		farmer:   walletAddr,
		filepath: path,
		netstats: netstats,
	}
	return self, nil
}

func (self *DBChunkstore) GetKeyManager() (km *KeyManager) {
	return self.km
}

func (self *DBChunkstore) StoreKChunk(u *SWARMDBUser, key []byte, val []byte, encrypted int) (err error) {
	self.netstats.StoreChunk()
	_, err = self.storeChunkInSwarm(u, val, encrypted, key)
	return err
}

func (self *DBChunkstore) StoreChunk(u *SWARMDBUser, val []byte, encrypted int) (key []byte, err error) {
	//self.netstats.StoreChunk() -- TODO: Review with Michael and Sourabh
	return self.storeChunkInSwarm(u, val, encrypted, key)
}

func (self *DBChunkstore) storeChunkInSwarm(u *SWARMDBUser, val []byte, encrypted int, k []byte) (key []byte, err error) {
	if len(val) < CHUNK_SIZE {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:StoreChunk] Chunk too small (< %s)| %x", CHUNK_SIZE, val), ErrorCode: 439, ErrorMessage: "Unable to Store Chunk"}
	}

	var chunk DBChunk

	swarmChunk := storage.NewChunk(k, make(chan bool))
	//? or swarmChunk := NewChunk(k, nil)

	var finalSdata []byte
	finalSdata = make([]byte, CHUNK_SIZE)
	recordData := val[CHUNK_START_CHUNKVAL : CHUNK_END_CHUNKVAL-40] //MAJOR TODO: figure out how we pass in to ensure <=4096
	if len(k) > 0 {
		key = k
		finalSdata = make([]byte, CHUNK_SIZE)
		//log.Debug(fmt.Sprintf("Key: [%x][%v] After Loop recordData length (%d) and start pos %d", key, key, len(recordData), CHUNK_START_CHUNKVAL))
		copy(finalSdata[0:CHUNK_START_CHUNKVAL], val[0:CHUNK_START_CHUNKVAL])
		if encrypted > 0 {
			//log.Debug(fmt.Sprintf("StoreChunk of length %d: VAL (encrypting 0 to %d) = %v", len(val), CHUNK_START_CHUNKVAL, val))
			encVal := self.km.EncryptData(u, recordData)
			log.Debug(fmt.Sprintf("EncVal is %+v", encVal))
			copy(finalSdata[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], encVal)
		} else {
			copy(finalSdata[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], recordData)
		}
		chunk.Enc = 1
		val = finalSdata
	} else {
		inp := make([]byte, hashChunkSize)
		copy(inp, val[0:hashChunkSize])
		key = ash.Computehash(inp)
		if encrypted > 0 {
			chunk.Enc = 1
			val = self.km.EncryptData(u, val)
		}
	}

	chunk.Val = val
	swarmChunk.SData = val
	swarmChunk.Size = 4096
	self.lstore.Put(swarmChunk)
	//log.Debug(fmt.Sprintf("Storing the following data: %v", val))

	return key, nil
}

/*
func (self *DBChunkstore) storeChunkInDB(u *SWARMDBUser, val []byte, encrypted int, k []byte) (key []byte, err error) {
	if len(val) < CHUNK_SIZE {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:StoreChunk] Chunk too small (< %s)| %x", CHUNK_SIZE, val), ErrorCode: 439, ErrorMessage: "Unable to Store Chunk"}
	}
	var chunk DBChunk
	var finalSdata []byte
	finalSdata = make([]byte, CHUNK_SIZE)
	recordData := val[CHUNK_START_CHUNKVAL : CHUNK_END_CHUNKVAL-40] //MAJOR TODO: figure out how we pass in to ensure <=4096
	if len(k) > 0 {
		key = k
		finalSdata = make([]byte, CHUNK_SIZE)
		//log.Debug(fmt.Sprintf("Key: [%x][%v] After Loop recordData length (%d) and start pos %d", key, key, len(recordData), CHUNK_START_CHUNKVAL))
		copy(finalSdata[0:CHUNK_START_CHUNKVAL], val[0:CHUNK_START_CHUNKVAL])
		if encrypted > 0 {
			//log.Debug(fmt.Sprintf("StoreChunk of length %d: VAL (encrypting 0 to %d) = %v", len(val), CHUNK_START_CHUNKVAL, val))
			encVal := self.km.EncryptData(u, recordData)
			log.Debug(fmt.Sprintf("EncVal is %+v", encVal))
			copy(finalSdata[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], encVal)
		} else {
			copy(finalSdata[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], recordData)
		}
		chunk.Enc = 1
		val = finalSdata
	} else {
		inp := make([]byte, hashChunkSize)
		copy(inp, val[0:hashChunkSize])
		key = ash.Computehash(inp)
		if encrypted > 0 {
			chunk.Enc = 1
			val = self.km.EncryptData(u, val)
		}
	}

	chunk.Val = val
	//log.Debug(fmt.Sprintf("Storing the following data: %v", val))
	data, err := rlp.EncodeToBytes(chunk)
	if err != nil {
		return key, err
	}

	//log.Debug(fmt.Sprintf("LDB Put with key %x", key))
	err = self.ldb.Put(key, data, nil)
	if err != nil {
		return key, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:StoreChunk] Exec %s | encrypted:%s", err.Error(), encrypted), ErrorCode: 439, ErrorMessage: "Unable to Store Chunk"}
	}
	//log.Debug(fmt.Sprintf("Stored chunk with key %x", key))
	//fmt.Printf("storeChunkInDB enc: %d [%x] -- %x\n", chunk.Enc, key, data)

	if len(k) > 0 {
		chunkHeader, errCh := ParseChunkHeader(chunk.Val)
		if errCh != nil {
			return key, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:StoreChunk] ParseChunkHeader %s ", err.Error()), ErrorCode: 439, ErrorMessage: "Unable to Parse Chunk"}
		}

		// TODO: the TS here should be the FIRST time the chunk is originally written
		ts := int64(chunkHeader.LastUpdatets)
		epochPrefix := epochBytesFromTimestamp(ts)
		ekey := append(epochPrefix, key...)
		// fmt.Printf("%d --> %x --> %x\n", ts, epochPrefix, ekey)

		secret := make([]byte, 32)
		rand.Read(secret)
		//log.Debug(fmt.Sprintf("Generating Ash for key %x", key))
		roothash, err := ash.GenerateAsh(secret, chunk.Val)
		//log.Debug(fmt.Sprintf("Ash Generated is: %+v", roothash))
		if err != nil {
			return key, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:storeChunkInDB] Exec %s | encrypted:%s", err.Error(), secret), ErrorCode: 450, ErrorMessage: "Unable to Generate Proper ASH"}
		}

		chunkAsh := ChunkAsh{Seed: secret, Root: roothash}
		chunkAsh.Renewal = byte(chunkHeader.AutoRenew) //Renew bool should be passed in here

		ashdata, err := rlp.EncodeToBytes(chunkAsh)
		if err != nil {
			return key, err
		}
		err = self.ldb.Put(ekey, ashdata, nil)
		if err != nil {
			return key, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:StoreChunk] Exec %s | encrypted:%s", err.Error(), encrypted), ErrorCode: 439, ErrorMessage: "Unable to Store Chunk"}
		}
	}
	return key, nil
}
*/

func (self *DBChunkstore) RetrieveRawChunk(key []byte) (val []byte, err error) {
	/*
		data, err := self.ldb.Get(key, nil)
		if err == leveldb.ErrNotFound {
			val = make([]byte, CHUNK_SIZE)
			return val, nil
		} else if err != nil {
			return val, err
			//TODO: make swarmdberror
		}
		c := new(DBChunk)
		err = rlp.Decode(bytes.NewReader(data), c)
		if err != nil {
			return val, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:RetrieveRawChunk] Prepare %s", err.Error()), ErrorCode: 440, ErrorMessage: "Unable to Retrieve Chunk"}
		}
		self.netstats.RetrieveChunk()
	*/
	return val, nil
}

func (self *DBChunkstore) RetrieveChunk(u *SWARMDBUser, key []byte) (val []byte, err error) {
	swarmChunk, err := self.lstore.Get(storage.Key(key))
	val = swarmChunk.SData
	if string(swarmChunk.SData[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE]) == "k" {
		//log.Debug(fmt.Sprintf("Retrieving the following data: %v", c.Val))
		val = val[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL]
	}
	/*
		if c.Enc > 0 {
			val, err = self.km.DecryptData(u, val)
			if err != nil {
				return val, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:RetrieveChunk] DecryptData %s", err.Error()), ErrorCode: 440, ErrorMessage: "Unable to Retrieve Chunk"}
			}
		}
	*/
	var fullVal []byte
	fullVal = make([]byte, CHUNK_SIZE)
	if string(swarmChunk.SData[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE]) == "k" {
		copy(fullVal[0:CHUNK_START_CHUNKVAL], swarmChunk.SData[0:CHUNK_START_CHUNKVAL])
		copy(fullVal[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], val)
		val = fullVal
		//log.Debug(fmt.Sprintf("Decrypted Retrieved K Node => %+v\n", val))
	}
	return val, nil
}

func (self *DBChunkstore) RetrieveKChunk(u *SWARMDBUser, key []byte) (val []byte, err error) {
	log.Debug(fmt.Sprintf("Retrieving KChunk with key %x", key))
	val, err = self.RetrieveChunk(u, key)
	if err != nil {
		log.Debug(fmt.Sprintf("Error retrieving KChunk: %s", err.Error()))
		return val, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[dbchunkstore:RetrieveChunk] DecryptData %s", err.Error()))
	}
	//log.Debug(fmt.Sprintf("Retrieved KChunk %+v", val))
	jsonRecord := val[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL]
	return bytes.TrimRight(jsonRecord, "\x00"), nil
}

func epochBytesFromTimestamp(ts int64) (out []byte) {
	return IntToByte(int(ts / epochSeconds))
}

func (self *DBChunkstore) GenerateFarmerLog(startTS int64, endTS int64) (log []string, err error) {
	self.netstats.GenerateFarmerLog()
	return self.GenerateBuyerLog(startTS, endTS)
}

func (self *DBChunkstore) GenerateBuyerLog(startTS int64, endTS int64) (log []string, err error) {
	/*
		for ts := startTS; ts < endTS; ts += epochSeconds {
			epochPrefix := epochBytesFromTimestamp(ts)
			iter := self.ldb.NewIterator(util.BytesPrefix(epochPrefix), nil)
			for iter.Next() {
				epochkey := iter.Key()
				key := epochkey[8:]
				//fmt.Printf("%x\n", key)

				chunkash := new(ChunkAsh)
				err = rlp.Decode(bytes.NewReader(iter.Value()), chunkash)
				if err != nil {
					return log, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:GenerateBuyerLog] EKEY: %x | Prepare %s", epochkey, err.Error()), ErrorCode: 451, ErrorMessage: "Unable to Decode Chunkash"}
				}

				chunkash.chunkID = key
				chunkash.epoch = bytes.TrimLeft(epochkey[0:8], "\x00")
				output, _ := json.Marshal(chunkash)
				log = append(log, fmt.Sprintf("%s\n", string(output)))

				// data, err := self.ldb.Get(key, nil)
				// chunklog, err := json.Marshal(c)
				// sql_readall := fmt.Sprintf("SELECT chunkKey,strftime('%s',chunkBirthDT) as chunkBirthTS, strftime('%s',chunkStoreDT) as chunkStoreTS, maxReplication, renewal FROM chunk where chunkBD >= %d and chunkBD < %d", time.Unix(startTS, 0).Format(time.RFC3339), time.Unix(endTS, 0).Format(time.RFC3339))
			}
			iter.Release()
			err = iter.Error()
		}

		self.netstats.GenerateBuyerLog()
	*/
	return log, nil
}

func (self *DBChunkstore) RetrieveAsh(key []byte, secret []byte, proofRequired bool, auditIndex int8) (res ash.AshResponse, err error) {
	request := ash.AshRequest{ChunkID: key, Seed: secret}
	request.Challenge = &ash.AshChallenge{ProofRequired: proofRequired, Index: auditIndex}
	chunkval := make([]byte, 4128)
	chunkval, err = self.RetrieveRawChunk(request.ChunkID)
	if err != nil {
		return res, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:RetrieveAsh] %s", err.Error()), ErrorCode: 470, ErrorMessage: "RawChunk Retrieval Error"}
	}
	res, err = ash.ComputeAsh(request, chunkval)
	if err != nil {
		return res, &sdbc.SWARMDBError{Message: fmt.Sprintf("[dbchunkstore:RetrieveAsh] %s", err.Error()), ErrorCode: 471, ErrorMessage: "RetrieveAsh Error"}
	}
	self.netstats.RetrieveAsh()
	return res, nil
}
