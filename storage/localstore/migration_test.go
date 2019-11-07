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
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

func TestOneMigration(t *testing.T) {
	DbSchemaCurrent = DbSchemaSanctuary
	defer func(v []migration) { allDbSchemaMigrations = v }(allDbSchemaMigrations)

	ran := false
	shouldNotRun := false
	allDbSchemaMigrations = []migration{
		{name: DbSchemaSanctuary, migrationFunc: func(db *DB) error {
			ran = true
			return nil
		}},
		{name: DbSchemaDiwali, migrationFunc: func(db *DB) error {
			shouldNotRun = true // this should not be executed
			return nil
		}},
	}

	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	// start the fresh localstore with the sanctuary schema name
	db, err := New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	DbSchemaCurrent = DbSchemaDiwali

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	schemaName, err := db.schemaName.Get()
	if err != nil {
		t.Fatal(err)
	}

	if schemaName != DbSchemaDiwali {
		t.Errorf("schema name mismatch. got '%s', want '%s'", schemaName, DbSchemaDiwali)
	}

	if !ran {
		t.Errorf("expected migration did not run")
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}

	err = db.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestManyMigrations(t *testing.T) {
	DbSchemaCurrent = DbSchemaSanctuary
	defer func(v []migration) { allDbSchemaMigrations = v }(allDbSchemaMigrations)

	shouldNotRun := false
	executionOrder := make([]int, 5)
	allDbSchemaMigrations = []migration{
		{name: DbSchemaSanctuary, migrationFunc: func(db *DB) error {
			executionOrder[0] = 0
			return nil
		}},
		{name: DbSchemaDiwali, migrationFunc: func(db *DB) error {
			executionOrder[1] = 1
			return nil
		}},
		{name: "coconut", migrationFunc: func(db *DB) error {
			executionOrder[2] = 2
			return nil
		}},
		{name: "mango", migrationFunc: func(db *DB) error {
			executionOrder[3] = 3
			return nil
		}},
		{name: "salvation", migrationFunc: func(db *DB) error {
			executionOrder[4] = 4
			shouldNotRun = true // this should not be executed
			return nil
		}},
	}

	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	// start the fresh localstore with the sanctuary schema name
	db, err := New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	DbSchemaCurrent = "salvation"

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	schemaName, err := db.schemaName.Get()
	if err != nil {
		t.Fatal(err)
	}

	if schemaName != "salvation" {
		t.Errorf("schema name mismatch. got '%s', want '%s'", schemaName, "salvation")
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}

	for i, v := range executionOrder {
		if i != v && i != len(executionOrder)-1 {
			t.Errorf("migration did not run in sequence, slot %d value %d", i, v)
		}
	}

	err = db.Close()
	if err != nil {
		t.Error(err)
	}
}

// TestMigrationFailFrom checks that local store boot should fail when the schema we're migrating from cannot be found
func TestMigrationFailFrom(t *testing.T) {
	DbSchemaCurrent = "koo-koo-schema"
	defer func(v []migration) { allDbSchemaMigrations = v }(allDbSchemaMigrations)

	shouldNotRun := false
	allDbSchemaMigrations = []migration{
		{name: "langur", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "coconut", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "chutney", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
	}

	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	// start the fresh localstore with the sanctuary schema name
	db, err := New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	DbSchemaCurrent = "foo"

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if err != errMissingCurrentSchema {
		t.Fatalf("expected errCannotFindSchema but got %v", err)
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}
}

// TestMigrationFailTo checks that local store boot should fail when the schema we're migrating to cannot be found
func TestMigrationFailTo(t *testing.T) {
	DbSchemaCurrent = "langur"
	defer func(v []migration) { allDbSchemaMigrations = v }(allDbSchemaMigrations)

	shouldNotRun := false
	allDbSchemaMigrations = []migration{
		{name: "langur", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "coconut", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "chutney", migrationFunc: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
	}

	dir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	// start the fresh localstore with the sanctuary schema name
	db, err := New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	DbSchemaCurrent = "foo"

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if err != errMissingTargetSchema {
		t.Fatalf("expected errMissingTargetSchema but got %v", err)
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}
}
