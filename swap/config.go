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
	"time"
)

// These are currently arbitrary values which have not been verified nor tested
// Need experimentation to arrive to values which make sense
const (
	// Thresholds which trigger payment or disconnection. The unit is in honey (internal accounting unit)
	// DefaultPaymentThreshold is set to be equivalent to requesting and serving 10mb of data (2441 chunks (4096 bytes) = 10 mb, 10^7 bytes = 10 mb)
	DefaultPaymentThreshold    = 2441*RetrieveRequestPrice + (10^7)*ChunkDeliveryPrice // 4096 * 2441 = 10 mb,
	DefaultDisconnectThreshold = 20 * DefaultPaymentThreshold
	// DefaultDepositAmount is the default amount to send to the contract when initially deploying
	// NOTE: deliberate value for now; needs experimentation
	DefaultDepositAmount = 0
	// This is the amount of time in seconds which an issuer has to wait to decrease the harddeposit of a beneficiary.
	// The smart-contract allows for setting this variable differently per beneficiary
	defaultHarddepositTimeoutDuration = 24 * time.Hour
	// Until we deploy swap officially, it's only allowed to be enabled under a specific network ID (use the --bzznetworkid flag to set it)
	AllowedNetworkID          = 5
	DefaultTransactionTimeout = 10 * time.Minute
)
