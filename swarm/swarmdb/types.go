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

//Comment for phobos-rebasetest 3/22 12:12 PM PT
//Comment for phobos-rebasetest 3/22 12:15 PM PT

package swarmdb

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/cznic/mathutil"
	//"github.com/ethereum/go-ethereum/log"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"math"
	"reflect"
	"strconv"
)

const (
	CHUNK_SIZE    = 4096
	VERSION_MAJOR = 0
	VERSION_MINOR = 1
	VERSION_PATCH = 3
	VERSION_META  = "poc"
)

var SWARMDBVersion = func() string {
	v := fmt.Sprintf("%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
	if VERSION_META != "" {
		v += "-" + VERSION_META
	}
	return v
}()

//for comparing rows in two different sets of data
//only 1 cell in the row has to be different in order for the rows to be different
func isDuplicateRow(row1 sdbc.Row, row2 sdbc.Row) bool {

	//if row1.primaryKeyValue == row2.primaryKeyValue {
	//	return true
	//}

	for k1, r1 := range row1 {
		if _, ok := row2[k1]; !ok {
			return false
		}
		if r1 != row2[k1] {
			return false
		}
	}

	for k2, r2 := range row2 {
		if _, ok := row1[k2]; !ok {
			return false
		}
		if r2 != row1[k2] {
			return false
		}
	}

	return true
}

type ChunkHeader struct {
	MsgHash        []byte
	Sig            []byte
	Payer          []byte
	NodeType       []byte
	MinReplication int
	MaxReplication int
	Birthts        int
	LastUpdatets   int
	Encrypted      int
	Version        int
	AutoRenew      int
	Key            []byte
	Owner          []byte
	Database       []byte
	Table          []byte
	//Epochts       []byte -- Do we need this in our Chunk?
	//Trailing Bytes
}

func ParseChunkHeader(chunk []byte) (ch ChunkHeader, err error) {
	/*
		if len(bytes.Trim(chunk, "\x00")) != CHUNK_SIZE {
			return ch, &sdbc.SWARMDBError{ Message: fmt.Sprintf("[types:ParseChunkHeader]"), ErrorCode: 480, ErrorMessage: fmt.Sprintf("Chunk of invalid size.  Expecting %d bytes, chunk is %d bytes", CHUNK_SIZE, len(chunk)) }
		}
	*/
	//fmt.Printf("Chunk is of size: %d and looking at %d to %d\n", len(chunk), CHUNK_START_MINREP, CHUNK_END_MINREP)
	//log.Debug(fmt.Sprintf("Chunk is of size: %d and looking at %d to %d ==> %+v\n%+v", CHUNK_SIZE, CHUNK_START_MINREP, CHUNK_END_MINREP, chunk[CHUNK_START_MINREP:CHUNK_END_MINREP], chunk))
	ch.MsgHash = chunk[CHUNK_START_MSGHASH:CHUNK_END_MSGHASH]
	ch.Sig = chunk[CHUNK_START_SIG:CHUNK_END_SIG]
	ch.Payer = chunk[CHUNK_START_PAYER:CHUNK_END_PAYER]
	ch.NodeType = chunk[CHUNK_START_CHUNKTYPE:CHUNK_END_CHUNKTYPE]
	ch.MinReplication = int(BytesToInt(chunk[CHUNK_START_MINREP:CHUNK_END_MINREP]))
	ch.MaxReplication = int(BytesToInt(chunk[CHUNK_START_MAXREP:CHUNK_END_MAXREP]))
	ch.Birthts = int(BytesToInt(chunk[CHUNK_START_BIRTHTS:CHUNK_END_BIRTHTS]))
	ch.LastUpdatets = int(BytesToInt(chunk[CHUNK_START_LASTUPDATETS:CHUNK_END_LASTUPDATETS]))
	ch.Encrypted = int(BytesToInt(chunk[CHUNK_START_ENCRYPTED:CHUNK_END_ENCRYPTED]))
	ch.Version = int(BytesToInt(chunk[CHUNK_START_VERSION:CHUNK_END_VERSION]))
	ch.AutoRenew = int(BytesToInt(chunk[CHUNK_START_RENEW:CHUNK_END_RENEW]))
	ch.Key = chunk[CHUNK_START_KEY:CHUNK_END_KEY]
	ch.Owner = chunk[CHUNK_START_OWNER:CHUNK_END_OWNER]
	ch.Database = chunk[CHUNK_START_DB:CHUNK_END_DB]
	ch.Table = chunk[CHUNK_START_TABLE:CHUNK_END_TABLE]
	//ch.Epochts = chunk[CHUNK_START_EPOCHTS:CHUNK_END_EPOCHTS])
	return ch, err
}

//gets data (Row) out of a slice of Rows, and rtns as one json.
func rowDataToJson(rows []sdbc.Row) (string, error) {
	var resRows []map[string]interface{}
	for _, row := range rows {
		resRows = append(resRows, row)
	}
	resBytes, err := json.Marshal(resRows)
	if err != nil {
		return "", err
	}
	return string(resBytes), nil
}

//json input string should be []map[string]interface{} format
func JsonDataToRow(in string) (rows []sdbc.Row, err error) {

	var jsonRows []map[string]interface{}
	if err = json.Unmarshal([]byte(in), &jsonRows); err != nil {
		return rows, err
	}
	for _, jRow := range jsonRows {
		row := sdbc.NewRow()
		row = jRow
		rows = append(rows, row)
	}
	return rows, nil
}

func stringToColumnType(in string, columnType sdbc.ColumnType) (out interface{}, err error) {
	switch columnType {
	case sdbc.CT_INTEGER:
		out, err = strconv.Atoi(in)
	case sdbc.CT_STRING:
		out = in
	case sdbc.CT_FLOAT:
		out, err = strconv.ParseFloat(in, 64)
	//case: sdbc.CT_BLOB:
	//?
	default:
		err = &sdbc.SWARMDBError{Message: "[types|stringToColumnType] columnType not found", ErrorCode: 434, ErrorMessage: fmt.Sprintf("ColumnType [%s] not SUPPORTED. Value [%s] rejected", columnType, in)}
	}
	return out, err
}

//gets only the specified Columns (column name and value) out of a single Row, returns as a Row with only the relevant data
func filterRowByColumns(row sdbc.Row, columns []sdbc.Column) (filteredRow sdbc.Row) {
	filteredRow = make(map[string]interface{})
	for _, col := range columns {
		if _, ok := row[col.ColumnName]; ok {
			filteredRow[col.ColumnName] = row[col.ColumnName]
		}
	}
	return filteredRow
}

func CheckColumnType(colType sdbc.ColumnType) bool {
	/*
		var ct uint8
		switch colType.(type) {
		case int:
			ct = uint8(colType.(int))
		case uint8:
			ct = colType.(uint8)
		case float64:
			ct = uint8(colType.(float64))
		case string:
			cttemp, _ := strconv.ParseUint(colType.(string), 10, 8)
			ct = uint8(cttemp)
		case ColumnType:
			ct = colType.(ColumnType)
		default:
			fmt.Printf("CheckColumnType not a type I can work with\n")
			return false
		}
	*/
	ct := colType
	if ct == sdbc.CT_INTEGER || ct == sdbc.CT_STRING || ct == sdbc.CT_FLOAT { //|| ct == sdbc.CT_BLOB {
		return true
	}
	return false
}

func CheckIndexType(it sdbc.IndexType) bool {
	if it == sdbc.IT_HASHTREE || it == sdbc.IT_BPLUSTREE { //|| it == sdbc.IT_FULLTEXT || it == sdbc.IT_FRACTALTREE || it == sdbc.IT_NONE {
		return true
	}
	return false
}

func StringToKey(columnType sdbc.ColumnType, key string) (k []byte) {
	k = make([]byte, 32)
	switch columnType {
	case sdbc.CT_INTEGER:
		// convert using atoi to int
		i, _ := strconv.Atoi(key)
		k8 := IntToByte(i) // 8 byte
		copy(k, k8)        // 32 byte
	case sdbc.CT_STRING:
		copy(k, []byte(key))
	case sdbc.CT_FLOAT:
		f, _ := strconv.ParseFloat(key, 64)
		k8 := FloatToByte(f) // 8 byte
		copy(k, k8)          // 32 byte
	case sdbc.CT_BLOB:
		// TODO: do this correctly with JSON treatment of binary
		copy(k, []byte(key))
	}
	return k
}

func KeyToString(columnType sdbc.ColumnType, k []byte) (out string) {
	switch columnType {
	case sdbc.CT_BLOB:
		return fmt.Sprintf("%v", k)
	case sdbc.CT_STRING:
		return fmt.Sprintf("%s", string(k))
	case sdbc.CT_INTEGER:
		a := binary.BigEndian.Uint64(k)
		return fmt.Sprintf("%d [%x]", a, k)
	case sdbc.CT_FLOAT:
		bits := binary.BigEndian.Uint64(k)
		f := math.Float64frombits(bits)
		return fmt.Sprintf("%f", f)
	}
	return "unknown key type"

}

func ValueToString(v []byte) (out string) {
	if IsHash(v) {
		return fmt.Sprintf("%x", v)
	} else {
		return fmt.Sprintf("%v", string(v))
	}
}

func EmptyBytes(hashid []byte) (valid bool) {
	valid = true
	for i := 0; i < len(hashid); i++ {
		if hashid[i] != 0 {
			return false
		}
	}
	return valid
}

func IsHash(hashid []byte) (valid bool) {
	cnt := 0
	for i := 0; i < len(hashid); i++ {
		if hashid[i] == 0 {
			cnt++
		}
	}
	if cnt > 3 {
		return false
	} else {
		return true
	}
}

func ByteToColumnType(b byte) (ct sdbc.ColumnType, err error) {
	switch b {
	case 1:
		return sdbc.CT_INTEGER, err
	case 2:
		return sdbc.CT_STRING, err
	case 3:
		return sdbc.CT_FLOAT, err
	case 4:
		return sdbc.CT_BLOB, err
	default:
		return sdbc.CT_INTEGER, &sdbc.SWARMDBError{Message: "Invalid Column Type", ErrorCode: 407, ErrorMessage: "Invalid Column Type"}
	}
}

func ByteToIndexType(b byte) (it sdbc.IndexType) {
	switch b {
	case 1:
		return sdbc.IT_HASHTREE
	case 2:
		return sdbc.IT_BPLUSTREE
	case 3:
		return sdbc.IT_FULLTEXT
	default:
		return sdbc.IT_NONE
	}
}

func ColumnTypeToInt(ct sdbc.ColumnType) (v int, err error) {
	switch ct {
	case sdbc.CT_INTEGER:
		return 1, err
	case sdbc.CT_STRING:
		return 2, err
	case sdbc.CT_FLOAT:
		return 3, err
	case sdbc.CT_BLOB:
		return 4, err
	default:
		return -1, &sdbc.SWARMDBError{Message: "[types|ColumnTypeToInt] columnType not found", ErrorCode: 434, ErrorMessage: fmt.Sprintf("ColumnType [%s] not SUPPORTED. Value [%s] rejected", ct, v)}
	}
}

func IndexTypeToInt(it sdbc.IndexType) (v int) {
	switch it {
	case sdbc.IT_HASHTREE:
		return 1
	case sdbc.IT_BPLUSTREE:
		return 2
	case sdbc.IT_FULLTEXT:
		return 3
	/*
		case "FRACTAL":
			//return sdbc.IT_FRACTALTREE
	*/
	case sdbc.IT_NONE:
		return 0
	default:
		return 0
	}
}

func IntToByte(i int) (k []byte) {
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, uint64(i))
	return k
}

