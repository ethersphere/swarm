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

	"github.com/ethereum/go-ethereum/common"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
)

/*
TestEmitCheque tests an end-to-end exchange between a debitor peer
and its creditor. The debitor issues the cheque, the creditor receives it
and responds with a confirmation
*/
func TestEmitCheque(t *testing.T) {
	var err error

	// setup test swap object
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

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
	cheque.Sig, err = swap.signContent(cheque)
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
					ContractAddress: common.Address{},
				},
				Peer: creditor.ID(),
			},
			{
				Code: 0,
				Msg: &HandshakeMsg{
					ContractAddress: common.Address{},
				},
				Peer: debitor.ID(),
			},
		},
	})

	// there should be no error at this point
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Test further exchanges
}
