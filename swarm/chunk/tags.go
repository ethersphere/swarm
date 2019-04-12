package chunk

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

var (
	errExists = errors.New("already exists")
	errNoETA  = errors.New("unable to calculate ETA")
)

// tags holds the tag infos indexed by name
type tags struct {
	tags *sync.Map
	rng  *rand.Rand
}

// NewTags creates a tags object
func newTags() *tags {

	return &tags{
		tags: &sync.Map{},
		rng:  rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// New creates a new tag, stores it by the name and returns it
// it returns an error if the tag with this name already exists
func (ts *tags) New(s string, total int) (*Tag, error) {
	t := &Tag{
		uid:       ts.rng.Uint32(),
		name:      s,
		startedAt: time.Now(),
		total:     uint32(total),
		State:     make(chan State, 5),
	}
	_, loaded := ts.tags.LoadOrStore(s, t)
	if loaded {
		return nil, errExists
	}
	return t, nil
}

// Inc increments the state count for a tag if tag is found
func (ts *tags) Inc(s string, f State) {
	t, ok := ts.tags.Load(s)
	if !ok {
		return
	}
	t.(*Tag).Inc(f)
}

// Get returns the state count for a tag
func (ts *tags) Get(s string, f State) int {
	t, _ := ts.tags.Load(s)
	return t.(*Tag).Get(f)
}

// State is the enum type for chunk states
type State = uint32

const (
	SPLIT  State = iota // chunk has been processed by filehasher/swarm safe call
	STORED              // chunk stored locally
	SENT                // chunk sent to neighbourhood
	SYNCED              // proof is received; chunk removed from sync db; chunk is available everywhere
)

// Tag represents info on the status of new chunks
type Tag struct {
	uid       uint32     // a unique identifier for this tag
	name      string     // a name tag for this tag
	total     uint32     // total chunks belonging to a tag
	split     uint32     // number of chunks already processed by splitter for hashing
	stored    uint32     // number of chunks already stored locally
	sent      uint32     // number of chunks sent for push syncing
	synced    uint32     // number of chunks synced with proof
	startedAt time.Time  // tag started to calculate ETA
	State     chan State // channel to signal completion
}

// New creates a new tag, stores it by the name and returns it
// it returns an error if the tag with this name already exists
func NewTag(uid uint32, s string, total uint32) *Tag {
	if len(s) == 0 {
		s = fmt.Sprintf("upload_%d", time.Now().Unix())
	}
	t := &Tag{
		uid:       uid,
		name:      s,
		startedAt: time.Now(),
		total:     uint32(total),
		State:     make(chan State, 5),
	}
	return t
}

// Inc increments the count for a state
func (t *Tag) Inc(state State) {
	var v *uint32
	switch state {
	case SPLIT:
		v = &t.split
	case STORED:
		v = &t.stored
	case SENT:
		v = &t.sent
	case SYNCED:
		v = &t.synced
	}
	n := atomic.AddUint32(v, 1)
	if int(n) == t.GetTotal() {
		t.State <- state
	}
}

// Get returns the count for a state on a tag
func (t *Tag) Get(state State) int {
	var v *uint32
	switch state {
	case SPLIT:
		v = &t.split
	case STORED:
		v = &t.stored
	case SENT:
		v = &t.sent
	case SYNCED:
		v = &t.synced
	}
	return int(atomic.LoadUint32(v))
}

// GetUid returns the unique identifier
func (t Tag) GetUid() uint32 {
	return t.uid
}

func (t Tag) GetName() string {
	return t.name
}

// GetTotal returns the total count
func (t *Tag) GetTotal() int {
	return int(atomic.LoadUint32(&t.total))
}

// SetTotal sets total count to SPLIT count
// is meant to be called when splitter finishes for input streams of unknown size
func (t *Tag) SetTotal() int {
	total := atomic.LoadUint32(&t.split)
	atomic.StoreUint32(&t.total, total)
	return int(total)
}

// Status returns the value of state and the total count
func (t *Tag) Status(state State) (int, int) {
	return t.Get(state), int(atomic.LoadUint32(&t.total))
}

// ETA returns the time of completion estimated based on time passed and rate of completion
func (t *Tag) ETA(state State) (time.Time, error) {
	cnt := t.Get(state)
	total := t.GetTotal()
	if cnt == 0 || total == 0 {
		return time.Time{}, errNoETA
	}
	diff := time.Since(t.startedAt)
	dur := time.Duration(total) * diff / time.Duration(cnt)
	return t.startedAt.Add(dur), nil
}

// WaitTill blocks until count for the State reaches total cnt
func (tg *Tag) WaitTill(ctx context.Context, s State) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case c := <-tg.State:
			if c == s {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Info("Status", "name", tg.name, "SENT", tg.Get(SENT), "SYNCED", tg.Get(SYNCED))
		}
	}
}

func (tag *Tag) MarshalBinary() (data []byte, err error) {
	intBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(intBuffer, tag.uid)
	buffer := append([]byte{}, intBuffer...)

	binary.BigEndian.PutUint32(intBuffer, tag.synced)
	buffer = append(buffer, intBuffer...)

	binary.BigEndian.PutUint32(intBuffer, tag.total)
	buffer = append(buffer, intBuffer...)

	intBuffer = make([]byte, 8)
	n := binary.PutVarint(intBuffer, tag.startedAt.Unix())
	buffer = append(buffer, intBuffer[:n]...)

	buffer = append(buffer, []byte(tag.name)...)

	return buffer, nil
}

func (tag *Tag) UnmarshalBinary(buffer []byte) error {
	if len(buffer) < 13 {
		return errors.New("buffer too short")
	}

	tag.uid = binary.BigEndian.Uint32(buffer[:4])
	tag.synced = binary.BigEndian.Uint32(buffer[4:8])
	tag.total = binary.BigEndian.Uint32(buffer[8:12])
	t, n := binary.Varint(buffer[12:])
	tag.startedAt = time.Unix(t, 0)
	tag.name = string(buffer[12+n:])

	return nil

}
