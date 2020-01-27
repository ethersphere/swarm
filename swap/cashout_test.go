// Copyright 2020 The Swarm Authors
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

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethersphere/swarm/boundedint"
	contract "github.com/ethersphere/swarm/contracts/swap"
)

// TestContractIntegration tests a end-to-end cheque interaction.
// First a simulated backend is created, then we deploy the issuer's swap contract.
// We issue a test cheque with the beneficiary address and on the issuer's contract,
// and immediately try to cash-in the cheque
// afterwards it attempts to cash-in a bouncing cheque
func TestContractIntegration(t *testing.T) {
	backend := newTestBackend(t)
	reset := setupContractTest()
	defer reset()

	payout := boundedint.FromUint64(42)

	chequebook, err := testDeployWithPrivateKey(context.Background(), backend, ownerKey, ownerAddress, payout)
	if err != nil {
		t.Fatal(err)
	}

	cheque, err := newSignedTestCheque(chequebook.ContractParams().ContractAddress, beneficiaryAddress, payout, ownerKey)
	if err != nil {
		t.Fatal(err)
	}

	opts := bind.NewKeyedTransactor(beneficiaryKey)

	tx, err := chequebook.CashChequeBeneficiaryStart(opts, beneficiaryAddress, payout, cheque.Signature)
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := contract.WaitForTransactionByHash(context.Background(), backend, tx.Hash())
	if err != nil {
		t.Fatal(err)
	}

	cashResult := chequebook.CashChequeBeneficiaryResult(receipt)
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}
	if cashResult.Bounced {
		t.Fatal("cashing bounced")
	}

	// check state, check that cheque is indeed there
	result, err := chequebook.PaidOut(nil, beneficiaryAddress)
	if err != nil {
		t.Fatal(err)
	}
	paidOut, err := boundedint.NewUint256().Set(*result)
	if err != nil {
		t.Fatal(err)
	}

	if !cheque.CumulativePayout.Equals(paidOut) {
		t.Fatalf("Wrong cumulative payout %v", paidOut)
	}
	log.Debug("cheques result", "result", result)

	// create a cheque that will bounce
	_, err = payout.Add(payout, boundedint.FromUint64(10000*RetrieveRequestPrice))
	if err != nil {
		t.Fatal(err)
	}

	bouncingCheque, err := newSignedTestCheque(chequebook.ContractParams().ContractAddress, beneficiaryAddress, payout, ownerKey)
	if err != nil {
		t.Fatal(err)
	}

	tx, err = chequebook.CashChequeBeneficiaryStart(opts, beneficiaryAddress, bouncingCheque.CumulativePayout, bouncingCheque.Signature)
	if err != nil {
		t.Fatal(err)
	}

	receipt, err = contract.WaitForTransactionByHash(context.Background(), backend, tx.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Status != 1 {
		t.Fatalf("Bad status %d", receipt.Status)
	}

	cashResult = chequebook.CashChequeBeneficiaryResult(receipt)
	if !cashResult.Bounced {
		t.Fatal("cheque did not bounce")
	}

}

// TestCashCheque creates a valid cheque and feeds it to cashoutProcessor.cashCheque
func TestCashCheque(t *testing.T) {
	backend := newTestBackend(t)
	reset := setupContractTest()
	defer reset()

	cashoutProcessor := newCashoutProcessor(backend, ownerKey)
	payout := boundedint.FromUint64(42)

	chequebook, err := testDeployWithPrivateKey(context.Background(), backend, ownerKey, ownerAddress, payout)
	if err != nil {
		t.Fatal(err)
	}

	testCheque, err := newSignedTestCheque(chequebook.ContractParams().ContractAddress, ownerAddress, payout, ownerKey)
	if err != nil {
		t.Fatal(err)
	}

	err = cashoutProcessor.cashCheque(context.Background(), &CashoutRequest{
		Cheque:      *testCheque,
		Destination: ownerAddress,
	})
	if err != nil {
		t.Fatal(err)
	}

	paidOut, err := chequebook.PaidOut(nil, ownerAddress)
	if err != nil {
		t.Fatal(err)
	}

	cumulativePayout := testCheque.CumulativePayout.Value()
	if paidOut.Cmp(&cumulativePayout) != 0 {
		t.Fatalf("paidOut does not equal the CumulativePayout: paidOut=%v expected=%v", paidOut, testCheque.CumulativePayout)
	}
}

// TestEstimatePayout creates a valid cheque and feeds it to cashoutProcessor.estimatePayout
func TestEstimatePayout(t *testing.T) {
	backend := newTestBackend(t)
	reset := setupContractTest()
	defer reset()

	cashoutProcessor := newCashoutProcessor(backend, ownerKey)
	payout := boundedint.FromUint64(42)

	chequebook, err := testDeployWithPrivateKey(context.Background(), backend, ownerKey, ownerAddress, payout)
	if err != nil {
		t.Fatal(err)
	}

	testCheque, err := newSignedTestCheque(chequebook.ContractParams().ContractAddress, ownerAddress, payout, ownerKey)
	if err != nil {
		t.Fatal(err)
	}

	expectedPayout, transactionCost, err := cashoutProcessor.estimatePayout(context.Background(), testCheque)
	if err != nil {
		t.Fatal(err)
	}

	if !expectedPayout.Equals(payout) {
		t.Fatalf("unexpected expectedPayout: got %v, wanted: %v", expectedPayout, payout)
	}

	// the gas price in the simulated backend is 1 therefore the total transactionCost should be 50000 * 1 = 50000
	if !transactionCost.Equals(boundedint.FromUint64(CashChequeBeneficiaryTransactionCost)) {
		t.Fatalf("unexpected transaction cost: got %v, wanted: %d", transactionCost, 0)
	}
}
