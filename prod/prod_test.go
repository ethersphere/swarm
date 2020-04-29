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
	"testing"

	"github.com/ethersphere/swarm/pss"
	"github.com/ethersphere/swarm/pss/trojan"
)

// TestRecoveryHook tests that a timeout in netstore
// invokes correctly recovery hook
func TestRecoveryHook(t *testing.T) {
	// setup recovery hook
	// verify that hook is correctly invoked
	ctx := context.TODO()

	handlerVerifier := 0 // test variable to check handler funcs are correctly retrieved

	// register first handler
	testHandler := func(ctx context.Context, targets [][]byte, topic trojan.Topic, payload []byte) (*pss.Monitor, error) {
		handlerVerifier = 1
		// what should I return?
		return nil, nil
	}

	//prod := NewProd(testHandler)

	// call prod Recovery and verify it's been called
	//prod.Recover(ctx, chunk.ZeroAddr)
	//if handlerVerifier != 1 {
	//	t.Fatalf("unexpected result for prod Recover func, expected test variable to have a value of %v but is %v instead", 1, handlerVerifier)
	//}

}

// TestSenderCall verifies that pss send is being called correctly
func TestSenderCall(t *testing.T) {
	// setup netstore
	// setup recovery hook with pss Sender
	// wait for timeout
	// verify that pss was actually called
}
