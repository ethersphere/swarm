// Copyright 2016 The go-ethereum Authors
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

package storage

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/swarm/bmt"
)

const MaxPO = 16
const AddressLength = 32

type Hasher func() hash.Hash
type SwarmHasher func() SwarmHash

// Peer is the recorded as Source on the chunk
// should probably not be here? but network should wrap chunk object
type Peer interface{}

type Address []byte

func (x Address) Size() uint {
	return uint(len(x))
}

func (x Address) isEqual(y Address) bool {
	return bytes.Equal(x, y)
}

func (h Address) bits(i, j uint) uint {
	ii := i >> 3
	jj := i & 7
	if ii >= h.Size() {
		return 0
	}

	if jj+j <= 8 {
		return uint((h[ii] >> jj) & ((1 << j) - 1))
	}

	res := uint(h[ii] >> jj)
	jj = 8 - jj
	j -= jj
	for j != 0 {
		ii++
		if j < 8 {
			res += uint(h[ii]&((1<<j)-1)) << jj
			return res
		}
		res += uint(h[ii]) << jj
		jj += 8
		j -= 8
	}
	return res
}

func Proximity(one, other []byte) (ret int) {
	b := (MaxPO-1)/8 + 1
	if b > len(one) {
		b = len(one)
	}
	m := 8
	for i := 0; i < b; i++ {
		oxo := one[i] ^ other[i]
		if i == b-1 {
			m = MaxPO % 8
		}
		for j := 0; j < m; j++ {
			if (oxo>>uint8(7-j))&0x01 != 0 {
				return i*8 + j
			}
		}
	}
	return MaxPO
}

func IsZeroAddr(addr Address) bool {
	return len(addr) == 0 || bytes.Equal(addr, ZeroAddr)
}

var ZeroAddr = Address(common.Hash{}.Bytes())

func MakeHashFunc(hash string) SwarmHasher {
	switch hash {
	case "SHA256":
		return func() SwarmHash { return &HashWithLength{crypto.SHA256.New()} }
	case "SHA3":
		return func() SwarmHash { return &HashWithLength{sha3.NewKeccak256()} }
	case "BMT":
		return func() SwarmHash {
			hasher := sha3.NewKeccak256
			pool := bmt.NewTreePool(hasher, bmt.DefaultSegmentCount, bmt.DefaultPoolSize)
			return bmt.New(pool)
		}
	}
	return nil
}

func (addr Address) Hex() string {
	return fmt.Sprintf("%064x", []byte(addr[:]))
}

func (addr Address) Log() string {
	if len(addr[:]) < 8 {
		return fmt.Sprintf("%x", []byte(addr[:]))
	}
	return fmt.Sprintf("%016x", []byte(addr[:8]))
}

func (addr Address) String() string {
	return fmt.Sprintf("%064x", []byte(addr)[:])
}

func (addr Address) MarshalJSON() (out []byte, err error) {
	return []byte(`"` + addr.String() + `"`), nil
}

func (addr *Address) UnmarshalJSON(value []byte) error {
	s := string(value)
	*addr = make([]byte, 32)
	h := common.Hex2Bytes(s[1 : len(s)-1])
	copy(*addr, h)
	return nil
}

type AddressCollection []Address

func NewAddressCollection(l int) AddressCollection {
	return make(AddressCollection, l)
}

func (c AddressCollection) Len() int {
	return len(c)
}

func (c AddressCollection) Less(i, j int) bool {
	return bytes.Compare(c[i], c[j]) == -1
}

