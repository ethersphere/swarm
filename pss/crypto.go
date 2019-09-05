// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.
package pss

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	crand "crypto/rand"
	"errors"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"io"
	mrand "math/rand"
	"strconv"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

// We need a singleton here?
var cryptoBackend defaultCryptoBackend

const (
	aesNonceLength = 12 // in bytes; for more info please see cipher.gcmStandardNonceSize & aesgcm.NonceSize()
)

type CryptoBackend interface {
	// Encrypt-Decrypt
	EncryptAsymmetric(rawBytes []byte, key *ecdsa.PublicKey) ([]byte, error)
	DecryptAsymmetric(rawBytes []byte, key *ecdsa.PrivateKey) ([]byte, error)
	EncryptSymmetric(rawBytes []byte, key []byte) ([]byte, error)
	DecryptSymmetric(rawBytes []byte, key []byte) (decrypted []byte, salt []byte, err error)

	//Signing and hashing
	Keccak256(data ...[]byte) []byte
	Sign(bytes []byte, key *ecdsa.PrivateKey) ([]byte, error) // Calculate hash and returns the signature
	SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error)      // SigToPub returns the public key that created the given signature.

	// Key store functions
	GetSymKey(id string) ([]byte, error)
	GenerateSymKey() (string, error)
	AddSymKeyDirect(bytes []byte) (string, error)

	// Key conversion
	FromECDSAPub(pub *ecdsa.PublicKey) []byte
	ImportECDSAPublic(key *ecdsa.PublicKey) *ecies.PublicKey // ecdsa pub key to ecies pub key

	// Key serialization
	UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error)
	CompressPubkey(pubkey *ecdsa.PublicKey) []byte
}

//Used only in tests
type CryptoUtils interface {
	GenerateKey() (*ecdsa.PrivateKey, error)
	NewKeyPair(ctx context.Context) (string, error)
	GetPrivateKey(id string) (*ecdsa.PrivateKey, error)
}

type defaultCryptoBackend struct {
	whisper *whisper.Whisper
	wapi    *whisper.PublicWhisperAPI
}

func NewCryptoBackend() CryptoBackend {
	w := whisper.New(&whisper.DefaultConfig)
	cryptoBackend = defaultCryptoBackend{
		whisper: w,
		wapi:    whisper.NewPublicWhisperAPI(w),
	}
	return &cryptoBackend
}

func NewCryptoUtils() CryptoUtils {
	if cryptoBackend.whisper == nil {
		NewCryptoBackend()
	}
	return &cryptoBackend
}

// === Encrypt-Decrypt ===

// encryptAsymmetric encrypts a message with a public key.
func (crypto *defaultCryptoBackend) EncryptAsymmetric(rawBytes []byte, key *ecdsa.PublicKey) ([]byte, error) {
	if !validatePublicKey(key) {
		return nil, errors.New("invalid public key provided for asymmetric encryption")
	}
	encrypted, err := crypto.encrypt(crand.Reader, crypto.ImportECDSAPublic(key), rawBytes, nil, nil)
	if err == nil {
		return encrypted, nil
	}
	return nil, err
}

func (crypto *defaultCryptoBackend) DecryptAsymmetric(rawBytes []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	return ecies.ImportECDSA(key).Decrypt(rawBytes, nil, nil)
}

func (crypto *defaultCryptoBackend) EncryptSymmetric(rawBytes []byte, key []byte) ([]byte, error) {
	if !crypto.validateDataIntegrity(key, aesKeyLength) {
		return nil, errors.New("invalid key provided for symmetric encryption, size: " + strconv.Itoa(len(key)))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	salt, err := crypto.generateSecureRandomData(aesNonceLength) // never use more than 2^32 random nonces with a given key
	if err != nil {
		return nil, err
	}
	encrypted := aesgcm.Seal(nil, salt, rawBytes, nil)
	encBytes := append(encrypted, salt...)
	return encBytes, nil
}

// decryptSymmetric decrypts a message with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func (crypto *defaultCryptoBackend) DecryptSymmetric(rawBytes []byte, key []byte) (decrypted []byte, salt []byte, err error) {
	// symmetric messages are expected to contain the 12-byte nonce at the end of the payload
	if len(rawBytes) < aesNonceLength {
		return nil, nil, errors.New("missing salt or invalid payload in symmetric message")
	}
	salt = rawBytes[len(rawBytes)-aesNonceLength:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	decrypted, err = aesgcm.Open(nil, salt, rawBytes[:len(rawBytes)-aesNonceLength], nil)
	if err != nil {
		return nil, nil, err
	}
	return
}

// === Signing and hashing ===

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func (crypto *defaultCryptoBackend) Keccak256(data ...[]byte) []byte {
	return ethCrypto.Keccak256(data...)
}

func (crypto *defaultCryptoBackend) Sign(bytes []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	hash := crypto.Keccak256(bytes)
	signature, err := crypto.signWithHash(hash, key)
	return signature, err
}

func (crypto *defaultCryptoBackend) SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	return ethCrypto.SigToPub(hash, sig)
}

