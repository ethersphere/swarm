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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	//sdbc "swarmdbcommon"
	sdbc "github.com/ethereum/go-ethereum/swarm/swarmdb/swarmdbcommon"
	_ "github.com/mattn/go-sqlite3"
	"math/big"
	"path/filepath"
	"sync"
	"time"
)

type SwapCheck struct {
	SwapID      []byte
	Sender      common.Address // address of sender
	Beneficiary common.Address
	Amount      *big.Int
	Timestamp   []byte
	Sig         []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

type SwapLog struct {
	SwapID      string `json:"swapID"`
	Sender      string `json:"sender"`
	Beneficiary string `json:"beneficiary"`
	Amount      int    `json:"amount"`
	Sig         string `json:"sig"` // of sender or beneficiary
	CheckBD     int    `json:"chunkSD"`
}

type Promise interface{}

// interface for the peer protocol for testing or external alternative payment
type Protocol interface {
	SDBPay(int, Promise) // units, payment proof
	Drop()
	String() string
}

type SwapDBStore struct {
	db       *sql.DB
	filepath string
	netstats *Netstats
	km       *KeyManager
}

type SwapDB struct {
	swapdbstore  *SwapDBStore
	proto        Protocol   // peer communication protocol
	lock         sync.Mutex // mutex for balance access
	balance      int        // units of chunk/retrieval request
	remotePayAt  uint       // remote peer's PayAt
	localAddress common.Address
	peerAddress  common.Address
}

func NewSwapDBStore(config *SWARMDBConfig, netstats *Netstats) (self *SwapDBStore, err error) {
	path := filepath.Join(config.ChunkDBPath, "swap.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil || db == nil {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:NewSwapDBStore] Open %s", err.Error())}
	}

	//Local Chunk table
	sql_table := `
    CREATE TABLE IF NOT EXISTS swap (
    swapID TEXT NOT NULL PRIMARY KEY,
    sender TEXT,
    beneficiary TEXT,
    amount INTEGER DEFAULT 1,
    sig    TEXT,
    checkBirthDT DATETIME
    );
    `
	_, err = db.Exec(sql_table)
	if err != nil {
		return nil, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:NewSwapDBStore] Exec - SQLite Chunk Table Creation %s", err.Error())}
	}

	self = &SwapDBStore{
		db:       db,
		filepath: path,
		netstats: netstats,
	}

	return self, nil
}

func (self *SwapDBStore) Issue(amount int, localAddress common.Address, peerAddress common.Address) (ch *SwapCheck, err error) {
	// compute the swapID = Keccak256(sender, beneficiary, amount, timestamp ...)
	timestamp := time.Now()
	timestampStr := timestamp.String()               // 2009-11-10 23:00:00 +0000 UTC m=+0.000000001
	timestampSubstring := string(timestampStr[0:19]) // 2009-11-10 23:00:00
	timestampByte := []byte(timestampSubstring)      // [50 48 48 57 45 49 49 45 49 48 32 50 51 58 48 48 58 48 48]      size 19 byte

	amount8 := IntToByte(amount)
	var raw []byte
	raw = make([]byte, 67)
	copy(raw[0:], localAddress[:20])
	copy(raw[20:], peerAddress[:20])
	copy(raw[40:], amount8[:8])
	copy(raw[48:], timestampByte[:19])
	swapID := crypto.Keccak256(raw)

	// TODO: use keymanager to sign message
	// sig, err = km.SignMessage(swapID)
	// if err != nil {
	//	return ch, &sdbc.SWARMDBError{message: fmt.Sprintf("[swapdb:Issue] SignMessage %s", err.Error())}
	// } else {
	var sig []byte
	sig = []byte{49, 50, 51}
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Issue swapID: %v sender: %v beneficiary: %v amount: %v  sig: %v", swapID, localAddress, peerAddress, amount, sig))

	sql_add := `INSERT OR REPLACE INTO swap ( swapID, sender, beneficiary, amount, sig, checkBirthDT) values(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
	stmt, err := self.db.Prepare(sql_add)
	if err != nil {
		return ch, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Issue] Prepare %s", err.Error())}
	}
	defer stmt.Close()

	swapID_str := fmt.Sprintf("%x", swapID)
	sender_str := fmt.Sprintf("%x", localAddress)
	beneficiary_str := fmt.Sprintf("%x", peerAddress)
	amount_int := amount
	sig_str := fmt.Sprintf("%x", sig)

	_, err = stmt.Exec(swapID_str, sender_str, beneficiary_str, amount_int, sig_str)
	if err != nil {
		return ch, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Issue] Exec %s", err.Error())}
	}
	stmt.Close()

	amountB := big.NewInt(int64(-amount))
	ch = &SwapCheck{
		SwapID:      swapID,        // []byte
		Sender:      localAddress,  // common.Address // address of sender
		Beneficiary: peerAddress,   // common.Address
		Amount:      amountB,       // int
		Timestamp:   timestampByte, //size 19 byte
		Sig:         sig,           // []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
	}

	self.netstats.AddIssue(amount)

	return ch, nil
}

func (self *SwapDBStore) Receive(units int, ch *SwapCheck) (err error) {
	swapIDIssue := ch.SwapID      // []byte
	sender := ch.Sender           // common.Address // address of sender
	beneficiary := ch.Beneficiary // common.Address
	amountB := ch.Amount          // big.NewIn
	timestamp := ch.Timestamp     //size 19 byte
	sig := ch.Sig

	amount8 := IntToByte(-units)
	var raw []byte
	raw = make([]byte, 67)
	copy(raw[0:], sender[:20])
	copy(raw[20:], beneficiary[:20])
	copy(raw[40:], amount8[:8])
	copy(raw[48:], timestamp[:19])
	swapIDReceive := crypto.Keccak256(raw)

	price := big.NewInt(int64(units))
	if price.Cmp(amountB) != 0 {
		return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Receive] units != amount")}
	} else {
		// TODO: check signature instead of hash
		if bytes.Equal(swapIDIssue, swapIDReceive) {
			log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Receive swapID: %v sender: %v beneficiary: %v amount: %v  sig: %v", swapIDReceive, sender, beneficiary, amountB, sig))

			sql_add := `INSERT OR REPLACE INTO swap ( swapID, sender, beneficiary, amount, sig, checkBirthDT) values(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
			stmt, err := self.db.Prepare(sql_add)
			if err != nil {
				return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Receive] Prepare %s", err.Error())}
			}
			defer stmt.Close()

			swapID_str := fmt.Sprintf("%x", swapIDReceive)
			sender_str := fmt.Sprintf("%x", sender)
			beneficiary_str := fmt.Sprintf("%x", beneficiary)
			amount_int := units // int
			sig_str := fmt.Sprintf("%x", sig)

			_, err = stmt.Exec(swapID_str, sender_str, beneficiary_str, amount_int, sig_str)
			if err != nil {
				return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Receive] Exec %s", err.Error())}
			}
			stmt.Close()

			self.netstats.AddReceive(units)

		} else {
			return &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Receive] sig != sig")}
		}
	}
	return nil
}

