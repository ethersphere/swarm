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

package protocols

import (
	"testing"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rlp"
)

//dummy Balance implementation
type dummyBalance struct {
	amount int64
	peer   *Peer
}

//a dummy message which needs size based accounting
//sender pays
type perBytesMsgSenderPays struct {
	Content string
}

//a dummy message which needs size based accounting
//receiver pays
type perBytesMsgReceiverPays struct {
	Content string
}

//a dummy message which is paid for per unit
//sender pays
type perUnitMsgSenderPays struct{}

//receiver pays
type perUnitMsgReceiverPays struct{}

//a dummy message which has zero as its price
type zeroPriceMsg struct{}

//a dummy message which has no accounting
type nilPriceMsg struct{}

func (m *perBytesMsgReceiverPays) Price() *Price {
	return &Price{
		PerByte: true,
		Value:   uint64(100),
		Payer:   Receiver,
	}
}

func (m *perBytesMsgSenderPays) Price() *Price {
	return &Price{
		PerByte: true,
		Value:   uint64(100),
		Payer:   Sender,
	}
}

func (m *perUnitMsgReceiverPays) Price() *Price {
	return &Price{
		PerByte: false,
		Value:   uint64(99),
		Payer:   Receiver,
	}
}

func (m *perUnitMsgSenderPays) Price() *Price {
	return &Price{
		PerByte: false,
		Value:   uint64(99),
		Payer:   Sender,
	}
}

func (m *zeroPriceMsg) Price() *Price {
	return &Price{
		PerByte: false,
		Value:   uint64(0),
		Payer:   Sender,
	}
}

//dummy accounting implementation, only stores values for later check
func (d *dummyBalance) Add(amount int64, peer *Peer) error {
	d.amount = amount
	d.peer = peer
	return nil
}

type testCase struct {
	msg        interface{}
	size       uint32
	sendResult int64
	recvResult int64
}

//lowest level unit test
func TestBalance(t *testing.T) {
	//create instances
	balance := &dummyBalance{}
	//create the spec
	spec := createTestSpec()
	//create the accounting hook for the spec
	acc := NewAccounting(balance)
	//create a peer
	id := adapters.RandomNodeConfig().ID
	p := p2p.NewPeer(id, "testPeer", nil)
	peer := NewPeer(p, &dummyRW{}, spec)
	//price depends on size, receiver pays
	msg := &perBytesMsgReceiverPays{Content: "testBalance"}
	size, _ := rlp.EncodeToBytes(msg)

	testCases := []testCase{
		{
			msg,
			uint32(len(size)),
			int64(len(size) * 100),
			int64(len(size) * -100),
		},
		{
			&perBytesMsgSenderPays{Content: "testBalance"},
			uint32(len(size)),
			int64(len(size) * -100),
			int64(len(size) * 100),
		},
		{
			&perUnitMsgSenderPays{},
			0,
			int64(-99),
			int64(99),
		},
		{
			&perUnitMsgReceiverPays{},
			0,
			int64(99),
			int64(-99),
		},
		{
			&zeroPriceMsg{},
			0,
			int64(0),
			int64(0),
		},
		{
			&nilPriceMsg{},
			0,
			int64(0),
			int64(0),
		},
	}
	checkAccountingTestCases(t, testCases, acc, peer, balance, true)
	checkAccountingTestCases(t, testCases, acc, peer, balance, false)
}

func checkAccountingTestCases(t *testing.T, cases []testCase, acc *Accounting, peer *Peer, balance *dummyBalance, send bool) {
	for _, c := range cases {
		var err error
		var expectedResult int64
		//reset balance before every check
		balance.amount = 0
		if send {
			err = acc.Send(peer, c.size, c.msg)
			expectedResult = c.sendResult
		} else {
			err = acc.Receive(peer, c.size, c.msg)
			expectedResult = c.recvResult
		}

		checkResults(t, err, balance, peer, expectedResult)
	}
}

func checkResults(t *testing.T, err error, balance *dummyBalance, peer *Peer, result int64) {
	if err != nil {
		t.Fatal(err)
	}
	if balance.peer != peer {
		t.Fatalf("expected Add to be called with peer %v, got %v", peer, balance.peer)
	}
	if balance.amount != result {
		t.Fatalf("Expected balance to be %d but is %d", result, balance.amount)
	}
}

//create a test spec
func createTestSpec() *Spec {
	spec := &Spec{
		Name:       "test",
		Version:    42,
		MaxMsgSize: 10 * 1024,
		Messages: []interface{}{
			&perBytesMsgReceiverPays{},
			&perBytesMsgSenderPays{},
			&perUnitMsgReceiverPays{},
			&perUnitMsgSenderPays{},
			&zeroPriceMsg{},
			&nilPriceMsg{},
		},
	}
	return spec
}
