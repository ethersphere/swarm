package file

import "sync"

// passed to a job to determine at which data lengths and levels a job should terminate
type target struct {
	size     int32         // bytes written
	sections int32         // sections written
	level    int32         // target level calculated from bytes written against branching factor and sector size
	resultC  chan []byte   // channel to receive root hash
	doneC    chan struct{} // when this channel is closed all jobs will calculate their end write count
	mu       sync.Mutex
}

func newTarget() *target {
	return &target{
		resultC: make(chan []byte),
		doneC:   make(chan struct{}),
	}
}

// Set is called when the final length of the data to be written is known
// TODO: method can be simplified to calculate sections and level internally
func (t *target) Set(size int, sections int, level int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.size = int32(size)
	t.sections = int32(sections)
	t.level = int32(level)
	//log.Trace("target set", "size", t.size, "section", t.sections, "level", t.level)
	close(t.doneC)
}

// Count returns the total section count for the target
// it should only be called after Set()
func (t *target) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return int(t.sections) + 1
}

func (t *target) Level() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return int(t.level)
}

func (t *target) Size() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return int(t.size)
}

// Done returns the channel in which the root hash will be sent
func (t *target) Done() <-chan []byte {
	return t.resultC
}
