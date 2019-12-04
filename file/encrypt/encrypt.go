package encrypt

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"hash"

	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage/encryption"
	"golang.org/x/crypto/sha3"
)

type Encrypt struct {
	key     []byte
	e       encryption.Encryption
	w       param.SectionWriter
	keyHash hash.Hash
	errFunc func(error)
}

func New(key []byte, initCtr uint32) (*Encrypt, error) {
	e := &Encrypt{
		e:       encryption.New(key, 0, initCtr, sha3.NewLegacyKeccak256),
		key:     make([]byte, encryption.KeyLength),
		keyHash: param.HashFunc(),
	}
	if key == nil {
		e.key = make([]byte, encryption.KeyLength)
		c, err := crand.Read(e.key)
		if err != nil {
			return nil, err
		}
		if c < encryption.KeyLength {
			return nil, fmt.Errorf("short read: %d", c)
		}
	} else {
		copy(e.key, key)
	}
	return e, nil
}

func (e *Encrypt) Init(_ context.Context, errFunc func(error)) {
	e.errFunc = errFunc
}

func (e *Encrypt) Link(writerFunc func() param.SectionWriter) {
	e.w = writerFunc()
}

func (e *Encrypt) Write(index int, b []byte) {
	cipherText, err := e.e.Encrypt(b)
	if err != nil {
		e.errFunc(err)
		return
	}
	e.w.Write(index, cipherText)
}

func (e *Encrypt) Reset(ctx context.Context) {
	e.w.Reset(ctx)
}

func (e *Encrypt) Sum(b []byte, length int, span []byte) []byte {
	oldKey := make([]byte, 32)
	copy(oldKey, e.key)
	e.keyHash.Reset()
	e.keyHash.Write(e.key)
	copy(e.key, e.keyHash.Sum(nil))
	s := e.w.Sum(b, length, span)
	return append(oldKey, s...)
}

func (e *Encrypt) DigestSize() int {
	return e.w.DigestSize() + encryption.KeyLength
}

func (e *Encrypt) SectionSize() int {
	return e.w.SectionSize()
}
