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
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
)

func TestRequestCheque(t *testing.T) {
	// temp datadir
	datadir, err := ioutil.TempDir("", "swap-test")
	if err != nil {
		t.Fatal(err)
	}
	removeDataDir := func() {
		os.RemoveAll(datadir)
	}
	prvkey, err := crypto.GenerateKey()
	if err != nil {
		removeDataDir()
		t.Fatal(err)
	}

	swap, dir := createTestSwap(t)
	defer os.RemoveAll(dir)
	ss := &SwapService{swap: swap}

	protocolTester := p2ptest.NewProtocolTester(prvkey, 2, ss.run)

	creditor := protocolTester.Nodes[0]

	swap.balances[creditor.ID()] = -42

	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Label: "TestRequestCheque",
		Triggers: []p2ptest.Trigger{
			{
				Code: 0,
				Msg: &ChequeRequestMsg{
					Peer:       creditor.ID(),
					PubKey:     crypto.FromECDSAPub(creditor.Pubkey()),
					LastCheque: &Cheque{},
				},
				Peer: creditor.ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 1,
				Msg:  &EmitChequeMsg{},
				Peer: creditor.ID(),
			},
		},
	})
}
