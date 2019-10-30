package swap

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/state"
)

type peerChequesTestCase struct {
	name            string
	peer            *protocols.Peer
	sentCheque      *Cheque
	receivedCheque  *Cheque
	expectedCheques PeerCheques
}

type chequesTestCase struct {
	name                 string
	protoPeers           []*protocols.Peer
	sentCheques          map[*protocols.Peer]*Cheque
	receivedCheques      map[*protocols.Peer]*Cheque
	storeSentCheques     map[enode.ID]*Cheque
	storeReceivedCheques map[enode.ID]*Cheque
	expectedCheques      map[enode.ID]*PeerCheques
}

// Test getting a peer's balance
func TestPeerBalance(t *testing.T) {
	// create a test swap account
	swap, testPeer, clean := newTestSwapAndPeer(t, ownerKey)
	testPeerID := testPeer.ID()
	defer clean()

	// test for correct balance
	setBalance(t, testPeer, 888)
	testPeerBalance(t, swap, testPeerID, 888)

	// test balance after change
	setBalance(t, testPeer, 17000)
	testPeerBalance(t, swap, testPeerID, 17000)

	// test balance for second peer
	testPeer2 := addPeer(t, swap)
	testPeer2ID := testPeer2.ID()

	setBalance(t, testPeer2, 4)
	testPeerBalance(t, swap, testPeer2ID, 4)

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
	testPeerBalance(t, swap, testPeer3ID, 777)

	// test previous results are still correct
	testPeerBalance(t, swap, testPeerID, 17000)
	testPeerBalance(t, swap, testPeer2ID, 4)
}

// tests that expected balance for peer matches the result of the Balance function
func testPeerBalance(t *testing.T, s *Swap, id enode.ID, expectedBalance int64) {
	t.Helper()
	b, err := s.PeerBalance(id)
	if err != nil {
		t.Fatal(err)
	}
	if b != expectedBalance {
		t.Fatalf("Expected peer's balance to be %d, but is %d", expectedBalance, b)
	}
}

func addPeer(t *testing.T, s *Swap) *Peer {
	t.Helper()
	peer, err := s.addPeer(newDummyPeer().Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}
	return peer
}

