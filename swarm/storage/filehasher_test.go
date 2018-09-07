package storage

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

var pool *bmt.TreePool

func init() {
	pool = bmt.NewTreePool(sha3.NewKeccak256, 128, bmt.PoolSize*32)
}

func newAsyncHasher() bmt.SectionWriter {
	h := bmt.New(pool)
	return h.NewAsyncWriter(false)
}

func newSerialData(l int) ([]byte, error) {
	data := make([]byte, l)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 255)
	}
	return data, nil
}

func newRandomData(l int) ([]byte, error) {
	data := make([]byte, l)
	c, err := crand.Read(data)
	if err != nil {
		return nil, err
	} else if c != len(data) {
		return nil, fmt.Errorf("short read (%d)", c)
	}
	return data, nil
}

func TestSum(t *testing.T) {

	var mismatch int
	dataFunc := newSerialData
	chunkSize := 128 * 32
	dataLengths := []int{31, 32, 33, 63, 64, 65, chunkSize, chunkSize + 31, chunkSize + 32, chunkSize + 63, chunkSize + 64, chunkSize * 2, chunkSize*2 + 32, chunkSize * 128, chunkSize*128 + 31, chunkSize*128 + 32, chunkSize * 129, chunkSize * 130, chunkSize * 128 * 128}

	for _, dl := range dataLengths {
		chunks := dl / chunkSize
		log.Debug("testing", "c", chunks, "s", dl%chunkSize)
		fh := NewFileHasher(newAsyncHasher, 128, 32)
		data, err := dataFunc(dl)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < len(data); i += 32 {
			max := i + 32
			if len(data) < max {
				max = len(data)
			}
			_, err := fh.WriteBuffer(i, data[i:max])
			if err != nil {
				t.Fatal(err)
			}
		}

		time.Sleep(time.Second * 1)
		fh.SetLength(int64(dl))
		h := fh.Sum(nil)

		p, err := referenceHash(data)

		if err != nil {
			t.Fatalf(err.Error())
		}
		eq := bytes.Equal(p, h)
		if !eq {
			mismatch++
		}
		t.Logf("[%3d + %2d]\t%v\t%x\t%x", chunks, dl%chunkSize, eq, p, h)
		t.Logf("[%3d + %2d]\t%x", chunks, dl%chunkSize, h)
	}
	if mismatch > 0 {
		t.Fatalf("%d/%d mismatches", mismatch, len(dataLengths))
	}
}

func referenceHash(data []byte) ([]byte, error) {
	//return []byte{}, nil
	putGetter := newTestHasherStore(&fakeChunkStore{}, BMTHash)
	p, _, err := PyramidSplit(context.TODO(), io.LimitReader(bytes.NewReader(data), int64(len(data))), putGetter, putGetter)
	return p, err
}
