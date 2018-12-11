package script

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/script/hexbytes"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

type Chunk struct {
	addr      storage.Address
	sdata     []byte
	scriptKey []byte
	scriptSig []byte
	payload   []byte
}

var ErrChunkTooBig = fmt.Errorf("Content for chunk is too big. Max size is %d", chunk.DefaultSize)

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

func (c *Chunk) Address() storage.Address {
	return c.addr
}

func (c *Chunk) Verify(addr storage.Address) error {
	if !bytes.Equal(addr, c.addr) {
		return errors.New("Address mismatch")
	}
	engine, err := vm.NewEngine(c.scriptKey, c.scriptSig, c.payload, vm.ScriptFlags(0))
	if err != nil {
		return err
	}
	return engine.Execute()
}

func (c *Chunk) UnmarshalBinary(data []byte) error {
	s := []*[]byte{&c.scriptKey, &c.scriptSig, &c.payload}
	dataLength := len(data)
	for i, offset := 0, 0; offset < dataLength && i < 3; i++ {
		length := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if length > dataLength-offset {
			return errors.New("Incorrect data length")
		}
		*s[i] = data[offset : offset+length]
		offset += length
	}
	c.sdata = data
	c.calcAddress()
	return nil
}

type chunkJSON struct {
	Address   hexbytes.HexBytes `json:"address,omitempty"`
	ScriptKey vm.Script         `json:"scriptKey"`
	ScriptSig vm.Script         `json:"scriptSig"`
	Data      hexbytes.HexBytes `json:"data"`
}

func (c *Chunk) MarshalJSON() ([]byte, error) {
	return json.Marshal(&chunkJSON{
		ScriptKey: c.scriptKey,
		ScriptSig: c.scriptSig,
		Data:      c.payload,
		Address:   hexbytes.HexBytes(c.addr),
	})
}

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
		return errors.New("Chunk address mismatch")
	}
	*c = *chunk
	return nil
}

func (c *Chunk) Data() []byte {
	return c.sdata
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
