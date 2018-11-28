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

package syncer

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 3, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// TestDBIteration tests the correct behaviour of DB, ie.
// in the context of inserting n chunks
// proof response model: after calling the sync function, the chunk hash appeared on
// the receiptsC channel after a random delay
// The test checks:
// - sync function is called on chunks in order of insertion (FIFO)
// - repeated calls are not before retryInterval time passed
// - already synced chunks are not resynced
// - if no more data inserted, the db is emptied shortly
func TestDBIteration(t *testing.T) {
	timeout := 30 * time.Second
	chunkCnt := 10000

	receiptsC := make(chan storage.Address)
	chunksSentAt := make([]*time.Time, chunkCnt)

	errc := make(chan error)
	quit := make(chan struct{})
	defer close(quit)
	errf := func(s string, vals ...interface{}) {
		select {
		case errc <- fmt.Errorf(s, vals...):
		case <-quit:
		}
	}

	var max uint64     // the highest index sent so far
	var complete int64 // number of chunks that got poc response
	// sync function is not called concurrently, so max need no lock
	// TODO: chunksSentAt array should use lock
	syncf := func(chunk storage.Chunk) error {
		cur := binary.BigEndian.Uint64(chunk.Address()[:8])
		if cur > max+1 {
			errf("incorrect order of chunks from db chunk #%d before #%d", cur, max+1)
		}
		now := time.Now()
		if cur < max+1 {
			sentAt := chunksSentAt[cur-1]
			if sentAt == nil {
				errf("resyncing already synced chunk #%d: %v", cur, sentAt)
				return nil
			}
			if (*sentAt).Add(retryInterval).After(now) {
				errf("resync chunk #%d too early", cur)
				return nil
			}
		} else {
			max = cur
		}
		chunksSentAt[cur-1] = &now
		// this go routine mimics the chunk sync - poc response roundrtip
		// with random delay (uniform within a fixed range)
		go func() {
			n := rand.Intn(1000)
			delay := time.Duration(n+5) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.TODO(), delay)
			defer cancel()
			select {
			case <-ctx.Done():
			case <-quit:
				return
			}
			receiptsC <- chunk.Address()
			chunksSentAt[cur-1] = nil
			delCnt := atomic.AddInt64(&complete, 1)
			if int(delCnt) == chunkCnt {
				close(errc)
				return
			}
		}()
		return nil
	}

	// initialise db, it starts all the go routines
	dbpath, err := ioutil.TempDir(os.TempDir(), "syncertest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbpath)
	db, err := NewDB(dbpath, nil, syncf, receiptsC)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// feed fake chunks into the db, hashes encode the order so that
	// it can be traced
	go func() {
		for i := 1; i <= chunkCnt; i++ {
			addr := make([]byte, 32)
			binary.BigEndian.PutUint64(addr, uint64(i))
			c := &item{
				Addr: addr,
			}
			db.Put(c)
		}
	}()

	// wait on errc for errors on any thread or close if success
	// otherwise time out
	select {
	case err := <-errc:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(timeout):
		t.Fatalf("timeout")
	}

	err = waitTillEmpty(db)
	if err != nil {
		t.Fatal(err)
	}
}

// TestDBIteration tests the correct behaviour of DB that while constantly inserting  chunks
//
func TestDBCompleteMultipleSessions(t *testing.T) {
	chunkCnt := 1000

	receiptsC := make(chan storage.Address)
	quit := make(chan struct{})
	defer close(quit)
	// sync function is not called concurrently, so max need no lock
	// TODO: chunksSentAt array should use lock
	sync := func(chunk storage.Chunk) error {

		// this go routine mimics the chunk sync - poc response roundrtip
		// with random delay (uniform within a fixed range)
		go func() {
			n := rand.Intn(1000)
			delay := time.Duration(n+5) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.TODO(), delay)
			defer cancel()
			select {
			case <-ctx.Done():
				receiptsC <- chunk.Address()
			case <-quit:
			}

		}()
		return nil
	}
	// initialise db, it starts all the go routines
	dbpath, err := ioutil.TempDir(os.TempDir(), "syncertest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbpath)
	db, err := NewDB(dbpath, nil, sync, receiptsC)
	if err != nil {
		t.Fatal(err)
	}

	// feed fake chunks into the db, hashes encode the order so that
	// it can be traced
	i := 1
	sessionChunkCnt := 10
	var round int
	ticker := time.NewTicker(10 * time.Microsecond)
	defer ticker.Stop()
	for range ticker.C {

		db.Put(&item{Addr: network.RandomAddr().OAddr})
		i++
		if i > chunkCnt {
			break
		}
		if i > sessionChunkCnt {
			round++
			log.Warn("session ends", "round", round, "chunks", i, "unsynced", db.Size())
			db.Close()
			db, err = NewDB(dbpath, nil, sync, receiptsC)
			if err != nil {
				t.Fatal(err)
			}
			sessionChunkCnt += rand.Intn(100)
		}
	}
	err = waitTillEmpty(db)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func waitTillEmpty(db *DB) error {
	checkticker := time.NewTicker(50 * time.Millisecond)
	defer checkticker.Stop()
	round := 0
	for range checkticker.C {
		size := db.Size()
		if size == 0 {
			break
		}
		if round > 50 {
			return fmt.Errorf("timeout waiting for db size 0, got %v", size)
		}
		round++
	}
	return nil
}
