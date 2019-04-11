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
func TestTagsMultipleConcurrentIncrements(t *testing.T) {
	tagMap := sync.Map{}
	n := 100
	wg := sync.WaitGroup{}
	wg.Add(10 * 4 * n)
	for i := 0; i < 10; i++ {
		s := string([]byte{uint8(i)})
		newTag := NewTag(uint32(i), s, n)
		tagMap.Store(i, newTag)
		for _, f := range allStates {
			go func(i int, f State) {
				for j := 0; j < n; j++ {
					go func() {
						tag, _ := tagMap.Load(i)
						tag.(*Tag).Inc(f)
						wg.Done()
					}()
				}
			}(i, f)
		}
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		for _, f := range allStates {
			v, _ := tagMap.Load(i)
			count := v.(*Tag).Get(f)
			if count != n {
				t.Fatalf("expected tag %d state %v to be %v, got %v", i, f, n, count)
			}
		}
	}
}
