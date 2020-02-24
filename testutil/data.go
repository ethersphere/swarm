package testutil

import (
	"bytes"
	"io"
)

func SerialData(l int, mod int, offset int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	for i := 0; i < len(slice); i++ {
		slice[i] = byte((i + offset) % mod)
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}