// sets the given balance for the given peer, fails if there are errors
func setBalance(t *testing.T, p *Peer, balance int64) {
	t.Helper()
	err := p.setBalance(balance)
	if err != nil {
		t.Fatal(err)
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
	// generate peers and cheques
	// peer 1
	testPeer := newDummyPeer().Peer
	testPeerSentCheque := newRandomTestCheque()
	testPeerReceivedCheque := newRandomTestCheque()
	testPeerSentCheque2 := newRandomTestCheque()
	// peer 2
	testPeer2 := newDummyPeer().Peer
	testPeer2SentCheque := newRandomTestCheque()
	testPeer2ReceivedCheque := newRandomTestCheque()
	testPeer2ReceivedCheque2 := newRandomTestCheque()
	// disconnected peer
	testPeer3ID := newDummyPeer().Peer.ID()
	testPeer3SentCheque := newRandomTestCheque()
	testPeer3SentCheque2 := newRandomTestCheque()
	testPeer3ReceivedCheque := newRandomTestCheque()
	testPeer3ReceivedCheque2 := newRandomTestCheque()

	// build test cases
	testCases := []chequesTestCase{
		{
			name:                 "no peers",
			protoPeers:           []*protocols.Peer{},
			sentCheques:          map[*protocols.Peer]*Cheque{},
			receivedCheques:      map[*protocols.Peer]*Cheque{},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques:      map[enode.ID]*PeerCheques{},
		},
		{
			name:                 "one peer",
			protoPeers:           []*protocols.Peer{testPeer},
			sentCheques:          map[*protocols.Peer]*Cheque{},
			receivedCheques:      map[*protocols.Peer]*Cheque{},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques:      map[enode.ID]*PeerCheques{},
		},
		{
			name:                 "one peer, one sent cheque",
			protoPeers:           []*protocols.Peer{testPeer},
			sentCheques:          map[*protocols.Peer]*Cheque{testPeer: testPeerSentCheque},
			receivedCheques:      map[*protocols.Peer]*Cheque{},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer.ID(): {testPeerSentCheque, nil},
			},
		},
		{
			name:                 "one peer, sent and received cheques",
			protoPeers:           []*protocols.Peer{testPeer},
			sentCheques:          map[*protocols.Peer]*Cheque{testPeer: testPeerSentCheque},
			receivedCheques:      map[*protocols.Peer]*Cheque{testPeer: testPeerReceivedCheque},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer.ID(): {testPeerSentCheque, testPeerReceivedCheque},
			},
		},
		{
			name:                 "two peers, sent and received cheques",
			protoPeers:           []*protocols.Peer{testPeer, testPeer2},
			sentCheques:          map[*protocols.Peer]*Cheque{testPeer: testPeerSentCheque, testPeer2: testPeer2SentCheque},
			receivedCheques:      map[*protocols.Peer]*Cheque{testPeer: testPeerReceivedCheque, testPeer2: testPeer2ReceivedCheque},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer.ID():  {testPeerSentCheque, testPeerReceivedCheque},
				testPeer2.ID(): {testPeer2SentCheque, testPeer2ReceivedCheque},
			},
		},
		{
			name:                 "two peers, successive sent and received cheques",
			protoPeers:           []*protocols.Peer{testPeer, testPeer2},
			sentCheques:          map[*protocols.Peer]*Cheque{testPeer: testPeerSentCheque, testPeer2: testPeer2SentCheque, testPeer: testPeerSentCheque2},
			receivedCheques:      map[*protocols.Peer]*Cheque{testPeer: testPeerReceivedCheque, testPeer2: testPeer2ReceivedCheque, testPeer2: testPeer2ReceivedCheque2},
			storeSentCheques:     map[enode.ID]*Cheque{},
			storeReceivedCheques: map[enode.ID]*Cheque{},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer.ID():  {testPeerSentCheque2, testPeerReceivedCheque},
				testPeer2.ID(): {testPeer2SentCheque, testPeer2ReceivedCheque2},
			},
		},
		{
			name:                 "disconnected node, sent and received cheques",
			protoPeers:           []*protocols.Peer{},
			sentCheques:          map[*protocols.Peer]*Cheque{},
			receivedCheques:      map[*protocols.Peer]*Cheque{},
			storeSentCheques:     map[enode.ID]*Cheque{testPeer3ID: testPeer3SentCheque},
			storeReceivedCheques: map[enode.ID]*Cheque{testPeer3ID: testPeer3ReceivedCheque},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer3ID: {testPeer3SentCheque, testPeer3ReceivedCheque},
			},
		},
		{
			name:                 "disconnected node, successive sent and received cheques",
			protoPeers:           []*protocols.Peer{},
			sentCheques:          map[*protocols.Peer]*Cheque{},
			receivedCheques:      map[*protocols.Peer]*Cheque{},
			storeSentCheques:     map[enode.ID]*Cheque{testPeer3ID: testPeer3SentCheque, testPeer3ID: testPeer3SentCheque2},
			storeReceivedCheques: map[enode.ID]*Cheque{testPeer3ID: testPeer3ReceivedCheque, testPeer3ID: testPeer3ReceivedCheque2},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer3ID: {testPeer3SentCheque2, testPeer3ReceivedCheque2},
			},
		},
		{
			name:                 "full",
			protoPeers:           []*protocols.Peer{testPeer, testPeer2},
			sentCheques:          map[*protocols.Peer]*Cheque{testPeer: testPeerSentCheque, testPeer2: testPeer2SentCheque, testPeer: testPeerSentCheque2},
			receivedCheques:      map[*protocols.Peer]*Cheque{testPeer: testPeerReceivedCheque, testPeer2: testPeer2ReceivedCheque, testPeer2: testPeer2ReceivedCheque2},
			storeSentCheques:     map[enode.ID]*Cheque{testPeer3ID: testPeer3SentCheque, testPeer3ID: testPeer3SentCheque2},
			storeReceivedCheques: map[enode.ID]*Cheque{testPeer3ID: testPeer3ReceivedCheque, testPeer3ID: testPeer3ReceivedCheque2},
			expectedCheques: map[enode.ID]*PeerCheques{
				testPeer.ID():  {testPeerSentCheque2, testPeerReceivedCheque},
				testPeer2.ID(): {testPeer2SentCheque, testPeer2ReceivedCheque2},
				testPeer3ID:    {testPeer3SentCheque2, testPeer3ReceivedCheque2},
			},
		},
	}
	// verify test cases
	testCheques(t, testCases)
}

func testCheques(t *testing.T, testCases []chequesTestCase) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create a test swap account
			swap, clean := newTestSwap(t, ownerKey, nil)
			defer clean()

			// add test case peers
			peersMapping := make(map[*protocols.Peer]*Peer)
			for _, pp := range tc.protoPeers {
				peer, err := swap.addPeer(pp, common.Address{}, common.Address{})
				if err != nil {
					t.Fatal(err)
				}
				peersMapping[pp] = peer
			}

			// add test case peer sent cheques
			for pp, sc := range tc.sentCheques {
				peer, ok := peersMapping[pp]
				if !ok {
					t.Fatalf("unexpected peer in test case sent cheques")
				}
				err := peer.setLastSentCheque(sc)
				if err != nil {
					t.Fatal(err)
				}
			}
			// add test case store sent cheques
			for p, sc := range tc.storeSentCheques {
				err := swap.saveLastSentCheque(p, sc)
				if err != nil {
					t.Fatal(err)
				}
			}

			// add test case peer received cheques
			for pp, rc := range tc.receivedCheques {
				peer, ok := peersMapping[pp]
				if !ok {
					t.Fatalf("unexpected peer in test case received cheques")
				}
				err := peer.setLastReceivedCheque(rc)
				if err != nil {
					t.Fatal(err)
				}
			}
			// add test case store received cheques
			for p, rc := range tc.storeReceivedCheques {
				err := swap.saveLastReceivedCheque(p, rc)
				if err != nil {
					t.Fatal(err)
				}
			}

			// verify results by calling Cheques function
			cheques, err := swap.Cheques()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expectedCheques, cheques) {
				t.Fatalf("expected cheques to be %v, but are %v", tc.expectedCheques, cheques)
			}
		})
	}
}

