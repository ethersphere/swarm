// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
package api

import (
	"context"
	"encoding/hex"
	"fmt"

	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

const (
	PinVersion     = "1.0"
	DONT_PIN       = 0
	WorkerChanSize = 8
)

var PinApiInstance *PinApi
var once sync.Once

type PinApi struct {
	db         *localstore.DB
	api        *API
	fileParams *storage.FileStoreParams
	tag        *chunk.Tags
}

func NewPinApi(lstore *localstore.DB, params *storage.FileStoreParams, tags *chunk.Tags) *PinApi {
	pinApi := &PinApi{
		db:         lstore,
		fileParams: params,
		tag:        tags,
	}
	once.Do(func() {
		PinApiInstance = pinApi
	})

	return pinApi
}

func (p *PinApi) SetApi(api *API) {
	p.api = api
}

func (p *PinApi) ShowDatabase() string {
	p.db.ShowDatabaseInformation()
	return "Check the swarm log file for the output"
}

func (p *PinApi) AddPinFile(hash []byte, isRaw bool) error {
	return p.db.AddToPinFileIndex(hash, isRaw)
}

func (p *PinApi) ListPinFiles() {
	p.db.ListPinnedFiles()
}

func (p *PinApi) UnpinFiles(rootHash string, credentials string) {
	p.showChunksOfRootHash(rootHash, credentials, true)
}

func (p *PinApi) WalkPinnedChunks(rootHash string, credentials string) {
	p.showChunksOfRootHash(rootHash, credentials, false)
}

func (p *PinApi) showChunksOfRootHash(rootHash string, credentials string, unPin bool) {

	fileWorkers := make(chan storage.Reference, WorkerChanSize)
	chunkWorkers := make(chan storage.Reference, WorkerChanSize)

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return
	}
	hashSize := len(addr)
	isEncrypted := len(addr) > hashFunc().Size()
	tag := chunk.NewTag(0, "show-chunks-tag", 0)
	getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, tag, DONT_PIN)

	// If the file is not raw.. then this file is a manifest
	// Manifests needs to be parsed and for each file this needs to print its chunks
	// If the file is Raw.. then only this file's chunks needs to be printed
	raw, err := p.db.IsPinnedFileRaw(addr)
	if err != nil {
		log.Error("Could not find root hash in pinFilesIndex" + err.Error())
		return
	}

	if !raw {

		// If it not a raw file... load the manifest and process the files inside one by one
		walker, err := p.api.NewManifestWalker(context.TODO(), storage.Address(addr),
			p.api.Decryptor(context.TODO(), credentials), nil)

		if err != nil {
			log.Error("Could not decode manifest. Reason: " + err.Error())
			return
		}

		err = walker.Walk(func(entry *ManifestEntry) error {

			fileAddr, err := hex.DecodeString(entry.Hash)
			if err != nil {
				log.Error("Error decoding hash present in manifest" + err.Error())
				return err
			}

			// send the file to file workers
			fileWorkers <- storage.Reference(fileAddr)
			return nil
		})

		if err != nil {
			log.Error("Error walking manifest. Reason: " + err.Error())
			return
		}

		// Finally, remove the manifest file too
		fileWorkers <- storage.Reference(addr)

		// Signal end of file stream
		fileWorkers <- storage.Reference(nil)

	} else {
		// Its a raw file.. no manifest.. so process only this hash
		fileWorkers <- storage.Reference(addr)

		// Singal end of file stream
		fileWorkers <- storage.Reference(nil)
	}

	doneFileWorker := make(chan struct{})
QuitFileFor:
	for {
		select {
		case <-doneFileWorker:
			break QuitFileFor

		case fileRef := <-fileWorkers:

			log.Info("UnPinning file from manifest", "Address", fmt.Sprintf("%0x", fileRef))

			if fileRef == nil {
				close(doneFileWorker)
				break
			}

			// Send the file to chunk workers
			chunkWorkers <- fileRef

			// See if the root chunk is pinned
			// If YES, then remember to remove the root chunk from the pinFIlesIndex,
			// once all chunks are removed from pinIndex
			isRootChunkPinned := p.db.IsChunkPinned(fileRef)

			actualFileSize := uint64(0)
			rcvdFileSize := uint64(0)
			doneChunkWorker := make(chan struct{})

		QuitChunkFor:
			for {
				select {
				case <-doneChunkWorker:
					break QuitChunkFor

				case ref := <-chunkWorkers:


					go func() {

						chunkData, err := getter.Get(context.TODO(), ref)
						if err != nil {
							log.Error("Error getting chunk data from localstore.")
							close(doneChunkWorker)
						}

						datalen := len(chunkData)
						if datalen < 9 {
							log.Error("Invalid chunk data from localstore.")
							close(doneChunkWorker)
						}

						subTreeSize := chunkData.Size()
						if actualFileSize < subTreeSize {
							actualFileSize = subTreeSize
							log.Info("File size ", "Size", actualFileSize)
						}

						if subTreeSize > chunk.DefaultSize {
							branches := (datalen - 8) / hashSize
							if unPin {
								err = p.db.UnpinChunk(ref)
								if err != nil {
									// TODO: if this happens, we should go back and revert the entire file's chunks
									log.Error("Could not unpin chunk. Addres " + fmt.Sprintf("%0x", ref))
								} else {
									log.Debug("Removing tree chunk", "Address", fmt.Sprintf("%0x", ref),
										"Branches", branches, "SubTreeSize", subTreeSize)
								}
							} else {
								log.Info("Tree chunk", "Address", fmt.Sprintf("%0x", ref),
									"Branches", branches, "SubTreeSize", subTreeSize)
							}
							for i := 0; i < branches; i++ {
								brAddr := make([]byte, hashSize)
								start := (i * hashSize) + 8
								end := ((i + 1) * hashSize) + 8
								copy(brAddr[:], chunkData[start:end])
								chunkWorkers <- storage.Reference(brAddr)
							}

						} else {
							if unPin {
								err := p.db.UnpinChunk(ref)
								if err != nil {
									// TODO: if this happens, we should go back and revert the entire file's chunks
									log.Error("Could not unpin chunk. Addres " + fmt.Sprintf("%0x", ref))
								} else {
									log.Debug("Removing data chunk", "Address", fmt.Sprintf("%0x", ref),
										"SubTreeSize", subTreeSize)
								}
							} else {
								log.Info("Data chunk", "Address", fmt.Sprintf("%0x", ref),
									"SubTreeSize", subTreeSize)
							}
							rcvdFileSize = rcvdFileSize + chunk.DefaultSize
							if rcvdFileSize > actualFileSize {
								close(doneChunkWorker)
							}
						}

					}()
				}
			}

			if unPin && !isRootChunkPinned {
				p.db.UnpinRootHash(fileRef)
				if err != nil {
					// TODO: if this happens, we should go back and revert the entire file's chunks
					log.Error("Could not unpin root chunk. Addres " + fmt.Sprintf("%0x", fileRef))
				}
			}
		}
	}
}
