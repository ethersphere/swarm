// Copyright 2019 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package swap

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ethersphere/swarm/network"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	contractFactory "github.com/ethersphere/go-sw3/contracts-v0-2-0/simpleswapfactory"
	"github.com/ethersphere/swarm/contracts/swap"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/testutil"
)

var (
	loglevel           = flag.Int("logleveld", 2, "verbosity of debug logs")
	ownerKey, _        = crypto.HexToECDSA("634fb5a872396d9693e5c9f9d7233cfa93f395c093371017ff44aa9ae6564cdd")
	ownerAddress       = crypto.PubkeyToAddress(ownerKey.PublicKey)
	beneficiaryKey, _  = crypto.HexToECDSA("6f05b0a29723ca69b1fc65d11752cee22c200cf3d2938e670547f7ae525be112")
	beneficiaryAddress = crypto.PubkeyToAddress(beneficiaryKey.PublicKey)
	testChequeSig      = common.Hex2Bytes("a53e7308bb5590b45cabf44538508ccf1760b53eea721dd50bfdd044547e38b412142da9f3c690a940d6ee390d3f365a38df02b2688cea17f303f6de01268c2e1c")
	testChequeContract = common.HexToAddress("0x4405415b2B8c9F9aA83E151637B8378dD3bcfEDD") // second contract created by ownerKey
)

// booking represents an accounting movement in relation to a particular node: `peer`
// if `amount` is positive, it means the node which adds this booking will be credited in respect to `peer`
// otherwise it will be debited
type booking struct {
	amount int64
	peer   *protocols.Peer
}

// swapTestBackend encapsulates the SimulatedBackend and can offer
// additional properties for the tests
type swapTestBackend struct {
	*backends.SimulatedBackend
	factoryAddress common.Address // address of the SimpleSwapFactory in the simulated network
	tokenAddress   common.Address // address of the token in the simulated network
	// the async cashing go routine needs synchronization for tests
	cashDone chan struct{}
}

func (b *swapTestBackend) Close() error {
	// Do not close SimulatedBackend as it is a global instance.
	return nil
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	// Close the global default backend
	// when tests are done.
	defaultBackend.Close()
	os.Exit(exitCode)
}

func init() {
	testutil.Init()
	mrand.Seed(time.Now().UnixNano())
	swapLog = log.Root()
}

var defaultBackend = backends.NewSimulatedBackend(core.GenesisAlloc{
	ownerAddress:       {Balance: big.NewInt(1000000000000000000)},
	beneficiaryAddress: {Balance: big.NewInt(1000000000000000000)},
}, 8000000)

// newTestBackend creates a new test backend instance
func newTestBackend(t *testing.T) *swapTestBackend {
	t.Helper()
	// commit the initial "pre-mined" accounts (issuer and beneficiary addresses)
	defaultBackend.Commit()

	// deploy the ERC20-contract
	// ignore receipt because if there is no error, we can assume everything is fine on a simulated backend
	tokenAddress, _, _, err := contractFactory.DeployERC20Mintable(bind.NewKeyedTransactor(ownerKey), defaultBackend)
	if err != nil {
		t.Fatal(err)
	}
	defaultBackend.Commit()

	// deploy a SimpleSwapFactoy
	// ignore receipt because if there is no error, we can assume everything is fine on a simulated backend
	// ignore factory instance, because the address is all we need at this point
	factoryAddress, _, _, err := contractFactory.DeploySimpleSwapFactory(bind.NewKeyedTransactor(ownerKey), defaultBackend, tokenAddress)
	if err != nil {
		t.Fatal(err)
	}
	defaultBackend.Commit()

	return &swapTestBackend{
		SimulatedBackend: defaultBackend,
		factoryAddress:   factoryAddress,
		tokenAddress:     tokenAddress,
	}
}

type storeKeysTestCase struct {
	nodeID                    enode.ID
	expectedBalanceKey        string
	expectedSentChequeKey     string
	expectedReceivedChequeKey string
	expectedPendingChequeKey  string
	expectedUsedChequebookKey string
}

// Test the getting balance and cheques store keys based on a node ID, and the reverse process as well
func TestStoreKeys(t *testing.T) {
	testCases := []storeKeysTestCase{
		{enode.HexID("f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56"), "balance_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "sent_cheque_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "received_cheque_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "pending_cheque_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "connected_chequebook"},
		{enode.HexID("93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c"), "balance_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "sent_cheque_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "received_cheque_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "pending_cheque_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "connected_chequebook"},
		{enode.HexID("c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44"), "balance_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "sent_cheque_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "received_cheque_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "pending_cheque_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "connected_chequebook"},
	}
	testStoreKeys(t, testCases)
}

func testStoreKeys(t *testing.T, testCases []storeKeysTestCase) {
	for _, testCase := range testCases {
		t.Run(fmt.Sprint(testCase.nodeID), func(t *testing.T) {
			actualBalanceKey := balanceKey(testCase.nodeID)
			actualSentChequeKey := sentChequeKey(testCase.nodeID)
			actualReceivedChequeKey := receivedChequeKey(testCase.nodeID)
			actualPendingChequeKey := pendingChequeKey(testCase.nodeID)
			actualUsedChequebookKey := connectedChequebookKey

			if actualBalanceKey != testCase.expectedBalanceKey {
				t.Fatalf("Expected balance key to be %s, but is %s instead.", testCase.expectedBalanceKey, actualBalanceKey)
			}
			if actualSentChequeKey != testCase.expectedSentChequeKey {
				t.Fatalf("Expected sent cheque key to be %s, but is %s instead.", testCase.expectedSentChequeKey, actualSentChequeKey)
			}
			if actualReceivedChequeKey != testCase.expectedReceivedChequeKey {
				t.Fatalf("Expected received cheque key to be %s, but is %s instead.", testCase.expectedReceivedChequeKey, actualReceivedChequeKey)
			}

			if actualPendingChequeKey != testCase.expectedPendingChequeKey {
				t.Fatalf("Expected pending cheque key to be %s, but is %s instead.", testCase.expectedPendingChequeKey, actualPendingChequeKey)
			}

			if actualUsedChequebookKey != testCase.expectedUsedChequebookKey {
				t.Fatalf("Expected used chequebook key to be %s, but is %s instead.", testCase.expectedUsedChequebookKey, actualUsedChequebookKey)
			}

			nodeID := keyToID(actualBalanceKey, balancePrefix)
			if nodeID != testCase.nodeID {
				t.Fatalf("Expected node ID to be %v, but is %v instead.", testCase.nodeID, nodeID)
			}
			nodeID = keyToID(actualSentChequeKey, sentChequePrefix)
			if nodeID != testCase.nodeID {
				t.Fatalf("Expected node ID to be %v, but is %v instead.", testCase.nodeID, nodeID)
			}
			nodeID = keyToID(actualReceivedChequeKey, receivedChequePrefix)
			if nodeID != testCase.nodeID {
				t.Fatalf("Expected node ID to be %v, but is %v instead.", testCase.nodeID, nodeID)
			}
		})
	}
}