// TestPeerCheques verifies that sent and received cheques data for a given peer is correct
func TestPeerCheques(t *testing.T) {
	// generate peers and cheques
	// peer 1
	testPeer := newDummyPeer().Peer
	testPeerSentCheque := newRandomTestCheque()
	testPeerReceivedCheque := newRandomTestCheque()
	// peer 2
	testPeer2 := newDummyPeer().Peer
	testPeer2ReceivedCheque := newRandomTestCheque()

	// build test cases
	testCases := []peerChequesTestCase{
		{
			name:            "peer 1 with no cheques",
			peer:            testPeer,
			sentCheque:      nil,
			receivedCheque:  nil,
			expectedCheques: PeerCheques{nil, nil},
		},
		{
			name:            "peer 1 with sent cheque",
			peer:            testPeer,
			sentCheque:      testPeerSentCheque,
			receivedCheque:  nil,
			expectedCheques: PeerCheques{testPeerSentCheque, nil},
		},
		{
			name:            "peer 1 with sent and received cheque",
			peer:            testPeer,
			sentCheque:      testPeerSentCheque,
			receivedCheque:  testPeerReceivedCheque,
			expectedCheques: PeerCheques{testPeerSentCheque, testPeerReceivedCheque},
		},
		{
			name:            "peer 2 with received cheque",
			peer:            testPeer2,
			sentCheque:      nil,
			receivedCheque:  testPeer2ReceivedCheque,
			expectedCheques: PeerCheques{nil, testPeer2ReceivedCheque},
		},
	}
	// verify test cases
	testPeerCheques(t, testCases)

	// verify cases for disconnected peers
	testPeer3ID := newDummyPeer().Peer.ID()
	testPeer3SentCheque := newRandomTestCheque()
	testPeer3ReceivedCheque := newRandomTestCheque()
	testPeer3ExpectedCheques := PeerCheques{testPeer3SentCheque, testPeer3ReceivedCheque}
	testPeerChequesDisconnected(t, testPeer3ID, testPeer3SentCheque, testPeer3ReceivedCheque, testPeer3ExpectedCheques)

	// verify cases for invalid peers
	invalidPeers := []enode.ID{adapters.RandomNodeConfig().ID, {}}
	testPeerChequesInvalid(t, invalidPeers)
}

func testPeerCheques(t *testing.T, testCases []peerChequesTestCase) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create a test swap account
			swap, clean := newTestSwap(t, ownerKey, nil)
			defer clean()

			// add test case peer
			peer, err := swap.addPeer(tc.peer, common.Address{}, common.Address{})
			if err != nil {
				t.Fatal(err)
			}

			// add test case peer sent cheque
			if tc.sentCheque != nil {
				err = peer.setLastSentCheque(tc.sentCheque)
				if err != nil {
					t.Fatal(err)
				}
			}

			// add test case peer received cheque
			if tc.receivedCheque != nil {
				err = peer.setLastReceivedCheque(tc.receivedCheque)
				if err != nil {
					t.Fatal(err)
				}
			}

			// verify results
			verifyCheques(t, swap, peer.ID(), tc.expectedCheques)
		})
	}
}

func testPeerChequesDisconnected(t *testing.T, peerID enode.ID, sentCheque *Cheque, receivedCheque *Cheque, expectedCheques PeerCheques) {
	t.Helper()
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// add store sent cheque
	err := swap.saveLastSentCheque(peerID, sentCheque)
	if err != nil {
		t.Fatal(err)
	}

	// add store received cheque
	err = swap.saveLastReceivedCheque(peerID, receivedCheque)
	if err != nil {
		t.Fatal(err)
	}

	verifyCheques(t, swap, peerID, expectedCheques)
}

func testPeerChequesInvalid(t *testing.T, invalidPeerIDs []enode.ID) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey, nil)
	defer clean()

	// verify results by calling PeerCheques function
	for _, invalidPeerID := range invalidPeerIDs {
		verifyCheques(t, swap, invalidPeerID, PeerCheques{nil, nil})
	}
}

// compares the result of the PeerCheques function with the expected parameter
func verifyCheques(t *testing.T, s *Swap, peer enode.ID, expectedCheques PeerCheques) {
	peerCheques, err := s.PeerCheques(peer)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedCheques, peerCheques) {
		t.Fatalf("Expected peer %v cheques to be %v, but are %v", peer, expectedCheques, peerCheques)
	}
}
