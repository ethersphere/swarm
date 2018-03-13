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
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	//sdbc "github.com/wolkdb/swarmdb/swarmdbcommon"
	"path/filepath"
	"strings"
	"github.com/ethereum/go-ethereum/swarm/swarmdb/ash"
	sdbp "github.com/ethereum/go-ethereum/swarm/swarmdb/sdbnetwork"
	"time"

	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/pss"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/swarmdb/ash"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"github.com/syndtr/goleveldb/leveldb"
)

type SwarmDB struct {
	tables       map[string]*Table
	dbchunkstore *DBChunkstore // Sqlite3 based
	ens          ENSSimulation
	swapdb       *SwapDBStore
	Netstats     *Netstats
	lstore		*storage.LocalStore
	api		*api.Api
	pss		*pss.Pss
	Sdbp		*sdbp.Sdbp
}

//for sql parsing
type QueryOption struct {
	Type           string //"Select" or "Insert" or "Update" probably should be an enum
	Owner          string
	Database       string
	Table          string
	Encrypted      int
	RequestColumns []sdbc.Column
	Inserts        []sdbc.Row
	Update         map[string]interface{} //'SET' portion: map[columnName]value
	Where          Where
	Ascending      int //1 true, 0 false (descending)
}

//for sql parsing
type Where struct {
	Left     string
	Right    string //all values are strings in query parsing
	Operator string //sqlparser.ComparisonExpr.Operator; sqlparser.BinaryExpr.Operator; sqlparser.IsExpr.Operator; sqlparser.AndExpr.Operator, sqlparser.OrExpr.Operator
}

type DBChunkstorage interface {
	RetrieveDBChunk(u *SWARMDBUser, key []byte) (val []byte, err error)
	StoreDBChunk(u *SWARMDBUser, val []byte, encrypted int) (key []byte, err error)
}

type Database interface {
	GetRootHash() []byte

	// Insert: adds key-value pair (value is an entire recrod)
	// ok - returns true if new key added
	// Possible Errors: KeySizeError, ValueSizeError, DuplicateKeyError, NetworkError, BufferOverflowError
	Insert(u *SWARMDBUser, key []byte, value []byte) (bool, error)

	// Put -- inserts/updates key-value pair (value is an entire record)
	// ok - returns true if new key added
	// Possible Errors: KeySizeError, ValueSizeError, NetworkError, BufferOverflowError
	Put(u *SWARMDBUser, key []byte, value []byte) (bool, error)

	// Get - gets value of key (value is an entire record)
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError
	Get(u *SWARMDBUser, key []byte) ([]byte, bool, error)

	// Delete - deletes key
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError, BufferOverflowError
	Delete(u *SWARMDBUser, key []byte) (bool, error)

	// Start/Flush - any buffered updates will be flushed to SWARM on FlushBuffer
	// ok - returns true if buffer started / flushed
	// Possible errors: NoBufferError, NetworkError
	StartBuffer(u *SWARMDBUser) (bool, error)
	FlushBuffer(u *SWARMDBUser) (bool, error)

	// Close - if buffering, then will flush buffer
	// ok - returns true if operation successful
	// Possible errors: NetworkError
	Close(u *SWARMDBUser) (bool, error)

	// prints what is in memory
	Print(u *SWARMDBUser)
}

type OrderedDatabase interface {
	Database
	// Seek -- moves cursor to key k
	// ok - returns true if key found, false if not found
	// Possible errors: KeySizeError, NetworkError
	Seek(u *SWARMDBUser, k []byte /*K*/) (e OrderedDatabaseCursor, ok bool, err error)
	SeekFirst(u *SWARMDBUser) (e OrderedDatabaseCursor, err error)
	SeekLast(u *SWARMDBUser) (e OrderedDatabaseCursor, err error)
}

type OrderedDatabaseCursor interface {
	Next(*SWARMDBUser) (k []byte /*K*/, v []byte /*V*/, err error)
	Prev(*SWARMDBUser) (k []byte /*K*/, v []byte /*V*/, err error)
}

const (
	DATABASE_NAME_LENGTH_MAX = 31
	TABLE_NAME_LENGTH_MAX    = 32
	DATABASES_PER_USER_MAX   = 30
	COLUMNS_PER_TABLE_MAX    = 30

	CHUNK_HASH_SIZE          = 32
	CHUNK_START_SIG          = 0
	CHUNK_END_SIG            = 65
	CHUNK_START_MSGHASH      = 65
	CHUNK_END_MSGHASH        = 97
	CHUNK_START_PAYER        = 97
	CHUNK_END_PAYER          = 129
	CHUNK_START_CHUNKTYPE    = 129
	CHUNK_END_CHUNKTYPE      = 130
	CHUNK_START_MINREP       = 130
	CHUNK_END_MINREP         = 131
	CHUNK_START_MAXREP       = 131
	CHUNK_END_MAXREP         = 132
	CHUNK_START_BIRTHTS      = 132
	CHUNK_END_BIRTHTS        = 140
	CHUNK_START_LASTUPDATETS = 140
	CHUNK_END_LASTUPDATETS   = 148
	CHUNK_START_ENCRYPTED    = 148
	CHUNK_END_ENCRYPTED      = 149
	CHUNK_START_VERSION      = 149
	CHUNK_END_VERSION        = 157
	CHUNK_START_RENEW        = 157
	CHUNK_END_RENEW          = 158
	CHUNK_START_KEY          = 158
	CHUNK_END_KEY            = 190
	CHUNK_START_OWNER        = 190
	CHUNK_END_OWNER          = 222
	CHUNK_START_DB           = 222
	CHUNK_END_DB             = 254
	CHUNK_START_TABLE        = 254
	CHUNK_END_TABLE          = 286
	//CHUNK_START_EPOCHTS      = 254
	//CHUNK_END_EPOCHTS        = 286
	CHUNK_START_ITERATOR = 416
	CHUNK_END_ITERATOR   = 512
	CHUNK_START_CHUNKVAL = 512
	CHUNK_END_CHUNKVAL   = 4096
)

func NewSwarmDB(config *SWARMDBConfig, lstore *storage.LocalStore, api *api.Api, pss *pss.Pss, sdbp *sdbp.Sdbp) (swdb *SwarmDB, err error) {
	sd := new(SwarmDB)
	sd.tables = make(map[string]*Table)
	sd.Sdbp = sdbp

	sd.Netstats = NewNetstats(config)
	//sd.ldb = lstore.DbStore.GetLDBDatabase().GetLevelDB()

	sd.lstore = lstore
	dbchunkstore, err := NewDBChunkStore(config, sd.lstore, sd.Netstats)
	if err != nil {
		return swdb, sdbc.GenerateSWARMDBError(err, `[swarmdb:NewSwarmDB] NewDBChunkStore `+err.Error())
	} else {
		sd.dbchunkstore = dbchunkstore
	}

	// default /tmp/ens.db
	ensdbFileName := "ens.db"
	ensdbFullPath := filepath.Join(config.ChunkDBPath, ensdbFileName)
	ens, errENS := NewENSSimulation(ensdbFullPath)
	if errENS != nil {
		return swdb, sdbc.GenerateSWARMDBError(errENS, `[swarmdb:NewSwarmDB] NewENSSimulation `+errENS.Error())
	}
	sd.ens = ens

	swapDBFileName := "swap.db"
	swapDBFullPath := filepath.Join(config.ChunkDBPath, swapDBFileName)
	swapdbObj, errSwapDB := NewSwapDBStore(config, sd.Netstats)
	if errSwapDB != nil {
		return swdb, sdbc.GenerateSWARMDBError(errSwapDB, `[swarmdb:NewSwarmDB] NewSwapDB `+swapDBFullPath+`|`+errSwapDB.Error())
	}
	sd.swapdb = swapdbObj

	sd.api = api
	sd.pss = pss
	return sd, nil
}

