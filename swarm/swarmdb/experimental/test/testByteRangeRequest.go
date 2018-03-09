//test for requesting custom byte slices of a file

//need to change in storage:
//memStore to MemStore in: localstore.go, netstore.go
//setCapacity to SetCapacity in: memstore.go, dbStore.go


package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

//change these for testing
const testDataSize = 0x1000000
const readDataLength = 0x0000111
const readOffset = 5

//can leave these alone...
const defaultHash = "SHA3" //SHA256 here does not work
const defaultDbCapacity = 5000000
const defaultRadius = 0 //some note about not yet used?
const defaultCacheCapacity = 5000

//makes random byte slice of length l
func testDataReaderAndSlice(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	if _, err := rand.Read(slice); err != nil {
		fmt.Printf("rand error - make byte slice checking error\n")
		return
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

//makes local temp db file
func initDbStore() *storage.DbStore {
	dir, err := ioutil.TempDir("", "bzz-storage-test")
	if err != nil {
		fmt.Printf(err.Error())
		return nil
	}
	m, err := storage.NewDbStore(dir, storage.MakeHashFunc(defaultHash), defaultDbCapacity, defaultRadius)
	if err != nil {
		fmt.Printf("can't create store: %v\n", err.Error())
		return nil
	}
	return m
}

//stores byte slice, reads back custom byte range, writes custom byte range to file, cleans up memstore
func TestDPArandom() {

	//storing
	dbStore := initDbStore()
	dbStore.SetCapacity(50000)
	memStore := storage.NewMemStore(dbStore, defaultCacheCapacity)
	localStore := &storage.LocalStore{
		memStore,
		dbStore,
	}
	chunker := storage.NewTreeChunker(storage.NewChunkerParams())
	dpa := &storage.DPA{
		Chunker:    chunker,
		ChunkStore: localStore,
	}
	dpa.Start()
	defer dpa.Stop()
	defer os.RemoveAll("/tmp/bzz")

	reader, slice := testDataReaderAndSlice(testDataSize)
	wg := &sync.WaitGroup{}
	fmt.Printf("testDataSize: %+v\n", testDataSize)
	key, err := dpa.Store(reader, testDataSize, wg, nil)
	if err != nil {
		fmt.Printf("Store error: %v\n", err)
	}
	wg.Wait()

	//reading
	resultReader := dpa.Retrieve(key)
	fmt.Printf("key: %v\nresultReader is : %+v\n", key, resultReader)

	if readDataLength+readOffset > testDataSize {
		fmt.Printf("params are not realistic. change and try again\n")
		return
	}
	fmt.Printf("length to read is: %+v from byte %v to byte %v \n", readDataLength, readOffset, readDataLength+readOffset)

	resultSlice := make([]byte, readDataLength)
	n, err := resultReader.ReadAt(resultSlice, readOffset)
	if err != nil && err != io.EOF {
		fmt.Printf("Retrieve error: %v\n", err)
	}
	if bytes.Equal(slice[readOffset:n+readOffset], resultSlice) {
		fmt.Printf("result piece is equivalent to starting byte slice\n")
	} else {
		fmt.Printf("Comparison error.\n")
		return
	}

	//writing byte slice read  to a file in /tmp/
	//fmt.Printf("going to write  n bytes: %+v\n", n)
	//ioutil.WriteFile("/tmp/slice.bzz.16M", slice, 0666)
	//ioutil.WriteFile("/tmp/result.bzz.16M", resultSlice, 0666)

	//cleaning up memstore
	memStore.SetCapacity(0)

	//checking to see if it's empty:
	dpa.ChunkStore = memStore
	resultReader = dpa.Retrieve(key)
	resultSlice_all := make([]byte, testDataSize)
	if _, err = resultReader.ReadAt(resultSlice_all, 0); err == nil {
		fmt.Printf("Was able to read %d bytes from an empty memStore.\n", len(slice))
	} else {
		fmt.Printf("no bytes read, memstore successfully deleted\n")
	}

	fmt.Printf("all done!\n\n")
	return
}

func main() {

	TestDPArandom()
	return
}
