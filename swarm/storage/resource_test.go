package storage

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/idna"

	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	blockCount = uint64(4200)
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

func TestResourceHandler(t *testing.T) {
	datadir, err := ioutil.TempDir("", "rh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(datadir)
	log.Trace("starttest", "dir", datadir)

	// starting the whole stack just to get blocknumbers is too cumbersome
	// so we fake the rpc server to get blocknumbers for testing
	ipcpath := filepath.Join(datadir, "test.ipc")
	ipcl, err := rpc.CreateIPCListener(ipcpath)
	if err != nil {
		t.Fatal(err)
	}
	rpcserver := rpc.NewServer()
	rpcserver.RegisterName("eth", &FakeRPC{
		blockcount: &blockCount,
	})
	go func() {
		rpcserver.ServeListener(ipcl)
	}()
	defer rpcserver.Stop()

	// connect to fake rpc
	rpcclient, err := rpc.Dial(ipcpath)
	if err != nil {
		t.Fatal(err)
	}

	rh, err := NewResourceHandler(datadir, &testCloudStore{}, rpcclient)
	if err != nil {
		t.Fatal(err)
	}

	// create a new resource
	resourcename := "føø.bar"
	resourcevalidname, err := idna.ToASCII(resourcename)
	if err != nil {
		t.Fatal(err)
	}
	resourcefrequency := uint64(42)
	_, err = rh.NewResource(resourcename, resourcefrequency)
	if err != nil {
		t.Fatal(err)
	}

	// check that the new resource is stored correctly
	namehash := ens.EnsNode(resourcevalidname)
	chunk, err := rh.ChunkStore.(*resourceChunkStore).localStore.(*LocalStore).memStore.Get(Key(namehash[:]))
	if err != nil {
		t.Fatal(err)
	}
	startblocknumber := binary.LittleEndian.Uint64(chunk.SData[8:16])
	chunkfrequency := binary.LittleEndian.Uint64(chunk.SData[16:])
	if startblocknumber != blockCount {
		t.Fatalf("stored block number %d does not match provided block number %d", startblocknumber, blockCount)
	}
	if chunkfrequency != resourcefrequency {
		t.Fatalf("stored frequency %d does not match provided frequency %d", chunkfrequency, resourcefrequency)
	}

	// update halfway to first period
	key := make(map[string]Key)
	blockCount = startblocknumber + (resourcefrequency / 2)
	key["blinky"], err = rh.Update(resourcename, []byte("blinky"))
	if err != nil {
		t.Fatal(err)
	}

	// update on first period
	blockCount = startblocknumber + resourcefrequency
	key["pinky"], err = rh.Update(resourcename, []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}

	// update on second period
	blockCount = startblocknumber + (resourcefrequency * 2)
	key["inky"], err = rh.Update(resourcename, []byte("inky"))
	if err != nil {
		t.Fatal(err)
	}

	// update just after second period
	blockCount = startblocknumber + (resourcefrequency * 2) + 1
	key["clyde"], err = rh.Update(resourcename, []byte("clyde"))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	rh.Close()

	// check we can retrieve the updates after close
	// it will match on second iteration startblocknumber + (resourcefrequency * 3)
	blockCount = startblocknumber + (resourcefrequency * 4)

	rh2, err := NewResourceHandler(datadir, &testCloudStore{}, rpcclient)
	_, err = rh2.OpenResource(resourcename, true)
	if err != nil {
		t.Fatal(err)
	}

	// last update should be "clyde", version two, blockheight startblocknumber + (resourcefrequency * 3)
	if !bytes.Equal(rh2.resources[resourcename].data, []byte("clyde")) {
		t.Fatalf("resource data was %v, expected %v", rh2.resources[resourcename].data, []byte("clyde"))
	}
	if rh2.resources[resourcename].version != 2 {
		t.Fatalf("resource version was %d, expected 2", rh2.resources[resourcename].version)
	}
	if rh2.resources[resourcename].lastblock != startblocknumber+(resourcefrequency*3) {
		t.Fatalf("resource blockheight was %d, expected %d", rh2.resources[resourcename].lastblock, startblocknumber+(resourcefrequency*3))
	}

	rsrc, err := NewResource(resourcename, startblocknumber, resourcefrequency)
	if err != nil {
		t.Fatal(err)
	}
	err = rh2.SetResource(rsrc, true)
	if err != nil {
		t.Fatal(err)
	}
	resource, err := rh2.OpenResource(resourcename, false) // if key is specified, refresh is implicit
	if err != nil {
		t.Fatal(err)
	}

	// check data
	if !bytes.Equal(resource.data, []byte("clyde")) {
		t.Fatalf("resource data was %v, expected %v", rh2.resources[resourcename].data, []byte("clyde"))
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
