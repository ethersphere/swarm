package mru

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

func getTestSignedResourceUpdate() *Request {
	return &Request{
		ResourceUpdate: *getTestResourceUpdate(),
	}
}

func TestUpdateChunkSerializationErrorChecking(t *testing.T) {

	// Test that parseUpdate fails if the chunk is too small
	var r Request
	if err := r.fromChunk(storage.ZeroAddr, make([]byte, minimumUpdateDataLength-1)); err == nil {
		t.Fatalf("Expected parseUpdate to fail when chunkData contains less than %d bytes", minimumUpdateDataLength)
	}

	r = *getTestSignedResourceUpdate()

	_, err := r.toChunk()
	if err == nil {
		t.Fatal("Expected newUpdateChunk to fail when there is no data")
	}
	r.data = []byte("Al bien hacer jamás le falta premio") // put some arbitrary length data
	_, err = r.toChunk()
	if err == nil {
		t.Fatal("expected newUpdateChunk to fail when there is no signature", err)
	}

	charlie := newCharlieSigner()
	if err := r.Sign(charlie); err != nil {
		t.Fatalf("error signing:%s", err)
	}

	chunk, err := r.toChunk()
	if err != nil {
		t.Fatalf("error creating update chunk:%s", err)
	}

	compareByteSliceToExpectedHex(t, "chunk", chunk.SData, "0x10dd205b00000000100e000000000000776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c4f000000da070000416c206269656e206861636572206a616dc3a173206c652066616c7461207072656d696f5214c16d60870afb21c679a6020d764c7bb0e7bd8a66cfa7d0aa06e9fbf9c6d0666b013ca47f7a9c44d8d54a14193ba3514a0e1f56234a0618e03f0a836801e100")

	var recovered Request
	l := len(chunk.SData)
	fmt.Println(l)
	recovered.fromChunk(chunk.Addr, chunk.SData)
	if !reflect.DeepEqual(recovered, r) {
		t.Fatal("Expected recovered SignedResource update to equal the original one")
	}
}

// check that signature address matches update signer address
func TestReverse(t *testing.T) {

	period := uint32(4)
	version := uint32(2)

	// make fake timeProvider
	timeProvider := &fakeTimeProvider{
		currentTime: startTime.Time,
	}

	// signer containing private key
	signer := newAliceSigner()

	// set up rpc and create resourcehandler
	_, _, teardownTest, err := setupTest(timeProvider, signer)
	if err != nil {
		t.Fatal(err)
	}
	defer teardownTest()

	view := View{
		Resource: Resource{
			Topic:     NewTopic("Cervantes quotes", nil),
			StartTime: startTime,
			Frequency: resourceFrequency,
		},
		User: signer.Address(),
	}

	data := []byte("Donde una puerta se cierra, otra se abre")

	update := new(Request)
	update.View = view
	update.Period = period
	update.Version = version
	update.data = data

	// generate a hash for t=4200 version 1
	key := update.UpdateAddr()

	if err = update.Sign(signer); err != nil {
		t.Fatal(err)
	}

	chunk, err := update.toChunk()
	if err != nil {
		t.Fatal(err)
	}

	// check that we can recover the owner account from the update chunk's signature
	var checkUpdate Request
	if err := checkUpdate.fromChunk(chunk.Addr, chunk.SData); err != nil {
		t.Fatal(err)
	}
	checkdigest, err := checkUpdate.GetDigest()
	if err != nil {
		t.Fatal(err)
	}
	recoveredaddress, err := getUserAddr(checkdigest, *checkUpdate.Signature)
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
	if period != checkUpdate.Period {
		t.Fatalf("Expected period '%d', was '%d'", period, checkUpdate.Period)
	}
	if version != checkUpdate.Version {
		t.Fatalf("Expected version '%d', was '%d'", version, checkUpdate.Version)
	}
	if !bytes.Equal(data, checkUpdate.data) {
		t.Fatalf("Expectedn data '%x', was '%x'", data, checkUpdate.data)
	}
}