// Test the correct storing of peer balances through the store after node balance updates
func TestStoreBalances(t *testing.T) {
	// create a test swap account
	s, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// modify balances both in memory and in store
	testPeer, err := s.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	testPeerID := testPeer.ID()
	peerBalance := int64(29)
	if err := testPeer.setBalance(peerBalance); err != nil {
		t.Fatal(err)
	}
	// store balance for peer should match
	comparePeerBalance(t, s, testPeerID, peerBalance)

	// update balances for second peer
	testPeer2, err := s.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	testPeer2ID := testPeer2.ID()
	peer2Balance := int64(-76)

	if err := testPeer2.setBalance(peer2Balance); err != nil {
		t.Fatal(err)
	}
	// store balance for each peer should match
	comparePeerBalance(t, s, testPeerID, peerBalance)
	comparePeerBalance(t, s, testPeer2ID, peer2Balance)
}

func comparePeerBalance(t *testing.T, s *Swap, peer enode.ID, expectedPeerBalance int64) {
	t.Helper()
	var peerBalance int64
	err := s.store.Get(balanceKey(peer), &peerBalance)
	if err != nil && err != state.ErrNotFound {
		t.Error("Unexpected peer balance retrieval failure.")
	}
	if peerBalance != expectedPeerBalance {
		t.Errorf("Expected peer store balance to be %d, but is %d instead.", expectedPeerBalance, peerBalance)
	}
}

// Test that repeated bookings do correct accounting
func TestRepeatedBookings(t *testing.T) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	var bookings []booking

	// credits to peer 1
	testPeer, err := swap.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	bookingAmount := int64(mrand.Intn(100))
	bookingQuantity := 1 + mrand.Intn(10)
	testPeerBookings(t, swap, &bookings, bookingAmount, bookingQuantity, testPeer.Peer)

	// debits to peer 2
	testPeer2, err := swap.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	bookingAmount = 0 - int64(mrand.Intn(100))
	bookingQuantity = 1 + mrand.Intn(10)
	testPeerBookings(t, swap, &bookings, bookingAmount, bookingQuantity, testPeer2.Peer)

	// credits and debits to peer 2
	mixedBookings := []booking{
		{int64(mrand.Intn(100)), testPeer2.Peer},
		{int64(0 - mrand.Intn(55)), testPeer2.Peer},
		{int64(0 - mrand.Intn(999)), testPeer2.Peer},
	}
	addBookings(swap, mixedBookings)
	verifyBookings(t, swap, append(bookings, mixedBookings...))
}

