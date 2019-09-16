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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	contract "github.com/ethersphere/swarm/contracts/swap"
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
	err = testDeploy(ctx, swap)
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
					ContractAddress: swap.GetParams().ContractAddress,
				},
				Peer: creditor.ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 0,
				Msg: &HandshakeMsg{
					ContractAddress: swap.GetParams().ContractAddress,
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
	err := testDeploy(ctx, creditorSwap)
	if err != nil {
		t.Fatal(err)
	}
	err = testDeploy(ctx, debitorSwap)
	if err != nil {
		t.Fatal(err)
	}

	log.Debug("create peer instances")

	// create the debitor peer
	dPtpPeer := p2p.NewPeer(enode.ID{}, "debitor", []p2p.Cap{})
	dProtoPeer := protocols.NewPeer(dPtpPeer, nil, Spec)
	debitor, err := creditorSwap.addPeer(dProtoPeer, debitorSwap.owner.address, debitorSwap.GetParams().ContractAddress)
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
			Contract:         debitorSwap.GetParams().ContractAddress,
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

	ctx := context.Background()
	err := testDeploy(ctx, debitorSwap)
	if err != nil {
		t.Fatal(err)
	}
	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	// create a dummy pper
	cPeer := newDummyPeer()
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
	cPeer := newDummyPeer()
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

// TestSwapRPC tests some basic things over RPC
// We want this so that we can check the API works
func TestSwapRPC(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip()
	}

	var (
		ipcPath = ".swap.ipc"
		err     error
	)

	swap, clean := newTestSwap(t, ownerKey)
	defer clean()

	// need to have a dummy contract or the call will fail at `GetParams` due to `NewAPI`
	swap.contract, err = contract.InstanceAt(common.Address{}, swap.backend)
	if err != nil {
		t.Fatal(err)
	}

	// start a service stack
	stack := createAndStartSvcNode(swap, ipcPath, t)
	defer func() {
		go stack.Stop()
	}()

	// use unique IPC path on windows
	ipcPath = filepath.Join(stack.DataDir(), ipcPath)

	// connect to the servicenode RPCs
	rpcclient, err := rpc.Dial(ipcPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(stack.DataDir())

	// create dummy peers so that we can artificially set balances and query
	dummyPeer1 := newDummyPeer()
	dummyPeer2 := newDummyPeer()
	id1 := dummyPeer1.ID()
	id2 := dummyPeer2.ID()

	// set some fake balances
	fakeBalance1 := int64(234)
	fakeBalance2 := int64(-100)

	// query a first time, should give error
	var balance int64
	err = rpcclient.Call(&balance, "swap_balance", id1)
	// at this point no balance should be there:  no peer registered with Swap
	if err == nil {
		t.Fatal("Expected error but no error received")
	}
	log.Debug("servicenode balance", "balance", balance)

	// ...thus balance should be zero
	if balance != 0 {
		t.Fatalf("Expected balance to be 0 but it is %d", balance)
	}

	peer1, err := swap.addPeer(dummyPeer1.Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}

	if err := peer1.setBalance(fakeBalance1); err != nil {
		t.Fatal(err)
	}

	peer2, err := swap.addPeer(dummyPeer2.Peer, common.Address{}, common.Address{})
	if err != nil {
		t.Fatal(err)
	}

	if err := peer2.setBalance(fakeBalance2); err != nil {
		t.Fatal(err)
	}

	// query them, values should coincide
	err = rpcclient.Call(&balance, "swap_balance", id1)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("balance1", "balance1", balance)
	if balance != fakeBalance1 {
		t.Fatalf("Expected balance %d to be equal to fake balance %d, but it is not", balance, fakeBalance1)
	}

	err = rpcclient.Call(&balance, "swap_balance", id2)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("balance2", "balance2", balance)
	if balance != fakeBalance2 {
		t.Fatalf("Expected balance %d to be equal to fake balance %d, but it is not", balance, fakeBalance2)
	}

	// now call all balances
	allBalances := make(map[enode.ID]int64)
	err = rpcclient.Call(&allBalances, "swap_balances")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("received balances", "allBalances", allBalances)

	var sum int64
	for _, v := range allBalances {
		sum += v
	}

	fakeSum := fakeBalance1 + fakeBalance2
	if sum != fakeSum {
		t.Fatalf("Expected total balance to be %d, but it %d", fakeSum, sum)
	}

	balances, err := swap.Balances()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(allBalances, balances) {
		t.Fatal("Balances are not deep equal")
	}
}

// createAndStartSvcNode setup a p2p service and start it
func createAndStartSvcNode(swap *Swap, ipcPath string, t *testing.T) *node.Node {
	stack, err := newServiceNode(ipcPath, 0, 0)
	if err != nil {
		t.Fatal("Create servicenode #1 fail", "err", err)
	}

	swapsvc := func(ctx *node.ServiceContext) (node.Service, error) {
		return swap, nil
	}

	err = stack.Register(swapsvc)
	if err != nil {
		t.Fatal("Register service in servicenode #1 fail", "err", err)
	}

	// start the nodes
	err = stack.Start()
	if err != nil {
		t.Fatal("servicenode #1 start failed", "err", err)
	}

	return stack
}

// newServiceNode creates a p2p.Service node stub
func newServiceNode(ipcPath string, httpport int, wsport int, modules ...string) (*node.Node, error) {
	var err error
	cfg := &node.DefaultConfig
	cfg.P2P.EnableMsgEvents = true
	cfg.P2P.NoDiscovery = true
	cfg.IPCPath = ipcPath
	cfg.DataDir, err = ioutil.TempDir("", "test-Service-node")
	if err != nil {
		return nil, err
	}
	if httpport > 0 {
		cfg.HTTPHost = node.DefaultHTTPHost
		cfg.HTTPPort = httpport
	}
	if wsport > 0 {
		cfg.WSHost = node.DefaultWSHost
		cfg.WSPort = wsport
		cfg.WSOrigins = []string{"*"}
		for i := 0; i < len(modules); i++ {
			cfg.WSModules = append(cfg.WSModules, modules[i])
		}
	}
	stack, err := node.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("ServiceNode create fail: %v", err)
	}
	return stack, nil
}
