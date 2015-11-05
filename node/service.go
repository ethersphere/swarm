// Copyright 2015 The go-ethereum Authors
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

package node

import "github.com/ethereum/go-ethereum/p2p"

// Service is an individual protocol that can be registered into a node.
//
// Notes:
//  - Service life-cycle management is delegated to the node. The service is
//    allowed to initialize itself upon creation, but no goroutines should be
//    spun up outside of the Start method.
//  - Restart logic is not required as the node will create a fresh instance
//    every time a service is started.
type Service interface {
	// Protocol retrieves the P2P protocols the service wishes to start.
	Protocols() []p2p.Protocol

	// Start spawns any goroutines required by the service.
	Start() error

	// Stop terminates all goroutines belonging to the service, blocking until they
	// are all terminated.
	Stop() error
}
