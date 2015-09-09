package bzz

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/xeth"
)

// SWAP Swarm Accounting Protocol
// implements a peer to peer micropayment system

const (
	autoCashInterval = 60 * time.Second // default interval for autocash
)

// rlp serializable config passed in handshake
type swapInfo struct {
	Sender, Recipient common.Address // addresses for swarm sales
	BuyAt, SellAt     *big.Int       // accepted max price, offered sale price
}

type swap struct {
	local      *swapInfo
	remote     *bzzProtocol
	sell, buy  bool       // boolean flags
	lock       sync.Mutex // mutex for balance access
	balance    int        // simple signed int sufficient in units of chunk/retrieval request
	quit       chan bool
	incoming   *chequebook.Chequebox  // incoming checkbox
	chequebook *chequebook.Chequebook // outgoing checkbook
}

func newSwap(chequebook *chequebook.Chequebook, local *swapInfo, remote *bzzProtocol) (self *swap, err error) {
	self = &swap{
		local:      local,
		remote:     remote,
		chequebook: chequebook,
	}

	// check prices are greater than zero
	if remote.SellAt.Cmp(local.BuyAt) <= 0 {
		self.buy = true
	}
	if local.SellAt.Cmp(remote.BuyAt) <= 0 {
		self.sell = true
	}

	// check if addresses are given to issue checks to and
	if self.sell {
		if (local.Recipient == common.Address{}) {
			return fmt.Errorf("peer is buyer but local recipient address missing")
		}
		if (remote.Sender == common.Address{}) {
			return fmt.Errorf("peer is buyer but remote sender address missing")
		}
		if self.buy {
			if (remote.Recipient == common.Address{}) {
				return fmt.Errorf("peer is buyer but remote recipient address missing")
			}
			if (local.Sender == common.Address{}) {
				return fmt.Errorf("peer is buyer but local sender address missing")
			}
		}
	}

	// set up checkbook for this recipient
	self.incoming = chequebook.NewChequebox(local.Recipient)
	self.quit = self.incoming.AutoCash(remote.backend, autoCashInterval)

	return
}

// add(n) called when sending chunks = receiving retrieve requests
//                 OR sending cheques
func (self *swap) add(n int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance += n
	if self.balance > self.disconnectThreshold {
		self.remote.Disconnect()
	}
}

// sub(n) called when receiving chunks = receiving delivery responses
//                 OR receiving cheques
func (self *swap) sub(n int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance -= n
	if self.buy && self.balance < self.paymentThreshold {
		amount := new(big.Int).SetInt64(int64(-self.balance))
		amount.Mul(amount, self.remote.SellAt)
		ch, err = createCheque(amount)
		if err != nil {
			glog.V().Warnf("[BZZ] cannot issue cheque. Sender: %v, Recipient: %v, Amount: %v", self.local.Sender, self.remote.Recipient, amount)
		} else {
			self.remote.sendCheque(-self.balance, ch)
			self.add(-self.balance)
		}
	}
	return
}

// createCheque(amount) is called when the local peer is in debt
// higher than Payment threshold
func (self *swap) createCheque(amount *big.Int) (ch *chequebook.Cheque, err error) {
	return self.chequebook.NewCheque(self.remote.Recipient, amount)
}

// processCheque is called by the network protocol when a check is received
func (self *swap) processCheque(units int, ch *chequebook.Cheque) error {
	if units <= 0 {
		return fmt.Errorf("invalid amount: %v <= 0", units)
	}
	sum := new(big.Int).SetInt64(int64(units))
	sum.Mul(sum, self.local.SellAt)
	if err = self.incoming.Receive(self.pubKey, ch, sum); err != nil {
		return fmt.Errorf("invalid cheque: %v", err)
	}
	self.sub(units)
	return nil
}

// chequebookPath(datadir, sender)
