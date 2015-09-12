package chequebook

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

/*
Chequebook package is a go API to the 'chequebook' ethereum smart contract
With convenience methods that allow using chequebook for
* issuing, receiving, verifying and (auto)cashing cheques in ether
* (auto)depositing ether to the chequebook contract
* watching peer solvency and notifying of bouncing cheques

Some functionality require interacting with the blockchain:
* setting current balance on peer's chequebook
* sending the transaction to cash the cheque
* depositing ether to the chequebook
* watching incoming ether

Backend is the interface for that
*/

const (
	gasToCash     = "500000"   // gas cost of a cash transaction using chequebook
	getSentAbiPre = "d75d691d" // sent amount accessor in the checkbook contract
	cashAbiPre    = "fbf788d6" // abi preamble signature for cash method of the chequebook
)

// rlp serialised cheque model for use with the chequebook
type Cheque struct {
	// the address of the contract itself needed to avoid cross-contract submission
	Sender    common.Address // contract address
	Recipient common.Address // beneficiary
	Amount    *big.Int       // cumulative amount of all funds sent
	Sig       []byte         // signature Sign(Sha3(sender, recipient, amount), prvKey)
}

// chequebook to create, sign cheques from single sender to multiple recipients
// outgoing payment handler for peer to peer micropayments
type Chequebook struct {
	path   string            // path to checkbook file
	prvKey *ecdsa.PrivateKey // private key to sign cheque with
	lock   sync.Mutex        //

	// persisted fields
	balance *big.Int                    // not synced with blockchain
	sender  common.Address              // contract address
	sent    map[common.Address]*big.Int //tallies for recipients
}

// NewChequebook(path, sender, balance, prvKey) creates a new Chequebook
func NewChequebook(path string, sender common.Address, prvKey *ecdsa.PrivateKey) (self *Chequebook, err error) {
	balance := new(big.Int)
	sent := make(map[common.Address]*big.Int)
	self = &Chequebook{
		balance: balance,
		sender:  sender,
		sent:    sent,
		path:    path,
		prvKey:  prvKey,
	}
	glog.V(logger.Detail).Infof("\nnew chequebook initialised from %v ", sender)
	return
}

// LoadChequebook(path, prvKey) loads a chequebook from disk (file path)
func LoadChequebook(path string, prvKey *ecdsa.PrivateKey) (self *Chequebook, err error) {
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	self, _ = NewChequebook(path, common.Address{}, prvKey)

	err = json.Unmarshal(data, self)
	if err != nil {
		return nil, err
	}

	return
}

type chequebookFile struct {
	Balance string
	Sender  string
	Sent    map[string]string
}

func (self *Chequebook) UnmarshalJSON(data []byte) error {
	var file chequebookFile
	err := json.Unmarshal(data, &file)
	if err != nil {
		return err
	}
	_, ok := self.balance.SetString(file.Balance, 10)
	if !ok {
		return fmt.Errorf("cumulative amount sent: unable to convert string to big integer: %v", file.Balance)
	}
	self.sender = common.HexToAddress(file.Sender)
	for addr, sent := range file.Sent {
		self.sent[common.HexToAddress(addr)], ok = new(big.Int).SetString(sent, 10)
		if !ok {
			return fmt.Errorf("recipient %v cumulative amount sent: unable to convert string to big integer: %v", addr, sent)
		}
	}
	return nil
}

func (self *Chequebook) MarshalJSON() ([]byte, error) {
	var file = &chequebookFile{
		Balance: self.balance.String(),
		Sender:  self.sender.Hex(),
		Sent:    make(map[string]string),
	}
	for addr, sent := range self.sent {
		file.Sent[addr.Hex()] = sent.String()
	}
	return json.Marshal(file)
}

