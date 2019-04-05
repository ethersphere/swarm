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
	"bytes"
	"errors"
	"fmt"
	"sort"
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

// TestGenericIndex_Iterate validates index Iterate
// functions for correctness.
func TestGenericIndex_Iterate(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	index, err := db.NewGenericIndex("retrieval", retrievalGenericIndexFuncs)
	if err != nil {
		t.Fatal(err)
	}

	items := []struct {
		k string
		v string
	}{
		{
			k: "iterate-hash-01",
			v: "data80",
		},
		{
			k: "iterate-hash-03",
			v: "data22",
		},
		{
			k: "iterate-hash-05",
			v: "data41",
		},
		{
			k: "iterate-hash-02",
			v: "data84",
		},
		{
			k: "iterate-hash-06",
			v: "data1",
		},
	}
	batch := new(leveldb.Batch)
	for _, i := range items {
		index.PutInBatch(batch, i.k, i.v)
	}
	err = db.WriteBatch(batch)
	if err != nil {
		t.Fatal(err)
	}
	k := "iterate-hash-04"
	v := "data0"
	err = index.Put(k, v)
	if err != nil {
		t.Fatal(err)
	}
	items = append(items, struct {
		k string
		v string
	}{k: k,
		v: v,
	})

	sort.SliceStable(items, func(i, j int) bool {
		return bytes.Compare([]byte(items[i].k), []byte(items[j].k)) < 0
	})
	t.Run("all", func(t *testing.T) {
		var i int
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("start from", func(t *testing.T) {
		startIndex := 2
		i := startIndex
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, &GenericIterateOptions{
			StartFrom: items[startIndex].k,
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("skip start from", func(t *testing.T) {
		startIndex := 2
		i := startIndex + 1
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, &GenericIterateOptions{
			StartFrom:         items[startIndex].k,
			SkipStartFromItem: true,
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("stop", func(t *testing.T) {
		var i int
		stopIndex := 3
		var count int
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			count++
			if i == stopIndex {
				return true, nil
			}
			i++
			return false, nil
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		wantItemsCount := stopIndex + 1
		if count != wantItemsCount {
			t.Errorf("got %v items, expected %v", count, wantItemsCount)
		}
	})

	t.Run("no overflow", func(t *testing.T) {
		secondIndex, err := db.NewGenericIndex("second-index", retrievalGenericIndexFuncs)
		if err != nil {
			t.Fatal(err)
		}

		secondItem := struct {
			k string
			v string
		}{
			k: "iterate-hash-10",
			v: "data-second",
		}
		err = secondIndex.Put(secondItem.k, secondItem.v)
		if err != nil {
			t.Fatal(err)
		}

		var i int
		err = index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, nil)
		if err != nil {
			t.Fatal(err)
		}

		i = 0
		err = secondIndex.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > 1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			checkItemString(t, v, secondItem.v)
			i++
			return false, nil
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}

// TestIndex_Iterate_withPrefix validates index Iterate
// function for correctness.
func TestGenericIndex_Iterate_withPrefix(t *testing.T) {
	db, cleanupFunc := newTestDB(t)
	defer cleanupFunc()

	index, err := db.NewGenericIndex("retrieval", retrievalGenericIndexFuncs)
	if err != nil {
		t.Fatal(err)
	}

	allItems := []struct{ k, v string }{
		{k: "want-hash-00", v: "data80"},
		{k: "skip-hash-01", v: "data81"},
		{k: "skip-hash-02", v: "data82"},
		{k: "skip-hash-03", v: "data83"},
		{k: "want-hash-04", v: "data84"},
		{k: "want-hash-05", v: "data85"},
		{k: "want-hash-06", v: "data86"},
		{k: "want-hash-07", v: "data87"},
		{k: "want-hash-08", v: "data88"},
		{k: "want-hash-09", v: "data89"},
		{k: "skip-hash-10", v: "data90"},
	}
	batch := new(leveldb.Batch)
	for _, i := range allItems {
		index.PutInBatch(batch, i.k, i.v)
	}
	err = db.WriteBatch(batch)
	if err != nil {
		t.Fatal(err)
	}

	prefix := []byte("want")

	items := make([]struct{ k, v string }, 0)
	for _, item := range allItems {
		if bytes.HasPrefix([]byte(item.k), prefix) {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return bytes.Compare([]byte(items[i].k), []byte(items[j].k)) < 0
	})

	t.Run("with prefix", func(t *testing.T) {
		var i int
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, &GenericIterateOptions{
			Prefix: prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if i != len(items) {
			t.Errorf("got %v items, want %v", i, len(items))
		}
	})

	t.Run("with prefix and start from", func(t *testing.T) {
		startIndex := 2
		var count int
		i := startIndex
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			count++
			return false, nil
		}, &GenericIterateOptions{
			StartFrom: items[startIndex].k,
			Prefix:    prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		wantCount := len(items) - startIndex
		if count != wantCount {
			t.Errorf("got %v items, want %v", count, wantCount)
		}
	})

	t.Run("with prefix and skip start from", func(t *testing.T) {
		startIndex := 2
		var count int
		i := startIndex + 1
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			count++
			return false, nil
		}, &GenericIterateOptions{
			StartFrom:         items[startIndex].k,
			SkipStartFromItem: true,
			Prefix:            prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		wantCount := len(items) - startIndex - 1
		if count != wantCount {
			t.Errorf("got %v items, want %v", count, wantCount)
		}
	})

	t.Run("stop", func(t *testing.T) {
		var i int
		stopIndex := 3
		var count int
		err := index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			count++
			if i == stopIndex {
				return true, nil
			}
			i++
			return false, nil
		}, &GenericIterateOptions{
			Prefix: prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		wantItemsCount := stopIndex + 1
		if count != wantItemsCount {
			t.Errorf("got %v items, expected %v", count, wantItemsCount)
		}
	})

	t.Run("no overflow", func(t *testing.T) {
		secondIndex, err := db.NewGenericIndex("second-index", retrievalGenericIndexFuncs)
		if err != nil {
			t.Fatal(err)
		}

		secondItem := struct{ k, v string }{
			k: "iterate-hash-10",
			v: "data-second",
		}
		err = secondIndex.Put(secondItem.k, secondItem.v)
		if err != nil {
			t.Fatal(err)
		}

		var i int
		err = index.Iterate(func(k, v interface{}) (stop bool, err error) {
			if i > len(items)-1 {
				return true, fmt.Errorf("got unexpected index item: %#v", v)
			}
			want := items[i].v
			checkItemString(t, v, want)
			i++
			return false, nil
		}, &GenericIterateOptions{
			Prefix: prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if i != len(items) {
			t.Errorf("got %v items, want %v", i, len(items))
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
