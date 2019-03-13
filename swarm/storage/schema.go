package storage

import (
	"fmt"

	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// The DB schema we want to use. The actual/current DB schema might differ
// until migrations are run.
const CurrentDbSchema = DbSchemaHalloween

// There was a time when we had no schema at all.
const DbSchemaNone = ""

// "purity" is the first formal schema of LevelDB we release together with Swarm 0.3.5
const DbSchemaPurity = "purity"

// "halloween" is here because we had a screw in the garbage collector index.
// Because of that we had to rebuild the GC index to get rid of erroneous
// entries and that takes a long time. This schema is used for bookkeeping,
// so rebuild index will run just once.
const DbSchemaHalloween = "halloween"

// returns true if legacy database is in the datadir
func IsLegacyDatabase(datadir string) bool {

	var (
		legacyDbSchemaKey = []byte{8}
		dbSchemaKey       = []byte{0}
	)

	db, err := leveldb.OpenFile(datadir, &opt.Options{OpenFilesCacheCapacity: 128})
	if err != nil {
		log.Error("error found", "err", err)
		return false
	}
	defer db.Close()

	data, err := db.Get(legacyDbSchemaKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {

			data, err := db.Get(dbSchemaKey, nil)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("getting some wtf")
			fmt.Println(string(data))

			return false
		}
	}

	fmt.Println(string(data))
	return string(data) == DbSchemaHalloween
}
