package storage

import (
	"bytes"
	crand "crypto/rand"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

func newAsyncHasher() bmt.SectionWriter {
	tp := bmt.NewTreePool(sha3.NewKeccak256, 128*128, 32)
	h := bmt.New(tp)
	return h.NewAsyncWriter(false)
}

func TestLevelFromOffset(t *testing.T) {
	fh := NewFileHasher(newAsyncHasher, 128, 32)
	sizes := []int{64, 127, 128, 129, 128*128 - 1, 128 * 128, 128 * 128 * 128 * 20}
	expects := []int{0, 0, 1, 1, 1, 2, 3}
	for i, sz := range sizes {
		offset := fh.ChunkSize() * sz
		lvl := fh.OffsetToLevelDepth(int64(offset))
		if lvl != expects[i] {
			t.Fatalf("offset %d (chunkcount %d), expected level %d, got %d", offset, sz, expects[i], lvl)
		}
	}
}

func TestWriteBuffer(t *testing.T) {
	data := []byte("0123456789abcdef")
	fh := NewFileHasher(newAsyncHasher, 2, 2)
	offsets := []int{12, 8, 4, 2, 6, 10, 0, 14}
	r := bytes.NewReader(data)
	for _, o := range offsets {
		log.Debug("writing", "o", o)
		r.Seek(int64(o), io.SeekStart)
		_, err := fh.WriteBuffer(o, r)
		if err != nil {
			t.Fatal(err)
		}
		//copy(buf, data[o:o+2])
	}

	batchone := fh.levels[0].getBatch(0)
	if !bytes.Equal(batchone.batchBuffer, data[:8]) {
		t.Fatalf("expected batch one data %x, got %x", data[:8], batchone.batchBuffer)
	}

	batchtwo := fh.levels[0].getBatch(1)
	if !bytes.Equal(batchtwo.batchBuffer, data[8:]) {
		t.Fatalf("expected batch two data %x, got %x", data[8:], batchtwo.batchBuffer)
	}

	time.Sleep(time.Second)
}

func TestSum(t *testing.T) {

	fh := NewFileHasher(newAsyncHasher, 128, 32)
	//data := make([]byte, 258*fh.ChunkSize())
	data := make([]byte, 128*fh.ChunkSize())
	c, err := crand.Read(data)
	if err != nil {
		t.Fatal(err)
	} else if c != len(data) {
		t.Fatalf("short read %d", c)
	}

	var offsets []int
	for i := 0; i < len(data)/32; i++ {
		offsets = append(offsets, i*32)
	}
	r := bytes.NewReader(data)
	for {
		if len(offsets) == 0 {
			break
		}
		lastIndex := len(offsets) - 1
		var c int
		if len(offsets) > 1 {
			c = rand.Intn(lastIndex)
		}
		offset := offsets[c]
		if c != lastIndex {
			offsets[c] = offsets[lastIndex]
		}
		offsets = offsets[:lastIndex]

		r.Seek(int64(offset), io.SeekStart)
		_, err := fh.WriteBuffer(offset, r)
		if err != nil {
			t.Fatal(err)
		}
	}
	fh.SetLength(int64(len(data)))
	h := fh.Sum(nil)
	t.Logf("hash: %x", h)

}
