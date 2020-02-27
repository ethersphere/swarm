// Copyright 2017 The go-ethereum Authors
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

/*
Package protocols is an extension to p2p. It offers a user friendly simple way to define
devp2p subprotocols by abstracting away code standardly shared by protocols.

* automate assignments of code indexes to messages
* automate RLP decoding/encoding based on reflecting
* provide the forever loop to read incoming messages
* standardise error handling related to communication
* standardised	handshake negotiation
* TODO: automatic generation of wire protocol specification for peers

*/
package protocols

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/tracing"
)

// MsgPauser can be used to pause run execution
// IMPORTANT: should be used only for tests
type MsgPauser interface {
	Pause()
	Resume()
	Wait()
}

//For accounting, the design is to allow the Spec to describe which and how its messages are priced
//To access this functionality, we provide a Hook interface which will call accounting methods
//NOTE: there could be more such (horizontal) hooks in the future
type Hook interface {
	// A hook for applying accounting
	Apply(peer *Peer, costToLocalNode int64, size uint32) error
	// Run some validation before applying accounting
	Validate(peer *Peer, size uint32, msg interface{}, payer Payer) (int64, error)
}

// Spec is a protocol specification including its name and version as well as
// the types of messages which are exchanged
type Spec struct {
	// Name is the name of the protocol, often a three-letter word
	Name string

	// Version is the version number of the protocol
	Version uint

	// MaxMsgSize is the maximum accepted length of the message payload
	MaxMsgSize uint32

	// Messages is a list of message data types which this protocol uses, with
	// each message type being sent with its array index as the code (so
	// [&foo{}, &bar{}, &baz{}] would send foo, bar and baz with codes
	// 0, 1 and 2 respectively)
	// each message must have a single unique data type
	Messages []interface{}

	//hook for accounting (could be extended to multiple hooks in the future)
	Hook Hook

	initOnce sync.Once
	codes    map[reflect.Type]uint64
	types    map[uint64]reflect.Type

	// if the protocol does not allow extending the p2p msg to propagate context
	// even if context not disabled, context will propagate only tracing is enabled
	DisableContext bool
}

func (s *Spec) init() {
	s.initOnce.Do(func() {
		s.codes = make(map[reflect.Type]uint64, len(s.Messages))
		s.types = make(map[uint64]reflect.Type, len(s.Messages))
		for i, msg := range s.Messages {
			code := uint64(i)
			typ := reflect.TypeOf(msg)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			s.codes[typ] = code
			s.types[code] = typ
		}
	})
}

// Length returns the number of message types in the protocol
func (s *Spec) Length() uint64 {
	return uint64(len(s.Messages))
}

// GetCode returns the message code of a type, and boolean second argument is
// false if the message type is not found
func (s *Spec) GetCode(msg interface{}) (uint64, bool) {
	s.init()
	typ := reflect.TypeOf(msg)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	code, ok := s.codes[typ]
	return code, ok
}

// NewMsg construct a new message type given the code
func (s *Spec) NewMsg(code uint64) (interface{}, bool) {
	s.init()
	typ, ok := s.types[code]
	if !ok {
		return nil, false
	}
	return reflect.New(typ).Interface(), true
}

// Peer represents a remote peer or protocol instance that is running on a peer connection with
// a remote peer
type Peer struct {
	*p2p.Peer                         // the p2p.Peer object representing the remote
	rw              p2p.MsgReadWriter // p2p.MsgReadWriter to send messages to and read messages from
	spec            *Spec
	encode          func(context.Context, interface{}) (interface{}, int, error)
	decode          func(p2p.Msg) (context.Context, []byte, error)
	wg              sync.WaitGroup
	running         bool         // if running is true async go routines are dispatched in the event loop
	mtx             sync.RWMutex // guards running
	handleMsgPauser MsgPauser    //  message pauser, should be used only in tests
}

// NewPeer constructs a new peer
// this constructor is called by the p2p.Protocol#Run function
// the first two arguments are the arguments passed to p2p.Protocol.Run function
// the third argument is the Spec describing the protocol
func NewPeer(peer *p2p.Peer, rw p2p.MsgReadWriter, spec *Spec) *Peer {
	encode := encodeWithContext
	decode := decodeWithContext
	if spec == nil || spec.DisableContext || !tracing.Enabled {
		encode = encodeWithoutContext
		decode = decodeWithoutContext
	}
	return &Peer{
		Peer:   peer,
		rw:     rw,
		spec:   spec,
		encode: encode,
		decode: decode,
	}
}

