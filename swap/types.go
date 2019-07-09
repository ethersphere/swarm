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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// TODO: add handshake protocol where we exchange last cheque (this is useful if one node disconnects)
// FIXME: Check the contract bytecode of the counterparty

type ChequeParams struct {
	Contract    common.Address // address of chequebook, needed to avoid cross-contract submission
	Beneficiary common.Address
	Serial      uint64 // cumulative amount of all funds sent
	Amount      uint64 // cumulative amount of all funds sent
	Timeout     uint64
}

// TODO: There should be a request cheque struct that only gives the Serial
// Cheque encapsulates the cheque information
type Cheque struct {
	ChequeParams
	Sig []byte // signature Sign(Keccak256(contract, beneficiary, amount), prvKey)
}

// ChequeRequestMsg is sent from a creditor to the debitor to solicit a cheque
type ChequeRequestMsg struct {
	Peer       enode.ID // TODO: Why is it here right now? Potentially not needed as everything goes through peer
	PubKey     []byte   // TODO: Also probably not needed
	LastCheque *Cheque  // TODO: maybe at most just a serial number rather than full cheque
}

// EmitChequeMsg is sent from the debitor to the creditor with the actual check
type EmitChequeMsg struct {
	Cheque *Cheque
}

// ErrorMsg is sent in case of an error TODO: specify error conditions and when this needs to be sent
type ErrorMsg struct{}

// ConfirmMsg is sent from the creditor to the debitor to confirm cheque reception
type ConfirmMsg struct {
	Cheque Cheque // TODO: probably not needed and if so likely should not include the full cheque
}
