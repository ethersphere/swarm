// Copyright 2018 The go-ethereum Authors
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

package mru

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var (
	loglevel          = flag.Int("loglevel", 3, "loglevel")
	testHasher        = storage.MakeHashFunc(storage.SHA3Hash)()
	startBlock        = uint64(4200)
	resourceFrequency = uint64(42)
	cleanF            func()
	domainName        = "føø.bar"
	safeName          string
	nameHash          common.Hash
	hashfunc          = storage.MakeHashFunc(storage.DefaultHash)
)

func init() {
	var err error
	flag.Parse()
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))
	safeName, err = ToSafeName(domainName)
	if err != nil {
		panic(err)
	}
	nameHash = ens.EnsNode(safeName)
}

// simulated backend does not have the blocknumber call
// so we use this wrapper to fake returning the block count
type fakeBackend struct {
	*backends.SimulatedBackend
	blocknumber int64
}

func (f *fakeBackend) Commit() {
	if f.SimulatedBackend != nil {
		f.SimulatedBackend.Commit()
	}
	f.blocknumber++
}

func (f *fakeBackend) HeaderByNumber(context context.Context, name string, bigblock *big.Int) (*types.Header, error) {
	f.blocknumber++
	biggie := big.NewInt(f.blocknumber)
	return &types.Header{
		Number: biggie,
	}, nil
}

