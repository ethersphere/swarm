package script

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/chunk"
	"github.com/ethereum/go-ethereum/swarm/storage"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

type Chunk struct {
	sdata     []byte
	addr      storage.Address
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
	chunk.sdata = make([]byte, binaryLength)
	buf := bytes.NewBuffer(chunk.sdata)
	buf.Reset()
	s := []*[]byte{&chunk.scriptKey, &chunk.scriptSig, &chunk.payload}
	for i, arr := range [][]byte{ScriptKey, ScriptSig, Payload} {
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(arr))); err != nil {
			return nil, err
		}
		fmt.Println(buf.Len(), buf.Len()+len(arr))
		*s[i] = chunk.sdata[buf.Len() : buf.Len()+len(arr)]
		if _, err := buf.Write(arr); err != nil {
			return nil, err
		}
	}
	chunk.calcAddress()
	return chunk, nil
}

// calcAddress calculates the chunk address corresponding to this request
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
	var offset int
	for i := 0; offset < len(data) && i < len(s); i++ {
		length := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
		offset += 2
		if length > len(data) {
			return errors.New("Incorrect data length")
		}
		*s[i] = data[offset : offset+length]
		offset += length
	}
	c.sdata = data
	c.calcAddress()
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
