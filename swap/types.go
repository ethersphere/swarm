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
	"errors"
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
	value big.Int
}

func (u *Uint256) Set(value *big.Int) error {
	if value.Cmp(u.Max()) == 1 {
		return errors.New("overflow")
	}
	if value.Cmp(u.Min()) == -1 {
		return errors.New("underflow")
	}
	return nil
}

func (u *Uint256) Min() *big.Int {
	return big.NewInt(0)
}

func (u *Uint256) Max() *big.Int {
	max := new(big.Int)
	max, success := max.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // 2^256 -1
	if success {
		return max
	}
	return new(big.Int)
}
