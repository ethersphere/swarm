package encrypt

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"hash"

	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage/encryption"
	"golang.org/x/crypto/sha3"
)

type Encrypt struct {
	key     []byte
	e       encryption.Encryption
	w       param.SectionWriter
	length  int
	span    int
	keyHash hash.Hash
	errFunc func(error)
}

func New(key []byte, initCtr uint32, hashFunc param.SectionWriterFunc) (*Encrypt, error) {
	if key == nil {
		key = make([]byte, encryption.KeyLength)
		c, err := crand.Read(key)
		if err != nil {
			return nil, err
		}
		if c < encryption.KeyLength {
			return nil, fmt.Errorf("short read: %d", c)
		}
	} else if len(key) != encryption.KeyLength {
		return nil, fmt.Errorf("encryption key must be %d bytes", encryption.KeyLength)
	}
	e := &Encrypt{
		e:       encryption.New(key, 0, initCtr, sha3.NewLegacyKeccak256),
		key:     make([]byte, encryption.KeyLength),
		keyHash: param.HashFunc(),
	}
	copy(e.key, key)
	return e, nil
}

func (e *Encrypt) SetWriter(hashFunc param.SectionWriterFunc) param.SectionWriter {
	e.w = hashFunc(nil)
	return e

}

func (e *Encrypt) Init(_ context.Context, errFunc func(error)) {
	e.errFunc = errFunc
}

func (e *Encrypt) SeekSection(offset int) {
	e.w.SeekSection(offset)
}

func (e *Encrypt) Write(b []byte) (int, error) {
	cipherText, err := e.e.Encrypt(b)
	if err != nil {
		e.errFunc(err)
		return 0, err
	}
	return e.w.Write(cipherText)
}

func (e *Encrypt) Reset() {
	e.w.Reset()
}

func (e *Encrypt) SetLength(length int) {
	e.length = length
	e.w.SetLength(length)
}

func (e *Encrypt) SetSpan(length int) {
	e.span = length
	e.w.SetSpan(length)
}

func (e *Encrypt) Sum(b []byte) []byte {
	// derive new key
	oldKey := make([]byte, encryption.KeyLength)
	copy(oldKey, e.key)
	e.keyHash.Reset()
	e.keyHash.Write(e.key)
	newKey := e.keyHash.Sum(nil)
	copy(e.key, newKey)
	s := e.w.Sum(b)
	log.Trace("key", "key", oldKey, "ekey", e.key, "newkey", newKey)
	return append(oldKey, s...)
}

// DigestSize implements param.SectionWriter
func (e *Encrypt) BlockSize() int {
	return e.Size()
}

// DigestSize implements param.SectionWriter
// TODO: cache these calculations
func (e *Encrypt) Size() int {
	return e.w.Size() + encryption.KeyLength
}

// SectionSize implements param.SectionWriter
func (e *Encrypt) SectionSize() int {
	return e.w.SectionSize()
}

// Branches implements param.SectionWriter
// TODO: cache these calculations
func (e *Encrypt) Branches() int {
	return e.w.Branches() / (e.Size() / e.w.SectionSize())
}