// === Key store functions ===

func (crypto *defaultCryptoBackend) GetSymKey(id string) ([]byte, error) {
	return crypto.whisper.GetSymKey(id)
}

func (crypto *defaultCryptoBackend) GenerateSymKey() (string, error) {
	return crypto.whisper.GenerateSymKey()
}

func (crypto *defaultCryptoBackend) AddSymKeyDirect(bytes []byte) (string, error) {
	return crypto.whisper.AddSymKeyDirect(bytes)
}

// === Key conversion ===

// FromECDSA exports a public key into a binary dump.
func (crypto *defaultCryptoBackend) FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	return ethCrypto.FromECDSAPub(pub)
}

func (crypto *defaultCryptoBackend) ImportECDSAPublic(key *ecdsa.PublicKey) *ecies.PublicKey {
	return ecies.ImportECDSAPublic(key)
}

// === Key serialization ===

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func (crypto *defaultCryptoBackend) UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	return ethCrypto.UnmarshalPubkey(pub)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func (crypto *defaultCryptoBackend) CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return ethCrypto.CompressPubkey(pubkey)
}

// == private methods ==

// signWithHash calculates an ECDSA signature.
func (crypto *defaultCryptoBackend) signWithHash(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return ethCrypto.Sign(hash, prv)
}

func (crypto *defaultCryptoBackend) encrypt(rand io.Reader, pub *ecies.PublicKey, m, s1, s2 []byte) (ct []byte, err error) {
	return ecies.Encrypt(rand, pub, m, s1, s2)
}

// generateSecureRandomData generates random data where extra security is required.
// The purpose of this function is to prevent some bugs in software or in hardware
// from delivering not-very-random data. This is especially useful for AES nonce,
// where true randomness does not really matter, but it is very important to have
// a unique nonce for every message.
func (crypto *defaultCryptoBackend) generateSecureRandomData(length int) ([]byte, error) {
	x := make([]byte, length)
	y := make([]byte, length)
	res := make([]byte, length)

	_, err := crand.Read(x)
	if err != nil {
		return nil, err
	} else if !crypto.validateDataIntegrity(x, length) {
		return nil, errors.New("crypto/rand failed to generate secure random data")
	}
	_, err = mrand.Read(y)
	if err != nil {
		return nil, err
	} else if !crypto.validateDataIntegrity(y, length) {
		return nil, errors.New("math/rand failed to generate secure random data")
	}
	for i := 0; i < length; i++ {
		res[i] = x[i] ^ y[i]
	}
	if !crypto.validateDataIntegrity(res, length) {
		return nil, errors.New("failed to generate secure random data")
	}
	return res, nil
}

// validateDataIntegrity returns false if the data have the wrong or contains all zeros,
// which is the simplest and the most common bug.
func (crypto *defaultCryptoBackend) validateDataIntegrity(k []byte, expectedSize int) bool {
	if len(k) != expectedSize {
		return false
	}
	if expectedSize > 3 && containsOnlyZeros(k) {
		return false
	}
	return true
}

// ValidatePublicKey checks the format of the given public key.
func validatePublicKey(k *ecdsa.PublicKey) bool {
	return k != nil && k.X != nil && k.Y != nil && k.X.Sign() != 0 && k.Y.Sign() != 0
}

// containsOnlyZeros checks if the data contain only zeros.
func containsOnlyZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// CryptoUtils

func (crypto *defaultCryptoBackend) GenerateKey() (*ecdsa.PrivateKey, error) {
	return ethCrypto.GenerateKey()
}

func (crypto *defaultCryptoBackend) NewKeyPair(ctx context.Context) (string, error) {
	return crypto.wapi.NewKeyPair(ctx)
}

func (crypto *defaultCryptoBackend) GetPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	return crypto.whisper.GetPrivateKey(id)
}
