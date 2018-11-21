package rushed

import (
	"errors"

	"github.com/ethereum/go-ethereum/swarm/shed"
)

const (
	iterBatchSize = 128
)

var (
	errCancelled = errors.New("cancelled")
)

type Subscription struct {
	cancel chan struct{} // cancel the subscription
	err    error
}

func Subscribe(index *shed.Index, buffer chan *shed.IndexItem, from *shed.IndexItem, trigger chan struct{}) *Subscription {
	cancel := make(chan struct{})
	f := func(item *shed.IndexItem) (bool, error) {
		select {
		case buffer <- item:
			return false, nil
		case <-cancel:
			return false, errCancelled
		}
	}
	s := &Subscription{
		cancel: cancel,
	}
	wait := func() (bool, error) {
		select {
		case <-trigger:
			return false, nil
		case <-cancel:
			return false, errCancelled
		}
	}
	go func() {
		defer close(buffer)
		s.err = Iterate(index, from, f, wait)
	}()
	return s
}

// iterate is a wrapper to shed.IterateFrom that periodically iterates starting from 'from'
// and remembering the last item on each round and continue the iteration from this on the
// following round
// once the items are retrieved in a fixed slice of max iterBatchSize elements
// it iterates over this slice and applies f to each element
// f returns a bool which when true terminates the iteration
// error returned from f result in terminating the iteration and returning the error
// if the iterator reached the last item in the index it calls the wait function
func Iterate(index *shed.Index, from *shed.IndexItem, f func(*shed.IndexItem) (bool, error), wait func() (bool, error)) error {
	items := make([]*shed.IndexItem, iterBatchSize)
	pos := 0
	cur := from
	size := 0
	// define feed function that populates the items slice
	feed := func(item shed.IndexItem) (bool, error) {
		// assign the item at pos
		items[pos] = &item
		pos++
		cur = &item
		// if reached the end, stop
		if pos == len(items) {
			return true, nil
		}
		return false, nil
	}
	// read when called triggers an IterateFrom on the index, populates the items slice
	read := func() (int, error) {
		defer func() { pos = 0 }()
		for {
			if err := index.IterateFrom(*cur, feed); err != nil {
				return size, err
			}
			if size > 0 {
				break
			}
			// if no items are available it calls wait and returns if stop or error
			stop, err := wait()
			if err != nil {
				return 0, err
			}
			if stop {
				return 0, nil
			}
		}
		return size, nil
	}
	cnt := 0
	for {
		if cnt == size {
			// retrieved items are all fed to buffer
			// get a new batch
			// if c is buffered channel, it can still get items while batch is read from disk
			// size items read, last is set if after size element no more needed
			var err error
			size, err = read()
			if err != nil {
				return err
			}
			cnt = 0
		}
		// calls f on the item
		stop, err := f(items[cnt])
		if err != nil {
			return err
		}
		if stop {
			break
		}
		cnt++
	}
	return nil
}