// DBChunkStore  API

func (self *SwarmDB) GenerateSwapLog(startts int64, endts int64) (log []string, err error) {
	log, err = self.swapdb.GenerateSwapLog(startts, endts)
	if err != nil {
		return log, sdbc.GenerateSWARMDBError(err, "Unable to GenerateSwapLog")
	}
	return log, nil
}

func (self *SwarmDB) GenerateBuyerLog(startts int64, endts int64) (log []string, err error) {
	log, err = self.dbchunkstore.GenerateBuyerLog(startts, endts)
	if err != nil {
		return log, sdbc.GenerateSWARMDBError(err, "Unable to GenerateBuyerLog")
	}
	return log, nil
}

func (self *SwarmDB) GenerateFarmerLog(startts int64, endts int64) (log []string, err error) {
	log, err = self.dbchunkstore.GenerateFarmerLog(startts, endts)
	if err != nil {
		return log, sdbc.GenerateSWARMDBError(err, "Unable to GenerateFarmerLog")
	}
	return log, nil
}

func (self *SwarmDB) GenerateAshResponse(chunkId []byte, seed []byte, proofRequired bool, index int8) (resp ash.AshResponse, err error) {
	resp, err = self.dbchunkstore.RetrieveAsh(chunkId, seed, proofRequired, index)
	// output, _ := json.Marshal(res)	fmt.Printf("%s\n", string(output))
	if err != nil {
		return resp, sdbc.GenerateSWARMDBError(err, "Unable to Retrieve Ash")
	}
	return resp, nil
}

func (self *SwarmDB) RetrieveDBChunk(u *SWARMDBUser, key []byte) (val []byte, err error) {
	val, err = self.dbchunkstore.RetrieveChunk(u, key)
	//TODO: SWARMDBError
	return val, err
}

func (self *SwarmDB) StoreDBChunk(u *SWARMDBUser, val []byte, encrypted int) (key []byte, err error) {
	key, err = self.dbchunkstore.StoreChunk(u, val, encrypted)
	//TODO: SWARMDBError
	return key, err
}

// ENSSimulation  API
func (self *SwarmDB) GetRootHash(u *SWARMDBUser, tblKey []byte /* GetTableKeyValue */) (roothash []byte, err error) {
	log.Debug(fmt.Sprintf("[GetRootHash] Getting Root Hash for (%s)[%x] ", tblKey, tblKey))
	return self.ens.GetRootHash(u, tblKey)
}

func (self *SwarmDB) StoreRootHash(u *SWARMDBUser, fullTableName []byte /* GetTableKey Value */, roothash []byte) (err error) {
	return self.ens.StoreRootHash(u, fullTableName, roothash)
}

// parse sql and return rows in bulk (order by, group by, etc.)
func (self *SwarmDB) QuerySelect(u *SWARMDBUser, query *QueryOption) (rows []sdbc.Row, err error) {
	table, err := self.GetTable(u, query.Owner, query.Database, query.Table)
	if err != nil {
		return rows, sdbc.GenerateSWARMDBError(err, `[swarmdb:QuerySelect] GetTable `+err.Error())
	}

	//var rawRows []sdbc.Row
	log.Debug(fmt.Sprintf("QueryOwner is: [%s]\n", query.Owner))
	colRows, err := self.Scan(u, query.Owner, query.Database, query.Table, table.primaryColumnName, query.Ascending)
	if err != nil {
		return rows, sdbc.GenerateSWARMDBError(err, `[swarmdb:QuerySelect] Scan `+err.Error())
	}
	//fmt.Printf("\nColRows = [%+v]", colRows)

	//apply WHERE
	whereRows, err := table.applyWhere(colRows, query.Where)
	if err != nil {
		return rows, sdbc.GenerateSWARMDBError(err, `[swarmdb:QuerySelect] applyWhere `+err.Error())
	}
	log.Debug(fmt.Sprintf("QuerySelect applied where rows: %+v and number of rows returned = %d", whereRows, len(whereRows)))

	//filter for requested columns
	for _, row := range whereRows {
		// fmt.Printf("QS b4 filterRowByColumns row: %+v\n", row)
		fRow := filterRowByColumns(row, query.RequestColumns)
		// fmt.Printf("QS after filterRowByColumns row: %+v\n", fRow)
		if len(fRow) > 0 {
			rows = append(rows, fRow)
		}
	}
	// fmt.Printf("\nNumber of FINAL rows returned : %d", len(rows))

	//TODO: Put it in order for Ascending/GroupBy
	// fmt.Printf("\nQS returning: %+v\n", rows)
	return rows, nil
}

// Insert is for adding new data to the table
// example: 'INSERT INTO tablename (col1, col2) VALUES (val1, val2)
func (self *SwarmDB) QueryInsert(u *SWARMDBUser, query *QueryOption) (affectedRows int, err error) {

	table, err := self.GetTable(u, query.Owner, query.Database, query.Table)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, `[swarmdb:QueryInsert] GetTable `+err.Error())
	}
	affectedRows = 0
	for _, row := range query.Inserts {
		// check if primary column exists in Row
		if _, ok := row[table.primaryColumnName]; !ok {
			return affectedRows, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:QueryInsert] Insert row %+v needs primary column '%s' value", row, table.primaryColumnName), ErrorCode: 446, ErrorMessage: fmt.Sprintf("Insert Query Missing Primary Key [%]", table.primaryColumnName)}
		}
		// check if Row already exists
		if _, ok := table.columns[table.primaryColumnName]; !ok {
			return affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryInsert] table.columns check - %s", err.Error()))
		}
		convertedKey, err := convertJSONValueToKey(table.columns[table.primaryColumnName].columnType, row[table.primaryColumnName])
		if err != nil {
			return affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryInsert] convertJSONValueToKey - %s", err.Error()))
		}
		_, ok, err := table.Get(u, convertedKey)
		//log.Debug(fmt.Sprintf("Row already exists | [%s] | [%+v] | [%d]", existingByteRow, existingByteRow, len(existingByteRow)))
		if ok {
			return affectedRows, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:QueryInsert] Insert row key %s already exists | Error: %s", row[table.primaryColumnName], err), ErrorCode: 434, ErrorMessage: fmt.Sprintf("Record with key [%s] already exists.  If you wish to modify, please use UPDATE SQL statement or PUT", bytes.Trim(convertedKey, "\x00"))}
		}
		if err != nil {
			//TODO: why is this uncommented?
			//return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:QueryInsert] Error: %s", err.Error()), ErrorCode: 434, ErrorMessage: fmt.Sprintf("Record with key [%s] already exists.  If you wish to modify, please use UPDATE SQL statement or PUT", bytes.Trim(convertedKey, "\x00")}
		}
		// put the new Row in
		err = table.Put(u, row)
		if err != nil {
			return affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryInsert] Put %s", err.Error()))
		}
		affectedRows = affectedRows + 1
	}
	return affectedRows, nil
}

