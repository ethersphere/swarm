package storage

import (
	"bytes"
	//crand "crypto/rand"
	"encoding/binary"
	"io"
	//"math/rand"
	"hash"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/bmt"
	"github.com/ethereum/go-ethereum/swarm/log"
)

func newAsyncHasher() bmt.SectionWriter {
	tp := bmt.NewTreePool(sha3.NewKeccak256, 128, 1)
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
		r.Seek(int64(o), io.SeekStart)
		_, err := fh.WriteBuffer(o, r)
		if err != nil {
			t.Fatal(err)
		}
	}

	batchone := fh.levels[0].getBatch(0)
	if !bytes.Equal(batchone.batchBuffer, data[:8]) {
		t.Fatalf("expected batch one data %x, got %x", data[:8], batchone.batchBuffer)
	}

	batchtwo := fh.levels[0].getBatch(1)
	if !bytes.Equal(batchtwo.batchBuffer, data[8:]) {
		t.Fatalf("expected batch two data %x, got %x", data[8:], batchtwo.batchBuffer)
	}
}

func TestSum(t *testing.T) {

	fh := NewFileHasher(newAsyncHasher, 128, 32)
	dataLength := 2 * fh.ChunkSize()
	data := make([]byte, dataLength)
	//c, err := crand.Read(data)
	//	if err != nil {
	//		t.Fatal(err)
	//	} else if c != len(data) {
	//		t.Fatalf("short read %d", c)
	//	}
	for i := 0; i < len(data); i++ {
		data[i] = byte(i % 256)
	}
	var offsets []int
	for i := 0; i < len(data)/32; i++ {
		offsets = append(offsets, i*32)
	}
	r := bytes.NewReader(data)
	//	for {
	//		if len(offsets) == 0 {
	//			break
	//		}
	//		lastIndex := len(offsets) - 1
	//		var c int
	//		if len(offsets) > 1 {
	//			c = rand.Intn(lastIndex)
	//		}
	//		offset := offsets[c]
	//		if c != lastIndex {
	//			offsets[c] = offsets[lastIndex]
	//		}
	//		offsets = offsets[:lastIndex]
	//
	//		r.Seek(int64(offset), io.SeekStart)
	//		_, err := fh.WriteBuffer(offset, r)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	for i := 0; i < len(offsets); i++ {
		//offset := offsets[i]
		offset := i * 32
		r.Seek(int64(offset), io.SeekStart)
		log.Warn("write", "o", offset)
		c, err := fh.WriteBuffer(offset, r)
		if err != nil {
			t.Fatal(err)
		} else if c < fh.BlockSize() {
			t.Fatalf("short read %d", c)
		}
	}

	hasher := func() hash.Hash {
		return sha3.NewKeccak256()
	}
	rb := bmt.NewRefHasher(hasher, dataLength)
	meta := make([]byte, 8)
	binary.BigEndian.PutUint64(meta, uint64(dataLength))
	res := rb.Hash(data)
	shasher := hasher()
	shasher.Reset()
	shasher.Write(meta)
	shasher.Write(res)
	x := shasher.Sum(nil)

	time.Sleep(time.Second)
	t.Logf("hash ref raw: %x", res)
	t.Logf("hash ref dosum: %x", x)
	fh.SetLength(int64(dataLength))
	h := fh.Sum(nil)
	t.Logf("hash: %x", h)
}
