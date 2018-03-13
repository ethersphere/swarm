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
	"context"
	"fmt"
    	"io/ioutil"
	"strings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	//"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/ethereum/go-ethereum/ethclient"
	elog "github.com/ethereum/go-ethereum/log"
	"path/filepath"
	"encoding/json"
	"time"

)

type ENSSimple struct {
	auth *bind.TransactOpts
	sens *Simplestens
	conn *ethclient.Client
	ldb  *leveldb.DB
}

type EnsData struct{
	Root []byte	 `json:"root"`
	Status	uint	 `json:"status"`
	PrevRoot []byte  `json:"prevroot,omitempty"`
}

type ENSSimpleConfig struct{
	Ipaddress	string	`json:"ipaddress,omitempty"`
}

func NewENSSimple(path string, config *SWARMDBConfig) (ens ENSSimple, err error) {
// TODO: using temporary config file
	elog.Debug(fmt.Sprintf("SimpleENS config %s %s %s", config.EnsIP, config.EnsKeyPath, config.EnsAddress))
	//ipaddress := config.EnsIP
//////debug
	var ipaddress string
	ipaddress = "/var/www/vhosts/data/geth.ipc"
	if len(config.EnsIP) > 0 {
		ipaddress = config.EnsIP
	}
	elog.Debug(fmt.Sprintf("SimpleENS ipaddress = %s", ipaddress))	
	
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial(ipaddress)
	if err != nil {
                return ens, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] NewENSSimple Connection `+err.Error())
	}
	ens.conn  = conn
	var ctx     context.Context
	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	h, err := conn.HeaderByNumber(ctx, nil)
	elog.Debug(fmt.Sprintf("SimpleENS h = %v err = %v", h, err))	

// TODO: need to get the dir (or filename) from config
//	k, err := ioutil.ReadFile(config.EnsKeyPath)
//debug
	keystoredir := "/var/www/vhosts/data/keystore"
	if len(config.EnsKeyPath) > 0{
		keystoredir = config.EnsKeyPath
	}
    	//files, err := ioutil.ReadDir("/var/www/vhosts/data/keystore")
    	files, err := ioutil.ReadDir(keystoredir)
	if err != nil {
                return ens, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] NewENSSimple Keystoredir `+err.Error())
	}
	var filename string
        for _, file := range files {
        	if strings.HasPrefix(file.Name(), "UTC") {
                	filename =  file.Name()
        	}
	}
        //fullpath := filepath.Join("/var/www/vhosts/data/keystore", filename)
        fullpath := filepath.Join(keystoredir, filename)
	k, err := ioutil.ReadFile(fullpath)
	if err != nil {
                return ens, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] NewENSSimple Keystorefile `+err.Error())
	}
	key := fmt.Sprintf("%s", k)
	
	auth, err := bind.NewTransactor(strings.NewReader(string(key)), "mdotm")
	if err != nil {
                return ens, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] NewENSSimple NewTransactor `+err.Error())
	}
	ens.auth = auth

	// Instantiate the contract and display its name
	//sens, err := NewSimplestens(common.HexToAddress("0x7e29ab7c40aaf6ca52270643b57c46c7766ca31d"), conn)
	sens, err := NewSimplestens(common.HexToAddress(config.EnsAddress), conn)
	if err != nil {
		elog.Debug(fmt.Sprintf("NewSimplestens failed %v", err))
		return ens, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] NewENSSimple NewSimplestens `+err.Error())
	}
	elog.Debug(fmt.Sprintf("NewSimplestens success %v", sens))
	ens.sens = sens

// TODO: get leveldb dir from config
	p := "/tmp/ensdb"
	if len(config.ChunkDBPath) > 0{
		p = filepath.Join(config.ChunkDBPath, "ensdb")
	}
	ldb, err := leveldb.OpenFile(p, nil)
	ens.ldb = ldb

	return ens, nil
}

func (self *ENSSimple) StoreRootHash(indexName []byte, roothash []byte) (err error) {
	var i32 [32]byte
	var r32 [32]byte
	copy(i32[0:], indexName)
	copy(r32[0:], roothash)
	elog.Debug(fmt.Sprintf("in ENSSimple StoreRootHash(len = %d) %x %x roothash (len = %d) %x %x ", len(indexName), indexName,i32, len(roothash), roothash, r32))
	fmt.Printf("ENSSimple StoreRootHash %x roothash %x\n", indexName, roothash)

	//status, err :=	self.sens.Content(self.auth, i32)
	//elog.Debug(fmt.Sprintf("ENSSimple StoreRootHash status %v err = %v", status, err))
/*
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	opts := &bind.CallOpts{Context: ctx}
	r, err := self.sens.SimplestensCaller.Context(opts, i32)
*/
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
//	h, err := self.conn.HeaderByNumber(ctx, nil)
//	elog.Debug(fmt.Sprintf("SimpleENS StoreRootHash h = %v err = %v", h, err))	
//	fmt.Printf("SimpleENS StoreRootHash h = %v err = %v", h, err)
        h, err := self.conn.HeaderByNumber(ctx, nil)
        elog.Debug(fmt.Sprintf("SimpleENS StoreRootHash self.conn.HeaderByNumber h = %v err = %v", h, err))

	tx, err := self.sens.SetContent(self.auth, i32, r32)
	if err != nil{
        	elog.Debug(fmt.Sprintf("SimpleENS StoreRootHash SetContent err = %v",err))
		self.StoreRootHashToLDB(indexName, roothash, 2, nil)
	}else{
//TODO: only for debugging
		self.StoreRootHashToLDB(indexName, roothash, 1, nil)
		elog.Debug(fmt.Sprintf("return store %x %v %x\n", tx.Hash(), err, tx))
/*
        	h, err = self.conn.HeaderByNumber(ctx, nil)
        	elog.Debug(fmt.Sprintf("SimpleENS StoreRootHash self.conn.HeaderByNumber h = %v err = %v", h,  err))
        	h, err = self.conn.HeaderByHash(ctx, tx.Hash())
        	elog.Debug(fmt.Sprintf("SimpleENS StoreRootHash self.conn.HeaderByHash h = %v err = %v", h, err))
		if err != nil {
			elog.Debug(fmt.Sprintf("ENSSimple StoreRootHash error %v", err2))
			return err // log.Fatalf("Failed to set Content: %v", err2)
		}
*/
	}

	elog.Debug(fmt.Sprintf("out ENSSimple StoreRootHash %x roothash %x", indexName, roothash))
	return nil
}

func (self *ENSSimple) StoreRootHashToLDB(indexName, roothash []byte, status uint, prevhash []byte)(err error){
	j, err := json.Marshal(EnsData{roothash, status, prevhash})
	elog.Debug(fmt.Sprintf("in ENSSimple StoreRootHashToLDB %v json = %v", indexName, j))
	if err != nil {
		return  GenerateSWARMDBError(err, `[swarmdb:ENSSimple] StoreRootHashToLDB `+err.Error())
	}
	err = self.ldb.Put(indexName, j , nil)
	if err != nil {
		return  GenerateSWARMDBError(err, `[swarmdb:ENSSimple] StoreRootHashToLDB `+err.Error())
	}
	return nil
}

func (self *ENSSimple) StoreRootHashWithStatus(indexName, roothash []byte, status uint)(err error){
	if status == 2{
		s := status
                err = self.StoreRootHash(indexName, roothash)
		var prevhash []byte
		if err == nil{
			s = 1
		}else{
			prevhash, _ = self.GetRootHash(indexName)
		}
                err = self.StoreRootHashToLDB(indexName, roothash, s, prevhash)
		return  GenerateSWARMDBError(err, `[swarmdb:ENSSimple] StoreRootHashWithStatus `+err.Error())
	}
        return self.StoreRootHashToLDB(indexName, roothash, status, nil)
}

func (self *ENSSimple) GetRootHashFromLDB(indexName []byte)(value []byte, status uint, err error){
	elog.Debug(fmt.Sprintf("in ENSSimple GetRootHashFromLDB %v", indexName))
        var d EnsData
        res, err := self.ldb.Get(indexName, nil)
	if err != nil && err != leveldb.ErrNotFound  {
		res, err = self.GetRootHash(indexName)
		return res, 0, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] GetRootHashFromLDB `+err.Error())
	}
	if err == nil{
        	err = json.Unmarshal(res, &d)
		elog.Debug(fmt.Sprintf("in ENSSimple GetRootHashFromLDB res = %v d = %v", res, d))
		if err != nil{
			return res, 0, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] GetRootHashFromLDB `+err.Error())
		}
	}
	return d.Root, d.Status, err
}

// status 
//   0: got data by GetRootHash (will be "need update")
//   1: updated by SetRootHash  (will be "just updated")
//   2: got error at SetRootHash(will be "got error at SetRootHash")

func (self *ENSSimple) GetRootHash(indexName []byte) (val []byte, err error) {
	elog.Debug(fmt.Sprintf("in ENSSimple GetRootHash %v", indexName))
	//status, err :=	self.sens.Content(self.auth, indexName)
	//elog.Debug(fmt.Sprintf("ENSSimple GetRootHash status %v err = %v", status, err)
	var d EnsData
	res, err := self.ldb.Get(indexName, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return res, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] GetRootHash `+err.Error())
	}
	if err == nil{
		err = json.Unmarshal(res, &d)
		if err != nil{
			return res, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] GetRootHash `+err.Error())
		}
	}
/*
	if d.Status == 1 {
		return d.Root, nil
	} 
*/
// TODO: check old value and decide to call store root hash
	if d.Status == 2{
		self.StoreRootHash(indexName, d.Root)
	}
	
	/*
		sql := `SELECT roothash FROM ens WHERE indexName = $1`
		stmt, err := self.db.Prepare(sql)
		if err != nil {
			return val, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(indexName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			err2 := rows.Scan(&val)
			if err2 != nil {
				return nil, err2
			}
			return val, nil
		}
	*/
	/*b, err := hex.DecodeString("9f5cd92e2589fadd191e7e7917b9328d03dc84b7a67773db26efb7d0a4635677")
	if err != nil {
		log.Fatalf("Failed to hexify %v", err)
	} */
	var b2 [32]byte
	copy(b2[0:], indexName)
	//s, err := sens.Content(b)
	s, err := self.sens.Content(nil, b2)
	if err != nil {
		elog.Debug(fmt.Sprintf("ENSSimple GetRootHash err %v %v", indexName, err))
		return val, GenerateSWARMDBError(err, `[swarmdb:ENSSimple] GetRootHash `+err.Error())
	}
	val = make([]byte, 32)
	for i := range s {
		val[i] = s[i]
		if i == 31 {
			break
		}
	}
	//copy(val[0:], s[0:32])
	elog.Debug(fmt.Sprintf("out ENSSimple GetRootHash %x s %x val %x", indexName, s, val))
	return val, nil
}
