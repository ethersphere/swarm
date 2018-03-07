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
package swarmdb_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"swarmdb"
	"testing"
)

func TestConfigGeneration(t *testing.T) {
	filename := swarmdb.SWARMDBCONF_FILE
	// if _, err := os.Stat(filename); os.IsNotExist(err) {
	_, err := swarmdb.NewKeyManagerWithoutConfig(filename, swarmdb.SWARMDBCONF_DEFAULT_PASSPHRASE)
	if err != nil {
		t.Fatalf("[keymanager_test:TestConfigGeneration] NewKeyManagerWithoutConfig %s", err.Error())
	}
	// }
}

// Test SignMessage and VerifyMessage
func TestSignVerifyMessage(t *testing.T) {
	// signatures are 65 bytes, 130 chars [this one is a bogus signature]
	sig_bytes, e1 := hex.DecodeString("1f7b169c846f218ab552fa82fbf86758bf5c97d2d2a313e4f95957818a7b3edca492f2b8a67697c4f91d9b9332e8234783de17bd7a25e0a9f6813976eadf26deb5")
	if e1 != nil {
		t.Fatal(e1)
	}

	// challenges are 32 bytes, 64 bytes
	challenge_bytes, e2 := hex.DecodeString("b0e33f362d4345fe36103d0f62f9ab8e480b0ed4467726b15733afed9a4d4cc1")
	if e2 != nil {
		t.Fatal(e2)
	}

	config, errConfig := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	if errConfig != nil {
		t.Fatal("Failure to open Config", errConfig)
	}

	km, err := swarmdb.NewKeyManager(config)
	if err != nil {
		t.Fatal("Failure to open KeyManager", err)
	}

	// Test if bogus signature is correctly rejected
	verified0, err2 := km.VerifyMessage(challenge_bytes, sig_bytes)
	if err2 != nil {
		fmt.Printf("Correct Reject0\n")
	} else if verified0 != nil {
		t.Fatal("Failure to Reject0: %s", err2)
	} else {
		t.Fatal("Failure to Reject0: %s", err2)
	}

	// take a variable length message, hash it into "msg_hash", sign it with SignMessage, and see if it is verified
	msg := "swarmdb"
	h256 := sha256.New()
	h256.Write([]byte(msg))
	msg_hash := h256.Sum(nil)

	sig, err4 := km.SignMessage(msg_hash)
	if err4 != nil {
		t.Fatal("sign err", err)
	}

	verified2, err5 := km.VerifyMessage(msg_hash, sig)
	if err5 != nil || (verified2 == nil) {
		t.Fatal("verify2 err", err)
	} else {
		fmt.Printf("Verified challenge %x signature %x\n", msg_hash, sig)
	}
}

// Test the KeyManager EncryptData and DecryptData
func TestEncryptDecrypt(t *testing.T) {

	// need a config file with a specific user
	config, errConfig := swarmdb.LoadSWARMDBConfig(swarmdb.SWARMDBCONF_FILE)
	if errConfig != nil {
		t.Fatal("Failure to open Config", errConfig)
	}

	km, err := swarmdb.NewKeyManager(config)
	if err != nil {
		t.Fatal("Failure to open KeyManager", err)
	}

	msg := "0123456789abcdef"
	r := []byte(msg)
	u := config.GetSWARMDBUser()

	// encrypt the msg using the specific users secret/private key
	encData := km.EncryptData(u, r)
	decData, err := km.DecryptData(u, encData)
	if err != nil {
		t.Fatal("Failure to decrypt", err.Error())
	}
	a := bytes.Compare(decData, r)
	if a != 0 {
		fmt.Printf("Encrypted data is [%v][%x]", encData, encData)
		fmt.Printf("Decrypted data is [%v][%s] => %d", decData, decData, a)
		t.Fatal("Failure to decrypt")
	} else {
		fmt.Printf("Success %s\n", msg)
	}
}
