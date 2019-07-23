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
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	p2ptest "github.com/ethereum/go-ethereum/p2p/testing"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/p2p/protocols"
)

/*
TestHandshake creates two mock nodes and initiates an exchange;
it expects a handshake to take place between the two nodes
(the handshake would fail because we don't actually use real nodes here)
*/
func TestHandshake(t *testing.T) {
	var err error

	// setup test swap object
	swap, dir := newTestSwap(t)
	defer os.RemoveAll(dir)

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
// and handles the cheque. We finally also send back a `ConfirmMsg` to the debitor
// TODO: The ConfirmMsg part is not definitevely specificied.
func TestEmitCheque(t *testing.T) {
	log.Debug("set up test swaps")
	creditorSwap, testDir1 := newTestSwap(t)
	debitorSwap, testDir2 := newTestSwap(t)
	defer os.RemoveAll(testDir1)
	defer os.RemoveAll(testDir2)

	ctx := context.Background()

	log.Debug("deploy to simulated backend")
	testDeploy(ctx, creditorSwap.backend, creditorSwap)
	testDeploy(ctx, debitorSwap.backend, debitorSwap)
	creditorSwap.backend.(*backends.SimulatedBackend).Commit()
	debitorSwap.backend.(*backends.SimulatedBackend).Commit()

	log.Debug("create peer instances")
	// create Peer instances
	// NOTE: remember that these are peer instances representing each **a model of the remote peer** for every local node
	// so creditor is the model of the remote mode for the debitor! (and vice versa)

	// in order to be able to model as realistically as possible sending and receiving, let's use a MsgPipe
	// a MsgPipe is a duplex read-write object, write to one end and read from the other

	// create the message pipe
	crw, drw := p2p.MsgPipe()
	// create the creditor peer
	cPtpPeer := p2p.NewPeer(enode.ID{}, "creditor", []p2p.Cap{})
	cProtoPeer := protocols.NewPeer(cPtpPeer, crw, Spec)
	// create the debitor peer
	dPtpPeer := p2p.NewPeer(enode.ID{}, "dreditor", []p2p.Cap{})
	dProtoPeer := protocols.NewPeer(dPtpPeer, drw, Spec)
	// create the Swap protocol peers
	creditor := NewPeer(cProtoPeer, debitorSwap, debitorSwap.backend, creditorSwap.owner.address, debitorSwap.owner.Contract)
	debitor := NewPeer(dProtoPeer, creditorSwap, creditorSwap.backend, debitorSwap.owner.address, debitorSwap.owner.Contract)

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
			Timeout:     0,
		},
	}
	cheque.Sig, err = debitorSwap.signContent(cheque)
	if err != nil {
		t.Fatal(err)
	}

	emitMsg := &EmitChequeMsg{
		Cheque: cheque,
	}
	log.Debug("send the message with the cheque to the beneficiary")
	go creditor.Send(ctx, emitMsg)

	log.Debug("read the message on the beneficiary through the pipe")
	msg, err := drw.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("convert the message to our message type (simulated p2p.protocols)")
	var wmsg protocols.WrappedMsg
	err = msg.Decode(&wmsg)
	if err != nil {
		t.Fatal(err)
	}
	msg.Discard()

	val, ok := Spec.NewMsg(msg.Code)
	if !ok {
		t.Fatalf("invalid message code: %v", msg.Code)
	}
	if err := rlp.DecodeBytes(wmsg.Payload, val); err != nil {
		t.Fatalf("decode error <= %v: %v", msg, err)
	}

	log.Debug("trigger reading the message on the beneficiary")
	// handleMsg is blocking as it sends a synchronous confirmation message back.
	// Therefore, we need a go-routine in order to check for the test,
	// and we need to synchronize the go-routines

	errc := make(chan error)
	// start the go-routine for handling the message
	go func(t *testing.T) {
		// this blocks
		err = debitor.handleMsg(ctx, val)
		if err != nil {
			errc <- err
		}
		/*
			TODO: When saving the cheque on creditor side is implemented,
			we should to re-enable this check
			if len(creditorSwap.cheques) != 1 {
				t.Fatalf("Expected exactly one cheque at creditor, but there are %d:", len(creditorSwap.cheques))
			}
		*/
		errc <- nil
	}(t)

	// handleMsg will block until we read a message
	log.Debug("read the message on the issuer")
	msg, err = crw.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}

	err = msg.Decode(&wmsg)
	if err != nil {
		t.Fatal(err)
	}

	val, ok = Spec.NewMsg(msg.Code)
	if !ok {
		t.Fatalf("invalid message code: %v", msg.Code)
	}
	if err := rlp.DecodeBytes(wmsg.Payload, val); err != nil {
		t.Fatalf("decode error <= %v: %v", msg, err)
	}

	// check that it is indeed a `ConfirmMsg`
	var conf *ConfirmMsg
	if conf, ok = val.(*ConfirmMsg); !ok {
		t.Fatal("Expected ConfirmMsg but it was not")
	}
	if bytes.Compare(conf.Cheque.Sig, cheque.Sig) != 0 {
		t.Fatal("Expected confirmation cheque to be the same, but it is no")
	}

	// wait until the go routine terminates
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}
