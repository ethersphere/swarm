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
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	contract "github.com/ethersphere/swarm/contracts/swap"
	p2ptest "github.com/ethersphere/swarm/p2p/testing"
	"github.com/ethersphere/swarm/uint256"
	colorable "github.com/mattn/go-colorable"
)

func init() {
	log.PrintOrigins(true)
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(*loglevel), log.StreamHandler(colorable.NewColorableStderr(), log.TerminalFormat(true))))
}

// protocol tester based on a swap instance
type swapTester struct {
	*p2ptest.ProtocolTester
	swap *Swap
}

// creates a new protocol tester for swap with a deployed chequebook
func newSwapTester(t *testing.T, backend *swapTestBackend, depositAmount *uint256.Uint256) (*swapTester, func(), error) {
	swap, clean := newTestSwap(t, ownerKey, backend)

	err := testDeploy(context.Background(), swap, depositAmount)
	if err != nil {
		return nil, nil, err
	}

	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester := p2ptest.NewProtocolTester(swap.owner.privateKey, 1, swap.run)
	return &swapTester{
		ProtocolTester: protocolTester,
		swap:           swap,
	}, clean, nil
}

// creates a test exchange for the handshakes
func HandshakeMsgExchange(lhs, rhs *HandshakeMsg, id enode.ID) []p2ptest.Exchange {
	return []p2ptest.Exchange{
		{
			Expects: []p2ptest.Expect{
				{
					Code: 0,
					Msg:  lhs,
					Peer: id,
				},
			},
		},
		{
			Triggers: []p2ptest.Trigger{
				{
					Code: 0,
					Msg:  rhs,
					Peer: id,
				},
			},
		},
	}
}

// helper function for testing the handshake
// lhs is the HandshakeMsg we expect to be sent, rhs the one we receive
// disconnects is a list of disconnect events to be expected
func (s *swapTester) testHandshake(lhs, rhs *HandshakeMsg, disconnects ...*p2ptest.Disconnect) error {
	if err := s.TestExchanges(HandshakeMsgExchange(lhs, rhs, s.Nodes[0].ID())...); err != nil {
		return err
	}

	if len(disconnects) > 0 {
		return s.TestDisconnected(disconnects...)
	}

	// If we don't expect disconnect, ensure peers remain connected
	err := s.TestDisconnected(&p2ptest.Disconnect{
		Peer:  s.Nodes[0].ID(),
		Error: nil,
	})

	if err == nil {
		return fmt.Errorf("Unexpected peer disconnect")
	}

	if err.Error() != "timed out waiting for peers to disconnect" {
		return err
	}

	return nil
}

// creates a new HandshakeMsg
func newSwapHandshakeMsg(contractAddress common.Address, chainID uint64) *HandshakeMsg {
	return &HandshakeMsg{
		ContractAddress: contractAddress,
		ChainID:         chainID,
	}
}

// creates the correct HandshakeMsg based on Swap instance
func correctSwapHandshakeMsg(swap *Swap) *HandshakeMsg {
	return newSwapHandshakeMsg(swap.GetParams().ContractAddress, swap.chainID)
}