// Save() persists the checkbook on disk
// remembers balance, contract address and
// cumulative amount of funds sent for each recipient
func (self *Chequebook) Save() (err error) {
	data, err := json.MarshalIndent(self, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(self.path, data, os.ModePerm)
}

// New(recipient, amount) will create a Cheque
// the cheque is signed by the checkbook owner's private key
// the signer commits to a contract (one that they own), a recipient and an amount
func (self *Chequebook) NewCheque(recipient common.Address, amount *big.Int) (ch *Cheque, err error) {
	defer self.lock.Unlock()
	self.lock.Lock()
	if amount.Cmp(common.Big0) <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero (%v)", amount)
	}
	if self.balance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficent funds to issue cheque for amount: %v. balance: %v", amount, self.balance)
	}
	var sig []byte
	sent, found := self.sent[recipient]
	if !found {
		sent = new(big.Int)
		self.sent[recipient] = sent
	}
	sum := new(big.Int).Set(sent)
	sum.Add(sum, amount)
	sig, err = crypto.Sign(sigHash(self.sender, recipient, sum), self.prvKey)
	if err == nil {
		ch = &Cheque{
			Sender:    self.sender,
			Recipient: recipient,
			Amount:    sum,
			Sig:       sig,
		}
		sent.Set(sum)                          // remember total sent
		self.balance.Sub(self.balance, amount) // subtract amount from balance
	}

	return
}

// data to sign: contract address, beneficiary, cumulative amount of funds ever sent
func sigHash(sender, recipient common.Address, sum *big.Int) []byte {
	bigamount := sum.Bytes()
	if len(bigamount) > 32 {
		return nil
	}
	var amount32 [32]byte
	copy(amount32[32-len(bigamount):32], bigamount)
	input := append(sender.Bytes(), recipient.Bytes()...)
	input = append(input, amount32[:]...)
	return crypto.Sha3(input)
}

// GetBalance() public accessor for Balance
func (self *Chequebook) Balance() *big.Int {
	defer self.lock.Unlock()
	self.lock.Lock()
	return new(big.Int).Set(self.balance)
}

// Deposit(amount) deposits amount to the checkbook account
// atm only used for bookkeeping
func (self *Chequebook) Deposit(amount *big.Int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance.Add(self.balance, amount)
}

// Backends is the interface to interact with the Ethereum blockchain
// implemented by xeth.XEth
type Backend interface {
	Transact(fromStr, toStr, nonceStr, valueStr, gasStr, gasPriceStr, codeStr string) (string, error)
	Call(fromStr, toStr, valueStr, gasStr, gasPriceStr, codeStr string) (string, string, error)
}

// type ChequeQueue struct {
//   recipient common.Address
//   last      map[string]*Chequebox
// }

// chequebox to deposit, verify and cash cheques
// from a single sender to single recipient
// incoming payment handler for peer to peer micropayments
type Chequebox struct {
	lock        sync.Mutex
	signer      *ecdsa.PublicKey
	sender      common.Address
	recipient   common.Address
	txhash      string
	backend     Backend
	quit        chan bool // when closed causes autocash to stop
	maxUncashed *big.Int
	cashed      *big.Int
	*Cheque
}

// NewChequebox(sender, recipient, signer, backend) constructor for Chequebox
// not persisted, cumulative sum updated from blockchain when first check received
// backend used to sync amount (Call) as well as cash the cheques (Transact)
func NewChequebox(sender, recipient common.Address, signer *ecdsa.PublicKey, backend Backend) (self *Chequebox, err error) {
	self = &Chequebox{
		sender:    sender,
		recipient: recipient,
		signer:    signer,
		backend:   backend,
		cashed:    new(big.Int).Set(common.Big0),
	}
	return
}

// Stop() quits the autocash loop if its running
func (self *Chequebox) Stop() {
	defer self.lock.Unlock()
	self.lock.Lock()
	if self.quit != nil {
		close(self.quit)
		self.quit = nil
	}
}

// AutoCash(cashInterval, maxUncashed) (re)sets maximum time and amount which
// triggers cashing of the last uncashed cheque
// if maxUncashed is set to 0, then autocash on receipt
func (self *Chequebox) AutoCash(cashInterval time.Duration, maxUncashed *big.Int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.maxUncashed = maxUncashed
	self.autoCash(cashInterval)
}

// autoCash(d) starts a loop that periodically clears the last check
// if the peer is trusted, clearing period could be 24h, or a week
// caller holds the lock
func (self *Chequebox) autoCash(cashInterval time.Duration) {
	if self.quit != nil {
		close(self.quit)
		self.quit = nil
	}
	// if maxUncashed is set to 0, then autocash on receipt
	if cashInterval == time.Duration(0) || self.maxUncashed != nil && self.maxUncashed.Cmp(common.Big0) == 0 {
		return
	}

	ticker := time.NewTicker(cashInterval)
	self.quit = make(chan bool)
	quit := self.quit
	go func() {
	FOR:
		for {
			select {
			case <-quit:
				break FOR
			case <-ticker.C:
				self.lock.Lock()
				if self.Cheque != nil && self.Amount.Cmp(self.cashed) != 0 {
					txhash, err := self.Cash(self.backend)
					if err == nil {
						self.cashed = self.Amount
						self.txhash = txhash
					}
				}
				self.lock.Unlock()
			}
		}
	}()
	return
}

