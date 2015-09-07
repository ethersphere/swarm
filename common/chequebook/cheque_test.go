package chequebook

import (
	// "crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func genAddr() common.Address {
	prvKey, _ := crypto.GenerateKey()
	return crypto.PubkeyToAddress(prvKey.PublicKey)
}

func TestChequebook(t *testing.T) {
	prvKey, _ := crypto.GenerateKey()
	sender := genAddr()
	chbook, err := NewChequebook(sender, prvKey)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	recipient := genAddr()
	amount := common.Big1
	ch, err := chbook.NewCheque(recipient, amount)
	if err == nil {
		t.Errorf("expected insufficient funds error, got none")
	}

	chbook.Deposit(common.Big1)
	if chbook.Balance().Cmp(common.Big1) != 0 {
		t.Errorf("expected: %v, got %v", "0", chbook.Balance())
	}

	ch, err = chbook.NewCheque(recipient, amount)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if chbook.Balance().Cmp(common.Big0) != 0 {
		t.Errorf("expected: %v, got %v", "0", chbook.Balance())
	}

	chbox, err := NewChequebox(sender, recipient, &prvKey.PublicKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	received, err := chbox.Receive(ch)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received.Cmp(common.Big1) != 0 {
		t.Errorf("expected: %v, got %v", "1", received)
	}

}
