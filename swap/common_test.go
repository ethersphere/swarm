package swap

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	contractFactory "github.com/ethersphere/go-sw3/contracts-v0-2-0/simpleswapfactory"
	contract "github.com/ethersphere/swarm/contracts/swap"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/swap/chain"
	mock "github.com/ethersphere/swarm/swap/chain/mock"
	"github.com/ethersphere/swarm/uint256"
)

// swapTestBackend encapsulates the SimulatedBackend and can offer
// additional properties for the tests
type swapTestBackend struct {
	*mock.TestBackend
	factoryAddress common.Address // address of the SimpleSwapFactory in the simulated network
	tokenAddress   common.Address // address of the token in the simulated network
}

var defaultBackend = backends.NewSimulatedBackend(core.GenesisAlloc{
	ownerAddress:       {Balance: big.NewInt(1000000000000000000)},
	beneficiaryAddress: {Balance: big.NewInt(1000000000000000000)},
}, 8000000)

// newTestBackend creates a new test backend instance
func newTestBackend(t *testing.T) *swapTestBackend {
	t.Helper()

	backend := mock.NewTestBackend(defaultBackend)
	// deploy the ERC20-contract
	// ignore receipt because if there is no error, we can assume everything is fine on a simulated backend
	tokenAddress, _, _, err := contractFactory.DeployERC20Mintable(bind.NewKeyedTransactor(ownerKey), backend)
	if err != nil {
		t.Fatal(err)
	}

	// deploy a SimpleSwapFactoy
	// ignore receipt because if there is no error, we can assume everything is fine on a simulated backend
	// ignore factory instance, because the address is all we need at this point
	factoryAddress, _, _, err := contractFactory.DeploySimpleSwapFactory(bind.NewKeyedTransactor(ownerKey), backend, tokenAddress)
	if err != nil {
		t.Fatal(err)
	}

	return &swapTestBackend{
		TestBackend:    backend,
		factoryAddress: factoryAddress,
		tokenAddress:   tokenAddress,
	}
}

// newDefaultParams creates a set of default params for tests
func newDefaultParams(t *testing.T) *Params {
	t.Helper()
	baseKey := make([]byte, 32)
	_, err := rand.Read(baseKey)
	if err != nil {
		t.Fatal(err)
	}
	return &Params{
		BaseAddrs:           network.NewBzzAddr(baseKey, nil),
		LogPath:             "",
		PaymentThreshold:    int64(DefaultPaymentThreshold),
		DisconnectThreshold: int64(DefaultDisconnectThreshold),
	}
}

// newBaseTestSwapWithParams creates a swap with the given params
func newBaseTestSwapWithParams(t *testing.T, key *ecdsa.PrivateKey, params *Params, backend *swapTestBackend) (*Swap, string) {
	t.Helper()
	dir, err := ioutil.TempDir("", "swap_test_store")
	if err != nil {
		t.Fatal(err)
	}
	stateStore, err := state.NewDBStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("creating simulated backend")
	owner := createOwner(key)
	swapLog = newSwapLogger(params.LogPath, params.BaseAddrs)
	factory, err := cswap.FactoryAt(backend.factoryAddress, backend)
	if err != nil {
		t.Fatal(err)
	}

	txqueue := chain.NewTxQueue(stateStore, "chain", &chain.DefaultTxSchedulerBackend{
		Backend: backend,
	}, owner.privateKey)
	swap := newSwapInstance(stateStore, owner, backend, 10, params, factory, txqueue)
	return swap, dir
}

// create a test swap account with a backend
// creates a stateStore for persistence and a Swap account
func newBaseTestSwap(t *testing.T, key *ecdsa.PrivateKey, backend *swapTestBackend) (*Swap, string) {
	params := newDefaultParams(t)
	return newBaseTestSwapWithParams(t, key, params, backend)
}

// create a test swap account with a backend
// creates a stateStore for persistence and a Swap account
// returns a cleanup function
func newTestSwap(t *testing.T, key *ecdsa.PrivateKey, backend *swapTestBackend) (*Swap, func()) {
	t.Helper()
	usedBackend := backend
	if backend == nil {
		usedBackend = newTestBackend(t)
	}
	swap, dir := newBaseTestSwap(t, key, usedBackend)
	swap.txScheduler.Start()
	clean := func() {
		swap.Close()
		// only close if created by newTestSwap to avoid double close
		if backend != nil {
			backend.Close()
		}
		os.RemoveAll(dir)
	}
	return swap, clean
}

type dummyPeer struct {
	*protocols.Peer
}

// creates a dummy protocols.Peer with dummy MsgReadWriter
func newDummyPeerWithSpec(spec *protocols.Spec) *dummyPeer {
	id := adapters.RandomNodeConfig().ID
	rw := &dummyMsgRW{}
	protoPeer := protocols.NewPeer(p2p.NewPeer(id, "testPeer", nil), rw, spec)
	dummy := &dummyPeer{
		Peer: protoPeer,
	}
	return dummy
}

// creates a dummy protocols.Peer with dummy MsgReadWriter
func newDummyPeer() *dummyPeer {
	return newDummyPeerWithSpec(nil)
}

// creates cheque structure for testing
func newTestCheque() *Cheque {
	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         testChequeContract,
			CumulativePayout: uint256.FromUint64(42),
			Beneficiary:      beneficiaryAddress,
		},
		Honey: uint64(42),
	}

	return cheque
}

