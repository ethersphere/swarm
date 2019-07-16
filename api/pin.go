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
	hashSize   int
}

func NewPinApi(lstore *localstore.DB, params *storage.FileStoreParams, tags *chunk.Tags) *PinApi {

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	pinApi := &PinApi{
		db:         lstore,
		fileParams: params,
		tag:        tags,
		hashSize:   hashFunc().Size(),
	}
	once.Do(func() {
		PinApiInstance = pinApi
	})

	return pinApi
}

func (p *PinApi) SetApi(api *API) {
	p.api = api
}

func (p *PinApi) ListPinFiles() {
	pinnedFiles := p.db.GetPinFilesIndex()
	for k,v := range  pinnedFiles {

		addr, err := hex.DecodeString(k)
		if err != nil {
			log.Error("Error decoding root hash" + err.Error())
			return
		}
		pinCounter, err := p.db.GetPinCounterOfChunk(p.removeDecryptionKeyFromChunkHash(addr))

		log.Info("Pinned file", "Address", k, "Size", v, "pinCounter", pinCounter)
	}
}



func (p *PinApi) PinFiles(rootHash string, isRaw bool, credentials string) error{

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return err
	}

	// Walk the root hash and pin all the chunks
	walkerFunction := func(ref storage.Reference)(error) {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		chunkToPin := chunk.NewChunk(chunkAddr,nil)
		_,err := p.db.Put(context.TODO(), chunk.ModePin, chunkToPin)
		if err != nil {
			log.Error("Could not pin chunk. Address " + hex.EncodeToString(chunkAddr))
			return err
		} else {
			log.Trace("Pinning chunk", "Address", hex.EncodeToString(chunkAddr))
		}
		return nil
	}
	p.WalkChunksFromRootHash(rootHash, isRaw, credentials, walkerFunction)


	// Check if the root hash is already pinned
	isFilePinned := p.db.IsFilePinned(addr)
	if  !isFilePinned {

		// Hack: If data is not nil, then this is a root chunk
		// also send the isRaw information in the data field
		data := []byte{0}
		if isRaw {
			data = []byte{1}
		}
		chunkToPin := chunk.NewChunk(addr,data)
		_,err := p.db.Put(context.TODO(), chunk.ModePin, chunkToPin)
		if err != nil {
			// TODO: if this happens, we should go back and revert the entire file's chunks
			log.Error("Could not unpin root chunk. Address " + fmt.Sprintf("%0x", addr))
			return err
		}
	}
	return nil
}

func (p *PinApi) UnpinFiles(rootHash string, credentials string) {

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return
	}

	isRawInDB, err := p.db.IsPinnedFileRaw(addr)
	if err != nil {
		log.Error("Root hash is not pinned" + err.Error())
		return
	}

	// Walk the root hash and unpin all the chunks
	walkerFunction := func(ref storage.Reference)(error) {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		chunkToPin := chunk.NewChunk(chunkAddr,nil)
		_,err := p.db.Put(context.TODO(), chunk.ModeUnpin, chunkToPin)
		if err != nil {
			log.Error("Could not unpin chunk. Address " + hex.EncodeToString(chunkAddr))
			return err
		} else {
			log.Trace("Unpinning chunk", "Address", fmt.Sprintf("%0x", chunkAddr))
		}
		return nil
	}
	p.WalkChunksFromRootHash(rootHash, isRawInDB, credentials, walkerFunction)


	// Check if the root chunk exists in pinIndex
	// If it is not.. then the pin counter became 0
	// so remove the root hash from pinFilesIndex
	isRootChunkPinned := p.db.IsChunkPinned(p.removeDecryptionKeyFromChunkHash(addr))
	if  !isRootChunkPinned {
		// Hack: If data is not nil, then this is a root chunk
		data := []byte{0}
		chunkToUnpin := chunk.NewChunk(addr,data)
		_,err := p.db.Put(context.TODO(), chunk.ModeUnpin, chunkToUnpin)
		if err != nil {
			// TODO: if this happens, we should go back and revert the entire file's chunks
			log.Error("Could not unpin root chunk. Address " + fmt.Sprintf("%0x", addr))
		}
	}
}

func (p *PinApi) LogPinnedChunks(rootHash string, credentials string) {

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return
	}

	isRawInDB, err := p.db.IsPinnedFileRaw(addr)
	if err != nil {
		log.Error("Root hash is not pinned" + err.Error())
		return
	}

	walkerFunction := func(ref storage.Reference) (error) {
		log.Info("Chunk", "Address", fmt.Sprintf("%0x", ref))
		return nil
	}
	p.WalkChunksFromRootHash(rootHash, isRawInDB, credentials, walkerFunction)
}

