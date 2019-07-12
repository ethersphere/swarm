// Copyright 2019 The go-ethereum Authors
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
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
)

/*
TestRequestCheque tests that a peer will respond with a
`EmitChequeMsg` containing a cheque for an expected amount
if it sends a `RequestChequeMsg` is sent to it
*/
func TestRequestCheque(t *testing.T) {
	var err error

	// setup test swap object
	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)

	// dummy object so we can run the protocol
	ss := swap

	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester := p2ptest.NewProtocolTester(swap.owner.privateKey, 2, ss.run)

	// shortcut to creditor node
	creditor := protocolTester.Nodes[0]

	// set balance artifially
	swap.balances[creditor.ID()] = -42

	// create the expected cheque to be received
	// NOTE: this may be improved, as it is essentially running the same
	// code as in production
	expectedCheque := swap.cheques[creditor.ID()]
	expectedCheque = &Cheque{
		ChequeParams: ChequeParams{
			Serial:      uint64(1),
			Amount:      uint64(42),
			Timeout:     defaultCashInDelay,
			Beneficiary: crypto.PubkeyToAddress(*creditor.Pubkey()),
		},
	}

	// sign the cheque
	expectedCheque.Sig, err = swap.signContent(expectedCheque)
	if err != nil {
		t.Fatal(err)
	}

	// run the exchange:
	// trigger a `ChequeRequestMsg`
	// expect a `EmitChequeMsg` with a valid cheque
	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Label: "TestRequestCheque",
		Triggers: []p2ptest.Trigger{
			{
				Code: 0,
				Msg: &ChequeRequestMsg{
					crypto.PubkeyToAddress(*creditor.Pubkey()),
				},
				Peer: creditor.ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 1,
				Msg: &EmitChequeMsg{
					Cheque: expectedCheque,
				},
				Peer: creditor.ID(),
			},
		},
	})

	// there should be no error at this point
	if err != nil {
		t.Fatal(err)
	}

	// now we request a new cheque;
	// the peer though should have already reset the balance,
	// so no new cheque should be issued
	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Label: "TestRequestNoCheque",
		Triggers: []p2ptest.Trigger{
			{
				Code: 0,
				Msg: &ChequeRequestMsg{
					crypto.PubkeyToAddress(*creditor.Pubkey()),
				},
				Peer: creditor.ID(),
			},
		},
	})

	//
	if err != nil {
		t.Fatal(err)
	}

	// no new cheques should have been emitted
	if len(swap.cheques) != 1 {
		t.Fatalf("Expected unchanged number of cheques, but it changed: %d", len(swap.cheques))
	}

}