// Run starts the forever loop that handles incoming messages.
// The handler argument is a function which is called for each message received
// from the remote peer, a returned error causes the loop to exit
// resulting in disconnection of the protocol
func (p *Peer) Run(handler func(ctx context.Context, msg interface{}) error) error {
	if err := p.run(handler); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// run receives messages from the peer and dispatches async routines to handle the messages
func (p *Peer) run(handler func(ctx context.Context, msg interface{}) error) error {
	p.mtx.Lock()
	p.running = true
	p.mtx.Unlock()

	for {
		msg, err := p.readMsg()
		if err != nil {
			return err
		}

		p.mtx.RLock()
		// if loop has been stopped, we don't dispatch any more async routines and discard (consume) the message
		if !p.running {
			_ = msg.Discard()
			p.mtx.RUnlock()
			continue
		}
		p.mtx.RUnlock()

		// handleMsgPauser should not be nil only in tests.
		// It does not use mutex lock protection and because of that
		// it must be set before the Registry is constructed and
		// reset when it is closed, in tests.
		// Production performance impact can be considered as
		// neglectable as nil check is a ns order operation.
		if p.handleMsgPauser != nil {
			p.handleMsgPauser.Wait()
		}

		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			err := p.handleMsg(msg, handler)
			if err != nil {
				var e *breakError
				if errors.As(err, &e) {
					p.Drop(err.Error())
				} else {
					log.Warn(err.Error())
				}
			}
		}()
	}
}

func (p *Peer) readMsg() (p2p.Msg, error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		if err != io.EOF {
			metrics.GetOrRegisterCounter("peer.readMsg.error", nil).Inc(1)
			return msg, fmt.Errorf("peer.readMsg, err: %w", err)
		}
	}

	return msg, err
}

// Drop disconnects a peer
// TODO: may need to implement protocol drop only? don't want to kick off the peer
func (p *Peer) Drop(reason string) {
	log.Error("dropping peer with DiscSubprotocolError", "peer", p.ID(), "reason", reason)
	p.Disconnect(p2p.DiscSubprotocolError)
}

// Stop stops the execution of new async jobs, and blocks until active jobs are finished or provided timeout passes.
// Returns nil if the active jobs are finished within the timeout duration, or error otherwise.
func (p *Peer) Stop(timeout time.Duration) error {
	p.mtx.Lock()
	if !p.running {
		return nil
	}

	p.running = false
	p.mtx.Unlock()

	done := make(chan bool)
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		log.Debug("peer shutdown with still active handlers: {}", p)
		return errors.New("shutdown timeout reached")
	}

	return nil
}

// Send takes a message, encodes it in RLP, finds the right message code and sends the
// message off to the peer
// this low level call will be wrapped by libraries providing routed or broadcast sends
// but often just used to forward and push messages to directly connected peers
func (p *Peer) Send(ctx context.Context, msg interface{}) error {
	defer metrics.GetOrRegisterResettingTimer("peer.send_t", nil).UpdateSince(time.Now())
	metrics.GetOrRegisterCounter("peer.send", nil).Inc(1)
	metrics.GetOrRegisterCounter(fmt.Sprintf("peer.send.%T", msg), nil).Inc(1)

	code, found := p.spec.GetCode(msg)
	if !found {
		return fmt.Errorf("invalid message type %v ", code)
	}

	wmsg, size, err := p.encode(ctx, msg)
	if err != nil {
		return err
	}

	// if size is not set by the wrapper, need to serialise
	if size == 0 {
		r, err := rlp.EncodeToBytes(msg)
		if err != nil {
			return err
		}
		size = len(r)
	}

	// if the accounting hook is set, do accounting logic
	if p.spec.Hook != nil {
		// validate that this operation would succeed...
		costToLocalNode, err := p.spec.Hook.Validate(p, uint32(size), wmsg, Sender)
		if err != nil {
			// ...because if it would fail, we return and don't send the message
			return err
		}
		// seems like accounting would succeed, thus send the message first...
		err = p2p.Send(p.rw, code, wmsg)
		if err != nil {
			return err
		}
		// ...and finally apply (write) the accounting change
		if err := p.spec.Hook.Apply(p, costToLocalNode, uint32(size)); err != nil {
			return err
		}
	} else {
		err = p2p.Send(p.rw, code, wmsg)
	}

	return nil
}

