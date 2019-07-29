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

package pin

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"sync"

	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

const (
	Version            = "1.0"
	WorkerChanSize     = 8             // Max no of goroutines when walking the file tree
	SwarmPinHeaderName = "x-swarm-pin" // Presence of this in header indicates pinning required
)

// FileInfo is the struct that stores the information about pinned files
// A map of this is stored in the state DB
type FileInfo struct {
	isRaw      bool
	fileSize   uint64
	pinCounter uint64
}

// API is the main object which implements all things pinning.
type API struct {
	db          *localstore.DB
	api         *api.API
	fileParams  *storage.FileStoreParams
	tag         *chunk.Tags
	hashSize    int
	state       state.Store         // the state store used to store info about pinned files
	pinnedFiles map[string]FileInfo // stores the root hashes and other info. about the pinned files
}

// NewAPI creates a API object that is required for pinning and unpinning
func NewApi(lstore *localstore.DB, stateStore state.Store, params *storage.FileStoreParams, tags *chunk.Tags, api *api.API) *API {
	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	pinnedFiles := make(map[string]FileInfo)
	err := loadPinnedFilesInfo(pinnedFiles, stateStore)
	if err != nil {
		log.Error("Error loading pinned files from state store.", "err", err)
		return nil
	}
	return &API{
		db:          lstore,
		api:         api,
		fileParams:  params,
		tag:         tags,
		hashSize:    hashFunc().Size(),
		state:       stateStore,
		pinnedFiles: pinnedFiles,
	}
}

// PinFiles is used to pin a RAW file or a collection (which hash manifest's)
// to the local Swarm node. It takes the root hash as the argument and walks
// down the merkle tree and pin all the chunks that are encountered on the
// way. It pins both data chunk and tree chunks. The pre-requisite is that
// the file should be present in the local database. This function is called
// from two places 1) Just after the file is uploaded 2) anytime after
// uploading the file using the pin command. This function can pin both
// encrypted and non-encrypted files.
func (p *API) PinFiles(rootHash string, isRaw bool, credentials string) error {
	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "rootHash", rootHash, "err", err)
		return err
	}

	hasChunk, err := p.db.Has(context.TODO(), chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
	if !hasChunk {
		log.Error("Could not pin hash. File not uploaded", "Hash", rootHash)
		return err
	}

	// Walk the root hash and pin all the chunks
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		err := p.db.Set(context.TODO(), chunk.ModeSetPin, chunkAddr)
		if err != nil {
			log.Error("Could not pin chunk. Address "+"Address", hex.EncodeToString(chunkAddr))
			return err
		} else {
			log.Trace("Pinning chunk", "Address", hex.EncodeToString(chunkAddr))
		}
		return nil
	}
	p.walkChunksFromRootHash(rootHash, isRaw, credentials, walkerFunction)

	// Check if the root hash is already pinned and add it to the fileInfo struct
	if val, ok := p.pinnedFiles[rootHash]; !ok {
		// Get the file size from the root chunk first 8 bytes
		hashFunc := storage.MakeHashFunc(storage.DefaultHash)
		isEncrypted := len(addr) > hashFunc().Size()
		getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, chunk.NewTag(0, "show-chunks-tag", 0))
		chunkData, err := getter.Get(context.TODO(), addr)
		if err != nil {
			log.Error("Error getting chunk data from localstore.", "Address", hex.EncodeToString(addr))
			return nil
		}
		fileSize := chunkData.Size()

		// Get the pin counter from the pinIndex
		pinCounter, err := p.getPinCounterOfChunk(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
		if err != nil {
			log.Error("Error getting pin counter of root hash.", "rootHash", rootHash, "err", err)
			return nil
		}

		// Store it in the fileinfo data structure
		// this is pushed to state DB when Close() is called
		p.pinnedFiles[rootHash] = FileInfo{
			isRaw:      isRaw,
			fileSize:   fileSize,
			pinCounter: pinCounter,
		}
	} else {
		// Get the pin counter from the pinIndex
		pinCounter, err := p.getPinCounterOfChunk(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
		if err != nil {
			log.Error("Error getting pin counter of root hash.", "rootHash", rootHash, "err", err)
			return nil
		}

		val.pinCounter = pinCounter
		p.pinnedFiles[rootHash] = val
	}

	// Store the pinned files in state DB
	err = storePinnedFilesInfo(p.pinnedFiles, p.state)
	if err != nil {
		log.Error("Error pinning file.", "rootHash", rootHash, "err", err)
		return nil
	}

	log.Debug("File pinned", "Address", rootHash)
	return nil
}

