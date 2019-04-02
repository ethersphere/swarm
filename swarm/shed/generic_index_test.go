// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package shed

import (
	"errors"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// Index functions for the index that is used in tests in this file.
var retrievalGenericIndexFuncs = GenericIndexFuncs{
	EncodeKey: func(fields interface{}) (key []byte, err error) {
		//marshal the fields as something, return a byte array
		val, ok := fields.(string)
		if !ok {
			return nil, errors.New("could not unmarshal field")
		}
		return []byte(val), nil
	},
	DecodeKey: func(key []byte) (e interface{}, err error) {
		str := string(key)
		return str, nil
	},
	EncodeValue: func(fields interface{}) (value []byte, err error) {
		val, ok := fields.(string)
		if !ok {
			return nil, errors.New("could not unmarshal value")
		}
		return []byte(val), nil
	},
	DecodeValue: func(keyItem interface{}, value []byte) (e interface{}, err error) {
		str := string(value)
		return str, nil
	},
}

// TestIndex validates put, get, has and delete functions of the Index implementation.
func TestGenericIndex(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	index, err := db.NewGenericIndex("retrieval", retrievalGenericIndexFuncs)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("put", func(t *testing.T) {
		wantK := "wantKey"
		wantV := "wantVal"
		err := index.Put(wantK, wantV)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(wantK)
		if err != nil {
			t.Fatal(err)
		}
		checkItemString(t, got, wantV)

		t.Run("overwrite", func(t *testing.T) {
			wantK := "wantKey"
			wantV := "wantNewVal"
			err = index.Put(wantK, wantV)
			if err != nil {
				t.Fatal(err)
			}
			got, err := index.Get(wantK)
			if err != nil {
				t.Fatal(err)
			}
			checkItemString(t, got, wantV)
		})
	})

	t.Run("put in batch", func(t *testing.T) {
		wantK := "wantKey"
		wantV := "anotherNewVal"
		batch := new(leveldb.Batch)
		index.PutInBatch(batch, wantK, wantV)
		err := db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(wantK)
		if err != nil {
			t.Fatal(err)
		}
		checkItemString(t, got, wantV)

		t.Run("overwrite", func(t *testing.T) {
			wantK := "wantKey"
			wantV := "overrideBatchVal"
			batch := new(leveldb.Batch)
			index.PutInBatch(batch, wantK, wantV)
			db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			got, err := index.Get(wantK)
			if err != nil {
				t.Fatal(err)
			}
			checkItemString(t, got, wantV)
		})
	})

	t.Run("put in batch twice", func(t *testing.T) {
		// ensure that the last item of items with the same db keys
		// is actually saved
		batch := new(leveldb.Batch)
		wantK := "doubleWantKey"
		firstWantV := "should override"
		secondWantV := "should persist"

		// put the first item
		index.PutInBatch(batch, wantK, firstWantV)

		// then put the item that will produce the same key
		// but different value in the database
		index.PutInBatch(batch, wantK, secondWantV)
		db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(wantK)
		if err != nil {
			t.Fatal(err)
		}
		checkItemString(t, got, secondWantV)
	})

	t.Run("has", func(t *testing.T) {
		wantK := "wantHasThis"
		wantV := "shouldHaveThis"
		dontWantK := "dontWantHasThis"
		err := index.Put(wantK, wantV)
		if err != nil {
			t.Fatal(err)
		}

		has, err := index.Has(wantK)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Error("item is not found")
		}

		has, err = index.Has(dontWantK)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Error("unwanted item is found")
		}
	})

	t.Run("delete", func(t *testing.T) {
		wantK := "wantDelete"
		wantV := "wantDeleteVal"
		err := index.Put(wantK, wantV)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(wantK)
		if err != nil {
			t.Fatal(err)
		}
		checkItemString(t, got, wantV)

		err = index.Delete(wantK)
		if err != nil {
			t.Fatal(err)
		}

		wantErr := leveldb.ErrNotFound
		got, err = index.Get(wantK)
		if err != wantErr {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})

	t.Run("delete in batch", func(t *testing.T) {
		wantK := "wantDelInBatch"
		wantV := "wantDelInBatchVal"
		err := index.Put(wantK, wantV)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(wantK)
		if err != nil {
			t.Fatal(err)
		}
		checkItemString(t, got, wantV)

		batch := new(leveldb.Batch)
		index.DeleteInBatch(batch, wantK)
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}

		wantErr := leveldb.ErrNotFound
		got, err = index.Get(wantK)
		if err != wantErr {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}

// checkItemString is a test helper function that compares if two generic items are the same string.
func checkItemString(t *testing.T, got, want interface{}) {
	t.Helper()

	g := got.(string)
	w := want.(string)

	if g != w {
		t.Errorf("got %s, expected %s", g, w)
	}
}
