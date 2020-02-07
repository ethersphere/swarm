package param

import "golang.org/x/crypto/sha3"

var (
	HashFunc = sha3.NewLegacyKeccak256
)
