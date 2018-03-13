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

package swarmdb

import (
	"encoding/json"
	"fmt"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	//sdbc "swarmdbcommon"
	"io/ioutil"
)

// SwarmDB Configuration Defaults
const (
	SWARMDBCONF_FILE                  = "/usr/local/swarmdb/etc/swarmdb.conf"
	SWARMDBCONF_DEFAULT_PASSPHRASE    = "wolk"
	SWARMDBCONF_CHUNKDB_PATH          = "/usr/local/swarmdb/data"
	SWARMDBCONF_KEYSTORE_PATH         = "/usr/local/swarmdb/data/keystore"
	SWARMDBCONF_ENSDOMAIN             = "ens.wolk.com"
	SWARMDBCONF_LISTENADDR            = "0.0.0.0"
	SWARMDBCONF_PORTTCP               = 2001
	SWARMDBCONF_PORTHTTP              = 8501
	SWARMDBCONF_PORTENS               = 8545
	SWARMDBCONF_CURRENCY              = "WLK"
	SWARMDBCONF_TARGET_COST_STORAGE   = 2.71828
	SWARMDBCONF_TARGET_COST_BANDWIDTH = 3.14159
	SWARMDBCONF_ENS_KEYSTORE	  = "/var/www/vhosts/data/keystore"
	SWARMDBCONF_ENS_IP		  = "/var/www/vhosts/data/geth.ipc"
	SWARMDBCONF_ENS_ADDRESS		  = ""
)

type SWARMDBUser struct {
	Address        string `json:"address,omitempty"`        //value of val, usually the whole json record
	Passphrase     string `json:"passphrase,omitempty"`     // password to unlock key in keystore directory
	MinReplication int    `json:"minReplication,omitempty"` // should this be in config
	MaxReplication int    `json:"maxReplication,omitempty"` // should this be in config
	AutoRenew      int    `json:"autoRenew,omitempty"`      // should this be in config
	pk             []byte
	sk             []byte
	publicK        [32]byte
	secretK        [32]byte
}

type SWARMDBConfig struct {
	ListenAddrTCP string `json:"listenAddrTCP,omitempty"` // IP for TCP server
	PortTCP       int    `json:"portTCP,omitempty"`       // port for TCP server

	ListenAddrHTTP string `json:"listenAddrHTTP,omitempty"` // IP for HTTP server
	PortHTTP       int    `json:"portHTTP,omitempty"`       // port for HTTP server

	Address    string `json:"address,omitempty"`    // the address that earns, must be in keystore directory
	PrivateKey string `json:"privateKey,omitempty"` // to access child chain

	ChunkDBPath    string        `json:"chunkDBPath,omitempty"`    // the directory of the SWARMDB local databases (SWARMDBCONF_CHUNKDB_PATH)
	KeystorePath   string        `json:"usersKeysPath,omitempty"`  // directory containing the keystore of Ethereum wallets (SWARMDBCONF_KEYSTORE_PATH)
	Authentication int           `json:"authentication,omitempty"` // 0 - authentication is not required, 1 - required 2 - only users data stored
	Users          []SWARMDBUser `json:"users,omitempty"`          // array of users with permissions

	Currency            string  `json:"currency,omitempty"`            //
	TargetCostStorage   float64 `json:"targetCostStorage,omitempty"`   //
	TargetCostBandwidth float64 `json:"targetCostBandwidth,omitempty"` //

	EnsKeyPath	    string  `json:"ensKeyPath,omitempty"`
	EnsIP		    string  `json:"ensIP,omitempty"`	
 	EnsAddress	    string  `json:"ensAddress,omitempty"`
}

func (self *SWARMDBConfig) GetNodeID() (out string) {
	// TODO: replace with public key of farmer
	return "abcd"
}

func (self *SWARMDBConfig) GetSWARMDBUser() (u *SWARMDBUser) {
	for _, user := range self.Users {
		return &user
	}
	return u
}

func GenerateSampleSWARMDBConfig(privateKey string, address string, passphrase string) (c SWARMDBConfig) {
	var u SWARMDBUser
	u.Address = address
	u.Passphrase = passphrase
	u.MinReplication = 3
	u.MaxReplication = 5
	u.AutoRenew = 1

	c.ListenAddrTCP = SWARMDBCONF_LISTENADDR
	c.PortTCP = SWARMDBCONF_PORTTCP

	c.ListenAddrHTTP = SWARMDBCONF_LISTENADDR
	c.PortHTTP = SWARMDBCONF_PORTHTTP

	c.Address = u.Address
	c.PrivateKey = privateKey

	c.Authentication = 1
	c.ChunkDBPath = SWARMDBCONF_CHUNKDB_PATH
	c.KeystorePath = SWARMDBCONF_KEYSTORE_PATH
	c.Users = append(c.Users, u)

	c.Currency = SWARMDBCONF_CURRENCY
	c.TargetCostStorage = SWARMDBCONF_TARGET_COST_STORAGE
	c.TargetCostBandwidth = SWARMDBCONF_TARGET_COST_BANDWIDTH
	return c
}

func SaveSWARMDBConfig(c SWARMDBConfig, filename string) (err error) {
	// save file
	cout, err1 := json.MarshalIndent(c, "", "\t")
	if err1 != nil {
		return &sdbc.SWARMDBError{Message: fmt.Sprintf("[config:SaveSWARMDBConfig] Marshal %s", err1.Error()), ErrorCode: 457, ErrorMessage: "Unable to Save Config File"}
	} else {
		err := ioutil.WriteFile(filename, cout, 0644)
		if err != nil {
			return &sdbc.SWARMDBError{Message: fmt.Sprintf("[config:SaveSWARMDBConfig] WriteFile %s", err.Error()), ErrorCode: 457, ErrorMessage: "Unable to Save Config File"}
		}
	}
	return nil
}

func LoadSWARMDBConfig(filename string) (c *SWARMDBConfig, err error) {
	// read file
	c = new(SWARMDBConfig)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, &sdbc.SWARMDBError{Message: fmt.Sprintf("[config:LoadSWARMDBConfig] ReadFile %s", err.Error()), ErrorCode: 458, ErrorMessage: "Unable to Load Config File"}
	}
	err = json.Unmarshal(dat, c)
	if err != nil {
		return c, &sdbc.SWARMDBError{Message: fmt.Sprintf("[config:LoadSWARMDBConfig] Unmarshal %s", err.Error()), ErrorCode: 458, ErrorMessage: "Unable to Load Config File"}
	}
	return c, nil
}
