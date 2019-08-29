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
	"crypto/ecdsa"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

var cryptoBackend defaultCryptoBackend

type CryptoBackend interface {
	GetSymKey(id string) ([]byte, error)
	GenerateSymKey() (string, error)
	AddSymKeyDirect(bytes []byte) (string, error)
	FromECDSAPub(pub *ecdsa.PublicKey) []byte
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

func (crypto *defaultCryptoBackend) GetSymKey(id string) ([]byte, error) {
	return crypto.whisper.GetSymKey(id)
}

func (crypto *defaultCryptoBackend) GenerateSymKey() (string, error) {
	return crypto.whisper.GenerateSymKey()
}

func (crypto *defaultCryptoBackend) AddSymKeyDirect(bytes []byte) (string, error) {
	return crypto.whisper.AddSymKeyDirect(bytes)
}

// FromECDSA exports a public key into a binary dump.
func (crypto *defaultCryptoBackend) FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	return ethCrypto.FromECDSAPub(pub)
}

// CompressPubkey encodes a public key to the 33-byte compressed format.
func (crypto *defaultCryptoBackend) CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return ethCrypto.CompressPubkey(pubkey)
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func (crypto *defaultCryptoBackend) UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	return ethCrypto.UnmarshalPubkey(pub)
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
