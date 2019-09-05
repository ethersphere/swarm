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
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethersphere/swarm/p2p/protocols"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
)

/*
TestHandshake creates two mock nodes and initiates an exchange;
it expects a handshake to take place between the two nodes
(the handshake would fail because we don't actually use real nodes here)
*/
func TestHandshake(t *testing.T) {
	var err error

	// setup test swap object
	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	ctx := context.Background()
	err = testDeploy(ctx, swap.backend, swap)
	if err != nil {
		t.Fatal(err)
	}
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester := p2ptest.NewProtocolTester(swap.owner.privateKey, 2, swap.run)

	// shortcut to creditor node
	debitor := protocolTester.Nodes[0]
	creditor := protocolTester.Nodes[1]

	// set balance artifially
	swap.saveBalance(creditor.ID(), -42)

	// create the expected cheque to be received
	cheque := newTestCheque()

	// sign the cheque
	cheque.Signature, err = cheque.Sign(swap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	// run the exchange:
	// trigger a `EmitChequeMsg`
	// expect HandshakeMsg on each node
	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Label: "TestHandshake",
		Triggers: []p2ptest.Trigger{
			{
				Code: 0,
				Msg: &HandshakeMsg{
					ContractAddress: swap.owner.Contract,
				},
				Peer: creditor.ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 0,
				Msg: &HandshakeMsg{
					ContractAddress: swap.owner.Contract,
				},
				Peer: debitor.ID(),
			},
		},
	})

	// there should be no error at this point
	if err != nil {
		t.Fatal(err)
	}
}

// TestEmitCheque is a full round of a cheque exchange between peers via the protocol.
// We create two swap, for the creditor (beneficiary) and debitor (issuer) each,
// and deploy them to the simulated backend.
// We then create Swap protocol peers with a MsgPipe to be able to directly write messages to each other.
// We have the debitor send a cheque via an `EmitChequeMsg`, then the creditor "reads" (pipe) the message
// and handles the cheque.
func TestEmitCheque(t *testing.T) {
	log.Debug("set up test swaps")
	creditorSwap, clean1 := newTestSwap(t, beneficiaryKey)
	debitorSwap, clean2 := newTestSwap(t, ownerKey)
	defer clean1()
	defer clean2()

	ctx := context.Background()

	log.Debug("deploy to simulated backend")
	err := testDeploy(ctx, creditorSwap.backend, creditorSwap)
	if err != nil {
		t.Fatal(err)
	}
	err = testDeploy(ctx, debitorSwap.backend, debitorSwap)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("create peer instances")

	// create the debitor peer
	dPtpPeer := p2p.NewPeer(enode.ID{}, "debitor", []p2p.Cap{})
	dProtoPeer := protocols.NewPeer(dPtpPeer, nil, Spec)
	debitor, err := creditorSwap.addPeer(dProtoPeer, debitorSwap.owner.address, debitorSwap.owner.Contract)
	if err != nil {
		t.Fatal(err)
	}

	// set balance artificially
	debitor.setBalance(42)
	log.Debug("balance", "balance", debitor.getBalance())
	// a safe check: at this point no cheques should be in the swap
	if debitor.getLastReceivedCheque() != nil {
		t.Fatalf("Expected no cheques at creditor, but there is %v:", debitor.getLastReceivedCheque())
	}

	log.Debug("create a cheque")
	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         debitorSwap.owner.Contract,
			Beneficiary:      creditorSwap.owner.address,
			CumulativePayout: 42,
		},
		Honey: 42,
	}
	cheque.Signature, err = cheque.Sign(debitorSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	emitMsg := &EmitChequeMsg{
		Cheque: cheque,
	}
	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// now we need to create the channel...
	testBackend.cashDone = make(chan struct{})
	err = creditorSwap.handleEmitChequeMsg(ctx, debitor, emitMsg)
	if err != nil {
		t.Fatal(err)
	}
	// ...on which we wait until the cashCheque is actually terminated (ensures proper nounce count)
	select {
	case <-testBackend.cashDone:
		log.Debug("cash transaction completed and committed")
	case <-time.After(4 * time.Second):
		t.Fatalf("Timeout waiting for cash transaction to complete")
	}
	log.Debug("balance", "balance", debitor.getBalance())
	// check that the balance has been reset
	if debitor.getBalance() != 0 {
		t.Fatalf("Expected debitor balance to have been reset to %d, but it is %d", 0, debitor.getBalance())
	}
	recvCheque := debitor.getLastReceivedCheque()
	log.Debug("expected cheque", "cheque", recvCheque)
	if recvCheque != cheque {
		t.Fatalf("Expected cheque at creditor, but it was %v:", recvCheque)
	}
}

