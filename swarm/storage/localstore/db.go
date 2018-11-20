package localstore

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/swarm/shed"
	"github.com/syndtr/goleveldb/leveldb"
)

/*
   types of access:
   - just get the data
   - increment access index

   - when uploaded or pull synced
   - when delivered
   - when push synced
   - when accessed
*/

var (
	errInvalidMode = errors.New("invalid mode")
)

// Modes of access/update
const (
	SYNCING rushed.Mode = iota
	UPLOAD
	REQUEST
	SYNCED
	ACCESS
	REMOVAL
)

// DB is a local chunkstore using disk storage
type DB struct {
	*rushed.DB
	// fields and indexes
	schemaName shed.StringField
	size       shed.Uint64Field
	retrieval  shed.Index
	push       shed.Index
	pull       shed.Index
	gc         shed.Index
}

// NewDB constructs a local chunks db
func NewDB(path string) (*DB, error) {
	db := new(DB)
	sdb := shed.NewDB(path)
	db.DB = rushed.NewDB(sdb, db.update, db.access)
	db.schemaName, err = idb.NewStringField("schema-name")
	if err != nil {
		return nil, err
	}
	db.size, err = idb.NewUint64Field("size")
	if err != nil {
		return nil, err
	}
	db.retrieval, err = idb.NewIndex("Hash->StoredTimestamp|AccessTimestamp|Data", shed.IndexFuncs{
		EncodeKey: func(fields shed.IndexItem) (key []byte, err error) {
			return fields.Hash, nil
		},
		DecodeKey: func(key []byte) (e shed.IndexItem, err error) {
			e.Hash = key
			return e, nil
		},
		EncodeValue: func(fields shed.IndexItem) (value []byte, err error) {
			b := make([]byte, 16)
			binary.BigEndian.PutUint64(b[:8], uint64(fields.StoreTimestamp))
			binary.BigEndian.PutUint64(b[8:16], uint64(fields.AccessTimestamp))
			value = append(b, fields.Data...)
			return value, nil
		},
		DecodeValue: func(value []byte) (e shed.IndexItem, err error) {
			e.StoredTimestamp = int64(binary.BigEndian.Uint64(value[:8]))
			e.AccessTimestamp = int64(binary.BigEndian.Uint64(value[8:16]))
			e.Data = value[16:]
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}
	// pull index allows history and live syncing per po bin
	db.pull, err = idb.NewIndex("PO|StoredTimestamp|Hash->nil", shed.IndexFuncs{
		EncodeKey: func(fields shed.IndexItem) (key []byte, err error) {
			key = make([]byte, 41)
			key[0] = byte(uint8(db.po(fields.Hash)))
			binary.BigEndian.PutUint64(key[1:9], fields.StoredTimestamp)
			copy(key[9:], fields.Hash[:])
			return key, nil
		},
		DecodeKey: func(key []byte) (e shed.IndexItem, err error) {
			e.Hash = key[9:]
			e.StoredTimestamp = int64(binary.BigEndian.Uint64(key[1:9]))
			return e, nil
		},
		EncodeValue: func(fields shed.IndexItem) (value []byte, err error) {
			return nil, nil
		},
		DecodeValue: func(value []byte) (e shed.IndexItem, err error) {
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}
	// push index contains as yet unsynced chunks
	db.push, err = idb.NewIndex("StoredTimestamp|Hash->nil", shed.IndexFuncs{
		EncodeKey: func(fields shed.IndexItem) (key []byte, err error) {
			key = make([]byte, 40)
			binary.BigEndian.PutUint64(key[:8], fields.StoredTimestamp)
			copy(key[8:], fields.Hash[:])
			return key, nil
		},
		DecodeKey: func(key []byte) (e shed.IndexItem, err error) {
			e.Hash = key[8:]
			e.StoredTimestamp = int64(binary.BigEndian.Uint64(key[:8]))
			return e, nil
		},
		EncodeValue: func(fields shed.IndexItem) (value []byte, err error) {
			return nil, nil
		},
		DecodeValue: func(value []byte) (e shed.IndexItem, err error) {
			e.AccessTimestamp = int64(binary.BigEndian.Uint64(value))
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}
	// gc index for removable chunk ordered by ascending last access time
	db.gcIndex, err = idb.NewIndex("AccessTimestamp|StoredTimestamp|Hash->nil", shed.IndexFuncs{
		EncodeKey: func(fields shed.IndexItem) (key []byte, err error) {
			b := make([]byte, 16, 16+len(fields.Hash))
			binary.BigEndian.PutUint64(b[:8], uint64(fields.AccessTimestamp))
			binary.BigEndian.PutUint64(b[8:16], uint64(fields.StoreTimestamp))
			key = append(b, fields.Hash...)
			return key, nil
		},
		DecodeKey: func(key []byte) (e shed.IndexItem, err error) {
			e.AccessTimestamp = int64(binary.BigEndian.Uint64(key[:8]))
			e.StoreTimestamp = int64(binary.BigEndian.Uint64(key[8:16]))
			e.Hash = key[16:]
			return e, nil
		},
		EncodeValue: func(fields shed.IndexItem) (value []byte, err error) {
			return nil, nil
		},
		DecodeValue: func(value []byte) (e shed.IndexItem, err error) {
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// access defines get accessors for different modes
func (db *DB) access(b *leveldb.Batch, mode rushed.Mode, item *shed.IndexItem) error {
	err := db.retrieve.Get(item)
	switch mode {
	case SYNCING:
	case TESTSYNCING:
	case REQUEST:
		return db.Update(context.TODO(), REQUEST, item)
	default:
		return errInvalidMode
	}
	return nil
}

// update defines set accessors for different modes
func (db *DB) update(b *rushed.Batch, mode rushed.Mode, item *shed.IndexItem) error {
	switch mode {
	case SYNCING:
		// put to indexes: retrieve, pull
		item.StoredTimestamp = now()
		item.AccessTimestamp = now()
		db.retrieve.PutInBatch(b, item)
		db.pull.PutInBatch(b, item)
		db.size.IncInBatch(b)

	case UPLOAD:
		// put to indexes: retrieve, push, pull
		item.StoredTimestamp = now()
		item.AccessTimestamp = now()
		db.retrieve.PutInBatch(b, item)
		db.pull.PutInBatch(b, item)
		db.push.PutInBatch(b, item)

	case REQUEST:
		// put to indexes: retrieve, gc
		item.StoredTimestamp = now()
		item.AccessTimestamp = now()
		db.retrieve.PutInBatch(b, item)
		db.gc.PutInBatch(b, item)

	case SYNCED:
		// delete from push, insert to gc
		db.push.DeleteInBatch(b, item)
		db.gc.PutInBatch(b, item)

	case ACCESS:
		// update accessTimeStamp in retrieve, gc
		db.gc.DeleteInBatch(b, item)
		item.AccessTimestamp = now()
		db.retrieve.PutInBatch(b, item)
		db.gc.PutInBatch(b, item)

	case REMOVAL:
		// delete from retrieve, pull, gc
		db.retrieve.DeleteInBatch(b, item)
		db.pull.DeleteInBatch(b, item)
		db.gc.DeleteInBatch(b, item)

	default:
		return errInvalidMode
	}
	return nil
}

func now() uint64 {
	return uint64(time.Now().UnixNano())
}
