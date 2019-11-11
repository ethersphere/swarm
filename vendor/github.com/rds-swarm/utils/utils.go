package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// DomainToHashedByteArray takes a string containing a domain, hashes it and returns it as an array of 32 bytes.
func DomainToHashedByteArray(domain string) [32]byte {
	var byteArrayAddress [32]byte

	hashedAddress := RnsNode(domain)
	byteSliceAddress := hashedAddress.Bytes()
	copy(byteArrayAddress[:], byteSliceAddress[:32])

	return byteArrayAddress
}

// RnsNode takes a string containing a domain, hashes it and returns it as a Keccak256Hash.
func RnsNode(name string) common.Hash {
	parentNode, parentLabel := rnsParentNode(name)
	return crypto.Keccak256Hash(parentNode[:], parentLabel[:])
}

func rnsParentNode(name string) (common.Hash, common.Hash) {
	parts := strings.SplitN(name, ".", 2)
	label := crypto.Keccak256Hash([]byte(parts[0]))
	if len(parts) == 1 {
		return [32]byte{}, label
	}
	parentNode, parentLabel := rnsParentNode(parts[1])
	return crypto.Keccak256Hash(parentNode[:], parentLabel[:]), label
}