// Update is for modifying existing data in the table (can use a Where clause)
// example: 'UPDATE tablename SET col1=value1, col2=value2 WHERE col3 > 0'
func (self *SwarmDB) QueryUpdate(u *SWARMDBUser, query *QueryOption) (affectedRows int, err error) {
	table, err := self.GetTable(u, query.Owner, query.Database, query.Table)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryUpdate] GetTable %s", err.Error()))
	}

	// get all rows with Scan, using primary key column
	rawRows, err := self.Scan(u, query.Owner, query.Database, query.Table, table.primaryColumnName, query.Ascending)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryUpdate] Scan %s", err.Error()))
	}

	// check to see if Update cols are in pulled set
	for colname, _ := range query.Update {
		if _, ok := table.columns[colname]; !ok {
			return 0, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:QueryUpdate] Update SET column name %s is not in table", colname), ErrorCode: 445, ErrorMessage: fmt.Sprintf("Attempting to update a column [%s] which is not in table [%s]", colname, table.tableName)}
		}
	}

	// apply WHERE clause
	filteredRows, err := table.applyWhere(rawRows, query.Where)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryUpdate] applyWhere %s", err.Error()))
	}

	// set the appropriate columns in filtered set
	for i, row := range filteredRows {
		for colname, value := range query.Update {
			if _, ok := row[colname]; !ok {
				//return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:QueryUpdate] Update SET column name %s is not in filtered rows", colname), ErrorCode: , ErrorMessage:""}
				//TODO: need to actually add this cell if it's an update query and the columnname is actually "valid"
				continue
			}
			filteredRows[i][colname] = value
		}
	}

	// put the changed rows back into the table
	affectedRows = 0
	for _, row := range filteredRows {
		if len(row) > 0 {
			err := table.Put(u, row)
			if err != nil {
				return affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryUpdate] Put %s", err.Error()))
			}
			affectedRows = affectedRows + 1
		}
	}
	return affectedRows, nil
}

//Delete is for deleting data rows (can use a Where clause, not just a key)
//example: 'DELETE FROM tablename WHERE col1 = value1'
func (self *SwarmDB) QueryDelete(u *SWARMDBUser, query *QueryOption) (affectedRows int, err error) {
	table, err := self.GetTable(u, query.Owner, query.Database, query.Table)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryDelete] GetTable %s", err.Error()))
	}

	//get all rows with Scan, using Where's specified col
	rawRows, err := self.Scan(u, query.Owner, query.Database, query.Table, query.Where.Left, query.Ascending)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryDelete] Scan %s", err.Error()))
	}

	//apply WHERE clause
	filteredRows, err := table.applyWhere(rawRows, query.Where)
	if err != nil {
		return 0, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryDelete] applyWhere %s", err.Error()))
	}

	//delete the selected rows
	for _, row := range filteredRows {
		if p, okp := row[table.primaryColumnName]; okp {
			ok, err := table.Delete(u, p)
			if err != nil {
				return affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:QueryDelete] Delete %s", err.Error()))
			}
			if !ok {
				// TODO: if !ok, what should happen? return appropriate response -- number of records affected
			} else {
				affectedRows = affectedRows + 1
			}
		}
	}
	return affectedRows, nil
}

func (self *SwarmDB) Query(u *SWARMDBUser, query *QueryOption) (rows []sdbc.Row, affectedRows int, err error) {
	switch query.Type {
	case "Select":
		rows, err = self.QuerySelect(u, query)
		if err != nil {
			return rows, len(rows), sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Query] QuerySelect %s", err.Error()))
		}
		return rows, len(rows), nil
	case "Insert":
		affectedRows, err = self.QueryInsert(u, query)
		if err != nil {
			return rows, affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Query] QueryInsert %s", err.Error()))
		}
		return rows, affectedRows, nil
	case "Update":
		affectedRows, err = self.QueryUpdate(u, query)
		if err != nil {
			return rows, affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Query] QueryUpdate %s", err.Error()))
		}
		return rows, affectedRows, nil
	case "Delete":
		affectedRows, err = self.QueryDelete(u, query)
		if err != nil {
			return rows, affectedRows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Query] QueryDelete %s", err.Error()))
		}
		return rows, affectedRows, nil
	}
	return rows, 0, nil
}

func (self *SwarmDB) Scan(u *SWARMDBUser, owner string, database string, tableName string, columnName string, ascending int) (rows []sdbc.Row, err error) {
	tblKey := self.GetTableKey(owner, database, tableName)
	tbl, ok := self.tables[tblKey]
	if !ok {
		//TODO: how would this ever happen?
		return rows, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:Scan] No such table to scan [%s:%s] - [%s]", owner, database, tblKey), ErrorCode: 403, ErrorMessage: fmt.Sprintf("Table Does Not Exist:  Table: [%s] Database [%s] Owner: [%s]", tableName, database, owner)}
	}
	rows, err = tbl.Scan(u, columnName, ascending)
	if err != nil {
		return rows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Scan] Error doing table scan: [%s] %s", columnName, err.Error()))
	}
	rows, err = tbl.assignRowColumnTypes(rows)
	if err != nil {
		return rows, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:Scan] Error assigning column types to row values"))
	}
	// fmt.Printf("swarmdb Scan finished ok: %+v\n", rows)
	return rows, nil
}

func (self *SwarmDB) GetTable(u *SWARMDBUser, owner string, database string, tableName string) (tbl *Table, err error) {
	if len(owner) == 0 {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetTable] owner missing "), ErrorCode: 430, ErrorMessage: "Owner Missing"}
	}
	if len(database) == 0 {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetTable] database missing "), ErrorCode: 500, ErrorMessage: "Database Missing"}
	}
	if len(tableName) == 0 {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetTable] tablename missing "), ErrorCode: 426, ErrorMessage: "Table Name Missing"}
	}
	tblKey := self.GetTableKey(owner, database, tableName)
	log.Debug(fmt.Sprintf("Getting Table [%s] with the Owner [%s] from TABLES [%v]", tableName, owner, self.tables))
	if tbl, ok := self.tables[tblKey]; ok {
		log.Debug(fmt.Sprintf("Table[%v] with Owner [%s] Database %s found in tables, it is: %+v\n", tblKey, owner, database, tbl))
		return tbl, nil
	} else {
		tbl = self.NewTable(owner, database, tableName)
		err = tbl.OpenTable(u)
		if err != nil {
			return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:GetTable] OpenTable %s", err.Error()))
		}
		self.RegisterTable(owner, database, tableName, tbl)
		return tbl, nil
	}
}

