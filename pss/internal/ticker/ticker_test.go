package ticker_test

import (
	"sync"
	"testing"
	"time"

	"github.com/epiclabs-io/ut"
	"github.com/ethersphere/swarm/pss/internal/ticker"
	"github.com/tilinna/clock"
)

// TestNewTicker tests whether the ticker calls a callback function periodically
func TestNewTicker(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()
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
	t.Ok(err)

	err = testTicker.Stop()
	t.MustFailWith(err, ticker.ErrAlreadyStopped)

}
