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
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/ethersphere/swarm/chunk"
)

func TestOneMigration(t *testing.T) {
	defer func(v []migration, s string) {
		schemaMigrations = v
		dbSchemaCurrent = s
	}(schemaMigrations, dbSchemaCurrent)

	dbSchemaCurrent = dbSchemaSanctuary

	ran := false
	shouldNotRun := false
	schemaMigrations = []migration{
		{name: dbSchemaSanctuary, fn: func(db *DB) error {
			shouldNotRun = true // this should not be executed
			return nil
		}},
		{name: dbSchemaDiwali, fn: func(db *DB) error {
			ran = true
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

	dbSchemaCurrent = dbSchemaDiwali

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	schemaName, err := db.schemaName.Get()
	if err != nil {
		t.Fatal(err)
	}

	if schemaName != dbSchemaDiwali {
		t.Errorf("schema name mismatch. got '%s', want '%s'", schemaName, dbSchemaDiwali)
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
	defer func(v []migration, s string) {
		schemaMigrations = v
		dbSchemaCurrent = s
	}(schemaMigrations, dbSchemaCurrent)

	dbSchemaCurrent = dbSchemaSanctuary

	shouldNotRun := false
	executionOrder := []int{-1, -1, -1, -1}

	schemaMigrations = []migration{
		{name: dbSchemaSanctuary, fn: func(db *DB) error {
			shouldNotRun = true // this should not be executed
			return nil
		}},
		{name: dbSchemaDiwali, fn: func(db *DB) error {
			executionOrder[0] = 0
			return nil
		}},
		{name: "coconut", fn: func(db *DB) error {
			executionOrder[1] = 1
			return nil
		}},
		{name: "mango", fn: func(db *DB) error {
			executionOrder[2] = 2
			return nil
		}},
		{name: "salvation", fn: func(db *DB) error {
			executionOrder[3] = 3
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

	dbSchemaCurrent = "salvation"

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

// TestGetMigrations validates the migration selection based on
// current and target schema names.
func TestGetMigrations(t *testing.T) {
	currentSchema := "current"
	defaultTargetSchema := "target"

	for _, tc := range []struct {
		name           string
		targetSchema   string
		migrations     []migration
		wantMigrations []migration
	}{
		{
			name:         "empty",
			targetSchema: "current",
			migrations: []migration{
				{name: "current"},
			},
		},
		{
			name: "single",
			migrations: []migration{
				{name: "current"},
				{name: "target"},
			},
			wantMigrations: []migration{
				{name: "target"},
			},
		},
		{
			name: "multiple",
			migrations: []migration{
				{name: "current"},
				{name: "middle"},
				{name: "target"},
			},
			wantMigrations: []migration{
				{name: "middle"},
				{name: "target"},
			},
		},
		{
			name: "between",
			migrations: []migration{
				{name: "current"},
				{name: "target"},
				{name: "future"},
			},
			wantMigrations: []migration{
				{name: "target"},
			},
		},
		{
			name: "between multiple",
			migrations: []migration{
				{name: "current"},
				{name: "middle"},
				{name: "target"},
				{name: "future"},
			},
			wantMigrations: []migration{
				{name: "middle"},
				{name: "target"},
			},
		},
		{
			name: "with previous",
			migrations: []migration{
				{name: "previous"},
				{name: "current"},
				{name: "target"},
			},
			wantMigrations: []migration{
				{name: "target"},
			},
		},
		{
			name: "with previous multiple",
			migrations: []migration{
				{name: "previous"},
				{name: "current"},
				{name: "middle"},
				{name: "target"},
			},
			wantMigrations: []migration{
				{name: "middle"},
				{name: "target"},
			},
		},
		{
			name: "breaking",
			migrations: []migration{
				{name: "current"},
				{name: "target", breaking: true},
			},
			wantMigrations: []migration{
				{name: "target", breaking: true},
			},
		},
		{
			name: "breaking multiple",
			migrations: []migration{
				{name: "current"},
				{name: "middle"},
				{name: "breaking", breaking: true},
				{name: "target"},
			},
			wantMigrations: []migration{
				{name: "breaking", breaking: true},
				{name: "target"},
			},
		},
		{
			name: "breaking with previous",
			migrations: []migration{
				{name: "previous"},
				{name: "current"},
				{name: "target", breaking: true},
			},
			wantMigrations: []migration{
				{name: "target", breaking: true},
			},
		},
		{
			name: "breaking multiple breaks",
			migrations: []migration{
				{name: "current"},
				{name: "middle", breaking: true},
				{name: "target", breaking: true},
			},
			wantMigrations: []migration{
				{name: "target", breaking: true},
			},
		},
		{
			name: "breaking multiple with middle",
			migrations: []migration{
				{name: "current"},
				{name: "breaking", breaking: true},
				{name: "middle"},
				{name: "target", breaking: true},
			},
			wantMigrations: []migration{
				{name: "target", breaking: true},
			},
		},
		{
			name: "breaking multiple between",
			migrations: []migration{
				{name: "current"},
				{name: "breaking", breaking: true},
				{name: "middle"},
				{name: "target", breaking: true},
				{name: "future"},
			},
			wantMigrations: []migration{
				{name: "target", breaking: true},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			targetSchema := tc.targetSchema
			if targetSchema == "" {
				targetSchema = defaultTargetSchema
			}
			got, err := getMigrations(
				currentSchema,
				targetSchema,
				tc.migrations,
			)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.wantMigrations) {
				t.Errorf("got migrations %v, want %v", got, tc.wantMigrations)
			}
		})
	}
}

// TestMigrationFailFrom checks that local store boot should fail when the schema we're migrating from cannot be found
func TestMigrationFailFrom(t *testing.T) {
	defer func(v []migration, s string) {
		schemaMigrations = v
		dbSchemaCurrent = s
	}(schemaMigrations, dbSchemaCurrent)

	dbSchemaCurrent = "koo-koo-schema"

	shouldNotRun := false
	schemaMigrations = []migration{
		{name: "langur", fn: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "coconut", fn: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "chutney", fn: func(db *DB) error {
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

	dbSchemaCurrent = "foo"

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if !strings.Contains(err.Error(), errMissingCurrentSchema.Error()) {
		t.Fatalf("expected errCannotFindSchema but got %v", err)
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}
}

// TestMigrationFailTo checks that local store boot should fail when the schema we're migrating to cannot be found
func TestMigrationFailTo(t *testing.T) {
	defer func(v []migration, s string) {
		schemaMigrations = v
		dbSchemaCurrent = s
	}(schemaMigrations, dbSchemaCurrent)

	dbSchemaCurrent = "langur"

	shouldNotRun := false
	schemaMigrations = []migration{
		{name: "langur", fn: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "coconut", fn: func(db *DB) error {
			shouldNotRun = true
			return nil
		}},
		{name: "chutney", fn: func(db *DB) error {
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

	dbSchemaCurrent = "foo"

	// start the existing localstore and expect the migration to run
	db, err = New(dir, baseKey, nil)
	if !strings.Contains(err.Error(), errMissingTargetSchema.Error()) {
		t.Fatalf("expected errMissingTargetSchema but got %v", err)
	}

	if shouldNotRun {
		t.Errorf("migration ran but shouldnt have")
	}
}

// TestMigrateSanctuaryFixture migrates an actual Sanctuary localstore
// to the most recent schema.
func TestMigrateSanctuaryFixture(t *testing.T) {

	tmpdir, err := ioutil.TempDir("", "localstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	dir := path.Join(".", "testdata", "sanctuary")
	if err != nil {
		t.Fatal(err)
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		err = copyFileContents(path.Join(dir, f.Name()), path.Join(tmpdir, f.Name()))
		if err != nil {
			t.Fatal(err)
		}
	}

	baseKey := make([]byte, 32)
	if _, err := rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	// start localstore with the copied fixture
	db, err := New(tmpdir, baseKey, &Options{Tags: chunk.NewTags()})
	if err != nil {
		t.Fatal(err)
	}
	schemaName, err := db.schemaName.Get()
	if err != nil {
		t.Fatal(err)
	}

	if schemaName != dbSchemaCurrent {
		t.Fatalf("schema name mismatch, want '%s' got '%s'", dbSchemaCurrent, schemaName)
	}

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
