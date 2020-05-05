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
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethersphere/swarm/api"
	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage/feed"
)

// RecoveryHook defines code to be executed upon trigger of failed to be retrieved chunks
type RecoveryHook func(ctx context.Context, chunkAddress chunk.Address) error

// sender is the function call for sending trojan chunks
type sender func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error)

// NewRecoveryHook returns a new RecoveryHook with the sender function defined
func NewRecoveryHook(send sender) RecoveryHook {
	return func(ctx context.Context, chunkAddress chunk.Address) error {
		targets, err := getPinners(publisher)
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

// getPinners returns the specific target pinners for a corresponding chunk address
func getPinners(publisher string) ([][]byte, error) {
	// get feed user from publisher
	publisherBytes, err := hex.DecodeString(publisher)
	if err != nil {
		return nil, api.ErrDecrypt
	}
	pubKey, err := crypto.DecompressPubkey(publisherBytes)
	addr := crypto.PubkeyToAddress(*pubKey)

	// TODO: use existing swarm handler
	fhParams := &feed.HandlerParams{}
	fh := feed.NewHandler(fhParams)

	// read feed
	// TODO: resolve sinful type conversions
	fd := feed.Feed{
		Topic: feed.Topic(trojan.NewTopic("RECOVERY")),
		User:  common.Address(addr),
	}

	// TODO: do we need query lookup before content fetch?
	// query := feed.NewQuery(&fd, uint64(time.Now().Unix()), lookup.NoClue)
	// TODO: context.WithCancel?
	// _, err = fh.Lookup(context.Background(), query)
	// if err != nil {
	// 	return nil, err
	// }

	// TODO: time-outs?
	_, content, err := fh.GetContent(&fd)
	if err != nil {
		return nil, err
	}

	// TODO: transform content into actual list of targets
	return [][]byte{content}, nil
}
