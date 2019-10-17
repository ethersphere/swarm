package swap

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/state"
)

// Test getting a peer's balance
func TestPeerBalance(t *testing.T) {
	// create a test swap account
	swap, testPeer, clean := newTestSwapAndPeer(t, ownerKey)
	testPeerID := testPeer.ID()
	defer clean()

	// test for correct balance
	setBalance(t, testPeer, 888)
	testBalance(t, swap, testPeerID, 888)

	// test balance after change
	setBalance(t, testPeer, 17000)
	testBalance(t, swap, testPeerID, 17000)

	// test balance for second peer
	testPeer2 := addPeer(t, swap)
	testPeer2ID := testPeer2.ID()

	setBalance(t, testPeer2, 4)
	testBalance(t, swap, testPeer2ID, 4)

	// test balance for inexistent node
	invalidPeerID := adapters.RandomNodeConfig().ID
	_, err := swap.PeerBalance(invalidPeerID)
	if err == nil {
		t.Fatal("Expected call to fail, but it didn't!")
	}
	if err != state.ErrNotFound {
		t.Fatalf("Expected test to fail with %s, but is %s", "ErrorNotFound", err.Error())
	}

	// test balance for disconnected node
	testPeer3 := newDummyPeer().Peer
	testPeer3ID := testPeer3.ID()
	err = swap.saveBalance(testPeer3ID, 777)
	testBalance(t, swap, testPeer3ID, 777)

	// test previous results are still correct
	testBalance(t, swap, testPeerID, 17000)
	testBalance(t, swap, testPeer2ID, 4)
}

// sets the given balance for the given peer, fails if there are errors
func setBalance(t *testing.T, p *Peer, balance int64) {
	t.Helper()
	err := p.setBalance(balance)
	if err != nil {
		t.Fatal(err)
	}
}

// tests that expected balance for peer matches the result of the PeerBalance function
func testBalance(t *testing.T, s *Swap, id enode.ID, expectedBalance int64) {
	t.Helper()
	b, err := s.PeerBalance(id)
	if err != nil {
		t.Fatal(err)
	}
	if b != expectedBalance {
		t.Fatalf("Expected peer's balance to be %d, but is %d", expectedBalance, b)
	}
}

// Test getting balances for all known peers
func TestBalances(t *testing.T) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// test balances are empty
	testBalances(t, swap, map[enode.ID]int64{})

	// add peer
	testPeer := addPeer(t, swap)
	testPeerID := testPeer.ID()

	// test balances with one peer
	setBalance(t, testPeer, 808)
	testBalances(t, swap, map[enode.ID]int64{testPeerID: 808})

	// add second peer
	testPeer2 := addPeer(t, swap)
	testPeer2ID := testPeer2.ID()

	// test balances with second peer
	setBalance(t, testPeer2, 123)
	testBalances(t, swap, map[enode.ID]int64{testPeerID: 808, testPeer2ID: 123})

	// test balances after balance change for peer
	setBalance(t, testPeer, 303)
	testBalances(t, swap, map[enode.ID]int64{testPeerID: 303, testPeer2ID: 123})
}

// tests that a map of peerID:balance matches the result of the Balances function
func testBalances(t *testing.T, s *Swap, expectedBalances map[enode.ID]int64) {
	t.Helper()
	actualBalances, err := s.Balances()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actualBalances, expectedBalances) {
		t.Fatalf("Expected node's balances to be %d, but are %d", expectedBalances, actualBalances)
	}
}

