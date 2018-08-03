package mru

import (
	"fmt"
	"hash"
	"strconv"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/swarm/storage"
)

// View represents a particular user's view of a resource
type View struct {
	Resource `json:"resource"`
	User     common.Address `json:"user"`
}

// View layout:
// ResourceLength bytes
// userAddr common.AddressLength bytes
const viewLength = ResourceLength + common.AddressLength

// NewView build a new resource "point of view" out of the provided resource and user
func NewView(resource *Resource, user common.Address) *View {
	return &View{
		Resource: *resource,
		User:     user,
	}
}

// mapKey calculates a unique id for this view for the cache map in `Handler`
func (u *View) mapKey() uint64 {
	serializedData := make([]byte, viewLength)
	u.binaryPut(serializedData)
	hasher := hashPool.Get().(hash.Hash)
	defer hashPool.Put(hasher)
	hasher.Reset()
	hasher.Write(serializedData)
	hash := hasher.Sum(nil)
	return *(*uint64)(unsafe.Pointer(&hash[0]))
}

// binaryPut serializes this View instance into the provided slice
func (u *View) binaryPut(serializedData []byte) error {
	if len(serializedData) != viewLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to serialize View. Expected %d, got %d", viewLength, len(serializedData))
	}
	var cursor int
	if err := u.Resource.binaryPut(serializedData[cursor : cursor+ResourceLength]); err != nil {
		return err
	}
	cursor += ResourceLength

	copy(serializedData[cursor:cursor+common.AddressLength], u.User[:])
	cursor += common.AddressLength

	return nil
}

// binaryLength returns the expected size of this structure when serialized
func (u *View) binaryLength() int {
	return viewLength
}

// binaryGet restores the current instance from the information contained in the passed slice
func (u *View) binaryGet(serializedData []byte) error {
	if len(serializedData) != viewLength {
		return NewErrorf(ErrInvalidValue, "Incorrect slice size to read View. Expected %d, got %d", viewLength, len(serializedData))
	}

	var cursor int
	if err := u.Resource.binaryGet(serializedData[cursor : cursor+ResourceLength]); err != nil {
		return err
	}
	cursor += ResourceLength

	copy(u.User[:], serializedData[cursor:cursor+common.AddressLength])
	cursor += common.AddressLength

	return nil
}

// Hex serializes the View to a hex string
func (u *View) Hex() string {
	serializedData := make([]byte, viewLength)
	u.binaryPut(serializedData)
	return hexutil.Encode(serializedData)
}

// FromValues deserializes this instance from a string key-value store
// useful to parse query strings
func (u *View) FromValues(values Values) error {
	startTime, err := strconv.ParseUint(values.Get("starttime"), 10, 64)
	if err != nil {
		return err
	}
	frequency, err := strconv.ParseUint(values.Get("frequency"), 10, 64)
	if err != nil {
		return err
	}
	topic := values.Get("topic")
	if topic != "" {
		if err = u.Topic.FromHex(values.Get("topic")); err != nil {
			return err
		}
	} else { // see if the user set name and relatedcontent
		name := values.Get("name")
		relatedContent, _ := hexutil.Decode(values.Get("relatedcontent"))
		if len(relatedContent) > 0 && len(relatedContent) < storage.KeyLength {
			return NewErrorf(ErrInvalidValue, "relatedcontent field must be a hex-encoded byte array exactly %d bytes long", storage.KeyLength)
		}
		u.Topic = NewTopic(name, relatedContent[:storage.KeyLength])
	}
	u.User = common.HexToAddress(values.Get("user"))
	u.Frequency = frequency
	u.StartTime.Time = startTime
	return nil
}

// ToValues serializes this structure into the provided string key-value store
// useful to build query strings
func (u *View) ToValues(values Values) {
	values.Set("starttime", fmt.Sprintf("%d", u.StartTime.Time))
	values.Set("frequency", fmt.Sprintf("%d", u.Frequency))
	values.Set("topic", u.Topic.Hex())
	values.Set("user", u.User.Hex())
}
