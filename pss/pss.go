// Copyright 2018 The go-ethereum Authors
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

package pss

import (
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethersphere/swarm/pss/message"
)

// Send a message without encryption
//TODO: CHANGE TARGETS TO APPROPIATE TYPE
//
func Send(targets []PssAddress, topic message.Topic, msg []byte) error {
	var errors error
	for target := range targets {
		error := p.send([]byte{}, topic, msg)
	}

}

// Send is payload agnostic, and will accept any byte slice as payload
// It generates an envelope for the specified recipient and topic,
// and wraps the message payload in it.
// TODO: Implement proper message padding
func send(to []byte, topic message.Topic, msg []byte) error {

	//construct message with tc
	//send chunk via localstore
	//Asymetric Crypto ()
	//Register Api
	//no Api for the moment
	//for second stage, use tags --> listen for response of recipient, recipient offline
	//Mock store
	//Call send

	metrics.GetOrRegisterCounter("globalpinning/send", nil).Inc(1)
	//construct message with tc
	//send chunk via localstore

	return nil
}
