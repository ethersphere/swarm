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
package api

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/storage"
	"github.com/ethersphere/swarm/storage/localstore"
)

const (
	PinVersion     = "1.0"
	WorkerChanSize = 8
)

var (
	errDecodingRootHash     = errors.New("error decoding root hash")
	errFileNotUploadedToPin = errors.New("file not uploaded")
)

// PinAPI is the main object which implements all things pinning.
type PinAPI struct {
	db         *localstore.DB
	api        *API
	fileParams *storage.FileStoreParams
	tag        *chunk.Tags
	hashSize   int
}

func NewPinApi(lstore *localstore.DB, params *storage.FileStoreParams, tags *chunk.Tags) *PinAPI {

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	pinApi := &PinAPI{
		db:         lstore,
		fileParams: params,
		tag:        tags,
		hashSize:   hashFunc().Size(),
	}

	return pinApi
}

func (p *PinAPI) SetApi(api *API) {
	p.api = api
}

// PinFiles is used to pin a RAW file or a collection (which hash manifest's)
// to the local Swarm node. It takes the root hash as the argument and walks
// down the merkle tree and pin all the chunks that are encountered on the
// way. It pins both data chunk and tree chunks. The pre-requisite is that
// the file should be present in the local database. This function is called
// from two places 1) Just after the file is uploaded 2) anytime after
// uploading the file using the pin command. This function can pin both
// encrypted and non-encrypted files.
func (p *PinAPI) PinFiles(rootHash string, isRaw bool, credentials string) error {

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "rootHash", rootHash, "Reason", err.Error())
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

	// Check if the root hash is already pinned
	if !p.db.IsFilePinned(chunk.Address(addr)) {
		if isRaw {
			err = p.db.Set(context.TODO(), chunk.ModeSetRawFile, addr)
		} else {
			err = p.db.Set(context.TODO(), chunk.ModeSetFile, addr)
		}
		if err != nil {
			log.Error("Could not unpin root chunk.", "Address", hex.EncodeToString(addr), "Reason", err.Error())
			return err
		}
	}

	log.Debug("File pinned", "Address", rootHash)

	return nil
}

