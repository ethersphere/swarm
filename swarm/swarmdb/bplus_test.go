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

package swarmdb_test

import (
	"bytes"
	"fmt"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	//sdbc "swarmdbcommon"
	"math/rand"
	"os"
	wolkdb "swarmdb"
	"testing"
)

const (
	TEST_ENCRYPTED = 1
)

var config *wolkdb.SWARMDBConfig
var swarmdb *wolkdb.SwarmDB

func TestMain(m *testing.M) {
	config, _ = wolkdb.LoadSWARMDBConfig(wolkdb.SWARMDBCONF_FILE)
	var err error
	swarmdb, err = wolkdb.NewSwarmDB(config)
	if err != nil {
		os.Exit(0) // m.Fatal("could not create SWARMDB", err)
	}
	wolkdb.NewKeyManager(config)
	code := m.Run()
	// do somethng in shutdown
	os.Exit(code)
}

func TestPutInteger(t *testing.T) {

	u := config.GetSWARMDBUser()

	fmt.Printf("---- TestPutInteger: generate 20 ints and enumerate them\n")
	hashid := make([]byte, 32)
	r, errB := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)
	if errB != nil {
		t.Fatal("could not create BplusTree", errB)
	}
	// write 20 values into B-tree (only kept in memory)
	r.StartBuffer(u)
	vals := rand.Perm(20)
	for _, i := range vals {
		k := wolkdb.IntToByte(i)
		v := []byte(fmt.Sprintf("valueof%06x", i))
		fmt.Printf("Insert [%d] - [%x] [%v]\n", i, string(k), string(v))
		_, errP := r.Put(u, k, v)
		if errP != nil {
			t.Fatal("failure to Put", k)
		}
	}

	r.Print(u)
	// flush B+tree in memory to SWARM
	_, errF := r.FlushBuffer(u)
	if errF != nil {
		t.Fatal("fail on FlushBuffer", errF)
	}

	hashid = r.GetRootHash()
	s, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	g, ok, err := s.Get(u, wolkdb.IntToByte(8))
	if !ok || err != nil {
		t.Fatal(g, err)
	} else if bytes.Contains(g, []byte("valueof000008")) {
		fmt.Printf("SUCC Get(8): [%s]\n", string(g))
	} else {
		fmt.Printf("FAIL Get(8): [%x] [%x]\n", g, []byte("valueof000008"))
		t.Fatal("Get(8) failure", g, "valueof000008")
	}

	g, ok, err = s.Get(u, wolkdb.IntToByte(1))
	if !ok || err != nil {
		t.Fatal("Get(1) not ok", err)
	} else if bytes.Contains(g, []byte("valueof000001")) {
		fmt.Printf("SUCC Get(1): [%s]\n", string(g))
	} else {
		t.Fatal("Get(1) failure")
	}

	g, ok, err = s.Get(u, wolkdb.IntToByte(12))
	if !ok || err != nil {
		t.Fatal("Get(12) not ok", err)
	} else if bytes.Contains(g, []byte("valueof00000c")) {
		fmt.Printf("SUCC Get(12): [%s]\n", string(g))
	} else {
		t.Fatal("Get(12) failure")
	}

	g, ok, err = s.Get(u, wolkdb.IntToByte(16))
	if !ok || err != nil {
		t.Fatal("Get(16) not ok", err)
	} else if bytes.Contains(g, []byte("valueof000010")) {
		fmt.Printf("SUCC Get(16): [%s]\n", string(g))
	} else {
		t.Fatal("Get(16) failure")
	}
	s.Print(u)

	// ENUMERATOR
	if false {
		res, _ := s.SeekFirst(u)
		records := 0
		for k, v, err := res.Next(u); err == nil; k, v, err = res.Next(u) {
			fmt.Printf(" *int*> %d: K: %s V: %v\n", records, wolkdb.KeyToString(sdbc.CT_INTEGER, k), string(v))
			records++
		}
		fmt.Printf("---- TestPutInteger Next (%d records)\n", records)
	}

	// ENUMERATOR
	if true {
		res, _ := s.SeekLast(u)
		records := 0
		for k, v, err := res.Prev(u); err == nil; k, v, err = res.Prev(u) {
			fmt.Printf(" *int*> %d: K: %s V: %v\n", records, wolkdb.KeyToString(sdbc.CT_INTEGER, k), string(v))
			records++
		}
		fmt.Printf("---- TestPutInteger Prev (%d records)\n", records)
	}
}

