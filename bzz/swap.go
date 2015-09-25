package bzz

import (
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
)

// SWAP Swarm Accounting Protocol
// implements a peer to peer micropayment system

// these should come from bzz config
var (
	autoCashInterval     = 300 * time.Second // default interval for autocash
	autoCashThreshold    = big.NewInt(100)   // threshold that triggers autocash (wei)
	autoDepositInterval  = 300 * time.Second // default interval for autocash
	autoDepositThreshold = big.NewInt(100)   // threshold that triggers autodeposit (wei)
	autoDepositBuffer    = big.NewInt(200)   // buffer that is surplus for fork protection etc (wei)
	acceptedPrice        = big.NewInt(2)     // maximum chunk price host is willing to pay (wei)
	offerPrice           = big.NewInt(2)     // minimum chunk price host requires (wei)
	paymentThreshold     = 10                // threshold that triggers payment request (units)
	disconnectThreshold  = 30                // threshold that triggers disconnect (units)
)

// rlp serializable config passed in handshake
type SwapData struct {
	ID          *ecdsa.PublicKey //
	Contract    common.Address   // address of chequebook contract
	Beneficiary common.Address   // addresses for swarm sales
	BuyAt       *big.Int         // accepted max price for chunk
	SellAt      *big.Int         // offered sale price for chunk
	PayAt       int              // threshold that triggers payment request
	DropAt      int              // threshold that triggers disconnect
}

func NewSwapData(contract common.Address, id *ecdsa.PublicKey) *SwapData {
	return &SwapData{
		ID:          id,
		Contract:    contract,
		Beneficiary: crypto.PubkeyToAddress(*id),
		BuyAt:       acceptedPrice,
		SellAt:      offerPrice,
		PayAt:       paymentThreshold,
		DropAt:      disconnectThreshold,
	}
}

// interface for the bzz protocol for testing
type paymentProtocol interface {
	payment(*paymentMsgData)
	paymentRequest(*paymentRequestMsgData)
	Drop()
}

// swap is the swarm accounting protocol instance
// * pairwise accounting and payments
type swap struct {
	local      *SwapData              // local peer's swap data
	remote     *SwapData              // remote peer's swap data
	lock       sync.Mutex             // mutex for balance access
	balance    int                    // units of chunk/retrieval request
	chequebox  *chequebook.Chequebox  // incoming chequebox  (one per connection)
	chequebook *chequebook.Chequebook // outgoing chequebook (shared amoung protocol insts)
	proto      paymentProtocol        //
}

// swap constructor, parameters
// * global chequebook, assumed deployed service and
// * the balance is at buffer.
func newSwap(chbook *chequebook.Chequebook, local, remote *SwapData, proto paymentProtocol) (self *swap, err error) {

	self = &swap{
		local:      local,
		remote:     remote,
		chequebook: chbook,
		proto:      proto,
	}

	// check if addresses are given to issue and receive cheques
	if self.sells() { // ie. host receives payment
		if (local.Beneficiary == common.Address{}) {
			return nil, fmt.Errorf("host is seller but local Beneficiary address missing")
		}
		if (remote.Contract == common.Address{}) {
			return nil, fmt.Errorf("peer is buyer but remote Contract address missing")
		}
	}

	if self.buys() { // ie/ remote peer receives payment
		if (local.Contract == common.Address{}) {
			return nil, fmt.Errorf("host is buyer but local Contract address missing")
		}
		if (remote.Beneficiary == common.Address{}) {
			return nil, fmt.Errorf("peer is seller but remote Beneficiary address missing")
		}
	}

	self.chequebox, err = chequebook.NewChequebox(remote.Contract, local.Beneficiary, local.ID, chbook.Backend())
	// call autocash
	self.chequebox.AutoCash(autoCashInterval, autoCashThreshold)

	glog.V(logger.Info).Infof("[BZZ] SWAP auto cash ON for %v -> %v: interval = %v, threshold = %v, peer = %v)", local.Contract.Hex()[:8], remote.Contract.Hex()[:8], autoCashInterval, autoCashThreshold)

	return
}

// true if host is buying.
func (self *swap) buys() bool {
	return self.remote.SellAt.Cmp(self.local.BuyAt) <= 0
}

// true iff host is selling.
func (self *swap) sells() bool {
	return self.local.SellAt.Cmp(self.remote.BuyAt) <= 0
}

// NewChequebook(path, Contract, prvKey*ecdsa.PrivateKey, backend) wraps the
// chequebook initialiser and sets up autoDeposit to cover spending.
func newChequebook(path string, Contract common.Address, prvKey *ecdsa.PrivateKey, backend chequebook.Backend) (chbook *chequebook.Chequebook, err error) {
	chbook, err = chequebook.LoadChequebook(path, prvKey, backend)
	if err != nil {
		chbook, err = chequebook.NewChequebook(path, Contract, prvKey, backend)
		if err == nil {
			glog.V(logger.Info).Infof("[BZZ] SWAP auto deposit ON: interval = %v, threshold = %v, buffer = %v)", autoDepositInterval, autoDepositThreshold, autoDepositBuffer)
			chbook.AutoDeposit(autoDepositInterval, autoDepositThreshold, autoDepositBuffer)
		}
	}
	return
}

// add(n) called when sending chunks = receiving retrieve requests
//                 OR sending cheques.
func (self *swap) add(n int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance += n
	if self.balance > self.local.DropAt {
		self.proto.Drop()
	} else {
		if self.balance > self.local.PayAt {
			self.proto.paymentRequest(&paymentRequestMsgData{self.balance})
		}
	}
}

// sub(n) called when receiving chunks = receiving delivery responses
//                 OR receiving cheques.
func (self *swap) sub(n int) {
	defer self.lock.Unlock()
	self.lock.Lock()
	self.balance -= n
}

// send(units) is called by the protocol when a Payment request is received.
// In case of insolvency no cheque is issued and sent, safe against fraud
// No return value: no error = payment is opportunistic = hang in till dropped
func (self *swap) send(n int) {
	if self.buys() && self.balance < 0 {
		amount := big.NewInt(int64(-self.balance))
		amount.Mul(amount, self.remote.SellAt)
		ch, err := self.chequebook.Issue(self.remote.Beneficiary, amount)
		if err != nil {
			glog.V(logger.Warn).Infof("[BZZ] cannot issue cheque. Contract: %v, Beneficiary: %v, Amount: %v", self.local.Contract, self.remote.Beneficiary, amount)
		} else {
			self.proto.payment(&paymentMsgData{-self.balance, ch})
			self.add(-self.balance)
		}
	}
}

// receive(units, cheque) is called by the protocol when a payment msg is received
// returns error if cheque is invalid.
func (self *swap) receive(units int, ch *chequebook.Cheque) error {
	if units <= 0 {
		return fmt.Errorf("invalid amount: %v <= 0", units)
	}

	// it could be easier to simply receive the price offer here, and negotiate
	// only chequebook address at handshake
	sum := new(big.Int).SetInt64(int64(units))
	sum.Mul(sum, self.local.SellAt)
	if sum.Cmp(ch.Amount) != 0 {
		return fmt.Errorf("invalid amount: %v (sent in msg) != %v (signed in cheque)", units)
	}

	if _, err := self.chequebox.Receive(ch); err != nil {
		return fmt.Errorf("invalid cheque: %v", err)
	}
	self.sub(units)
	return nil
}

// stop() causes autocash loop to terminate.
// Called after protocol handle loop terminates.
func (self *swap) stop() {
	self.chequebox.Stop()
}
