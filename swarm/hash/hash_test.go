package swarmhash

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestHasher(t *testing.T) {
	initTest()
	h := GetHash()
	dh := h.Hash(data)
	if !bytes.Equal(dh, dataHash) {
		t.Fatalf("Expected hash %x, got %x", dataHash, dh)
	}

	l := make([]byte, 8)
	binary.LittleEndian.PutUint64(l, uint64(len(data)))
	dh = h.HashWithLength(l, data)
	if !bytes.Equal(dh, dataHashWithLength) {
		t.Fatalf("Expected hashwithlength %x, got %x", dataHash, dh)
	}
}

func BenchmarkHash(b *testing.B) {
	b.Run("1/32", benchmarkHash)
	b.Run("1/4096", benchmarkHash)
	b.Run("1/1024000", benchmarkHash)

	b.Run("10/32", benchmarkHash)
	b.Run("10/4096", benchmarkHash)
	b.Run("10/1024000", benchmarkHash)

	b.Run("100/32", benchmarkHash)
	b.Run("100/4096", benchmarkHash)
	b.Run("100/1024000", benchmarkHash)
}

func benchmarkHash(b *testing.B) {
	initTest()
	args := strings.Split(b.Name(), "/")
	fmt.Println(args)
	threads, _ := strconv.ParseInt(args[1], 10, 0)
	dataLength, _ := strconv.ParseInt(args[2], 10, 0)
	dataToHash := make([]byte, dataLength)
	rand.Read(dataToHash)
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		for j := 0; j < int(threads); j++ {
			wg.Add(1)
			go func() {
				h := GetHashByName("bar")
				h.Hash(dataToHash)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