func TestPutString(t *testing.T) {
	fmt.Printf("---- TestPutString: generate 20 strings and enumerate them\n")
	u := config.GetSWARMDBUser()

	hashid := make([]byte, 32)
	r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_STRING, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	r.StartBuffer(u)
	vals := rand.Perm(20)
	// write 20 values into B-tree (only kept in memory)
	for _, i := range vals {
		k := []byte(fmt.Sprintf("%06x", i))
		v := []byte(fmt.Sprintf("valueof%06x", i))
		// fmt.Printf("Insert %d %v %v\n", i, string(k), string(v))
		r.Put(u, k, v)
	}
	// this writes B+tree to SWARM
	r.FlushBuffer(u)
	// r.Print()

	hashid = r.GetRootHash()
	s, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_STRING, false, sdbc.CT_STRING, TEST_ENCRYPTED)
	g, _, _ := s.Get(u, []byte("000008"))
	fmt.Printf("Get(000008): %v\n", string(g))

	h, _, _ := s.Get(u, []byte("000001"))
	fmt.Printf("Get(000001): %v\n", string(h))
	// s.Print()

	// ENUMERATOR
	res, _, _ := r.Seek(u, []byte("000004"))
	records := 0
	for k, v, err := res.Next(u); err == nil; k, v, err = res.Next(u) {
		fmt.Printf(" *string*> %d K: %s V: %v\n", records, wolkdb.KeyToString(sdbc.CT_STRING, k), string(v))
		records++
	}
	fmt.Printf("---- TestPutString DONE (%d records)\n", records)
}

func TestPutFloat(t *testing.T) {
	fmt.Printf("---- TestPutFloat: generate 20 floats and enumerate them\n")
	u := config.GetSWARMDBUser()

	hashid := make([]byte, 32)
	r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_FLOAT, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	r.StartBuffer(u)
	vals := rand.Perm(20)
	// write 20 values into B-tree (only kept in memory)
	for _, i := range vals {
		k := wolkdb.FloatToByte(float64(i) + .314159)
		v := []byte(fmt.Sprintf("valueof%06x", i))
		// fmt.Printf("Insert %d %v %v\n", i, wolkdb.KeyToString(sdbc.CT_FLOAT, k), string(v))
		r.Put(u, k, v)
	}
	// this writes B+tree to SWARM
	r.FlushBuffer(u)
	// r.Print()

	hashid = r.GetRootHash()
	s, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_FLOAT, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	// ENUMERATOR
	res, _, _ := s.Seek(u, wolkdb.FloatToByte(3.14159))
	records := 0
	for k, v, err := res.Next(u); err == nil; k, v, err = res.Next(u) {
		fmt.Printf(" *float*> %d: K: %s V: %v\n", records, wolkdb.KeyToString(sdbc.CT_FLOAT, k), string(v))
		records++
	}
}

func TestSetGetString(t *testing.T) {
	u := config.GetSWARMDBUser()

	hashid := make([]byte, 32)
	r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_STRING, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	// put
	key := []byte("42")
	val := wolkdb.SHA256("314")
	r.Put(u, key, val)

	// check put with get
	g, ok, err := r.Get(u, key)
	if !ok || err != nil {
		t.Fatal(ok)
	}
	if bytes.Compare(g, val) != 0 {
		t.Fatal(g, val)
	}
	//r.Print()
	hashid = r.GetRootHash()

	// r2 put
	r2, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_STRING, false, sdbc.CT_STRING, TEST_ENCRYPTED)
	val2 := wolkdb.SHA256("278")
	r2.Put(u, key, val2)
	//r2.Print()

	// check put with get
	g2, ok, err := r2.Get(u, key)
	if !ok || err != nil {
		t.Fatal(ok)
	}
	if bytes.Compare(g2, val2) != 0 {
		t.Fatal(g2, val2)
	}
	hashid = r2.GetRootHash()

	// r3 put
	r3, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_STRING, false, sdbc.CT_STRING, TEST_ENCRYPTED)
	key2 := []byte("420")
	val3 := wolkdb.SHA256("bbb")
	r3.Put(u, key2, val3)

	// check put with get
	g3, ok, err := r3.Get(u, key2)
	//r3.Print()
	if !ok || err != nil {
		t.Fatal(ok)
	}
	if bytes.Compare(g3, val3) != 0 {
		t.Fatal(g3, val3)
	}
	fmt.Printf("PASS\n")

}

