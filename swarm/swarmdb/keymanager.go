// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// KeyManager is used to abstract the Ethereum wallet keystore (a local directory holding raw JSON files) for SWARMDB to:
// (a) sign and verify messages [e.g. in SWARMDB TCP/IP client-server communications]  with SignMessage and VerifyMessage
// (b) encrypt and decrypt chunks

package swarmdb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	"golang.org/x/crypto/nacl/box"
	// "os"
)

type KeyManager struct {
	config   *SWARMDBConfig
	keystore *keystore.KeyStore
}

func SignHash(unencrypted []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(unencrypted), unencrypted)
	return crypto.Keccak256([]byte(msg))
}

// KeyManager requires a swarmdb.conf loaded into SWARMDBConfig.  This config specifies a directory specified in the ChunkDBPath
//  i.e.  "chunkDBPath": "/swarmdb/data/keystore"
// Inside the config file there are wallet passphrases (in **plaintext** so the config file MUST be SECURED)
// which are used to unlock any users wallet and get at their publicKey and secretKey
func NewKeyManager(c *SWARMDBConfig) (keymgr KeyManager, err error) {

	if c != nil {
		// for all users specified in the config file, set up their { sk, pk }  in the config
		keymgr.config = c
		//keymgr.keystore = keystore.NewKeyStore(c.ChunkDBPath, keystore.StandardScryptN, keystore.StandardScryptP)
		keymgr.keystore = keystore.NewKeyStore(c.KeystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
		wallets := keymgr.keystore.Wallets()
		for _, u := range c.Users {
			address := common.HexToAddress(u.Address)
			for _, w := range wallets {
				accounts := w.Accounts()
				for _, a := range accounts {
					if bytes.Compare(a.Address.Bytes(), address.Bytes()) == 0 {
						err := keymgr.keystore.Unlock(a, u.Passphrase)
						// TODO: what if people supply a secretkey instead of a passphrase?
						_, k, err := keymgr.keystore.WgetDecryptedKey(a, u.Passphrase)
						if err != nil {
							return keymgr, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:NewKeyManager] WgetDecryptedKey - Error Decrypting Account: %s", err.Error()), ErrorCode: 452, ErrorMessage: "Error Decrypting Account"}
						} else {
							u.sk = crypto.FromECDSA(k.PrivateKey)
							u.pk = crypto.FromECDSAPub(&k.PrivateKey.PublicKey)

							copy(u.publicK[0:], u.pk[0:])
							copy(u.secretK[0:], u.sk[0:])
						}
					}
				}
			}
		}
	}

	return keymgr, nil // fmt.Errorf("No keystore file found", ownerAddress)
}

// server will call this when the configuration file does not exist
func NewKeyManagerWithoutConfig(filename string, passphrase string) (keymgr KeyManager, err error) {
	// keystore created here
	keymgr.keystore = keystore.NewKeyStore(SWARMDBCONF_KEYSTORE_PATH, keystore.StandardScryptN, keystore.StandardScryptP)

	// create new account with passphrase using keystore
	account, err := keymgr.keystore.NewAccount(passphrase)
	if err != nil {
		return keymgr, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:NewKeyManagerWithoutConfig] NewAccount %s", err.Error()), ErrorCode: 453, ErrorMessage: "Error creating new account"}
	}
	fmt.Printf("Account: %v\n", account)

	// get address of the new account
	// address := common.HexToAddress(u.Address)
	address := fmt.Sprintf("%x", account.Address.Bytes())
	fmt.Printf("Address: %v\n", address)

	// unlocking the account using the passphrase
	err = keymgr.keystore.Unlock(account, passphrase)
	if err != nil {
		return keymgr, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:NewKeyManagerWithoutConfig] Unlock %s", err.Error()), ErrorCode: 454, ErrorMessage: "Error Unlocking Account"}
	}
	fmt.Printf("Unlocked account with passphras %s\n", passphrase)

	// get the Key of the new account account from the keystore
	_, k, err := keymgr.keystore.WgetDecryptedKey(account, passphrase)
	if err != nil {
		return keymgr, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:NewKeyManagerWithoutConfig] WgetDecryptedKey %s", err.Error()), ErrorCode: 452, ErrorMessage: "Error Decrypting Account"}
	}
	fmt.Printf("Key: %v\n", k)

	// make a config using the { privatekey, address, passphrase }
	privateKey := hex.EncodeToString(crypto.FromECDSA(k.PrivateKey))

	config := GenerateSampleSWARMDBConfig(privateKey, address, passphrase)
	fmt.Printf("SWARMDBConfig: %v\n", config)

	// save it!
	err = SaveSWARMDBConfig(config, filename)
	if err != nil {
		return keymgr, sdbc.GenerateSWARMDBError(err, fmt.Sprintf("[keymanager:NewKeyManagerWithoutConfig] SaveSWARMDBConfig %s", err.Error()))
	}
	fmt.Printf("Saved config in %s:\n", filename)
	return keymgr, nil
}

