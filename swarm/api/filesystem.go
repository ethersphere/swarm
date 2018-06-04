// Copyright 2016 The go-ethereum Authors
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

package api

import (
	"bufio"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/swarm/storage"
)

const maxParallelFiles = 5

type FileSystem struct {
	api *Api
}

func NewFileSystem(api *Api) *FileSystem {
	return &FileSystem{api}
}

func retrieveToFile(dpa *storage.DPAAPI, addr storage.Address, path string) error {
	f, err := os.Create(path) // TODO: basePath separators
	if err != nil {
		return err
	}
	reader, _ := fileStore.Retrieve(addr)
	writer := bufio.NewWriter(f)
	size, err := reader.Size()
	if err != nil {
		return err
	}
	if _, err = io.CopyN(writer, reader, size); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return f.Close()
}
