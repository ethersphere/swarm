package mru

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

	// Note: the following is signed by Charlie (see private key in resource_test.go)
	// the rootAddr refers to a metadata chunk to Charlie's address that we used in resourcemetadata_test.go
	// Therefore, only Charlie can sign resource updates for this resource.

	metaHash, _ := hexutil.Decode("0x38e401814e98b251612e40f070fddb756315705fa8f674b8ab00b2b5fa091988")
	rootAddr, _ := hexutil.Decode("0xa884c9583d9f86e8009bfd5fe7d892790071c2d6cf8acd2c3e16e5f17e9b143e")
	const expectedSignature = "0x83bfd16ad1fd37208f4024d230c74543a6f1d35829d77b8dc189acb8e45ae00002d1649c7f9dbca6eb916e11fab0b456a2425b5cccb0acb2cde80c019f3da66c00"

	const expectedJSON = `{"ownerAddr":"0x0000000000000000000000000000000000000000","rootAddr":"0xa884c9583d9f86e8009bfd5fe7d892790071c2d6cf8acd2c3e16e5f17e9b143e","metaHash":"0x38e401814e98b251612e40f070fddb756315705fa8f674b8ab00b2b5fa091988","version":1,"period":7,"data":"0x5468697320686f75722773207570646174653a20537761726d2039392e3020686173206265656e2072656c656173656421","multiHash":false}`

	signer := newCharlieSigner()  //Charlie, our good guy
	falseSigner := newBobSigner() //Bob will play the bad guy again

	data := []byte("This hour's update: Swarm 99.0 has been released!")
	request := &UpdateRequest{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceData: resourceData{
				version:   1,
				period:    7,
				multihash: false,
				data:      data,
				metaHash:  metaHash,
				rootAddr:  rootAddr,
			},
		},
	}

	messageRawData, err := EncodeUpdateRequest(request)
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

	recoveredRequest, err := DecodeUpdateRequest(messageRawData)
	if err != nil {
		t.Fatalf("Error decoding update request: %s", err)
	}

	if err := recoveredRequest.Sign(signer); err != nil {
		t.Fatalf("Error signing request: %s", err)
	}

	compareByteSliceToExpectedHex(t, "signature", recoveredRequest.signature[:], expectedSignature)

	// mess with the signature
	var j updateRequestJSON
	if err := json.Unmarshal([]byte(expectedJSON), &j); err != nil {
		t.Fatal("Error unmarshalling test json, check expectedJSON constant")
	}
	j.Signature = "Certainly not a signature"
	corruptMessage, _ := json.Marshal(j)
	_, err = DecodeUpdateRequest(corruptMessage)
	if err == nil {
		t.Fatal("Expected DecodeUpdateRequest to fail when trying to interpret a corrupt message with an invalid signature")
	}

	// Now encode a signed request to see if it works
	// note that we are signing it with the attacker's private key

	if err := request.Sign(falseSigner); err != nil {
		t.Fatalf("Error signing: %s", err)
	}

	messageRawData, err = EncodeUpdateRequest(request)
	if err != nil {
		t.Fatalf("Error encoding message:%s", err)
	}

	recoveredRequest, err = DecodeUpdateRequest(messageRawData)
	if err != nil {
		t.Fatalf("Error decoding message:%s", err)
	}

	//mess with the signature big time to see if Verify catches it
	saveSignature := *recoveredRequest.signature
	binary.LittleEndian.PutUint64(recoveredRequest.signature[5:], 556845463424)
	if err = recoveredRequest.Verify(); err == nil {
		t.Fatal("Expected Verify to fail on corrupt signature")
	}

	// restore the signature with the not corrupt attacker's signature
	*recoveredRequest.signature = saveSignature
	if err = recoveredRequest.Verify(); err == nil {
		t.Fatalf("Expected Verify to fail because this resource belongs to Charlie, not Bob the attacker:%s", err)
	}

	//Sign with our friend Charlie's private key
	if err := recoveredRequest.Sign(signer); err != nil {
		t.Fatalf("Error signing with the correct private key: %s", err)
	}

	if err = recoveredRequest.Verify(); err != nil {
		t.Fatalf("Error verifying that Charlie, the good guy, can sign his resource:%s", err)
	}
}
