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

// PriceOracle is the interface through which Oracles will deliver prices
type PriceOracle interface {
	GetPrice(honey uint64) (uint64, error)
}

// NewPriceOracle returns the actual oracle to be used for discovering the price
// For now it will return a default one
func NewPriceOracle() PriceOracle {
	return &FixedPriceOracle{
		honeyPrice: defaultHoneyPrice,
	}
}

// FixedPriceOracle is a price oracle which which returns a fixed price.
// It is the default price oracle used as a placeholder for this iteration of the implementation.
// In production this should probably be some on-chain oracle called remotely
type FixedPriceOracle struct {
	honeyPrice uint64
}

// GetPrice returns the actual price for honey
func (cpo *FixedPriceOracle) GetPrice(honey uint64) (uint64, error) {
	// otherwise don't refresh the rate and return the latest price
	return honey * cpo.honeyPrice, nil
}
