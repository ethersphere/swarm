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

// Config params to wrap and encrypt a message.
// For asymmetric encryption Receiver is needed.
// For symmetric, SymmetricKey is needed. Sender is not mandatory but used to sign the message in both schemes.
type WrapParams struct {
	Sender       *ecdsa.PrivateKey // Private key of sender used for signature
	Receiver     *ecdsa.PublicKey  // Public key of receiver for encryption
	SymmetricKey []byte            // Symmetric key for encryption
}

// Config params to unwrap and decrypt a message.
// For asymmetric encryption Receiver is needed.
// For symmetric, SymmetricKey is needed. Sender is not mandatory but used to sign the message in both schemes.
type UnwrapParams struct {
	Sender       *ecdsa.PublicKey  // Private key of sender used for signature validation
	Receiver     *ecdsa.PrivateKey // Public key of receiver for decryption
	SymmetricKey []byte            // Symmetric key for decryption
}

// Contains a successfully decrypted message prior to parsing and validating
type ReceivedMessage interface {
	GetPayload() ([]byte, error)
	GetSender() *ecdsa.PublicKey
}

// Crypto contains methods from Message and KeyStore
type Crypto interface {
	Message
	KeyStore
}

// Message contains methods for wrapping(encrypting) and unwrapping(decrypting) messages
type Message interface {
	Wrap(plaintext []byte, params *WrapParams) (data []byte, err error)
	UnWrap(ciphertext []byte, unwrapParams *UnwrapParams) (ReceivedMessage, error)
}

// KeyStore contains key manipulation methods
type KeyStore interface {

	// Asymmetric key management
	GetSymKey(id string) ([]byte, error)
	GenerateSymKey() (string, error)
	AddSymKey(bytes []byte) (string, error)

	// Key serialization
	SerializePublicKey(pub *ecdsa.PublicKey) []byte
	UnmarshalPublicKey(pub []byte) (*ecdsa.PublicKey, error)
	CompressPublicKey(pub *ecdsa.PublicKey) []byte
}

var (
	errInvalidPubkey               = errors.New("invalid public key provided for asymmetric encryption")
	errInvalidSymkey               = errors.New("invalid key provided for symmetric encryption")
	errMissingSaltOrInvalidPayload = errors.New("missing salt or invalid payload in symmetric message")
	errSecureRandomData            = errors.New("failed to generate secure random data")
	errNoKey                       = errors.New("unable to encrypt the message: neither symmetric nor asymmetric key provided")

	// Validation and Parse errors
	errEmptyMessage       = errors.New("empty message")
	errEmptySignature     = errors.New("empty expected signature")
	errIncorrectSignature = errors.New("incorrect signature")
	errIncorrectSize      = errors.New("incorrect payload size")
)

type defaultCryptoBackend struct {
	symKeys map[string][]byte // Symmetric key storage
	keyMu   sync.RWMutex      // Mutex associated with key storage
}

// receivedMessage represents a data packet to be received
// and successfully decrypted.
type receivedMessage struct {
	Payload   []byte // Parsed and validated message content
	Raw       []byte // Unparsed but decrypted message content
	Signature []byte // Signature from the sender
	Salt      []byte // Protect against plaintext attacks
	Padding   []byte // Padding applied

	Sender *ecdsa.PublicKey // Source public key used for signing the message

	validateOnce  sync.Once
	validateError error

	crypto *defaultCryptoBackend
}

// Returns the sender public key of the message
func (msg *receivedMessage) GetSender() *ecdsa.PublicKey {
	msg.validateOnce.Do(
		func() {
			msg.validateError = msg.validateAndParse()
		})
	return msg.Sender
}

// GetPayload validate and parse the payload of the message.
func (msg *receivedMessage) GetPayload() ([]byte, error) {
	msg.validateOnce.Do(
		func() {
			msg.validateError = msg.validateAndParse()
		})
	return msg.Payload, msg.validateError
}

// validateAndParse checks that the format and the signature are correct. It also set Payload as the parsed message
func (msg *receivedMessage) validateAndParse() error {
	end := len(msg.Raw)
	if end == 0 {
		return errEmptyMessage
	}
	if isMessageSigned(msg.Raw[0]) {
		end -= signatureLength
		if end <= 1 {
			return errEmptySignature
		}
		msg.Signature = msg.Raw[end : end+signatureLength]
		sz := len(msg.Raw) - signatureLength
		pub, err := msg.crypto.sigToPub(msg.Raw[:sz], msg.Signature)
		if err != nil {
			log.Error("failed to recover public key from signature", "err", err)
			return errIncorrectSignature
		} else {
			msg.Sender = pub
		}
	}

	beg := 1
	payloadSize := 0
	sizeOfPayloadSizeField := int(msg.Raw[0] & SizeMask) // number of bytes indicating the size of payload
	if sizeOfPayloadSizeField != 0 {
		log.Warn("Size of payload field", "size", sizeOfPayloadSizeField)
		payloadSize = int(bytesToUintLittleEndian(msg.Raw[beg : beg+sizeOfPayloadSizeField]))
		if payloadSize+1 > end {
			return errIncorrectSize
		}
		beg += sizeOfPayloadSizeField
		msg.Payload = msg.Raw[beg : beg+payloadSize]
	}

	beg += payloadSize
	msg.Padding = msg.Raw[beg:end]
	return nil
}

