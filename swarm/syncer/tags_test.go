package syncer

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var (
	allStates = []State{SPLIT, STORED, SENT, SYNCED}
)

// TestTagSingleIncrements tests if Inc increments the tag state value
func TestTagSingleIncrements(t *testing.T) {
	tg := &Tag{total: 10}
	for _, f := range allStates {
		tg.Inc(f)
		if tg.Get(f) != 1 {
			t.Fatalf("not incremented")
		}
		cnt, total := tg.Status(f)
		if cnt != 1 {
			t.Fatalf("expected count 1 for state %v, got %v", f, cnt)
		}
		if total != 10 {
			t.Fatalf("expected total count %v for state %v, got %v", 10, f, cnt)
		}
	}
}

// tests ETA is precise
func TestTagETA(t *testing.T) {
	now := time.Now()
	maxDiff := 100000 // 100 microsecond
	tg := &Tag{total: 10, startedAt: now}
	time.Sleep(100 * time.Millisecond)
	tg.Inc(SPLIT)
	eta, err := tg.ETA(SPLIT)
	if err != nil {
		t.Fatal(err)
	}
	diff := time.Until(eta) - 9*time.Since(now)
	if int(diff) > maxDiff || int(diff) < -maxDiff {
		t.Fatalf("ETA is not precise, got diff %v > .1ms", diff)
	}
}

// TestTagConcurrentIncrements tests Inc calls concurrently
func TestTagConcurrentIncrements(t *testing.T) {
	tg := &Tag{}
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(4 * n)
	for _, f := range allStates {
		go func(f State) {
			for j := 0; j < n; j++ {
				go func() {
					tg.Inc(f)
					wg.Done()
				}()
			}
		}(f)
	}
	wg.Wait()
	for _, f := range allStates {
		v := tg.Get(f)
		if v != n {
			t.Fatalf("expected state %v to be %v, got %v", f, n, v)
		}
	}
}

// TestTagsMultipleConcurrentIncrements tests Inc calls concurrently
func TestTagsMultipleConcurrentIncrements(t *testing.T) {
	ts := newTags()
	n := 100
	wg := sync.WaitGroup{}
	wg.Add(10 * 4 * n)
	for i := 0; i < 10; i++ {
		s := string([]byte{uint8(i)})
		ts.New(s, n)
		for _, f := range allStates {
			go func(s string, f State) {
				for j := 0; j < n; j++ {
					go func() {
						ts.Inc(s, f)
						wg.Done()
					}()
				}
			}(s, f)
		}
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		s := string([]byte{uint8(i)})
		for _, f := range allStates {
			v := ts.Get(s, f)
			if v != n {
				t.Fatalf("expected tag %v state %v to be %v, got %v", s, f, n, v)
			}
		}
	}
}

// tests the correct behaviour of tags while using the DB
func TestDBWithTags(t *testing.T) {
	names := []string{"1", "2", "3", "4"}
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
	defer db.Close()

	// feed fake chunks into the db, hashes encode the order so that
	// it can be traced
	for i, name := range names {
		total := i*100 + 100
		db.tags.New(name, total)
		go func(name string, total int) {
			for j := 0; j < total; j++ {
				db.Put(&item{Addr: network.RandomAddr().OAddr, Tag: name})
			}
		}(name, total)
	}

	err = waitTillEmpty(db)
	if err != nil {
		t.Fatal(err)
	}

	states := []State{STORED, SENT, SYNCED}
	var cnt int
	for i, name := range names {
		total := i*100 + 100
		for _, state := range states {
			cnt = db.tags.Get(name, state)
			if cnt != total {
				t.Fatalf("expected tag %v state %v to count %v, got %v", name, state, total, cnt)
			}
		}
	}
}
