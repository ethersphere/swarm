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
	"github.com/ethersphere/swarm/testutil"
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

// Test getting a peer's balance
func TestPeerBalance(t *testing.T) {
	// create a test swap account
	swap, testDir := newTestSwap(t)
	defer os.RemoveAll(testDir)

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

func TestAllBalances(t *testing.T) {
	// create a test swap account
	swap, testDir := newTestSwap(t)
	defer os.RemoveAll(testDir)

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
	balances, err := swap.Balances()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(balances, expectedBalances) {
		t.Fatalf("Expected node's balances to be %d, but are %d", expectedBalances, balances)
	}
}

// Test that repeated bookings do correct accounting
func TestRepeatedBookings(t *testing.T) {
	// create a test swap account
	swap, testDir := newTestSwap(t)
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
	creditorSwap, testDir1 := newTestSwap(t)
	debitorSwap, testDir2 := newTestSwap(t)
	defer os.RemoveAll(testDir1)
	defer os.RemoveAll(testDir2)

	ctx := context.Background()
	// deploying would strictly speaking not be necessary, as the signing would also just work
	// with empty contract addresses. Nevertheless to avoid later suprises and for
	// coherence and clarity we deploy here so that we get a simulated contract address
	testDeploy(ctx, creditorSwap.backend, creditorSwap)
	testDeploy(ctx, debitorSwap.backend, debitorSwap)
	creditorSwap.backend.(*backends.SimulatedBackend).Commit()
	debitorSwap.backend.(*backends.SimulatedBackend).Commit()

	// create Peer instances
	// NOTE: remember that these are peer instances representing each **a model of the remote peer** for every local node
	// so creditor is the model of the remote mode for the debitor! (and vice versa)
	cPeer := newDummyPeerWithSpec(Spec)
	dPeer := newDummyPeerWithSpec(Spec)
	creditor := NewPeer(cPeer.Peer, debitorSwap, debitorSwap.backend, creditorSwap.owner.address, debitorSwap.owner.Contract)
	debitor := NewPeer(dPeer.Peer, creditorSwap, creditorSwap.backend, debitorSwap.owner.address, debitorSwap.owner.Contract)

	// set balances arbitrarily
	testAmount := int64(DefaultPaymentThreshold + 42)
	creditorSwap.balances[debitor.ID()] = testAmount
	debitorSwap.balances[creditor.ID()] = 0 - testAmount

	// set the peers into each other's list
	creditorSwap.peers[debitor.ID()] = debitor
	debitorSwap.peers[creditor.ID()] = creditor

	// now simulate sending the cheque to the creditor from the debitor
	debitorSwap.sendCheque(creditor.ID())
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
	// ...and trigger message handling on the receiver side (creditor)
	// remember that debitor is the model of the remote node for the creditor...
	err = debitor.handleEmitChequeMsg(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}
	// finally check that the creditor also successfully reset the balances
	if creditorSwap.balances[debitor.ID()] != 0 {
		t.Fatalf("unexpected balance to be 0, but it is %d", creditorSwap.balances[debitor.ID()])
	}
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

// try restoring a balance from state store
// this is simulated by creating a node,
// assigning it an arbitrary balance,
// then closing the state store.
// Then we re-open the state store and check that
// the balance is still the same
func TestRestoreBalanceFromStateStore(t *testing.T) {
	// create a test swap account
	swap, testDir := newTestSwap(t)
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

	// compare the balances
	if tmpBalance != newBalance {
		t.Fatal(fmt.Sprintf("Unexpected balance value after sending cheap message test. Expected balance: %d, balance is: %d",
			tmpBalance, newBalance))
	}
}

// create a test swap account with a backend
// creates a stateStore for persistence and a Swap account
func newTestSwap(t *testing.T) (*Swap, string) {
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

	log.Debug("creating simulated backend")

	gasLimit := uint64(8000000)
	owner := crypto.PubkeyToAddress(key.PublicKey)
	defaultBackend := backends.NewSimulatedBackend(core.GenesisAlloc{
		owner:              {Balance: big.NewInt(1000000000)},
		ownerAddress:       {Balance: big.NewInt(1000000000)},
		beneficiaryAddress: {Balance: big.NewInt(1000000000)},
	}, gasLimit)

	swap := New(stateStore, key, common.Address{}, defaultBackend)
	defaultBackend.Commit()
	return swap, dir
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
	rw := &testutil.DummyMsgRW{}
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

// tests if encodeCheque encodes the cheque as expected
func TestChequeEncode(t *testing.T) {
	expectedCheque := newTestCheque()

	// encode the cheque
	encoded := expectedCheque.Encode()
	// expected value (computed through truffle/js)
	expected := common.Hex2Bytes("4405415b2b8c9f9aa83e151637b8378dd3bcfeddb8d424e9662fe0837fb1d728f1ac97cebb1085fe0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002a0000000000000000000000000000000000000000000000000000000000000000")
	if !bytes.Equal(encoded, expected) {
		t.Fatalf("Unexpected encoding of cheque. Expected encoding: %x, result is: %x",
			expected, encoded)
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
		t.Fatal(fmt.Sprintf("Unexpected sigHash of cheque. Expected: %x, result is: %x",
			expected, hash))
	}
}

// tests if signContent computes the correct signature
func TestSignContent(t *testing.T) {
	// setup test swap object
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

	expectedCheque := newTestCheque()

	var err error

	// set the owner private key to a known key so we always get the same signature
	swap.owner.privateKey = ownerKey

	// sign the cheque
	sig, err := swap.signContent(expectedCheque)
	// expected value (computed through truffle/js)
	expected := testChequeSig
	if err != nil {
		t.Fatal(fmt.Sprintf("Error in signing: %s", err))
	}
	if !bytes.Equal(sig, expected) {
		t.Fatal(fmt.Sprintf("Unexpected signature for cheque. Expected: %x, result is: %x",
			expected, sig))
	}
}

// tests if verifyChequeSig accepts a correct signature
func TestVerifyChequeSig(t *testing.T) {
	expectedCheque := newTestCheque()
	expectedCheque.Sig = testChequeSig

	if err := expectedCheque.VerifySig(ownerAddress); err != nil {
		t.Fatalf("Invalid signature: %v", err)
	}
}

// tests if verifyChequeSig reject a signature produced by another key
func TestVerifyChequeSigWrongSigner(t *testing.T) {
	expectedCheque := newTestCheque()
	expectedCheque.Sig = testChequeSig

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
	expectedCheque.Sig = manipulateSignature(testChequeSig)

	if err := expectedCheque.VerifySig(ownerAddress); err == nil {
		t.Fatal("Valid signature, should have been invalid")
	}
}

// tests if verifyContract accepts an address with the correct bytecode
func TestVerifyContract(t *testing.T) {
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

	// deploy a new swap contract
	opts := bind.NewKeyedTransactor(ownerKey)
	addr, _, _, err := cswap.Deploy(opts, swap.backend, ownerAddress, 0*time.Second)
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	swap.backend.(*backends.SimulatedBackend).Commit()

	if err = swap.verifyContract(context.TODO(), addr); err != nil {
		t.Fatalf("Contract verification failed: %v", err)
	}
}

// tests if verifyContract rejects an address with different bytecode
func TestVerifyContractWrongContract(t *testing.T) {
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

	opts := bind.NewKeyedTransactor(ownerKey)

	// we deploy the ECDSA library of OpenZeppelin which has a different bytecode than swap
	addr, _, _, err := contract.DeployECDSA(opts, swap.backend)
	if err != nil {
		t.Fatalf("Error in deploy: %v", err)
	}

	swap.backend.(*backends.SimulatedBackend).Commit()

	// since the bytecode is different this should throw an error
	if err = swap.verifyContract(context.TODO(), addr); err != cswap.ErrNotASwapContract {
		t.Fatalf("Contract verification verified wrong contract: %v", err)
	}
}

// TestContractIntegration tests a end-to-end cheque interaction.
// First a simulated backend is created, then we deploy the issuer's swap contract.
// We issue a test cheque with the beneficiary address and on the issuer's contract,
// and immediately try to cash-in the cheque
func TestContractIntegration(t *testing.T) {

	log.Debug("creating test swap")

	issuerSwap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

	issuerSwap.owner.address = ownerAddress
	issuerSwap.owner.privateKey = ownerKey

	backend := issuerSwap.backend.(*backends.SimulatedBackend)

	log.Debug("deploy issuer swap")

	ctx := context.TODO()
	err := testDeploy(ctx, backend, issuerSwap)
	if err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	log.Debug("deployed. signing cheque")

	cheque := newTestCheque()
	cheque.ChequeParams.Contract = issuerSwap.owner.Contract
	cheque.Sig, err = issuerSwap.signContent(cheque)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("sending cheque...")

	opts := bind.NewKeyedTransactor(beneficiaryKey)
	opts.Value = big.NewInt(0)
	opts.Context = ctx

	tx, err := issuerSwap.contractReference.Instance.SubmitChequeBeneficiary(
		opts,
		big.NewInt(int64(cheque.Serial)),
		big.NewInt(int64(cheque.Amount)),
		big.NewInt(int64(cheque.Timeout)),
		cheque.Sig)

	backend.Commit()
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("getting receipt")
	receipt, err := backend.TransactionReceipt(context.TODO(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}

	// check if success
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}

	log.Debug("check cheques state")

	// check state, check that cheque is indeed there
	result, err := issuerSwap.contractReference.Instance.Cheques(nil, beneficiaryAddress)
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
	backend.AdjustTime(30 * time.Second)

	payoutAmount := int64(20)
	// test cashing in, for this we need balance in the contract
	// => send some money
	log.Debug("send money to contract")
	depoTx := types.NewTransaction(
		1,
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
	backend.SendTransaction(context.TODO(), depoTxs)

	log.Debug("cash-in the cheque")
	tx, err = issuerSwap.contractReference.Instance.CashChequeBeneficiary(opts, beneficiaryAddress, big.NewInt(payoutAmount))
	if err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	log.Debug("check tx receipt")
	receipt, err = backend.TransactionReceipt(context.TODO(), tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}

	// check again the status, check paid out is increase by amount
	result, err = issuerSwap.contractReference.Instance.Cheques(nil, beneficiaryAddress)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("cheques result", "result", result)
	if result.PaidOut.Int64() != payoutAmount {
		t.Fatalf("Expected paid out amount to be %d, but is %d", payoutAmount, result.PaidOut)
	}
}

// deploy for testing (needs simulated backend commit)
func testDeploy(ctx context.Context, backend cswap.Backend, swap *Swap) (err error) {
	opts := bind.NewKeyedTransactor(swap.owner.privateKey)
	opts.Value = big.NewInt(int64(swap.params.InitialDepositAmount))
	opts.Context = ctx

	if swap.owner.Contract, swap.contractReference, _, err = cswap.Deploy(opts, backend, swap.owner.address, defaultHarddepositTimeoutDuration); err != nil {
		return err
	}
	return nil
}

// TestSaveAndLoadLastReceivedCheque tests if a saved last received cheque can be loaded again later using the swap functions
func TestSaveAndLoadLastReceivedCheque(t *testing.T) {
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

	testID := newDummyPeer().Peer.ID()
	testCheque := newTestCheque()

	if err := swap.saveLastReceivedCheque(testID, testCheque); err != nil {
		t.Fatalf("Error while saving: %s", err.Error())
	}

	returnedCheque := swap.loadLastReceivedCheque(testID)

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
func newTestSwapAndPeer(t *testing.T) (*Swap, *Peer, string) {
	swap, dir := newTestSwap(t)
	// owner address is the beneficary (counterparty) for the peer
	// that's because we expect cheques we receive to be signed by the address we would issue cheques to
	peer := NewPeer(newDummyPeer().Peer, swap, nil, ownerAddress, testChequeContract)
	// we need to adjust the owner address on swap because we will issue cheques to beneficiaryAddress
	swap.owner.address = beneficiaryAddress
	return swap, peer, dir
}

// TestPeerSaveAndLoadLastReceivedCheque tests if a saved last received cheque can be loaded again later using the peer functions
func TestPeerSaveAndLoadLastReceivedCheque(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	testCheque := newTestCheque()

	if err := peer.saveLastReceivedCheque(testCheque); err != nil {
		t.Fatalf("Error while saving: %s", err.Error())
	}

	returnedCheque := peer.loadLastReceivedCheque()

	if returnedCheque == nil {
		t.Fatal("Could not find saved cheque")
	}

	if !returnedCheque.Equal(testCheque) {
		t.Fatal("Returned cheque was different")
	}
}

// TestPeerVerifyChequeProperties tests that verifyChequeProperties will accept a valid cheque
func TestPeerVerifyChequeProperties(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	testCheque := newTestCheque()
	testCheque.Sig = testChequeSig

	if err := peer.verifyChequeProperties(testCheque); err != nil {
		t.Fatalf("failed to verify cheque properties: %s", err.Error())
	}
}

// TestPeerVerifyChequeProperties tests that verifyChequeProperties will reject invalid cheques
func TestPeerVerifyChequePropertiesInvalidCheque(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	// cheque with an invalid signature
	testCheque := newTestCheque()
	testCheque.Sig = manipulateSignature(testChequeSig)
	if err := peer.verifyChequeProperties(testCheque); err == nil {
		t.Fatalf("accepted cheque with invalid signature")
	}

	// cheque with wrong contract
	testCheque = newTestCheque()
	testCheque.Contract = beneficiaryAddress
	testCheque.Sig, _ = testCheque.Sign(ownerKey)
	if err := peer.verifyChequeProperties(testCheque); err == nil {
		t.Fatalf("accepted cheque with wrong contract")
	}

	// cheque with wrong beneficiary
	testCheque = newTestCheque()
	testCheque.Beneficiary = ownerAddress
	testCheque.Sig, _ = testCheque.Sign(ownerKey)
	if err := peer.verifyChequeProperties(testCheque); err == nil {
		t.Fatalf("accepted cheque with wrong beneficiary")
	}

	// cheque with non-zero timeout
	testCheque = newTestCheque()
	testCheque.Timeout = 10
	testCheque.Sig, _ = testCheque.Sign(ownerKey)
	if err := peer.verifyChequeProperties(testCheque); err == nil {
		t.Fatalf("accepted cheque with non-zero timeout")
	}
}

// TestPeerVerifyChequeAgainstLast tests that verifyChequeAgainstLast accepts a cheque with higher serial and amount
func TestPeerVerifyChequeAgainstLast(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	increase := uint64(10)
	oldCheque := newTestCheque()
	newCheque := newTestCheque()

	newCheque.Serial = oldCheque.Serial + 1
	newCheque.Amount = oldCheque.Amount + increase

	if err := peer.verifyChequeAgainstLast(newCheque, oldCheque, increase); err != nil {
		t.Fatalf("failed to verify cheque compared to old cheque: %s", err.Error())
	}
}

// TestPeerVerifyChequeAgainstLastInvalid tests that verifyChequeAgainstLast rejects cheques with lower serial or amount or an unexpected value
func TestPeerVerifyChequeAgainstLastInvalid(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	increase := uint64(10)

	// cheque with higher amount but same serial
	oldCheque := newTestCheque()
	newCheque := newTestCheque()
	newCheque.Amount = oldCheque.Amount + increase

	if err := peer.verifyChequeAgainstLast(newCheque, oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with same serial")
	}

	// cheque with higher serial but same amount
	oldCheque = newTestCheque()
	newCheque = newTestCheque()
	newCheque.Serial = oldCheque.Serial + 1

	if err := peer.verifyChequeAgainstLast(newCheque, oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with same amount")
	}

	// cheque with amount != increase
	oldCheque = newTestCheque()
	newCheque = newTestCheque()
	newCheque.Serial = oldCheque.Serial + 1
	newCheque.Amount = oldCheque.Amount + increase + 5

	if err := peer.verifyChequeAgainstLast(newCheque, oldCheque, increase); err == nil {
		t.Fatal("accepted a cheque with unexpected amount")
	}
}

// TestPeerProcessAndVerifyCheque tests that processAndVerifyCheque accepts a valid cheque and also saves it
func TestPeerProcessAndVerifyCheque(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	// create test cheque and process
	cheque := newTestCheque()
	cheque.Sig, _ = cheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(cheque); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	// verify that it was indeed saved
	if peer.loadLastReceivedCheque().Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
	}

	// create another cheque with higher serial and amount
	otherCheque := newTestCheque()
	otherCheque.Serial = cheque.Serial + 1
	otherCheque.Amount = cheque.Amount + 10
	otherCheque.Honey = 10
	otherCheque.Sig, _ = otherCheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(otherCheque); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	// verify that it was indeed saved
	if peer.loadLastReceivedCheque().Serial != otherCheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, otherCheque.Serial)
	}
}

// TestPeerProcessAndVerifyChequeInvalid verifies that processAndVerifyCheque does not accept cheques incompatible with the last cheque
// it first tries to process an invalid cheque
// then it processes a valid cheque
// then rejects one with lower serial
// then rejects one with lower amount
func TestPeerProcessAndVerifyChequeInvalid(t *testing.T) {
	_, peer, dir := newTestSwapAndPeer(t)
	defer os.RemoveAll(dir)

	// invalid cheque because wrong recipient
	cheque := newTestCheque()
	cheque.Beneficiary = ownerAddress
	cheque.Sig, _ = cheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(cheque); err == nil {
		t.Fatal("accecpted an invalid cheque as first cheque")
	}

	// valid cheque with serial 5
	cheque = newTestCheque()
	cheque.Serial = 5
	cheque.Sig, _ = cheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(cheque); err != nil {
		t.Fatalf("failed to process cheque: %s", err)
	}

	if peer.loadLastReceivedCheque().Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
	}

	// invalid cheque because serial is lower
	otherCheque := newTestCheque()
	otherCheque.Serial = cheque.Serial - 1
	otherCheque.Amount = cheque.Amount + 10
	otherCheque.Honey = 10
	otherCheque.Sig, _ = otherCheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(otherCheque); err == nil {
		t.Fatal("accepted a cheque with lower serial")
	}

	// invalid cheque because amount is lower
	otherCheque = newTestCheque()
	otherCheque.Serial = cheque.Serial + 1
	otherCheque.Amount = cheque.Amount - 10
	otherCheque.Honey = 10
	otherCheque.Sig, _ = otherCheque.Sign(ownerKey)

	if err := peer.processAndVerifyCheque(otherCheque); err == nil {
		t.Fatal("accepted a cheque with lower amount")
	}

	// check that no invalid cheque was saved
	if peer.loadLastReceivedCheque().Serial != cheque.Serial {
		t.Fatalf("last received cheque has wrong serial, was: %d, expected: %d", peer.lastReceivedCheque.Serial, cheque.Serial)
	}
}
