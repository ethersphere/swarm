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

/*
This module contains the pricing for message types as constants.

Pricing in Swarm is defined as an internal unit (called `honey`).
Honey acts as a unit of relative message pricing; that is,
it allows setting prices relative to message types, irrespective of a reference currency.

The expectation is then that an external, probably on-chain, **oracle**
would be queried with the total amount of honey for a message,
for which the oracle would return the price in a given currency.

Currently the expected currency from the oracle would be wei,
but it could potentially be any currency the oracle and Swarm support,
allowing for a multi-currency design.
*/

//TODO: this calculations make little sense now, after update to ERC20-enabled chequebook
// Placeholder prices
// Based on a very crude calculation: average monthly cost for bandwidth in the US / average monthly usage of bandwidth in the US
// $67 / 190GB = $0.35 / GB
// 0.35 / (1.073.741.824) = $3.259629e^-10 / byte
// 3.259629e^-10/ (166 * 10^18) = 19636319 Wei / byte, where 166 is the current Ether price in Dollar
// per byte of data transferred, we account for 1 chunkDelivery price (accounted per byte), and 1/4096 retrieveRequest (accounted per message)
// RetrieveRequestPrice = 0.1 * 19636319 * 4096 = 8043036262, where 0.1 is a bogus factor
// ChunkDeliveryPrice = 0.9 * 19636319 = 17672687, where 0.9 is a bogus factor
const (
	RetrieveRequestPrice = uint64(8043036262)
	ChunkDeliveryPrice   = uint64(17672687)
	// default conversion of honey into output currency - currently ETH in Wei
	defaultHoneyPrice = uint64(1)
)
