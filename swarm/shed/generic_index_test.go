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
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// Index functions for the index that is used in tests in this file.
var retrievalIndexFuncs = GenericIndexFuncs{
	EncodeKey: func(fields interface{}) (key []byte, err error) {
		return nil, nil
	},
	DecodeKey: func(key []byte) (e interface{}, err error) {
		return e, nil
	},
	EncodeValue: func(fields interface{}) (value []byte, err error) {
		return value, nil
	},
	DecodeValue: func(keyItem interface{}, value []byte) (e interface{}, err error) {
		return nil, nil
	},
}

// TestIndex validates put, get, has and delete functions of the Index implementation.
func TestGenericIndex(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	index, err := db.NewGenericIndex("retrieval", retrievalIndexFuncs)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("put", func(t *testing.T) {
		want := //
		err := index.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(...)
		if err != nil {
			t.Fatal(err)
		}
		//checkItem(t, got, want)

		t.Run("overwrite", func(t *testing.T) {
			want := ...
			err = index.Put(want)
			if err != nil {
				t.Fatal(err)
			}
			got, err := index.Get(...)
			if err != nil {
				t.Fatal(err)
			}
			checkItem(t, got, want)
		})
	})

	t.Run("put in batch", func(t *testing.T) {
		want :=...
		batch := new(leveldb.Batch)
		index.PutInBatch(batch, want)
		err := db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(...)
		if err != nil {
			t.Fatal(err)
		}
		checkItem(t, got, want)

		t.Run("overwrite", func(t *testing.T) {
			want :=...
			batch := new(leveldb.Batch)
			index.PutInBatch(batch, want)
			db.WriteBatch(batch)
			if err != nil {
				t.Fatal(err)
			}
			got, err := index.Get(...)
			if err != nil {
				t.Fatal(err)
			}
			checkItem(t, got, want)
		})
	})

	t.Run("put in batch twice", func(t *testing.T) {
		// ensure that the last item of items with the same db keys
		// is actually saved
		batch := new(leveldb.Batch)
		address := []byte("put-in-batch-twice-hash")

		// put the first item
		index.PutInBatch(batch, Item{
			Address:        address,
			Data:           []byte("DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		})

		want := Item{
			Address:        address,
			Data:           []byte("New DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		}
		// then put the item that will produce the same key
		// but different value in the database
		index.PutInBatch(batch, want)
		db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(Item{
			Address: address,
		})
		if err != nil {
			t.Fatal(err)
		}
		checkItem(t, got, want)
	})

	t.Run("has", func(t *testing.T) {
		want := Item{
			Address:        []byte("has-hash"),
			Data:           []byte("DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		}

		dontWant := Item{
			Address:        []byte("do-not-has-hash"),
			Data:           []byte("DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		}

		err := index.Put(want)
		if err != nil {
			t.Fatal(err)
		}

		has, err := index.Has(want)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Error("item is not found")
		}

		has, err = index.Has(dontWant)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Error("unwanted item is found")
		}
	})

	t.Run("delete", func(t *testing.T) {
		want := Item{
			Address:        []byte("delete-hash"),
			Data:           []byte("DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		}

		err := index.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(Item{
			Address: want.Address,
		})
		if err != nil {
			t.Fatal(err)
		}
		checkItem(t, got, want)

		err = index.Delete(Item{
			Address: want.Address,
		})
		if err != nil {
			t.Fatal(err)
		}

		wantErr := leveldb.ErrNotFound
		got, err = index.Get(Item{
			Address: want.Address,
		})
		if err != wantErr {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})

	t.Run("delete in batch", func(t *testing.T) {
		want := Item{
			Address:        []byte("delete-in-batch-hash"),
			Data:           []byte("DATA"),
			StoreTimestamp: time.Now().UTC().UnixNano(),
		}

		err := index.Put(want)
		if err != nil {
			t.Fatal(err)
		}
		got, err := index.Get(Item{
			Address: want.Address,
		})
		if err != nil {
			t.Fatal(err)
		}
		checkItem(t, got, want)

		batch := new(leveldb.Batch)
		index.DeleteInBatch(batch, Item{
			Address: want.Address,
		})
		err = db.WriteBatch(batch)
		if err != nil {
			t.Fatal(err)
		}

		wantErr := leveldb.ErrNotFound
		got, err = index.Get(Item{
			Address: want.Address,
		})
		if err != wantErr {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}

// checkItem is a test helper function that compares if two Index items are the same.
/*func checkItem(t *testing.T, got, want Item) {
	t.Helper()

	if !bytes.Equal(got.Address, want.Address) {
		t.Errorf("got hash %q, expected %q", string(got.Address), string(want.Address))
	}
	if !bytes.Equal(got.Data, want.Data) {
		t.Errorf("got data %q, expected %q", string(got.Data), string(want.Data))
	}
	if got.StoreTimestamp != want.StoreTimestamp {
		t.Errorf("got store timestamp %v, expected %v", got.StoreTimestamp, want.StoreTimestamp)
	}
	if got.AccessTimestamp != want.AccessTimestamp {
		t.Errorf("got access timestamp %v, expected %v", got.AccessTimestamp, want.AccessTimestamp)
	}
}*/