func (p *PinApi) WalkChunksFromRootHash(rootHash string, isRaw bool, credentials string, executeFunc func(storage.Reference) error) {

	fileWorkers := make(chan storage.Reference, WorkerChanSize)
	chunkWorkers := make(chan storage.Reference, WorkerChanSize)

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return
	}

	hashFunc := storage.MakeHashFunc(storage.DefaultHash)
	hashSize := len(addr)
	isEncrypted := len(addr) > hashFunc().Size()
	tag := chunk.NewTag(0, "show-chunks-tag", 0)
	getter := storage.NewHasherStore(p.db, hashFunc, isEncrypted, tag)


	go func() {
		if !isRaw {

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
	}()



	doneFileWorker := make(chan struct{})
QuitFileFor:
	for {
		select {
		case <-doneFileWorker:
			break QuitFileFor

		case fileRef := <-fileWorkers:

			if fileRef == nil {
				close(doneFileWorker)
				break
			}

			// Send the file to chunk workers
			chunkWorkers <- fileRef

			actualFileSize := uint64(0)
			rcvdFileSize := uint64(0)
			doneChunkWorker := make(chan struct{})
			var cwg sync.WaitGroup  // Wait group to wait for chunk processing to complete

		QuitChunkFor:
			for {
				select {
				case <-doneChunkWorker:
					break QuitChunkFor

				case ref := <-chunkWorkers:
					cwg.Add(1)

					go func() {

						//fmt.Println("Inside chunk pinner")
						chunkData, err := getter.Get(context.TODO(), ref)
						if err != nil {
							log.Error("Error getting chunk data from localstore.")
							close(doneChunkWorker)
							return
						}

						//fmt.Println("read chunkData")

						datalen := len(chunkData)
						if datalen < 9 { // Atleast 1 data byte. first 8 bytes are address
							log.Error("Invalid chunk data from localstore.")
							close(doneChunkWorker)
							return
						}

						subTreeSize := chunkData.Size()
						if actualFileSize < subTreeSize {
							actualFileSize = subTreeSize
							log.Info("File size ", "Size", actualFileSize)
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
							log.Error("Could not unpin chunk. Address " + fmt.Sprintf("%0x", ref))
						}
						cwg.Done()
					}()
				}
			}

			// Wait for all the chunks to finish execution
			cwg.Wait()
		}
	}
}


func (p *PinApi) ShowDatabase() string {
	p.db.ShowDatabaseInformation()
	return "Check the swarm log file for the output"
}

func (p *PinApi) removeDecryptionKeyFromChunkHash(ref []byte) []byte{

	// remove the decryption key from the encrypted file hash
	isEncrypted := len(ref) > p.hashSize
	if isEncrypted {
		chunkAddr := make([]byte,p.hashSize)
		copy(chunkAddr, ref[0:p.hashSize])
		return chunkAddr
	}
	return ref
}


// Used in testing
func (p *PinApi) GetPinnedFiles() map[string]uint64 {
	return p.db.GetPinFilesIndex()
}

func (p *PinApi) GetPinnedChunks() map[string]uint64 {
	return p.db.GetPinnedChunks()
}

func (p *PinApi) GetAllChunksFromDB() map[string]int {
	return p.db.GetAllChunksInDB()
}

func (p *PinApi) CollectPinnedChunks(rootHash string, credentials string) map[string]uint64{

	var lock = sync.RWMutex{}

	addr, err := hex.DecodeString(rootHash)
	if err != nil {
		log.Error("Error decoding root hash" + err.Error())
		return nil
	}

	isRawInDB, err := p.db.IsPinnedFileRaw(addr)
	if err != nil {
		log.Error("Root hash is not pinned" + err.Error())
		return nil
	}

	pinnedChunks := make(map[string]uint64)
	walkerFunction := func(ref storage.Reference) (error) {
		chunkAddr := p.removeDecryptionKeyFromChunkHash(ref)
		pinCounter, err := p.db.GetPinCounterOfChunk(chunkAddr)
		if err != nil {
			return err
		}
		lock.Lock()
		defer lock.Unlock()
		pinnedChunks[hex.EncodeToString(chunkAddr)] = pinCounter
		return nil
	}
	p.WalkChunksFromRootHash(rootHash, isRawInDB, credentials, walkerFunction)

	return pinnedChunks
}

