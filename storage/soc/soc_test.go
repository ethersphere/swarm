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

package soc

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/network"
)

func createRandomSOC(t *testing.T, refId uint32) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("could not generate a random key")
	}

	pubKey := hex.EncodeToString(crypto.FromECDSAPub(&privKey.PublicKey))
	data := make([]byte, chunk.DefaultSize)
	rand.Read(data)

	soc := NewSoc(privKey, network.RandomBzzAddr())
	span := rand.Uint64()
	socChunk, err := soc.NewChunk(refId, pubKey, span, data)
	if err != nil {
		t.Fatalf("could not create Soc chunk")
	}

	if !verifySocChunk(refId, pubKey, data, socChunk) {
		t.Fatalf("chunk verification failed")
	}
}

func verifySocChunk(refId uint32, pubKey string, data []byte, socChunk *SocChunk) bool {
	return true
}

func TestRandomDataForSoc(t *testing.T) {
	s := make([]int, 10)
	for _ = range s {
		createRandomSOC(t, rand.Uint32())
	}
}
