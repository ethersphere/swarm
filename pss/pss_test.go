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

package pss

import (
	"context"
	"crypto/rand"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/chunk"
	trojan "github.com/ethersphere/swarm/pss/trojan"
	"github.com/ethersphere/swarm/storage/localstore"
)

// TestTrojanChunkRetrieval creates a trojan chunk
// mocks the localstore
// calls pss.Send method and verifies it's properly stored
func TestTrojanChunkRetrieval(t *testing.T) {
	var err error
	ctx := context.TODO()

	localStore := newMockLocalStore(t)
	pss := NewPss(localStore)

	testTargets := [][]byte{
		{57, 120},
		{209, 156},
		{156, 38},
		{89, 19},
		{22, 129}}
	payload := []byte("RECOVERY CHUNK")
	topic := trojan.NewTopic("RECOVERY")

	var ch chunk.Chunk

	// call Send to store trojan chunk in localstore
	if ch, err = pss.Send(ctx, testTargets, topic, payload); err != nil {
		t.Fatal(err)
	}

	// verify store, that trojan chunk has been stored correctly
	var storedChunk chunk.Chunk
	if storedChunk, err = localStore.Get(ctx, chunk.ModeGetRequest, ch.Address()); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ch, storedChunk) {
		t.Fatalf("store chunk does not match sent chunk")
	}

	// check if pinning makes a difference

}

func newMockLocalStore(t *testing.T) *localstore.DB {
	dir, err := ioutil.TempDir("", "swarm-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseKey := make([]byte, 32)
	if _, err = rand.Read(baseKey); err != nil {
		t.Fatal(err)
	}

	localStore, err := localstore.New(dir, baseKey, &localstore.Options{})
	if err != nil {
		t.Fatal(err)
	}

	return localStore
}

// TestRegister verifies that handler funcs are able to be registered correctly in pss
func TestRegister(t *testing.T) {
	localStore := newMockLocalStore(t)
	pss := NewPss(localStore)

	// pss handlers should be empty
	if len(pss.handlers) != 0 {
		t.Fatalf("expected pss handlers to contain 0 elements, but its length is %d", len(pss.handlers))
	}

	handlerVerifier := 0 // test variable to check handler funcs are correctly retrieved

	// register first handler
	testHandler := func(m trojan.Message) error {
		handlerVerifier = 1
		return nil
	}
	testTopic := trojan.NewTopic("FIRST_HANDLER")
	pss.Register(testTopic, testHandler)

	if len(pss.handlers) != 1 {
		t.Fatalf("expected pss handlers to contain 1 element, but its length is %d", len(pss.handlers))
	}

	registeredHandler := pss.getHandler(testTopic)
	registeredHandler(trojan.Message{}) // call handler to verify the retrieved func is correct

	if handlerVerifier != 1 {
		t.Fatal("unexpected handler retrieved")
	}

	// register second handler
	testHandler = func(m trojan.Message) error {
		handlerVerifier = 2
		return nil
	}
	testTopic = trojan.NewTopic("SECOND_HANDLER")
	pss.Register(testTopic, testHandler)
	if len(pss.handlers) != 2 {
		t.Fatalf("expected pss handlers to contain 2 elements, but its length is %d", len(pss.handlers))
	}

	registeredHandler = pss.getHandler(testTopic)
	registeredHandler(trojan.Message{}) // call handler to verify the retrieved func is correct

	if handlerVerifier != 2 {
		t.Fatalf("unexpected handler retrieved, verifier variable should be %d but is %d instead", 2, handlerVerifier)
	}
}

func TestDeliver(t *testing.T) {
	localStore := newMockLocalStore(t)
	pss := NewPss(localStore)

	// test message
	topic := trojan.NewTopic("footopic")
	payload := []byte("foopayload")
	msg, err := trojan.NewMessage(topic, payload)
	if err != nil {
		t.Fatal(err)
	}
	// test chunk
	targets := [][]byte{{255}}
	chunk, err := msg.Wrap(targets)
	if err != nil {
		t.Fatal(err)
	}

	// create and register handler
	var tt trojan.Topic // test variable to check handler func was correctly called
	hndlr := func(m trojan.Message) error {
		tt = m.Topic // copy the message topic to the test variable
		return nil
	}
	pss.Register(topic, hndlr)

	// deliver on chunk and verify test topic variable value change
	pss.Deliver(chunk)
	if tt != msg.Topic {
		t.Fatalf("unexpected result for pss Deliver func, expected test variable to have a value of %v but is %v instead", msg.Topic, tt)
	}

}

// TODO: later test could be a simulation test for 2 nodes, localstore + netstore
