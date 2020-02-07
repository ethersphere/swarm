package bmt

import (
	"fmt"
	"testing"
)

func BenchmarkBMTUsed(t *testing.B) {
	size := 4096
	t.Run(fmt.Sprintf("%v_size_%v", "BMT", size), func(t *testing.B) {
		benchmarkBMT(t, size)
	})
}
