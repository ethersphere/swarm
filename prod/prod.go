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

// RecoveryHook defines code to be executed upon trigger of lost chunks
type RecoveryHook func(ctx context.Context, chunkAddress chunk.Address) error

type sender func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error)

// NewRecoveryHook TODO: better comment please
func NewRecoveryHook(send sender) RecoveryHook {
	return func(ctx context.Context, chunkAddress chunk.Address) error {
		targets, err := getPinners(chunkAddress)
		if err != nil {
			return err
		}
		payload := chunkAddress
		topic := trojan.NewTopic("RECOVERY")

		if _, err := send(ctx, targets, topic, payload); err != nil {
			return err
		}
		return nil
	}
}

// // Recover invokes underlying pss.Send as the first step of global pinning
// func Recover(ctx context.Context, chunkAddress chunk.Address) error {

// 	// TODO: REMOVE FROM HERE DO IT OUTSIDE
// 	// TODO: where to get chunk from??, localstore/netstore etc?
// 	// TODO: does it obtain it from RequestFromPeer?
// 	// for {
// 	// 	select {
// 	// 	case <-time.After(timeouts.FetcherGlobalTimeout):
// 	// 		return nil, errors.New("unable to retreive globally pinned chunk")
// 	// 	case <-time.After(timeouts.SearchTimeout):
// 	// 		break
// 	// 	}
// 	// }

// 	// should it wait for the chunk and timeout if not
// 	// using the monitor of pss until this is received
// 	return nil
// }

func getPinners(chunkAddress chunk.Address) ([][]byte, error) {
	//this should get the feed and return correct target of pinners
	return [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}, nil
}