//TestNewSwapFailure attempts to initialze SWAP with (a combination of) parameters which are not allowed. The test checks whether there are indeed failures
func TestNewSwapFailure(t *testing.T) {
	testBackend := newTestBackend(t)
	defer testBackend.Close()
	dir, err := ioutil.TempDir("", "swarmSwap")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// a simple rpc endpoint for testing dialing
	ipcEndpoint := path.Join(dir, "TestSwarmSwap.ipc")

	// windows namedpipes are not on filesystem but on NPFS
	if runtime.GOOS == "windows" {
		b := make([]byte, 8)
		rand.Read(b)
		ipcEndpoint = `\\.\pipe\TestSwarm-` + hex.EncodeToString(b)
	}

	_, server, err := rpc.StartIPCEndpoint(ipcEndpoint, nil)
	if err != nil {
		t.Error(err)
	}
	defer server.Stop()

	prvKey, err := crypto.GenerateKey()
	if err != nil {
		t.Error(err)
	}

	params := newDefaultParams(t)
	chequebookAddress := testChequeContract
	Deposit := uint64(1)

	type testSwapConfig struct {
		dbPath            string
		prvkey            *ecdsa.PrivateKey
		backendURL        string
		params            *Params
		chequebookAddress common.Address
		skipDeposit       bool
		deposit           uint64
		factoryAddress    common.Address
	}

	var config testSwapConfig

	for _, tc := range []struct {
		name      string
		configure func(*testSwapConfig)
		check     func(*testing.T, *testSwapConfig)
	}{
		{
			name: "no backedURL",
			configure: func(config *testSwapConfig) {
				config.dbPath = dir
				config.prvkey = prvKey
				config.backendURL = ""
				config.params = params
				config.chequebookAddress = chequebookAddress
				config.deposit = Deposit
				config.factoryAddress = testBackend.factoryAddress
			},
			check: func(t *testing.T, config *testSwapConfig) {
				_, err := New(
					config.dbPath,
					config.prvkey,
					config.backendURL,
					config.params,
					config.chequebookAddress,
					config.skipDeposit,
					config.deposit,
					config.factoryAddress,
				)
				if !strings.Contains(err.Error(), "no backend URL given") {
					t.Fatal("no backendURL, but created SWAP")
				}
			},
		},
		{
			name: "disconnect threshold lower than payment threshold",
			configure: func(config *testSwapConfig) {
				config.dbPath = dir
				config.prvkey = prvKey
				config.backendURL = ipcEndpoint
				params.PaymentThreshold = params.DisconnectThreshold + 1
				config.factoryAddress = testBackend.factoryAddress
			},
			check: func(t *testing.T, config *testSwapConfig) {
				_, err := New(
					config.dbPath,
					config.prvkey,
					config.backendURL,
					config.params,
					config.chequebookAddress,
					config.skipDeposit,
					config.deposit,
					config.factoryAddress,
				)
				if !strings.Contains(err.Error(), "disconnect threshold lower or at payment threshold") {
					t.Fatal("disconnect threshold lower than payment threshold, but created SWAP", err.Error())
				}
			},
		},
		{
			name: "no deposit and given deposit amount",
			configure: func(config *testSwapConfig) {
				config.params = newDefaultParams(t)
				config.chequebookAddress = chequebookAddress
				config.skipDeposit = true
				config.deposit = Deposit
				config.factoryAddress = testBackend.factoryAddress
			},
			check: func(t *testing.T, config *testSwapConfig) {
				defer os.RemoveAll(config.dbPath)
				_, err := New(
					config.dbPath,
					config.prvkey,
					config.backendURL,
					config.params,
					config.chequebookAddress,
					config.skipDeposit,
					config.deposit,
					config.factoryAddress,
				)
				if !strings.Contains(err.Error(), ErrSkipDeposit.Error()) {
					t.Fatal("skipDeposit true and non-zero depositAmount, but created SWAP", err)
				}
			},
		},
		{
			name: "invalid backendURL",
			configure: func(config *testSwapConfig) {
				config.prvkey = prvKey
				config.backendURL = "invalid backendURL"
				params.PaymentThreshold = int64(DefaultPaymentThreshold)
				config.skipDeposit = false
				config.factoryAddress = testBackend.factoryAddress
			},
			check: func(t *testing.T, config *testSwapConfig) {
				defer os.RemoveAll(config.dbPath)
				_, err := New(
					config.dbPath,
					config.prvkey,
					config.backendURL,
					config.params,
					config.chequebookAddress,
					config.skipDeposit,
					config.deposit,
					config.factoryAddress,
				)
				if !strings.Contains(err.Error(), "error connecting to Ethereum API") {
					t.Fatal("invalid backendURL, but created SWAP", err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "swarmSwap")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			config.dbPath = dir

			logDir, err := ioutil.TempDir("", "swap_test_log")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(logDir)

			tc.configure(&config)
			if tc.check != nil {
				tc.check(t, &config)
			}
		})

	}
}

func TestStartChequebookFailure(t *testing.T) {
	type chequebookConfig struct {
		passIn        common.Address
		expectedError error
		testBackend   *swapTestBackend
	}

	var config chequebookConfig

	for _, tc := range []struct {
		name      string
		configure func(*chequebookConfig)
		check     func(*testing.T, *chequebookConfig)
	}{
		{
			name: "with pass in and save",
			configure: func(config *chequebookConfig) {
				config.passIn = testChequeContract
				config.expectedError = fmt.Errorf("Attempting to connect to provided chequebook, but different chequebook used before")
			},
			check: func(t *testing.T, config *chequebookConfig) {
				// create SWAP
				swap, clean := newTestSwap(t, ownerKey, config.testBackend)
				defer clean()
				// deploy a chequebook
				err := testDeploy(context.TODO(), swap, big.NewInt(0))
				if err != nil {
					t.Fatal(err)
				}
				// save chequebook on SWAP
				err = swap.saveChequebook(swap.GetParams().ContractAddress)
				if err != nil {
					t.Fatal(err)
				}
				// try to connect with a different address
				_, err = swap.StartChequebook(config.passIn)
				if err.Error() != config.expectedError.Error() {
					t.Fatal(fmt.Errorf("Expected error not equal to actual error. Expected: %v. Actual: %v", config.expectedError, err))
				}
			},
		},
		{
			name: "with wrong pass in",
			configure: func(config *chequebookConfig) {
				config.passIn = common.HexToAddress("0x4405415b2B8c9F9aA83E151637B8370000000000") // address without deployed chequebook
				config.expectedError = fmt.Errorf("contract validation for %v failed: %v", config.passIn.Hex(), swap.ErrNotDeployedByFactory)
			},
			check: func(t *testing.T, config *chequebookConfig) {
				// create SWAP
				swap, clean := newTestSwap(t, ownerKey, config.testBackend)
				defer clean()
				// try to connect with an address not containing a chequebook instance
				_, err := swap.StartChequebook(config.passIn)
				if err.Error() != config.expectedError.Error() {
					t.Fatal(fmt.Errorf("Expected error not equal to actual error. Expected: %v. Actual: %v", config.expectedError, err))
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testBackend := newTestBackend(t)
			defer testBackend.Close()
			tc.configure(&config)
			config.testBackend = testBackend
			if tc.check != nil {
				tc.check(t, &config)
			}
		})
	}
}

func TestStartChequebookSuccess(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *swapTestBackend)
	}{
		{
			name: "with same pass in as previously used",
			check: func(t *testing.T, testBackend *swapTestBackend) {
				// create SWAP
				swap, clean := newTestSwap(t, ownerKey, testBackend)
				defer clean()

				// deploy a chequebook
				err := testDeploy(context.TODO(), swap, big.NewInt(0))
				if err != nil {
					t.Fatal(err)
				}

				// save chequebook on SWAP
				err = swap.saveChequebook(swap.GetParams().ContractAddress)
				if err != nil {
					t.Fatal(err)
				}

				// start chequebook with same pass in as deployed
				_, err = swap.StartChequebook(swap.GetParams().ContractAddress)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "with correct pass in",
			check: func(t *testing.T, testBackend *swapTestBackend) {
				// create SWAP
				swap, clean := newTestSwap(t, ownerKey, testBackend)
				defer clean()

				// deploy a chequebook
				err := testDeploy(context.TODO(), swap, big.NewInt(0))
				if err != nil {
					t.Fatal(err)
				}

				// start chequebook with same pass in as deployed
				_, err = swap.StartChequebook(swap.GetParams().ContractAddress)
				if err != nil {
					t.Fatal(err)
				}

				// err should be nil
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testBackend := newTestBackend(t)
			defer testBackend.Close()
			if tc.check != nil {
				tc.check(t, testBackend)
			}
		})
	}
}

//TestDisconnectThreshold tests that the disconnect threshold is reached when adding the DefaultDisconnectThreshold amount to the peers balance
func TestDisconnectThreshold(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()
	testDeploy(context.Background(), swap, big.NewInt(0))

	testPeer := newDummyPeer()
	swap.addPeer(testPeer.Peer, swap.owner.address, swap.GetParams().ContractAddress)

	// leave balance exactly at disconnect threshold
	swap.Add(int64(DefaultDisconnectThreshold), testPeer.Peer)
	// account for traffic which increases debt
	err := swap.Add(1, testPeer.Peer)
	if err == nil {
		t.Fatal("expected accounting operation to fail, but it didn't")
	}
	if !strings.Contains(err.Error(), "disconnect threshold") {
		t.Fatal(err)
	}
	// account for traffic which reduces debt, which should be allowed even when over the threshold
	err = swap.Add(-1, testPeer.Peer)
	if err != nil {
		t.Fatalf("expected accounting operation to succeed, but it failed with %v", err)
	}
}

//TestPaymentThreshold tests that the payment threshold is reached when subtracting the DefaultPaymentThreshold amount from the peers balance
func TestPaymentThreshold(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()
	testDeploy(context.Background(), swap, big.NewInt(int64(DefaultPaymentThreshold)))
	testPeer := newDummyPeerWithSpec(Spec)
	swap.addPeer(testPeer.Peer, swap.owner.address, swap.GetParams().ContractAddress)
	if err := swap.Add(-int64(DefaultPaymentThreshold), testPeer.Peer); err != nil {
		t.Fatal()
	}

	var cheque *Cheque
	_ = swap.store.Get(pendingChequeKey(testPeer.Peer.ID()), &cheque)
	if cheque.CumulativePayout != DefaultPaymentThreshold {
		t.Fatal()
	}
}

// TestResetBalance tests that balances are correctly reset
// The test deploys creates swap instances for each node,
// deploys simulated contracts, sets the balance of each
// other node to some arbitrary number above thresholds,
// and then calls both `sendCheque` on one side and
// `handleEmitChequeMsg` in order to simulate a roundtrip
// and see that both have reset the balance correctly
func TestResetBalance(t *testing.T) {
	testBackend := newTestBackend(t)
	defer testBackend.Close()
	// create both test swap accounts
	creditorSwap, clean1 := newTestSwap(t, beneficiaryKey, testBackend)
	debitorSwap, clean2 := newTestSwap(t, ownerKey, testBackend)
	defer clean1()
	defer clean2()

	testAmount := int64(DefaultPaymentThreshold + 42)

	ctx := context.Background()
	err := testDeploy(ctx, creditorSwap, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	err = testDeploy(ctx, debitorSwap, big.NewInt(testAmount))
	if err != nil {
		t.Fatal(err)
	}

	// create Peer instances
	// NOTE: remember that these are peer instances representing each **a model of the remote peer** for every local node
	// so creditor is the model of the remote mode for the debitor! (and vice versa)
	cPeer := newDummyPeerWithSpec(Spec)
	dPeer := newDummyPeerWithSpec(Spec)
	creditor, err := debitorSwap.addPeer(cPeer.Peer, creditorSwap.owner.address, debitorSwap.GetParams().ContractAddress)
	if err != nil {
		t.Fatal(err)
	}
	debitor, err := creditorSwap.addPeer(dPeer.Peer, debitorSwap.owner.address, debitorSwap.GetParams().ContractAddress)
	if err != nil {
		t.Fatal(err)
	}

	// set balances arbitrarily
	debitor.setBalance(testAmount)
	creditor.setBalance(-testAmount)

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// now simulate sending the cheque to the creditor from the debitor
	creditor.sendCheque()

	debitorSwap.handleConfirmChequeMsg(ctx, creditor, &ConfirmChequeMsg{
		Cheque: creditor.getPendingCheque(),
	})
	// the debitor should have already reset its balance
	if creditor.getBalance() != 0 {
		t.Fatalf("unexpected balance to be 0, but it is %d", creditor.getBalance())
	}

	// now load the cheque that the debitor created...
	cheque := creditor.getLastSentCheque()
	if cheque == nil {
		t.Fatal("expected to find a cheque, but it was empty")
	}
	// ...create a message...
	msg := &EmitChequeMsg{
		Cheque: cheque,
	}
	// now we need to create the channel...
	testBackend.cashDone = make(chan struct{})
	// ...and trigger message handling on the receiver side (creditor)
	// remember that debitor is the model of the remote node for the creditor...
	err = creditorSwap.handleEmitChequeMsg(ctx, debitor, msg)
	if err != nil {
		t.Fatal(err)
	}
	// ...on which we wait until the cashCheque is actually terminated (ensures proper nounce count)
	select {
	case <-testBackend.cashDone:
		log.Debug("cash transaction completed and committed")
	case <-time.After(4 * time.Second):
		t.Fatalf("Timeout waiting for cash transactions to complete")
	}
	// finally check that the creditor also successfully reset the balances
	if debitor.getBalance() != 0 {
		t.Fatalf("unexpected balance to be 0, but it is %d", debitor.getBalance())
	}
}

// generate bookings based on parameters, apply them to a Swap struct and verify the result
// append generated bookings to slice pointer
func testPeerBookings(t *testing.T, s *Swap, bookings *[]booking, bookingAmount int64, bookingQuantity int, peer *protocols.Peer) {
	t.Helper()
	peerBookings := generateBookings(bookingAmount, bookingQuantity, peer)
	*bookings = append(*bookings, peerBookings...)
	addBookings(s, peerBookings)
	verifyBookings(t, s, *bookings)
}

// generate as many bookings as specified by `quantity`, each one with the indicated `amount` and `peer`
func generateBookings(amount int64, quantity int, peer *protocols.Peer) (bookings []booking) {
	for i := 0; i < quantity; i++ {
		bookings = append(bookings, booking{amount, peer})
	}
	return
}

// take a Swap struct and a list of bookings, and call the accounting function for each of them
func addBookings(swap *Swap, bookings []booking) {
	for i := 0; i < len(bookings); i++ {
		booking := bookings[i]
		swap.Add(booking.amount, booking.peer)
	}
}

// take a Swap struct and a list of bookings, and verify the resulting balances are as expected
func verifyBookings(t *testing.T, s *Swap, bookings []booking) {
	t.Helper()
	expectedBalances := calculateExpectedBalances(s, bookings)
	realBalances, err := s.Balances()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedBalances, realBalances) {
		t.Fatalf("After %d bookings, expected balance to be %v, but is %v", len(bookings), stringifyBalance(expectedBalances), stringifyBalance(realBalances))
	}
}

// converts a balance map to a one-line string representation
func stringifyBalance(balance map[enode.ID]int64) string {
	marshaledBalance, err := json.Marshal(balance)
	if err != nil {
		return err.Error()
	}
	return string(marshaledBalance)
}

// take a swap struct and a list of bookings, and calculate the expected balances.
// the result is a map which stores the balance for all the peers present in the bookings,
// from the perspective of the node that loaded the Swap struct.
func calculateExpectedBalances(swap *Swap, bookings []booking) map[enode.ID]int64 {
	expectedBalances := make(map[enode.ID]int64)
	for i := 0; i < len(bookings); i++ {
		booking := bookings[i]
		peerID := booking.peer.ID()
		peerBalance := expectedBalances[peerID]
		// peer balance should only be affected if debt is being reduced or if balance is smaller than disconnect threshold
		if peerBalance < swap.params.DisconnectThreshold || booking.amount < 0 {
			peerBalance += booking.amount
		}
		expectedBalances[peerID] = peerBalance
	}
	return expectedBalances
}

// try restoring a balance from state store
// this is simulated by creating a node,
// assigning it an arbitrary balance,
// then closing the state store.
// Then we re-open the state store and check that
// the balance is still the same
func TestRestoreBalanceFromStateStore(t *testing.T) {
	testBackend := newTestBackend(t)
	defer testBackend.Close()

	// create a test swap account
	swap, testDir := newBaseTestSwap(t, ownerKey, testBackend)
	defer os.RemoveAll(testDir)

	testPeer, err := swap.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	testPeer.setBalance(-8888)

	tmpBalance := testPeer.getBalance()
	swap.store.Put(testPeer.ID().String(), &tmpBalance)

	err = swap.store.Close()
	if err != nil {
		t.Fatal(err)
	}
	swap.store = nil

	stateStore, err := state.NewDBStore(testDir)
	defer stateStore.Close()
	if err != nil {
		t.Fatal(err)
	}

	var newBalance int64
	stateStore.Get(testPeer.Peer.ID().String(), &newBalance)

	// compare the balances
	if tmpBalance != newBalance {
		t.Fatalf("Unexpected balance value after sending cheap message test. Expected balance: %d, balance is: %d", tmpBalance, newBalance)
	}
}

// During tests, because the cashing in of cheques is async, we should wait for the function to be returned
// Otherwise if we call `handleEmitChequeMsg` manually, it will return before the TX has been committed to the `SimulatedBackend`,
// causing subsequent TX to possibly fail due to nonce mismatch
func testCashCheque(s *Swap, otherSwap cswap.Contract, opts *bind.TransactOpts, cheque *Cheque) {
	cashCheque(s, otherSwap, opts, cheque)
	// send to the channel, signals to clients that this function actually finished
	if stb, ok := s.backend.(*swapTestBackend); ok {
		if stb.cashDone != nil {
			stb.cashDone <- struct{}{}
		}
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
	swap := newSwapInstance(stateStore, owner, backend, 10, params, factory)
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
func newDummyPeer() *dummyPeer {
	return newDummyPeerWithSpec(nil)
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

// creates cheque structure for testing
func newTestCheque() *Cheque {
	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         testChequeContract,
			CumulativePayout: uint64(42),
			Beneficiary:      beneficiaryAddress,
		},
		Honey: uint64(42),
	}

	return cheque
}

// creates a randomized cheque structure for testing
func newRandomTestCheque() *Cheque {
	amount := mrand.Intn(100)

	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         testChequeContract,
			CumulativePayout: uint64(amount),
			Beneficiary:      beneficiaryAddress,
		},
		Honey: uint64(amount),
	}

	return cheque
}

