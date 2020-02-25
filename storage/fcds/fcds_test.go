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

package fcds

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethersphere/swarm/chunk"
	chunktesting "github.com/ethersphere/swarm/chunk/testing"
	"github.com/ethersphere/swarm/testutil"
)

func init() {
	testutil.Init()
}

func TestStoreGrow(t *testing.T) {
	path, err := ioutil.TempDir("", "swarm-fcds")
	if err != nil {
		t.Fatal(err)
	}
	defer func(sc uint8) {
		ShardCount = sc
	}(ShardCount)

	ShardCount = 8
	capacity := 10000
	gcTarget := 100
	insert := 11000
	ms, err := NewMetaStore("", true)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(path, chunk.DefaultSize, ms, WithCache(false))
	if err != nil {
		os.RemoveAll(path)
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.RemoveAll(path)
	}()
	inserted := 0
	gcRuns := 0
	var mtx sync.Mutex
	sem := make(chan struct{}, 1)
	for i := 0; i < insert; i++ {
		ch := chunktesting.GenerateTestRandomChunk()
		err = s.Put(ch)
		if err != nil {
			t.Fatal(err)
		}
		mtx.Lock()
		inserted++
		mtx.Unlock()
		if inserted > capacity {
			select {
			case sem <- struct{}{}:
				gcRuns++
				go func() {
					defer func() {
						//time.Sleep(1 * time.Second)
						<-sem
					}()
					count := 0
					var wg sync.WaitGroup
					err := s.Iterate(func(c chunk.Chunk) (stop bool, err error) {
						fmt.Println("get chunk to iterate on ", c.Address().String())
						count++
						wg.Add(1)
						go func(c chunk.Address) {
							defer wg.Done()
							e := s.Delete(c)
							if e != nil {
								fmt.Println("error deleting", e, "c", c)
							}
							mtx.Lock()
							inserted--
							mtx.Unlock()
						}(c.Address())
						if count >= gcTarget {
							return true, nil
						}
						return false, nil
					})
					if err != nil {
						fmt.Println("iterator err", err)
					}
					wg.Wait()
				}()
			default:

			}

		}
		if i%1000 == 0 {
			mtx.Lock()
			ss := getShardsSum(s.shards)
			ssmb := ss / (1024 * 1024)
			insertedmb := i * 4096 / (1024 * 1024)
			expectedSum := capacity * 4096 / (1024 * 1024)

			fmt.Println("inserted", i, "insertedMB", insertedmb, "expectedSum", expectedSum, "shardsum", ss, "mb", ssmb, "gcruns", gcRuns)
			mtx.Unlock()
		}
	}

}

func getShardsSum(s []shard) int {
	sum := 0
	elems := make([]int64, len(s))
	for i, sh := range s {
		v, err := sh.f.Stat()
		if err != nil {
			panic(err)
		}
		elems[i] = v.Size()
		sum += int(v.Size())
	}

	spew.Dump("elements", elems)

	return sum
}
