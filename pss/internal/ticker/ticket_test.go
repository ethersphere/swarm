package ticker_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ethersphere/swarm/pss/internal/ticker"
	"github.com/tilinna/clock"
)

// TestNewTicker tests whether the ticker calls a callback function periodically
func TestNewTicker(t *testing.T) {
	var err error

	testClock := clock.NewMock(time.Unix(0, 0))
	interval := 10 * time.Second

	wg := sync.WaitGroup{}
	wg.Add(10)
	tickWait := make(chan bool)

	testTicker := ticker.New(&ticker.Config{
		Interval: interval,
		Clock:    testClock,
		Callback: func() {
			wg.Done()
			tickWait <- true
		},
	})

	for i := 0; i < 10; i++ {
		testClock.Add(interval)
		<-tickWait
	}

	wg.Wait()
	err = testTicker.Stop()
	if err != nil {
		t.Fatal(err)
	}

	err = testTicker.Stop()
	if err != ticker.ErrAlreadyStopped {
		t.Fatal("Expected Stop() to return ticker.ErrAlreadyStopped when trying to stop an already stopped ticker")
	}
}