// tests if encodeForSignature encodes the cheque as expected
func TestChequeEncodeForSignature(t *testing.T) {
	expectedCheque := newTestCheque()

	// encode the cheque
	encoded := expectedCheque.encodeForSignature()
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("4405415b2b8c9f9aa83e151637b8378dd3bcfeddb8d424e9662fe0837fb1d728f1ac97cebb1085fe000000000000000000000000000000000000000000000000000000000000002a")
	if !bytes.Equal(encoded, expected) {
		t.Fatalf("Unexpected encoding of cheque. Expected encoding: %x, result is: %x", expected, encoded)
	}
}

// tests if sigHashCheque computes the correct hash to sign
func TestChequeSigHash(t *testing.T) {
	expectedCheque := newTestCheque()

	// compute the hash that will be signed
	hash := expectedCheque.sigHash()
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("354a78a181b24d0beb1606cd9f525e6068e8e5dd96747468c21f2ecc89cb0bad")
	if !bytes.Equal(hash, expected) {
		t.Fatalf("Unexpected sigHash of cheque. Expected: %x, result is: %x", expected, hash)
	}
}

// tests if signContent computes the correct signature
func TestSignContent(t *testing.T) {
	// setup test swap object
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	expectedCheque := newTestCheque()

	var err error

	// set the owner private key to a known key so we always get the same signature
	swap.owner.privateKey = ownerKey

	// sign the cheque
	sig, err := expectedCheque.Sign(swap.owner.privateKey)
	// expected value (computed through truffle/js)
	expected := testChequeSig
	if err != nil {
		t.Fatalf("Error in signing: %s", err)
	}
	if !bytes.Equal(sig, expected) {
		t.Fatalf("Unexpected signature for cheque. Expected: %x, result is: %x", expected, sig)
	}
}