// TODO: when there are errors, the error must be parsable make user friendly developer errors that can be trapped by Node.js, Go library, JS CLI
func (self *SwarmDB) SelectHandler(u *SWARMDBUser, data string) (resp sdbc.SWARMDBResponse, err error) {

	log.Debug(fmt.Sprintf("SelectHandler Input: %s\n", data))
	d, err := parseData(data)
	if err != nil {
		return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] parseData %s", err.Error()))
	}

	switch d.RequestType {
	case sdbc.RT_CREATE_DATABASE:
		err = self.CreateDatabase(u, d.Owner, d.Database, d.Encrypted)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] CreateDatabase %s", err.Error()))
		}

		return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil

	case sdbc.RT_DROP_DATABASE:
		ok, err := self.DropDatabase(u, d.Owner, d.Database)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] DropDatabase %s", err.Error()))
		}
		if ok {
			return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil
		} else {
			return sdbc.SWARMDBResponse{AffectedRowCount: 0}, nil
		}

	case sdbc.RT_LIST_DATABASES:
		self.Sdbp.SendTest()
		databases, err := self.ListDatabases(u, d.Owner)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] ListDatabases %s", err.Error()))
		}
		resp.Data = databases
		resp.MatchedRowCount = len(databases)
		return resp, nil

	case sdbc.RT_CREATE_TABLE:
		if len(d.Table) == 0 || len(d.Columns) == 0 {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] empty table and column"), ErrorCode: 417, ErrorMessage: "Invalid [CreateTable] Request: Missing Table and/or Columns"}
		}
		//TODO: Upon further review, could make a NewTable and then call this from tbl. ---
		_, err := self.CreateTable(u, d.Owner, d.Database, d.Table, d.Columns)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] CreateTable %s", err.Error()))
		}
		return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil

	case sdbc.RT_DROP_TABLE:
		ok, err := self.DropTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] DropTable %s", err.Error()))
		}
		if ok {
			return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil
		} else {
			return sdbc.SWARMDBResponse{AffectedRowCount: 0}, nil
		}

	case sdbc.RT_SCAN:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		rawRows, err := self.Scan(u, d.Owner, d.Database, d.Table, tbl.primaryColumnName, 1)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		resp.Data = rawRows
		resp.AffectedRowCount = len(resp.Data)
		return resp, nil

	case sdbc.RT_DESCRIBE_TABLE:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		tblcols, err := tbl.DescribeTable()
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] DescribeTable %s", err.Error()))
		}
		if len(tblcols) == 0 {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Table [%s] not found", d.Table), ErrorCode: 482, ErrorMessage: fmt.Sprintf("Cannot Describe Table [%s] as it was not found", d.Table)}
		}
		for _, colInfo := range tblcols {
			r := sdbc.NewRow()
			r["ColumnName"] = colInfo.ColumnName
			r["IndexType"] = colInfo.IndexType
			r["Primary"] = colInfo.Primary
			r["ColumnType"] = colInfo.ColumnType
			resp.Data = append(resp.Data, r)
		}
		return resp, nil

	case sdbc.RT_LIST_TABLES:
		tableNames, err := self.ListTables(u, d.Owner, d.Database)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] ListDatabases %s", err.Error()))
		}
		resp.Data = tableNames
		resp.MatchedRowCount = len(tableNames)
		log.Debug(fmt.Sprintf("returning resp %+v tablenames %+v Mrc %+v", resp, tableNames, len(tableNames)))
		return resp, nil
	case sdbc.RT_PUT:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		tblInfo, err := tbl.DescribeTable()
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] DescribeTable %s", err.Error()))
		}
		d.Rows, err = tbl.assignRowColumnTypes(d.Rows)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] assignRowColumnTypes %s", err.Error()))
		}

		//error checking for primary column, and valid columns
		for _, row := range d.Rows {
			log.Debug(fmt.Sprintf("checking row %v\n", row))
			if _, ok := row[tbl.primaryColumnName]; !ok {
				return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Put row %+v needs primary column '%s' value", row, tbl.primaryColumnName), ErrorCode: 428, ErrorMessage: "Row missing primary key"}
			}
			for columnName, _ := range row {
				if _, ok := tblInfo[columnName]; !ok {
					return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Put row %+v has unknown column %s", row, columnName), ErrorCode: 429, ErrorMessage: fmt.Sprintf("Row contains unknown column [%s]", columnName)}
				}
			}
			// check to see if row already exists in table (no overwriting, TODO: check if that is right??)
			/* TODO: we want to have PUT blindly update.  INSERT will fail on duplicate and need to confirm what to do if multiple rows attempted to be inserted and just some are dupes
			if _, ok := tbl.columns[tbl.primaryColumnName]; !ok {
				return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Put row %+v has unknown column %s", row, columnName), ErrorCode: 429, ErrorMessage: fmt.Sprintf("Row contains unknown column [%s]", columnName)}
			}
			primaryColumnType := tbl.columns[tbl.primaryColumnName].columnType
			convertedKey, err := convertJSONValueToKey(primaryColumnType, row[tbl.primaryColumnName])
			if err != nil {
				return resp, sdbc.GenerateSWARMDBError( err, fmt.Sprintf("[swarmdb:SelectHandler] convertJSONValueToKey %s", err.Error()) )
			}
			validBytes, err := tbl.Get(u, convertedKey)
			if err == nil {
				validRow, err2 := tbl.byteArrayToRow(validBytes)
				if err2 != nil {
					return resp, sdbc.GenerateSWARMDBError( err2, fmt.Sprintf("[swarmdb:SelectHandler] byteArrayToRow %s", err2.Error()) )
				}
				return resp sdbc.GenerateSWARMDBError( err, fmt.Sprintf("[swarmdb:SelectHandler] Row with that primary key already exists: %+v", validRow) )
			} else {
				fmt.Printf("good, row wasn't found\n")
			}
			*/
		}

		//put the rows in
		successfulRows := 0
		for _, row := range d.Rows {
			err = tbl.Put(u, row)
			if err != nil {
				return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] Put %s", err.Error()))
			}
			successfulRows++
		}
		return sdbc.SWARMDBResponse{AffectedRowCount: successfulRows}, nil

	case sdbc.RT_GET:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		if isNil(d.Key) {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Get - Missing Key"), ErrorCode: 433, ErrorMessage: "GET Request Missing Key"}
		}
		if _, ok := tbl.columns[tbl.primaryColumnName]; !ok {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Get - Primary Key Not found in Column Definition"), ErrorCode: 479, ErrorMessage: "Table Definition Missing Primary Key"}
		}
		primaryColumnType := tbl.columns[tbl.primaryColumnName].columnType
		convertedKey, err := convertJSONValueToKey(primaryColumnType, d.Key)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] convertJSONValueToKey %s", err.Error()))
		}
		byteRow, ok, err := tbl.Get(u, convertedKey)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] Get %s", err.Error()))
		}

		if ok {
			validRow, err2 := tbl.byteArrayToRow(byteRow)
			if err2 != nil {
				return resp, sdbc.GenerateSWARMDBError(err2, fmt.Sprintf("[swarmdb:SelectHandler] byteArrayToRow %s", err2.Error()))
			}
			resp.Data = append(resp.Data, validRow)
			resp.MatchedRowCount = 1
		}
		return resp, nil

	case sdbc.RT_DELETE:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		if isNil(d.Key) {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Delete is Missing Key"), ErrorCode: 448, ErrorMessage: "Delete Statement missing KEY"}
		}
		ok, err := tbl.Delete(u, d.Key)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] Delete %s", err.Error()))
		}
		if ok {
			return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil
		}
		return sdbc.SWARMDBResponse{AffectedRowCount: 0}, nil

	case sdbc.RT_START_BUFFER:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		err = tbl.StartBuffer(u)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] StartBuffer %s", err.Error()))
		}
		//TODO: update to use real "count"
		return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil

	case sdbc.RT_FLUSH_BUFFER:
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		err = tbl.FlushBuffer(u)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] FlushBuffer %s", err.Error()))
		}
		//TODO: update to use real "count"
		return sdbc.SWARMDBResponse{AffectedRowCount: 1}, nil

	case sdbc.RT_QUERY:
		if len(d.RawQuery) == 0 {
			return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] RawQuery is blank"), ErrorCode: 425, ErrorMessage: "Invalid Query Request. Missing Rawquery"}
		}
		query, err := ParseQuery(d.RawQuery)
		query.Encrypted = d.Encrypted
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] ParseQuery [%s] %s", d.RawQuery, err.Error()))
		}
		query.Owner = d.Owner
		query.Database = d.Database
		if len(d.Table) == 0 {
			//TODO: check if empty even after query.Table check
			d.Table = query.Table //since table is specified in the query we do not have get it as a separate input
		}
		tbl, err := self.GetTable(u, d.Owner, d.Database, d.Table)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] GetTable %s", err.Error()))
		}
		tblInfo, err := tbl.DescribeTable()
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] DescribeTable %s", err.Error()))
		}

		//checking validity of columns
		for _, reqCol := range query.RequestColumns {
			if _, ok := tblInfo[reqCol.ColumnName]; !ok {
				return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Requested col [%s] does not exist in table [%+v]", reqCol.ColumnName, tblInfo), ErrorCode: 404, ErrorMessage: fmt.Sprintf("Column Does Not Exist in table definition: [%s]", reqCol.ColumnName)}
			}
		}

		//checking the Where clause
		if query.Type == "Select" && len(query.Where.Left) > 0 {
			if _, ok := tblInfo[query.Where.Left]; !ok {
				return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Query col [%s] does not exist in table", query.Where.Left), ErrorCode: 432, ErrorMessage: fmt.Sprintf("WHERE Clause contains invalid column [%s]", query.Where.Left)}
			}

			//checking if the query is just a primary key Get
			if query.Where.Left == tbl.primaryColumnName && query.Where.Operator == "=" {
				// fmt.Printf("Calling Get from Query\n")
				if _, ok := tbl.columns[tbl.primaryColumnName]; !ok {
					return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] Query col [%s] does not exist in table", tbl.primaryColumnName), ErrorCode: 432, ErrorMessage: fmt.Sprintf("Primary key [%s] not defined in table", tbl.primaryColumnName)}
				}
				convertedKey, err := convertJSONValueToKey(tbl.columns[tbl.primaryColumnName].columnType, query.Where.Right)
				if err != nil {
					return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] convertJSONValueToKey %s", err.Error()))
				}

				byteRow, ok, err := tbl.Get(u, convertedKey)
				if err != nil {
					return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] Get %s", err.Error()))
				}
				if ok {
					row, err := tbl.byteArrayToRow(byteRow)
					// fmt.Printf("Response row from Get: %s (%v)\n", row, row)
					if err != nil {
						return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] byteArrayToRow %s", err.Error()))
					}

					filteredRow := filterRowByColumns(row, query.RequestColumns)
					// fmt.Printf("\nResponse filteredrow from Get: %s (%v)", filteredRow, filteredRow)
					resp.Data = append(resp.Data, filteredRow)
				}
				return resp, nil
			}
		}

		// process the query
		qRows, affectedRows, err := self.Query(u, &query)
		if err != nil {
			return resp, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:SelectHandler] Query [%+v] %s", query, err.Error()))
		}
		return sdbc.SWARMDBResponse{AffectedRowCount: affectedRows, Data: qRows}, nil

	} //end switch

	return resp, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:SelectHandler] RequestType invalid: [%s]", d.RequestType), ErrorCode: 418, ErrorMessage: "Request Invalid"}

}

