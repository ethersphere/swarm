package chequebook

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	// "github.com/ethereum/go-ethereum/xeth"
)

const (
	gasToCash  = "500000"   // gas cost of a cash transaction using chequebook
	cashAbiPre = "fbf788d6" // abi preamble signature for cash method of the chequebook
)

// rlp serialised cheque model for use with the chequebook
type Cheque struct {
	// the address of the contract itself needed to avoid cross-contract submission
	Sender    common.Address
	Recipient common.Address
	Amount    *big.Int
	Sig       []byte
}

// chequebook to create, sign cheques from single sender to multiple recipients
// outgoing payment handler for peer to peer micropayments
type Chequebook struct {
	*Cheque
	balance *big.Int          // not synced with blockchain
	prvKey  *ecdsa.PrivateKey // not to keep in memory
	sender  common.Address
	sent    map[common.Address]*big.Int
	lock    sync.Mutex
}

// New(sender, balance, prvKeyFunc) creates a new Chequebook
func NewChequebook(sender common.Address, prvKey *ecdsa.PrivateKey) (self *Chequebook, err error) {
	balance := new(big.Int)                   // should read from blockchain or file
	sent := make(map[common.Address]*big.Int) // should read from blockchain or file
	self = &Chequebook{
		balance: balance,
		prvKey:  prvKey,
		sender:  sender,
		sent:    sent,
	}
	glog.V(logger.Detail).Infof("\nnew chequebook initialised from %v ", sender)
	return
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

func (self *Chequebook) Balance() *big.Int {
	defer self.lock.Unlock()
	self.lock.Lock()
	return self.balance
}

// Deposit(amount) deposits amount to the checkbook account
// atm only used for bookkeeping
func (self *Chequebook) Deposit(amount *big.Int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance.Add(self.balance, amount)
}

type Backend interface {
	Transact(fromStr, toStr, nonceStr, valueStr, gasStr, gasPriceStr, codeStr string) (string, error)
	// Call(fromStr, toStr, valueStr, gasStr, gasPriceStr, codeStr string) (string, string, error)
}

// type ChequeQueue struct {
//   recipient common.Address
//   last      map[string]*Chequebox
// }

// chequebox to deposit, verify and cash cheques
// from a single sender to single recipient
// incoming payment handler for peer to peer micropayments
type Chequebox struct {
	lock      sync.Mutex
	signer    *ecdsa.PublicKey
	sender    common.Address
	recipient common.Address
	cashed    bool
	txhash    string
	*Cheque
}

func NewChequebox(sender, recipient common.Address, signer *ecdsa.PublicKey) (self *Chequebox, err error) {
	self = &Chequebox{
		sender:    sender,
		recipient: recipient,
		signer:    signer,
	}
	return
}

// AutoCash(d) starts a loop that periodically clears the last check
// if the peer is trusted, clearing period could be 24h, or a week
func (self *Chequebox) AutoCash(be Backend, cashInterval time.Duration) (quit chan bool) {
	quit = make(chan bool)
	ticker := time.NewTicker(cashInterval)
	go func() {
	FOR:
		for {
			select {
			case <-quit:
				break FOR
			case <-ticker.C:
				self.lock.Lock()
				if self.Cheque != nil && !self.cashed {
					txhash, err := self.Cheque.Cash(be)
					if err == nil {
						self.cashed = true
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
	if self.Cheque != nil {
		sum = self.Cheque.Amount
	}

	amount, err := ch.Verify(self.signer, self.sender, self.recipient, sum)
	if err == nil {
		self.Cheque = ch
		self.cashed = false
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

// abi encoding of a cheque to send as eth tx data
func (self *Cheque) abiEncode() string {
	v, r, s := sig2rsv(self.Sig)
	// cashAbiPre, sender, recipient, amount, v, r, s
	bigamount := self.Amount.Bytes()
	if len(bigamount) > 32 {
		glog.V(logger.Detail).Infof("number too big: %v (>32 bytes)", self.Amount)
		return ""
	}
	var amount32, vabi [32]byte
	copy(amount32[32-len(bigamount):32], bigamount)
	vabi[31] = v
	return cashAbiPre + self.Sender.Hex()[2:] + self.Recipient.Hex()[2:] + common.Bytes2Hex(amount32[:]) +
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
		if amount.Cmp(common.Big0) > 0 {
			return nil, fmt.Errorf("incorrect amount: %v <= %v", self.Amount, sum)
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
func (self *Cheque) Cash(b Backend) (string, error) {
	return b.Transact(self.Recipient.Hex(), self.Sender.Hex(), "", "", "", gasToCash, self.abiEncode())
}
