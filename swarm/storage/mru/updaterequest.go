package mru

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

const signatureLength = 65

// Signature is an alias for a static byte array with the size of a signature
type Signature [signatureLength]byte

type mruRequestJSON struct {
	Name      string `json:"name"`
	Frequency uint64 `json:"frequency"`
	StartTime uint64 `json:"startTime,omitempty"`
	OwnerAddr string `json:"ownerAddr"`
	RootAddr  string `json:"rootAddr,omitempty"`
	MetaHash  string `json:"metaHash,omitempty"`
	Version   uint32 `json:"version"`
	Period    uint32 `json:"period"`
	Data      string `json:"data,omitempty"`
	Multihash bool   `json:"multiHash"`
	Signature string `json:"signature,omitempty"`
}

type SignedResourceUpdate struct {
	resourceUpdate
	signature *Signature
	key       storage.Address
}

type UpdateRequest struct {
	SignedResourceUpdate
	resourceMetadata
}

func (mj *mruRequestJSON) decode() (smr *UpdateRequest, err error) {

	// make sure name only contains ascii values
	if !isSafeName(mj.Name) {
		return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s'", mj.Name))
	}

	mr := &UpdateRequest{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				version:   mj.Version,
				period:    mj.Period,
				multihash: mj.Multihash,
			},
		},
		resourceMetadata: resourceMetadata{
			name:      mj.Name,
			frequency: mj.Frequency,
			startTime: mj.StartTime,
		},
	}

	ownerAddrBytes, err := hexutil.Decode(mj.OwnerAddr)
	if err != nil || len(ownerAddrBytes) != common.AddressLength {
		return nil, NewError(ErrInvalidValue, "Cannot decode ownerAddr")
	}
	copy(mr.ownerAddr[:], ownerAddrBytes)

	if mj.Data != "" {
		mr.data, err = hexutil.Decode(mj.Data)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode data")
		}
	}
	if mj.RootAddr != "" {
		mr.rootAddr, err = hexutil.Decode(mj.RootAddr)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode rootAddr")
		}
	}

	if mj.MetaHash != "" {
		mr.metaHash, err = hexutil.Decode(mj.MetaHash)
		if err != nil {
			return nil, NewError(ErrInvalidValue, "Cannot decode metaHash")
		}
	}

	if mj.Signature != "" {
		sigBytes, err := hexutil.Decode(mj.Signature)
		if err != nil || len(sigBytes) != signatureLength {
			return nil, NewError(ErrInvalidSignature, "Cannot decode signature")
		}
		mr.signature = new(Signature)
		copy(mr.signature[:], sigBytes)
	}
	return mr, nil
}

func NewUpdateRequest(name string, frequency, startTime uint64, ownerAddr common.Address, data []byte, multihash bool) (*UpdateRequest, error) {
	if !isSafeName(name) {
		return nil, NewError(ErrInvalidValue, fmt.Sprintf("Invalid name: '%s' when creating NewMruRequest", name))
	}

	if startTime == 0 {
		startTime = uint64(time.Now().Unix())
	}

	updateRequest := &UpdateRequest{
		SignedResourceUpdate: SignedResourceUpdate{
			resourceUpdate: resourceUpdate{
				version:   1,
				period:    1,
				data:      data,
				multihash: multihash,
			},
		},
		resourceMetadata: resourceMetadata{
			name:      name,
			frequency: frequency,
			startTime: startTime,
			ownerAddr: ownerAddr,
		},
	}

	updateRequest.rootAddr, updateRequest.metaHash, _ = updateRequest.resourceMetadata.hash()

	return updateRequest, nil
}

func (mr *UpdateRequest) Verify() error {
	key := resourceHash(mr.period, mr.version, mr.rootAddr)
	digest := keyDataHash(key, mr.metaHash, mr.data)
	// get the address of the signer (which also checks that it's a valid signature)
	addr, err := getAddressFromDataSig(digest, *mr.signature)
	if err != nil {
		return err
	}
	if addr != mr.ownerAddr {
		return NewError(ErrInvalidSignature, "Signature address does not match with ownerAddr")
	}
	_, err = mr.SignedResourceUpdate.Verify(key)
	return err
}

