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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/storage"
)

const (
	DefaultHash       = "SHA3" // http://golang.org/pkg/hash/#Hash
	PayloadHash       = "BMT"
	PublicKeySize     = 32  //
	ReferenceIdSize   = 32   //
	SOCDataHeaderSize = 128 // // ID  + Signature + padding + span  = 128
	SignatureSize     = 65
	PaddingSize       = 23
	SpanLength        = 8
)


func NewSOCAddress(ownerKey string, refId []byte) ([]byte, error) {
	hasher := storage.MakeHashFunc(DefaultHash)
	hasher().Reset()
	saddr := make([]byte, PublicKeySize+ReferenceIdSize)

	copy(saddr[:PublicKeySize], []byte(ownerKey))
	copy(saddr[PublicKeySize:], refId)
	_, err := hasher().Write(saddr)
	if err != nil {
		return nil, err
	}
	return hasher().Sum(nil), nil
}

func NewSOCData(refId []byte, span uint64, data []byte, pkey *ecdsa.PrivateKey) ([]byte, error) {
	if data == nil {
		return nil, errors.New("Invalid data length")
	}

	//(32)  +   (65)    +  (23)   + (8)   =  128
	socData := make([]byte, SOCDataHeaderSize+len(data))

	// 1 - Add refId
	copy(socData[:ReferenceIdSize], refId)


	// 2 - Add Signature
	idAndDataHash, err := getIdAndDataHash(span, data, refId)
	if err != nil {
		return nil, err
	}
	sig, err := crypto.Sign(idAndDataHash, pkey)
	if err != nil {
		return nil, err
	}
	copy(socData[ReferenceIdSize:], sig)

	// 3 - Add Padding
	padding := make([]byte, PaddingSize)
	copy(socData[ReferenceIdSize+SignatureSize:], padding)

	// 4 - Span
	binary.BigEndian.PutUint64(socData[ReferenceIdSize+SignatureSize+PaddingSize:], span)

	// 5 - data
	copy(socData[:SOCDataHeaderSize], data)

	return socData, nil
}

func getIdAndDataHash(span uint64, data []byte, refId []byte) ([]byte, error){
	//BMT (span + data)
	payloadHasher := storage.MakeHashFunc(PayloadHash)
	payloadHasher().Reset()
	payload := make([]byte, 8 + len(data))
	binary.BigEndian.PutUint64(payload[:8], span)
	copy(payload[8:],data)
	_, err := payloadHasher().Write(payload)
	if err != nil {
		return nil, err
	}
	dataHash := payloadHasher().Sum(nil)

	// Sha3 ( refId + BMT (data))
	hasher := storage.MakeHashFunc(DefaultHash)
	hasher().Reset()
	saddr := make([]byte, ReferenceIdSize+len(dataHash))
	copy(saddr[:ReferenceIdSize], refId)
	copy(saddr[ReferenceIdSize:], dataHash)
	_, err = hasher().Write(saddr)
	if err != nil {
		return nil, err
	}
	IdandDataHash := hasher().Sum(nil)

	return IdandDataHash, nil
}