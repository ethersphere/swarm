package api

import (
	"crypto/ecdsa"
	"io/ioutil"
	"os"
	"fmt"
	"path/filepath"
	"encoding/json"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	port = "8500"
)

/***
 * \todo in swarm impl privatekey is stored in swap package does that make a difference for security?
 */
type Config struct {
	// serialised/persisted fields
	Path      string
	Port      string
	PublicKey string
	PrivateKey	*ecdsa.PrivateKey
	METAKey    string
	NetworkId uint64
}


// config is agnostic to where private key is coming from
// so managing accounts is outside META and left to wrappers
func NewConfig(path string, prvKey *ecdsa.PrivateKey, networkId uint64) (self *Config, err error) {
	address := crypto.PubkeyToAddress(prvKey.PublicKey) // default beneficiary address
	dirpath := filepath.Join(path, "META-"+common.Bytes2Hex(address.Bytes()))
	err = os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		return
	}
	confpath := filepath.Join(dirpath, "config.json")
	var data []byte
	pubkey := crypto.FromECDSAPub(&prvKey.PublicKey)
	pubkeyhex := common.ToHex(pubkey)
	keyhex := crypto.Sha3Hash(pubkey).Hex()

	self = &Config{
		Port:          port,
		Path:          dirpath,
		PrivateKey:		prvKey,
		PublicKey:     pubkeyhex,
		METAKey:       keyhex,
		NetworkId:     networkId,
	}
	data, err = ioutil.ReadFile(confpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		// file does not exist
		// write out config file
		err = self.Save()
		if err != nil {
			err = fmt.Errorf("error writing config: %v", err)
		}
		return
	}
	// file exists, deserialise
	err = json.Unmarshal(data, self)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %v", err)
	}
	// check public key
	if pubkeyhex != self.PublicKey {
		return nil, fmt.Errorf("public key does not match the one in the config file %v != %v", pubkeyhex, self.PublicKey)
	}
	if keyhex != self.METAKey {
		return nil, fmt.Errorf("META key does not match the one in the config file %v != %v", keyhex, self.METAKey)
	}

	return
}


func (self *Config) Save() error {
	data, err := json.MarshalIndent(self, "", "    ")
	if err != nil {
		return err
	}
	err = os.MkdirAll(self.Path, os.ModePerm)
	if err != nil {
		return err
	}
	confpath := filepath.Join(self.Path, "config.json")
	return ioutil.WriteFile(confpath, data, os.ModePerm)
}
