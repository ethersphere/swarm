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
	"flag"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
	colorable "github.com/mattn/go-colorable"
)

var (
	loglevel = flag.Int("loglevel", 2, "verbosity of logs")
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

//Test that repeated bookings do correct accounting
func TestRepeatedBookings(t *testing.T) {
	//create a test swap account
	swap, testDir := createTestSwap(t)
	defer os.RemoveAll(testDir)

	testPeer := newDummyPeer()
	amount := mrand.Intn(100)
	cnt := 1 + mrand.Intn(10)
	var bookings []booking
	for i := 0; i < cnt; i++ {
		bookings = append(bookings, booking{int64(amount), testPeer.Peer})
	}
	addBookings(swap, bookings)
	balancesAfterBookings := calculateExpectedBalances(swap, bookings)
	expectedBalance := balancesAfterBookings[testPeer.Peer.ID()]
	realBalance := swap.balances[testPeer.ID()]
	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After %d credits of %d, expected balance to be: %d, but is: %d", cnt, amount, expectedBalance, realBalance))
	}

	testPeer2 := newDummyPeer()
	amount = mrand.Intn(100)
	cnt = 1 + mrand.Intn(10)
	for i := 0; i < cnt; i++ {
		bookings = append(bookings, booking{0 - int64(amount), testPeer2.Peer})
	}
	addBookings(swap, bookings[len(bookings)-cnt:])
	balancesAfterBookings = calculateExpectedBalances(swap, bookings)
	expectedBalance = balancesAfterBookings[testPeer2.Peer.ID()]
	realBalance = swap.balances[testPeer2.ID()]
	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After %d debits of %d, expected balance to be: %d, but is: %d", cnt, amount, expectedBalance, realBalance))
	}

	//mixed debits and credits
	mixedBookings := []booking{
		booking{int64(mrand.Intn(100)), testPeer2.Peer},
		booking{int64(0 - mrand.Intn(55)), testPeer2.Peer},
		booking{int64(0 - mrand.Intn(999)), testPeer2.Peer},
	}
	addBookings(swap, mixedBookings)
	balancesAfterBookings = calculateExpectedBalances(swap, append(bookings, mixedBookings...))
	expectedBalance = balancesAfterBookings[testPeer2.Peer.ID()]
	realBalance = swap.balances[testPeer2.ID()]
	if expectedBalance != realBalance {
		t.Fatal(fmt.Sprintf("After mixed debits and credits, expected balance to be: %d, but is: %d", expectedBalance, realBalance))
	}
}

func addBookings(swap *Swap, bookings []booking) {
	for i := 0; i < len(bookings); i++ {
		booking := bookings[i]
		swap.Add(booking.amount, booking.peer)
	}
}

// take a swap pointer and a list of bookings, and calculate the expected balances.
// the result is a map which stores the balance for all the peers present in the bookings,
// from the perspective of the node that loaded the swap struct.
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
	swap := New(stateStore)
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
