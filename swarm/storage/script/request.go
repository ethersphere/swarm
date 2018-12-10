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

type Request struct {
	ScriptKey vm.Script
	ScriptSig vm.Script
	Payload   []byte
}

var ErrRequestTooBig = fmt.Errorf("Request is too big. Max size is %d", chunk.DefaultSize)

// Addr calculates the feed update chunk address corresponding to this ID
func (r *Request) Addr() storage.Address {
	hasher := sha3.NewKeccak256()
	hasher.Write(r.ScriptKey)
	return hasher.Sum(nil)
}

func (r *Request) Verify() error {
	engine, err := vm.NewEngine(r.ScriptKey, r.ScriptSig, r.Payload, vm.ScriptFlags(0))
	if err != nil {
		return err
	}
	return engine.Execute()
}

func (r *Request) MarshalBinary() ([]byte, error) {
	binaryLength := len(r.ScriptKey) + len(r.ScriptSig) + len(r.Payload) + 3*2
	if binaryLength > chunk.DefaultSize {
		return nil, ErrRequestTooBig
	}
	buf := new(bytes.Buffer)
	for _, arr := range [][]byte{r.ScriptKey, r.ScriptSig, r.Payload} {
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(arr))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(arr); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (r *Request) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	dataRead := 0
	var content [][]byte
	for len(content) < 3 && buf.Len() > 0 {
		var length uint16
		if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
			return err
		}
		dataRead += 2
		if int(length) > buf.Len() {
			return errors.New("Incorrect data length")
		}
		arr := make([]byte, length)
		if _, err := buf.Read(arr); err != nil {
			return err
		}
		content = append(content, arr)
	}
	if len(content) < 3 {
		return errors.New("Incorrect script chunk")
	}
	r.ScriptKey = content[0]
	r.ScriptSig = content[1]
	r.Payload = content[2]

	return nil
}
