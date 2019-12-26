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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ChequeParams encapsulate all cheque parameters
type ChequeParams struct {
	Contract         common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary      common.Address // address of the beneficiary, the contract which will redeem the cheque
	CumulativePayout uint64         // cumulative amount of the cheque in currency
}

// Cheque encapsulates the parameters and the signature
type Cheque struct {
	ChequeParams
	Honey     uint64 // amount of honey which resulted in the cumulative currency difference
	Signature []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

// HandshakeMsg is exchanged on peer handshake
type HandshakeMsg struct {
	ChainID         uint64         // chain id of the blockchain the peer is connected to
	ContractAddress common.Address // chequebook contract address of the peer
}

// EmitChequeMsg is sent from the debitor to the creditor with the actual cheque
type EmitChequeMsg struct {
	Cheque *Cheque
}

// ConfirmChequeMsg is sent from the creditor to the debitor with the cheque to confirm successful processing
type ConfirmChequeMsg struct {
	Cheque *Cheque
}

// Uint256 represents an unsigned integer of 256 bits
type Uint256 struct {
	value *big.Int
}

var minUint256 = big.NewInt(0)
var maxUint256, _ = new(big.Int).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // 2^256 - 1 (base 10)

// NewUint256 returns a new Uint256 struct with a value based on the given uint64 param
func NewUint256(i uint64) *Uint256 {
	var u *Uint256
	u.value = new(big.Int).SetUint64(i) // any uint64 is good enough for a uint256
	return u
}

// Value returns the underlying big.Int pointer for the Uint256 struct
func (u *Uint256) Value() *big.Int {
	return u.value
}

// Set assigns a new value to the underlying pointer within the unsigned 256-bit integer range
func (u *Uint256) Set(value *big.Int) error {
	if value.Cmp(minUint256) == 1 {
		return fmt.Errorf("cannot set uint256 to %v as it overflows max value of %v", value, maxUint256)
	}
	if value.Cmp(maxUint256) == -1 {
		return fmt.Errorf("cannot set uint256 to %v as it underflows min value of %v", value, minUint256)
	}
	u.value = value
	return nil
}

// Add attempts to add the given addend to an unsigned 256-bit integer
func (u *Uint256) Add(addend *big.Int) error {
	var summand *big.Int
	summand.Add(u.Value(), addend)
	return u.Set(summand)
}

// Sub attempts to subtract the given subtrahend from an unsigned 256-bit integer
func (u *Uint256) Sub(subtrahend *big.Int) error {
	var difference *big.Int
	difference.Sub(u.Value(), subtrahend)
	return u.Set(difference)
}