// TestHandshake tests the correct handshake scenario
func TestHandshake(t *testing.T) {
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester, clean, err := newSwapTester(t, nil, uint256.FromUint64(0))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}

	err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(protocolTester.swap),
		correctSwapHandshakeMsg(protocolTester.swap),
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestHandshakeInvalidChainID tests that a handshake with the wrong chain id is rejected
func TestHandshakeInvalidChainID(t *testing.T) {
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester, clean, err := newSwapTester(t, nil, uint256.FromUint64(0))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}

	err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(protocolTester.swap),
		newSwapHandshakeMsg(protocolTester.swap.GetParams().ContractAddress, 1234),
		&p2ptest.Disconnect{
			Peer:  protocolTester.Nodes[0].ID(),
			Error: fmt.Errorf("message handler: (msg code 0): %v", ErrDifferentChainID),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestHandshakeEmptyContract tests that a handshake with an empty contract address is rejected
func TestHandshakeEmptyContract(t *testing.T) {
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester, clean, err := newSwapTester(t, nil, uint256.FromUint64(0))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}

	err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(protocolTester.swap),
		newSwapHandshakeMsg(common.Address{}, 1234),
		&p2ptest.Disconnect{
			Peer:  protocolTester.Nodes[0].ID(),
			Error: fmt.Errorf("message handler: (msg code 0): %v", ErrEmptyAddressInSignature),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestHandshakeInvalidContract tests that a handshake with an address that's not a valid chequebook
func TestHandshakeInvalidContract(t *testing.T) {
	// setup the protocolTester, which will allow protocol testing by sending messages
	protocolTester, clean, err := newSwapTester(t, nil, uint256.FromUint64(0))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}

	err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(protocolTester.swap),
		newSwapHandshakeMsg(ownerAddress, protocolTester.swap.chainID),
		&p2ptest.Disconnect{
			Peer:  protocolTester.Nodes[0].ID(),
			Error: fmt.Errorf("message handler: (msg code 0): %v", contract.ErrNotDeployedByFactory),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestEmitCheque tests the correct processing of EmitChequeMsg messages
// One protocol tester is created which will receive the EmitChequeMsg
// A second swap instance is created for easy creation of a chequebook contract which is deployed to the simulated backend
// We send a EmitChequeMsg to the creditor which handles the cheque and sends a ConfirmChequeMsg
func TestEmitCheque(t *testing.T) {
	testBackend := newTestBackend(t)

	protocolTester, clean, err := newSwapTester(t, testBackend, uint256.FromUint64(0))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}
	creditorSwap := protocolTester.swap

	debitorSwap, cleanDebitorSwap := newTestSwap(t, beneficiaryKey, testBackend)
	defer cleanDebitorSwap()

	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	log.Debug("deploy to simulated backend")

	// cashCheque cashes a cheque when the reward of doing so is twice the transaction costs.
	// gasPrice on testBackend == 1
	// estimated gas costs == 50000
	// cheque should be sent if the accumulated amount of uncashed cheques is worth more than 100000
	balance := uint256.FromUint64(100001)
	balanceValue := balance.Value()

	if err := testDeploy(context.Background(), debitorSwap, balance); err != nil {
		t.Fatal(err)
	}

	if err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(creditorSwap),
		correctSwapHandshakeMsg(debitorSwap),
	); err != nil {
		t.Fatal(err)
	}

	debitor := creditorSwap.getPeer(protocolTester.Nodes[0].ID())
	// set balance artificially
	if err = debitor.setBalance(balanceValue.Int64()); err != nil {
		t.Fatal(err)
	}

	// a safe check: at this point no cheques should be in the swap
	if debitor.getLastReceivedCheque() != nil {
		t.Fatalf("Expected no cheques at creditor, but there is %v:", debitor.getLastReceivedCheque())
	}

	cheque := &Cheque{
		ChequeParams: ChequeParams{
			Contract:         debitorSwap.GetParams().ContractAddress,
			Beneficiary:      creditorSwap.owner.address,
			CumulativePayout: balance,
		},
		Honey: balanceValue.Uint64(),
	}
	cheque.Signature, err = cheque.Sign(debitorSwap.owner.privateKey)
	if err != nil {
		t.Fatal(err)
	}

	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Triggers: []p2ptest.Trigger{
			{
				Code: 1,
				Msg: &EmitChequeMsg{
					Cheque: cheque,
				},
				Peer: protocolTester.Nodes[0].ID(),
			},
		},
		Expects: []p2ptest.Expect{
			{
				Code: 2,
				Msg: &ConfirmChequeMsg{
					Cheque: cheque,
				},
				Peer: protocolTester.Nodes[0].ID(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// check that the balance has been reset
	if debitor.getBalance() != 0 {
		t.Fatalf("Expected debitor balance to have been reset to %d, but it is %d", 0, debitor.getBalance())
	}
	recvCheque := debitor.getLastReceivedCheque()
	log.Debug("expected cheque", "cheque", recvCheque)
	if !recvCheque.Equal(cheque) {
		t.Fatalf("Expected cheque %v at creditor, but it was %v:", cheque, recvCheque)
	}

	// we wait until the cashCheque is actually terminated (ensures proper nonce count)
	select {
	case <-testBackend.cashDone:
		log.Debug("cash transaction completed and committed")
	case <-time.After(4 * time.Second):
		t.Fatalf("Timeout waiting for cash transaction to complete")
	}
}

// TestTriggerPaymentThreshold is to test that the whole cheque protocol is triggered
// when we reach the payment threshold
// One protocol tester is created and then Add with a value above the payment threshold is called for another node
// we expect a EmitChequeMsg to be sent, then we send a ConfirmChequeMsg to the swap instance
func TestTriggerPaymentThreshold(t *testing.T) {
	testBackend := newTestBackend(t)
	log.Debug("create test swap")
	protocolTester, clean, err := newSwapTester(t, testBackend, uint256.FromUint64(DefaultPaymentThreshold*2))
	defer clean()
	if err != nil {
		t.Fatal(err)
	}
	debitorSwap := protocolTester.swap

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// setup the wait for mined transaction function for testing
	cleanup := setupContractTest()
	defer cleanup()

	if err = protocolTester.testHandshake(
		correctSwapHandshakeMsg(debitorSwap),
		correctSwapHandshakeMsg(debitorSwap),
	); err != nil {
		t.Fatal(err)
	}

	creditor := debitorSwap.getPeer(protocolTester.Nodes[0].ID())

	// set the balance to manually be at PaymentThreshold
	overDraft := 42
	expectedAmount := uint64(overDraft) + DefaultPaymentThreshold
	if err = creditor.setBalance(-int64(DefaultPaymentThreshold)); err != nil {
		t.Fatal(err)
	}

	// we expect a cheque at the end of the test, but not yet
	if creditor.getLastSentCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v:", creditor.getLastSentCheque())
	}
	// do some accounting, no error expected, just a WARN
	err = debitorSwap.Add(int64(-overDraft), creditor.Peer)
	if err != nil {
		t.Fatal(err)
	}

	// balance should be reset now
	if creditor.getBalance() != 0 {
		t.Fatalf("Expected debitorSwap balance to be 0, but is %d", creditor.getBalance())
	}

	// pending cheque should now be set
	pending := creditor.getPendingCheque()
	if pending == nil {
		t.Fatal("Expected pending cheque")
	}

	if !pending.CumulativePayout.Equals(uint256.FromUint64(expectedAmount)) {
		t.Fatalf("Expected cheque cumulative payout to be %d, but is %v", expectedAmount, pending.CumulativePayout)
	}

	if pending.Honey != expectedAmount {
		t.Fatalf("Expected cheque honey to be %d, but is %d", expectedAmount, pending.Honey)
	}

	if pending.Beneficiary != creditor.beneficiary {
		t.Fatalf("Expected cheque beneficiary to be %x, but is %x", creditor.beneficiary, pending.Beneficiary)
	}

	if pending.Contract != debitorSwap.contract.ContractParams().ContractAddress {
		t.Fatalf("Expected cheque contract to be %x, but is %x", debitorSwap.contract.ContractParams().ContractAddress, pending.Contract)
	}

	// we expect a EmitChequeMsg to be sent, then we trigger a ConfirmChequeMsg for the same cheque
	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Expects: []p2ptest.Expect{
			{
				Code: 1,
				Msg: &EmitChequeMsg{
					Cheque: creditor.getPendingCheque(),
				},
				Peer: protocolTester.Nodes[0].ID(),
			},
		},
	}, p2ptest.Exchange{
		Triggers: []p2ptest.Trigger{
			{
				Code: 2,
				Msg: &ConfirmChequeMsg{
					Cheque: creditor.getPendingCheque(),
				},
				Peer: protocolTester.Nodes[0].ID(),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// we wait until the confirm message has been processed
loop:
	for {
		select {
		case <-ctx.Done():
			t.Fatal("Expected one cheque, but there is none")
		default:
			creditor.lock.Lock()
			lastSentCheque := creditor.getLastSentCheque()
			creditor.lock.Unlock()
			if lastSentCheque != nil {
				break loop
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	cheque := creditor.getLastSentCheque()

	if !cheque.Equal(pending) {
		t.Fatalf("Expected sent cheque to be the last pending one. expected: %v, but is %v", pending, cheque)
	}

	// because no other accounting took place in the meantime the balance should be exactly 0
	if creditor.getBalance() != 0 {
		t.Fatalf("Expected debitorSwap balance to be 0, but is %d", creditor.getBalance())
	}

	// do some accounting again to trigger a second cheque
	if err = debitorSwap.Add(-int64(DefaultPaymentThreshold), creditor.Peer); err != nil {
		t.Fatal(err)
	}

	// we expect a cheque to be sent
	err = protocolTester.TestExchanges(p2ptest.Exchange{
		Expects: []p2ptest.Expect{
			{
				Code: 1,
				Msg: &EmitChequeMsg{
					Cheque: creditor.getPendingCheque(),
				},
				Peer: protocolTester.Nodes[0].ID(),
			},
		},
	})
	if err != nil {
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
	creditorSwap, clean := newTestSwap(t, beneficiaryKey, nil)
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
	if err = debitor.setBalance(expectedBalance); err != nil {
		t.Fatal(err)
	}
	// we also don't expect any cheques yet
	if debitor.getPendingCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v", debitor.getPendingCheque())
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
	if debitor.getPendingCheque() != nil {
		t.Fatalf("Expected still no cheques yet, but there is %v", debitor.getPendingCheque())
	}

	// let's do the whole thing again (actually a bit silly, it's somehow simulating the peer would have been dropped)
	err = creditorSwap.Add(int64(overDraft), debitor.Peer)
	if err == nil {
		t.Fatal("Expected an error due to overdraft, but did not get any")
	}

	if debitor.getBalance() != expectedBalance {
		t.Fatalf("Expected balance to be %d, but is %d", expectedBalance, debitor.getBalance())
	}

	if debitor.getPendingCheque() != nil {
		t.Fatalf("Expected no cheques yet, but there is %v", debitor.getPendingCheque())
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

	swap, clean := newTestSwap(t, ownerKey, nil)
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
	err = rpcclient.Call(&balance, "swap_peerBalance", id1)
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
	err = rpcclient.Call(&balance, "swap_peerBalance", id1)
	if err != nil {
		t.Fatal(err)
	}
	log.Debug("balance1", "balance1", balance)
	if balance != fakeBalance1 {
		t.Fatalf("Expected balance %d to be equal to fake balance %d, but it is not", balance, fakeBalance1)
	}

	err = rpcclient.Call(&balance, "swap_peerBalance", id2)
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
