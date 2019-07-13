// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package swap

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	contracts "github.com/ethersphere/swarm/contracts/swap/contract"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel           = flag.Int("loglevel", 2, "verbosity of logs")
	ownerKey, _        = crypto.HexToECDSA("634fb5a872396d9693e5c9f9d7233cfa93f395c093371017ff44aa9ae6564cdd")
	ownerAddress       = crypto.PubkeyToAddress(ownerKey.PublicKey)
	beneficiaryKey, _  = crypto.HexToECDSA("6f05b0a29723ca69b1fc65d11752cee22c200cf3d2938e670547f7ae525be112")
	beneficiaryAddress = crypto.PubkeyToAddress(beneficiaryKey.PublicKey)
	testSwapAdress     = common.HexToAddress("0x4405415b2B8c9F9aA83E151637B8378dD3bcfEDD")
	chequeSig          = common.Hex2Bytes("d985613f7d8bfcf0f96f4bb00a21111beb9a675477f47e4d9b79c89f880cf99c5ab9ef4cdec7186debc51b898fe4d062a835de61fba6db390316db13d50d23941c")
)

// booking represents an accounting movement in relation to a particular node: `peer`
// if `amount` is positive, it means the node which adds this booking will be credited in respect to `peer`
// otherwise it will be debited
type booking struct {
	amount int64
	peer   *protocols.Peer
}

func init() {
	flag.Parse()
	mrand.Seed(time.Now().UnixNano())

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

//Test getting a peer's balance
func TestGetPeerBalance(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	//test for correct value
	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = 888
	b, err := swap.GetPeerBalance(testPeer.ID())
	if err != nil {
		t.Fatal(err)
	}
	if b != 888 {
		t.Fatalf("Expected peer's balance to be %d, but is %d", 888, b)
	}

	//test for inexistent node
	id := adapters.RandomNodeConfig().ID
	_, err = swap.GetPeerBalance(id)
	if err == nil {
		t.Fatal("Expected call to fail, but it didn't!")
	}
	if err.Error() != "Peer not found" {
		t.Fatalf("Expected test to fail with %s, but is %s", "Peer not found", err.Error())
	}
}

func TestGetAllBalances(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	if len(swap.balances) != 0 {
		t.Fatalf("Expected balances to be empty, but are %v", swap.balances)
	}

	//test balance addition for peer
	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = 808
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 808})

	//test successive balance addition for peer
	testPeer2 := newDummyPeer()
	swap.balances[testPeer2.ID()] = 909
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 808, testPeer2.ID(): 909})

	//test balance change for peer
	swap.balances[testPeer.ID()] = 303
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 303, testPeer2.ID(): 909})
}

func testBalances(t *testing.T, swap *Swap, expectedBalances map[enode.ID]int64) {
	balances := swap.GetAllBalances()
	if !reflect.DeepEqual(balances, expectedBalances) {
		t.Fatalf("Expected node's balances to be %d, but are %d", expectedBalances, balances)
	}
}

//Test that repeated bookings do correct accounting
func TestRepeatedBookings(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	var bookings []booking

	// credits to peer 1
	testPeer := newDummyPeer()
	bookingAmount := int64(mrand.Intn(100))
	bookingQuantity := 1 + mrand.Intn(10)
	testPeerBookings(t, swap, &bookings, bookingAmount, bookingQuantity, testPeer.Peer)

	// debits to peer 2
	testPeer2 := newDummyPeer()
	bookingAmount = 0 - int64(mrand.Intn(100))
	bookingQuantity = 1 + mrand.Intn(10)
	testPeerBookings(t, swap, &bookings, bookingAmount, bookingQuantity, testPeer2.Peer)

	// credits and debits to peer 2
	mixedBookings := []booking{
		booking{int64(mrand.Intn(100)), testPeer2.Peer},
		booking{int64(0 - mrand.Intn(55)), testPeer2.Peer},
		booking{int64(0 - mrand.Intn(999)), testPeer2.Peer},
	}
	addBookings(swap, mixedBookings)
	verifyBookings(t, swap, append(bookings, mixedBookings...))
}

// generate bookings based on parameters, apply them to a Swap struct and verify the result
// append generated bookings to slice pointer
func testPeerBookings(t *testing.T, swap *Swap, bookings *[]booking, bookingAmount int64, bookingQuantity int, peer *protocols.Peer) {
	peerBookings := generateBookings(bookingAmount, bookingQuantity, peer)
	*bookings = append(*bookings, peerBookings...)
	addBookings(swap, peerBookings)
	verifyBookings(t, swap, *bookings)
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
func verifyBookings(t *testing.T, swap *Swap, bookings []booking) {
	expectedBalances := calculateExpectedBalances(swap, bookings)
	realBalances := swap.balances
	if !reflect.DeepEqual(expectedBalances, realBalances) {
		t.Fatal(fmt.Sprintf("After %d bookings, expected balance to be %v, but is %v", len(bookings), stringifyBalance(expectedBalances), stringifyBalance(realBalances)))
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
		// balance is not expected to be affected once past the disconnect threshold
		if peerBalance < swap.disconnectThreshold {
			peerBalance += booking.amount
		}
		expectedBalances[peerID] = peerBalance
	}
	return expectedBalances
}

//try restoring a balance from state store
//this is simulated by creating a node,
//assigning it an arbitrary balance,
//then closing the state store.
//Then we re-open the state store and check that
//the balance is still the same
func TestRestoreBalanceFromStateStore(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = -8888

	tmpBalance := swap.balances[testPeer.ID()]
	swap.stateStore.Put(testPeer.ID().String(), &tmpBalance)

	swap.stateStore.Close()
	swap.stateStore = nil

	stateStore, err := state.NewDBStore(testDir)
	if err != nil {
		t.Fatal(err)
	}

	var newBalance int64
	stateStore.Get(testPeer.ID().String(), &newBalance)

	//compare the balances
	if tmpBalance != newBalance {
		t.Fatal(fmt.Sprintf("Unexpected balance value after sending cheap message test. Expected balance: %d, balance is: %d",
			tmpBalance, newBalance))
	}
}

//create a test swap account
//creates a stateStore for persistence and a Swap account
func createTestSwap(t *testing.T) (*Swap, string) {
	dir, err := ioutil.TempDir("", "swap_test_store")
	if err != nil {
		t.Fatal(err)
	}
	stateStore, err2 := state.NewDBStore(dir)
	if err2 != nil {
		t.Fatal(err2)
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	contractBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		ownerAddress: {Balance: big.NewInt(1000000000)},
	}, 8000000)
	swap := New(stateStore, key, common.Address{}, contractBackend)
	return swap, dir
}

