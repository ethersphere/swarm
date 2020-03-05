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
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/chunk"
)

func createRandomSOC(t *testing.T, refId []byte) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("could not generate a random key")
	}

	pubKey := hex.EncodeToString(crypto.FromECDSAPub(&privKey.PublicKey))
	data := make([]byte, chunk.DefaultSize)
	rand.Read(data)

	span := rand.Uint64()

	socAddr, err := NewSOCAddress(pubKey, refId)
	if err != nil {
		t.Fatalf("could not create Soc Address")
	}

	socData, err := NewSOCData(refId, span, data, privKey)
	if err != nil {
		t.Fatalf("could not create Soc Data")
	}

	verifySocChunk(t, refId, pubKey, span, data, privKey, socAddr, socData)
}

func verifySocChunk(t *testing.T, refId []byte, pubKey string, span uint64, data []byte, pkey *ecdsa.PrivateKey,
	socAddr []byte, socData []byte) {
	t.Helper()

	if len(socAddr) != 32 {
		t.Fatalf("soc address length is not 32 bytes")
	}

	if len(socData) !=  (DataHeaderSize + len(data)) {
		t.Fatalf("soc payload length is not 128 + actual data length")
	}

	extractedRefId := data[:ReferenceIdSize]
	if bytes.EqualFold(extractedRefId, refId) {
		t.Fatalf("invalid reference id")
	}


	// verify signatures and public key
	extractedSignature := make([]byte, SignatureSize)
	extractedSignature = data[ReferenceIdSize:ReferenceIdSize+SignatureSize]
	dataUsedForSignature, err := getIdAndDataHash(span, data, refId)
	if err != nil {
		t.Fatalf("could not get the data used for signature")
	}
	extractedPublicKey, err := crypto.SigToPub(dataUsedForSignature, extractedSignature)
	extractedPublicKeyBytes := crypto.FromECDSAPub(extractedPublicKey)
	pubkey := crypto.FromECDSAPub(&pkey.PublicKey)
	if bytes.Equal(pubkey,extractedPublicKeyBytes) {
		t.Fatalf("error in signature")
	}
}


func TestRandomDataForSoc(t *testing.T) {
	s := make([]int, 10)
	for _ = range s {
		refId := make([]byte, 32)
		_,_ = rand.Read(refId)
		createRandomSOC(t, refId)
	}
}