// check that signature address matches update signer address
func TestReverse(t *testing.T) {

	period := uint32(4)
	version := uint32(2)

	// make fake backend, set up rpc and create resourcehandler
	backend := &fakeBackend{
		blocknumber: int64(startBlock),
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up rpc and create resourcehandler
	rh, _, teardownTest, err := setupTest(backend, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// generate a hash for block 4200 version 1
	key := rh.resourceHash(period, version, ens.EnsNode(safeName), signer.Address())

	// generate some bogus data for the chunk and sign it
	data := make([]byte, 8)
	_, err = rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	testHasher.Reset()
	testHasher.Write(data)
	digest := rh.keyDataHash(key, data)
	sig, err := rh.signer.Sign(digest)
	if err != nil {
		t.Fatal(err)
	}

	chunk := newUpdateChunk(key, &sig, period, version, safeName, data, len(data))

	// check that we can recover the owner account from the update chunk's signature
	checksig, checkperiod, checkversion, checkname, checkdata, _, err := rh.parseUpdate(chunk.SData)
	if err != nil {
		t.Fatal(err)
	}
	checkdigest := rh.keyDataHash(chunk.Addr, checkdata)
	recoveredaddress, err := getAddressFromDataSig(checkdigest, *checksig)
	if err != nil {
		t.Fatalf("Retrieve address from signature fail: %v", err)
	}
	originaladdress := crypto.PubkeyToAddress(signer.PrivKey.PublicKey)

	// check that the metadata retrieved from the chunk matches what we gave it
	if recoveredaddress != originaladdress {
		t.Fatalf("addresses dont match: %x != %x", originaladdress, recoveredaddress)
	}

	if !bytes.Equal(key[:], chunk.Addr[:]) {
		t.Fatalf("Expected chunk key '%x', was '%x'", key, chunk.Addr)
	}
	if period != checkperiod {
		t.Fatalf("Expected period '%d', was '%d'", period, checkperiod)
	}
	if version != checkversion {
		t.Fatalf("Expected version '%d', was '%d'", version, checkversion)
	}
	if safeName != checkname {
		t.Fatalf("Expected name '%s', was '%s'", safeName, checkname)
	}
	if !bytes.Equal(data, checkdata) {
		t.Fatalf("Expectedn data '%x', was '%x'", data, checkdata)
	}
}

// make updates and retrieve them based on periods and versions
func TestResourceHandler(t *testing.T) {

	// make fake backend, set up rpc and create resourcehandler
	backend := &fakeBackend{
		blocknumber: int64(startBlock),
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	rh, datadir, teardownTest, err := setupTest(backend, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create a new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rootChunkKey, _, err := rh.New(ctx, safeName, resourceFrequency)
	if err != nil {
		t.Fatal(err)
	}

	chunk, err := rh.chunkStore.Get(storage.Address(rootChunkKey))
	if err != nil {
		t.Fatal(err)
	} else if len(chunk.SData) < 16 {
		t.Fatalf("chunk data must be minimum 16 bytes, is %d", len(chunk.SData))
	}
	startblocknumber := binary.LittleEndian.Uint64(chunk.SData[8:16])
	chunkfrequency := binary.LittleEndian.Uint64(chunk.SData[16:24])
	if startblocknumber != uint64(backend.blocknumber) {
		t.Fatalf("stored block number %d does not match provided block number %d", startblocknumber, backend.blocknumber)
	}
	if chunkfrequency != resourceFrequency {
		t.Fatalf("stored frequency %d does not match provided frequency %d", chunkfrequency, resourceFrequency)
	}

	// data for updates:
	updates := []string{
		"blinky",
		"pinky",
		"inky",
		"clyde",
	}

	// update halfway to first period
	resourcekey := make(map[string]storage.Address)
	fwdBlocks(int(resourceFrequency/2), backend)
	data := []byte(updates[0])
	resourcekey[updates[0]], err = rh.Update(ctx, rootChunkKey, data)
	if err != nil {
		t.Fatal(err)
	}

	// update on first period
	fwdBlocks(int(resourceFrequency/2), backend)
	data = []byte(updates[1])
	resourcekey[updates[1]], err = rh.Update(ctx, rootChunkKey, data)
	if err != nil {
		t.Fatal(err)
	}

	// update on second period
	fwdBlocks(int(resourceFrequency), backend)
	data = []byte(updates[2])
	resourcekey[updates[2]], err = rh.Update(ctx, rootChunkKey, data)
	if err != nil {
		t.Fatal(err)
	}

	// update just after second period
	fwdBlocks(1, backend)
	data = []byte(updates[3])
	resourcekey[updates[3]], err = rh.Update(ctx, rootChunkKey, data)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	rh.Close()

	// check we can retrieve the updates after close
	// it will match on second iteration startblocknumber + (resourceFrequency * 3)
	fwdBlocks(int(resourceFrequency*2)-1, backend)

	rhparams := &HandlerParams{
		Signer:       signer,
		HeaderGetter: rh.headerGetter,
	}

	rh2, err := NewTestHandler(datadir, rhparams)
	if err != nil {
		t.Fatal(err)
	}
	rsrc2, err := rh2.Load(rootChunkKey)
	if err != nil {
		t.Fatal(err)
	}
	lookupParams := &LookupParams{
		Root: rootChunkKey,
	}
	_, err = rh2.Lookup(ctx, lookupParams)
	if err != nil {
		t.Fatal(err)
	}

	// last update should be "clyde", version two, blockheight startblocknumber + (resourcefrequency * 3)
	if !bytes.Equal(rsrc2.data, []byte(updates[len(updates)-1])) {
		t.Fatalf("resource data was %v, expected %v", rsrc2.data, updates[len(updates)-1])
	}
	if rsrc2.version != 2 {
		t.Fatalf("resource version was %d, expected 2", rsrc2.version)
	}
	if rsrc2.lastPeriod != 3 {
		t.Fatalf("resource period was %d, expected 3", rsrc2.lastPeriod)
	}
	log.Debug("Latest lookup", "period", rsrc2.lastPeriod, "version", rsrc2.version, "data", rsrc2.data)

	// specific block, latest version
	lookupParams.Period = 3
	rsrc, err := rh2.Lookup(ctx, lookupParams)
	if err != nil {
		t.Fatal(err)
	}
	// check data
	if !bytes.Equal(rsrc.data, []byte(updates[len(updates)-1])) {
		t.Fatalf("resource data (historical) was %v, expected %v", rsrc2.data, updates[len(updates)-1])
	}
	log.Debug("Historical lookup", "period", rsrc2.lastPeriod, "version", rsrc2.version, "data", rsrc2.data)

	// specific block, specific version
	lookupParams.Version = 1
	rsrc, err = rh2.Lookup(ctx, lookupParams)
	if err != nil {
		t.Fatal(err)
	}
	// check data
	if !bytes.Equal(rsrc.data, []byte(updates[2])) {
		t.Fatalf("resource data (historical) was %v, expected %v", rsrc2.data, updates[2])
	}
	log.Debug("Specific version lookup", "period", rsrc2.lastPeriod, "version", rsrc2.version, "data", rsrc2.data)

	// we are now at third update
	// check backwards stepping to the first
	for i := 1; i >= 0; i-- {
		//rsrc, err := rh2.LookupPreviousByName(ctx, safeName, rh2.queryMaxPeriods)
		rsrc, err := rh2.LookupPrevious(ctx, lookupParams)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rsrc.data, []byte(updates[i])) {
			t.Fatalf("resource data (previous) was %v, expected %v", rsrc2.data, updates[i])

		}
	}

	// beyond the first should yield an error
	rsrc, err = rh2.LookupPrevious(ctx, lookupParams)
	if err == nil {
		t.Fatalf("expeected previous to fail, returned period %d version %d data %v", rsrc2.lastPeriod, rsrc2.version, rsrc2.data)
	}

}

func TestMultihash(t *testing.T) {

	// make fake backend, set up rpc and create resourcehandler
	backend := &fakeBackend{
		blocknumber: int64(startBlock),
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up rpc and create resourcehandler
	rh, datadir, teardownTest, err := setupTest(backend, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create a new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rootChunkAddr, _, err := rh.New(ctx, safeName, resourceFrequency)
	if err != nil {
		t.Fatal(err)
	}

	// we're naïvely assuming keccak256 for swarm hashes
	// if it ever changes this test should also change
	multihashbytes := ens.EnsNode("foo")
	multihashmulti := multihash.ToMultihash(multihashbytes.Bytes())
	multihashkey, err := rh.UpdateMultihash(ctx, rootChunkAddr, multihashmulti)
	if err != nil {
		t.Fatal(err)
	}

	sha1bytes := make([]byte, multihash.MultihashLength)
	sha1multi := multihash.ToMultihash(sha1bytes)
	sha1key, err := rh.UpdateMultihash(ctx, rootChunkAddr, sha1multi)
	if err != nil {
		t.Fatal(err)
	}

	// invalid multihashes
	_, err = rh.UpdateMultihash(ctx, rootChunkAddr, multihashmulti[1:])
	if err == nil {
		t.Fatalf("Expected update to fail with first byte skipped")
	}
	_, err = rh.UpdateMultihash(ctx, rootChunkAddr, multihashmulti[:len(multihashmulti)-2])
	if err == nil {
		t.Fatalf("Expected update to fail with last byte skipped")
	}

	data, err := getUpdateDirect(rh.Handler, multihashkey)
	if err != nil {
		t.Fatal(err)
	}
	multihashdecode, err := multihash.FromMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(multihashdecode, multihashbytes.Bytes()) {
		t.Fatalf("Decoded hash '%x' does not match original hash '%x'", multihashdecode, multihashbytes.Bytes())
	}
	data, err = getUpdateDirect(rh.Handler, sha1key)
	if err != nil {
		t.Fatal(err)
	}
	shadecode, err := multihash.FromMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(shadecode, sha1bytes) {
		t.Fatalf("Decoded hash '%x' does not match original hash '%x'", shadecode, sha1bytes)
	}
	rh.Close()

	rhparams := &HandlerParams{
		Signer:       signer,
		HeaderGetter: rh.headerGetter,
	}
	// test with signed data
	rh2, err := NewTestHandler(datadir, rhparams)
	if err != nil {
		t.Fatal(err)
	}
	rootChunkAddr, _, err = rh2.New(ctx, safeName, resourceFrequency)
	if err != nil {
		t.Fatal(err)
	}
	multihashsignedkey, err := rh2.UpdateMultihash(ctx, rootChunkAddr, multihashmulti)
	if err != nil {
		t.Fatal(err)
	}
	sha1signedkey, err := rh2.UpdateMultihash(ctx, rootChunkAddr, sha1multi)
	if err != nil {
		t.Fatal(err)
	}

	data, err = getUpdateDirect(rh2.Handler, multihashsignedkey)
	if err != nil {
		t.Fatal(err)
	}
	multihashdecode, err = multihash.FromMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(multihashdecode, multihashbytes.Bytes()) {
		t.Fatalf("Decoded hash '%x' does not match original hash '%x'", multihashdecode, multihashbytes.Bytes())
	}
	data, err = getUpdateDirect(rh2.Handler, sha1signedkey)
	if err != nil {
		t.Fatal(err)
	}
	shadecode, err = multihash.FromMultihash(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(shadecode, sha1bytes) {
		t.Fatalf("Decoded hash '%x' does not match original hash '%x'", shadecode, sha1bytes)
	}
}

// \TODO verify testing of signature validation and enforcement
func TestValidator(t *testing.T) {

	// make fake backend, set up rpc and create resourcehandler
	backend := &fakeBackend{
		blocknumber: int64(startBlock),
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// fake signer for false results
	falseSigner, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up rpc and create resourcehandler with ENS sim backend
	rh, _, teardownTest, err := setupTest(backend, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	key, rsrc, err := rh.New(ctx, safeName, resourceFrequency)
	if err != nil {
		t.Fatalf("Create resource fail: %v", err)
	}

	// chunk with address
	data := []byte("foo")
	key = rh.resourceHash(1, 1, rsrc.nameHash, signer.Address())
	digest := rh.keyDataHash(key, data)
	sig, err := rh.signer.Sign(digest)
	if err != nil {
		t.Fatalf("sign fail: %v", err)
	}
	chunk := newUpdateChunk(key, &sig, 1, 1, safeName, data, len(data))
	if !rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator fail on update chunk")
	}

	// chunk with address made from different publickey
	key = rh.resourceHash(1, 1, rsrc.nameHash, falseSigner.Address())
	digest = rh.keyDataHash(key, data)
	sig, err = rh.signer.Sign(digest)
	if err != nil {
		t.Fatalf("sign fail: %v", err)
	}
	chunk = newUpdateChunk(key, &sig, 1, 1, safeName, data, len(data))
	if rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator did not fail on update chunk with false address")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	startBlock, err := rh.getBlock(ctx, safeName)
	if err != nil {
		t.Fatal(err)
	}
	chunk = rh.newMetaChunk(safeName, startBlock, resourceFrequency, signer.Address())
	if rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator fail on metadata chunk")
	}
}

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// there is some redundancy in this test as it also tests content addressed chunks,
// which should be evaluated as invalid chunks by this validator
func TestValidatorInStore(t *testing.T) {

	// make fake backend, set up rpc and create resourcehandler
	backend := &fakeBackend{
		blocknumber: int64(startBlock),
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up localstore
	datadir, err := ioutil.TempDir("", "storage-testresourcevalidator")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(datadir)

	params := storage.NewDefaultLocalStoreParams()
	params.Init(datadir)
	store, err := storage.NewLocalStore(params, nil)
	if err != nil {
		t.Fatal(err)
	}

	// set up resource handler and add is as a validator to the localstore
	rhParams := &HandlerParams{
		HeaderGetter: backend,
		Signer:       signer,
	}
	rh, err := NewHandler(rhParams)
	if err != nil {
		t.Fatal(err)
	}
	store.Validators = append(store.Validators, rh)

	// create content addressed chunks, one good, one faulty
	chunks := storage.GenerateRandomChunks(storage.DefaultChunkSize, 2)
	goodChunk := chunks[0]
	badChunk := chunks[1]
	badChunk.SData = goodChunk.SData

	// create a resource update chunk with correct publickey
	key := rh.resourceHash(42, 1, ens.EnsNode("xyzzy.eth"), signer.Address())
	data := []byte("bar")
	digestToSign := rh.keyDataHash(key, data)
	digestSignature, err := signer.Sign(digestToSign)
	uglyChunk := newUpdateChunk(key, &digestSignature, 42, 1, "xyzzy.eth", data, len(data))

	// put the chunks in the store and check their error status
	storage.PutChunks(store, goodChunk, badChunk, uglyChunk)
	if goodChunk.GetErrored() == nil {
		t.Fatal("expected error on good content address chunk with resource validator only, but got nil")
	}
	if badChunk.GetErrored() == nil {
		t.Fatal("expected error on bad content address chunk with resource validator only, but got nil")
	}
	if err := uglyChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on resource update chunk with resource validator only, but got: %s", err)
	}
}

// fast-forward blockheight
func fwdBlocks(count int, backend *fakeBackend) {
	for i := 0; i < count; i++ {
		backend.Commit()
	}
}

// create rpc and resourcehandler
func setupTest(backend headerGetter, signer Signer) (rh *TestHandler, datadir string, teardown func(), err error) {

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

	// temp datadir
	datadir, err = ioutil.TempDir("", "rh")
	if err != nil {
		return nil, "", nil, err
	}
	fsClean = func() {
		os.RemoveAll(datadir)
	}

	rhparams := &HandlerParams{
		Signer:       signer,
		HeaderGetter: backend,
	}
	rh, err = NewTestHandler(datadir, rhparams)
	return rh, datadir, cleanF, err
}

func newTestSigner() (*GenericSigner, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return NewGenericSigner(privKey), nil
}

func getUpdateDirect(rh *Handler, addr storage.Address) ([]byte, error) {
	chunk, err := rh.chunkStore.Get(addr)
	if err != nil {
		return nil, err
	}
	_, _, _, _, data, _, err := rh.parseUpdate(chunk.SData)
	if err != nil {
		return nil, err
	}
	return data, nil
}
