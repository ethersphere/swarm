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
	"errors"

	"github.com/ethersphere/swarm/log"
)

var errMissingCurrentSchema = errors.New("could not find current db schema")
var errMissingTargetSchema = errors.New("could not find target db schema")

func (db *DB) migrate(schemaName string) error {
	migrations, err := getMigrations(schemaName, DbSchemaCurrent, allDbSchemaMigrations)
	if err != nil {
		log.Error("error getting data migrations for current localstore version", "schemaName", schemaName)
		return err
	}

	if migrations != nil {
		log.Info("need to run data migrations on localstore", "numMigrations", len(migrations), "schemaName", schemaName)
		for i := 0; i < len(migrations)-1; i++ {
			err := migrations[i].migrationFunc(db)
			if err != nil {
				return err
			}
			if i != len(migrations)-1 {
				err = db.schemaName.Put(migrations[i+1].name) // put the name of the next schema
				if err != nil {
					return err
				}
			}
			schemaName, err = db.schemaName.Get()
			if err != nil {
				return err
			}
			log.Info("successfully ran migration", "migrationId", i, "currentSchema", schemaName)
		}
	}
	return nil
}

// migrationFn is a function that takes a localstore.DB and
// returns an error if a migration has failed
type migrationFn func(db *DB) error

// getMigrations returns an ordered list of migrations that need be executed
// with no errors in order to bring the localstore to the most up-to-date
// schema definition
func getMigrations(currentSchema, targetSchema string, allSchemeMigrations []migration) (migrations []migration, err error) {
	foundCurrent := false
	foundTarget := false
	if currentSchema == DbSchemaCurrent {
		return nil, nil
	}
	for i, v := range allSchemeMigrations {
		switch v.name {
		case currentSchema:
			if foundCurrent {
				return nil, errors.New("found schema name for the second time when looking for migrations")
			}
			foundCurrent = true
			log.Info("found current localstore schema", "currentSchema", currentSchema, "migrateTo", DbSchemaCurrent, "total migrations", len(allSchemeMigrations)-i-1)
		case targetSchema:
			foundTarget = true
		}
		if foundCurrent {
			migrations = append(migrations, v)
		}
	}
	if !foundCurrent {
		return nil, errMissingCurrentSchema
	}
	if !foundTarget {
		return nil, errMissingTargetSchema
	}
	return migrations, nil
}

// this function migrates Sanctuary schema to the Diwali schema
func migrateSanctuary(db *DB) error {
	// just rename the pull index
	renamed, err := db.shed.RenameIndex("PO|BinID->Hash", "PO|BinID->Hash|Tag")
	if err != nil {
		return err
	}
	if !renamed {
		return errors.New("pull index was not successfully renamed!")
	}

	return nil
}
