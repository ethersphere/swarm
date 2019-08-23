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

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
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
	testDeploy(ctx, swap.backend, swap)
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester := p2ptest.NewProtocolTester(swap.owner.privateKey, 2, swap.run)

	// shortcut to creditor node
	debitor := protocolTester.Nodes[0]
	creditor := protocolTester.Nodes[1]

	// set balance artifially
	swap.balances[creditor.ID()] = -42

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
				Code: 1,
				Msg: &EmitChequeMsg{
					Cheque: cheque,
				},
				Peer: debitor.ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 0,
				Msg: &HandshakeMsg{
					ContractAddress: swap.owner.Contract,
				},
				Peer: creditor.ID(),
			},
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
	testDeploy(ctx, creditorSwap.backend, creditorSwap)
	testDeploy(ctx, debitorSwap.backend, debitorSwap)
	creditorSwap.backend.(*backends.SimulatedBackend).Commit()
	debitorSwap.backend.(*backends.SimulatedBackend).Commit()

	log.Debug("create peer instances")

	// create the debitor peer
	dPtpPeer := p2p.NewPeer(enode.ID{}, "debitor", []p2p.Cap{})
	dProtoPeer := protocols.NewPeer(dPtpPeer, nil, Spec)
	debitor := NewPeer(dProtoPeer, creditorSwap, debitorSwap.owner.address, debitorSwap.owner.Contract)

	// set balance artificially
	creditorSwap.balances[debitor.ID()] = 42
	log.Debug("balance", "balance", creditorSwap.balances[debitor.ID()])
	// a safe check: at this point no cheques should be in the swap
	if len(creditorSwap.cheques) != 0 {
		t.Fatalf("Expected no cheques at creditor, but there are %d:", len(creditorSwap.cheques))
	}

	log.Debug("create a cheque")
	var err error
	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:    debitorSwap.owner.Contract,
			Beneficiary: creditorSwap.owner.address,
			Amount:      42,
			Honey:       42,
			Timeout:     0,
		},
	}
	cheque.Signature, err = cheque.Sign(debitorSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	emitMsg := &EmitChequeMsg{
		Cheque: cheque,
	}

	err = creditorSwap.handleEmitChequeMsg(ctx, debitor, emitMsg)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("balance", "balance", creditorSwap.balances[debitor.ID()])
	// check that the balance has been reset
	if creditorSwap.balances[debitor.ID()] != 0 {
		t.Fatalf("Expected debitor balance to have been reset to %d, but it is %d", 0, creditorSwap.balances[debitor.ID()])
	}
	/*
			TODO: This test actually fails now, because the two Swaps create independent backends,
			thus when handling the cheque, it will actually complain (check ERROR log output)
			with `error="no contract code at given address"`.
			Therefore, the `lastReceivedCheque` is not being saved, and this check would fail.
			So TODO is to find out how to address this (should be by having same backend when creating the Swap)
		if creditorSwap.loadLastReceivedCheque(debitor.ID()) != cheque {
			t.Fatalf("Expected exactly one cheque at creditor, but there are %d:", len(creditorSwap.cheques))
		}
	*/
}

// TestTriggerPaymentThreshold is to test that the whole cheque protocol is triggered
// when we reach the payment threshold
// It is the debitor who triggers cheques
func TestTriggerPaymentThreshold(t *testing.T) {
	log.Debug("create test swap")
	debitorSwap, clean := newTestSwap(t, ownerKey)
	defer clean()

	// create a dummy pper
	cPeer := newDummyPeerWithSpec(Spec)
	creditor := NewPeer(cPeer.Peer, debitorSwap, common.Address{}, common.Address{})
	// set the creditor as peer into the debitor's swap
	debitorSwap.peers[creditor.ID()] = creditor

	// set the balance to manually be at PaymentThreshold
	overDraft := 42
	debitorSwap.balances[creditor.ID()] = 0 - DefaultPaymentThreshold

	// we expect a cheque at the end of the test, but not yet
	lenCheques := len(debitorSwap.cheques)
	if lenCheques != 0 {
		t.Fatalf("Expected no cheques yet, but there are %d", lenCheques)
	}
	// do some accounting, no error expected, just a WARN
	err := debitorSwap.Add(int64(-overDraft), creditor.Peer)
	if err != nil {
		t.Fatal(err)
	}

	// we should now have a cheque
	lenCheques = len(debitorSwap.cheques)
	if lenCheques != 1 {
		t.Fatalf("Expected one cheque, but there are %d", lenCheques)
	}
	cheque := debitorSwap.cheques[creditor.ID()]
	expectedAmount := uint64(overDraft) + DefaultPaymentThreshold
	if cheque.Amount != expectedAmount {
		t.Fatalf("Expected cheque amount to be %d, but is %d", expectedAmount, cheque.Amount)
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
	debitor := NewPeer(cPeer.Peer, creditorSwap, common.Address{}, common.Address{})
	// set the debitor as peer into the creditor's swap
	creditorSwap.peers[debitor.ID()] = debitor

	// set the balance to manually be at DisconnectThreshold
	overDraft := 42
	expectedBalance := int64(DefaultDisconnectThreshold)
	// we don't expect any change after the test
	creditorSwap.balances[debitor.ID()] = expectedBalance
	// we also don't expect any cheques yet
	lenCheques := len(creditorSwap.cheques)
	if lenCheques != 0 {
		t.Fatalf("Expected no cheques yet, but there are %d", lenCheques)
	}
	// now do some accounting
	err := creditorSwap.Add(int64(overDraft), debitor.Peer)
	// it should fail due to overdraft
	if err == nil {
		t.Fatal("Expected an error due to overdraft, but did not get any")
	}
	// no balance change expected
	if creditorSwap.balances[debitor.ID()] != expectedBalance {
		t.Fatalf("Expected balance to be %d, but is %d", expectedBalance, creditorSwap.balances[debitor.ID()])
	}
	// still no cheques expected
	lenCheques = len(creditorSwap.cheques)
	if lenCheques != 0 {
		t.Fatalf("Expected still no cheque, but there are %d", lenCheques)
	}

	// let's do the whole thing again (actually a bit silly, it's somehow simulating the peer would have been dropped)
	err = creditorSwap.Add(int64(overDraft), debitor.Peer)
	if err == nil {
		t.Fatal("Expected an error due to overdraft, but did not get any")
	}

	if creditorSwap.balances[debitor.ID()] != expectedBalance {
		t.Fatalf("Expected balance to be %d, but is %d", expectedBalance, creditorSwap.balances[debitor.ID()])
	}

	lenCheques = len(creditorSwap.cheques)
	if lenCheques != 0 {
		t.Fatalf("Expected still no cheque, but there are %d", lenCheques)
	}
}
