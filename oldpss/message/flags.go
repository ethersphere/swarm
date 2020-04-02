package message

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
)

// Flags represents the possible PSS message flags
type Flags struct {
	Raw       bool // message is flagged as raw or with external encryption
	Symmetric bool // message is symmetrically encrypted
}

const flagsLength = 1
const flagSymmetric = 1 << 0
const flagRaw = 1 << 1

// ErrIncorrectFlagsFieldLength is returned when the incoming flags field length is incorrect
var ErrIncorrectFlagsFieldLength = errors.New("Incorrect flags field length in message")

// DecodeRLP implements the rlp.Decoder interface
func (f *Flags) DecodeRLP(s *rlp.Stream) error {
	flagsBytes, err := s.Bytes()
	if err != nil {
		return err
	}
	if len(flagsBytes) != flagsLength {
		return ErrIncorrectFlagsFieldLength
	}
	f.Symmetric = flagsBytes[0]&flagSymmetric != 0
	f.Raw = flagsBytes[0]&flagRaw != 0
	return nil
}

// EncodeRLP implements the rlp.Encoder interface
func (f *Flags) EncodeRLP(w io.Writer) error {
	var flags byte
	if f.Raw {
		flags |= flagRaw
	}
	if f.Symmetric {
		flags |= flagSymmetric
	}

	return rlp.Encode(w, []byte{flags})
}
