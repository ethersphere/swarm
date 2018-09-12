package storage

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
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

func newSerialData(l int, offset int) ([]byte, error) {
	data := make([]byte, l)
	for i := 0; i < len(data); i++ {
		data[i] = byte((i + offset) % 255)
	}
	return data, nil
}

// offset doesn't matter here
func newRandomData(l int, offset int) ([]byte, error) {
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
	chunkSize := 128 * 32
	serialOffset := 0
	//dataLengths := []int{31, 32, 33, 63, 64, 65, chunkSize, chunkSize + 31, chunkSize + 32, chunkSize + 63, chunkSize + 64, chunkSize * 2, chunkSize*2 + 32, chunkSize * 128, chunkSize*128 + 31, chunkSize*128 + 32, chunkSize*128 + 64, chunkSize * 129, chunkSize * 130, chunkSize * 128 * 128}
	dataLengths := []int{chunkSize * 2}

	for _, dl := range dataLengths {
		chunks := dl / chunkSize
		log.Debug("testing", "c", chunks, "s", dl%chunkSize)
		fh := NewFileHasher(newAsyncHasher, 128, 32)
		data, err := newSerialData(dl, serialOffset)
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

func TestAnomaly(t *testing.T) {

	correctData := []byte{48, 71, 216, 65, 7, 120, 152, 194, 107, 190, 107, 230, 82, 162, 236, 89, 10, 93, 155, 215, 205, 69, 210, 144, 234, 66, 81, 27, 72, 117, 60, 9, 129, 179, 29, 154, 127, 108, 55, 117, 35, 232, 118, 157, 176, 33, 9, 29, 242, 62, 221, 159, 215, 189, 107, 205, 241, 26, 34, 245, 24, 219, 96, 6}
	correctHex := "b8e1804e37a064d28d161ab5f256cc482b1423d5cd0a6b30fde7b0f51ece9199"
	var dataLength uint64 = 4096*128 + 4096

	data := make([]byte, dataLength)
	for i := uint64(0); i < dataLength; i++ {
		data[i] = byte(i % 255)
	}

	leftChunk := make([]byte, 4096)

	h := bmt.New(pool)
	meta := make([]byte, 8)
	binary.LittleEndian.PutUint64(meta, 4096)
	for i := 0; i < 128; i++ {
		h.ResetWithLength(meta)
		h.Write(data[i*4096 : i*4096+4096])
		copy(leftChunk[i*32:], h.Sum(nil))
	}
	binary.LittleEndian.PutUint64(meta, 4096*128)
	h.ResetWithLength(meta)
	h.Write(leftChunk)
	leftChunkHash := h.Sum(nil)
	t.Logf("%x %v %v", leftChunkHash, bytes.Equal(correctData[:32], leftChunkHash), meta)

	binary.LittleEndian.PutUint64(meta, 4096)
	h.ResetWithLength(meta)
	h.Write(data[4096*128:])
	rightChunkHash := h.Sum(nil)
	t.Logf("%x %v %v", rightChunkHash, bytes.Equal(correctData[32:], rightChunkHash), meta)

	binary.LittleEndian.PutUint64(meta, dataLength)
	h.ResetWithLength(meta)
	h.Write(leftChunkHash)
	h.Write(rightChunkHash)
	resultHex := fmt.Sprintf("%x", h.Sum(nil))
	t.Logf("%v %v %v", resultHex, resultHex == correctHex, meta)
}

func TestReferenceFileHasher(t *testing.T) {
	h := bmt.New(pool)
	var mismatch int
	chunkSize := 128 * 32
	dataLengths := []int{31, 32, 33, 63, 64, 65, chunkSize, chunkSize + 31, chunkSize + 32, chunkSize + 63, chunkSize + 64, chunkSize * 2, chunkSize*2 + 32, chunkSize * 128, chunkSize*128 + 31, chunkSize*128 + 32, chunkSize*128 + 64, chunkSize * 129} //, chunkSize * 130, chunkSize * 128 * 128}
	//dataLengths := []int{31}
	for _, dataLength := range dataLengths {
		fh := NewReferenceFileHasher(h, 128)
		data, _ := newSerialData(dataLength, 0)
		refHash := fh.Hash(bytes.NewReader(data), len(data)).Bytes()

		pyramidHash, err := referenceHash(data)
		if err != nil {
			t.Fatalf(err.Error())
		}

		eq := bytes.Equal(pyramidHash, refHash)
		if !eq {
			mismatch++
		}
		t.Logf("[%7d+%4d]\tref: %x\tpyr: %x", dataLength/chunkSize, dataLength%chunkSize, refHash, pyramidHash)
	}
	if mismatch > 0 {
		t.Fatalf("failed have %d mismatch", mismatch)
	}
}