// UnPinFiles is used to unpin an already pinned file. It takes the root
// hash of the file and walks down the merkle tree unpinning all the chunks
// that are encountered on the way. The pre-requisite is that the file should
// have been already pinned using the PinFiles function. This function can
// be called only from an external command.
func (p *API) UnpinFiles(rootHash string, credentials string) error {
	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "rootHash", rootHash, "err", err)
		return err
	}

	fileInfo, ok := p.pinnedFiles[rootHash]
	if !ok {
		log.Error("Root hash is not pinned", "rootHash", rootHash)
		return err
	}

	// Walk the root hash and unpin all the chunks
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		err := p.db.Set(context.TODO(), chunk.ModeSetUnpin, chunkAddr)
		if err != nil {
			log.Error("Could not unpin chunk", "Address", hex.EncodeToString(chunkAddr))
			return err
		} else {
			log.Trace("Unpinning chunk", "Address", hex.EncodeToString(chunkAddr))
		}
		return nil
	}
	p.walkChunksFromRootHash(rootHash, fileInfo.isRaw, credentials, walkerFunction)

	// Update the state DB
	pinCounter, err := p.getPinCounterOfChunk(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
	if err != nil {
		delete(p.pinnedFiles, rootHash)

		// Store the pinned files in state DB
		err = storePinnedFilesInfo(p.pinnedFiles, p.state)
		if err != nil {
			log.Error("Error unpinning file.", "rootHash", rootHash, "err", err)
			return nil
		}
	} else {
		fileInfo.pinCounter = pinCounter
		p.pinnedFiles[rootHash] = fileInfo
	}
	log.Debug("File unpinned", "Address", rootHash)
	return nil
}

// ListPinFiles functions logs information of all the files that are pinned
// in the current local node. It displays the root hash of the pinned file
// or collection. It also display three vital information's
//     1) Whether the file is a RAW file or not
//     2) Size of the pinned file or collection
//     3) the number of times that particular file or collection is pinned.
func (p *API) ListPinFiles() map[string]FileInfo {
	for hash, fileInfo := range p.pinnedFiles {
		log.Info("Pinned file", "Address", hash, "IsRAW", fileInfo.isRaw,
			"fileSize", fileInfo.fileSize, "pinCounter", fileInfo.fileSize)
	}
	return p.pinnedFiles
}