func (self *SwapDBStore) GenerateSwapLog(startts int64, endts int64) (log []string, err error) {
	rows, err := self.db.Query("SELECT swapID, sender, beneficiary, amount, sig FROM swap")
	if err != nil {
		return log, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:GenerateSwapLog] Query %s", err.Error())}
	}

	defer rows.Close()

	for rows.Next() {
		c := SwapLog{}
		err = rows.Scan(&c.SwapID, &c.Sender, &c.Beneficiary, &c.Amount, &c.Sig)
		if err != nil {
			return log, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:GenerateSwapLog] Scan %s", err.Error())}
		}

		l, err2 := json.Marshal(c)
		if err2 != nil {
			return log, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:GenerateSwapLog] Marshal %s", err2.Error())}

		}

		s := fmt.Sprintf("%s\n", l)
		log = append(log, s)
	}
	rows.Close()
	self.netstats.GenerateSwapLog()
	return log, nil
}

func NewSwapDB(swapdbstore *SwapDBStore, proto Protocol, remotePayAt uint, localAddress common.Address, peerAddress common.Address) (self *SwapDB, err error) {
	localAddressHex := localAddress.Hex()
	peerAddressHex := peerAddress.Hex()
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.NewSwapDB Beneficiary: local: %v peer: %v", localAddressHex, peerAddressHex))

	self = &SwapDB{
		swapdbstore:  swapdbstore,
		proto:        proto,
		balance:      0,
		remotePayAt:  remotePayAt,
		localAddress: localAddress,
		peerAddress:  peerAddress,
	}
	return self, nil
}

// Add(n)
// n > 0 called when promised/provided n units of service
// n < 0 called when used/requested n units of service
func (self *SwapDB) Add(n int) error {
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Add amount: %v", n))
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance += n
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Add self.balance: %v self.remotePayAt: %v", self.balance, self.remotePayAt))
	if self.balance <= -int(self.remotePayAt) {
		self.Issue()
	}
	return nil
}

// Issue creates a new signed by the farmer's private key for the beneficiary and amount
//func (self *SwapDB) Issue(km *KeyManager, u *SWARMDBUser, beneficiary common.Address, amount int) (ch *SwapCheck, err error) {
func (self *SwapDB) Issue() (err error) {
	localAddressHex := self.localAddress.Hex()
	peerAddressHex := self.peerAddress.Hex()
	amount := self.balance
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Issue local: %v peer: %v self.balance: %v", localAddressHex, peerAddressHex, amount))

	//	defer self.lock.Unlock()
	//	self.lock.Lock()
	//		if amount < 0 {
	//		return ch, &sdbc.SWARMDBError{Message: fmt.Sprintf("[swapdb:Issue] Check Amount must be positive %d", amount)}
	//		}
	ch, err := self.swapdbstore.Issue(self.balance, self.localAddress, self.peerAddress)
	if err != nil {
	}
	self.balance = self.balance - amount
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Issue self.balance: %v", self.balance))
	self.proto.SDBPay(-amount, ch)
	return err
}

// receive(units, promise) is called by the protocol when a payment msg is received
// returns error if promise is invalid.
func (self *SwapDB) Receive(units int, promise Promise) (err error) {
	// TODO: address case where type cast fails
	ch := promise.(*SwapCheck)
	err = self.swapdbstore.Receive(units, ch)
	if err != nil {
		return err
	}
	self.balance = self.balance - units
	log.Debug(fmt.Sprintf("[wolk-cloudstore] swapdb.Receive self.balance: %v", self.balance))
	return nil
}