// TestTriggerPaymentThreshold is to test that the whole cheque protocol is triggered
// when we reach the payment threshold
// It is the debitor who triggers cheques
func TestTriggerPaymentThreshold(t *testing.T) {
	log.Debug("create test swap")
	debitorSwap, clean := newTestSwap(t, ownerKey)
	defer clean()

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// create a dummy pper
	cPeer := newDummyPeerWithSpec(Spec)
	creditor, err := debitorSwap.addPeer(cPeer.Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}

	// set the balance to manually be at PaymentThreshold
	overDraft := 42
	creditor.setBalance(-DefaultPaymentThreshold)

	// we expect a cheque at the end of the test, but not yet
	if creditor.getLastSentCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v:", creditor.getLastSentCheque())
	}
	// do some accounting, no error expected, just a WARN
	err = debitorSwap.Add(int64(-overDraft), creditor.Peer)
	if err != nil {
		t.Fatal(err)
	}

	// we should now have a cheque
	if creditor.getLastSentCheque() == nil {
		t.Fatal("Expected one cheque, but there is none")
	}

	cheque := creditor.getLastSentCheque()
	expectedAmount := uint64(overDraft) + DefaultPaymentThreshold
	if cheque.CumulativePayout != expectedAmount {
		t.Fatalf("Expected cheque cumulative payout to be %d, but is %d", expectedAmount, cheque.CumulativePayout)
	}

	// because no other accounting took place in the meantime the balance should be exactly 0
	if creditor.getBalance() != 0 {
		t.Fatalf("Expected debitorSwap balance to be 0, but is %d", creditor.getBalance())
	}

	// do some accounting again to trigger a second cheque
	if err = debitorSwap.Add(int64(-DefaultPaymentThreshold), creditor.Peer); err != nil {
		t.Fatal(err)
	}

	if creditor.getBalance() != 0 {
		t.Fatalf("Expected debitorSwap balance to be 0, but is %d", creditor.getBalance())
	}
}

// TestTriggerDisconnectThreshold is to test that no further accounting takes place
// when we reach the disconnect threshold
// It is the creditor who triggers the disconnect from a overdraft creditor
func TestTriggerDisconnectThreshold(t *testing.T) {
	log.Debug("create test swap")
	creditorSwap, clean := newTestSwap(t, beneficiaryKey)
	defer clean()

	// create a dummy pper
	cPeer := newDummyPeerWithSpec(Spec)
	debitor, err := creditorSwap.addPeer(cPeer.Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}

	// set the balance to manually be at DisconnectThreshold
	overDraft := 42
	expectedBalance := int64(DefaultDisconnectThreshold)
	// we don't expect any change after the test
	debitor.setBalance(expectedBalance)
	// we also don't expect any cheques yet
	if debitor.getLastSentCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v", debitor.getLastSentCheque())
	}
	// now do some accounting
	err = creditorSwap.Add(int64(overDraft), debitor.Peer)
	// it should fail due to overdraft
	if err == nil {
		t.Fatal("Expected an error due to overdraft, but did not get any")
	}
	// no balance change expected
	if debitor.getBalance() != expectedBalance {
		t.Fatalf("Expected balance to be %d, but is %d", expectedBalance, debitor.getBalance())
	}
	// still no cheques expected
	if debitor.getLastSentCheque() != nil {
		t.Fatalf("Expected still no cheques yet, but there is %v", debitor.getLastSentCheque())
	}

	// let's do the whole thing again (actually a bit silly, it's somehow simulating the peer would have been dropped)
	err = creditorSwap.Add(int64(overDraft), debitor.Peer)
	if err == nil {
		t.Fatal("Expected an error due to overdraft, but did not get any")
	}

	if debitor.getBalance() != expectedBalance {
		t.Fatalf("Expected balance to be %d, but is %d", expectedBalance, debitor.getBalance())
	}

	if debitor.getLastSentCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v", debitor.getLastSentCheque())
	}
}