// client libraries call this to sign Messages (hashed to 32 bytes) with the config's secret
func (self *KeyManager) SignMessage(msg_hash []byte) (sig []byte, err error) {
	secretKey, err := crypto.HexToECDSA(self.config.PrivateKey)
	if err != nil {
		return sig, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:SignMessage] HexToECDSA  %s", err.Error()), ErrorCode: 455, ErrorMessage: "Keymanager Unable to Sign Message"}
	} else {
		sig, err2 := crypto.Sign(msg_hash, secretKey)
		if err2 != nil {
			return sig, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:SignMessage] Sign %s", err.Error()), ErrorCode: 455, ErrorMessage: "Keymanager Unable to Sign Message"}
		}
		return sig, nil
	}
}

// Given a 32 byte hash of a message and a signature [signed with SignMessage above]
// returns the specific SWARMDBUser in the keystore
// If no user
// This is used in SWARMDB TCP server + HTTP use client response to a challenge to determine which account the user is
func (self *KeyManager) VerifyMessage(msg_hash []byte, sig []byte) (u *SWARMDBUser, err error) {
	// signatures are 65 byte RSV form - but RSV has the last 1-byte V [web3 (1b/1c) vs go client (00/01)]
	if len(sig) >= 65 {
		if sig[64] > 4 {
			sig[64] -= 27
		}
	} else {
		return u, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:VerifyMessage] Invalid signature length %d [%x]", len(sig), sig), ErrorCode: 419, ErrorMessage: "Invalid Signature Length: Must be 65 characters"}
	}
	pubKey, err := crypto.SigToPub(msg_hash, sig)
	if err != nil {
		return u, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:VerifyMessage] Invalid signature - Cannot get public key"), ErrorCode: 420, ErrorMessage: "Invalid Signature: Unable to Retrieve Public Key"}
	} else {
		address := crypto.PubkeyToAddress(*pubKey)
		for _, u0 := range self.config.Users {
			a := common.HexToAddress(u0.Address)
			if bytes.Compare(a.Bytes(), address.Bytes()) == 0 {
				return &u0, nil
			}
		}
		return u, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:VerifyMessage] Address not found: %x", address.Bytes()), ErrorCode: 421, ErrorMessage: "User Address not configured on connected node."}
	}

}

// using a users public/secret key, decrypt the data
func (self *KeyManager) DecryptData(u *SWARMDBUser, data []byte) (b []byte, err error) {
	var decryptNonce [24]byte
	copy(decryptNonce[:], data[:24])

	decrypted, ok := box.Open(nil, data[24:], &decryptNonce, &u.publicK, &u.secretK)
	if !ok {
		return b, &sdbc.SWARMDBError{Message: fmt.Sprintf("[keymanager:DecryptData] box.Open"), ErrorCode: 456, ErrorMessage: "Failure Decrypting Data"}
	}
	return decrypted, nil
}

// using a users public/secret key, decrypt the data
func (self *KeyManager) EncryptData(u *SWARMDBUser, data []byte) []byte {
	var nonce [24]byte
	// TODO: make nonce random
	nonce = [24]byte{4, 0, 50, 203, 12, 81, 11, 49, 236, 255, 155, 11, 101, 6, 97, 233, 94, 169, 107, 4, 37, 57, 106, 151}
	msg := data
	encrypted := box.Seal(nonce[:], msg, &nonce, &u.publicK, &u.secretK)
	return encrypted
}
