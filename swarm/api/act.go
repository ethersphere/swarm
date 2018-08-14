package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/swarm/log"
	"github.com/ethereum/go-ethereum/swarm/sctx"
	"github.com/ethereum/go-ethereum/swarm/storage"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	ErrDecrypt                = errors.New("cant decrypt - forbidden")
	ErrUnknownAccessType      = errors.New("unknown access type (or not implemented)")
	ErrDecryptDomainForbidden = errors.New("decryption request domain forbidden - can only decrypt on localhost")
	AllowedDecryptDomains     = []string{
		"localhost",
		"127.0.0.1",
	}
)

const EMPTY_CREDENTIALS = ""

func (a *API) doDecrypt(ctx context.Context, credentials string, pk *ecdsa.PrivateKey) DecryptFunc {
	return func(m *ManifestEntry) error {
		if m.Access == nil {
			return nil
		}

		allowed := false
		requestDomain := sctx.GetHost(ctx)
		for _, v := range AllowedDecryptDomains {
			if strings.Contains(requestDomain, v) {
				allowed = true
			}
		}

		if !allowed {
			return ErrDecryptDomainForbidden
		}

		switch m.Access.Type {
		case "pass":
			if credentials != "" {
				// decrypt
				key, err := NewSessionKeyPassword(credentials, m.Access)
				if err != nil {
					return err
				}

				ref, err := hex.DecodeString(m.Hash)
				if err != nil {
					return err
				}

				enc := NewRefEncryption(len(ref) - 8)
				decodedRef, err := enc.Decrypt(ref, key)
				if err != nil {
					// Return ErrDecrypt to be able to detect
					// invalid decryption in hinger levels of code.
					return ErrDecrypt
				}

				m.Hash = hex.EncodeToString(decodedRef)
				m.Access = nil
				return nil
			}
			return ErrDecrypt
		case "pk":
			publisherBytes, err := hex.DecodeString(m.Access.Publisher)
			if err != nil {
				return ErrDecrypt
			}
			publisher, err := crypto.DecompressPubkey(publisherBytes)
			if err != nil {
				return ErrDecrypt
			}
			key, err := a.NodeSessionKey(pk, publisher, m.Access.Salt)
			if err != nil {
				return ErrDecrypt
			}
			ref, err := hex.DecodeString(m.Hash)
			if err != nil {
				return err
			}

			enc := NewRefEncryption(len(ref) - 8)
			decodedRef, err := enc.Decrypt(ref, key)
			if err != nil {
				// Return ErrDecrypt to be able to detect
				// invalid decryption in hinger levels of code.
				return ErrDecrypt
			}

			m.Hash = hex.EncodeToString(decodedRef)
			m.Access = nil
			return nil
		case "act":
			publisherBytes, err := hex.DecodeString(m.Access.Publisher)
			if err != nil {
				return ErrDecrypt
			}
			publisher, err := crypto.DecompressPubkey(publisherBytes)
			if err != nil {
				return ErrDecrypt
			}

			sessionKey, err := a.NodeSessionKey(pk, publisher, m.Access.Salt)
			if err != nil {
				return ErrDecrypt
			}

			hasher := sha3.NewKeccak256()
			hasher.Write(append(sessionKey, 0))
			lookupKey := hasher.Sum(nil)

			hasher.Reset()

			hasher.Write(append(sessionKey, 1))
			accessKeyDecryptionKey := hasher.Sum(nil)

			lk := hex.EncodeToString(lookupKey)
			log.Error("lookup", "lk", lk, "act", m.Access.Act)
			list, err := a.GetManifestList(ctx, NOOPDecrypt, storage.Address(common.Hex2Bytes(m.Access.Act)), lk)

			found := ""
			for _, v := range list.Entries {
				if v.Path == lk {
					found = v.Hash
				}
			}

			if found == "" {
				return ErrDecrypt
			}

			v, err := hex.DecodeString(found)
			if err != nil {
				return err
			}
			enc := NewRefEncryption(len(v) - 8)
			decodedRef, err := enc.Decrypt(v, accessKeyDecryptionKey)
			if err != nil {
				// Return ErrDecrypt to be able to detect
				// invalid decryption in hinger levels of code.
				return ErrDecrypt
			}

			ref, err := hex.DecodeString(m.Hash)
			if err != nil {
				return err
			}

			enc = NewRefEncryption(len(ref) - 8)
			decodedMainRef, err := enc.Decrypt(ref, decodedRef)
			if err != nil {
				// Return ErrDecrypt to be able to detect
				// invalid decryption in hinger levels of code.
				return ErrDecrypt
			}
			m.Hash = hex.EncodeToString(decodedMainRef)
			m.Access = nil
			return nil
		}
		return ErrUnknownAccessType
	}
}

