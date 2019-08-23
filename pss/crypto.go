// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package pss

import (
	"context"
	"crypto/ecdsa"

	whisper "github.com/ethereum/go-ethereum/whisper/whisperv6"
)

type CryptoBackend interface {
	GetSymKey(id string) ([]byte, error)
	GenerateSymKey() (string, error)
	AddSymKeyDirect(bytes []byte) (string, error)
	GetPrivateKey(id string) (*ecdsa.PrivateKey, error)
	NewKeyPair(ctx context.Context) (string, error)
}

type WhisperCryptoBackend struct {
	whisper *whisper.Whisper
	wapi    *whisper.PublicWhisperAPI
}

func NewCryptoBackend() CryptoBackend {
	w := whisper.New(&whisper.DefaultConfig)
	return &WhisperCryptoBackend{
		whisper: w,
		wapi:    whisper.NewPublicWhisperAPI(w),
	}
}

func (crypto *WhisperCryptoBackend) GetSymKey(id string) ([]byte, error) {
	return crypto.whisper.GetSymKey(id)
}

func (crypto *WhisperCryptoBackend) GenerateSymKey() (string, error) {
	return crypto.whisper.GenerateSymKey()
}

func (crypto *WhisperCryptoBackend) AddSymKeyDirect(bytes []byte) (string, error) {
	return crypto.whisper.AddSymKeyDirect(bytes)
}

func (crypto *WhisperCryptoBackend) GetPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	return crypto.whisper.GetPrivateKey(id)
}

func (crypto *WhisperCryptoBackend) NewKeyPair(ctx context.Context) (string, error) {
	return crypto.wapi.NewKeyPair(ctx)
}