type dummyPeer struct {
	*protocols.Peer
}

//creates a dummy protocols.Peer with dummy MsgReadWriter
func newDummyPeer() *dummyPeer {
	id := adapters.RandomNodeConfig().ID
	protoPeer := protocols.NewPeer(p2p.NewPeer(id, "testPeer", nil), nil, nil)
	dummy := &dummyPeer{
		Peer: protoPeer,
	}
	return dummy
}

func newTestCheque() *Cheque {
	contract := common.HexToAddress("0x4405415b2B8c9F9aA83E151637B8378dD3bcfEDD")
	cashInDelay := 10

	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:    contract,
			Serial:      uint64(1),
			Amount:      uint64(42),
			Timeout:     uint64(cashInDelay),
			Beneficiary: beneficiaryAddress,
		},
	}

	return cheque
}

func TestEncodeCheque(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()

	// encode the cheque
	encoded := swap.encodeCheque(expectedCheque)
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("4405415b2b8c9f9aa83e151637b8378dd3bcfeddb8d424e9662fe0837fb1d728f1ac97cebb1085fe0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002a000000000000000000000000000000000000000000000000000000000000000a")
	if !bytes.Equal(encoded, expected) {
		t.Fatalf("Unexpected encoding of cheque. Expected encoding: %x, result is: %x",
			expected, encoded)
	}
}

func TestSigHashCheque(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()

	// compute the hash that will be signed
	hash := swap.sigHashCheque(expectedCheque)
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("e431e83bed105cb66d9aa5878cb010bc21365d2e328ce7c36671f0cbd44070ae")
	if !bytes.Equal(hash, expected) {
		t.Fatal(fmt.Sprintf("Unexpected sigHash of cheque. Expected: %x, result is: %x",
			expected, hash))
	}
}

func TestSignContent(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()

	var err error

	swap.owner.privateKey = ownerKey

	// sign the cheque
	sig, err := swap.signContent(expectedCheque)
	// expected value (computed through truffle/js)
	expected := chequeSig
	if err != nil {
		t.Fatal(fmt.Sprintf("Error in signing: %s", err))
	}
	if !bytes.Equal(sig, expected) {
		t.Fatal(fmt.Sprintf("Unexpected signature for cheque. Expected: %x, result is: %x",
			expected, sig))
	}
}

func TestVerifyChequeSig(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()
	expectedCheque.Sig = chequeSig

	err := swap.verifyChequeSig(expectedCheque, ownerAddress)

	if err != nil {
		t.Fatalf("Invalid signature: %v", err)
	}

}

func TestVerifyChequeSigWrongSigner(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()
	expectedCheque.Sig = chequeSig

	err := swap.verifyChequeSig(expectedCheque, beneficiaryAddress)

	if err == nil {
		t.Fatalf("Valid signature, should have been invalid")
	}
}

func TestVerifyChequeInvalidSignature(t *testing.T) {
	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()

	invalidSig := chequeSig[:]
	// change one byte in the signature
	invalidSig[27] += 2
	expectedCheque.Sig = invalidSig

	err := swap.verifyChequeSig(expectedCheque, ownerAddress)

	if err == nil {
		t.Fatalf("Valid signature, should have been invalid")
	}
}

func TestVerifyContract(t *testing.T) {
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	opts := bind.NewKeyedTransactor(ownerKey)
	addr, _, _, err := cswap.Deploy(opts, swap.backend, ownerAddress)

	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	swap.backend.(*backends.SimulatedBackend).Commit()

	err = swap.verifyContract(context.TODO(), addr)

	if err != nil {
		t.Fatalf("Contract verification failed: %v", err)
	}
}

func TestVerifyContractWrongContract(t *testing.T) {
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	opts := bind.NewKeyedTransactor(ownerKey)

	addr, _, _, err := contracts.DeployECDSA(opts, swap.backend)

	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	swap.backend.(*backends.SimulatedBackend).Commit()

	err = swap.verifyContract(context.TODO(), addr)

	if err != ErrNotASwapContract {
		t.Fatalf("Contract verification verified wrong contract: %v", err)
	}
}