// tests if verifyChequeSig accepts a correct signature
func TestVerifyChequeSig(t *testing.T) {
	expectedCheque := newTestCheque()
	expectedCheque.Signature = testChequeSig

	if err := expectedCheque.VerifySig(ownerAddress); err != nil {
		t.Fatalf("Invalid signature: %v", err)
	}
}

// tests if verifyChequeSig reject a signature produced by another key
func TestVerifyChequeSigWrongSigner(t *testing.T) {
	expectedCheque := newTestCheque()
	expectedCheque.Signature = testChequeSig

	// We expect the signer to be beneficiaryAddress but chequeSig is the signature from the owner
	if err := expectedCheque.VerifySig(beneficiaryAddress); err == nil {
		t.Fatal("Valid signature, should have been invalid")
	}
}

// helper function to make a signature "invalid"
func manipulateSignature(sig []byte) []byte {
	invalidSig := make([]byte, len(sig))
	copy(invalidSig, sig)
	// change one byte in the signature
	invalidSig[27] += 2
	return invalidSig
}

// tests if verifyChequeSig reject an invalid signature
func TestVerifyChequeInvalidSignature(t *testing.T) {
	expectedCheque := newTestCheque()
	expectedCheque.Signature = manipulateSignature(testChequeSig)

	if err := expectedCheque.VerifySig(ownerAddress); err == nil {
		t.Fatal("Valid signature, should have been invalid")
	}
}

// tests if TestValidateCode accepts an address with the correct bytecode
func TestVerifyContract(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// deploy a new swap contract
	err := testDeploy(context.TODO(), swap, big.NewInt(0))
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	if err = swap.chequebookFactory.VerifyContract(swap.GetParams().ContractAddress); err != nil {
		t.Fatalf("Contract verification failed: %v", err)
	}
}

