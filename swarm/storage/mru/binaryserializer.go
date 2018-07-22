package mru

import "github.com/ethereum/go-ethereum/common/hexutil"

type binarySerializer interface {
	binaryPut(serializedData []byte) error
	binaryLength() int
	binaryGet(serializedData []byte) error
}

func BinHex(bin binarySerializer) string {
	b := make([]byte, bin.binaryLength())
	bin.binaryPut(b)
	return hexutil.Encode(b)
}
