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
	"flag"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
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
)

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

//Test that repeated bookings do correct accounting
func TestRepeatedBookings(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	testPeer := newDummyPeer()
	amount := mrand.Intn(100)
	cnt := 1 + mrand.Intn(10)
	for i := 0; i < cnt; i++ {
		swap.Add(int64(amount), testPeer.Peer)
	}
	expectedBalance := int64(cnt * amount)
	realBalance := swap.balances[testPeer.ID()]
	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After %d credits of %d, expected balance to be: %d, but is: %d", cnt, amount, expectedBalance, realBalance))
	}

	testPeer2 := newDummyPeer()
	amount = mrand.Intn(100)
	cnt = 1 + mrand.Intn(10)
	for i := 0; i < cnt; i++ {
		swap.Add(0-int64(amount), testPeer2.Peer)
	}
	expectedBalance = int64(0 - (cnt * amount))
	realBalance = swap.balances[testPeer2.ID()]
	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After %d debits of %d, expected balance to be: %d, but is: %d", cnt, amount, expectedBalance, realBalance))
	}

	//mixed debits and credits
	amount1 := mrand.Intn(100)
	amount2 := mrand.Intn(55)
	amount3 := mrand.Intn(999)
	swap.Add(int64(amount1), testPeer2.Peer)
	swap.Add(int64(0-amount2), testPeer2.Peer)
	swap.Add(int64(0-amount3), testPeer2.Peer)

	expectedBalance = expectedBalance + int64(amount1-amount2-amount3)
	realBalance = swap.balances[testPeer2.ID()]

	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After mixed debits and credits, expected balance to be: %d, but is: %d", expectedBalance, realBalance))
	}
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
	swap := New(stateStore, key, common.Address{})
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
	expected := common.Hex2Bytes("4405415b2b8c9f9aa83e151637b8378dd3bcfedd0000000000000000000000000000000000000000000000000000000000000001b8d424e9662fe0837fb1d728f1ac97cebb1085fe000000000000000000000000000000000000000000000000000000000000002a000000000000000000000000000000000000000000000000000000000000000a")
	if !bytes.Equal(encoded, expected) {
		t.Fatal(fmt.Sprintf("Unexpected encoding of cheque. Expected encoding: %x, result is: %x",
			expected, encoded))
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
	expected := common.Hex2Bytes("305cf876a5c6a24430743695fa5a42d40f8d59e174921520c8efe2d01c9b2a6a")
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

	swap.owner.privateKey, err = crypto.HexToECDSA("634fb5a872396d9693e5c9f9d7233cfa93f395c093371017ff44aa9ae6564cdd")

	// sign the cheque
	sig, err := swap.signContent(expectedCheque)
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("833ffa1515b545ce75f4cbe520e6d22bcd76f8e688920e82cf9800a2a7891dda7d4b9f702d6da1b026c3bd860b00028ecce0003daba887c22b5926c26452136b1c")
	if err != nil {
		t.Fatal(fmt.Sprintf("Error in signing: %s", err))
	}
	if !bytes.Equal(sig, expected) {
		t.Fatal(fmt.Sprintf("Unexpected signature for cheque. Expected: %x, result is: %x",
			expected, sig))
	}
}
