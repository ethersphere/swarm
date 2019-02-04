package pss

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/network/simulation"
	"github.com/ethereum/go-ethereum/swarm/pot"
	"github.com/ethereum/go-ethereum/swarm/state"
)

const (
	minProxBinSize = 2
)

var (
	runNodes      = flag.Int("nodes", 0, "nodes to start in the network")
	runMessages   = flag.Int("messages", 0, "messages to send during test")
	stableTimeout = flag.Int("timeout.stable", 10, "timeout in seconds for network to stabilize")
)

// needed to make the enode id of the receiving node available to the handler for triggers
type handlerContextFunc func(*adapters.NodeConfig) *handler

// struct to notify reception of messages to simulation driver
// TODO To make code cleaner:
// - consider a separate pss unwrap to message event in sim framework (this will make eventual message propagation analysis with pss easier/possible in the future)
// - consider also test api calls to inspect handling results of messages
type handlerNotification struct {
	id     enode.ID
	serial uint64
}

func TestProxNetwork(t *testing.T) {
	if (*runNodes > 0 && *runMessages == 0) || (*runMessages > 0 && *runNodes == 0) {
		log.Crit("cannot specify only one of flags --nodes and --messages")
	}

	if *runNodes > 0 {
		t.Run(fmt.Sprintf("%d/%d", *runMessages, *runNodes), testProxNetwork)
		return
	}
	t.Run("1/32", testProxNetwork)
}