func TestSetGetInt(t *testing.T) {
	u := config.GetSWARMDBUser()

	const N = 4
	hashid := make([]byte, 32)
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x333333, 0xcccccc, 0x314159} {
		r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)

		a := make([]int, N)
		for i := range a {
			a[i] = (i ^ x) << 1
		}
		fmt.Printf("%v\n", a)
		for _, k := range a {
			r.Put(u, wolkdb.IntToByte(k), wolkdb.SHA256(fmt.Sprintf("%v", k^x)))
		}

		for i, k := range a {
			v, ok, err := r.Get(u, wolkdb.IntToByte(k))
			if !ok || err != nil {
				t.Fatal(i, k, v, ok)
			}

			val := wolkdb.SHA256(fmt.Sprintf("%v", k^x))
			if bytes.Compare([]byte(val), v) != 0 {
				t.Fatal(i, val, v)
			}

			k |= 1

			_, ok, _ = r.Get(u, wolkdb.IntToByte(k))
			if ok {
				t.Fatal(i, k)
			}

		}

		for _, k := range a {
			r.Put(u, wolkdb.IntToByte(k), wolkdb.SHA256(fmt.Sprintf("%v", k^x+42)))
		}

		for i, k := range a {
			v, ok, err := r.Get(u, wolkdb.IntToByte(k))
			if !ok || err != nil {
				t.Fatal(i, k, v, ok)
			}

			val := wolkdb.SHA256(fmt.Sprintf("%v", k^x+42))
			if bytes.Compare([]byte(val), v) != 0 {
				t.Fatal(i, v, val)
			}

			k |= 1
			_, ok, _ = r.Get(u, wolkdb.IntToByte(k))
			if ok {
				t.Fatal(i, k)
			}
		}

	}
}

func TestDelete0(t *testing.T) {
	// TODO: make this test work!
	t.SkipNow()

	u := config.GetSWARMDBUser()

	hashid := make([]byte, 32)
	r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)

	key0 := wolkdb.IntToByte(0)
	key1 := wolkdb.IntToByte(1)

	val0 := wolkdb.SHA256("0")
	val1 := wolkdb.SHA256("1")

	if ok, _ := r.Delete(u, key0); ok {
		t.Fatal(ok)
	}

	r.Put(u, key0, val0)
	if ok, _ := r.Delete(u, key1); ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key0); !ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key0); ok {
		t.Fatal(ok)
	}

	r.Put(u, key0, val0)
	r.Put(u, key1, val1)
	if ok, _ := r.Delete(u, key1); !ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key1); ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key0); !ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key0); ok {
		t.Fatal(ok)
	}

	r.Put(u, key0, val0)
	r.Put(u, key1, val1)
	if ok, _ := r.Delete(u, key0); !ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key0); ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key1); !ok {
		t.Fatal(ok)
	}

	if ok, _ := r.Delete(u, key1); ok {
		t.Fatal(ok)
	}
}

func TestDelete1(t *testing.T) {
	// TODO: make this test work!
	t.SkipNow()

	u := config.GetSWARMDBUser()

	hashid := make([]byte, 32)
	const N = 130
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x333333, 0xcccccc, 0x314159} {
		r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)
		a := make([]int, N)
		for i := range a {
			a[i] = (i ^ x) << 1
		}
		for _, k := range a {
			r.Put(u, wolkdb.IntToByte(k), wolkdb.SHA256("0"))
		}

		for i, k := range a {
			ok, _ := r.Delete(u, wolkdb.IntToByte(k))
			if !ok {
				fmt.Printf("YIPE%s\n", k)
				t.Fatal(i, x, k)
			}
		}
	}
}

func TestDelete2(t *testing.T) {
	// TODO: make this test work!
	t.SkipNow()
	u := config.GetSWARMDBUser()

	const N = 100
	hashid := make([]byte, 32)
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x333333, 0xcccccc, 0x314159} {
		r, _ := wolkdb.NewBPlusTreeDB(u, swarmdb, hashid, sdbc.CT_INTEGER, false, sdbc.CT_STRING, TEST_ENCRYPTED)
		a := make([]int, N)
		rng := wolkdb.Rng()
		for i := range a {
			a[i] = (rng.Next() ^ x) << 1
		}
		for _, k := range a {
			r.Put(u, wolkdb.IntToByte(k), wolkdb.SHA256("0"))
		}
		for i, k := range a {
			ok, _ := r.Delete(u, wolkdb.IntToByte(k))
			if !ok {
				t.Fatal(i, x, k)
			}
		}
	}
}
