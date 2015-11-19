package bzz

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type testSyncDb struct {
	*syncDb
	c         int
	deliverC  chan bool
	t         *testing.T
	delivered [][]byte
	saved     []bool
	dbdir     string
	total     int
	at        int
}

func newTestSyncDb(priority, bufferSize int, dbdir string, t *testing.T) *testSyncDb {
	if len(dbdir) == 0 {
		tmp, err := ioutil.TempDir(os.TempDir(), "syndbtest")
		if err != nil {
			t.Fatalf("unable to create temporary direcory %v: %v", tmp, err)
		}
		dbdir = tmp
	}
	db, err := NewLDBDatabase(filepath.Join(dbdir, "requestdb"))
	if err != nil {
		t.Fatalf("unable to create db: %v", err)
	}
	self := &testSyncDb{
		deliverC: make(chan bool),
		dbdir:    dbdir,
		t:        t,
	}
	h := crypto.Sha3Hash([]byte{0})
	key := Key(h[:])
	self.syncDb = newSyncDb(db, key, uint(priority), uint(bufferSize), self.deliver)
	// kick off db iterator right away, if no items on db this will allow
	// reading from the buffer
	go self.syncDb.iterate(self.deliver)
	return self

}

func (self *testSyncDb) close() {
	os.RemoveAll(self.dbdir)
}

func (self *testSyncDb) push(n int) {
	for i := 0; i < n; i++ {
		self.buffer <- Key(crypto.Sha3([]byte{byte(self.c)}))
		self.c++
	}
}

func (self *testSyncDb) deliver(req interface{}, quit chan bool) bool {
	select {
	case <-self.deliverC:
		_, ok := req.(*syncDbEntry)
		self.saved = append(self.saved, ok)
		key, _, _, _, err := parseRequest(req)
		if err != nil {
			self.t.Fatalf("unexpected error of key %v: %v", key, err)
		}
		r := make([]byte, 32)
		copy(r, key[:])
		self.delivered = append(self.delivered, r)
		return true
	case <-quit:
		return false
	}
}

func (self *testSyncDb) expect(n int, db bool) {
	var attempts int
	var ok bool
	// for n items
	for i := 0; i < n; {
		// wait till item appears for attempts * 10 ms
		if self.at == len(self.saved) {
			if attempts == 0 {
				self.deliverC <- true
			}
			attempts++
			if attempts == 100 {
				self.t.Fatalf("timed out: expected %v, got %v", len(self.saved), len(self.delivered))
			}
			time.Sleep(1 * time.Millisecond)
			continue
		}
		attempts = 0
		i++
		ok = self.saved[self.at]
		if !ok && db {
			self.t.Fatalf("expected delivery %v/%v from db", self.at, self.total)
		}
		if ok && !db {
			self.t.Fatalf("expected delivery %v/%v from cache", self.at, self.total)
		}
		self.at++
	}
}

func TestSyncDb(t *testing.T) {
	priority := High
	bufferSize := 5
	s := newTestSyncDb(priority, bufferSize, "", t)
	defer s.stop()
	defer s.close()

	s.push(4)
	s.expect(1, false)

	s.push(2)
	s.expect(5, false)

	s.push(4)
	s.expect(3, true)

	s.push(3)
	s.expect(4, true)

	s.push(3)
	s.expect(3, false)

	s.push(5)
	s.expect(4, false)

	s.push(1)
	s.expect(1, false)
	s.expect(1, true)
}

func TestSaveSyncDb(t *testing.T) {
	amount := 500
	priority := High
	bufferSize := amount
	s := newTestSyncDb(priority, bufferSize, "", t)
	s.push(amount)
	s.stop()

	s = newTestSyncDb(priority, bufferSize, s.dbdir, t)
	s.expect(amount, true)
	for i, key := range s.delivered {
		expKey := crypto.Sha3([]byte{byte(i)})
		if !bytes.Equal(key, expKey) {
			t.Fatalf("delivery %v expected to be key %x, got %x", i, expKey, key)
		}
	}

	s.push(amount)
	s.expect(amount, false)
	for i := amount; i < 2*amount; i++ {
		key := s.delivered[i]
		expKey := crypto.Sha3([]byte{byte(i - amount)})
		if !bytes.Equal(key, expKey) {
			t.Fatalf("delivery %v expected to be key %x, got %x", i, expKey, key)
		}
	}
	s.stop()

	s = newTestSyncDb(priority, bufferSize, s.dbdir, t)
	defer s.stop()
	defer s.close()

	s.push(1)
	s.expect(1, false)

}
