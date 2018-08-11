package mru

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"

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
	if err := r.fromChunk(storage.ZeroAddr, make([]byte, minimumUpdateDataLength-1+signatureLength)); err == nil {
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

	compareByteSliceToExpectedHex(t, "chunk", chunk.SData, "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c0000000000000019416c206269656e206861636572206a616dc3a173206c652066616c7461207072656d696f52c3face89e9440d46d914c77c57e2b1f69b46b9319f31977368f4bb1cffb4dc7ed6cf70a765fcfb3ca1925f1749ed626a8b8a50e7d6fb62194310c6c7a45e6400")

	var recovered Request
	recovered.fromChunk(chunk.Addr, chunk.SData)
	if !reflect.DeepEqual(recovered, r) {
		t.Fatal("Expected recovered SignedResource update to equal the original one")
	}
}

// check that signature address matches update signer address
func TestReverse(t *testing.T) {

	epoch := lookup.Epoch{
		Time:  7888,
		Level: 6,
	}

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
			Topic: NewTopic("Cervantes quotes", nil),
		},
		User: signer.Address(),
	}

	data := []byte("Donde una puerta se cierra, otra se abre")

	update := new(Request)
	update.View = view
	update.Epoch = epoch
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
	if epoch != checkUpdate.Epoch {
		t.Fatalf("Expected epoch to be '%s', was '%s'", epoch.String(), checkUpdate.Epoch.String())
	}
	if !bytes.Equal(data, checkUpdate.data) {
		t.Fatalf("Expectedn data '%x', was '%x'", data, checkUpdate.data)
	}
}
