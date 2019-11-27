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
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/testutil"
)

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