func (c AddressCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Chunk interface implemented by context.Contexts and data chunks
type Chunk interface {
	Address() Address
	Payload() []byte
	SpanBytes() []byte
	Span() int64
	Data() []byte
	Chunk() *chunk
}

type chunk struct {
	addr  Address
	sdata []byte
	span  int64
}

func NewChunk(addr Address, data []byte) *chunk {
	return &chunk{
		addr:  addr,
		sdata: data,
	}
}

func (c *chunk) Address() Address {
	return c.addr
}

func (c *chunk) SpanBytes() []byte {
	return c.sdata[:8]
}

func (c *chunk) Span() int64 {
	// if c.span == 0 {
	c.span = int64(binary.LittleEndian.Uint64(c.sdata[:8]))
	// }
	return c.span
}

func (c *chunk) Data() []byte {
	return c.sdata
}

func (c *chunk) Payload() []byte {
	return c.sdata[8:]
}

func (c *chunk) Chunk() *chunk {
	return c
}

// String() for pretty printing
func (self *chunk) String() string {
	return fmt.Sprintf("Address: %v TreeSize: %v Chunksize: %v", self.addr.Log(), self.span, len(self.sdata))
}

func GenerateRandomChunk(dataSize int64) Chunk {
	hasher := MakeHashFunc(DefaultHash)()
	sdata := make([]byte, dataSize+8)
	rand.Read(sdata[8:])
	binary.LittleEndian.PutUint64(sdata[:8], uint64(dataSize))
	hasher.ResetWithLength(sdata[:8])
	hasher.Write(sdata[8:])
	return NewChunk(hasher.Sum(nil), sdata)
}

func GenerateRandomChunks(dataSize int64, count int) (chunks []Chunk) {
	if dataSize > DefaultChunkSize {
		dataSize = DefaultChunkSize
	}
	for i := 0; i < count; i++ {
		ch := GenerateRandomChunk(DefaultChunkSize)
		chunks = append(chunks, ch)
	}
	return chunks
}

func GenerateRandomData(l int) (r io.Reader, slice []byte) {
	slice, err := ioutil.ReadAll(io.LimitReader(rand.Reader, int64(l)))
	if err != nil {
		panic("rand error")
	}
	// log.Warn("generate random data", "len", len(slice), "data", common.Bytes2Hex(slice))
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return r, slice
}

// Size, Seek, Read, ReadAt
type LazySectionReader interface {
	Size() (int64, error)
	io.Seeker
	io.Reader
	io.ReaderAt
}

type LazyTestSectionReader struct {
	*io.SectionReader
}

func (self *LazyTestSectionReader) Size(chan bool) (int64, error) {
	return self.SectionReader.Size(), nil
}

type StoreParams struct {
	Hash          SwarmHasher `toml:"-"`
	DbCapacity    uint64
	CacheCapacity uint
	BaseKey       []byte
}

func NewDefaultStoreParams() *StoreParams {
	return NewStoreParams(defaultLDBCapacity, defaultCacheCapacity, nil, nil)
}

func NewStoreParams(ldbCap uint64, cacheCap uint, hash SwarmHasher, basekey []byte) *StoreParams {
	if basekey == nil {
		basekey = make([]byte, 32)
	}
	if hash == nil {
		hash = MakeHashFunc(DefaultHash)
	}
	return &StoreParams{
		Hash:          hash,
		DbCapacity:    ldbCap,
		CacheCapacity: cacheCap,
		BaseKey:       basekey,
	}
}

type ChunkData []byte

type Reference []byte

// Putter is responsible to store data and create a reference for it
type Putter interface {
	Put(ChunkData) (Reference, error)
	// RefSize returns the length of the Reference created by this Putter
	RefSize() int64
	// Close is to indicate that no more chunk data will be Put on this Putter
	Close()
	// Wait returns if all data has been store and the Close() was called.
	Wait(ctx context.Context) error
}

// Getter is an interface to retrieve a chunk's data by its reference
type Getter interface {
	Get(context.Context, Reference) (ChunkData, error)
}

// NOTE: this returns invalid data if chunk is encrypted
func (c ChunkData) Size() uint64 {
	return uint64(binary.LittleEndian.Uint64(c[:8]))
}

func (c ChunkData) Data() []byte {
	return c[8:]
}

type ChunkValidator interface {
	Validate(addr Address, data []byte) bool
}

// Provides method for validation of content address in chunks
// Holds the corresponding hasher to create the address
type ContentAddressValidator struct {
	Hasher SwarmHasher
}

// Constructor
func NewContentAddressValidator(hasher SwarmHasher) *ContentAddressValidator {
	return &ContentAddressValidator{
		Hasher: hasher,
	}
}

// Validate that the given key is a valid content address for the given data
func (self *ContentAddressValidator) Validate(addr Address, data []byte) bool {
	hasher := self.Hasher()
	hasher.ResetWithLength(data[:8])
	hasher.Write(data[8:])
	hash := hasher.Sum(nil)

	if !bytes.Equal(hash, addr[:]) {
		log.Error("invalid content address", "expected", fmt.Sprintf("%x", hash), "have", addr)
		return false
	}
	return true
}

type ChunkStore interface {
	Put(ch Chunk) (waitToStore func(ctx context.Context) error, err error)
	Get(rctx context.Context, ref Address) (ch Chunk, err error)
	Close()
}
