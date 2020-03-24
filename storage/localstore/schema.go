// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package localstore

import (
	"github.com/ethersphere/swarm/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// dbSchemaCurrent is the schema name of the current implementation.
// The actual/current DB schema might differ until migrations are run.
var dbSchemaCurrent = dbSchemaForky

const (
	// dbSchemaSanctuary is the first storage/localstore schema.
	dbSchemaSanctuary = "sanctuary"
	// dbSchemaDiwali migration simply renames the pullIndex in localstore.
	dbSchemaDiwali = "diwali"
	// dbSchemaForky migration implements FCDS storage and requires manual import and export.
	dbSchemaForky = "forky"
)

// IsLegacyDatabase returns true if legacy database is in the data directory.
func IsLegacyDatabase(datadir string) bool {

	// "purity" is the first formal schema of LevelDB we release together with Swarm 0.3.5
	const dbSchemaPurity = "purity"

	// "halloween" is here because we had a screw in the garbage collector index.
	// Because of that we had to rebuild the GC index to get rid of erroneous
	// entries and that takes a long time. This schema is used for bookkeeping,
	// so rebuild index will run just once.
	const dbSchemaHalloween = "halloween"

	var legacyDBSchemaKey = []byte{8}

	db, err := leveldb.OpenFile(datadir, &opt.Options{OpenFilesCacheCapacity: 128})
	if err != nil {
		log.Error("open leveldb", "path", datadir, "err", err)
		return false
	}
	defer db.Close()

	data, err := db.Get(legacyDBSchemaKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			// if we haven't found anything under the legacy db schema key- we are not on legacy
			return false
		}

		log.Error("get legacy name from", "err", err)
	}
	schema := string(data)
	log.Trace("checking if database scheme is legacy", "schema name", schema)
	return schema == dbSchemaHalloween || schema == dbSchemaPurity
}
