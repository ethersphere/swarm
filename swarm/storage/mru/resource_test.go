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
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/ens"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/multihash"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var (
	loglevel          = flag.Int("loglevel", 3, "loglevel")
	testHasher        = storage.MakeHashFunc(resourceHashAlgorithm)()
	startTime         = uint64(4200)
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

// simulated timeProvider
type fakeTimeProvider struct {
	currentTime uint64
}

func (f *fakeTimeProvider) Tick() {
	f.currentTime++
}

func (f *fakeTimeProvider) GetCurrentTime() uint64 {
	return f.currentTime
}

// check that signature address matches update signer address
func TestReverse(t *testing.T) {

	period := uint32(4)
	version := uint32(2)

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime,
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up rpc and create resourcehandler
	_, _, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	metadata := resourceMetadata{
		name:      safeName,
		startTime: startTime,
		frequency: resourceFrequency,
		ownerAddr: signer.Address(),
	}

	rootAddr, metaHash, _ := metadata.hash()

	// generate a hash for block 4200 version 1
	key := resourceHash(period, version, rootAddr)

	// generate some bogus data for the chunk and sign it
	data := make([]byte, 8)
	_, err = rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	testHasher.Reset()
	testHasher.Write(data)

	update := &SignedResourceUpdate{
		resourceUpdate: resourceUpdate{
			period:   period,
			version:  version,
			metaHash: metaHash,
			rootAddr: rootAddr,
			data:     data,
		},
	}
	if err = update.Sign(signer); err != nil {
		t.Fatal(err)
	}

	chunk := newUpdateChunk(update)

	// check that we can recover the owner account from the update chunk's signature
	checkUpdate, err := parseUpdate(chunk.SData)
	if err != nil {
		t.Fatal(err)
	}
	checkdigest := keyDataHash(chunk.Addr, metaHash, checkUpdate.data)
	recoveredaddress, err := getAddressFromDataSig(checkdigest, *checkUpdate.signature)
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
	if period != checkUpdate.period {
		t.Fatalf("Expected period '%d', was '%d'", period, checkUpdate.period)
	}
	if version != checkUpdate.version {
		t.Fatalf("Expected version '%d', was '%d'", version, checkUpdate.version)
	}
	if !bytes.Equal(data, checkUpdate.data) {
		t.Fatalf("Expectedn data '%x', was '%x'", data, checkUpdate.data)
	}
}

// make updates and retrieve them based on periods and versions
func TestResourceHandler(t *testing.T) {

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime,
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	rh, datadir, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create a new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request, err := NewUpdateRequest(safeName, resourceFrequency, timeProvider.currentTime, signer.Address(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	err = rh.New(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	chunk, err := rh.chunkStore.Get(context.TODO(), storage.Address(request.rootAddr))
	if err != nil {
		t.Fatal(err)
	} else if len(chunk.SData) < 16 {
		t.Fatalf("chunk data must be minimum 16 bytes, is %d", len(chunk.SData))
	}
	startblocknumber := binary.LittleEndian.Uint64(chunk.SData[8:16])
	chunkfrequency := binary.LittleEndian.Uint64(chunk.SData[16:24])
	if startblocknumber != timeProvider.currentTime {
		t.Fatalf("stored block number %d does not match provided block number %d", startblocknumber, timeProvider.currentTime)
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

	// update halfway to first period. period=1, version=1
	resourcekey := make(map[string]storage.Address)
	fwdBlocks(int(resourceFrequency/2), timeProvider)
	data := []byte(updates[0])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[0]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// update on first period with version = 1 to make it fail since there is already one update with version=1
	request, err = rh.NewUpdateRequest(ctx, request.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	if request.version != 2 || request.period != 1 {
		t.Fatal("Suggested period should be 1 and version should be 2")
	}

	request.version = 1 // force version 1 instead of 2 to make it fail
	data = []byte(updates[1])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[1]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err == nil {
		t.Fatal("Expected update to fail since this version already exists")
	}

	// update on second period with version = 1, correct. period=2, version=1
	fwdBlocks(int(resourceFrequency/2), timeProvider)
	request, err = rh.NewUpdateRequest(ctx, request.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[1]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// first attempt to update on third period, setting version to something different than 1 to make it fail
	fwdBlocks(int(resourceFrequency), timeProvider)
	request, err = rh.NewUpdateRequest(ctx, request.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	request.version = 79 //put something different than 1 to make it fail
	data = []byte(updates[2])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[2]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err == nil {
		t.Fatal("Expected update to fail since this is the first version of this period and we didn't set version=1")
	}

	// second attempt to update on third period, now with version =1 should work since it is the first one
	request, err = rh.NewUpdateRequest(ctx, request.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(updates[2])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[2]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// update just after second period
	fwdBlocks(1, timeProvider)
	data = []byte(updates[3])
	request.SetData(data)

	request.version = 2
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[3]], err = rh.Update(ctx, request.rootAddr, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	rh.Close()

	// check we can retrieve the updates after close
	// it will match on second iteration startblocknumber + (resourceFrequency * 3)
	fwdBlocks(int(resourceFrequency*2)-1, timeProvider)

	rhparams := &HandlerParams{
		TimestampProvider: timeProvider,
	}

	rh2, err := NewTestHandler(datadir, rhparams)
	if err != nil {
		t.Fatal(err)
	}

	rsrc2, err := rh2.Load(context.TODO(), request.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	lookupParams := &LookupParams{
		Root: request.rootAddr,
	}
	_, err = rh2.Lookup(ctx, lookupParams)
	if err != nil {
		t.Fatal(err)
	}

	// last update should be "clyde", version two, blockheight startblocknumber + (resourcefrequency * 3)
	if !bytes.Equal(rsrc2.data, []byte(updates[len(updates)-1])) {
		t.Fatalf("resource data was %v, expected %v", string(rsrc2.data), updates[len(updates)-1])
	}
	if rsrc2.version != 2 {
		t.Fatalf("resource version was %d, expected 2", rsrc2.version)
	}
	if rsrc2.period != 3 {
		t.Fatalf("resource period was %d, expected 3", rsrc2.period)
	}
	log.Debug("Latest lookup", "period", rsrc2.period, "version", rsrc2.version, "data", rsrc2.data)

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
	log.Debug("Historical lookup", "period", rsrc2.period, "version", rsrc2.version, "data", rsrc2.data)

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
	log.Debug("Specific version lookup", "period", rsrc2.period, "version", rsrc2.version, "data", rsrc2.data)

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
		t.Fatalf("expeected previous to fail, returned period %d version %d data %v", rsrc2.period, rsrc2.version, rsrc2.data)
	}

}

func TestMultihash(t *testing.T) {

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime,
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up rpc and create resourcehandler
	rh, datadir, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create a new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mr, err := NewUpdateRequest(safeName, resourceFrequency, timeProvider.currentTime, signer.Address(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	err = rh.New(ctx, mr)
	if err != nil {
		t.Fatal(err)
	}

	// we're naïvely assuming keccak256 for swarm hashes
	// if it ever changes this test should also change
	multihashbytes := ens.EnsNode("foo")
	multihashmulti := multihash.ToMultihash(multihashbytes.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	mr.SetData(multihashmulti)
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	multihashkey, err := rh.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	sha1bytes := make([]byte, multihash.MultihashLength)
	sha1multi := multihash.ToMultihash(sha1bytes)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = rh.NewUpdateRequest(ctx, mr.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr.SetData(sha1multi)
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	sha1key, err := rh.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// invalid multihashes
	mr, err = rh.NewUpdateRequest(ctx, mr.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr.SetData(multihashmulti[1:])
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rh.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
	if err == nil {
		t.Fatalf("Expected update to fail with first byte skipped")
	}
	mr, err = rh.NewUpdateRequest(ctx, mr.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr.SetData(multihashmulti[:len(multihashmulti)-2])
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}

	_, err = rh.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
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
		TimestampProvider: rh.timestampProvider,
	}
	// test with signed data
	rh2, err := NewTestHandler(datadir, rhparams)
	if err != nil {
		t.Fatal(err)
	}
	mr, err = NewUpdateRequest(safeName, resourceFrequency, timeProvider.currentTime, signer.Address(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	err = rh2.New(ctx, mr)
	if err != nil {
		t.Fatal(err)
	}

	mr.SetData(multihashmulti)
	mr.Sign(signer)

	if err != nil {
		t.Fatal(err)
	}
	multihashsignedkey, err := rh2.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	mr, err = rh2.NewUpdateRequest(ctx, mr.rootAddr)
	if err != nil {
		t.Fatal(err)
	}
	mr.SetData(sha1multi)
	mr.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}

	sha1signedkey, err := rh2.Update(ctx, mr.rootAddr, &mr.SignedResourceUpdate)
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

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime,
	}

	// signer containing private key
	signer, err := newTestSigner()
	if err != nil {
		t.Fatal(err)
	}

	// fake signer for false results
	falseSigner, err := newFalseSigner()
	if err != nil {
		t.Fatal(err)
	}

	// set up  sim timeProvider
	rh, _, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mr, err := NewUpdateRequest(safeName, resourceFrequency, timeProvider.currentTime, signer.Address(), nil, false)
	if err != nil {
		t.Fatal(err)
	}
	mr.Sign(signer)

	err = rh.New(ctx, mr)
	if err != nil {
		t.Fatalf("Create resource fail: %v", err)
	}

	// chunk with address
	data := []byte("foo")
	mr.SetData(data)
	if err := mr.Sign(signer); err != nil {
		t.Fatalf("sign fail: %v", err)
	}
	chunk := newUpdateChunk(&mr.SignedResourceUpdate)
	if !rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator fail on update chunk")
	}

	// chunk with address made from different publickey
	if err := mr.Sign(falseSigner); err != nil {
		t.Fatalf("sign fail: %v", err)
	}

	chunk = newUpdateChunk(&mr.SignedResourceUpdate)
	if rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator did not fail on update chunk with false address")
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	startTime := rh.getCurrentTime(ctx)

	chunk, _ = rh.newMetaChunk(&resourceMetadata{
		name:      safeName,
		startTime: startTime,
		frequency: resourceFrequency,
		ownerAddr: signer.Address(),
	})
	if !rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator fail on metadata chunk")
	}
}

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// there is some redundancy in this test as it also tests content addressed chunks,
// which should be evaluated as invalid chunks by this validator
func TestValidatorInStore(t *testing.T) {

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime,
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
		TimestampProvider: timeProvider,
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

	rootChunk, metaHash := rh.newMetaChunk(&resourceMetadata{
		startTime: startTime,
		name:      "xyzzy",
		frequency: resourceFrequency,
		ownerAddr: signer.Address(),
	})

	// create a resource update chunk with correct publickey
	key := resourceHash(42, 1, rootChunk.Addr)
	data := []byte("bar")
	digestToSign := keyDataHash(key, metaHash, data)
	digestSignature, err := signer.Sign(digestToSign)
	uglyChunk := newUpdateChunk(&SignedResourceUpdate{
		key:       key,
		signature: &digestSignature,
		resourceUpdate: resourceUpdate{
			period:   42,
			version:  1,
			data:     data,
			rootAddr: rootChunk.Addr,
			metaHash: metaHash,
		},
	})

	// put the chunks in the store and check their error status
	storage.PutChunks(store, goodChunk)
	if goodChunk.GetErrored() == nil {
		t.Fatal("expected error on good content address chunk with resource validator only, but got nil")
	}
	storage.PutChunks(store, badChunk)
	if badChunk.GetErrored() == nil {
		t.Fatal("expected error on bad content address chunk with resource validator only, but got nil")
	}
	storage.PutChunks(store, uglyChunk)
	if err := uglyChunk.GetErrored(); err != nil {
		t.Fatalf("expected no error on resource update chunk with resource validator only, but got: %s", err)
	}
}

// fast-forward clock
func fwdBlocks(count int, timeProvider *fakeTimeProvider) {
	for i := 0; i < count; i++ {
		timeProvider.Tick()
	}
}

// create rpc and resourcehandler
func setupTest(timeProvider timestampProvider, signer Signer) (rh *TestHandler, datadir string, teardown func(), err error) {

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
		TimestampProvider: timeProvider,
	}
	rh, err = NewTestHandler(datadir, rhparams)
	return rh, datadir, cleanF, err
}

func newTestSigner() (*GenericSigner, error) {
	privKey, err := crypto.HexToECDSA("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		return nil, err
	}
	return NewGenericSigner(privKey), nil
}

func newFalseSigner() (*GenericSigner, error) {
	privKey, err := crypto.HexToECDSA("accedeaccedeaccedeaccedeaccedeaccedeaccedeaccedeaccedeaccedecaca")
	if err != nil {
		return nil, err
	}
	return NewGenericSigner(privKey), nil
}

func getUpdateDirect(rh *Handler, addr storage.Address) ([]byte, error) {
	chunk, err := rh.chunkStore.Get(context.TODO(), addr)
	if err != nil {
		return nil, err
	}
	mr, err := parseUpdate(chunk.SData)
	if err != nil {
		return nil, err
	}
	return mr.data, nil
}