// tests if ValidateCode rejects an address with different bytecode
func TestVerifyContractNotDeployedByFactory(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	opts := bind.NewKeyedTransactor(ownerKey)

	addr, _, _, err := contractFactory.DeployERC20SimpleSwap(opts, swap.backend, ownerAddress, common.Address{}, big.NewInt(int64(defaultHarddepositTimeoutDuration)))
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	if err = swap.chequebookFactory.VerifyContract(addr); err != cswap.ErrNotDeployedByFactory {
		t.Fatalf("Contract verification verified wrong contract: %v", err)
	}
}

// setupContractTest is a helper function for setting up the
// blockchain wait function for testing
func setupContractTest() func() {
	// we overwrite the waitForTx function with one which the simulated backend
	// immediately commits
	currentWaitFunc := cswap.WaitFunc
	// we also need to store the previous cashCheque function in case this is called multiple times
	currentCashCheque := defaultCashCheque
	defaultCashCheque = testCashCheque
	// overwrite only for the duration of the test, so...
	cswap.WaitFunc = testWaitForTx
	return func() {
		// ...we need to set it back to original when done
		cswap.WaitFunc = currentWaitFunc
		defaultCashCheque = currentCashCheque
	}
}

// TestFactoryAddressForNetwork tests that an address is found for ropsten
// and no address if for network 32145
func TestFactoryAddressForNetwork(t *testing.T) {
	address, err := cswap.FactoryAddressForNetwork(3)
	if err != nil {
		t.Fatal("didn't find address for ropsten")
	}
	if (address == common.Address{}) {
		t.Fatal("address for ropsten is empty")
	}
	_, err = cswap.FactoryAddressForNetwork(32145)
	if err == nil {
		t.Fatal("got address for a not supported network")
	}
}

// TestFactoryVerifySelf tests that it returns no error for a real factory
// and expects errors for a different contract or no contract
func TestFactoryVerifySelf(t *testing.T) {
	testBackend := newTestBackend(t)
	defer testBackend.Close()

	factory, err := cswap.FactoryAt(testBackend.factoryAddress, testBackend)
	if err != nil {
		t.Fatal(err)
	}
	if factory.VerifySelf() != nil {
		t.Fatal(err)
	}
}

// TestContractIntegration tests a end-to-end cheque interaction.
// First a simulated backend is created, then we deploy the issuer's swap contract.
// We issue a test cheque with the beneficiary address and on the issuer's contract,
// and immediately try to cash-in the cheque
// afterwards it attempts to cash-in a bouncing cheque
func TestContractIntegration(t *testing.T) {
	log.Debug("creating test swap")

	issuerSwap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	issuerSwap.owner.address = ownerAddress
	issuerSwap.owner.privateKey = ownerKey

	log.Debug("deploy issuer swap")

	cheque := newTestCheque()

	ctx := context.TODO()
	err := testDeploy(ctx, issuerSwap, big.NewInt(int64(cheque.CumulativePayout)))
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("deployed. signing cheque")
	cheque.ChequeParams.Contract = issuerSwap.GetParams().ContractAddress
	cheque.Signature, err = cheque.Sign(issuerSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("sending cheque...")

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	opts := bind.NewKeyedTransactor(beneficiaryKey)
	opts.Context = ctx

	log.Debug("cash-in the cheque")

	cashResult, receipt, err := issuerSwap.contract.CashChequeBeneficiary(opts, beneficiaryAddress, big.NewInt(int64(cheque.CumulativePayout)), cheque.Signature)

	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}
	if cashResult.Bounced {
		t.Fatal("cashing bounced")
	}

	// check state, check that cheque is indeed there
	result, err := issuerSwap.contract.PaidOut(nil, beneficiaryAddress)
	if err != nil {
		t.Fatal(err)
	}
	if result.Uint64() != cheque.CumulativePayout {
		t.Fatalf("Wrong cumulative payout %d", result)
	}
	log.Debug("cheques result", "result", result)

	// create a cheque that will bounce
	bouncingCheque := newTestCheque()
	bouncingCheque.ChequeParams.Contract = issuerSwap.GetParams().ContractAddress
	bouncingCheque.CumulativePayout = bouncingCheque.CumulativePayout + 10000*RetrieveRequestPrice
	bouncingCheque.Signature, err = bouncingCheque.Sign(issuerSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("try to cash-in the bouncing cheque")
	cashResult, receipt, err = issuerSwap.contract.CashChequeBeneficiary(opts, beneficiaryAddress, big.NewInt(int64(bouncingCheque.CumulativePayout)), bouncingCheque.Signature)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}
	if !cashResult.Bounced {
		t.Fatal("cheque did not bounce")
	}
}

// when testing, we don't need to wait for a transaction to be mined
func testWaitForTx(ctx context.Context, backend cswap.Backend, tx *types.Transaction) (*types.Receipt, error) {

	var stb *swapTestBackend
	var ok bool
	if stb, ok = backend.(*swapTestBackend); !ok {
		return nil, errors.New("not the expected test backend")
	}
	stb.Commit()

	receipt, err := backend.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, cswap.ErrTransactionReverted
	}
	return receipt, nil
}

// deploy for testing (needs simulated backend commit)
func testDeploy(ctx context.Context, swap *Swap, depositAmount *big.Int) (err error) {
	opts := bind.NewKeyedTransactor(swap.owner.privateKey)
	opts.Context = ctx

	var stb *swapTestBackend
	var ok bool
	if stb, ok = swap.backend.(*swapTestBackend); !ok {
		return errors.New("not the expected test backend")
	}

	factory, err := cswap.FactoryAt(stb.factoryAddress, stb)
	if err != nil {
		return err
	}

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()
	swap.contract, err = factory.DeploySimpleSwap(opts, swap.owner.address, big.NewInt(int64(defaultHarddepositTimeoutDuration)))
	if err != nil {
		return err
	}
	stb.Commit()

	// send money into the new chequebook
	token, err := contractFactory.NewERC20Mintable(stb.tokenAddress, stb)
	if err != nil {
		return err
	}

	tx, err := token.Mint(bind.NewKeyedTransactor(ownerKey), swap.contract.ContractParams().ContractAddress, depositAmount)
	if err != nil {
		return err
	}
	stb.Commit()
	receipt, err := stb.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return err
	}

	if receipt.Status != 1 {
		return errors.New("token transfer reverted")
	}

	return nil
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

