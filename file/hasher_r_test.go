package file

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ethersphere/swarm/bmt"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/testutil"
	"golang.org/x/crypto/sha3"
)

func TestReferenceFileHasher(t *testing.T) {
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, 128, bmt.PoolSize)
	h := bmt.New(pool)
	var mismatch int
	for i := start; i < end; i++ {
		dataLength := dataLengths[i]
		log.Info("start", "i", i, "len", dataLength)
		fh := NewReferenceFileHasher(h, 128)
		_, data := testutil.SerialData(dataLength, 255, 0)
		refHash := fh.Hash(bytes.NewReader(data), len(data)).Bytes()
		eq := true
		if expected[i] != fmt.Sprintf("%x", refHash) {
			mismatch++
			eq = false
		}
		t.Logf("[%7d+%4d]\t%v\tref: %x\texpect: %s", dataLength/chunkSize, dataLength%chunkSize, eq, refHash, expected[i])
	}
	if mismatch > 0 {
		t.Fatalf("mismatches: %d/%d", mismatch, end-start)
	}
}

func BenchmarkReferenceHasher(b *testing.B) {
	for i := start; i < end; i++ {
		b.Run(fmt.Sprintf("%d", dataLengths[i]), benchmarkReferenceFileHasher)
	}
}

func benchmarkReferenceFileHasher(b *testing.B) {
	params := strings.Split(b.Name(), "/")
	dataLength, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		b.Fatal(err)
	}
	_, data := testutil.SerialData(int(dataLength), 255, 0)
	pool := bmt.NewTreePool(sha3.NewLegacyKeccak256, 128, bmt.PoolSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := bmt.New(pool)
		fh := NewReferenceFileHasher(h, 128)
		fh.Hash(bytes.NewReader(data), len(data)).Bytes()
	}
}
