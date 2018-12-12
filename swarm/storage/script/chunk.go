package script

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

// Chunk represents a self-validating chunk
type Chunk struct {
	addr      storage.Address
	sdata     []byte
	scriptKey []byte
	scriptSig []byte
	payload   []byte
}

// ErrChunkTooBig is returned when scripts + payload exceed the max amount of data that fits in a chunk
var ErrChunkTooBig = fmt.Errorf("Content for chunk is too big. Max size is %d", chunk.DefaultSize)

// ErrChunkAddressMismatch is returned when the ScriptKey hash does not match the chunk address
var ErrChunkAddressMismatch = errors.New("Chunk address mismatch")

// ErrIncorrectChunkEncoding is returned when the chunk cannot be decoded due to encoding mistakes
// or simply when this is not a script chunk
var ErrIncorrectChunkEncoding = errors.New("Incorrect chunk encoding")

// NewChunk builds a new bzz-script chunk
func NewChunk(ScriptKey, ScriptSig vm.Script, Payload []byte) (*Chunk, error) {
	binaryLength := len(ScriptKey) + len(ScriptSig) + len(Payload) + 3*2
	if binaryLength > chunk.DefaultSize {
		return nil, ErrChunkTooBig
	}
	chunk := new(Chunk)
	sdata := make([]byte, binaryLength)
	chunk.sdata = sdata
	offset := 0

	s := []*[]byte{&chunk.scriptKey, &chunk.scriptSig, &chunk.payload}
	for i, src := range [][]byte{ScriptKey, ScriptSig, Payload} {
		length := len(src)
		binary.LittleEndian.PutUint16(sdata[offset:offset+2], uint16(length))
		offset += 2
		dst := sdata[offset : offset+length]
		offset += length
		copy(dst, src)
		*s[i] = dst
	}
	chunk.calcAddress()
	return chunk, nil
}

// calcAddress calculates the chunk address according to the scriptKey.
func (c *Chunk) calcAddress() {
	hasher := sha3.NewKeccak256()
	hasher.Write(c.scriptKey)
	c.addr = hasher.Sum(nil)
}

// Address returns the chunk's address. Implements storage.Chunk.Address()
func (c *Chunk) Address() storage.Address {
	return c.addr
}

// Data implements the storage.Chunk.Data() method and returns the chunk byte contents
func (c *Chunk) Data() []byte {
	return c.sdata
}

// Verify executes the validation script. Returns nil if the script
// returns true.
func (c *Chunk) Verify(addr storage.Address) error {
	if !bytes.Equal(addr, c.addr) {
		return ErrChunkAddressMismatch
	}
	engine, err := vm.NewEngine(c.scriptKey, c.scriptSig, c.payload, vm.ScriptFlags(0))
	if err != nil {
		return err
	}
	return engine.Execute()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler to rebuild
// a chunk out of a binary representation
func (c *Chunk) UnmarshalBinary(data []byte) error {
	if len(data) > chunk.DefaultSize {
		return ErrChunkTooBig
	}
	s := []*[]byte{&c.scriptKey, &c.scriptSig, &c.payload}
	dataLength := len(data)
	for i, offset := 0, 0; offset < dataLength && i < 3; i++ {
		length := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if length > dataLength-offset {
			return ErrIncorrectChunkEncoding
		}
		*s[i] = data[offset : offset+length]
		offset += length
	}
	c.sdata = data
	c.calcAddress()
	return nil
}

// chunkJSON is a helper struct to serialize a chunk as JSON
type chunkJSON struct {
	Address   hexutil.Bytes `json:"address,omitempty"`
	ScriptKey vm.Script     `json:"scriptKey"`
	ScriptSig vm.Script     `json:"scriptSig"`
	Data      hexutil.Bytes `json:"data"`
}

// MarshalJSON implements the json.Marshaller interface
func (c *Chunk) MarshalJSON() ([]byte, error) {
	return json.Marshal(&chunkJSON{
		ScriptKey: c.scriptKey,
		ScriptSig: c.scriptSig,
		Data:      c.payload,
		Address:   hexutil.Bytes(c.addr),
	})
}

// UnmarshalJSON implements the json.Unmarshaller interface
func (c *Chunk) UnmarshalJSON(data []byte) error {
	var cj chunkJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return err
	}
	chunk, err := NewChunk(cj.ScriptKey, cj.ScriptSig, cj.Data)
	if err != nil {
		return err
	}
	if len(cj.Address) != 0 && !bytes.Equal(chunk.addr, cj.Address) {
		return ErrChunkAddressMismatch
	}
	*c = *chunk
	return nil
}

func (c *Chunk) SpanBytes() []byte {
	panic("unused")
}

func (c *Chunk) Span() int64 {
	panic("unused")
}

func (c *Chunk) Payload() []byte {
	panic("unused")
}