// This tests generates a sequenced number of messages with random addresses
// It then calculates which nodes in the network have the address of each message within their nearest neighborhood depth, and stores them as recipients
// Upon sending the messages, it verifies that the respective message is passed to the message handlers of these recipients
// It will fail if a recipient handles a message it should not, or if after propagation not all expected messages are handled (timeout)
func testProxNetwork(t *testing.T) {

	args := strings.Split(t.Name(), "/")
	msgCount, err := strconv.ParseInt(args[1], 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	nodeCount, err := strconv.ParseInt(args[2], 10, 16)
	if err != nil {
		t.Fatal(err)
	}

	topic := BytesToTopic([]byte{0x00, 0x00, 0x06, 0x82})

	// passes message from pss message handler to simulation driver
	handlerC := make(chan handlerNotification)

	// set to true on termination of the simulation run
	var handlerDone bool

	// keeps handlerDonc in sync
	mu := &sync.Mutex{}

	// message handler for pss
	handlerContextFuncs := map[Topic]handlerContextFunc{
		topic: func(ctx *adapters.NodeConfig) *handler {
			return &handler{
				f: func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {

					// using simple serial in message body
					// makes it easy to keep track of who's getting what
					serial, c := binary.Uvarint(msg)
					if c <= 0 {
						t.Fatalf("corrupt message received by %x (uvarint parse returned %d)", ctx.ID, c)
					}

					// terminate if sim is over
					mu.Lock()
					if handlerDone {
						mu.Unlock()
						return errors.New("handlers aborted")
					}
					mu.Unlock()

					// pass message context to the listener in the simulation
					handlerC <- handlerNotification{
						id:     ctx.ID,
						serial: serial,
					}
					return nil
				},
				caps: &handlerCaps{
					raw:  true, // we use raw messages for simplicity
					prox: true,
				},
			}
		},
	}

	// TODO refactor swarm sim to enable access to kademlias from the sim obj
	kademlias := make(map[enode.ID]*network.Kademlia)
	sim := simulation.New(newProxServices(true, handlerContextFuncs, kademlias))
	defer sim.Close()

	// start network
	// TODO: use snapshot and skip until SKIP
	_, err = sim.AddNodesAndConnectRing(int(nodeCount))
	if err != nil {
		t.Fatal(err)
	}

	// make predictable overlay addresses from the generated random enode ids
	nodeAddrs := make(map[enode.ID][]byte)
	for _, nodeId := range sim.NodeIDs() {
		nodeAddrs[nodeId] = nodeIDToAddr(nodeId)
	}

	// at least in cyberspace we can be health freaks
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ill, err := sim.WaitTillHealthy(ctx, minProxBinSize)
	if err != nil {
		// inspect the latest detected not healthy kademlias
		for id, kad := range ill {
			log.Debug("Node not healthy", id)
			log.Trace(kad.String())
		}
		t.Fatal(err)
	}

	// wait until network is stable, too
	secondMultiplier := 5
	pulse, err := time.ParseDuration(fmt.Sprintf("%ds", secondMultiplier))
	if err != nil {
		t.Fatal(err)
	}
	timeoutMultiplier := *stableTimeout
	timeout, err := time.ParseDuration(fmt.Sprintf("%ds", timeoutMultiplier))
	if err != nil {
		t.Fatal(err)
	}
	if !serenityNowPlease(sim, pulse, timeout) {
		t.Fatalf("network not stable after %ds", timeoutMultiplier)
	}
	log.Debug("network stable", "threshold", pulse)
	// SKIP <- skip till here

	// generate messages and index them
	pof := pot.DefaultPof(256)

	// recipient addresses of messages
	var msgs [][]byte

	// for logging output only
	recipients := make(map[int][]enode.ID)

	// message serials we expect respective nodes to receive
	expectedMsgs := make(map[enode.ID][]uint64)

	// originating nodes of the messages
	// intention is to choose as far as possible from the receiving neighborhood
	senders := make(map[int]enode.ID)

	// total count of messages to receive, used for terminating the simulation run
	var msgsToReceive int

	for i := 0; i < int(msgCount); i++ {

		// we choose message addresses by random
		msgAddr := pot.RandomAddress()
		msgs = append(msgs, msgAddr.Bytes())
		smallestPo := 256

		// loop through all nodes and add the message to receipient indices
		for _, nod := range sim.Net.GetNodes() {
			po, _ := pof(msgs[i], nodeAddrs[nod.ID()], 0)
			depth := kademlias[nod.ID()].NeighbourhoodDepth()

			// node has message address within nearest neighborhood depth
			// that means it is a recipient
			if po >= depth {
				recipients[i] = append(recipients[i], nod.ID())
				expectedMsgs[nod.ID()] = append(expectedMsgs[nod.ID()], uint64(i))
				msgsToReceive++
			}

			// keep track of the smallest po value in the iteration
			// the first node in the smallest value bin
			// will be the sender
			if po < smallestPo {
				smallestPo = po
				senders[i] = nod.ID()
			}
		}
		log.Debug("nn for msg", "rcptcount", len(recipients[i]), "msgidx", i, "msg", common.Bytes2Hex(msgs[i]), "sender", senders[i], "senderpo", smallestPo)
	}
	log.Debug("msgs to receive", "count", msgsToReceive)

	// simulation run function
	runFunc := func(ctx context.Context, sim *simulation.Simulation) error {

		// terminates the handler channel listener
		doneC := make(chan struct{})

		// error to pass to main sim thread
		errC := make(chan error)

		// message receipt notification to main sim thread
		msgC := make(chan handlerNotification)

		// handler channel listener
		go func(errC chan error, doneC chan struct{}, msgC chan handlerNotification) {
			for {
				select {

				// everything a-ok
				case <-doneC:
					mu.Lock()
					handlerDone = true
					mu.Unlock()
					errC <- nil
					return

				// timeout or cancel
				case <-ctx.Done():
					mu.Lock()
					handlerDone = true
					mu.Unlock()
					errC <- ctx.Err()
					return

				// incoming message from pss message handler
				case handlerNotification := <-handlerC:

					// for syntax brevity below
					xMsgs := expectedMsgs[handlerNotification.id]

					// check if recipient has already received all its messages
					// and notify to fail the test if so
					if len(xMsgs) == 0 {
						mu.Lock()
						handlerDone = true
						mu.Unlock()
						errC <- fmt.Errorf("too many messages received by recipient %x", handlerNotification.id)
						return
					}

					// check if message serial is in expected messages for this recipient
					// and notify to fail the test if not
					idx := -1
					for i, msg := range xMsgs {
						if handlerNotification.serial == msg {
							idx = i
							break
						}
					}
					if idx == -1 {
						mu.Lock()
						handlerDone = true
						mu.Unlock()
						errC <- fmt.Errorf("message %d received by wrong recipient %v", handlerNotification.serial, handlerNotification.id)
						return
					}

					// message is ok, so remove that message serial from the recipient expectation array and notify the main sim thread
					xMsgs[idx] = xMsgs[len(xMsgs)-1]
					xMsgs = xMsgs[:len(xMsgs)-1]
					msgC <- handlerNotification
				}
			}
		}(errC, doneC, msgC)

		// send the messages
		go func(msgs [][]byte, senders map[int]enode.ID, sim *simulation.Simulation) {
			for i, msg := range msgs {
				log.Debug("sending msg", "idx", i, "from", senders[i])
				nodeClient, err := sim.Net.GetNode(senders[i]).Client()
				if err != nil {
					t.Fatal(err)
				}
				var uvarByte [8]byte
				binary.PutUvarint(uvarByte[:], uint64(i))
				nodeClient.Call(nil, "pss_sendRaw", hexutil.Encode(msg), topic, uvarByte[:])
			}
		}(msgs, senders, sim)

		// collect incoming messages
		// and terminate with corresponding status
		// when message handler listener ends
		msgsCountdown := msgsToReceive
	OUTER:
		for {
			select {
			case err := <-errC:
				if err != nil {
					return err
				}
				break OUTER
			case hn := <-msgC:
				msgsCountdown--
				log.Debug("msg left", "count", msgsCountdown, "total", msgsToReceive, "id", hn.id, "serial", hn.serial)
				if msgsCountdown == 0 {
					close(doneC)
				}
			}
		}

		return nil
	}

	// run the sim
	result := sim.Run(ctx, runFunc)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	t.Logf("completed %d", result.Duration)
}

// an adaptation of the same services setup as in pss_test.go
// replaces pss_test.go when those tests are rewritten to the new swarm/network/simulation package
func newProxServices(allowRaw bool, handlerContextFuncs map[Topic]handlerContextFunc, kademlias map[enode.ID]*network.Kademlia) map[string]simulation.ServiceFunc {
	stateStore := state.NewInmemoryStore()
	kademlia := func(id enode.ID) *network.Kademlia {
		if k, ok := kademlias[id]; ok {
			return k
		}
		params := network.NewKadParams()
		params.MinProxBinSize = minProxBinSize
		params.MaxBinSize = 3
		params.MinBinSize = 1
		params.MaxRetries = 1000
		params.RetryExponent = 2
		params.RetryInterval = 1000000
		kademlias[id] = network.NewKademlia(id[:], params)
		return kademlias[id]
	}
	return map[string]simulation.ServiceFunc{
		"pss": func(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			// execadapter does not exec init()
			initTest()

			// create keys in whisper and set up the pss object
			ctxlocal, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			keys, err := wapi.NewKeyPair(ctxlocal)
			privkey, err := w.GetPrivateKey(keys)
			pssp := NewPssParams().WithPrivateKey(privkey)
			pssp.AllowRaw = allowRaw
			pskad := kademlia(ctx.Config.ID)
			ps, err := NewPss(pskad, pssp)
			if err != nil {
				return nil, nil, err
			}

			// register the handlers we've been passed
			var deregisters []func()
			for tpc, hndlrFunc := range handlerContextFuncs {
				deregisters = append(deregisters, ps.Register(&tpc, hndlrFunc(ctx.Config)))
			}

			// if handshake mode is set, add the controller
			// TODO: This should be hooked to the handshake test file
			if useHandshake {
				SetHandshakeController(ps, NewHandshakeParams())
			}

			// we expose some api calls for cheating
			ps.addAPI(rpc.API{
				Namespace: "psstest",
				Version:   "0.3",
				Service:   NewAPITest(ps),
				Public:    false,
			})

			// return Pss and cleanups
			return ps, func() {
				// run the handler deregister functions in reverse order
				for i := len(deregisters); i > 0; i-- {
					deregisters[i-1]()
				}
			}, nil
		},
		"bzz": func(ctx *adapters.ServiceContext, b *sync.Map) (node.Service, func(), error) {
			// normally translation of enode id to swarm address is concealed by the network package
			// however, we need to keep track of it in the test driver aswell.
			// if the translation in the network package changes, that can cause thiese tests to unpredictably fail
			// therefore we keep a local copy of the translation here
			addr := network.NewAddr(ctx.Config.Node())
			addr.OAddr = nodeIDToAddr(ctx.Config.Node().ID())

			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			return network.NewBzz(config, kademlia(ctx.Config.ID), stateStore, nil, nil), nil, nil
		},
	}
}

// makes sure we create the addresses the same way in driver and service setup
func nodeIDToAddr(id enode.ID) []byte {
	return id.Bytes()
}

// temporary function for polling a "stable" network
// stability here means no conns or drops in network within a "serenity" duration
// timeout is max time to wait for a "stable" network
// TODO: remove when replaced with snapshot
func serenityNowPlease(sim *simulation.Simulation, serenity time.Duration, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	eventFilter := simulation.NewPeerEventsFilter().Connect().Drop()

	eventC := sim.PeerEvents(ctx, sim.NodeIDs(), eventFilter)

	timer := time.NewTimer(serenity)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-timer.C:
			return true
		case <-eventC:
			timer.Reset(serenity)
		}
	}
}
