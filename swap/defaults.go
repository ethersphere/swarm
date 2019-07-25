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

import "time"

// These are currently arbitrary values which have not been verified nor tested
// Need experimentation to arrive to values which make sense
const (
	// Thresholds which trigger payment or disconnection. The unit is in honey (internal accounting unit)
	DefaultPaymentThreshold    = 1000000
	DefaultDisconnectThreshold = 1500000
	// DefaultInitialDepositAmount is the default amount to send to the contract when initially deploying
	// TODO: deliberate value for now; needs experimentation
	DefaultInitialDepositAmount = 0

	deployRetries = 5
	// delay between retries
	deployDelay = 1 * time.Second
	// Default timeout until cashing in cheques is possible - TODO: deliberate value, experiment
	// Should be non-zero once we implement waivers
	defaultCashInDelay = uint64(0)
	// This is the amount of time in seconds which an issuer has to wait to decrease the harddeposit of a beneficiary.
	// The smart-contract allows for setting this variable differently per beneficiary
	defaultHarddepositTimeoutDuration = 24 * time.Hour

	// While Swap is unstable, it's only allowed to be run under a specific network ID
	AllowedNetworkID = 5
)
