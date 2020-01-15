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
	"bytes"
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/uint256"
)

// encodeForSignature encodes the cheque params in the format used in the signing procedure
func (cheque *ChequeParams) encodeForSignature() []byte {
	cumulativePayoutBytes := make([]byte, 32)
	// we need to write the last 8 bytes as we write a uint64 into a 32-byte array
	// encoded in BigEndian because EVM uses BigEndian encoding
	cumulativePayout := cheque.CumulativePayout.Value()
	chequePayoutBytes := (&cumulativePayout).Bytes()
	copy(cumulativePayoutBytes[32-len(chequePayoutBytes):], chequePayoutBytes)

	// construct the actual cheque
	input := cheque.Contract.Bytes()
	input = append(input, cheque.Beneficiary.Bytes()...)
	input = append(input, cumulativePayoutBytes[:]...)
	return input
}

// sigHash hashes the cheque params using the prefix that would be added by eth_Sign
func (cheque *ChequeParams) sigHash() []byte {
	// we can ignore the error because it is always nil
	encoded := cheque.encodeForSignature()
	input := crypto.Keccak256(encoded)
	withPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(input), input)
	return crypto.Keccak256([]byte(withPrefix))
}

// VerifySig verifies the signature on the cheque
func (cheque *Cheque) VerifySig(expectedSigner common.Address) error {
	sigHash := cheque.sigHash()

	if cheque.Signature == nil {
		return fmt.Errorf("tried to verify signature on cheque with sig nil")
	}

	if len(cheque.Signature) != 65 {
		return fmt.Errorf("signature has invalid length: %d", len(cheque.Signature))
	}
	// copy signature to avoid modifying the original
	sig := make([]byte, len(cheque.Signature))
	copy(sig, cheque.Signature)
	// reduce the v value of the signature by 27 (see Sign)
	sig[len(sig)-1] -= 27
	pubKey, err := crypto.SigToPub(sigHash, sig)
	if err != nil {
		return err
	}

	if crypto.PubkeyToAddress(*pubKey) != expectedSigner {
		return ErrInvalidChequeSignature
	}

	return nil
}

// Sign returns the cheque's signature with supplied private key
func (cheque *ChequeParams) Sign(prv *ecdsa.PrivateKey) ([]byte, error) {
	sig, err := crypto.Sign(cheque.sigHash(), prv)
	if err != nil {
		return nil, err
	}
	// increase the v value by 27 as crypto.Sign produces 0 or 1 but the contract only accepts 27 or 28
	// this is to prevent malleable signatures. while not strictly necessary in this case the ECDSA implementation from Openzeppelin expects it.
	sig[len(sig)-1] += 27
	return sig, nil
}

// Equal checks if other has the same fields
func (cheque *Cheque) Equal(other *Cheque) bool {
	if cheque.Beneficiary != other.Beneficiary {
		return false
	}

	if !cheque.CumulativePayout.Equals(other.CumulativePayout) {
		return false
	}

	if cheque.Honey != other.Honey {
		return false
	}

	if !bytes.Equal(cheque.Signature, other.Signature) {
		return false
	}

	return true
}

// verifyChequeProperties verifies the signature and if the cheque fields are appropriate for this peer
// it does not verify anything that requires knowing the previous cheque
func (cheque *Cheque) verifyChequeProperties(p *Peer, expectedBeneficiary common.Address) error {
	if cheque.Contract != p.contractAddress {
		return fmt.Errorf("wrong cheque parameters: expected contract: %x, was: %x", p.contractAddress, cheque.Contract)
	}

	// the beneficiary is the owner of the counterparty swap contract
	if err := cheque.VerifySig(p.beneficiary); err != nil {
		return err
	}

	if cheque.Beneficiary != expectedBeneficiary {
		return fmt.Errorf("wrong cheque parameters: expected beneficiary: %x, was: %x", expectedBeneficiary, cheque.Beneficiary)
	}

	return nil
}

// verifyChequeAgainstLast verifies that the amount is higher than in the previous cheque and the increase is as expected
// returns the actual amount received in this cheque
func (cheque *Cheque) verifyChequeAgainstLast(lastCheque *Cheque, expectedAmount *uint256.Uint256) (*uint256.Uint256, error) {
	actualAmount := uint256.New().Copy(cheque.CumulativePayout)

	if lastCheque != nil {
		if cheque.CumulativePayout.Cmp(lastCheque.CumulativePayout) < 1 {
			return nil, fmt.Errorf("wrong cheque parameters: expected cumulative payout larger than %v, was: %v", lastCheque.CumulativePayout, cheque.CumulativePayout)
		}

		actualAmount.Sub(actualAmount, lastCheque.CumulativePayout)
	}

	if !expectedAmount.Equals(actualAmount) {
		return nil, fmt.Errorf("unexpected amount for honey, expected %v was %v", expectedAmount, actualAmount)
	}

	return actualAmount, nil
}

func (cheque *Cheque) String() string {
	return fmt.Sprintf("Contract: %x Beneficiary: %x CumulativePayout: %v Honey: %d", cheque.Contract, cheque.Beneficiary, cheque.CumulativePayout, cheque.Honey)
}
