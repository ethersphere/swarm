package rushed

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/shed"
	"github.com/ethereum/go-ethereum/swarm/storage"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 3, "verbosity of logs")
)

func init() {
	flag.Parse()
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

type tester struct {
	index shed.Index
	db    *DB
}

func (t *tester) access(mode Mode, item *shed.IndexItem) error {
	it, err := t.index.Get(*item)
	if err != nil {
		return err
	}
	*item = it
	return nil
}

// update defines set accessors for different modes
func (t *tester) update(b *Batch, mode Mode, item *shed.IndexItem) error {
	if mode != 0 {
		return errors.New("no such mode")
	}
	return t.index.PutInBatch(b.Batch, *item)
}

func newTester(path string) (*tester, error) {
	tester := new(tester)
	sdb, err := shed.NewDB(path)
	if err != nil {
		return nil, err
	}
	tester.db = New(sdb, tester.update, tester.access)
	tester.index, err = sdb.NewIndex("Hash->StoredTimestamp|AccessTimestamp|Data", shed.IndexFuncs{
		EncodeKey: func(fields shed.IndexItem) (key []byte, err error) {
			return fields.Address, nil
		},
		DecodeKey: func(key []byte) (e shed.IndexItem, err error) {
			e.Address = key
			return e, nil
		},
		EncodeValue: func(fields shed.IndexItem) (value []byte, err error) {
			b := make([]byte, 16)
			binary.BigEndian.PutUint64(b[:8], uint64(fields.StoreTimestamp))
			binary.BigEndian.PutUint64(b[8:16], uint64(fields.AccessTimestamp))
			value = append(b, fields.Data...)
			return value, nil
		},
		DecodeValue: func(value []byte) (e shed.IndexItem, err error) {
			e.StoreTimestamp = int64(binary.BigEndian.Uint64(value[:8]))
			e.AccessTimestamp = int64(binary.BigEndian.Uint64(value[8:16]))
			e.Data = value[16:]
			return e, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return tester, nil
}

func TestPutGet(t *testing.T) {
	path, err := ioutil.TempDir("", "rushed-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	tester, err := newTester(path)
	if err != nil {
		t.Fatal(err)
	}
	defer tester.db.Close()
	s := tester.db.Mode(0)
	ch := storage.GenerateRandomChunk(chunk.DefaultSize)
	log.Debug("put")
	err = s.Put(context.Background(), ch)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("get")
	got, err := s.Get(context.Background(), ch.Address())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Data(), ch.Data()) {
		t.Fatal("chunk data mismatch after retrieval")
	}
}

type putter interface {
	Put(context.Context, storage.Chunk) error
}

func (t *tester) Put(_ context.Context, ch storage.Chunk) error {
	return t.index.Put(*(newItemFromChunk(ch)))
}
func BenchmarkPut(b *testing.B) {
	n := 128	
	for j := 0; j < 5; j++ {
		n *= 2
		in := time.Nanosecond
		for i := 0; i < 3; i++ {
			for _, name := range []string{"shed", "rushed"} {
				path, err := ioutil.TempDir("", "rushed-test")
				if err != nil {
					b.Fatal(err)
				}
				defer os.RemoveAll(path)
				tester, err := newTester(path)
				if err != nil {
					b.Fatal(err)
				}
				defer tester.db.Close()
				var db putter
				if name == "shed" {
					db = tester
				} else {
					db = tester.db.Mode(0)
				}
				b.Run(fmt.Sprintf("N=%v Interval=%v, DB=%v", n, in, name), func(t *testing.B) {
					benchmarkPut(t, n, in, db)
				})
			}
			in *= time.Duration(10)
		}
	}
}

func benchmarkPut(b *testing.B, n int, in time.Duration, db putter) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(n)
		for j := 0; j < n; j++ {
			go func() {
				defer wg.Done()
				db.Put(context.Background(), storage.GenerateRandomChunk(chunk.DefaultSize))
			}()
		}
		wg.Wait()
	}
}
