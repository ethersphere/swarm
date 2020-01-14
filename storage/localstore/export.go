// Copyright 2019 The go-ethereum Authors
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

package localstore

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/shed"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	// filename in tar archive that holds the information
	// about exported data format version
	exportVersionFilename = ".swarm-export-version"
	// legacy version for previous LDBStore
	legacyExportVersion = "1"
	// current export format version
	currentExportVersion = "2"
	// tags
	tagsFilenamePrefix  = "tags-"
	exportTagsFileLimit = 1000
)

// Export writes a tar structured data to the writer of
// all chunks in the retrieval data index. It returns the
// number of chunks exported.
func (db *DB) Export(w io.Writer) (count int64, err error) {
	tw := tar.NewWriter(w)
	defer tw.Close()

	if err := tw.WriteHeader(&tar.Header{
		Name: exportVersionFilename,
		Mode: 0644,
		Size: int64(len(currentExportVersion)),
	}); err != nil {
		return 0, err
	}
	if _, err := tw.Write([]byte(currentExportVersion)); err != nil {
		return 0, err
	}

	// tags export
	var (
		tagsFileNumber int
		tagsCounter    int
		tags           []byte

		writeTags = func() (err error) {
			l := len(tags)
			if l == 0 {
				return nil
			}

			tagsCounter = 0
			tagsFileNumber++

			if err := tw.WriteHeader(&tar.Header{
				Name: fmt.Sprintf("%s%v", tagsFilenamePrefix, tagsFileNumber),
				Mode: 0644,
				Size: int64(l),
			}); err != nil {
				return err
			}

			_, err = tw.Write(tags)
			return err
		}
	)
	err = db.pinIndex.Iterate(func(item shed.Item) (stop bool, err error) {
		tags = append(tags, encodeExportPin(item.Address, item.PinCounter)...)
		tags = append(tags, '\n')
		if tagsCounter == exportTagsFileLimit {
			if err := writeTags(); err != nil {
				return true, err
			}
		}
		return false, nil
	}, nil)
	if err != nil {
		return 0, err
	}
	if err := writeTags(); err != nil {
		return 0, err
	}

	exportchunk := func(addr chunk.Address, data []byte) error {
		hdr := &tar.Header{
			Name: hex.EncodeToString(addr),
			Mode: 0644,
			Size: int64(len(data)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
		count++
		return nil
	}

	// Export legacy (pre fcds) data index.
	// This is required as a manual step in migrateDiwali migration.
	err = db.retrievalDataIndex.Iterate(func(item shed.Item) (stop bool, err error) {
		err = exportchunk(item.Address, item.Data)
		return false, err
	}, nil)
	if err != nil {
		return 0, err
	}

	err = db.data.Iterate(func(ch chunk.Chunk) (stop bool, err error) {
		err = exportchunk(ch.Address(), ch.Data())
		return false, err
	})

	return count, err
}

// Import reads a tar structured data from the reader and
// stores chunks in the database. It returns the number of
// chunks imported.
func (db *DB) Import(r io.Reader, legacy bool) (count int64, err error) {
	tr := tar.NewReader(r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errC := make(chan error)
	doneC := make(chan struct{})
	tokenPool := make(chan struct{}, 100)
	var wg sync.WaitGroup
	go func() {
		var (
			// if exportVersionFilename file is not present
			// assume legacy version
			version = legacyExportVersion
		)
		for {
			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			// get the export file format version
			if hdr.Name == exportVersionFilename {
				data, err := ioutil.ReadAll(tr)
				if err != nil {
					select {
					case errC <- err:
					case <-ctx.Done():
					}
				}
				version = string(data)
				continue
			}
			// set pinned chunks
			if strings.HasPrefix(hdr.Name, tagsFilenamePrefix) {
				// All chunks are put before tag files are iterated on
				// because of tagsFilenamePrefix starts with "t"
				// which is ordered later in the tar file then
				// hex characters of chunk addresses.
				//
				// Wait for chunks to be stored before continuing.
				wg.Wait()

				scanner := bufio.NewScanner(tr)
				batch := new(leveldb.Batch)
				for scanner.Scan() {
					addr, counter := decodeExportPin(scanner.Bytes())
					if addr == nil {
						continue
					}
					if err := db.setPin(batch, addr, counter); err != nil {
						select {
						case errC <- err:
						case <-ctx.Done():
						}
					}
				}

				if err := scanner.Err(); err != nil {
					select {
					case errC <- err:
					case <-ctx.Done():
					}
				}

				if err := db.shed.WriteBatch(batch); err != nil {
					select {
					case errC <- err:
					case <-ctx.Done():
					}
				}
				continue
			}

			if len(hdr.Name) != 64 {
				log.Warn("ignoring non-chunk file", "name", hdr.Name)
				continue
			}

			keybytes, err := hex.DecodeString(hdr.Name)
			if err != nil {
				log.Warn("ignoring invalid chunk file", "name", hdr.Name, "err", err)
				continue
			}

			data, err := ioutil.ReadAll(tr)
			if err != nil {
				select {
				case errC <- err:
				case <-ctx.Done():
				}
			}
			key := chunk.Address(keybytes)

			var ch chunk.Chunk
			switch version {
			case legacyExportVersion:
				// LDBStore Export exported chunk data prefixed with the chunk key.
				// That is not necessary, as the key is in the chunk filename,
				// but backward compatibility needs to be preserved.
				ch = chunk.NewChunk(key, data[32:])
			case currentExportVersion:
				ch = chunk.NewChunk(key, data)
			default:
				select {
				case errC <- fmt.Errorf("unsupported export data version %q", version):
				case <-ctx.Done():
				}
			}
			tokenPool <- struct{}{}
			wg.Add(1)

			go func() {
				_, err := db.Put(ctx, chunk.ModePutUpload, ch)
				select {
				case errC <- err:
				case <-ctx.Done():
					wg.Done()
					<-tokenPool
				default:
					_, err := db.Put(ctx, chunk.ModePutUpload, ch)
					if err != nil {
						errC <- err
					}
					wg.Done()
					<-tokenPool
				}
			}()

			count++
		}
		wg.Wait()
		close(doneC)
	}()

	// wait for all chunks to be stored
	for {
		select {
		case err := <-errC:
			if err != nil {
				return count, err
			}
		case <-ctx.Done():
			return count, ctx.Err()
		default:
			select {
			case <-doneC:
				return count, nil
			default:
			}
		}
	}
}

func encodeExportPin(addr chunk.Address, counter uint64) (data []byte) {
	data = make([]byte, 8, 8+len(addr))
	binary.BigEndian.PutUint64(data[:8], counter)
	data = append(data, addr...)
	return data
}

func decodeExportPin(data []byte) (addr chunk.Address, counter uint64) {
	if len(data) < 8 {
		return nil, 0
	}
	counter = binary.BigEndian.Uint64(data[:8])
	addr = chunk.Address(data[8:])
	return addr, counter
}
