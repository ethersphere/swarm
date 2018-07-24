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
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

var (
	loglevel   = flag.Int("loglevel", 3, "loglevel")
	testHasher = storage.MakeHashFunc(resourceHashAlgorithm)()
	startTime  = Timestamp{
		Time: uint64(4200),
	}
	resourceFrequency = uint64(42)
	cleanF            func()
	resourceName      = "føø.bar"
	hashfunc          = storage.MakeHashFunc(storage.DefaultHash)
)

func init() {
	flag.Parse()
	log.Root().SetHandler(log.CallerFileHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(os.Stderr, log.TerminalFormat(true)))))
}

// simulated timeProvider
type fakeTimeProvider struct {
	currentTime uint64
}

func (f *fakeTimeProvider) Tick() {
	f.currentTime++
}

func (f *fakeTimeProvider) Now() Timestamp {
	return Timestamp{
		Time: f.currentTime,
	}
}

// make updates and retrieve them based on periods and versions
func TestResourceHandler(t *testing.T) {

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime.Time,
	}

	// signer containing private key
	signer := newAliceSigner()

	rh, datadir, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create a new resource
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	viewID := ResourceViewID{
		resourceID: ResourceID{
			Topic:     NewTopic("Mess with mru code and see what ghost catches you", nil),
			StartTime: startTime,
			Frequency: resourceFrequency,
		},
		ownerAddr: signer.Address(),
	}

	request, err := NewCreateUpdateRequest(&viewID.resourceID)
	if err != nil {
		t.Fatal(err)
	}
	request.Sign(signer)
	if err != nil {
		t.Fatal(err)
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
	fwdClock(int(resourceFrequency/2), timeProvider)
	data := []byte(updates[0])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[0]], err = rh.Update(ctx, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// update on first period with version = 1 to make it fail since there is already one update with version=1
	request, err = rh.NewUpdateRequest(ctx, &request.viewID)
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
	resourcekey[updates[1]], err = rh.Update(ctx, &request.SignedResourceUpdate)
	if err == nil {
		t.Fatal("Expected update to fail since this version already exists")
	}

	// update on second period with version = 1, correct. period=2, version=1
	fwdClock(int(resourceFrequency/2), timeProvider)
	request, err = rh.NewUpdateRequest(ctx, &request.viewID)
	if err != nil {
		t.Fatal(err)
	}
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[1]], err = rh.Update(ctx, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	fwdClock(int(resourceFrequency), timeProvider)
	// Update on third period, with version = 1
	request, err = rh.NewUpdateRequest(ctx, &request.viewID)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(updates[2])
	request.SetData(data)
	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[2]], err = rh.Update(ctx, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	// update just after third period
	fwdClock(1, timeProvider)
	request, err = rh.NewUpdateRequest(ctx, &request.viewID)
	if err != nil {
		t.Fatal(err)
	}
	if request.period != 3 || request.version != 2 {
		t.Fatal("Suggested period should be 3 and version should be 2")
	}
	data = []byte(updates[3])
	request.SetData(data)

	if err := request.Sign(signer); err != nil {
		t.Fatal(err)
	}
	resourcekey[updates[3]], err = rh.Update(ctx, &request.SignedResourceUpdate)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	rh.Close()

	// check we can retrieve the updates after close
	// it will match on second iteration startTime + (resourceFrequency * 3)
	fwdClock(int(resourceFrequency*2)-1, timeProvider)

	rhparams := &HandlerParams{}

	rh2, err := NewTestHandler(datadir, rhparams)
	if err != nil {
		t.Fatal(err)
	}

	rsrc2, err := rh2.Lookup(ctx, LookupLatest(&request.viewID))
	if err != nil {
		t.Fatal(err)
	}

	// last update should be "clyde", version two, time= startTime + (resourcefrequency * 3)
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

	// specific period, latest version
	rsrc, err := rh2.Lookup(ctx, LookupLatestVersionInPeriod(&request.viewID, 3))
	if err != nil {
		t.Fatal(err)
	}
	// check data
	if !bytes.Equal(rsrc.data, []byte(updates[len(updates)-1])) {
		t.Fatalf("resource data (historical) was %v, expected %v", string(rsrc2.data), updates[len(updates)-1])
	}
	log.Debug("Historical lookup", "period", rsrc2.period, "version", rsrc2.version, "data", rsrc2.data)

	// specific period, specific version
	lookupParams := LookupVersion(&request.viewID, 3, 1)
	rsrc, err = rh2.Lookup(ctx, lookupParams)
	if err != nil {
		t.Fatal(err)
	}
	// check data
	if !bytes.Equal(rsrc.data, []byte(updates[2])) {
		t.Fatalf("resource data (historical) was %v, expected %v", string(rsrc2.data), updates[2])
	}
	log.Debug("Specific version lookup", "period", rsrc2.period, "version", rsrc2.version, "data", rsrc2.data)

	// we are now at third update
	// check backwards stepping to the first
	for i := 1; i >= 0; i-- {
		rsrc, err := rh2.LookupPrevious(ctx, lookupParams)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rsrc.data, []byte(updates[i])) {
			t.Fatalf("resource data (previous) was %v, expected %v", rsrc.data, updates[i])

		}
	}

	// beyond the first should yield an error
	rsrc, err = rh2.LookupPrevious(ctx, lookupParams)
	if err == nil {
		t.Fatalf("expected previous to fail, returned period %d version %d data %v", rsrc.period, rsrc.version, rsrc.data)
	}

}