// UnPinFiles is used to unpin an already pinned file. It takes the root
// hash of the file and walks down the merkle tree unpinning all the chunks
// that are encountered on the way. The pre-requisite is that the file should
// have been already pinned using the PinFiles function. This function can
// be called only from an external command.
func (p *PinAPI) UnpinFiles(rootHash string, credentials string) error {

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "rootHash", rootHash, "Reason", err.Error())
		return err
	}

	isRawInDB, err := p.db.IsPinnedFileRaw(chunk.Address(addr))
	if err != nil {
		log.Error("Root hash is not pinned", "rootHash", rootHash, "Reason", err.Error())
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
	p.walkChunksFromRootHash(rootHash, isRawInDB, credentials, walkerFunction)

	// Check if the root chunk exists in pinIndex
	// If it is not.. then the pin counter became 0
	// so remove the root hash from pinFilesIndex
	isRootChunkPinned := p.db.IsChunkPinned(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
	if !isRootChunkPinned {
		err := p.db.Set(context.TODO(), chunk.ModeSetUnpinFile, addr)
		if err != nil {
			log.Error("Could not unpin root chunk", "Address", hex.EncodeToString(addr), "Reason", err.Error())
			return err
		}
	}

	log.Debug("File unpinned", "Address", rootHash)

	return nil
}

// ListPinFiles functions logs information of all the files that are pinned
// in the current local node. It displays the root hash of the pinned file
// or collection. It also display two vital information
//     1) Size of the pinned file or collection
//     2) the number of times that particular file or collection is pinned.
func (p *PinAPI) ListPinFiles() {
	pinnedFiles := p.db.GetPinFilesIndex()
	for k, v := range pinnedFiles {
		addr, err := hex.DecodeString(k)
		if err != nil {
			log.Error("Error decoding root hash.", "rootHash", k, "Reason", err.Error())
			return
		}

		// This iteration can drain CPU. Figure out a way to store size of the pinned file
		// in the pinFileIndex itself by changing the refactoring the Set method to get additional parameters.
		pinCounter, err := p.db.GetPinCounterOfChunk(chunk.Address(p.removeDecryptionKeyFromChunkHash(addr)))
		if err != nil {
			log.Error("Error getting pin counter of root hash.", "rootHash", k, "Reason", err.Error())
			return
		}

		isRaw := false
		if v > 0 {
			isRaw = true
		}
		noOfChunks := p.getNoOfChunks(k, isRaw, "")

		log.Info("Pinned file", "Address", k, "NoOfChunks", noOfChunks, "pinCounter", pinCounter)
	}
}

func (p *PinAPI) walkChunksFromRootHash(rootHash string, isRaw bool, credentials string, executeFunc func(storage.Reference) error) {

	fileWorkers := make(chan storage.Reference, WorkerChanSize)
	chunkWorkers := make(chan storage.Reference, WorkerChanSize)

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "Reason", err.Error())
		return
	}

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	hashSize := len(addr)
	isEncrypted := len(addr) > hashFunc().Size()
	getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, chunk.NewTag(0, "show-chunks-tag", 0))

	go func() {
		if !isRaw {

			// If it not a raw file... load the manifest and process the files inside one by one
			walker, err := p.api.NewManifestWalker(context.TODO(), storage.Address(addr),
				p.api.Decryptor(context.TODO(), credentials), nil)

			if err != nil {
				log.Error("Could not decode manifest.", "Reason", err.Error())
				return
			}

			err = walker.Walk(func(entry *ManifestEntry) error {

				fileAddr, err := hex.DecodeString(entry.Hash)
				if err != nil {
					log.Error("Error decoding hash present in manifest", "Reason", err.Error())
					return err
				}

				// send the file to file workers
				fileWorkers <- storage.Reference(fileAddr)

				return nil
			})

			if err != nil {
				log.Error("Error walking manifest", "Reaon", err.Error())
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
						if rcvdFileSize > actualFileSize {
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

func (p *PinAPI) removeDecryptionKeyFromChunkHash(ref []byte) []byte {

	// remove the decryption key from the encrypted file hash
	isEncrypted := len(ref) > p.hashSize
	if isEncrypted {
		chunkAddr := make([]byte, p.hashSize)
		copy(chunkAddr, ref[0:p.hashSize])
		return chunkAddr
	}
	return ref
}

func (p *PinAPI) getNoOfChunks(rootHash string, isRaw bool, credentials string) uint64 {

	noOfChunks := uint64(0)
	walkerFunction := func(ref storage.Reference) error {
		noOfChunks += 1
		return nil
	}
	p.walkChunksFromRootHash(rootHash, isRaw, credentials, walkerFunction)

	return noOfChunks
}

//
// Functions used in testing
//

// getPinnedFiles is used in testing to get root hashes of all the pinned content
// including info whether they are raw files or manifests
func (p *PinAPI) GetPinnedFiles() map[string]uint8 {
	return p.db.GetPinFilesIndex()
}

// getAllChunksFromDB is used in testing to generate the truth dataset about all the chunks
// that are present in the DB.
func (p *PinAPI) GetAllChunksFromDB() map[string]int {
	return p.db.GetAllChunksInDB()
}

// collectPinnedChunks is used to collect all the chunks that are pinned as part of the
// given root hash.
func (p *PinAPI) CollectPinnedChunks(rootHash string, credentials string) map[string]uint64 {

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash", "Reason", err.Error())
		return nil
	}

	isRawInDB, err := p.db.IsPinnedFileRaw(chunk.Address(addr))
	if err != nil {
		log.Error("Root hash is not pinned", "Reason", err.Error())
		return nil
	}

	pinnedChunks := make(map[string]uint64)
	var lock = sync.RWMutex{}
	walkerFunction := func(ref storage.Reference) error {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		pinCounter, err := p.db.GetPinCounterOfChunk(chunk.Address(chunkAddr))
		if err != nil {
			return err
		}
		lock.Lock()
		defer lock.Unlock()
		pinnedChunks[hex.EncodeToString(chunkAddr)] = pinCounter
		return nil
	}
	p.walkChunksFromRootHash(rootHash, isRawInDB, credentials, walkerFunction)

	return pinnedChunks
}
