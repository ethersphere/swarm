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
	"crypto/ecdsa"
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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	cswap "github.com/ethersphere/swarm/contracts/swap"
	"github.com/ethersphere/swarm/contracts/swap/contract"
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
	testChequeSig      = common.Hex2Bytes("fd3f73c7a708bb4e42471b76dabee2a0c1b9af29efb7eadb37f206bf871b81cf0c7987ad89633be930a63eba9e793cc77896131de7d9740b49da80c23c217c621c")
	testChequeContract = common.HexToAddress("0x4405415b2B8c9F9aA83E151637B8378dD3bcfEDD")
	gasLimit           = uint64(8000000)
	testBackend        *swapTestBackend
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
	// the async submit and cashing go routine needs synchronization for tests
	submitDone chan struct{}
}

func init() {
	flag.Parse()
	mrand.Seed(time.Now().UnixNano())

	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))

	// create a single backend for all tests
	testBackend = newTestBackend()
	// commit the initial "pre-mined" accounts (issuer and beneficiary addresses)
	testBackend.Commit()
}

// newTestBackend creates a new test backend instance
func newTestBackend() *swapTestBackend {
	defaultBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		ownerAddress:       {Balance: big.NewInt(1000000000)},
		beneficiaryAddress: {Balance: big.NewInt(1000000000)},
	}, gasLimit)
	return &swapTestBackend{
		SimulatedBackend: defaultBackend,
	}
}

// Test getting a peer's balance
func TestPeerBalance(t *testing.T) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	// test for correct value
	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = 888
	b, err := swap.Balance(testPeer.ID())
	if err != nil {
		t.Fatal(err)
	}
	if b != 888 {
		t.Fatalf("Expected peer's balance to be %d, but is %d", 888, b)
	}

	// test for inexistent node
	id := adapters.RandomNodeConfig().ID
	_, err = swap.Balance(id)
	if err == nil {
		t.Fatal("Expected call to fail, but it didn't!")
	}
	if err != state.ErrNotFound {
		t.Fatalf("Expected test to fail with %s, but is %s", "ErrorNotFound", err.Error())
	}
}

// Test getting balances for all known peers
func TestAllBalances(t *testing.T) {
	// create a test swap account
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	if len(swap.balances) != 0 {
		t.Fatalf("Expected balances to be empty, but are %v", swap.balances)
	}

	// test balance addition for peer
	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = 808
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 808})

	// test successive balance addition for peer
	testPeer2 := newDummyPeer()
	swap.balances[testPeer2.ID()] = 909
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 808, testPeer2.ID(): 909})

	// test balance change for peer
	swap.balances[testPeer.ID()] = 303
	testBalances(t, swap, map[enode.ID]int64{testPeer.ID(): 303, testPeer2.ID(): 909})
}

func testBalances(t *testing.T, swap *Swap, expectedBalances map[enode.ID]int64) {
	t.Helper()
	balances, err := swap.Balances()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(balances, expectedBalances) {
		t.Fatalf("Expected node's balances to be %d, but are %d", expectedBalances, balances)
	}
}

type storeKeysTestCases struct {
	nodeID                    enode.ID
	expectedBalanceKey        string
	expectedSentChequeKey     string
	expectedReceivedChequeKey string
}