// TestPeerSetAndGetLastReceivedCheque tests if a saved last received cheque can be loaded again later using the peer functions
func TestPeerSetAndGetLastReceivedCheque(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	testCheque := newTestCheque()

	if err := peer.setLastReceivedCheque(testCheque); err != nil {
		t.Fatalf("Error while saving: %s", err.Error())
	}

	returnedCheque := peer.getLastReceivedCheque()
	if returnedCheque == nil {
		t.Fatal("Could not find saved cheque")
	}

	if !returnedCheque.Equal(testCheque) {
		t.Fatal("Returned cheque was different")
	}

	// create a new swap peer for the same underlying peer to force a database load
	samePeer, err := swap.addPeer(peer.Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}

	returnedCheque = samePeer.getLastReceivedCheque()
	if returnedCheque == nil {
		t.Fatal("Could not find saved cheque")
	}

	if !returnedCheque.Equal(testCheque) {
		t.Fatal("Returned cheque was different")
	}
}

// TestPeerVerifyChequeProperties tests that verifyChequeProperties will accept a valid cheque
func TestPeerVerifyChequeProperties(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	testCheque := newTestCheque()
	testCheque.Signature = testChequeSig

	if err := testCheque.verifyChequeProperties(peer, swap.owner.address); err != nil {
		t.Fatalf("failed to verify cheque properties: %s", err.Error())
	}
}

// TestPeerVerifyChequeProperties tests that verifyChequeProperties will reject invalid cheques
func TestPeerVerifyChequePropertiesInvalidCheque(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	// cheque with an invalid signature
	testCheque := newTestCheque()
	testCheque.Signature = manipulateSignature(testChequeSig)
	if err := testCheque.verifyChequeProperties(peer, swap.owner.address); err == nil {
		t.Fatalf("accepted cheque with invalid signature")
	}

	// cheque with wrong contract
	testCheque = newTestCheque()
	testCheque.Contract = beneficiaryAddress
	testCheque.Signature, _ = testCheque.Sign(ownerKey)
	if err := testCheque.verifyChequeProperties(peer, swap.owner.address); err == nil {
		t.Fatalf("accepted cheque with wrong contract")
	}

	// cheque with wrong beneficiary
	testCheque = newTestCheque()
	testCheque.Beneficiary = ownerAddress
	testCheque.Signature, _ = testCheque.Sign(ownerKey)
	if err := testCheque.verifyChequeProperties(peer, swap.owner.address); err == nil {
		t.Fatalf("accepted cheque with wrong beneficiary")
	}
}

// TestPeerVerifyChequeAgainstLast tests that verifyChequeAgainstLast accepts a cheque with higher amount
func TestPeerVerifyChequeAgainstLast(t *testing.T) {
	increase := uint64(10)
	oldCheque := newTestCheque()
	newCheque := newTestCheque()

	newCheque.CumulativePayout = oldCheque.CumulativePayout + increase

	actualAmount, err := newCheque.verifyChequeAgainstLast(oldCheque, increase)
	if err != nil {
		t.Fatalf("failed to verify cheque compared to old cheque: %v", err)
	}

	if actualAmount != increase {
		t.Fatalf("wrong actual amount, expected: %d, was: %d", increase, actualAmount)
	}
}

// TestPeerVerifyChequeAgainstLastInvalid tests that verifyChequeAgainstLast rejects cheques with lower amount or an unexpected value
func TestPeerVerifyChequeAgainstLastInvalid(t *testing.T) {
	increase := uint64(10)

	// cheque with same or lower amount
	oldCheque := newTestCheque()
	newCheque := newTestCheque()

	if _, err := newCheque.verifyChequeAgainstLast(oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with same amount")
	}

	// cheque with amount != increase
	oldCheque = newTestCheque()
	newCheque = newTestCheque()
	newCheque.CumulativePayout = oldCheque.CumulativePayout + increase + 5

	if _, err := newCheque.verifyChequeAgainstLast(oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with unexpected amount")
	}
}

// TestPeerProcessAndVerifyCheque tests that processAndVerifyCheque accepts a valid cheque and also saves it
func TestPeerProcessAndVerifyCheque(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	// create test cheque and process
	cheque := newTestCheque()
	cheque.Signature, _ = cheque.Sign(ownerKey)

	actualAmount, err := swap.processAndVerifyCheque(cheque, peer)
	if err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	if actualAmount != cheque.CumulativePayout {
		t.Fatalf("computed wrong actual amount: was %d, expected: %d", actualAmount, cheque.CumulativePayout)
	}

	// verify that it was indeed saved
	if peer.getLastReceivedCheque().CumulativePayout != cheque.CumulativePayout {
		t.Fatalf("last received cheque has wrong cumulative payout, was: %d, expected: %d", peer.lastReceivedCheque.CumulativePayout, cheque.CumulativePayout)
	}

	// create another cheque with higher amount
	otherCheque := newTestCheque()
	otherCheque.CumulativePayout = cheque.CumulativePayout + 10
	otherCheque.Honey = 10
	otherCheque.Signature, _ = otherCheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(otherCheque, peer); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	// verify that it was indeed saved
	if peer.getLastReceivedCheque().CumulativePayout != otherCheque.CumulativePayout {
		t.Fatalf("last received cheque has wrong cumulative payout, was: %d, expected: %d", peer.lastReceivedCheque.CumulativePayout, otherCheque.CumulativePayout)
	}
}

// TestPeerProcessAndVerifyChequeInvalid verifies that processAndVerifyCheque does not accept cheques incompatible with the last cheque
// it first tries to process an invalid cheque
// then it processes a valid cheque
// then rejects one with lower amount
func TestPeerProcessAndVerifyChequeInvalid(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	// invalid cheque because wrong recipient
	cheque := newTestCheque()
	cheque.Beneficiary = ownerAddress
	cheque.Signature, _ = cheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(cheque, peer); err == nil {
		t.Fatal("accecpted an invalid cheque as first cheque")
	}

	// valid cheque
	cheque = newTestCheque()
	cheque.Signature, _ = cheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(cheque, peer); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	if peer.getLastReceivedCheque().CumulativePayout != cheque.CumulativePayout {
		t.Fatalf("last received cheque has wrong cumulative payout, was: %d, expected: %d", peer.lastReceivedCheque.CumulativePayout, cheque.CumulativePayout)
	}

	// invalid cheque because amount is lower
	otherCheque := newTestCheque()
	otherCheque.CumulativePayout = cheque.CumulativePayout - 10
	otherCheque.Honey = 10
	otherCheque.Signature, _ = otherCheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(otherCheque, peer); err == nil {
		t.Fatal("accepted a cheque with lower amount")
	}

	// check that no invalid cheque was saved
	if peer.getLastReceivedCheque().CumulativePayout != cheque.CumulativePayout {
		t.Fatalf("last received cheque has wrong cumulative payout, was: %d, expected: %d", peer.lastReceivedCheque.CumulativePayout, cheque.CumulativePayout)
	}
}

