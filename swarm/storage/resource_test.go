package storage

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	blockCount = uint64(4200)
	cleanF     func()
)

func init() {
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))
}

type FakeRPC struct {
	blockcount *uint64
}

func (r *FakeRPC) BlockNumber() (string, error) {
	return strconv.FormatUint(*r.blockcount, 10), nil
}

func TestResourceValidContent(t *testing.T) {

	rh, privkey, _, err := setupTest()
	if err != nil {
		teardownTest(t, err.Error())
	}

	// create a new resource
	validname, err := idna.ToASCII("føø.bar")
	if err != nil {
		teardownTest(t, err.Error())
	}

	// generate a hash for block 4200 version 1
	key := rh.resourceHash(ens.EnsNode(validname), 4200, 1)
	chunk := NewChunk(key, nil)

	// generate some bogus data for the chunk and sign it
	data := make([]byte, 8)
	_, err = rand.Read(data)
	if err != nil {
		teardownTest(t, err.Error())
	}
	rh.hasher.Reset()
	rh.hasher.Write(data)
	datahash := rh.hasher.Sum(nil)
	sig, err := crypto.Sign(datahash, privkey)
	if err != nil {
		teardownTest(t, err.Error())
	}

	// put sig and data in the chunk
	chunk.SData = make([]byte, 8+SIGNATURE_LENGTH)
	copy(chunk.SData[:SIGNATURE_LENGTH], sig)
	copy(chunk.SData[SIGNATURE_LENGTH:], data)

	// check that we can recover the owner account from the update chunk's signature
	// TODO: change this to verifyContent on ENS integration
	recoveredaddress, err := rh.getContentAccount(chunk.SData)
	if err != nil {
		teardownTest(t, err.Error())
	}
	originaladdress := crypto.PubkeyToAddress(privkey.PublicKey)

	if recoveredaddress != originaladdress {
		teardownTest(t, fmt.Sprintf("addresses dont match: %x != %x", originaladdress, recoveredaddress))
	}
	teardownTest(t, "")
}

