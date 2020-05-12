// Copyright 2020 The Swarm Authors
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

package prod

import (
	"context"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
)

// RecoveryHook defines code to be executed upon trigger of failed to be retrieved chunks
type RecoveryHook func(ctx context.Context, chunkAddress chunk.Address) error

// sender is the function call for sending trojan chunks
type sender func(ctx context.Context, targets trojan.Targets, topic trojan.Topic, payload []byte) (*pss.Monitor, error)

// NewRecoveryHook returns a new RecoveryHook with the sender function defined
func NewRecoveryHook(send sender) RecoveryHook {
	return func(ctx context.Context, chunkAddress chunk.Address) error {
		targets, err := getPinners(chunkAddress)
		if err != nil {
			return err
		}
		payload := chunkAddress
		topic := trojan.NewTopic("RECOVERY")

		// TODO: monitor return should
		if _, err := send(ctx, targets, topic, payload); err != nil {
			return err
		}
		return nil
	}
}

// TODO: refactor this method to implement feed of target pinners
// getPinners returns the specific target pinners for a corresponding chunk address
func getPinners(chunkAddress chunk.Address) (trojan.Targets, error) {

	// TODO: dummy targets for now
	t1 := trojan.Target([]byte{57, 120})
	t2 := trojan.Target([]byte{209, 156})
	t3 := trojan.Target([]byte{156, 38})
	return trojan.Targets([]trojan.Target{t1, t2, t3}), nil
}
