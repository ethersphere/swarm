package encrypt

import (
	"context"

	"github.com/ethersphere/swarm/param"
	"github.com/ethersphere/swarm/storage/encryption"
	"golang.org/x/crypto/sha3"
)

type Encrypt struct {
	e encryption.Encryption
	w param.SectionWriter
}

func New(key []byte, initCtr uint32) *Encrypt {
	return &Encrypt{
		e: encryption.New(key, 0, initCtr, sha3.NewLegacyKeccak256),
	}
}

func (e *Encrypt) Init(_ context.Context, errFunc func(error)) {
}

func (e *Encrypt) Link(writerFunc func() param.SectionWriter) {
	e.w = writerFunc()
}

func (e *Encrypt) Write(index int, b []byte) {
	e.w.Write(index, b)
}

func (e *Encrypt) Reset(ctx context.Context) {
	e.w.Reset(ctx)
}

func (e *Encrypt) Sum(b []byte, length int, span []byte) []byte {
	return e.w.Sum(b, length, span)
}

func (e *Encrypt) DigestSize() int {
	return e.w.DigestSize() + encryption.KeyLength
}

func (e *Encrypt) SectionSize() int {
	return e.w.SectionSize()
}