func TestResourceHandler(t *testing.T) {

	rh, privkey, datadir, err := setupTest()
	if err != nil {
		teardownTest(t, err.Error())
	}

	// create a new resource
	resourcename := "føø.bar"
	resourcevalidname, err := idna.ToASCII(resourcename)
	if err != nil {
		teardownTest(t, err.Error())
	}
	resourcefrequency := uint64(42)
	_, err = rh.NewResource(resourcename, resourcefrequency)
	if err != nil {
		teardownTest(t, err.Error())
	}

	// check that the new resource is stored correctly
	namehash := ens.EnsNode(resourcevalidname)
	chunk, err := rh.ChunkStore.(*resourceChunkStore).localStore.(*LocalStore).memStore.Get(Key(namehash[:]))
	if err != nil {
		teardownTest(t, err.Error())
	}
	startblocknumber := binary.LittleEndian.Uint64(chunk.SData[8:16])
	chunkfrequency := binary.LittleEndian.Uint64(chunk.SData[16:])
	if startblocknumber != blockCount {
		teardownTest(t, fmt.Sprintf("stored block number %d does not match provided block number %d", startblocknumber, blockCount))
	}
	if chunkfrequency != resourcefrequency {
		teardownTest(t, fmt.Sprintf("stored frequency %d does not match provided frequency %d", chunkfrequency, resourcefrequency))
	}

	// update halfway to first period
	key := make(map[string]Key)
	blockCount = startblocknumber + (resourcefrequency / 2)
	key["blinky"], err = rh.Update(resourcename, []byte("blinky"))
	if err != nil {
		teardownTest(t, err.Error())
	}

	// update on first period
	blockCount = startblocknumber + resourcefrequency
	key["pinky"], err = rh.Update(resourcename, []byte("pinky"))
	if err != nil {
		teardownTest(t, err.Error())
	}

	// update on second period
	blockCount = startblocknumber + (resourcefrequency * 2)
	key["inky"], err = rh.Update(resourcename, []byte("inky"))
	if err != nil {
		teardownTest(t, err.Error())
	}

	// update just after second period
	blockCount = startblocknumber + (resourcefrequency * 2) + 1
	key["clyde"], err = rh.Update(resourcename, []byte("clyde"))
	if err != nil {
		teardownTest(t, err.Error())
	}
	time.Sleep(time.Second)
	rh.Close()

	// check we can retrieve the updates after close
	// it will match on second iteration startblocknumber + (resourcefrequency * 3)
	blockCount = startblocknumber + (resourcefrequency * 4)

	rh2, err := NewResourceHandler(privkey, datadir, &testCloudStore{}, rh.ethapi)
	_, err = rh2.OpenResource(resourcename, true)
	if err != nil {
		teardownTest(t, err.Error())
	}

	// last update should be "clyde", version two, blockheight startblocknumber + (resourcefrequency * 3)
	if !bytes.Equal(rh2.resources[resourcename].data, []byte("clyde")) {
		teardownTest(t, fmt.Sprintf("resource data was %v, expected %v", rh2.resources[resourcename].data, []byte("clyde")))
	}
	if rh2.resources[resourcename].version != 2 {
		teardownTest(t, fmt.Sprintf("resource version was %d, expected 2", rh2.resources[resourcename].version))
	}
	if rh2.resources[resourcename].lastBlock != startblocknumber+(resourcefrequency*3) {
		teardownTest(t, fmt.Sprintf("resource blockheight was %d, expected %d", rh2.resources[resourcename].lastBlock, startblocknumber+(resourcefrequency*3)))
	}

	rsrc, err := NewResource(resourcename, startblocknumber, resourcefrequency)
	if err != nil {
		teardownTest(t, err.Error())
	}
	err = rh2.SetResource(rsrc, true)
	if err != nil {
		teardownTest(t, err.Error())
	}
	resource, err := rh2.OpenResource(resourcename, false) // if key is specified, refresh is implicit
	if err != nil {
		teardownTest(t, err.Error())
	}

	// check data
	if !bytes.Equal(resource.data, []byte("clyde")) {
		teardownTest(t, fmt.Sprintf("resource data was %v, expected %v", rh2.resources[resourcename].data, []byte("clyde")))
	}
	teardownTest(t, "")

}

func setupTest() (rh *ResourceHandler, privkey *ecdsa.PrivateKey, datadir string, err error) {

	var fsClean func()
	var rpcClean func()
	cleanF = func() {
		if fsClean != nil {
			fsClean()
		}
		if rpcClean != nil {
			rpcClean()
		}
	}

	// privkey for signing updates
	privkey, err = crypto.GenerateKey()
	if err != nil {
		return
	}

	// temp datadir
	datadir, err = ioutil.TempDir("", "rh")
	if err != nil {
		return
	}
	fsClean = func() {
		os.RemoveAll(datadir)
	}

	// starting the whole stack just to get blocknumbers is too cumbersome
	// so we fake the rpc server to get blocknumbers for testing
	ipcpath := filepath.Join(datadir, "test.ipc")
	ipcl, err := rpc.CreateIPCListener(ipcpath)
	if err != nil {
		return
	}
	rpcserver := rpc.NewServer()
	rpcserver.RegisterName("eth", &FakeRPC{
		blockcount: &blockCount,
	})
	go func() {
		rpcserver.ServeListener(ipcl)
	}()
	rpcClean = func() {
		rpcserver.Stop()
	}

	// connect to fake rpc
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		return
	}

	rh, err = NewResourceHandler(privkey, datadir, &testCloudStore{}, rpcclient)
	return

}

func teardownTest(t *testing.T, errstr string) {
	cleanF()
	if errstr != "" {
		t.Fatal(errstr)
	}
}

type testCloudStore struct {
}

func (c *testCloudStore) Store(*Chunk) {
}

func (c *testCloudStore) Deliver(*Chunk) {
}

func (c *testCloudStore) Retrieve(*Chunk) {
}
