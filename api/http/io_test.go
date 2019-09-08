package http

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/testutil"
)

func BenchmarkDownload(b *testing.B) {
	for _, peek := range []int{1} {
		for _, bufSize := range []int{4096, 16 * 4096} {
			for _, size := range []int{4096 * 32} {
				for _, delay := range []int{500, 1000} {
					Delay = delay
					b.Run(fmt.Sprintf("buf=%v,size=%v,delay=%v,la=%v", bufSize, size, delay, peek), func(b *testing.B) {
						for n := 0; n < b.N; n++ {
							benchmarkDownload(b, bufSize, size, peek)
						}
					})
				}
			}
		}
	}
}

func benchmarkDownload(b *testing.B, bufSize int, expLen int, peek int) {
	defer func(peek int) {
		PeekSize = peek
	}(PeekSize)
	PeekSize = peek
	defer func(bufSize int) {
		BufferSize = bufSize
	}(BufferSize)
	BufferSize = bufSize

	mode := Baseline
	if peek > 0 {
		mode = Peeker
	} else if bufSize > 0 {
		mode = Buffered
	}
	defer func(mode int) {
		Mode = mode
	}(Mode)
	Mode = mode
	srv, err := newTestSwarmServer(serverFunc, nil, nil)
	if err != nil {
		b.Fatal(err)
	}
	defer srv.Close()

	r := testutil.RandomReader(3, expLen)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ref, wait, err := srv.FileStore.Store(ctx, r, int64(expLen), false)
	wait(ctx)
	log.Warn("uploaded data", "ref", hex.EncodeToString(ref))

	getBzzURL := fmt.Sprintf("%s/bzz-raw:/%s", srv.URL, ref)

	b.StartTimer()
	getResp, err := http.Get(getBzzURL)
	if err != nil {
		b.Fatal(err)
	}
	if getResp.StatusCode != http.StatusOK {
		b.Fatalf("err %s", getResp.Status)
	}
	if err != nil {
		b.Fatal(err)
	}
	var read int
	buf := make([]byte, 10000)
	defer getResp.Body.Close()
	for {
		n, err := getResp.Body.Read(buf)
		read += n
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			b.Fatalf("unexpected error %v", err)
		}
	}
	// read, err = io.ReadFull(getResp.Body, buf)
	b.StopTimer()
	if err != nil && err != io.ErrUnexpectedEOF {
		b.Fatal(err)
	}

	if read != expLen {
		b.Fatalf("expected %v, got %v", expLen, read)
	}
}
