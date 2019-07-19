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

const (
	// maximum validity of the honey price
	maxPriceAge       = 1 * time.Hour // TODO: this is currently an arbitrary value - currently irrelevant as prices are fixed
	defaultHoneyPrice = uint64(1)     // default convertion of honey into output currency - currently ETH
)

// Currency represents a string identifier for currencies
type Currency string

// PriceOracle is the interface through which Oracles will deliver prices
type PriceOracle interface {
	GetPrice(honey int64) (int64, error)
}

// NewPriceOracle returns the actual oracle to be used for discovering the price
// TODO: Add a config flag so that this can be configured via command line
// For now it will return a default one
func NewPriceOracle() PriceOracle {
	cpo := &ConfigurablePriceOracle{
		honeyPrice: defaultHoneyPrice,
	}
	cpo.refreshRate()
	return cpo
}

// ConfigurablePriceOracle is a price oracle which can be customized.
// It is the default price oracle used as a placeholder for this iteration of the implementation.
// In production this should probably be some on-chain oracle called remotely
type ConfigurablePriceOracle struct {
	honeyPrice  uint64
	lastUpdated time.Time
}

// GetPrice returns the actual price for honey
func (cpo *ConfigurablePriceOracle) GetPrice(honey int64) (int64, error) {
	if time.Now().After(cpo.lastUpdated.Add(maxPriceAge)) {
		cpo.refreshRate()
	}
	return honey * int64(cpo.honeyPrice), nil
}

// refreshRate refreshes the honey to output unit price (currently ETH)
func (cpo *ConfigurablePriceOracle) refreshRate() (uint64, error) {
	return cpo.honeyPrice, nil
}