func (p *API) walkChunksFromRootHash(rootHash string, isRaw bool, credentials string, executeFunc func(storage.Reference) error) {
	fileWorkers := make(chan storage.Reference, WorkerChanSize)
	chunkWorkers := make(chan storage.Reference, WorkerChanSize)

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "err", err)
		return
	}

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	hashSize := len(addr)
	isEncrypted := len(addr) > hashFunc().Size()
	getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, chunk.NewTag(0, "show-chunks-tag", 0))

	go func() {
		if !isRaw {

			// If it is not a raw file... load the manifest and process the files inside one by one
			walker, err := p.api.NewManifestWalker(context.TODO(), storage.Address(addr),
				p.api.Decryptor(context.TODO(), credentials), nil)

			if err != nil {
				log.Error("Could not decode manifest.", "err", err)
				return
			}

			err = walker.Walk(func(entry *api.ManifestEntry) error {

				fileAddr, err := hex.DecodeString(entry.Hash)
				if err != nil {
					log.Error("Error decoding hash present in manifest", "err", err)
					return err
				}

				// send the file to file workers
				fileWorkers <- storage.Reference(fileAddr)

				return nil
			})

			if err != nil {
				log.Error("Error walking manifest", "err", err)
				return
			}

			// Finally, remove the manifest file too
			fileWorkers <- storage.Reference(addr)

			// Signal end of file stream
			close(fileWorkers)

		} else {
			// Its a raw file.. no manifest.. so process only this hash
			fileWorkers <- storage.Reference(addr)

			// Signal end of file stream
			close(fileWorkers)

		}
	}()

	for fileRef := range fileWorkers {

		// Send the file to chunk workers
		chunkWorkers <- fileRef

		actualFileSize := uint64(0)
		rcvdFileSize := uint64(0)
		doneChunkWorker := make(chan struct{})
		var cwg sync.WaitGroup // Wait group to wait for chunk processing to complete

	QuitChunkFor:
		for {
			select {
			case <-doneChunkWorker:
				break QuitChunkFor

			case ref := <-chunkWorkers:
				cwg.Add(1)

				go func() {

					defer cwg.Done()

					chunkData, err := getter.Get(context.TODO(), ref)
					if err != nil {
						log.Error("Error getting chunk data from localstore.", "Address", hex.EncodeToString(ref))
						close(doneChunkWorker)
						return
					}

					datalen := len(chunkData)
					if datalen < 9 { // Atleast 1 data byte. first 8 bytes are address
						log.Error("Invalid chunk data from localstore.", "Address", hex.EncodeToString(ref))
						close(doneChunkWorker)
						return
					}

					subTreeSize := chunkData.Size()
					if actualFileSize < subTreeSize {
						actualFileSize = subTreeSize
					}

					if subTreeSize > chunk.DefaultSize {
						// this is a tree chunk
						// load the tree's branches
						branches := (datalen - 8) / hashSize
						for i := 0; i < branches; i++ {
							brAddr := make([]byte, hashSize)
							start := (i * hashSize) + 8
							end := ((i + 1) * hashSize) + 8
							copy(brAddr[:], chunkData[start:end])
							chunkWorkers <- storage.Reference(brAddr)
						}

					} else {
						// this is a data chunk
						rcvdFileSize = rcvdFileSize + chunk.DefaultSize
						if rcvdFileSize >= actualFileSize {
							close(doneChunkWorker)
						}
					}

					// process the chunk (pin / unpin / display)
					err = executeFunc(ref)
					if err != nil {
						// TODO: if this happens, we should go back and revert the entire file's chunks
						log.Error("Could not unpin chunk.", "Address", hex.EncodeToString(ref))
					}
				}()
			}
		}

		// Wait for all the chunks to finish execution
		cwg.Wait()

	}
}

func (p *API) removeDecryptionKeyFromChunkHash(ref []byte) []byte {
	// remove the decryption key from the encrypted file hash
	isEncrypted := len(ref) > p.hashSize
	if isEncrypted {
		chunkAddr := make([]byte, p.hashSize)
		copy(chunkAddr, ref[0:p.hashSize])
		return chunkAddr
	}
	return ref
}

func (p *API) getPinCounterOfChunk(addr chunk.Address) (uint64, error) {
	pinnedChunk, err := p.db.Get(context.Background(), chunk.ModeGetPin, p.removeDecryptionKeyFromChunkHash(addr))
	if err != nil {
		return 0, err
	}
	// Pin counter is passed in the Data field... decode and use it
	return binary.BigEndian.Uint64(pinnedChunk.Data()[:8]), nil
}

func loadPinnedFilesInfo(pinnedFiles map[string]FileInfo, stateStore state.Store) error {
	pinnedFiles = make(map[string]FileInfo)
	err := stateStore.Get("pin-files", &pinnedFiles)
	if err != nil {
		if err == state.ErrNotFound {
			log.Info("No pinned files found")
			return nil
		}
		return err
	}
	log.Info("Pinned files loaded")
	return nil
}

func storePinnedFilesInfo(pinnedFiles map[string]FileInfo, stateStore state.Store) error {
	if err := stateStore.Put("pin-files", pinnedFiles); err != nil {
		return err
	}
	return nil
}
