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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethersphere/swarm/log"
)

// We need a singleton here?
var cryptoBackend defaultCryptoBackend
var createPadding = true

const (
	aesNonceLength = 12 // in bytes; for more info please see cipher.gcmStandardNonceSize & aesgcm.NonceSize()
	keyIDSize      = 32 // in bytes

	flagsLength             = 1
	payloadSizeFieldMaxSize = 4
	signatureLength         = 65      // in bytes
	padSizeLimit            = 256     // just an arbitrary number, could be changed without breaking the protocol
	SizeMask                = byte(3) // mask used to extract the size of payload size field from the flags
	signatureFlag           = byte(4)
	aesKeyLength            = 32 // in bytes

	defaultPaddingByteSize = 16
)

type WrapParams struct {
	Src    *ecdsa.PrivateKey
	Dst    *ecdsa.PublicKey
	KeySym []byte
}

type ReceivedMessage interface {
	ValidateAndParse() bool
	GetPayload() []byte
	GetSrc() *ecdsa.PublicKey
}

type Crypto interface {
	MessageCrypto
	CryptoBackend
	CryptoUtils
}

type MessageCrypto interface {
	WrapMessage(msgData []byte, params *WrapParams) (data []byte, err error)
	UnWrapSymmetric(encryptedData, key []byte) (ReceivedMessage, error)
	UnWrapAsymmetric(encryptedData []byte, key *ecdsa.PrivateKey) (ReceivedMessage, error)
}

