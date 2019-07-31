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
	Version        = "1.0"
	WorkerChanSize = 8             // Max no of goroutines when walking the file tree
	HeaderName     = "x-swarm-pin" // Presence of this in header indicates pinning required
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
	db         *localstore.DB
	api        *api.API
	fileParams *storage.FileStoreParams
	tag        *chunk.Tags
	hashSize   int
	state      state.Store // the state store used to store info about pinned files
}

// NewAPI creates a API object that is required for pinning and unpinning
func NewAPI(lstore *localstore.DB, stateStore state.Store, params *storage.FileStoreParams, tags *chunk.Tags, api *api.API) *API {
	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	return &API{
		db:         lstore,
		api:        api,
		fileParams: params,
		tag:        tags,
		hashSize:   hashFunc().Size(),
		state:      stateStore,
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

	hasChunk, err := p.db.Has(context.Background(), chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
	if !hasChunk {
		log.Error("Could not pin hash. File not uploaded", "Hash", rootHash)
		return err
	}

	// Walk the root hash and pin all the chunks
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		err := p.db.Set(context.Background(), chunk.ModeSetPin, chunkAddr)
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
	fileInfo, err := p.getPinnedFile(rootHash)
	if err != nil {
		// Get the file size from the root chunk first 8 bytes
		hashFunc := storage.MakeHashFunc(storage.DefaultHash)
		isEncrypted := len(addr) > hashFunc().Size()
		getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, chunk.NewTag(0, "show-chunks-tag", 0))
		chunkData, err := getter.Get(context.Background(), addr)
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

		fileInfo = FileInfo{
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
		fileInfo.pinCounter = pinCounter
	}

	// Store the pinned files in state DB
	err = p.savePinnedFile(rootHash, fileInfo)
	if err != nil {
		log.Error("Error saving pinned file info to state store.", "rootHash", rootHash, "err", err)
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

	fileInfo, err := p.getPinnedFile(rootHash)
	if err != nil {
		log.Error("Root hash is not pinned", "rootHash", rootHash, "err", err)
		return err
	}

	// Walk the root hash and unpin all the chunks
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		err := p.db.Set(context.Background(), chunk.ModeSetUnpin, chunkAddr)
		if err != nil {
			log.Error("Could not unpin chunk", "Address", hex.EncodeToString(chunkAddr))
			return err
		} else {
			log.Trace("Unpinning chunk", "Address", hex.EncodeToString(chunkAddr))
		}
		return nil
	}
	p.walkChunksFromRootHash(rootHash, fileInfo.isRaw, credentials, walkerFunction)

	// Delete or Update the state DB
	pinCounter, err := p.getPinCounterOfChunk(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
	if err != nil {
		err := p.removePinnedFile(rootHash)
		if err != nil {
			log.Error("Error unpinning file.", "rootHash", rootHash, "err", err)
			return nil
		}
	} else {
		fileInfo.pinCounter = pinCounter
		err = p.savePinnedFile(rootHash, fileInfo)
		if err != nil {
			log.Error("Error updating file info to state store.", "rootHash", rootHash, "err", err)
			return nil
		}
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
func (p *API) ListPinFiles() (map[string]FileInfo, error) {
	pinnedFiles := make(map[string]FileInfo)
	iterFunc := func(key []byte, value []byte) {
		hash := string(key[4:])
		fileInfo := FileInfo{}
		err := fileInfo.UnmarshalBinary(value)
		if err != nil {
			log.Debug("Error unmarshaling fileinfo from state store", "Address", hash)
		}
		log.Info("Pinned file", "Address", hash, "IsRAW", fileInfo.isRaw,
			"fileSize", fileInfo.fileSize, "pinCounter", fileInfo.pinCounter)
		pinnedFiles[hash] = fileInfo
	}
	err := p.state.Iterate("pin_", iterFunc)
	if err != nil {
		log.Error("Error iterating pinned files", "err", err)
		return nil, err
	}
	return pinnedFiles, nil
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
			walker, err := p.api.NewManifestWalker(context.Background(), storage.Address(addr),
				p.api.Decryptor(context.Background(), credentials), nil)

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

					chunkData, err := getter.Get(context.Background(), ref)
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
	return pinnedChunk.PinCounter(), nil
}

func (p *API) savePinnedFile(rootHash string, fileInfo FileInfo) error {
	key := "pin_" + rootHash
	err := p.state.Put(key, &fileInfo)
	return err
}

func (p *API) removePinnedFile(rootHash string) error {
	key := "pin_" + rootHash
	err := p.state.Delete(key)
	return err
}

func (p *API) getPinnedFile(rootHash string) (FileInfo, error) {
	key := "pin_" + rootHash
	fileInfo := FileInfo{}
	err := p.state.Get(key, &fileInfo)
	return fileInfo, err
}

// MarshalBinary encodes the FileInfo object in to a binary form for storage
func (f *FileInfo) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 17)
	if f.isRaw {
		data[0] = 1
	} else {
		data[0] = 0
	}
	binary.BigEndian.PutUint64(data[1:], uint64(f.fileSize))
	binary.BigEndian.PutUint64(data[9:], uint64(f.pinCounter))
	return data, nil
}

// UnmarshalBinary decodes the binary form from the state store to the FileInfo object
func (f *FileInfo) UnmarshalBinary(data []byte) error {
	if data[0] == 1 {
		f.isRaw = true
	} else {
		f.isRaw = false
	}
	f.fileSize = binary.BigEndian.Uint64(data[1:])
	f.pinCounter = binary.BigEndian.Uint64(data[9:])
	return nil
}
