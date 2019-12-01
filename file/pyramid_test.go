package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

// TestPyramidHasherVector executes the file hasher algorithms on serial input data of periods of 0-254
// of lengths defined in common_test.go
func TestPyramidHasherVector(t *testing.T) {
	t.Skip("only provided for easy reference to bug in case chunkSize*129")
	var mismatch int
	for i := start; i < end; i++ {
		eq := true
		dataLength := dataLengths[i]
		log.Info("pyramidvector start", "i", i, "l", dataLength)
		buf, _ := testutil.SerialData(dataLength, 255, 0)
		putGetter := storage.NewHasherStore(&storage.FakeChunkStore{}, storage.MakeHashFunc(storage.BMTHash), false, chunk.NewTag(0, "foo", 0, false))

		ctx := context.Background()
		ref, wait, err := storage.PyramidSplit(ctx, buf, putGetter, putGetter, chunk.NewTag(0, "foo", int64(dataLength/4096+1), false))
		if err != nil {
			t.Fatalf(err.Error())
		}
		err = wait(ctx)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if ref.Hex() != expected[i] {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %s\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, ref, expected[i])
	}

	if mismatch != 1 {
		t.Fatalf("mismatches: %d/%d", mismatch, end-start)
	}
}

// BenchmarkPyramidHasher establishes the benchmark BenchmarkHasher should be compared to
func BenchmarkPyramidHasher(b *testing.B) {

	for i := start; i < end; i++ {
		b.Run(fmt.Sprintf("%d", dataLengths[i]), benchmarkPyramidHasher)
	}
}

func benchmarkPyramidHasher(b *testing.B) {
	params := strings.Split(b.Name(), "/")
	dataLength, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		b.Fatal(err)
	}
	_, data := testutil.SerialData(int(dataLength), 255, 0)
	buf := bytes.NewReader(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Seek(0, io.SeekStart)
		//putGetter := newTestHasherStore(&storage.FakeChunkStore{}, storage.BMTHash)
		putGetter := storage.NewHasherStore(&storage.FakeChunkStore{}, storage.MakeHashFunc(storage.BMTHash), false, chunk.NewTag(0, "foo", 0, false))

		ctx := context.Background()
		_, wait, err := storage.PyramidSplit(ctx, buf, putGetter, putGetter, chunk.NewTag(0, "foo", dataLength/4096+1, false))
		if err != nil {
			b.Fatalf(err.Error())
		}
		err = wait(ctx)
		if err != nil {
			b.Fatalf(err.Error())
		}
	}
}