type CryptoBackend interface {

	// Key store functions
	GetSymKey(id string) ([]byte, error)
	GenerateSymKey() (string, error)
	AddSymKeyDirect(bytes []byte) (string, error)

	// Key conversion
	FromECDSAPub(pub *ecdsa.PublicKey) []byte

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

var (
	errInvalidPubkey               = errors.New("invalid public key provided for asymmetric encryption")
	errInvalidSymkey               = errors.New("invalid key provided for symmetric encryption")
	errMissingSaltOrInvalidPayload = errors.New("missing salt or invalid payload in symmetric message")
	errSecureRandomData            = errors.New("failed to generate secure random data")
	errNeitherSymNorAsymKeysProv   = errors.New("unable to encrypt the message: neither symmetric nor asymmetric key provided")
)

type defaultCryptoBackend struct {
	privateKeys map[string]*ecdsa.PrivateKey // Private key storage
	symKeys     map[string][]byte            // Symmetric key storage
	keyMu       sync.RWMutex                 // Mutex associated with key storages
}

// receivedMessage represents a data packet to be received
// and successfully decrypted.
type receivedMessage struct {
	Payload   []byte // Parsed and validated message content
	Raw       []byte // Unparsed but decrypted message content
	Signature []byte // Signature from the sender
	Salt      []byte // Salt used for validation
	Padding   []byte // Padding applied

	Src *ecdsa.PublicKey // Source public key used for signing the message
	Dst *ecdsa.PublicKey // Destination public key used for encrypting the msessage asymmetrically

	crypto defaultCryptoBackend
}

func (msg receivedMessage) GetSrc() *ecdsa.PublicKey {
	return msg.Src
}

func (msg *receivedMessage) ValidateAndParse() bool {
	end := len(msg.Raw)
	if end < 1 {
		return false
	}
	if isMessageSigned(msg.Raw[0]) {
		end -= signatureLength
		if end <= 1 {
			return false
		}
		msg.Signature = msg.Raw[end : end+signatureLength]
		msg.Src = msg.sigToPubKey()
		if msg.Src == nil {
			return false
		}
	}

	beg := 1
	payloadSize := 0
	sizeOfPayloadSizeField := int(msg.Raw[0] & SizeMask) // number of bytes indicating the size of payload
	if sizeOfPayloadSizeField != 0 {
		payloadSize = int(bytesToUintLittleEndian(msg.Raw[beg : beg+sizeOfPayloadSizeField]))
		if payloadSize+1 > end {
			return false
		}
		beg += sizeOfPayloadSizeField
		msg.Payload = msg.Raw[beg : beg+payloadSize]
	}

	beg += payloadSize
	msg.Padding = msg.Raw[beg:end]
	return true
}

func (msg receivedMessage) GetPayload() []byte {
	return msg.Payload
}

// sigToPubKey returns the public key associated to the message's signature.
// should only be called id message has been signed
func (msg *receivedMessage) sigToPubKey() *ecdsa.PublicKey {
	defer func() { recover() }() // in case of invalid signature

	signedBytes := msg.Raw
	if isMessageSigned(msg.Raw[0]) {
		sz := len(msg.Raw) - signatureLength
		signedBytes = msg.Raw[:sz]
	}
	pub, err := msg.crypto.sigToPub(signedBytes, msg.Signature)
	if err != nil {
		log.Error("failed to recover public key from signature", "err", err)
		return nil
	}
	return pub
}

func New() Crypto {
	cryptoBackend = defaultCryptoBackend{
		symKeys:     make(map[string][]byte),
		privateKeys: make(map[string]*ecdsa.PrivateKey),
	}
	return &cryptoBackend
}

func NewCryptoUtils() CryptoUtils {
	if cryptoBackend.privateKeys == nil {
		New()
	}
	return &cryptoBackend
}

// == Message Crypto ==

func (crypto defaultCryptoBackend) WrapMessage(payload []byte, params *WrapParams) (data []byte, err error) {
	var padding []byte
	if createPadding {
		padding, err = crypto.generateSecureRandomData(defaultPaddingByteSize)
		if err != nil {
			return
		}
	} else {
		padding = make([]byte, 0)
	}
	// Message structure is flags+PayloadSize+Payload+padding+signature
	rawBytes := make([]byte, 1,
		flagsLength+payloadSizeFieldMaxSize+len(payload)+len(padding)+signatureLength+padSizeLimit)
	// set flags byte
	rawBytes[0] = 0 // set all the flags to zero
	// add payloadSizeField
	rawBytes = crypto.addPayloadSizeField(rawBytes, payload)
	// add payload
	rawBytes = append(rawBytes, payload...)
	// add padding
	rawBytes = append(rawBytes, padding...)
	// sign
	if params.Src != nil {
		if rawBytes, err = crypto.sign(rawBytes, params.Src); err != nil {
			return
		}
	}
	// encrypt
	if params.Dst != nil {
		rawBytes, err = crypto.encryptAsymmetric(rawBytes, params.Dst)
	} else if params.KeySym != nil {
		rawBytes, err = crypto.encryptSymmetric(rawBytes, params.KeySym)
	} else {
		err = errNeitherSymNorAsymKeysProv
	}
	if err != nil {
		return
	}

	data = rawBytes
	return
}

func (crypto defaultCryptoBackend) UnWrapSymmetric(encryptedData, key []byte) (ReceivedMessage, error) {
	decrypted, salt, err := crypto.decryptSymmetric(encryptedData, key)
	if err != nil {
		return nil, err
	}
	msg := newDecryptedMessage(decrypted, salt)
	return msg, err
}

func (crypto defaultCryptoBackend) UnWrapAsymmetric(encryptedData []byte, key *ecdsa.PrivateKey) (ReceivedMessage, error) {
	decrypted, err := crypto.decryptAsymmetric(encryptedData, key)
	switch err {
	case nil:
		message := newDecryptedMessage(decrypted, nil)
		return message, nil
	case ecies.ErrInvalidPublicKey: // addressed to somebody else
		return nil, err
	default:
		return nil, fmt.Errorf("unable to open envelope, decrypt failed: %v", err)
	}
}

func newDecryptedMessage(decrypted []byte, salt []byte) *receivedMessage {
	return &receivedMessage{
		Raw:  decrypted,
		Salt: salt,
	}
}

func (crypto defaultCryptoBackend) addPayloadSizeField(rawBytes rawMessage, payload []byte) rawMessage {
	fieldSize := getSizeOfPayloadSizeField(payload)
	field := make([]byte, 4)
	binary.LittleEndian.PutUint32(field, uint32(len(payload)))
	field = field[:fieldSize]
	rawBytes = append(rawBytes, field...)
	rawBytes[0] |= byte(fieldSize)
	return rawBytes
}

// appendPadding appends the padding specified in params.
// If no padding is provided in params, then random padding is generated.
func (crypto defaultCryptoBackend) appendPadding(rawBytes, payload []byte, src *ecdsa.PrivateKey) (rawMessage, error) {
	rawSize := flagsLength + getSizeOfPayloadSizeField(payload) + len(payload)
	if src != nil {
		rawSize += signatureLength
	}
	odd := rawSize % padSizeLimit
	paddingSize := padSizeLimit - odd
	pad := make([]byte, paddingSize)
	_, err := crand.Read(pad)
	if err != nil {
		return nil, err
	}

	if len(pad) != paddingSize || isAllZeroes(pad) {
		return nil, errors.New("failed to generate random padding of size " + strconv.Itoa(paddingSize))
	}
	rawBytes = append(rawBytes, pad...)
	return rawBytes, nil
}

// sign calculates and sets the cryptographic signature for the message,
// also setting the sign flag.
func (crypto defaultCryptoBackend) sign(rawBytes rawMessage, key *ecdsa.PrivateKey) (rawMessage, error) {
	if isMessageSigned(rawBytes[0]) {
		// this should not happen, but no reason to panic
		log.Error("failed to sign the message: already signed")
		return rawBytes, nil
	}
	rawBytes[0] |= signatureFlag // it is important to set this flag before signing

	hash := crypto.keccak256(rawBytes)
	signature, err := crypto.signHash(hash, key)

	if err != nil {
		rawBytes[0] &= (0xFF ^ signatureFlag) // clear the flag
		return nil, err
	}
	rawBytes = append(rawBytes, signature...)
	return rawBytes, nil
}

func isMessageSigned(flags byte) bool {
	return (flags & signatureFlag) != 0
}

// === Key store functions ===

func (crypto *defaultCryptoBackend) GetSymKey(id string) ([]byte, error) {
	crypto.keyMu.RLock()
	defer crypto.keyMu.RUnlock()
	if crypto.symKeys[id] != nil {
		return crypto.symKeys[id], nil
	}
	return nil, fmt.Errorf("non-existent key ID")
}

func (crypto *defaultCryptoBackend) GenerateSymKey() (string, error) {
	key, err := crypto.generateSecureRandomData(aesKeyLength)
	if err != nil {
		return "", err
	} else if !crypto.validateDataIntegrity(key, aesKeyLength) {
		return "", fmt.Errorf("error in GenerateSymKey: crypto/rand failed to generate random data")
	}

	id, err := crypto.generateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	crypto.keyMu.Lock()
	defer crypto.keyMu.Unlock()

	if crypto.symKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	crypto.symKeys[id] = key
	return id, nil
}

func (crypto *defaultCryptoBackend) AddSymKeyDirect(key []byte) (string, error) {
	if len(key) != aesKeyLength {
		return "", fmt.Errorf("wrong key size: %d", len(key))
	}

	id, err := crypto.generateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	crypto.keyMu.Lock()
	defer crypto.keyMu.Unlock()

	if crypto.symKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	crypto.symKeys[id] = key
	return id, nil
}

// === Key conversion ===

// FromECDSA exports a public key into a binary dump.
func (crypto *defaultCryptoBackend) FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	return ethCrypto.FromECDSAPub(pub)
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

// === Encrypt-Decrypt ===

// decryptSymmetric decrypts a message with a topic key, using AES-GCM-256.
// nonce size should be 12 bytes (see cipher.gcmStandardNonceSize).
func (crypto *defaultCryptoBackend) decryptSymmetric(rawBytes []byte, key []byte) (decrypted []byte, salt []byte, err error) {
	// symmetric messages are expected to contain the 12-byte nonce at the end of the payload
	if len(rawBytes) < aesNonceLength {
		return nil, nil, errMissingSaltOrInvalidPayload
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

// encryptAsymmetric encrypts a message with a public key.
func (crypto *defaultCryptoBackend) encryptAsymmetric(rawBytes []byte, key *ecdsa.PublicKey) ([]byte, error) {
	if !validatePublicKey(key) {
		return nil, errInvalidPubkey
	}
	encrypted, err := ecies.Encrypt(crand.Reader, crypto.importECDSAPublic(key), rawBytes, nil, nil)
	if err == nil {
		return encrypted, nil
	}
	return nil, err
}

func (crypto *defaultCryptoBackend) decryptAsymmetric(rawBytes []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	return ecies.ImportECDSA(key).Decrypt(rawBytes, nil, nil)
}

func (crypto *defaultCryptoBackend) encryptSymmetric(rawBytes []byte, key []byte) ([]byte, error) {
	if !crypto.validateDataIntegrity(key, aesKeyLength) {
		return nil, errInvalidSymkey
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

// signHash calculates an ECDSA signature.
func (crypto *defaultCryptoBackend) signHash(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	return ethCrypto.Sign(hash, prv)
}

// Obtain public key from the signed message and the signature
func (crypto *defaultCryptoBackend) sigToPub(signed, sig []byte) (*ecdsa.PublicKey, error) {
	hash := crypto.keccak256(signed)
	return ethCrypto.SigToPub(hash, sig)
}

// GenerateRandomID generates a random string, which is then returned to be used as a key id
func (crypto *defaultCryptoBackend) generateRandomID() (id string, err error) {
	buf, err := crypto.generateSecureRandomData(keyIDSize)
	if err != nil {
		return "", err
	}
	if !crypto.validateDataIntegrity(buf, keyIDSize) {
		return "", fmt.Errorf("error in generateRandomID: crypto/rand failed to generate random data")
	}
	id = common.Bytes2Hex(buf)
	return id, err
}

// keccak256 calculates and returns the keccak256 hash of the input data.
func (crypto *defaultCryptoBackend) keccak256(data ...[]byte) []byte {
	return ethCrypto.Keccak256(data...)
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
		return nil, errSecureRandomData
	}
	_, err = mrand.Read(y)
	if err != nil {
		return nil, err
	} else if !crypto.validateDataIntegrity(y, length) {
		return nil, errSecureRandomData
	}
	for i := 0; i < length; i++ {
		res[i] = x[i] ^ y[i]
	}
	if !crypto.validateDataIntegrity(res, length) {
		return nil, errSecureRandomData
	}
	return res, nil
}

func (crypto *defaultCryptoBackend) importECDSAPublic(key *ecdsa.PublicKey) *ecies.PublicKey {
	return ecies.ImportECDSAPublic(key)
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

// CryptoUtils

func (crypto *defaultCryptoBackend) GenerateKey() (*ecdsa.PrivateKey, error) {
	return ethCrypto.GenerateKey()
}

// NewKeyPair generates a new cryptographic identity for the client, and injects
// it into the known identities for message decryption. Returns ID of the new key pair.
func (crypto *defaultCryptoBackend) NewKeyPair(ctx context.Context) (string, error) {
	key, err := ethCrypto.GenerateKey()
	if err != nil || !validatePrivateKey(key) {
		key, err = crypto.GenerateKey() // retry once
	}
	if err != nil {
		return "", err
	}
	if !validatePrivateKey(key) {
		return "", fmt.Errorf("failed to generate valid key")
	}

	id, err := crypto.generateRandomID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %s", err)
	}

	crypto.keyMu.Lock()
	defer crypto.keyMu.Unlock()

	if crypto.privateKeys[id] != nil {
		return "", fmt.Errorf("failed to generate unique ID")
	}
	crypto.privateKeys[id] = key
	return id, nil
}

func (crypto *defaultCryptoBackend) GetPrivateKey(id string) (*ecdsa.PrivateKey, error) {
	crypto.keyMu.RLock()
	defer crypto.keyMu.RUnlock()
	key := crypto.privateKeys[id]
	if key == nil {
		return nil, fmt.Errorf("invalid id")
	}
	return key, nil
}

// Util functions

// validatePrivateKey checks the format of the given private key.
func validatePrivateKey(k *ecdsa.PrivateKey) bool {
	if k == nil || k.D == nil || k.D.Sign() == 0 {
		return false
	}
	return validatePublicKey(&k.PublicKey)
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

// bytesToUintLittleEndian converts the slice to 64-bit unsigned integer.
func bytesToUintLittleEndian(b []byte) (res uint64) {
	mul := uint64(1)
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) * mul
		mul *= 256
	}
	return res
}

func isAllZeroes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// getSizeOfPayloadSizeField returns the number of bytes necessary to encode the size of payload
func getSizeOfPayloadSizeField(payload []byte) int {
	s := 1
	for i := len(payload); i >= 256; i /= 256 {
		s++
	}
	return s
}

type rawMessage []byte
