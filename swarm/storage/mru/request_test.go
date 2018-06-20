package mru

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func areEqualJSON(s1, s2 string) (bool, error) {
	//credit for the trick: turtlemonvh https://gist.github.com/turtlemonvh/e4f7404e28387fadb8ad275a99596f67
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

// TestEncodingDecodingUpdateRequests ensures that requests are serialized properly
// while also checking cryptographically that only the owner of a resource can update it.
func TestEncodingDecodingUpdateRequests(t *testing.T) {

	signer := newCharlieSigner()  //Charlie, our good guy
	falseSigner := newBobSigner() //Bob will play the bad guy again

	// Create a resource to our good guy Charlie's name, however, Charlie's
	// keyboard was upside down and generated invalid characters for his resource name
	// Only plain ASCII characters are allowed.
	createRequest, err := NewCreateRequest("sɹǝʇɔɐɹɐɥɔ ǝƃuɐɹʇs ɥʇᴉʍ ʇnq 'ǝɔɹnosǝɹ ǝɯosǝʍɐ ʎW",
		300, 1528900000, signer.Address(), nil, false)
	if err == nil {
		t.Fatal("Expected create request to fail since the name contains bad characters")
	}

	// Now Charlie tries again, this time with a proper resource name
	createRequest, err = NewCreateRequest("a good resource name",
		300, 1528900000, signer.Address(), nil, false)
	if err != nil {
		t.Fatalf("Error creating resource name: %s", err)
	}

	// We now encode the create message to simulate we send it over the wire
	messageRawData, err := EncodeUpdateRequest(createRequest)
	if err != nil {
		t.Fatalf("Error encoding create resource request: %s", err)
	}

	// ... the message arrives and is decoded...
	recoveredCreateRequest, err := DecodeUpdateRequest(messageRawData)
	if err != nil {
		t.Fatalf("Error decoding create resource request: %s", err)
	}

	// ... but verification should fail because it is not signed!
	if err := recoveredCreateRequest.Verify(); err == nil {
		t.Fatal("Expected Verify to fail since the message is not signed")
	}

	// We now assume that the resource was created and propagated. With rootAddr we can retrieve the resource metadata
	// and recover the information above. To sign an update, we need the rootAddr and the metaHash to construct
	// proof of ownership

	metaHash := createRequest.metaHash
	rootAddr := createRequest.rootAddr
	const expectedSignature = "0x7862349bc47bc985ca25f23f7aa61954d27ef94f750a9a30b1fb94e42370b5c83e92ff8a1bf71fec4bff64c23b535d7365700334826e0a4c4457b3e27576c5bc00"
	const expectedJSON = `{"rootAddr":"0x612dd6ea96897ebed0373432ab5ed902ce811bde8ad7dcf51300007188730a4a","metaHash":"0x1672f55af019dd612e594ee084540c72d6b82073eadf9cb2cc7721f50c833f7d","version":1,"period":7,"data":"0x5468697320686f75722773207570646174653a20537761726d2039392e3020686173206265656e2072656c656173656421","multiHash":false}`

	//Put together an unsigned update request that we will serialize to send it to the signer.
	data := []byte("This hour's update: Swarm 99.0 has been released!")
	request := &Request{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				updateHeader: updateHeader{
					UpdateLookup: UpdateLookup{
						period:   7,
						version:  1,
						rootAddr: rootAddr,
					},
					multihash: false,
					metaHash:  metaHash,
				},
				data: data,
			},
		},
	}

	messageRawData, err = EncodeUpdateRequest(request)
	if err != nil {
		t.Fatalf("Error encoding update request: %s", err)
	}

	equalJSON, err := areEqualJSON(string(messageRawData), expectedJSON)
	if err != nil {
		t.Fatalf("Error decoding update request JSON: %s", err)
	}
	if !equalJSON {
		t.Fatalf("Received a different JSON message. Expected %s, got %s", expectedJSON, string(messageRawData))
	}

	// now the encoded message messageRawData is sent over the wire and arrives to the signer

	//Attempt to extract an UpdateRequest out of the encoded message
	recoveredRequest, err := DecodeUpdateRequest(messageRawData)
	if err != nil {
		t.Fatalf("Error decoding update request: %s", err)
	}

	//sign the request and see if it matches our predefined signature above.
	if err := recoveredRequest.Sign(signer); err != nil {
		t.Fatalf("Error signing request: %s", err)
	}

	compareByteSliceToExpectedHex(t, "signature", recoveredRequest.signature[:], expectedSignature)

	// mess with the signature and see what happens. To alter the signature, we briefly decode it as JSON
	// to alter the signature field.
	var j updateRequestJSON
	if err := json.Unmarshal([]byte(expectedJSON), &j); err != nil {
		t.Fatal("Error unmarshalling test json, check expectedJSON constant")
	}
	j.Signature = "Certainly not a signature"
	corruptMessage, _ := json.Marshal(j) // encode the message with the bad signature
	_, err = DecodeUpdateRequest(corruptMessage)
	if err == nil {
		t.Fatal("Expected DecodeUpdateRequest to fail when trying to interpret a corrupt message with an invalid signature")
	}

	// Now imagine Evil Bob (why always Bob, poor Bob) attempts to update Charlie's resource,
	// signing a message with his private key
	if err := request.Sign(falseSigner); err != nil {
		t.Fatalf("Error signing: %s", err)
	}

	// Now Bob encodes the message to send it over the wire...
	messageRawData, err = EncodeUpdateRequest(request)
	if err != nil {
		t.Fatalf("Error encoding message:%s", err)
	}

	// ... the message arrives to our Swarm node and it is decoded.
	recoveredRequest, err = DecodeUpdateRequest(messageRawData)
	if err != nil {
		t.Fatalf("Error decoding message:%s", err)
	}

	// Before discovering Bob's misdemeanor, let's see what would happen if we mess
	// with the signature big time to see if Verify catches it
	savedSignature := *recoveredRequest.signature                               // save the signature for later
	binary.LittleEndian.PutUint64(recoveredRequest.signature[5:], 556845463424) // write some random data to break the signature
	if err = recoveredRequest.Verify(); err == nil {
		t.Fatal("Expected Verify to fail on corrupt signature")
	}

	// restore the Evil Bob's signature from corruption
	*recoveredRequest.signature = savedSignature

	// Now the signature is not corrupt, however Verify should now fail because Bob doesn't own the resource
	if err = recoveredRequest.Verify(); err == nil {
		t.Fatalf("Expected Verify to fail because this resource belongs to Charlie, not Bob the attacker:%s", err)
	}

	// Sign with our friend Charlie's private key
	if err := recoveredRequest.Sign(signer); err != nil {
		t.Fatalf("Error signing with the correct private key: %s", err)
	}

	// And now, Verify should work since this resource belongs to Charlie
	if err = recoveredRequest.Verify(); err != nil {
		t.Fatalf("Error verifying that Charlie, the good guy, can sign his resource:%s", err)
	}
}