// Reveive(cheque) called to deposit latest cheque to incoming Chequebox
func (self *Chequebox) Receive(ch *Cheque) (*big.Int, error) {
	defer self.lock.Unlock()
	self.lock.Lock()
	var sum *big.Int
	if self.Cheque == nil {
		// the sum is checked against the blockchain once a check is received
		tally, _, err := self.backend.Call(self.recipient.Hex(), self.sender.Hex(), "", "", "", getSentAbiEncode(ch.Sender))
		if err != nil {
			return nil, fmt.Errorf("checkbox: error calling backend to set amount: %v", err)
		}
		var ok bool
		sum, ok = new(big.Int).SetString(tally, 10)
		if !ok {
			return nil, fmt.Errorf("checkbox: cannot convert amount to integer")
		}

	} else {
		sum = self.Amount
	}

	amount, err := ch.Verify(self.signer, self.sender, self.recipient, sum)
	if err == nil {
		self.Cheque = ch

		if self.maxUncashed != nil {
			uncashed := new(big.Int).Sub(ch.Amount, self.cashed)
			if self.maxUncashed.Cmp(uncashed) < 0 {
				ch.Cash(self.backend)
				self.cashed = ch.Amount
			}
		}
	}

	return amount, err
}

// RSV representation of signature
func sig2rsv(sig []byte) (v byte, r, s []byte) {
	v = sig[64] + 27
	r = sig[:32]
	s = sig[32:64]
	return
}

func getSentAbiEncode(recipient common.Address) string {
	return getSentAbiPre + recipient.Hex()[2:]
}

// abi encoding of a cheque to send as eth tx data
func (self *Cheque) cashAbiEncode() string {
	v, r, s := sig2rsv(self.Sig)
	// cashAbiPre, recipient, amount, v, r, s
	bigamount := self.Amount.Bytes()
	if len(bigamount) > 32 {
		glog.V(logger.Detail).Infof("number too big: %v (>32 bytes)", self.Amount)
		return ""
	}
	var amount32, vabi [32]byte
	copy(amount32[32-len(bigamount):32], bigamount)
	vabi[31] = v
	return cashAbiPre + self.Recipient.Hex()[2:] + common.Bytes2Hex(amount32[:]) +
		common.Bytes2Hex(vabi[:]) + common.Bytes2Hex(r) + common.Bytes2Hex(s)
}

// Verify(cheque) verifies cheque for signer, sender, recipient, amount, valid signature
func (self *Cheque) Verify(signerKey *ecdsa.PublicKey, sender, recipient common.Address, sum *big.Int) (*big.Int, error) {
	if self.Recipient != recipient {
		return nil, fmt.Errorf("recipient mismatch: %v != %v", self.Recipient.Hex(), recipient.Hex())
	}
	if self.Sender != sender {
		return nil, fmt.Errorf("sender mismatch: %v != %v", self.Sender.Hex(), sender.Hex())
	}

	amount := new(big.Int).Set(self.Amount)
	if sum != nil {
		amount.Sub(self.Amount, sum)
		if amount.Cmp(common.Big0) <= 0 {
			return nil, fmt.Errorf("incorrect amount: %v <= 0", amount)
		}
	}

	pubKey, err := crypto.SigToPub(sigHash(self.Sender, recipient, self.Amount), self.Sig)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %v", err)
	}
	if !bytes.Equal(crypto.FromECDSAPub(pubKey), crypto.FromECDSAPub(signerKey)) {
		return nil, fmt.Errorf("signer mismatch: %x != %x", pubKey, signerKey)
	}
	return amount, nil
}

// Cash(backend) will cash the check using xeth backend to send a transaction
// Recipient address should be unlocked
func (self *Cheque) Cash(backend Backend) (string, error) {
	return backend.Transact(self.Recipient.Hex(), self.Sender.Hex(), "", "", "", gasToCash, self.cashAbiEncode())
}
