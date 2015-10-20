package bzz

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/chequebook"
	"github.com/ethereum/go-ethereum/common/swap"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
)

// SwAP Swarm Accounting Protocol with
//      Swift Automatic  Payments
// using chequebook pkg for delayed payments
// default parameters

var (
	autoCashInterval     = 300 * time.Second       // default interval for autocash
	autoCashThreshold    = big.NewInt(10000000000) // threshold that triggers autocash (wei)
	autoDepositInterval  = 300 * time.Second       // default interval for autocash
	autoDepositThreshold = big.NewInt(10000000000) // threshold that triggers autodeposit (wei)
	autoDepositBuffer    = big.NewInt(20000000000) // buffer that is surplus for fork protection etc (wei)
	buyAt                = big.NewInt(20000000)    // maximum chunk price host is willing to pay (wei)
	sellAt               = big.NewInt(20000000)    // minimum chunk price host requires (wei)
	payAt                = 100                     // threshold that triggers payment request (units)
	dropAt               = 150                     // threshold that triggers disconnect (units)
)

type swapParams struct {
	*swap.Params
	*payProfile
}

type swapProfile struct {
	*swap.Profile
	*payProfile
}

type payProfile struct {
	PublicKey   string                 // check againsst signature of promise
	Contract    common.Address         // address of chequebook contract
	Beneficiary common.Address         // recipient address for swarm sales revenue
	privateKey  *ecdsa.PrivateKey      `json:"-"`
	publicKey   *ecdsa.PublicKey       `json:"-"`
	chequebook  *chequebook.Chequebook `json:"-"`
}

func defaultSwapParams(contract common.Address, prvkey *ecdsa.PrivateKey) *swapParams {
	pubkey := &prvkey.PublicKey
	return &swapParams{
		payProfile: &payProfile{
			PublicKey:   common.ToHex(crypto.FromECDSAPub(pubkey)),
			Contract:    contract,
			Beneficiary: crypto.PubkeyToAddress(*pubkey),
			privateKey:  prvkey,
			publicKey:   pubkey,
		},
		Params: &swap.Params{
			Profile: &swap.Profile{
				BuyAt:  buyAt,
				SellAt: sellAt,
				PayAt:  uint(payAt),
				DropAt: uint(dropAt),
			},
			Strategy: &swap.Strategy{
				AutoCashInterval:     autoCashInterval,
				AutoCashThreshold:    autoCashThreshold,
				AutoDepositInterval:  autoDepositInterval,
				AutoDepositThreshold: autoDepositThreshold,
				AutoDepositBuffer:    autoDepositBuffer,
			},
		},
	}
}

// swap constructor, parameters
// * global chequebook, assumed deployed service and
// * the balance is at buffer.
// swap.Add(n) called in netstore
// n > 0 called when sending chunks = receiving retrieve requests
//                 OR sending cheques.
// n < 0  called when receiving chunks = receiving delivery responses
//                 OR receiving cheques.

func newSwap(local *swapParams, remote *swapProfile, proto swap.Protocol) (self *swap.Swap, err error) {

	out := chequebook.NewOutbox(local.chequebook, remote.Beneficiary)

	in, err := chequebook.NewInbox(remote.Contract, local.Beneficiary, crypto.ToECDSAPub(common.FromHex(remote.PublicKey)), local.chequebook.Backend())
	if err != nil {
		return
	}

	self, err = swap.New(local.Params, out, in, proto)
	if err != nil {
		return
	}
	// remote profile given (first) in handshake
	self.SetRemote(remote.Profile)

	glog.V(logger.Info).Infof("[BZZ] SWAP auto cash ON for %v -> %v: interval = %v, threshold = %v, peer = %v)", local.Contract.Hex()[:8], local.Beneficiary.Hex()[:8], local.AutoCashInterval, local.AutoCashThreshold, proto)

	return
}

// setChequebook(path, backend) wraps the
// chequebook initialiser and sets up autoDeposit to cover spending.
func (self *swapParams) setChequebook(path string, backend chequebook.Backend) (err error) {
	hexkey := common.Bytes2Hex(self.Contract.Bytes())
	err = os.MkdirAll(filepath.Join(path, "chequebooks"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory for chequebooks: %v", err)
	}
	chbookpath := filepath.Join(path, "chequebooks", hexkey+".json")
	self.chequebook, err = chequebook.LoadChequebook(chbookpath, self.privateKey, backend)

	if err != nil {
		self.chequebook, err = chequebook.NewChequebook(chbookpath, self.Contract, self.privateKey, backend)
		if err != nil {
			return
		}
	}
	self.chequebook.AutoDeposit(self.AutoDepositInterval, self.AutoDepositThreshold, self.AutoDepositBuffer)
	glog.V(logger.Info).Infof("[BZZ] SWAP auto deposit ON for %v -> %v: interval = %v, threshold = %v, buffer = %v)", self.Beneficiary.Hex()[:8], self.Contract.Hex()[:8], self.AutoDepositInterval, self.AutoDepositThreshold, self.AutoDepositBuffer)

	return
}