func FloatToByte(f float64) (k []byte) {
	bits := math.Float64bits(f)
	k = make([]byte, 8)
	binary.BigEndian.PutUint64(k, bits)
	return k
}

func BytesToFloat(b []byte) (f float64) {
	bits := binary.BigEndian.Uint64(b)
	f = math.Float64frombits(bits)
	return f
}

func BytesToInt64(b []byte) (i int64) {
	if len(bytes.Trim(b, "\x00")) == 0 {
		return 0
	}
	i = int64(binary.BigEndian.Uint64(b))
	return i
}

func BytesToInt(b []byte) (i int) {
	if len(bytes.Trim(b, "\x00")) == 0 {
		return 0
	}
	i = int(binary.BigEndian.Uint64(b))
	return i
}

func SHA256(inp string) (k []byte) {
	h := sha256.New()
	h.Write([]byte(inp))
	k = h.Sum(nil)
	return k
}

func convertJSONValueToKey(columnType sdbc.ColumnType, pvalue interface{}) (k []byte, err error) {
	// fmt.Printf(" *** convertJSONValueToKey: CONVERT %v (columnType %v)\n", pvalue, columnType)
	switch svalue := pvalue.(type) {
	case (int):
		i := fmt.Sprintf("%d", svalue)
		k = StringToKey(columnType, i)
	case (float64):
		f := ""
		switch columnType {
		case sdbc.CT_INTEGER:
			f = fmt.Sprintf("%d", int(svalue))
		case sdbc.CT_FLOAT:
			f = fmt.Sprintf("%f", svalue)
		case sdbc.CT_STRING:
			f = fmt.Sprintf("%f", svalue)
		}
		k = StringToKey(columnType, f)
	case (string):
		k = StringToKey(columnType, svalue)
	default:
		return k, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swarmdb:convertJSONValueToKey] Unknown Type: %v", reflect.TypeOf(svalue)), ErrorCode: 429, ErrorMessage: fmt.Sprintf("Column Value is an unsupported type of [%s]", svalue)}
	}
	return k, nil
}

func isNil(a interface{}) bool {
	if a == nil { // || reflect.ValueOf(a).IsNil()  {
		return true
	}
	return false
}

func Rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32/4, math.MaxInt32/4, false)
	if err != nil {
		panic(err)
	}
	return x
}