func TestValidator(t *testing.T) {

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime.Time,
	}

	// signer containing private key. Alice will be the good girl
	signer := newAliceSigner()

	// set up  sim timeProvider
	rh, _, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	// create new resource
	viewID := ResourceViewID{
		resourceID: ResourceID{
			Topic:     NewTopic(resourceName, nil),
			StartTime: timeProvider.Now(),
			Frequency: resourceFrequency,
		},
		ownerAddr: signer.Address(),
	}
	mr, err := NewCreateUpdateRequest(&viewID.resourceID)
	if err != nil {
		t.Fatal(err)
	}

	// chunk with address
	data := []byte("foo")
	mr.SetData(data)
	if err := mr.Sign(signer); err != nil {
		t.Fatalf("sign fail: %v", err)
	}

	chunk, err := mr.SignedResourceUpdate.toChunk()
	if err != nil {
		t.Fatal(err)
	}
	if !rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Chunk validator fail on update chunk")
	}

	// mess with the address
	chunk.Addr[0] = 11
	chunk.Addr[15] = 99

	if rh.Validate(chunk.Addr, chunk.SData) {
		t.Fatal("Expected Validate to fail with false chunk address")
	}
}

// tests that the content address validator correctly checks the data
// tests that resource update chunks are passed through content address validator
// there is some redundancy in this test as it also tests content addressed chunks,
// which should be evaluated as invalid chunks by this validator
func TestValidatorInStore(t *testing.T) {

	// make fake timeProvider
	TimestampProvider = &fakeTimeProvider{
		currentTime: startTime.Time,
	}

	// signer containing private key
	signer := newAliceSigner()

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
	rhParams := &HandlerParams{}
	rh := NewHandler(rhParams)
	store.Validators = append(store.Validators, rh)

	// create content addressed chunks, one good, one faulty
	chunks := storage.GenerateRandomChunks(chunk.DefaultSize, 2)
	goodChunk := chunks[0]
	badChunk := chunks[1]
	badChunk.SData = goodChunk.SData

	viewID := ResourceViewID{
		resourceID: ResourceID{
			Topic:     NewTopic("xyzzy", nil),
			StartTime: startTime,
			Frequency: resourceFrequency,
		},
		ownerAddr: signer.Address(),
	}

	// create a resource update chunk with correct publickey
	updateLookup := UpdateLookup{
		period:  42,
		version: 1,
		viewID:  viewID,
	}

	updateAddr := updateLookup.UpdateAddr()
	data := []byte("bar")

	r := new(SignedResourceUpdate)
	r.updateAddr = updateAddr
	r.resourceUpdate.UpdateLookup = updateLookup
	r.data = data

	r.Sign(signer)

	uglyChunk, err := r.toChunk()
	if err != nil {
		t.Fatal(err)
	}

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
func fwdClock(count int, timeProvider *fakeTimeProvider) {
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

	TimestampProvider = timeProvider
	rhparams := &HandlerParams{}
	rh, err = NewTestHandler(datadir, rhparams)
	return rh, datadir, cleanF, err
}

func newAliceSigner() *GenericSigner {
	privKey, _ := crypto.HexToECDSA("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	return NewGenericSigner(privKey)
}

func newBobSigner() *GenericSigner {
	privKey, _ := crypto.HexToECDSA("accedeaccedeaccedeaccedeaccedeaccedeaccedeaccedeaccedeaccedecaca")
	return NewGenericSigner(privKey)
}

func newCharlieSigner() *GenericSigner {
	privKey, _ := crypto.HexToECDSA("facadefacadefacadefacadefacadefacadefacadefacadefacadefacadefaca")
	return NewGenericSigner(privKey)
}

func getUpdateDirect(rh *Handler, addr storage.Address) ([]byte, error) {
	chunk, err := rh.chunkStore.Get(context.TODO(), addr)
	if err != nil {
		return nil, err
	}
	var r SignedResourceUpdate
	if err := r.fromChunk(addr, chunk.SData); err != nil {
		return nil, err
	}
	return r.data, nil
}
