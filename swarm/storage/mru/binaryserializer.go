package mru

import "github.com/ethereum/go-ethereum/common/hexutil"

type binarySerializer interface {
	binaryPut(serializedData []byte) error
	binaryLength() int
	binaryGet(serializedData []byte) error
}

// Hex serializes the structure and converts it to a hex string
func Hex(bin binarySerializer) string {
	b := make([]byte, bin.binaryLength())
	bin.binaryPut(b)
	return hexutil.Encode(b)
}