func TestSwapLogToFile(t *testing.T) {
	// create a log dir
	logDirDebitor, err := ioutil.TempDir("", "swap_test_log")
	log.Debug("creating swap log dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(logDirDebitor)

	// set the log dir to the params
	params := newDefaultParams(t)
	params.LogPath = logDirDebitor

	testBackend := newTestBackend(t)
	defer testBackend.Close()
	// create both test swap accounts
	creditorSwap, storeDirCreditor := newBaseTestSwap(t, beneficiaryKey, testBackend)
	// we are only checking one of the two nodes for logs
	debitorSwap, storeDirDebitor := newBaseTestSwapWithParams(t, ownerKey, params, testBackend)

	clean := func() {
		creditorSwap.Close()
		debitorSwap.Close()
		os.RemoveAll(storeDirCreditor)
		os.RemoveAll(storeDirDebitor)
	}
	defer clean()

	testAmount := int64(DefaultPaymentThreshold + 42)

	ctx := context.Background()
	err = testDeploy(ctx, creditorSwap, big.NewInt(testAmount))
	if err != nil {
		t.Fatal(err)
	}
	err = testDeploy(ctx, debitorSwap, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}

	// create Peer instances
	// NOTE: remember that these are peer instances representing each **a model of the remote peer** for every local node
	// so creditor is the model of the remote mode for the debitor! (and vice versa)
	cPeer := newDummyPeerWithSpec(Spec)
	dPeer := newDummyPeerWithSpec(Spec)
	creditor, err := debitorSwap.addPeer(cPeer.Peer, creditorSwap.owner.address, debitorSwap.GetParams().ContractAddress)
	if err != nil {
		t.Fatal(err)
	}
	debitor, err := creditorSwap.addPeer(dPeer.Peer, debitorSwap.owner.address, debitorSwap.GetParams().ContractAddress)
	if err != nil {
		t.Fatal(err)
	}

	// set balances arbitrarily
	debitor.setBalance(testAmount)
	creditor.setBalance(-testAmount)

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// now simulate sending the cheque to the creditor from the debitor
	creditor.sendCheque()

	if logDirDebitor == "" {
		t.Fatal("Swap Log Dir is not defined")
	}

	files, err := ioutil.ReadDir(logDirDebitor)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("expected at least 1 file in the log directory, found none")
	}

	logFile := path.Join(logDirDebitor, files[0].Name())

	var b []byte
	b, err = ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	logString := string(b)
	if !strings.Contains(logString, "sending cheque") {
		t.Fatalf("expected the log to contain \"sending cheque\"")
	}
}

func TestPeerGetLastSentCumulativePayout(t *testing.T) {
	_, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	if peer.getLastSentCumulativePayout() != 0 {
		t.Fatalf("last cumulative payout should be 0 in the beginning, was %d", peer.getLastSentCumulativePayout())
	}

	cheque := newTestCheque()
	if err := peer.setLastSentCheque(cheque); err != nil {
		t.Fatal(err)
	}

	if peer.getLastSentCumulativePayout() != cheque.CumulativePayout {
		t.Fatalf("last cumulative payout should be the payout of the last sent cheque, was: %d, expected %d", peer.getLastSentCumulativePayout(), cheque.CumulativePayout)
	}
}

func TestAvailableBalance(t *testing.T) {
	testBackend := newTestBackend(t)
	defer testBackend.Close()
	swap, clean := newTestSwap(t, ownerKey, testBackend)
	defer clean()
	cleanup := setupContractTest()
	defer cleanup()

	depositAmount := big.NewInt(9000 * int64(RetrieveRequestPrice))

	// deploy a chequebook
	err := testDeploy(context.TODO(), swap, depositAmount)
	if err != nil {
		t.Fatal(err)
	}
	// create a peer
	peer, err := swap.addPeer(newDummyPeerWithSpec(Spec).Peer, swap.owner.address, swap.GetParams().ContractAddress)
	if err != nil {
		t.Fatal(err)
	}

	// verify that available balance equals depositAmount (we deposit during deployment)
	availableBalance, err := swap.AvailableBalance()
	if err != nil {
		t.Fatal(err)
	}
	if availableBalance != depositAmount.Uint64() {
		t.Fatalf("availableBalance not equal to deposited amount. availableBalance: %d, depositAmount: %d", availableBalance, depositAmount)
	}
	// withdraw 50
	withdrawAmount := big.NewInt(50)
	netDeposit := depositAmount.Uint64() - withdrawAmount.Uint64()
	opts := bind.NewKeyedTransactor(swap.owner.privateKey)
	opts.Context = context.TODO()
	rec, err := swap.contract.Withdraw(opts, withdrawAmount)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != types.ReceiptStatusSuccessful {
		t.Fatal("Transaction reverted")
	}

	// verify available balance
	availableBalance, err = swap.AvailableBalance()
	if err != nil {
		t.Fatal(err)
	}
	if availableBalance != netDeposit {
		t.Fatalf("availableBalance not equal to deposited minus withdraw. availableBalance: %d, deposit minus withdrawn: %d", availableBalance, depositAmount.Uint64()-withdrawAmount.Uint64())
	}

	// send a cheque worth 42
	chequeAmount := int64(42)
	// create a dummy peer. Note: the peer's contract address and the peers address are resp the swap contract and the swap owner
	peer.setBalance(-chequeAmount)
	if err = peer.sendCheque(); err != nil {
		t.Fatal(err)
	}

	availableBalance, err = swap.AvailableBalance()
	if err != nil {
		t.Fatal(err)
	}
	// verify available balance
	if availableBalance != (netDeposit - uint64(chequeAmount)) {
		t.Fatalf("availableBalance not equal to deposited minus withdraw. availableBalance: %d, deposit minus withdrawn: %d", availableBalance, depositAmount.Uint64()-withdrawAmount.Uint64())
	}
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
