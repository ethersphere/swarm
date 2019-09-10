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
package crypto

import (
	"crypto/ecdsa"
	"fmt"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv6"
	"sync"
)

// Utils contains utility methods only for testing. Generates and stores asymmetric keys
type Utils interface {
	GenerateKey() (*ecdsa.PrivateKey, error)
	NewKeyPair() (string, error)
	GetPrivateKey(id string) (*ecdsa.PrivateKey, error)
}

type utils struct {
	privateKeys map[string]*ecdsa.PrivateKey // Private key storage
	keyMu       sync.RWMutex                 // Mutex associated with private key storage
}

func NewUtils() Utils {
	return &utils{
		privateKeys: make(map[string]*ecdsa.PrivateKey),
	}
}

// Generates a new private key right away. It is not stored for later retrieval.
func (utils *utils) GenerateKey() (*ecdsa.PrivateKey, error) {
	return ethCrypto.GenerateKey()
}

// NewKeyPair generates a new cryptographic identity for the client, and injects
// it into the known identities for message decryption. Returns ID of the new key pair.
func (utils *utils) NewKeyPair() (string, error) {
	key, err := ethCrypto.GenerateKey()
	if err != nil || !validatePrivateKey(key) {
		key, err = utils.GenerateKey() // retry once
	}
	if err != nil {
		return "", err
	}
	if !validatePrivateKey(key) {
		return "", fmt.Errorf("failed to generate valid key")
	}

	id, err := generateRandomKeyID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	utils.keyMu.Lock()
	defer utils.keyMu.Unlock()

	if utils.privateKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	utils.privateKeys[id] = key
	return id, nil
}

// GetPrivateKey return a PrivateKey previously generated in NewKeyPair by id
func (utils *utils) GetPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	whisperv6.NewSentMessage()
	utils.keyMu.RLock()
	defer utils.keyMu.RUnlock()
	key := utils.privateKeys[id]
	if key == nil {
		return nil, fmt.Errorf("invalid id")
	}
	return key, nil
}
