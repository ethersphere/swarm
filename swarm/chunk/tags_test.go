package chunk

import (
	"sync"
	"testing"
	"time"
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
func TestTagsMultipleConcurrentIncrementsSyncMap(t *testing.T) {
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

func TestMarshalling(t *testing.T) {
	tg := NewTag(111, "test/tag", 10)
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

	b, err := tg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	unmarshalledTag := &Tag{}
	err = unmarshalledTag.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}

	if unmarshalledTag.GetUid() != tg.GetUid() {
		t.Fatalf("tag uids not equal. want %d got %d", tg.GetUid(), unmarshalledTag.GetUid())
	}

	if unmarshalledTag.GetName() != tg.GetName() {
		t.Fatalf("tag names not equal. want %s got %s", tg.GetName(), unmarshalledTag.GetName())
	}

	if unmarshalledTag.Get(SYNCED) != tg.Get(SYNCED) {
		t.Fatalf("tag names not equal. want %d got %d", tg.Get(SYNCED), unmarshalledTag.Get(SYNCED))
	}

	if unmarshalledTag.GetTotal() != tg.GetTotal() {
		t.Fatalf("tag names not equal. want %d got %d", tg.GetTotal(), unmarshalledTag.GetTotal())
	}
}