// TestCheques verifies that sent and received cheques data for all known swap peers is correct
func TestCheques(t *testing.T) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// check cheques are empty
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{})

	// add peer
	testPeer := addPeer(t, swap)
	testPeerID := testPeer.ID()

	// test no cheques are present
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: nil, sentChequeResponseKey: nil}})

	// test sent cheque for peer
	sentCheque := setNewSentCheque(t, testPeer)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: nil, sentChequeResponseKey: sentCheque}})

	// test received cheque for peer
	receivedCheque := setNewReceivedCheque(t, testPeer)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}})

	// add second peer
	testPeer2 := addPeer(t, swap)
	testPeer2ID := testPeer2.ID()

	// test sent cheque for second peer
	sentCheque2 := setNewSentCheque(t, testPeer2)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: nil, sentChequeResponseKey: sentCheque2}})

	// test received cheque for second peer
	receivedCheque2 := setNewReceivedCheque(t, testPeer2)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: receivedCheque2, sentChequeResponseKey: sentCheque2}})

	// test sent cheque change for second peer
	receivedCheque3 := setNewReceivedCheque(t, testPeer2)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: receivedCheque3, sentChequeResponseKey: sentCheque2}})

	// test cheques for disconnected node
	testPeer3ID := newDummyPeer().Peer.ID()

	// test sent cheque for disconnected node
	sentCheque3 := saveNewSentCheque(t, swap, testPeer3ID)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: receivedCheque3, sentChequeResponseKey: sentCheque2}, testPeer3ID: {sentChequeResponseKey: sentCheque3, receivedChequeResponseKey: nil}})

	// test received cheque for disconnected node
	receivedCheque4 := saveNewReceivedCheque(t, swap, testPeer3ID)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: receivedCheque3, sentChequeResponseKey: sentCheque2}, testPeer3ID: {sentChequeResponseKey: sentCheque3, receivedChequeResponseKey: receivedCheque4}})

	// test cheque change for disconnected node
	sentCheque4 := saveNewSentCheque(t, swap, testPeer3ID)
	testChequesByPeerAndType(t, swap, map[enode.ID]map[string]*Cheque{testPeerID: {receivedChequeResponseKey: receivedCheque, sentChequeResponseKey: sentCheque}, testPeer2ID: {receivedChequeResponseKey: receivedCheque3, sentChequeResponseKey: sentCheque2}, testPeer3ID: {sentChequeResponseKey: sentCheque4, receivedChequeResponseKey: receivedCheque4}})
}

// tests that a nested map of peerID:{typeOfCheque:cheque} matches the result of the Cheques function
func testChequesByPeerAndType(t *testing.T, s *Swap, expectedCheques map[enode.ID]map[string]*Cheque) {
	t.Helper()
	cheques, err := s.Cheques()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedCheques, cheques) {
		t.Fatalf("Expected cheques to be %v, but are %v", expectedCheques, cheques)
	}
}

// generates a cheque and adds it as the last sent cheque for the given peer, fails if there are errors
func setNewSentCheque(t *testing.T, p *Peer) *Cheque {
	t.Helper()
	return setNewCheque(t, p.setLastSentCheque)
}

// generates a cheque and adds it as the last received cheque for the given peer, fails if there are errors
func setNewReceivedCheque(t *testing.T, p *Peer) *Cheque {
	t.Helper()
	return setNewCheque(t, p.setLastReceivedCheque)
}

func setNewCheque(t *testing.T, setChequeFunction func(*Cheque) error) *Cheque {
	t.Helper()
	newCheque := newRandomTestCheque()
	err := setChequeFunction(newCheque)
	if err != nil {
		t.Fatal(err)
	}
	return newCheque
}

// generates a cheque and saves it as the last sent cheque for a peer in the given swap struct, fails if there are errors
func saveNewSentCheque(t *testing.T, s *Swap, id enode.ID) *Cheque {
	t.Helper()
	return saveNewCheque(t, id, s.saveLastSentCheque)
}

// generates a cheque and saves it as the last received cheque for a peer in the given swap struct, fails if there are errors
func saveNewReceivedCheque(t *testing.T, s *Swap, id enode.ID) *Cheque {
	t.Helper()
	return saveNewCheque(t, id, s.saveLastReceivedCheque)
}