func parseData(data string) (*sdbc.RequestOption, error) {
	udata := new(sdbc.RequestOption)
	if err := json.Unmarshal([]byte(data), udata); err != nil {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:parseData] Unmarshal %s", err.Error()), ErrorCode: 432, ErrorMessage: "Unable to Parse Request"}
	}
	return udata, nil
}

func (self *SwarmDB) NewTable(owner string, database string, tableName string) *Table {
	t := new(Table)
	t.swarmdb = self
	t.Owner = owner
	t.Database = database
	t.tableName = tableName
	t.columns = make(map[string]*ColumnInfo)

	return t
}

func (self *SwarmDB) RegisterTable(owner string, database string, tableName string, t *Table) {
	// register the Table in SwarmDB
	tblKey := self.GetTableKey(owner, database, tableName)
	self.tables[tblKey] = t
}

func (self *SwarmDB) UnregisterTable(owner string, database string, tableName string) {
	// register the Table in SwarmDB
	tblKey := self.GetTableKey(owner, database, tableName)
	delete(self.tables, tblKey)
}

func (self *SwarmDB) BuildChunkHeader(u *SWARMDBUser, owner []byte, database []byte, tableName []byte, key []byte, value []byte, birthts int, version int, nodeType []byte, encrypted int) (ch []byte, err error) {
	ch = make([]byte, CHUNK_START_CHUNKVAL)
	copy(ch[CHUNK_START_OWNER:CHUNK_END_OWNER], owner)
	copy(ch[CHUNK_START_DB:CHUNK_END_DB], database)
	copy(ch[CHUNK_START_TABLE:CHUNK_END_TABLE], tableName)
	copy(ch[CHUNK_START_KEY:CHUNK_END_KEY], key)
	copy(ch[CHUNK_START_PAYER:CHUNK_END_PAYER], u.Address)
	copy(ch[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE], nodeType) // O = OWNER | D = Database | T = Table | x,h,d,k = various data nodes
	copy(ch[CHUNK_START_RENEW:CHUNK_END_RENEW], IntToByte(u.AutoRenew))
	copy(ch[CHUNK_START_MINREP:CHUNK_END_MINREP], IntToByte(u.MinReplication))
	copy(ch[CHUNK_START_MAXREP:CHUNK_END_MAXREP], IntToByte(u.MaxReplication))
	copy(ch[CHUNK_START_ENCRYPTED:CHUNK_END_ENCRYPTED], IntToByte(encrypted))
	copy(ch[CHUNK_START_BIRTHTS:CHUNK_END_BIRTHTS], IntToByte(birthts))

	lastupdatets := int(time.Now().Unix())
	copy(ch[CHUNK_START_LASTUPDATETS:CHUNK_END_LASTUPDATETS], IntToByte(lastupdatets))

	copy(ch[CHUNK_START_VERSION:CHUNK_END_VERSION], IntToByte(version))

	rawMetadata := ch[CHUNK_END_MSGHASH:CHUNK_START_CHUNKVAL]
	msg_hash := SignHash(rawMetadata)

	//TODO: msg_hash --
	copy(ch[CHUNK_START_MSGHASH:CHUNK_END_MSGHASH], msg_hash)

	km := self.dbchunkstore.GetKeyManager()
	sdataSig, errSign := km.SignMessage(msg_hash)
	if errSign != nil {
		return ch, &sdbc.SWARMDBError{Message: `[kademliadb:buildSdata] SignMessage ` + errSign.Error(), ErrorCode: 455, ErrorMessage: "Keymanager Unable to Sign Message"}
	}

	//TODO: Sig -- document this
	copy(ch[CHUNK_START_SIG:CHUNK_END_SIG], sdataSig)
	//log.Debug(fmt.Sprintf("Metadata is [%+v]", ch))

	return ch, err

	/*
	   mergedBodycontent = make([]byte, CHUNK_SIZE)
	   copy(mergedBodycontent[:], metadataBody)
	   copy(mergedBodycontent[CHUNK_START_CHUNKVAL:CHUNK_END_CHUNKVAL], value) // expected to be the encrypted body content

	   log.Debug(fmt.Sprintf("Merged Body Content: [%v]", mergedBodycontent))
	   return mergedBodycontent, err
	*/
}