func newReceivedMessage(decrypted []byte, salt []byte, crypto *defaultCryptoBackend) *receivedMessage {
	return &receivedMessage{
		Raw:          decrypted,
		Salt:         salt,
		crypto:       crypto,
		validateOnce: sync.Once{},
	}
}

// Return the default implementation of Crypto
func New() Crypto {
	return newDefaultCryptoBackend()

}

func newDefaultCryptoBackend() *defaultCryptoBackend {
	return &defaultCryptoBackend{
		symKeys: make(map[string][]byte),
	}
}

// == Message Crypto ==

// Wrap creates a message adding signature, padding and other control fields and then it is encrypted using params
func (crypto *defaultCryptoBackend) Wrap(plaintext []byte, params *WrapParams) (data []byte, err error) {
	var padding []byte
	if createPadding {
		padding, err = generateSecureRandomData(defaultPaddingByteSize)
		if err != nil {
			return
		}
	} else {
		padding = make([]byte, 0)
	}
	// Message structure is flags+PayloadSize+Payload+padding+signature
	rawBytes := make([]byte, 1,
		flagsLength+payloadSizeFieldMaxSize+len(plaintext)+len(padding)+signatureLength+padSizeLimit)
	// set flags byte
	rawBytes[0] = 0 // set all the flags to zero
	// add payloadSizeField
	rawBytes = crypto.addPayloadSizeField(rawBytes, plaintext)
	// add payload
	rawBytes = append(rawBytes, plaintext...)
	// add padding
	rawBytes = append(rawBytes, padding...)
	// sign
	if params.Sender != nil {
		if rawBytes, err = crypto.sign(rawBytes, params.Sender); err != nil {
			return
		}
	}
	// encrypt
	if params.Receiver != nil {
		rawBytes, err = crypto.encryptAsymmetric(rawBytes, params.Receiver)
	} else if params.SymmetricKey != nil {
		rawBytes, err = crypto.encryptSymmetric(rawBytes, params.SymmetricKey)
	} else {
		err = errNoKey
	}
	if err != nil {
		return
	}

	data = rawBytes
	return
}

// Unwrap decrypts and compose a received message ready to be parsed and validated.
// It selects symmetric/asymmetric crypto depending on unwrapParams
func (crypto *defaultCryptoBackend) UnWrap(ciphertext []byte, unwrapParams *UnwrapParams) (ReceivedMessage, error) {
	if unwrapParams.SymmetricKey != nil {
		return crypto.unWrapSymmetric(ciphertext, unwrapParams.SymmetricKey)
	} else if unwrapParams.Receiver != nil {
		return crypto.unWrapAsymmetric(ciphertext, unwrapParams.Receiver)
	} else {
		return nil, errNoKey
	}
}

func (crypto *defaultCryptoBackend) unWrapSymmetric(ciphertext, key []byte) (ReceivedMessage, error) {
	decrypted, salt, err := crypto.decryptSymmetric(ciphertext, key)
	if err != nil {
		return nil, err
	}
	msg := newReceivedMessage(decrypted, salt, crypto)
	return msg, err
}

func (crypto *defaultCryptoBackend) unWrapAsymmetric(ciphertext []byte, key *ecdsa.PrivateKey) (ReceivedMessage, error) {
	plaintext, err := crypto.decryptAsymmetric(ciphertext, key)
	switch err {
	case nil:
		message := newReceivedMessage(plaintext, nil, crypto)
		return message, nil
	case ecies.ErrInvalidPublicKey: // addressed to somebody else
		return nil, err
	default:
		return nil, fmt.Errorf("unable to open envelope, decrypt failed: %v", err)
	}
}

func (crypto *defaultCryptoBackend) addPayloadSizeField(rawBytes []byte, payload []byte) []byte {
	fieldSize := getSizeOfPayloadSizeField(payload)
	field := make([]byte, 4)
	binary.LittleEndian.PutUint32(field, uint32(len(payload)))
	field = field[:fieldSize]
	rawBytes = append(rawBytes, field...)
	rawBytes[0] |= byte(fieldSize)
	return rawBytes
}

