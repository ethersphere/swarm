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
package outbox

import (
	"github.com/ethersphere/swarm/log"
	"github.com/ethersphere/swarm/pss/message"
)

const (
	defaultOutboxCapacity = 1000
)

var mockForwardFunction = func(msg *message.Message) error {

	log.Debug("Forwarded message", "msg", msg)
	return nil
}

// NewMock creates an Outbox mock. Config can be nil, in that case default values will be set.
func NewMock(config *Config) (outboxMock *Outbox) {
	if config == nil {
		config = &Config{
			NumberSlots: defaultOutboxCapacity,
			Forward:     mockForwardFunction,
		}
	} else {
		if config.Forward == nil {
			config.Forward = mockForwardFunction
		}
		if config.NumberSlots == 0 {
			config.NumberSlots = defaultOutboxCapacity
		}
	}
	return NewOutbox(config)
}