// creating a database results in a new entry, e.g. "videos" in the owners ENS e.g. "wolktoken.eth" stored in a single chunk
// e.g.  key 1: wolktoken.eth (up to 64 chars)
//       key 2: videos     => 32 byte hash, pointing to tables of "video'
func (self *SwarmDB) CreateDatabase(u *SWARMDBUser, owner string, database string, encrypted int) (err error) {
	// this is the 32 byte version of the database name
	if len(database) > DATABASE_NAME_LENGTH_MAX {
		return &sdbc.SWARMDBError{Message: "[swarmdb:CreateDatabase] Database exists already", ErrorCode: 500, ErrorMessage: "Database Name too long (max is 32 chars)"}
	}

	ownerHash := crypto.Keccak256([]byte(owner))
	newDBName := make([]byte, DATABASE_NAME_LENGTH_MAX) //TODO: confirm use of constant ok -- making consistent with other DB names
	copy(newDBName[0:], database)

	// look up what databases the owner has already
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateDatabase] GetRootHash %s", err))
	}

	ownerChunk := make([]byte, CHUNK_SIZE)
	log.Debug(fmt.Sprintf("[swarmdb:CreateDatabase] Getting Root Hash using ownerHash [%x] and got [%x]", ownerHash, ownerDatabaseChunkID))

	if EmptyBytes(ownerDatabaseChunkID) {
		// put the 32-byte ownerHash in the first 32 bytes
		log.Debug(fmt.Sprintf("Creating new %s - %x\n", owner, ownerHash))
		//Create New Owner Chunk
		copy(ownerChunk[0:CHUNK_HASH_SIZE], []byte(ownerHash))
	} else {
		//Retrieve Owner Chunk "O" Chunk
		ownerChunk, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateDatabase] RetrieveDBChunk %s", err))
		}

		// the first 32 bytes of the ownerChunk should match
		if bytes.Compare(ownerChunk[0:CHUNK_HASH_SIZE], ownerHash[0:CHUNK_HASH_SIZE]) != 0 {
			return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateDatabase] Invalid owner %x != %x", ownerHash, ownerChunk[0:32]), ErrorCode: 450, ErrorMessage: fmt.Sprintf("Owner [%s] is invalid", owner)}
			//TODO: understand how/when this would occur
		}

		// check if there is already a database entry
		for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
			if bytes.Equal(ownerChunk[i:(i+DATABASE_NAME_LENGTH_MAX)], newDBName) {
				return &sdbc.SWARMDBError{Message: "[swarmdb:CreateDatabase] Database exists already", ErrorCode: 500, ErrorMessage: "Database Exists Already"}
			}
		}
	}

	//ownerChunkHeader :=
	for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
		// find the first 000 byte entry
		if EmptyBytes(ownerChunk[i:(i + 64)]) {
			// make a new database chunk, with the first 32 bytes of the chunk being the database name (the next keys will be the tables)
			bufDB := make([]byte, CHUNK_SIZE)
			copy(bufDB[0:DATABASE_NAME_LENGTH_MAX], newDBName[0:DATABASE_NAME_LENGTH_MAX])

			newDBHash, err := self.StoreDBChunk(u, bufDB, encrypted)
			if err != nil {
				return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateDatabase] StoreDBChunk %s", err.Error()))
			}

			// save the owner chunk, with the name + new DB hash
			copy(ownerChunk[i:(i+DATABASE_NAME_LENGTH_MAX)], newDBName[0:DATABASE_NAME_LENGTH_MAX])
			log.Debug(fmt.Sprintf("Saving Database with encrypted bit of %d at possition: %d", encrypted, i+DATABASE_NAME_LENGTH_MAX))
			if encrypted > 0 {
				ownerChunk[i+DATABASE_NAME_LENGTH_MAX] = 1
			} else {
				ownerChunk[i+DATABASE_NAME_LENGTH_MAX] = 0
			}
			copy(ownerChunk[(i+CHUNK_HASH_SIZE):(i+CHUNK_HASH_SIZE+32)], newDBHash[0:CHUNK_HASH_SIZE])
			log.Debug(fmt.Sprintf("Buffer has encrypted bit of %d ", ownerChunk[i+DATABASE_NAME_LENGTH_MAX]))

			ownerDatabaseChunkID, err = self.StoreDBChunk(u, ownerChunk, 0) // this could be a function of the top level domain .pri/.eth
			if err != nil {
				return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateDatabase] StoreDBChunk %s", err.Error()))
			}

			err = self.StoreRootHash(u, ownerHash, ownerDatabaseChunkID)
			if err != nil {
				return sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateDatabase] StoreRootHash %s", err.Error()))
			}
			return nil
		}
	}
	return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateDatabase] Database could not be created -- exceeded allocation"), ErrorCode: 451, ErrorMessage: fmt.Sprintf("Database could not be created -- exceeded allocation of %d", DATABASE_NAME_LENGTH_MAX)}
}

func (self *SwarmDB) ListDatabases(u *SWARMDBUser, owner string) (ret []sdbc.Row, err error) {
	ownerHash := crypto.Keccak256([]byte(owner))
	// look up what databases the owner has
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return ret, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:ListDatabases] GetRootHash %s", err))
	}

	ownerChunk := make([]byte, CHUNK_SIZE)
	if EmptyBytes(ownerDatabaseChunkID) {

	} else {
		ownerChunk, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return ret, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:ListDatabases] RetrieveDBChunk %s", err))
		}

		// the first 32 bytes of the ownerChunk should match
		if bytes.Compare(ownerChunk[0:32], ownerHash[0:32]) != 0 {
			return ret, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListDatabases] Invalid owner %x != %x", ownerHash, ownerChunk[0:CHUNK_HASH_SIZE]), ErrorCode: 450, ErrorMessage: "Invalid Owner Specified"}
		}

		// check if there is already a database entry
		for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
			if EmptyBytes(ownerChunk[i:(i + DATABASE_NAME_LENGTH_MAX)]) {
			} else {
				r := sdbc.NewRow()
				db := string(bytes.Trim(ownerChunk[i:(i+DATABASE_NAME_LENGTH_MAX)], "\x00"))
				log.Debug(fmt.Sprintf("DB: %s | %v BUF %s | %v ", db, db, ownerChunk[i:(i+32)], ownerChunk[i:(i+32)]))
				//rowstring := fmt.Sprintf("{\"database\":\"%s\"}", db)
				r["database"] = db
				ret = append(ret, r)
			}
		}
	}

	return ret, nil
}

// dropping a database removes the ENS entry
func (self *SwarmDB) DropDatabase(u *SWARMDBUser, owner string, database string) (ok bool, err error) {
	if len(database) > DATABASE_NAME_LENGTH_MAX {
		return false, &sdbc.SWARMDBError{Message: "[swarmdb:CreateDatabase] Database exists already", ErrorCode: 500, ErrorMessage: "Database Name too long (max is 32 chars)"}
	}

	// this is the 32 byte version of the database name
	ownerHash := crypto.Keccak256([]byte(owner))
	dropDBName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(dropDBName[0:], database)

	// look up what databases the owner has already
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] GetRootHash %s", err)}
	}

	ownerChunk := make([]byte, CHUNK_SIZE)
	if EmptyBytes(ownerDatabaseChunkID) {
		return false, nil // No error returned.  Just 'nil' it.  &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] No database %s", err)}
	} else {
		ownerChunk, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] RetrieveDBChunk %s", err)}
		}

		// the first 32 bytes of the ownerChunk should match
		if bytes.Compare(ownerChunk[0:CHUNK_HASH_SIZE], ownerHash[0:CHUNK_HASH_SIZE]) != 0 {
			return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] Invalid owner %x != %x", ownerHash, ownerChunk[0:CHUNK_HASH_SIZE])}
		}

		// check for the database entry
		for i := CHUNK_START_CHUNKVAL + 64; i < CHUNK_SIZE; i += 64 {
			if bytes.Compare(ownerChunk[i:(i+DATABASE_NAME_LENGTH_MAX)], dropDBName) == 0 {
				// found it, zero out the database
				copy(ownerChunk[i:(i+64)], make([]byte, 64))
				ownerDatabaseChunkID, err = self.StoreDBChunk(u, ownerChunk, 0) // TODO: .eth disc
				if err != nil {
					return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:DropDatabase] StoreDBChunk %s", err.Error()))
				}
				err = self.StoreRootHash(u, ownerHash, ownerDatabaseChunkID)
				if err != nil {
					return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:DropDatabase] StoreRootHash %s", err.Error()))
				}
				return true, nil
			}
		}
	}
	return false, nil // &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] Database could not be found")}
}