// SetMsgPauser sets message pauser for this peer
// IMPORTANT: to be used only for testing
func (p *Peer) SetMsgPauser(pauser MsgPauser) {
	p.handleMsgPauser = pauser
}

// receive is a sync call that handles incoming message with provided message handler
func (p *Peer) receive(handler func(ctx context.Context, msg interface{}) error) error {
	msg, err := p.readMsg()
	if err != nil {
		return err
	}

	return p.handleMsg(msg, handler)
}

// handleMsg is handling message with provided handler. It:
// * checks message size,
// * checks for out-of-range message codes,
// * handles decoding with reflection,
// * call handlers as callbacks
func (p *Peer) handleMsg(msg p2p.Msg, handle func(ctx context.Context, msg interface{}) error) error {
	// make sure that the payload has been fully consumed
	defer msg.Discard()

	if msg.Size > p.spec.MaxMsgSize {
		return Break(fmt.Errorf("message too long: %v > %v", msg.Size, p.spec.MaxMsgSize))
	}

	val, ok := p.spec.NewMsg(msg.Code)
	if !ok {
		return Break(fmt.Errorf("invalid message code: %v", msg.Code))
	}

	ctx, msgBytes, err := p.decode(msg)
	if err != nil {
		return Break(fmt.Errorf("invalid message (RLP error): %v err=%w", msg.Code, err))
	}

	if err := rlp.DecodeBytes(msgBytes, val); err != nil {
		return Break(fmt.Errorf("invalid message (RLP error): <= %v: %w", msg, err))
	}

	// if the accounting hook is set, do accounting logic
	if p.spec.Hook != nil {
		size := uint32(len(msgBytes))

		// validate that the accounting call would succeed...
		costToLocalNode, err := p.spec.Hook.Validate(p, size, val, Receiver)
		if err != nil {
			// ...because if it would fail, we return and don't handle the message
			return Break(err)
		}

		// seems like accounting would be fine, so handle the message
		if err := handle(ctx, val); err != nil {
			return fmt.Errorf("message handler: (msg code %v): %w", msg.Code, err)
		}

		// handling succeeded, finally apply accounting
		if err := p.spec.Hook.Apply(p, costToLocalNode, size); err != nil {
			return Break(err)
		}
	} else {
		// call the registered handler callbacks
		// a registered callback take the decoded message as argument as an interface
		// which the handler is supposed to cast to the appropriate type
		// it is entirely safe not to check the cast in the handler since the handler is
		// chosen based on the proper type in the first place
		if err := handle(ctx, val); err != nil {
			return fmt.Errorf("message handler: (msg code %v): %w", msg.Code, err)
		}
	}

	return nil
}

// Handshake negotiates a handshake on the peer connection
// * arguments
//   * context
//   * the local handshake to be sent to the remote peer
//   * function to be called on the remote handshake (can be nil)
// * expects a remote handshake back of the same type
// * the dialing peer needs to send the handshake first and then waits for remote
// * the listening peer waits for the remote handshake and then sends it
// returns the remote handshake and an error
func (p *Peer) Handshake(ctx context.Context, hs interface{}, verify func(interface{}) error) (interface{}, error) {
	if _, ok := p.spec.GetCode(hs); !ok {
		return nil, fmt.Errorf("unknown handshake message type: %T", hs)
	}

	var rhs interface{}
	errc := make(chan error, 2)

	send := func() { errc <- p.Send(ctx, hs) }
	receive := func() {
		errc <- p.receive(func(ctx context.Context, msg interface{}) error {
			rhs = msg
			if verify != nil {
				return verify(rhs)
			}
			return nil
		})
	}

	go func() {
		if p.Inbound() {
			receive()
			send()
		} else {
			send()
			receive()
		}
	}()

	for i := 0; i < 2; i++ {
		var err error
		select {
		case err = <-errc:
		case <-ctx.Done():
			err = ctx.Err()
		}
		if err != nil {
			return nil, err
		}
	}
	return rhs, nil
}

// HasCap returns true if Peer has a capability
// with provided name.
func (p *Peer) HasCap(capName string) (yes bool) {
	if p == nil || p.Peer == nil {
		return false
	}
	for _, c := range p.Caps() {
		if c.Name == capName {
			return true
		}
	}
	return false
}