func GenerateAccessControlManifest(ctx *cli.Context, ref string, accessKey []byte, ae *AccessEntry) (*Manifest, error) {
	refBytes, err := hex.DecodeString(ref)
	if err != nil {
		return nil, err
	}
	// encrypt ref with accessKey
	enc := NewRefEncryption(len(refBytes))
	encrypted, err := enc.Encrypt(refBytes, accessKey)
	if err != nil {
		return nil, err
	}

	m := &Manifest{
		Entries: []ManifestEntry{
			{
				Hash:        hex.EncodeToString(encrypted),
				ContentType: ManifestType,
				ModTime:     time.Now(),
				Access:      ae,
			},
		},
	}

	return m, nil
}

func DoPKNew(ctx *cli.Context, privateKey *ecdsa.PrivateKey, granteePublicKey string, salt []byte) (sessionKey []byte, ae *AccessEntry, err error) {
	if granteePublicKey == "" {
		return nil, nil, errors.New("need a grantee Public Key")
	}
	b, err := hex.DecodeString(granteePublicKey)
	if err != nil {
		log.Error("error decoding grantee public key", "err", err)
		return nil, nil, err
	}

	granteePub, err := crypto.DecompressPubkey(b)
	if err != nil {
		log.Error("error decompressing grantee public key", "err", err)
		return nil, nil, err
	}

	sessionKey, err = NewSessionKeyPK(privateKey, granteePub, salt)
	if err != nil {
		log.Error("error getting session key", "err", err)
		return nil, nil, err
	}

	ae, err = NewAccessEntryPK(hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey)), salt)
	if err != nil {
		log.Error("error generating access entry", "err", err)
		return nil, nil, err
	}

	return sessionKey, ae, nil
}

func DoACTNew(ctx *cli.Context, privateKey *ecdsa.PrivateKey, salt []byte, grantees []string) (accessKey []byte, ae *AccessEntry, actManifest *Manifest, err error) {
	if len(grantees) == 0 {
		utils.Fatalf("did not get any grantee public keys")
	}

	publisherPub := hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey))
	grantees = append(grantees, publisherPub)

	accessKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	if _, err := io.ReadFull(rand.Reader, accessKey); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}

	lookupPathEncryptedAccessKeyMap := make(map[string]string)
	i := 0
	for _, v := range grantees {
		i++
		if v == "" {
			return nil, nil, nil, errors.New("need a grantee Public Key")
		}
		b, err := hex.DecodeString(v)
		if err != nil {
			log.Error("error decoding grantee public key", "err", err)
			return nil, nil, nil, err
		}

		granteePub, err := crypto.DecompressPubkey(b)
		if err != nil {
			log.Error("error decompressing grantee public key", "err", err)
			return nil, nil, nil, err
		}
		sessionKey, err := NewSessionKeyPK(privateKey, granteePub, salt)

		hasher := sha3.NewKeccak256()
		hasher.Write(append(sessionKey, 0))
		lookupKey := hasher.Sum(nil)

		hasher.Reset()
		hasher.Write(append(sessionKey, 1))

		accessKeyEncryptionKey := hasher.Sum(nil)

		enc := NewRefEncryption(len(accessKey))
		encryptedAccessKey, err := enc.Encrypt(accessKey, accessKeyEncryptionKey)

		lookupPathEncryptedAccessKeyMap[hex.EncodeToString(lookupKey)] = hex.EncodeToString(encryptedAccessKey)
	}

	m := &Manifest{
		Entries: []ManifestEntry{},
	}

	for k, v := range lookupPathEncryptedAccessKeyMap {
		m.Entries = append(m.Entries, ManifestEntry{
			Path:        k,
			Hash:        v,
			ContentType: "text/plain",
		})
	}

	ae, err = NewAccessEntryACT(hex.EncodeToString(crypto.CompressPubkey(&privateKey.PublicKey)), salt, "")
	if err != nil {
		return nil, nil, nil, err
	}

	return accessKey, ae, m, nil
}

func DoPasswordNew(ctx *cli.Context, password string, salt []byte) (sessionKey []byte, ae *AccessEntry, err error) {
	ae, err = NewAccessEntryPassword(salt, DefaultKdfParams)
	if err != nil {
		return nil, nil, err
	}

	sessionKey, err = NewSessionKeyPassword(password, ae)
	if err != nil {
		return nil, nil, err
	}
	return sessionKey, ae, nil
}