func (self *SwarmDB) DropTable(u *SWARMDBUser, owner string, database string, tableName string) (ok bool, err error) {
	log.Debug(fmt.Sprintf("Attempting to Drop table [%s]", tableName))
	if len(tableName) > TABLE_NAME_LENGTH_MAX {
		return false, &sdbc.SWARMDBError{Message: "[swarmdb:DropTable] Tablename length", ErrorCode: 500, ErrorMessage: "Table Name too long (max is 32 chars)"}
	}

	// this is the 32 byte version of the database name
	ownerHash := crypto.Keccak256([]byte(owner))
	dbName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(dbName[0:], database)

	dropTableName := make([]byte, TABLE_NAME_LENGTH_MAX)
	copy(dropTableName[0:], tableName)

	// look up what databases the owner has already
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] GetRootHash %s", err)}
	}

	buf := make([]byte, CHUNK_SIZE)
	if EmptyBytes(ownerDatabaseChunkID) {
		return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] No owner found %s", err)}
	} else {
		buf, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] RetrieveDBChunk %s", err)}
		}

		// the first 32 bytes of the buf should match
		if bytes.Compare(buf[0:32], ownerHash[0:32]) != 0 {
			return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] Invalid owner %x != %x", ownerHash, buf[0:32])}
		}

		// check for the database entry
		foundTable := false
		for i := 64; i < CHUNK_SIZE; i += 64 {
			if bytes.Compare(buf[i:(i+DATABASE_NAME_LENGTH_MAX)], dbName) == 0 {
				// found it - read the encryption level
				encrypted := 0
				if buf[i+DATABASE_NAME_LENGTH_MAX] > 0 {
					encrypted = 1
				}

				databaseHash := make([]byte, 32)
				copy(databaseHash[:], buf[(i+32):(i+64)])

				// bufDB has the tables!
				bufDB := make([]byte, CHUNK_SIZE)
				bufDB, err = self.RetrieveDBChunk(u, databaseHash)
				if err != nil {
					return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] RetrieveDBChunk %s", err)}
				}

				// nuke the table name in bufDB and write the updated bufDB
				for j := 64; j < CHUNK_SIZE; j += 64 {
					if bytes.Compare(bufDB[j:(j+TABLE_NAME_LENGTH_MAX)], dropTableName) == 0 {
						foundTable = true
						log.Debug(fmt.Sprintf("Found Table: %s - Attempting to delete from DB Chunk"))
						blankN := make([]byte, TABLE_NAME_LENGTH_MAX)
						copy(bufDB[j:(j+TABLE_NAME_LENGTH_MAX)], blankN[0:TABLE_NAME_LENGTH_MAX])
						databaseHash, err := self.StoreDBChunk(u, bufDB, encrypted)
						log.Debug(fmt.Sprintf("Update DB Chunk after blanking out [%s]", dropTableName))
						if err != nil {
							return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] StoreDBChunk %s", err)}
						}
						// update the database hash in the owner's databases
						copy(buf[(i+32):(i+64)], databaseHash[0:32])
						ownerDatabaseChunkID, err = self.StoreDBChunk(u, buf, 0) // TODO: review
						log.Debug(fmt.Sprintf("Updating Owner Chunk with new DB hash of [%s]", buf[(i+32):(i+64)]))
						if err != nil {
							return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropTable] StoreDBChunk %s", err)}
						}

						log.Debug(fmt.Sprintf("Storing new OwnerDatabaseChunkID of [%s]", ownerDatabaseChunkID))
						err = self.StoreRootHash(u, ownerHash, ownerDatabaseChunkID)
						if err != nil {
							return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:DropTable] StoreRootHash %s", err.Error()))
						}
						break
					}
				}
			}
		}
		if !foundTable {
			return false, nil
		}
		//Drop Table from ENS hash as well as db columns
		tblKey := self.GetTableKey(owner, database, tableName)
		emptyRootHash := make([]byte, 64)
		err = self.StoreRootHash(u, []byte(tblKey), emptyRootHash)
		//TODO: Empty out column info?
		if err != nil {
			return false, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[table:OpenTable] GetRootHash for table [%s]: %v", tblKey, err))
		}
		self.UnregisterTable(owner, database, tableName)
		return true, nil
	}
	return false, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:DropDatabase] Database could not be found")}
}

func (self *SwarmDB) ListTables(u *SWARMDBUser, owner string, database string) (tableNames []sdbc.Row, err error) {
	// this is the 32 byte version of the database name
	ownerHash := crypto.Keccak256([]byte(owner))
	dbName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	copy(dbName[0:], database)

	// look up what databases the owner has already
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] GetRootHash %s", err)}
	}

	buf := make([]byte, CHUNK_SIZE)
	if EmptyBytes(ownerDatabaseChunkID) {
		return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] Requested owner [%s] not found", owner), ErrorCode: 477, ErrorMessage: fmt.Sprintf("Requested owner [%s] not found", owner)}
	} else {
		buf, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] RetrieveDBChunk %s", err)}
		}

		// the first 32 bytes of the buf should match
		if bytes.Compare(buf[0:32], ownerHash[0:32]) != 0 {
			return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] Invalid owner %x != %x", ownerHash, buf[0:32])}
		}

		// check for the database entry
		for i := 64; i < CHUNK_SIZE; i += 64 {
			if bytes.Compare(buf[i:(i+DATABASE_NAME_LENGTH_MAX)], dbName) == 0 {
				// found it - read the encryption level
				databaseHash := make([]byte, 32)
				copy(databaseHash[:], buf[(i+32):(i+64)])

				// bufDB has the tables!
				bufDB := make([]byte, CHUNK_SIZE)
				bufDB, err = self.RetrieveDBChunk(u, databaseHash)
				if err != nil {
					return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] RetrieveDBChunk %s", err)}
				}

				for j := 64; j < CHUNK_SIZE; j += 64 {
					if EmptyBytes(bufDB[j:(j + TABLE_NAME_LENGTH_MAX)]) {
					} else {
						r := sdbc.NewRow()
						r["table"] = string(bytes.Trim(bufDB[j:(j+TABLE_NAME_LENGTH_MAX)], "\x00"))
						tableNames = append(tableNames, r)
					}
				}
				return tableNames, nil
			}
		}
	}
	return tableNames, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:ListTables] Did not find database %s", database), ErrorCode: 476, ErrorMessage: fmt.Sprintf("Database [%s] Not Found", database)}
}