// pad appends the padding specified in params.
func (crypto *defaultCryptoBackend) pad(rawBytes, payload []byte, signed bool) ([]byte, error) {
	rawSize := flagsLength + getSizeOfPayloadSizeField(payload) + len(payload)
	if signed {
		rawSize += signatureLength
	}
	odd := rawSize % padSizeLimit
	paddingSize := padSizeLimit - odd
	pad := make([]byte, paddingSize)
	_, err := crand.Read(pad)
	if err != nil {
		return nil, err
	}

	if len(pad) != paddingSize || containsOnlyZeros(pad) {
		return nil, errors.New("failed to generate random padding of size " + strconv.Itoa(paddingSize))
	}
	rawBytes = append(rawBytes, pad...)
	return rawBytes, nil
}

// sign calculates and sets the cryptographic signature for the message,
// also setting the sign flag.
func (crypto *defaultCryptoBackend) sign(rawBytes []byte, key *ecdsa.PrivateKey) ([]byte, error) {
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

// GetSymKey retrieves symmetric key by id from the store
func (crypto *defaultCryptoBackend) GetSymKey(id string) ([]byte, error) {
	crypto.keyMu.RLock()
	defer crypto.keyMu.RUnlock()
	if crypto.symKeys[id] != nil {
		return crypto.symKeys[id], nil
	}
	return nil, fmt.Errorf("non-existent key ID")
}

// GenerateSymKey creates a new symmetric, stores it and return its id
func (crypto *defaultCryptoBackend) GenerateSymKey() (string, error) {
	key, err := generateSecureRandomData(aesKeyLength)
	if err != nil {
		return "", err
	} else if !validateDataIntegrity(key, aesKeyLength) {
		return "", fmt.Errorf("error in GenerateSymKey: crypto/rand failed to generate random data")
	}

	id, err := generateRandomID()
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

// Add a symmetric key ti the store generating an id and returning it
func (crypto *defaultCryptoBackend) AddSymKey(key []byte) (string, error) {
	if len(key) != aesKeyLength {
		return "", fmt.Errorf("wrong key size: %d", len(key))
	}

	id, err := generateRandomID()
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
func (crypto *defaultCryptoBackend) SerializePublicKey(pub *ecdsa.PublicKey) []byte {
	return ethCrypto.FromECDSAPub(pub)
}

// === Key serialization ===

// UnmarshalPublicKey converts bytes to a secp256k1 public key.
func (crypto *defaultCryptoBackend) UnmarshalPublicKey(pub []byte) (*ecdsa.PublicKey, error) {
	return ethCrypto.UnmarshalPubkey(pub)
}

// CompressPublicKey encodes a public key to the 33-byte compressed format.
func (crypto *defaultCryptoBackend) CompressPublicKey(pubkey *ecdsa.PublicKey) []byte {
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
	if !validateDataIntegrity(key, aesKeyLength) {
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
	salt, err := generateSecureRandomData(aesNonceLength) // never use more than 2^32 random nonces with a given key
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
	defer func() { recover() }() // in case of invalid signature
	hash := crypto.keccak256(signed)
	return ethCrypto.SigToPub(hash, sig)
}

func (crypto *defaultCryptoBackend) importECDSAPublic(key *ecdsa.PublicKey) *ecies.PublicKey {
	return ecies.ImportECDSAPublic(key)
}

// GenerateRandomID generates a random string, which is then returned to be used as a key id
func generateRandomID() (id string, err error) {
	buf, err := generateSecureRandomData(keyIDSize)
	if err != nil {
		return "", err
	}
	if !validateDataIntegrity(buf, keyIDSize) {
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
func generateSecureRandomData(length int) ([]byte, error) {
	x := make([]byte, length)
	y := make([]byte, length)
	res := make([]byte, length)

	_, err := crand.Read(x)
	if err != nil {
		return nil, err
	} else if !validateDataIntegrity(x, length) {
		return nil, errSecureRandomData
	}
	_, err = mrand.Read(y)
	if err != nil {
		return nil, err
	} else if !validateDataIntegrity(y, length) {
		return nil, errSecureRandomData
	}
	for i := 0; i < length; i++ {
		res[i] = x[i] ^ y[i]
	}
	if !validateDataIntegrity(res, length) {
		return nil, errSecureRandomData
	}
	return res, nil
}

// validateDataIntegrity returns false if the data have the wrong size or contains all zeros,
// which is the simplest and the most common bug.
func validateDataIntegrity(k []byte, expectedSize int) bool {
	if len(k) != expectedSize {
		return false
	}
	if expectedSize > 3 && containsOnlyZeros(k) {
		return false
	}
	return true
}

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

// getSizeOfPayloadSizeField returns the number of bytes necessary to encode the size of payload
func getSizeOfPayloadSizeField(payload []byte) int {
	s := 1
	for i := len(payload); i >= 256; i /= 256 {
		s++
	}
	return s
}
