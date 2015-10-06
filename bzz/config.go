package bzz

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	port = "8500"
)

// separate bzz directories
// allow several bzz nodes running in parallel
type Config struct {
	// serialised/persisted fields
	Swap *swapParams
	Path string
	Port string
	// not serialised/not persisted fields
	// address // node address
}

// SetBackend is meant to be called once
// Backend interface implemented by xeth.XEth or JSON-IPC client
func (self *Config) SetBackend(backend chequebook.Backend) (err error) {
	self.Swap.setBackend(self.Path, backend)
	return
}

// config is agnostic to where private key is coming from
// so managing accounts is outside swarm and left to wrappers
func NewConfig(path string, contract common.Address, prvKey *ecdsa.PrivateKey) (self *Config, err error) {

	address := crypto.PubkeyToAddress(prvKey.PublicKey) // default beneficiary address
	confpath := filepath.Join(path, common.Bytes2Hex(address.Bytes())+".json")
	var data []byte

	data, err = ioutil.ReadFile(confpath)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		// file does not exist

		config := &Config{
			Port: port,
			Path: path,
			Swap: defaultSwapParams(contract, prvKey),
		}
		// write out config file
		data, err = json.MarshalIndent(config, "", "    ")
		if err != nil {
			return nil, fmt.Errorf("error writing config: %v", err)
		}
		err = ioutil.WriteFile(confpath, data, os.ModePerm)

	} else {
		// file exists, deserialise
		self = &Config{}
		err = json.Unmarshal(data, self)
		if err != nil {
			return nil, err
		}
		// check public key
		pubkey := common.ToHex(crypto.FromECDSAPub(&prvKey.PublicKey))
		if pubkey != self.Swap.PublicKey {
			return nil, fmt.Errorf("key does not match the one in the config file %v != %v", pubkey, self.Swap.PublicKey)
		}

	}

	return
}