// TODO: Review adding owner string, database string input parameters where the goal is to get database.owner/table/key type HTTP urls like:
//       https://swarm.wolk.com/wolkinc.eth => GET: ListDatabases
//       https://swarm.wolk.com/videos.wolkinc.eth => GET; ListTables
//       https://swarm.wolk.com/videos.wolkinc.eth/user => GET: DescribeTable
//       https://swarm.wolk.com/videos.wolkinc.eth/user/sourabhniyogi => GET: Get
// TODO: check for the existence in the owner-database combination before creating.
// TODO: need to make sure the types of the columns are correct
func (self *SwarmDB) CreateTable(u *SWARMDBUser, owner string, database string, tableName string, columns []sdbc.Column) (tbl *Table, err error) {
	columnsMax := COLUMNS_PER_TABLE_MAX
	primaryColumnName := ""
	if len(columns) > columnsMax {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] Max Allowed Columns for a table is %s and you submit %s", columnsMax, len(columns)), ErrorCode: 409, ErrorMessage: fmt.Sprintf("Max Allowed Columns exceeded - [%d] supplied, max is [MaxNumColumns]", len(columns), columnsMax)}
	}

	if len(tableName) > TABLE_NAME_LENGTH_MAX {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] Maximum length of table name exceeded (max %d chars)", TABLE_NAME_LENGTH_MAX), ErrorCode: 500, ErrorMessage: fmt.Sprintf("Max table name length exceeded")}
	}

	//error checking
	for _, columninfo := range columns {
		if columninfo.Primary > 0 {
			if len(primaryColumnName) > 0 {
				return tbl, &sdbc.SWARMDBError{Message: "[swarmdb:CreateTable] More than one primary column", ErrorCode: 406, ErrorMessage: "Multiple Primary keys specified in Create Table"}
			}
			primaryColumnName = columninfo.ColumnName
		}
		if !CheckColumnType(columninfo.ColumnType) {
			return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] bad columntype"), ErrorCode: 407, ErrorMessage: "Invalid ColumnType: [columnType]"}
		}
		if !CheckIndexType(columninfo.IndexType) {
			return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] bad indextype"), ErrorCode: 408, ErrorMessage: "Invalid IndexType: [indexType]"}
		}
	}
	if len(primaryColumnName) == 0 {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] no primary column indicated"), ErrorCode: 405, ErrorMessage: "No Primary Key specified in Create Table"}
	}

	// creating a database results in a new entry, e.g. "videos" in the owners ENS e.g. "wolktoken.eth" stored in a single chunk
	// e.g.  key 1: wolktoken.eth (up to 64 chars)
	//       key 2: videos     => 32 byte hash, pointing to tables of "video'
	ownerHash := crypto.Keccak256([]byte(owner))
	databaseName := make([]byte, DATABASE_NAME_LENGTH_MAX)
	databaseHash := make([]byte, 32)
	copy(databaseName[0:], database)

	// look up what databases the owner has already
	ownerDatabaseChunkID, err := self.ens.GetRootHash(u, ownerHash)
	if err != nil {
		return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:GetDatabase] GetRootHash %s", err))
	}
	log.Debug(fmt.Sprintf("[swarmdb:CreateTable] GetRootHash using ownerHash (%x) for DBChunkID => (%x)", ownerHash, ownerDatabaseChunkID))
	var buf []byte
	var bufDB []byte
	dbi := 0
	encrypted := 0
	if EmptyBytes(ownerDatabaseChunkID) {
		return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetDatabase] No database", err), ErrorCode: 443, ErrorMessage: "Database Specified Not Found"}
	} else {
		found := false
		// buf holds a list of the owner's databases
		buf, err = self.RetrieveDBChunk(u, ownerDatabaseChunkID)
		if err != nil {
			return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:GetDatabase] RetrieveDBChunk %s", err))
		}

		// the first 32 bytes of the buf should match the ownerHash
		if bytes.Compare(buf[0:32], ownerHash[0:32]) != 0 {
			return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetDatabase] Invalid owner %x != %x", ownerHash, buf[0:32]), ErrorCode: 450, ErrorMessage: "Invalid Owner Specified"}
		}

		// look for the database
		for i := 64; i < CHUNK_SIZE; i += 64 {
			if (bytes.Compare(buf[i:(i+DATABASE_NAME_LENGTH_MAX)], databaseName) == 0) && (found == false) {
				log.Debug(fmt.Sprintf("Found Database [%s] and it's encrypted bit is: [%+v]", databaseName, buf[i+DATABASE_NAME_LENGTH_MAX]))
				if buf[i+DATABASE_NAME_LENGTH_MAX] > 0 {
					encrypted = 1
				}
				// database is found, so we have the databaseHash now
				dbi = i
				databaseHash = make([]byte, 32)
				copy(databaseHash[:], buf[(i+32):(i+64)])
				// bufDB has the tables
				log.Debug(fmt.Sprintf("Pulled bufDB using [%x]", databaseHash))
				bufDB, err = self.RetrieveDBChunk(u, databaseHash)
				if err != nil {
					return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:GetDatabase] RetrieveDBChunk %s", err))
				}
				found = true
				break //TODO: think this should be ok?
			}
		}
		if !found {
			return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:GetDatabase] Database could not be found"), ErrorCode: 443, ErrorMessage: "Database Specified Not Found"}
		}
	}

	// add table to bufDB
	found := false
	for i := 64; i < CHUNK_SIZE; i += 64 {
		if EmptyBytes(bufDB[i:(i + 32)]) {
			if found == true {
			} else {
				// update the table name in bufDB and write the chunk
				tblN := make([]byte, 32)
				copy(tblN[0:32], tableName)
				copy(bufDB[i:(i+32)], tblN[0:32])
				log.Debug(fmt.Sprintf("Copying tableName [%s] to bufDB [%s]", tblN[0:32], bufDB[i:(i+32)]))
				newdatabaseHash, err := self.StoreDBChunk(u, bufDB, encrypted)
				if err != nil {
					return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] StoreDBChunk %s", err))
				}

				// update the database hash in the owner's databases
				copy(buf[(dbi+32):(dbi+64)], newdatabaseHash[0:32])
				ownerDatabaseChunkID, err = self.StoreDBChunk(u, buf, 0) // TODO

				if err != nil {
					return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] StoreDBChunk %s", err))
				}
				log.Debug(fmt.Sprintf("[swarmdb:CreateTable] Storing Hash of (%x) and ChunkID: [%s]", ownerHash, ownerDatabaseChunkID))
				err = self.StoreRootHash(u, ownerHash, ownerDatabaseChunkID)
				if err != nil {
					return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] StoreRootHash %s", err.Error()))
				}
				found = true
				break //TODO: This ok?
			}
		} else {
			tbl0 := string(bytes.Trim(bufDB[i:(i+32)], "\x00"))
			log.Debug(fmt.Sprintf("Comparing tableName [%s](%+v) to tbl0 [%s](%+v)", tableName, tableName, tbl0, tbl0))
			if strings.Compare(tableName, tbl0) == 0 {
				return tbl, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:CreateTable] table exists already"), ErrorCode: 500, ErrorMessage: "Table exists already"}
			}
		}
	}

	// ok now make the table!
	log.Debug(fmt.Sprintf("Creating Table [%s] - Owner [%s] Database [%s]\n", tableName, owner, database))
	tbl = self.NewTable(owner, database, tableName)
	tbl.encrypted = encrypted
	for i, columninfo := range columns {
		copy(buf[2048+i*64:], columninfo.ColumnName)
		b := make([]byte, 1)
		b[0] = byte(columninfo.Primary)
		copy(buf[2048+i*64+26:], b)

		intColumnInfo, _ := ColumnTypeToInt(columninfo.ColumnType)
		//TODO: check this
		b[0] = byte(intColumnInfo)
		copy(buf[2048+i*64+28:], b)

		intIndexType := IndexTypeToInt(columninfo.IndexType)
		b[0] = byte(intIndexType)
		copy(buf[2048+i*64+30:], b) // columninfo.IndexType
		// fmt.Printf(" column: %v\n", columninfo)
	}

	//Could (Should?) be less bytes, but leaving space in case more is to be there
	copy(buf[4000:4024], IntToByte(tbl.encrypted))

	log.Debug(fmt.Sprintf("Storing Table with encrypted bit set to %d [%v]", tbl.encrypted, buf[4000:4024]))
	swarmhash, err := self.StoreDBChunk(u, buf, tbl.encrypted)
	if err != nil {
		return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] StoreDBChunk %s", err.Error()))
	}
	tbl.primaryColumnName = primaryColumnName
	tbl.roothash = swarmhash

	tblKey := self.GetTableKey(tbl.Owner, tbl.Database, tbl.tableName)

	log.Debug(fmt.Sprintf("**** CreateTable (owner [%s] database [%s] tableName: [%s]) Primary: [%s] tblKey: [%s] Roothash:[%x]\n", tbl.Owner, tbl.Database, tbl.tableName, tbl.primaryColumnName, tblKey, swarmhash))
	err = self.StoreRootHash(u, []byte(tblKey), []byte(swarmhash))
	if err != nil {
		return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] StoreRootHash %s", err.Error()))
	}
	err = tbl.OpenTable(u)
	if err != nil {
		return tbl, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[swarmdb:CreateTable] OpenTable %s", err.Error()))
	}
	self.RegisterTable(owner, database, tableName, tbl)
	return tbl, nil
}

func (self *SwarmDB) GetTableKey(owner string, database string, tableName string) (key string) {
	return fmt.Sprintf("%s|%s|%s", owner, database, tableName)
}
