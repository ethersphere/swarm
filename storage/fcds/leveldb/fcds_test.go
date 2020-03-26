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

package leveldb_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage/fcds"
	"github.com/ethersphere/swarm/storage/fcds/leveldb"
)

const (
	ConcurrentThreads = 128
)

func newDB(b *testing.B) (db fcds.Storer, clean func()) {
	b.Helper()

	path, err := ioutil.TempDir("/tmp/swarm", "swarm-shed")
	if err != nil {
		b.Fatal(err)
	}
	metaStore, err := leveldb.NewMetaStore(filepath.Join(path, "meta"))
	if err != nil {
		b.Fatal(err)
	}

	db, err = fcds.New(path, 4096, metaStore)
	if err != nil {
		os.RemoveAll(path)
		b.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(path)
	}
}

func getChunks(count int, chunkCache []chunk.Chunk) []chunk.Chunk {
	l := len(chunkCache)
	if l == 0 {
		chunkCache = make([]chunk.Chunk, count)
		for i := 0; i < count; i++ {
			chunkCache[i] = GenerateTestRandomChunk()
		}
		return chunkCache
	}
	if l < count {
		for i := 0; i < count-l; i++ {
			chunkCache = append(chunkCache, GenerateTestRandomChunk())
		}
		return chunkCache
	}
	return chunkCache[:count]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateTestRandomChunk() chunk.Chunk {
	data := make([]byte, chunk.DefaultSize)
	rand.Read(data)
	key := make([]byte, 32)
	rand.Read(key)
	return chunk.NewChunk(key, data)
}

func createBenchBaseline(b *testing.B, baseChunksCount int) (db fcds.Storer, clean func(), baseChunks []chunk.Chunk) {
	db, clean = newDB(b)

	if baseChunksCount > 0 {
		baseChunks = getChunks(baseChunksCount, baseChunks)
		//start := time.Now()
		sem := make(chan struct{}, ConcurrentThreads)
		var wg sync.WaitGroup
		wg.Add(baseChunksCount)
		for i, ch := range baseChunks {
			sem <- struct{}{}
			go func(i int, ch chunk.Chunk) {
				defer func() {
					<-sem
					wg.Done()
				}()
				if _, err := db.Put(ch); err != nil {
					panic(err)
				}
			}(i, ch)
		}
		wg.Wait()
		//elapsed := time.Since(start)
		//fmt.Println("-- adding base chunks took, ", elapsed)
	}

	rand.Shuffle(baseChunksCount, func(i, j int) {
		baseChunks[i], baseChunks[j] = baseChunks[j], baseChunks[i]
	})

	return db, clean, baseChunks
}

// Benchmarkings

func runBenchmark(b *testing.B, db fcds.Storer, basechunks []chunk.Chunk, baseChunksCount int, writeChunksCount int, readChunksCount int, deleteChunksCount int) {
	var writeElapsed time.Duration
	var readElapsed time.Duration
	var deleteElapsed time.Duration

	var writeChunks []chunk.Chunk
	writeChunks = getChunks(writeChunksCount, writeChunks)
	b.StartTimer()

	var jobWg sync.WaitGroup
	if writeChunksCount > 0 {
		jobWg.Add(1)
		go func() {
			start := time.Now()
			sem := make(chan struct{}, ConcurrentThreads)
			var wg sync.WaitGroup
			wg.Add(writeChunksCount)
			for i, ch := range writeChunks {
				sem <- struct{}{}
				go func(i int, ch chunk.Chunk) {
					defer func() {
						<-sem
						wg.Done()
					}()
					if _, err := db.Put(ch); err != nil {
						panic(err)
					}
				}(i, ch)
			}
			wg.Wait()
			elapsed := time.Since(start)
			//fmt.Println("-- writing chunks took , ", elapsed)
			writeElapsed += elapsed
			jobWg.Done()
		}()
	}

	if readChunksCount > 0 {
		jobWg.Add(1)
		go func() {
			errCount := 0
			start := time.Now()
			sem := make(chan struct{}, ConcurrentThreads*4)
			var wg sync.WaitGroup
			wg.Add(readChunksCount)
			for i, ch := range basechunks {
				if i >= readChunksCount {
					break
				}
				sem <- struct{}{}
				go func(i int, ch chunk.Chunk) {
					defer func() {
						<-sem
						wg.Done()
					}()
					_, err := db.Get(ch.Address())
					if err != nil {
						//panic(err)
						errCount++
					}
				}(i, ch)
			}
			wg.Wait()
			elapsed := time.Since(start)
			//fmt.Println("-- reading chunks took , ", elapsed)
			readElapsed += elapsed
			jobWg.Done()
		}()
	}

	if deleteChunksCount > 0 {
		jobWg.Add(1)
		go func() {
			start := time.Now()
			sem := make(chan struct{}, ConcurrentThreads)
			var wg sync.WaitGroup
			wg.Add(deleteChunksCount)
			for i, ch := range basechunks {
				if i >= deleteChunksCount {
					break
				}
				sem <- struct{}{}
				go func(i int, ch chunk.Chunk) {
					defer func() {
						<-sem
						wg.Done()
					}()
					if err := db.Delete(ch.Address()); err != nil {
						panic(err)
					}
				}(i, ch)
			}
			wg.Wait()
			elapsed := time.Since(start)
			//fmt.Println("-- deleting chunks took , ", elapsed)
			deleteElapsed += elapsed
			jobWg.Done()
		}()
	}

	jobWg.Wait()

	//if writeElapsed > 0 {
	//	fmt.Println("- Average write  time : ", writeElapsed.Nanoseconds()/int64(iterationCount), " ns/op")
	//}
	//if readElapsed > 0 {
	//	//fmt.Println("- Average read time : ", readElapsed.Nanoseconds()/int64(iterationCount), " ns/op")
	//}
	//if deleteElapsed > 0 {
	//	//fmt.Println("- Average delete time : ", deleteElapsed.Nanoseconds()/int64(iterationCount), " ns/op")
	//}
}

func BenchmarkWrite(b *testing.B) {
	for i := 10000; i <= 1000000; i *= 10 {
		b.Run(fmt.Sprintf("baseline_%d", i), func(b *testing.B) {
			// for each baseline, insert 10k, 20k, 50k, 100k
			for _, k := range []int{10000, 20000, 50000, 100000} {
				b.Run(fmt.Sprintf("add_%d", k), func(b *testing.B) {
					for j := 0; j < b.N; j++ {
						b.StopTimer()
						db, clean, baseChunks := createBenchBaseline(b, i)
						b.StartTimer()

						runBenchmark(b, db, baseChunks, 0, k, 0, 0)
						b.StopTimer()
						clean()
						b.StartTimer()
					}
				})
			}
		})
	}
}

func BenchmarkRead(b *testing.B) {
	for i := 10000; i <= 1000000; i *= 10 {
		b.Run(fmt.Sprintf("baseline_%d", i), func(b *testing.B) {
			b.StopTimer()
			db, clean, baseChunks := createBenchBaseline(b, i)
			b.StartTimer()

			for k := 10000; k <= i; k *= 10 {
				b.Run(fmt.Sprintf("read_%d", k), func(b *testing.B) {
					for j := 0; j < b.N; j++ {
						runBenchmark(b, db, baseChunks, 0, 0, k, 0)
					}
				})
			}
			b.StopTimer()
			clean()
			b.StartTimer()
		})
	}
}

//func BenchmarkWriteOverClean_100000(t *testing.B) { runBenchmark(t, 0, 100000, 0, 0, 6) }
//func BenchmarkWriteOverClean_1000000(t *testing.B) { runBenchmark(t, 0, 1000000, 0, 0, 4) }

//func BenchmarkWriteOver1Million_10000(t *testing.B) {
//for i := 0; i < t.N; i++ {
//runBenchmark(t, 1000000, 10000, 0, 0, 8)
//}

//}

//func BenchmarkWriteOver1Million_100000(t *testing.B) { runBenchmark(t, 1000000, 100000, 0, 0,6) }
//func BenchmarkWriteOver1Million_1000000(t *testing.B) { runBenchmark(t, 1000000, 1000000, 0, 0, 4) }
//func BenchmarkWriteOver1Million_5000000(t *testing.B) { runBenchmark(t, 5000000, 1000000, 0, 0, 4) }
//func BenchmarkWriteOver1Million_10000000(t *testing.B) { runBenchmark(t, 10000000, 1000000, 0, 0, 4) }

//func BenchmarkReadOver1Million_10000(t *testing.B) { runBenchmark(t, 1000000, 0, 10000, 0,8) }
//func BenchmarkReadOver1Million_100000(t *testing.B) { runBenchmark(t, 1000000, 0, 100000, 0, 6) }
//func BenchmarkReadOver1Million_1000000(t *testing.B) { runBenchmark(t, 1000000, 0, 1000000, 0, 4) }

//func BenchmarkDeleteOver1Million_10000(t *testing.B) { runBenchmark(t, 1000000, 0, 0, 10000,8) }
//func BenchmarkDeleteOver1Million_100000(t *testing.B) { runBenchmark(t, 1000000, 0, 0, 100000,6) }
//func BenchmarkDeleteOver1Million_1000000(t *testing.B) { runBenchmark(t, 1000000, 0, 0, 1000000, 4) }

//func BenchmarkWriteReadOver1Million_10000(t *testing.B) { runBenchmark(t, 1000000, 10000, 10000, 0,8) }
//func BenchmarkWriteReadOver1Million_100000(t *testing.B) { runBenchmark(t, 1000000, 100000, 100000, 0,6) }
//func BenchmarkWriteReadOver1Million_1000000(t *testing.B) {runBenchmark(t, 1000000, 1000000, 1000000, 0, 4)}

//func BenchmarkWriteReadDeleteOver1Million_10000(t *testing.B) { runBenchmark(t, 1000000, 10000, 10000, 10000,8) }
//func BenchmarkWriteReadDeleteOver1Million_100000(t *testing.B) { runBenchmark(t, 1000000, 100000, 100000, 100000,6) }
//func BenchmarkWriteReadDeleteOver1Million_1000000(t *testing.B) {runBenchmark(t, 1000000, 1000000, 1000000, 1000000, 4)}
