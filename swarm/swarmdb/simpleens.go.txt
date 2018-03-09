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
	// "database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
	// "encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ENSSimple struct {
	auth *bind.TransactOpts
	sens *Simplestens
}

func NewENSSimple(path string) (ens ENSSimple, err error) {
	// Create an IPC based RPC connection to a remote node
	//y	conn, err := ethclient.Dial("/home/karalabe/.ethereum/testnet/geth.ipc")

	//conn, err := ethclient.Dial("/var/www/vhosts/data/geth.ipc")      // this is working OK
	//conn, err := ethclient.Dial("http://127.0.0.1:8545")              // this is working OK	   //  JSON-RPC Endpoint   https://github.com/ethereum/wiki/wiki/JSON-RPC
	//conn, err := ethclient.Dial("http://35.224.194.195:8545")
	conn, err := ethclient.Dial("http://ens.wolk.com:8545")

	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	var key = `{"address":"90fb0de606507e989247797c6a30952cae4d5cbe","crypto":{"cipher":"aes-128-ctr","ciphertext":"54396d6ed0335e4b4874cd4440d24eabeca895fcbafb15d310c25c6b1e4bb306","cipherparams":{"iv":"e3a2457cf8420d3072e5adf118d31df8"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"d25987f2f2429e53f51d87eb6474e3f12a67c63603fd860b558657cee19a6ea9"},"mac":"023fc8a29a6e323db43e0c7795d2d59d0c1f295a62cbb9bc625951fca9c385dd"},"id":"dc849ada-c6be-4f12-bfa2-5200ec560c2e","version":3}`
	auth, err := bind.NewTransactor(strings.NewReader(key), "mdotm")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	} else {
		ens.auth = auth
	}

	// Instantiate the contract and display its name
	sens, err := NewSimplestens(common.HexToAddress("0x6120c3f1fdcd20c384b82eb20d93eef7838e0363"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Simplestens contract: %v", err)
	} else {
		ens.sens = sens
	}

	// -------------------
	/*
		db, err := sql.Open("sqlite3", path)
		if err != nil {
			return ens, err
		}
		if db == nil {
			return ens, err
		}
		ens.db = db
		ens.filepath = path

		sql_table := `
		CREATE TABLE IF NOT EXISTS ens (
		indexName TEXT NOT NULL PRIMARY KEY,
		roothash BLOB,
		storeDT DATETIME
		);
		`

		_, err = db.Exec(sql_table)
		if err != nil {
			return ens, err
		}
	*/
	return ens, nil
}

func (self *ENSSimple) StoreRootHash(indexName []byte, roothash []byte) (err error) {
	var i32 [32]byte
	var r32 [32]byte
	copy(i32[0:], indexName)
	copy(r32[0:], roothash)

	tx, err2 := self.sens.SetContent(self.auth, i32, r32)
	if err2 != nil {
		return err // log.Fatalf("Failed to set Content: %v", err2)
	}
	fmt.Printf("i32: %x r32: %x tx: %v\n", i32, r32, tx.Hash())

	/*
		sql_add := `INSERT OR REPLACE INTO ens ( indexName, roothash, storeDT ) values(?, ?, CURRENT_TIMESTAMP)`
		stmt, err := self.db.Prepare(sql_add)
		if err != nil {
			return (err)
		}
		defer stmt.Close()

		_, err2 := stmt.Exec(indexName, roothash)
		if err2 != nil {
			return (err2)
		}
	*/
	return nil
}

func (self *ENSSimple) GetRootHash(indexName []byte) (val []byte, err error) {
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
		fmt.Printf("GetContent failed:  %v", err)
		return val, err
	}
	val = make([]byte, 32)
	for i := range s {
		val[i] = s[i]
		if i == 31 {
			break
		}
	}
	//copy(val[0:], s[0:32])
	fmt.Printf("indexName: [%x] => s: [%x] val: [%x]\n", indexName, s, val)
	return val, nil
}
