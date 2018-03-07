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
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	//sdbc "swarmdbcommon"
	_ "github.com/mattn/go-sqlite3"
)

type ENSSimulation struct {
	filepath string
	db       *sql.DB
}

func NewENSSimulation(path string) (ens ENSSimulation, err error) {
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
	return ens, nil
}

func (self *ENSSimulation) StoreRootHash(u *SWARMDBUser, indexName []byte, roothash []byte) (err error) {
	log.Debug(fmt.Sprintf("[enssimulation:StoreRootHash] indexName: (%s)[%x] => roothash[%x]", indexName, indexName, roothash))
	sql_add := `INSERT OR REPLACE INTO ens ( indexName, roothash, storeDT ) values(?, ?, CURRENT_TIMESTAMP)`
	stmt, err := self.db.Prepare(sql_add)
	if err != nil {
		log.Debug("Error storing RootHash")
		return &sdbc.SWARMDBError{Message: fmt.Sprintf("[enssimulation:StoreRootHash] sql.db.Prepare [%s]", err.Error()), ErrorCode: 441, ErrorMessage: "Error Storing RootHash"}
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(indexName, roothash)
	if err2 != nil {
		log.Debug("Error storing RootHash")
		return &sdbc.SWARMDBError{Message: fmt.Sprintf("[enssimulation:StoreRootHash] stmt.Exec [%s]", err2.Error()), ErrorCode: 441, ErrorMessage: "Error Storing RootHash"}
	}
	return nil
}

func (self *ENSSimulation) GetRootHash(u *SWARMDBUser, indexName []byte) (val []byte, err error) {
	//TODO: why are we passing in 'u' but not using?
	log.Debug(fmt.Sprintf("[enssimulation:GetRootHash] indexName: (%s)[%x] => roothash[%x]", indexName, indexName)) //, roothash))
	sql := `SELECT roothash FROM ens WHERE indexName = $1`
	stmt, err := self.db.Prepare(sql)
	if err != nil {
		return val, &sdbc.SWARMDBError{Message: fmt.Sprintf("[enssimulation:GetRootHash] sql.db.Prepare [%s]", err.Error()), ErrorCode: 442, ErrorMessage: "Error Retrieving RootHash"}
	}
	defer stmt.Close()

	rows, err := stmt.Query(indexName)
	if err != nil {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[enssimulation:GetRootHash] sql.db.Prepare [%s]", err.Error()), ErrorCode: 442, ErrorMessage: "Error Retrieving RootHash"}
	}
	defer rows.Close()

	for rows.Next() {
		err2 := rows.Scan(&val)
		if err2 != nil {
			return nil, err2
		}
		log.Debug(fmt.Sprintf("[enssimulation:GetRootHash] Success Query indexName: [%x] => roothash: [%x]", indexName, val))
		return val, nil
	}
	return val, nil
}