// Test the getting balance and cheques store keys based on a node ID, and the reverse process as well
func TestStoreKeys(t *testing.T) {
	testCases := []storeKeysTestCases{
		{enode.HexID("f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56"), "balance_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "sent_cheque_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56", "received_cheque_f6876a1f73947b0495d36e648aeb74f952220c3b03e66a1cc786863f6104fa56"},
		{enode.HexID("93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c"), "balance_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "sent_cheque_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c", "received_cheque_93a3309412ff6204ec9b9469200742f62061932009e744def79ef96492673e6c"},
		{enode.HexID("c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44"), "balance_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "sent_cheque_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44", "received_cheque_c19ecf22f02f77f4bb320b865d3f37c6c592d32a1c9b898efb552a5161a1ee44"},
	}
	testStoreKeys(t, testCases)
}

func testStoreKeys(t *testing.T, testCases []storeKeysTestCases) {
	for _, testCase := range testCases {
		t.Run(fmt.Sprint(testCase.nodeID), func(t *testing.T) {
			actualBalanceKey := balanceKey(testCase.nodeID)
			actualSentChequeKey := sentChequeKey(testCase.nodeID)
			actualReceivedChequeKey := receivedChequeKey(testCase.nodeID)

			if actualBalanceKey != testCase.expectedBalanceKey {
				t.Fatalf("Expected balance key to be %s, but is %s instead.", testCase.expectedBalanceKey, actualBalanceKey)
			}
			if actualSentChequeKey != testCase.expectedSentChequeKey {
				t.Fatalf("Expected sent cheque key to be %s, but is %s instead.", testCase.expectedSentChequeKey, actualSentChequeKey)
			}
			if actualReceivedChequeKey != testCase.expectedReceivedChequeKey {
				t.Fatalf("Expected received cheque key to be %s, but is %s instead.", testCase.expectedReceivedChequeKey, actualReceivedChequeKey)
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
	s, clean := newTestSwap(t, ownerKey)
	defer clean()

	var err error

	// modify balances in memory but not in store
	testPeerID := enode.HexID("8418f4eeb20735630cfff00459b7fc7ec1674c9f77f19bab23e895706bfd1032")
	s.balances[testPeerID] = 144
	comparePeerBalance(t, s, testPeerID, 0)

	// modify balances both in memory and in store
	peerBalance, err := s.updateBalance(testPeerID, 29)
	if err != nil {
		t.Fatal(err)
	}
	// store balance for peer should match
	comparePeerBalance(t, s, testPeerID, peerBalance)

	// update balances for second peer
	testPeer2ID := enode.HexID("fdbb55b4d9b0011c93e736bb1b736013943890f2a08640f89de00e738a8b7986")
	s.balances[testPeer2ID] = 144
	peer2Balance, err := s.updateBalance(testPeer2ID, -76)
	if err != nil {
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
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

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
		{int64(mrand.Intn(100)), testPeer2.Peer},
		{int64(0 - mrand.Intn(55)), testPeer2.Peer},
		{int64(0 - mrand.Intn(999)), testPeer2.Peer},
	}
	addBookings(swap, mixedBookings)
	verifyBookings(t, swap, append(bookings, mixedBookings...))
}

// TestResetBalance tests that balances are correctly reset
// The test deploys creates swap instances for each node,
// deploys simulated contracts, sets the balance of each
// other node to some arbitrary number above thresholds,
// and then calls both `sendCheque` on one side and
// `handleEmitChequeMsg` in order to simulate a roundtrip
// and see that both have reset the balance correctly
func TestResetBalance(t *testing.T) {
	// create both test swap accounts
	creditorSwap, clean1 := newTestSwap(t, beneficiaryKey)
	debitorSwap, clean2 := newTestSwap(t, ownerKey)
	defer clean1()
	defer clean2()

	ctx := context.Background()
	// deploying would strictly speaking not be necessary, as the signing would also just work
	// with empty contract addresses. Nevertheless to avoid later suprises and for
	// coherence and clarity we deploy here so that we get a simulated contract address
	testDeploy(ctx, creditorSwap.backend, creditorSwap)
	testBackend.Commit()
	testDeploy(ctx, debitorSwap.backend, debitorSwap)
	testBackend.Commit()

	// create Peer instances
	// NOTE: remember that these are peer instances representing each **a model of the remote peer** for every local node
	// so creditor is the model of the remote mode for the debitor! (and vice versa)
	cPeer := newDummyPeerWithSpec(Spec)
	dPeer := newDummyPeerWithSpec(Spec)
	creditor := NewPeer(cPeer.Peer, debitorSwap, creditorSwap.owner.address, debitorSwap.owner.Contract)
	debitor := NewPeer(dPeer.Peer, creditorSwap, debitorSwap.owner.address, debitorSwap.owner.Contract)

	// set balances arbitrarily
	testAmount := int64(DefaultPaymentThreshold + 42)
	creditorSwap.balances[debitor.ID()] = testAmount
	debitorSwap.balances[creditor.ID()] = 0 - testAmount

	// set the peers into each other's list
	creditorSwap.peers[debitor.ID()] = debitor
	debitorSwap.peers[creditor.ID()] = creditor

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// now simulate sending the cheque to the creditor from the debitor
	debitorSwap.sendCheque(creditor)
	// the debitor should have already reset its balance
	if debitorSwap.balances[creditor.ID()] != 0 {
		t.Fatalf("unexpected balance to be 0, but it is %d", debitorSwap.balances[creditor.ID()])
	}

	var err error
	// now load the cheque that the debitor created...
	cheque := debitorSwap.cheques[creditor.ID()]
	if cheque == nil {
		t.Fatal("expected to find a cheque, but it was empty")
	}
	// ...create a message...
	msg := &EmitChequeMsg{
		Cheque: cheque,
	}
	// now we need to create the channel...
	testBackend.submitDone = make(chan struct{})
	// ...and trigger message handling on the receiver side (creditor)
	// remember that debitor is the model of the remote node for the creditor...
	err = creditorSwap.handleEmitChequeMsg(ctx, debitor, msg)
	if err != nil {
		t.Fatal(err)
	}
	// ...on which we wait until the submitChequeAndCash is actually terminated (ensures proper nounce count)
	select {
	case <-testBackend.submitDone:
		log.Debug("submit and cash transactions completed and committed")
	case <-time.After(4 * time.Second):
		t.Fatalf("Timeout waiting for submit and cash transactions to complete")
	}
	// finally check that the creditor also successfully reset the balances
	if creditorSwap.balances[debitor.ID()] != 0 {
		t.Fatalf("unexpected balance to be 0, but it is %d", creditorSwap.balances[debitor.ID()])
	}
}

// generate bookings based on parameters, apply them to a Swap struct and verify the result
// append generated bookings to slice pointer
func testPeerBookings(t *testing.T, swap *Swap, bookings *[]booking, bookingAmount int64, bookingQuantity int, peer *protocols.Peer) {
	t.Helper()
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
	t.Helper()
	expectedBalances := calculateExpectedBalances(swap, bookings)
	realBalances := swap.balances
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
		// balance is not expected to be affected once past the disconnect threshold
		if peerBalance < swap.disconnectThreshold {
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
	// create a test swap account
	swap, testDir := newBaseTestSwap(t, ownerKey)
	defer os.RemoveAll(testDir)

	testPeer := newDummyPeer()
	swap.balances[testPeer.ID()] = -8888

	tmpBalance := swap.balances[testPeer.ID()]
	swap.store.Put(testPeer.ID().String(), &tmpBalance)

	err := swap.store.Close()
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
	stateStore.Get(testPeer.ID().String(), &newBalance)

	// compare the balances
	if tmpBalance != newBalance {
		t.Fatalf("Unexpected balance value after sending cheap message test. Expected balance: %d, balance is: %d", tmpBalance, newBalance)
	}
}

// During tests, because the submit and cashing in of cheques are async, we should wait for the function to be returned
// Otherwise if we call `handleEmitChequeMsg` manually, it will return before the TX has been committed to the `SimulatedBackend`,
// causing subsequent TX to possibly fail due to nonce mismatch
func testSubmitChequeAndCash(s *Swap, otherSwap cswap.Contract, opts *bind.TransactOpts, actualAmount uint64, cheque *Cheque) {
	submitChequeAndCash(s, otherSwap, opts, actualAmount, cheque)
	// close the channel, signals to clients that this function actually finished
	if stb, ok := s.backend.(*swapTestBackend); ok {
		if stb.submitDone != nil {
			close(stb.submitDone)
		}
	}
}

// create a test swap account with a backend
// creates a stateStore for persistence and a Swap account
func newBaseTestSwap(t *testing.T, key *ecdsa.PrivateKey) (*Swap, string) {
	t.Helper()
	dir, err := ioutil.TempDir("", "swap_test_store")
	if err != nil {
		t.Fatal(err)
	}
	stateStore, err2 := state.NewDBStore(dir)
	if err2 != nil {
		t.Fatal(err2)
	}
	log.Debug("creating simulated backend")

	swap := New(stateStore, key, common.Address{}, testBackend)
	return swap, dir
}

// create a test swap account with a backend
// creates a stateStore for persistence and a Swap account
// returns a cleanup function
func newTestSwap(t *testing.T, key *ecdsa.PrivateKey) (*Swap, func()) {
	t.Helper()
	swap, dir := newBaseTestSwap(t, key)
	clean := func() {
		swap.Close()
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
			Contract:    testChequeContract,
			Serial:      uint64(1),
			Amount:      uint64(42),
			Honey:       uint64(42),
			Timeout:     uint64(0),
			Beneficiary: beneficiaryAddress,
		},
	}

	return cheque
}

// tests if encodeForSignature encodes the cheque as expected
func TestChequeEncodeForSignature(t *testing.T) {
	expectedCheque := newTestCheque()

	// encode the cheque
	encoded := expectedCheque.encodeForSignature()
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("4405415b2b8c9f9aa83e151637b8378dd3bcfeddb8d424e9662fe0837fb1d728f1ac97cebb1085fe0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002a0000000000000000000000000000000000000000000000000000000000000000")
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
	expected := common.Hex2Bytes("291619739fc0008915f09989411d22a29ea62eb39d86ed094ef51d6a420a1358")
	if !bytes.Equal(hash, expected) {
		t.Fatalf("Unexpected sigHash of cheque. Expected: %x, result is: %x", expected, hash)
	}
}

// tests if signContent computes the correct signature
func TestSignContent(t *testing.T) {
	// setup test swap object
	swap, clean := newTestSwap(t, ownerKey)
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

// tests if verifyContract accepts an address with the correct bytecode
func TestVerifyContract(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	// deploy a new swap contract
	opts := bind.NewKeyedTransactor(ownerKey)
	addr, _, _, err := cswap.Deploy(opts, swap.backend, ownerAddress, 0*time.Second)
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	testBackend.Commit()

	if err = swap.verifyContract(context.TODO(), addr); err != nil {
		t.Fatalf("Contract verification failed: %v", err)
	}
}

// tests if verifyContract rejects an address with different bytecode
func TestVerifyContractWrongContract(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	opts := bind.NewKeyedTransactor(ownerKey)

	// we deploy the ECDSA library of OpenZeppelin which has a different bytecode than swap
	addr, _, _, err := contract.DeployECDSA(opts, swap.backend)
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	testBackend.Commit()

	// since the bytecode is different this should throw an error
	if err = swap.verifyContract(context.TODO(), addr); err != cswap.ErrNotASwapContract {
		t.Fatalf("Contract verification verified wrong contract: %v", err)
	}
}

// setupContractTest is a helper function for setting up the
// blockchain wait function for testing
func setupContractTest() func() {
	// we overwrite the waitForTx function with one which the simulated backend
	// immediately commits
	currentWaitFunc := cswap.WaitFunc
	defaultSubmitChequeAndCash = testSubmitChequeAndCash
	// overwrite only for the duration of the test, so...
	cswap.WaitFunc = testWaitForTx
	return func() {
		// ...we need to set it back to original when done
		cswap.WaitFunc = currentWaitFunc
		defaultSubmitChequeAndCash = submitChequeAndCash
	}
}

// TestContractIntegration tests a end-to-end cheque interaction.
// First a simulated backend is created, then we deploy the issuer's swap contract.
// We issue a test cheque with the beneficiary address and on the issuer's contract,
// and immediately try to cash-in the cheque
func TestContractIntegration(t *testing.T) {

	log.Debug("creating test swap")

	issuerSwap, clean := newTestSwap(t, ownerKey)
	defer clean()

	issuerSwap.owner.address = ownerAddress
	issuerSwap.owner.privateKey = ownerKey

	log.Debug("deploy issuer swap")

	ctx := context.TODO()
	err := testDeploy(ctx, testBackend, issuerSwap)
	if err != nil {
		t.Fatal(err)
	}
	testBackend.Commit()

	log.Debug("deployed. signing cheque")

	cheque := newTestCheque()
	cheque.ChequeParams.Contract = issuerSwap.owner.Contract
	cheque.Signature, err = cheque.Sign(issuerSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("sending cheque...")

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	opts := bind.NewKeyedTransactor(beneficiaryKey)
	opts.Value = big.NewInt(0)
	opts.Context = ctx

	receipt, err := issuerSwap.contract.SubmitChequeBeneficiary(
		opts,
		testBackend,
		big.NewInt(int64(cheque.Serial)),
		big.NewInt(int64(cheque.Amount)),
		big.NewInt(int64(cheque.Timeout)),
		cheque.Signature)

	testBackend.Commit()

	// check if success
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}

	log.Debug("check cheques state")

	// check state, check that cheque is indeed there
	result, err := issuerSwap.contract.Cheques(nil, beneficiaryAddress)
	if err != nil {
		t.Fatal(err)
	}
	if result.Serial.Uint64() != cheque.Serial {
		t.Fatalf("Wrong serial %d", result.Serial)
	}
	if result.Amount.Uint64() != cheque.Amount {
		t.Fatalf("Wrong amount %d", result.Amount)
	}
	log.Debug("cheques result", "result", result)

	// go forward in time
	testBackend.AdjustTime(30 * time.Second)

	payoutAmount := int64(20)
	// test cashing in, for this we need balance in the contract
	// => send some money
	log.Debug("send money to contract")
	nonce, err := testBackend.NonceAt(ctx, issuerSwap.owner.address, nil)
	if err != nil {
		t.Fatal(err)
	}
	depoTx := types.NewTransaction(
		nonce,
		issuerSwap.owner.Contract,
		big.NewInt(payoutAmount),
		50000,
		big.NewInt(int64(0)),
		[]byte{},
	)
	depoTxs, err := types.SignTx(depoTx, types.HomesteadSigner{}, issuerSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}
	testBackend.SendTransaction(context.TODO(), depoTxs)

	log.Debug("cash-in the cheque")
	receipt, err = issuerSwap.contract.CashChequeBeneficiary(opts, testBackend, beneficiaryAddress, big.NewInt(payoutAmount))
	testBackend.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}

	// check again the status, check paid out is increase by amount
	result, err = issuerSwap.contract.Cheques(nil, beneficiaryAddress)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("cheques result", "result", result)
	if result.PaidOut.Int64() != payoutAmount {
		t.Fatalf("Expected paid out amount to be %d, but is %d", payoutAmount, result.PaidOut)
	}
}

// when testing, we don't need to wait for a transaction to be mined
func testWaitForTx(auth *bind.TransactOpts, backend cswap.Backend, tx *types.Transaction) (*types.Receipt, error) {

	testBackend.Commit()
	receipt, err := backend.TransactionReceipt(context.TODO(), tx.Hash())
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

// deploy for testing (needs simulated backend commit)
func testDeploy(ctx context.Context, backend cswap.Backend, swap *Swap) (err error) {
	opts := bind.NewKeyedTransactor(swap.owner.privateKey)
	opts.Value = big.NewInt(int64(swap.params.InitialDepositAmount))
	opts.Context = ctx

	swap.owner.Contract, swap.contract, _, err = cswap.Deploy(opts, backend, swap.owner.address, defaultHarddepositTimeoutDuration)

	return err
}

// TestSaveAndLoadLastReceivedCheque tests if a saved last received cheque can be loaded again later using the swap functions
func TestSaveAndLoadLastReceivedCheque(t *testing.T) {
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	testPeer := NewPeer(newDummyPeer().Peer, swap, common.Address{}, common.Address{})
	testCheque := newTestCheque()

	if err := swap.saveLastReceivedCheque(testPeer, testCheque); err != nil {
		t.Fatalf("Error while saving: %s", err.Error())
	}

	returnedCheque := swap.loadLastReceivedCheque(testPeer)

	if returnedCheque == nil {
		t.Fatalf("Could not find saved cheque")
	}

	if !returnedCheque.Equal(testCheque) {
		t.Fatalf("Returned cheque was different")
	}
}

// newTestSwapAndPeer is a helper function to create a swap and a peer instance that fit together
// the owner of this swap is the beneficiaryAddress
// hence the owner of this swap would sign cheques with beneficiaryKey and receive cheques from ownerKey (or another party) which is NOT the owner of this swap
func newTestSwapAndPeer(t *testing.T, key *ecdsa.PrivateKey) (*Swap, *Peer, func()) {
	swap, clean := newTestSwap(t, key)
	// owner address is the beneficiary (counterparty) for the peer
	// that's because we expect cheques we receive to be signed by the address we would issue cheques to
	peer := NewPeer(newDummyPeer().Peer, swap, ownerAddress, testChequeContract)
	// we need to adjust the owner address on swap because we will issue cheques to beneficiaryAddress
	swap.owner.address = beneficiaryAddress
	return swap, peer, clean
}

// TestPeerSaveAndLoadLastReceivedCheque tests if a saved last received cheque can be loaded again later using the peer functions
func TestPeerSaveAndLoadLastReceivedCheque(t *testing.T) {
	swap, peer, clean := newTestSwapAndPeer(t, ownerKey)
	defer clean()

	testCheque := newTestCheque()

	if err := swap.saveLastReceivedCheque(peer, testCheque); err != nil {
		t.Fatalf("Error while saving: %s", err.Error())
	}

	returnedCheque := swap.loadLastReceivedCheque(peer)

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

	// cheque with non-zero timeout
	testCheque = newTestCheque()
	testCheque.Timeout = 10
	testCheque.Signature, _ = testCheque.Sign(ownerKey)
	if err := testCheque.verifyChequeProperties(peer, swap.owner.address); err == nil {
		t.Fatalf("accepted cheque with non-zero timeout")
	}
}

// TestPeerVerifyChequeAgainstLast tests that verifyChequeAgainstLast accepts a cheque with higher serial and amount
func TestPeerVerifyChequeAgainstLast(t *testing.T) {
	increase := uint64(10)
	oldCheque := newTestCheque()
	newCheque := newTestCheque()

	newCheque.Serial = oldCheque.Serial + 1
	newCheque.Amount = oldCheque.Amount + increase

	actualAmount, err := newCheque.verifyChequeAgainstLast(oldCheque, increase)
	if err != nil {
		t.Fatalf("failed to verify cheque compared to old cheque: %v", err)
	}

	if actualAmount != increase {
		t.Fatalf("wrong actual amount, expected: %d, was: %d", increase, actualAmount)
	}
}

// TestPeerVerifyChequeAgainstLastInvalid tests that verifyChequeAgainstLast rejects cheques with lower serial or amount or an unexpected value
func TestPeerVerifyChequeAgainstLastInvalid(t *testing.T) {
	increase := uint64(10)

	// cheque with higher amount but same serial
	oldCheque := newTestCheque()
	newCheque := newTestCheque()
	newCheque.Amount = oldCheque.Amount + increase

	if _, err := newCheque.verifyChequeAgainstLast(oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with same serial")
	}

	// cheque with higher serial but same amount
	oldCheque = newTestCheque()
	newCheque = newTestCheque()
	newCheque.Serial = oldCheque.Serial + 1

	if _, err := newCheque.verifyChequeAgainstLast(oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with same amount")
	}

	// cheque with amount != increase
	oldCheque = newTestCheque()
	newCheque = newTestCheque()
	newCheque.Serial = oldCheque.Serial + 1
	newCheque.Amount = oldCheque.Amount + increase + 5

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

	if actualAmount != cheque.Amount {
		t.Fatalf("computed wrong actual amount: was %d, expected: %d", actualAmount, cheque.Amount)
	}

	// verify that it was indeed saved
	if swap.loadLastReceivedCheque(peer).Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
	}

	// create another cheque with higher serial and amount
	otherCheque := newTestCheque()
	otherCheque.Serial = cheque.Serial + 1
	otherCheque.Amount = cheque.Amount + 10
	otherCheque.Honey = 10
	otherCheque.Signature, _ = otherCheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(otherCheque, peer); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	// verify that it was indeed saved
	if swap.loadLastReceivedCheque(peer).Serial != otherCheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, otherCheque.Serial)
	}
}

// TestPeerProcessAndVerifyChequeInvalid verifies that processAndVerifyCheque does not accept cheques incompatible with the last cheque
// it first tries to process an invalid cheque
// then it processes a valid cheque
// then rejects one with lower serial
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

	// valid cheque with serial 5
	cheque = newTestCheque()
	cheque.Serial = 5
	cheque.Signature, _ = cheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(cheque, peer); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	if swap.loadLastReceivedCheque(peer).Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
	}

	// invalid cheque because serial is lower
	otherCheque := newTestCheque()
	otherCheque.Serial = cheque.Serial - 1
	otherCheque.Amount = cheque.Amount + 10
	otherCheque.Honey = 10
	otherCheque.Signature, _ = otherCheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(otherCheque, peer); err == nil {
		t.Fatal("accepted a cheque with lower serial")
	}

	// invalid cheque because amount is lower
	otherCheque = newTestCheque()
	otherCheque.Serial = cheque.Serial + 1
	otherCheque.Amount = cheque.Amount - 10
	otherCheque.Honey = 10
	otherCheque.Signature, _ = otherCheque.Sign(ownerKey)

	if _, err := swap.processAndVerifyCheque(otherCheque, peer); err == nil {
		t.Fatal("accepted a cheque with lower amount")
	}

	// check that no invalid cheque was saved
	if swap.loadLastReceivedCheque(peer).Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
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
