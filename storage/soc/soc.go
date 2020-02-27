// Copyright 2018 The go-ethereum Authors
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

package soc

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/storage"
)

const (
	DefaultHash       = "SHA3" // http://golang.org/pkg/hash/#Hash
	PayloadHash       = "BMT"
	PublicKeySize     = 32  //
	ReferenceIdSize   = 32  //
	SOCDataHeaderSize = 128 // // ID  + Signature + padding + span  = 128
	SignatureSize     = 65
	PaddingSize       = 23
)

type Soc struct {
	pkey        *ecdsa.PrivateKey
	overlayAddr *network.BzzAddr
}

type SocChunk struct {
	address []byte
	data    []byte
}

func NewSoc(key *ecdsa.PrivateKey, bzzAddr *network.BzzAddr) *Soc {
	return &Soc{
		pkey:        key,
		overlayAddr: bzzAddr,
	}
}

func (s *Soc) NewChunk(refId uint32, ownerPubKey string, span uint64, payload []byte) (*SocChunk, error) {
	add, err := s.setChunkAddress(ownerPubKey, refId)
	if err != nil {
		return nil, err
	}

	payload, err = s.setChunkData(refId, span, payload)
	if err != nil {
		return nil, err
	}

	return &SocChunk{
		address: add,
		data:    payload,
	}, nil
}

func (s *Soc) setChunkAddress(ownerKey string, refId uint32) ([]byte, error) {
	hasher := storage.MakeHashFunc(DefaultHash)
	hasher().Reset()
	saddr := make([]byte, PublicKeySize+ReferenceIdSize)

	copy(saddr[:PublicKeySize], []byte(ownerKey))
	binary.BigEndian.PutUint32(saddr[:PublicKeySize], refId)
	_, err := hasher().Write(saddr)
	if err != nil {
		return nil, err
	}
	return hasher().Sum(nil), nil
}

func (s *Soc) setChunkData(refId uint32, span uint64, data []byte) ([]byte, error) {
	if data == nil {
		return nil, errors.New("Invalid data length")
	}

	//(32)  +   (65)    +  (23)   + (8)   =  128
	socData := make([]byte, SOCDataHeaderSize+len(data))

	// 1 - Add refId
	binary.BigEndian.PutUint32(socData[:ReferenceIdSize], refId)

	//BMT (data)
	payloadHasher := storage.MakeHashFunc(PayloadHash)
	payloadHasher().Reset()
	_, err := payloadHasher().Write(data)
	if err != nil {
		return nil, err
	}
	dataHash := payloadHasher().Sum(nil)

	// Sha3 ( refId + BMT (data))
	hasher := storage.MakeHashFunc(DefaultHash)
	hasher().Reset()
	saddr := make([]byte, ReferenceIdSize+len(dataHash))
	binary.BigEndian.PutUint32(saddr[:ReferenceIdSize], refId)
	copy(saddr[ReferenceIdSize:], dataHash)
	_, err = hasher().Write(saddr)
	if err != nil {
		return nil, err
	}
	IdandDataHash := hasher().Sum(nil)

	sig, err := crypto.Sign(IdandDataHash, s.pkey)
	if err != nil {
		return nil, err
	}

	// 2 - Add Signature
	copy(socData[ReferenceIdSize:], sig)

	// 3 - Add Padding
	padding := make([]byte, PaddingSize)
	rand.Read(padding)
	copy(socData[ReferenceIdSize+SignatureSize:], padding)

	// 4 - Span
	binary.BigEndian.PutUint64(socData[ReferenceIdSize+SignatureSize+PaddingSize:], span)

	// 5 - data
	copy(socData[:SOCDataHeaderSize], data)

	return saddr, nil
}
