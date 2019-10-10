// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package langos

import "bufio"

// BufferedReader wraps bufio.Reader to expose Seek method
// from the provided Reader in NewBufferedReader.
type BufferedReader struct {
	Reader
	r *bufio.Reader
}

// NewBufferedReader creates a new instance of BufferedReader,
// out of Reader. Argument `size` is the size of the read buffer.
func NewBufferedReader(reader Reader, size int) BufferedReader {
	return BufferedReader{
		Reader: reader,
		r:      bufio.NewReaderSize(reader, size),
	}
}

func (b BufferedReader) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}

func (b BufferedReader) Seek(offset int64, whence int) (int64, error) {
	n, err := b.Reader.Seek(offset, whence)
	b.r.Reset(b.Reader)
	return n, err
}
