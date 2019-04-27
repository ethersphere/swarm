// Copyright 2019 The go-ethereum Authors
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

package chunk

import (
	"sync"
	"testing"
	"time"
)

var (
	allStates = []State{SPLIT, STORED, SEEN, SENT, SYNCED}
)

// TestTagSingleIncrements tests if Inc increments the tag state value
func TestTagSingleIncrements(t *testing.T) {
	tg := &Tag{total: 10}
	for _, f := range allStates {
		tg.Inc(f)
		if tg.Get(f) != 1 {
			t.Fatalf("not incremented")
		}
	}
}

// TestTagStatus is a unit test to cover Tag.Status method functionality
func TestTagStatus(t *testing.T) {
	tg := &Tag{total: 10}
	tg.Inc(SEEN)
	tg.Inc(SENT)
	tg.Inc(SYNCED)
	for i := 0; i < 10; i++ {
		tg.Inc(SPLIT)
	}
	for i := 0; i < 10; i++ {
		tg.Inc(STORED)
	}
	for _, v := range []struct {
		state    State
		expVal   int
		expTotal int
	}{
		{state: STORED, expVal: 10, expTotal: 10},
		{state: SPLIT, expVal: 10, expTotal: 10},
		{state: SEEN, expVal: 1, expTotal: 10},
		{state: SENT, expVal: 1, expTotal: 9},
		{state: SYNCED, expVal: 1, expTotal: 9},
	} {
		val, total, err := tg.Status(v.state)
		if err != nil {
			t.Fatal(err)
		}
		if val != v.expVal {
			t.Fatalf("should be %d, got %d", v.expVal, val)
		}
		if total != v.expTotal {
			t.Fatalf("expected total to be %d, got %d", v.expTotal, total)
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
	wg.Add(5 * n)
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
func TestTagsMultipleConcurrentIncrementsSyncMap(t *testing.T) {
	ts := newTags()
	n := 100
	wg := sync.WaitGroup{}
	wg.Add(10 * 5 * n)
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

// TestMarshalling tests that a Tag gets correctly marshalled and unmarshalled to and from a byte slice
func TestMarshalling(t *testing.T) {
	tg := NewTag(111, "test/tag", 10)
	for _, f := range allStates {
		tg.Inc(f)
		if tg.Get(f) != 1 {
			t.Fatalf("not incremented")
		}
	}

	b, err := tg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	unmarshalledTag := &Tag{}
	err = unmarshalledTag.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}

	if unmarshalledTag.Uid != tg.Uid {
		t.Fatalf("tag uids not equal. want %d got %d", tg.Uid, unmarshalledTag.Uid)
	}

	if unmarshalledTag.Name != tg.Name {
		t.Fatalf("tag names not equal. want %s got %s", tg.Name, unmarshalledTag.Name)
	}

	if unmarshalledTag.Get(SYNCED) != tg.Get(SYNCED) {
		t.Fatalf("tag names not equal. want %d got %d", tg.Get(SYNCED), unmarshalledTag.Get(SYNCED))
	}

	if unmarshalledTag.Total() != tg.Total() {
		t.Fatalf("tag names not equal. want %d got %d", tg.Total(), unmarshalledTag.Total())
	}
}