func saveNewCheque(t *testing.T, id enode.ID, saveChequeFunction func(enode.ID, *Cheque) error) *Cheque {
	t.Helper()
	newCheque := newRandomTestCheque()
	err := saveChequeFunction(id, newCheque)
	if err != nil {
		t.Fatal(err)
	}
	return newCheque
}

// TestPeerCheques verifies that sent and received cheques data for a given peer is correct
func TestPeerCheques(t *testing.T) {
	testBackend := newTestBackend()
	defer testBackend.Close()
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, testBackend)
	defer clean()

	// add peer
	testPeer := addPeer(t, swap)
	testPeerID := testPeer.ID()

	// test peer cheques are nil
	testChequesByType(t, swap, testPeerID, map[string]*Cheque{sentChequeResponseKey: nil, receivedChequeResponseKey: nil})

	// test sent cheque for peer
	sentCheque := setNewSentCheque(t, testPeer)
	testChequesByType(t, swap, testPeerID, map[string]*Cheque{sentChequeResponseKey: sentCheque, receivedChequeResponseKey: nil})

	// test received cheque for peer
	receivedCheque := setNewReceivedCheque(t, testPeer)
	testChequesByType(t, swap, testPeerID, map[string]*Cheque{sentChequeResponseKey: sentCheque, receivedChequeResponseKey: receivedCheque})

	// add second peer
	testPeer2 := addPeer(t, swap)
	testPeer2ID := testPeer2.ID()

	// test sent cheque for second peer
	sentCheque2 := setNewSentCheque(t, testPeer2)
	testChequesByType(t, swap, testPeer2ID, map[string]*Cheque{sentChequeResponseKey: sentCheque2, receivedChequeResponseKey: nil})

	// test received cheque for second peer
	receivedCheque2 := setNewReceivedCheque(t, testPeer2)
	testChequesByType(t, swap, testPeer2ID, map[string]*Cheque{sentChequeResponseKey: sentCheque2, receivedChequeResponseKey: receivedCheque2})

	// check previous cheques are still correct
	testChequesByType(t, swap, testPeerID, map[string]*Cheque{sentChequeResponseKey: sentCheque, receivedChequeResponseKey: receivedCheque})

	// check change in cheques for peer
	sentCheque3 := setNewSentCheque(t, testPeer)
	testChequesByType(t, swap, testPeerID, map[string]*Cheque{sentChequeResponseKey: sentCheque3, receivedChequeResponseKey: receivedCheque})

	// check previous cheques for second peer are still correct
	testChequesByType(t, swap, testPeer2ID, map[string]*Cheque{sentChequeResponseKey: sentCheque2, receivedChequeResponseKey: receivedCheque2})

	// test cheques for invalid peer
	testChequesByType(t, swap, adapters.RandomNodeConfig().ID, map[string]*Cheque{sentChequeResponseKey: nil, receivedChequeResponseKey: nil})

	// test cheques for disconnected node
	testPeer3ID := newDummyPeer().Peer.ID()

	// test sent cheque for disconnected node
	sentCheque4 := saveNewSentCheque(t, swap, testPeer3ID)
	testChequesByType(t, swap, testPeer3ID, map[string]*Cheque{sentChequeResponseKey: sentCheque4, receivedChequeResponseKey: nil})

	// test received cheque for disconnected node
	receivedCheque3 := saveNewReceivedCheque(t, swap, testPeer3ID)
	testChequesByType(t, swap, testPeer3ID, map[string]*Cheque{sentChequeResponseKey: sentCheque4, receivedChequeResponseKey: receivedCheque3})
}

// tests that map of typeOfCheque:cheque matches the result of the PeerCheques function
func testChequesByType(t *testing.T, s *Swap, id enode.ID, expectedCheques map[string]*Cheque) {
	t.Helper()
	peerCheques, err := s.PeerCheques(id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedCheques, peerCheques) {
		t.Fatalf("Expected peer cheques to be %v, but are %v", expectedCheques, peerCheques)
	}
}
