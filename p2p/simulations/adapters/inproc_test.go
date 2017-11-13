// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package adapters

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestSocketPipe(t *testing.T) {
	c1, c2, _ := socketPipe()

	done := make(chan struct{})

	go func() {
		msgs := 20
		for i := 0; i < msgs; i++ {
			msg := make([]byte, 8)
			_ = binary.PutUvarint(msg, uint64(i))

			_, err := c1.Write(msg)
			if err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < msgs; i++ {
			msg := make([]byte, 8)
			_ = binary.PutUvarint(msg, uint64(i))

			out := make([]byte, 8)
			_, err := c2.Read(out)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(msg, out) != 0 {
				t.Fatalf("expected %#v, got %#v", msg, out)
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("test timeout")
	}
}

func TestTcpPipe(t *testing.T) {
	c1, c2, _ := tcpPipe()

	done := make(chan struct{})

	go func() {
		msgs := 50
		for i := 0; i < msgs; i++ {
			msg := make([]byte, 1024)
			_ = binary.PutUvarint(msg, uint64(i))

			_, err := c1.Write(msg)
			if err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < msgs; i++ {
			msg := make([]byte, 1024)
			_ = binary.PutUvarint(msg, uint64(i))

			out := make([]byte, 1024)
			_, err := c2.Read(out)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(msg, out) != 0 {
				t.Fatalf("expected %#v, got %#v", msg, out)
			}
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("test timeout")
	}
}

func TestNetPipe(t *testing.T) {
	c1, c2, _ := netPipe()

	done := make(chan struct{})

	go func() {
		msgs := 50
		// netPipe is blocking, so writes are emitted asynchronously
		go func() {
			for i := 0; i < msgs; i++ {
				msg := make([]byte, 1024)
				_ = binary.PutUvarint(msg, uint64(i))

				_, err := c1.Write(msg)
				if err != nil {
					t.Fatal(err)
				}
			}
		}()

		for i := 0; i < msgs; i++ {
			msg := make([]byte, 1024)
			_ = binary.PutUvarint(msg, uint64(i))

			out := make([]byte, 1024)
			_, err := c2.Read(out)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(msg, out) != 0 {
				t.Fatalf("expected %#v, got %#v", msg, out)
			}
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("test timeout")
	}
}
