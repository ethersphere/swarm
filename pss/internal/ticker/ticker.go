package ticker

import (
	"errors"
	"time"

	"github.com/tilinna/clock"
)

// Config defines the necessary information and dependencies to instantiate a Ticker
type Config struct {
	Clock    clock.Clock
	Interval time.Duration
	Callback func()
}

// Ticker represents a periodic timer that invokes a callback
type Ticker struct {
	quitC chan struct{}
}

// ErrAlreadyStopped is returned if this service was already stopped and Stop() is called again
var ErrAlreadyStopped = errors.New("Already stopped")

// New builds a ticker that will call the given callback function periodically
func New(config *Config) *Ticker {

	tk := &Ticker{
		quitC: make(chan struct{}),
	}
	ticker := config.Clock.NewTicker(config.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				config.Callback()
			case <-tk.quitC:
				return
			}
		}
	}()
	return tk
}

// Stop stops the timer and releases the goroutine running it.
func (tk *Ticker) Stop() error {
	if tk.quitC == nil {
		return ErrAlreadyStopped
	}
	close(tk.quitC)
	tk.quitC = nil
	return nil
}