func newSignedTestCheque(testChequeContract common.Address, beneficiaryAddress common.Address, cumulativePayout *uint256.Uint256, signingKey *ecdsa.PrivateKey) (*Cheque, error) {
	cp := cumulativePayout.Value()
	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         testChequeContract,
			CumulativePayout: cumulativePayout,
			Beneficiary:      beneficiaryAddress,
		},
		Honey: (&cp).Uint64(),
	}

	sig, err := cheque.Sign(signingKey)
	if err != nil {
		return nil, err
	}
	cheque.Signature = sig
	return cheque, nil
}

// creates a randomized cheque structure for testing
func newRandomTestCheque() *Cheque {
	amount := uint64(mrand.Intn(100))

	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         testChequeContract,
			CumulativePayout: uint256.FromUint64(amount),
			Beneficiary:      beneficiaryAddress,
		},
		Honey: amount,
	}

	return cheque
}

// deploy for testing (needs simulated backend commit)
func testDeployWithPrivateKey(ctx context.Context, backend chain.Backend, privateKey *ecdsa.PrivateKey, ownerAddress common.Address, depositAmount *uint256.Uint256) (cswap.Contract, error) {
	opts := bind.NewKeyedTransactor(privateKey)
	opts.Context = ctx

	var stb *swapTestBackend
	var ok bool
	if stb, ok = backend.(*swapTestBackend); !ok {
		return nil, errors.New("not the expected test backend")
	}

	factory, err := cswap.FactoryAt(stb.factoryAddress, stb)
	if err != nil {
		return nil, err
	}

	contract, err := factory.DeploySimpleSwap(opts, ownerAddress, big.NewInt(int64(defaultHarddepositTimeoutDuration)))
	if err != nil {
		return nil, err
	}

	// send money into the new chequebook
	token, err := contractFactory.NewERC20Mintable(stb.tokenAddress, stb)
	if err != nil {
		return nil, err
	}

	deposit := depositAmount.Value()
	tx, err := token.Mint(bind.NewKeyedTransactor(ownerKey), contract.ContractParams().ContractAddress, &deposit)
	if err != nil {
		return nil, err
	}

	receipt, err := chain.WaitMined(ctx, stb, tx.Hash())
	if err != nil {
		return nil, err
	}

	if receipt.Status != 1 {
		return nil, errors.New("token transfer reverted")
	}

	return contract, nil
}

// deploy for testing (needs simulated backend commit)
func testDeploy(ctx context.Context, swap *Swap, depositAmount *uint256.Uint256) (err error) {
	swap.contract, err = testDeployWithPrivateKey(ctx, swap.backend, swap.owner.privateKey, swap.owner.address, depositAmount)
	return err
}

// newTestSwapAndPeer is a helper function to create a swap and a peer instance that fit together
// the owner of this swap is the beneficiaryAddress
// hence the owner of this swap would sign cheques with beneficiaryKey and receive cheques from ownerKey (or another party) which is NOT the owner of this swap
func newTestSwapAndPeer(t *testing.T, key *ecdsa.PrivateKey) (*Swap, *Peer, func()) {
	swap, clean := newTestSwap(t, key, nil)
	// owner address is the beneficiary (counterparty) for the peer
	// that's because we expect cheques we receive to be signed by the address we would issue cheques to
	peer, err := swap.addPeer(newDummyPeer().Peer, ownerAddress, testChequeContract)
	if err != nil {
		t.Fatal(err)
	}
	// we need to adjust the owner address on swap because we will issue cheques to beneficiaryAddress
	swap.owner.address = beneficiaryAddress
	return swap, peer, clean
}

// dummyMsgRW implements MessageReader and MessageWriter
// but doesn't do anything. Useful for dummy message sends
type dummyMsgRW struct{}

// ReadMsg is from the MessageReader interface
func (d *dummyMsgRW) ReadMsg() (p2p.Msg, error) {
	return p2p.Msg{}, nil
}

// WriteMsg is from the MessageWriter interface
func (d *dummyMsgRW) WriteMsg(msg p2p.Msg) error {
	return nil
}

// struct used by the testCashoutResultHandler
type cashChequeDoneData struct {
	request *CashoutRequest
	result  *contract.CashChequeResult
	receipt *types.Receipt
}

// testCashoutResultHandler is a CashoutResultHandler which writes to a channel after forwarding the result to swap
type testCashoutResultHandler struct {
	swap           *Swap
	cashChequeDone chan cashChequeDoneData
}

func newTestCashoutResultHandler(swap *Swap) *testCashoutResultHandler {
	return &testCashoutResultHandler{
		swap:           swap,
		cashChequeDone: make(chan cashChequeDoneData),
	}
}

// HandleCashoutResult forwards the result to swap if set and afterwards sends it to its channel
func (h *testCashoutResultHandler) HandleCashoutResult(request *CashoutRequest, result *contract.CashChequeResult, receipt *types.Receipt) error {
	if h.swap != nil {
		if err := h.swap.HandleCashoutResult(request, result, receipt); err != nil {
			return err
		}
	}
	h.cashChequeDone <- cashChequeDoneData{
		request: request,
		result:  result,
		receipt: receipt,
	}
	return nil
}

// helper function to override the cashoutHandler for a cashout processor of swap instance
func overrideCashoutResultHandler(swap *Swap) *testCashoutResultHandler {
	cashoutResultHandler := newTestCashoutResultHandler(swap)
	swap.cashoutProcessor.cashoutResultHandler = cashoutResultHandler
	return cashoutResultHandler
}
