package mru

import (
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func NewView(resource *Resource, owner common.Address) *View {
	return &View{
		Resource: *resource,
		User:     owner,
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

func (u *View) Hex() string {
	serializedData := make([]byte, viewLength)
	u.binaryPut(serializedData)
	return hexutil.Encode(serializedData)
}

func (u *View) FromURL(url *url.URL) error {
	query := url.Query()
	startTime, err := strconv.ParseUint(query.Get("starttime"), 10, 64)
	if err != nil {
		return err
	}
	frequency, err := strconv.ParseUint(query.Get("frequency"), 10, 64)
	if err != nil {
		return err
	}
	if err = u.Topic.FromHex(query.Get("topic")); err != nil {
		return err
	}
	u.User = common.HexToAddress(query.Get("user"))
	u.Frequency = frequency
	u.StartTime.Time = startTime
	return nil
}

func (u *View) ToURL(url *url.URL) {
	query := url.Query()
	query.Set("starttime", fmt.Sprintf("%d", u.StartTime.Time))
	query.Set("frequency", fmt.Sprintf("%d", u.Frequency))
	query.Set("topic", u.Topic.Hex())
	query.Set("user", u.User.Hex())
	url.RawQuery = query.Encode()
}