func (mr *UpdateRequest) Sign(signer Signer) error {
	if err := mr.SignedResourceUpdate.Sign(signer); err != nil {
		return err
	}
	mr.ownerAddr = signer.Address()
	return nil
}

func (mr *UpdateRequest) Frequency() uint64 {
	return mr.frequency
}

func (mr *UpdateRequest) Name() string {
	return mr.name
}

func (mr *UpdateRequest) Multihash() bool {
	return mr.multihash
}

func (mr *UpdateRequest) Version() uint32 {
	return mr.version
}
func (mr *UpdateRequest) Period() uint32 {
	return mr.period
}
func (mr *UpdateRequest) StartTime() uint64 {
	return mr.startTime
}
func (mr *UpdateRequest) OwnerAddr() common.Address {
	return mr.ownerAddr
}
func (mr *UpdateRequest) RootAddr() storage.Address {
	return mr.rootAddr
}

func (mu *SignedResourceUpdate) Verify(key storage.Address) (ownerAddr common.Address, err error) {
	if len(mu.data) == 0 {
		return ownerAddr, NewError(ErrInvalidValue, "I refuse to waste swarm space for updates with empty values, amigo (data length is 0)")
	}
	if mu.signature == nil {
		return ownerAddr, NewError(ErrInvalidSignature, "Missing signature field")
	}

	digest := keyDataHash(key, mu.metaHash, mu.data)

	// get the address of the signer (which also checks that it's a valid signature)
	ownerAddr, err = getAddressFromDataSig(digest, *mu.signature)
	if err != nil {
		return ownerAddr, err
	}

	if !bytes.Equal(key, resourceHash(mu.period, mu.version, mu.rootAddr)) {
		return ownerAddr, NewError(ErrInvalidSignature, "Signature address does not match with ownerAddr")
	}
	/*	if !verifyResourceOwnership(ownerAddr, mu.metaHash, mu.rootAddr) {
			return NewError(ErrInvalidSignature, "Signature address does not match with ownerAddr")
		}
	*/
	mu.key = key
	return ownerAddr, nil
}

func (mu *SignedResourceUpdate) Sign(signer Signer) error {

	key := resourceHash(mu.period, mu.version, mu.rootAddr)

	digest := keyDataHash(key, mu.metaHash, mu.data)
	signature, err := signer.Sign(digest)
	if err != nil {
		return err
	}
	ownerAddress, err := getAddressFromDataSig(digest, signature)
	if err != nil {
		return NewError(ErrInvalidSignature, "Error verifying signature")
	}
	if ownerAddress != signer.Address() {
		return NewError(ErrInvalidSignature, "Signer address does not match private key")
	}
	mu.signature = &signature
	mu.key = key
	return nil
}

func (mr *UpdateRequest) SetData(data []byte) {
	mr.signature = nil
	mr.data = data
	mr.frequency = 0 //mark as update
}

func DecodeMruRequest(rawData []byte) (*UpdateRequest, error) {
	var requestJSON mruRequestJSON
	if err := json.Unmarshal(rawData, &requestJSON); err != nil {
		return nil, err
	}
	return requestJSON.decode()
}

func EncodeMruRequest(mruRequest *UpdateRequest) (rawData []byte, err error) {
	var signatureString, dataHashString, rootAddrString, metaHashString string
	if mruRequest.signature != nil {
		signatureString = hexutil.Encode(mruRequest.signature[:])
	}
	if mruRequest.data != nil {
		dataHashString = hexutil.Encode(mruRequest.data)
	}
	if mruRequest.rootAddr != nil {
		rootAddrString = hexutil.Encode(mruRequest.rootAddr)
	}
	if mruRequest.metaHash != nil {
		metaHashString = hexutil.Encode(mruRequest.metaHash)
	}

	requestJSON := &mruRequestJSON{
		Name:      mruRequest.name,
		Frequency: mruRequest.frequency,
		StartTime: mruRequest.startTime,
		Version:   mruRequest.version,
		Period:    mruRequest.period,
		OwnerAddr: hexutil.Encode(mruRequest.ownerAddr[:]),
		Data:      dataHashString,
		Multihash: mruRequest.multihash,
		Signature: signatureString,
		RootAddr:  rootAddrString,
		MetaHash:  metaHashString,
	}

	return json.Marshal(requestJSON)
}
