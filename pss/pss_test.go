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
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethersphere/swarm/network"
	"github.com/ethersphere/swarm/network/simulation"
	"github.com/ethersphere/swarm/p2p/protocols"
	"github.com/ethersphere/swarm/pot"
	"github.com/ethersphere/swarm/pss/crypto"
	"github.com/ethersphere/swarm/state"
	"github.com/ethersphere/swarm/testutil"
)

var (
	initOnce        = sync.Once{}
	psslogmain      log.Logger
	pssprotocols    map[string]*protoCtrl
	useHandshake    bool
	noopHandlerFunc = func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		return nil
	}
)

func init() {
	testutil.Init()
	rand.Seed(time.Now().Unix())
	initTest()
}

func initTest() {
	initOnce.Do(
		func() {
			psslogmain = log.New("psslog", "*")

			pssprotocols = make(map[string]*protoCtrl)
		},
	)
}

// test that topic conversion functions give predictable results
func TestTopic(t *testing.T) {

	api := &API{}

	topicstr := strings.Join([]string{PingProtocol.Name, strconv.Itoa(int(PingProtocol.Version))}, ":")

	// bytestotopic is the authoritative topic conversion source
	topicobj := BytesToTopic([]byte(topicstr))

	// string to topic and bytes to topic must match
	topicapiobj, _ := api.StringToTopic(topicstr)
	if topicobj != topicapiobj {
		t.Fatalf("bytes and string topic conversion mismatch; %s != %s", topicobj, topicapiobj)
	}

	// string representation of topichex
	topichex := topicobj.String()

	// protocoltopic wrapper on pingtopic should be same as topicstring
	// check that it matches
	pingtopichex := PingTopic.String()
	if topichex != pingtopichex {
		t.Fatalf("protocol topic conversion mismatch; %s != %s", topichex, pingtopichex)
	}

	// json marshal of topic
	topicjsonout, err := topicobj.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(topicjsonout)[1:len(topicjsonout)-1] != topichex {
		t.Fatalf("topic json marshal mismatch; %s != \"%s\"", topicjsonout, topichex)
	}

	// json unmarshal of topic
	var topicjsonin Topic
	topicjsonin.UnmarshalJSON(topicjsonout)
	if topicjsonin != topicobj {
		t.Fatalf("topic json unmarshal mismatch: %x != %x", topicjsonin, topicobj)
	}
}

// test bit packing of message control flags
func TestMsgParams(t *testing.T) {
	var ctrl byte
	ctrl |= pssControlRaw
	p := newMsgParamsFromBytes([]byte{ctrl})
	m := newPssMsg(p)
	if !m.isRaw() || m.isSym() {
		t.Fatal("expected raw=true and sym=false")
	}
	ctrl |= pssControlSym
	p = newMsgParamsFromBytes([]byte{ctrl})
	m = newPssMsg(p)
	if !m.isRaw() || !m.isSym() {
		t.Fatal("expected raw=true and sym=true")
	}
	ctrl &= 0xff &^ pssControlRaw
	p = newMsgParamsFromBytes([]byte{ctrl})
	m = newPssMsg(p)
	if m.isRaw() || !m.isSym() {
		t.Fatal("expected raw=false and sym=true")
	}
}

// test if we can insert into cache, match items with cache and cache expiry
func TestCache(t *testing.T) {
	var err error
	to, _ := hex.DecodeString("08090a0b0c0d0e0f1011121314150001020304050607161718191a1b1c1d1e1f")
	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	ps := newTestPss(privkey, nil, nil)
	defer ps.Stop()
	pp := NewParams().WithPrivateKey(privkey)
	data := []byte("foo")
	datatwo := []byte("bar")
	datathree := []byte("baz")
	wparams := &crypto.WrapParams{
		Sender:   privkey,
		Receiver: &privkey.PublicKey,
	}
	env, err := ps.Crypto.Wrap(data, wparams)
	msg := &PssMsg{
		Payload: env,
		To:      to,
		Topic:   PingTopic,
	}
	envtwo, err := ps.Crypto.Wrap(datatwo, wparams)
	msgtwo := &PssMsg{
		Payload: envtwo,
		To:      to,
		Topic:   PingTopic,
	}
	envthree, err := ps.Crypto.Wrap(datathree, wparams)
	msgthree := &PssMsg{
		Payload: envthree,
		To:      to,
		Topic:   PingTopic,
	}

	digestone := ps.msgDigest(msg)
	if err != nil {
		t.Fatalf("could not store cache msgone: %v", err)
	}
	digesttwo := ps.msgDigest(msgtwo)
	if err != nil {
		t.Fatalf("could not store cache msgtwo: %v", err)
	}
	digestthree := ps.msgDigest(msgthree)
	if err != nil {
		t.Fatalf("could not store cache msgthree: %v", err)
	}

	if digestone == digesttwo {
		t.Fatalf("different msgs return same hash: %d", digesttwo)
	}

	// check the cache
	err = ps.addFwdCache(msg)
	if err != nil {
		t.Fatalf("write to pss expire cache failed: %v", err)
	}

	if !ps.checkFwdCache(msg) {
		t.Fatalf("message %v should have EXPIRE record in cache but checkCache returned false", msg)
	}

	if ps.checkFwdCache(msgtwo) {
		t.Fatalf("message %v should NOT have EXPIRE record in cache but checkCache returned true", msgtwo)
	}

	time.Sleep(pp.CacheTTL + 1*time.Second)
	err = ps.addFwdCache(msgthree)
	if err != nil {
		t.Fatalf("write to pss expire cache failed: %v", err)
	}

	if ps.checkFwdCache(msg) {
		t.Fatalf("message %v should have expired from cache but checkCache returned true", msg)
	}

	if _, ok := ps.fwdCache[digestthree]; !ok {
		t.Fatalf("unexpired message should be in the cache: %v", digestthree)
	}

	if _, ok := ps.fwdCache[digesttwo]; ok {
		t.Fatalf("expired message should have been cleared from the cache: %v", digesttwo)
	}
}

// matching of address hints; whether a message could be or is for the node
func TestAddressMatch(t *testing.T) {

	localaddr := network.RandomAddr().Over()
	copy(localaddr[:8], []byte("deadbeef"))
	remoteaddr := []byte("feedbeef")
	kadparams := network.NewKadParams()
	kad := network.NewKademlia(localaddr, kadparams)
	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatalf("Could not generate private key: %v", err)
	}
	pssp := NewParams().WithPrivateKey(privkey)
	ps, err := New(kad, pssp)
	if err != nil {
		t.Fatal(err.Error())
	}

	pssmsg := &PssMsg{
		To: remoteaddr,
	}

	// differ from first byte
	if ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient true but %x != %x", remoteaddr, localaddr)
	}
	if ps.isSelfPossibleRecipient(pssmsg, false) {
		t.Fatalf("isSelfPossibleRecipient true but %x != %x", remoteaddr[:8], localaddr[:8])
	}

	// 8 first bytes same
	copy(remoteaddr[:4], localaddr[:4])
	if ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient true but %x != %x", remoteaddr, localaddr)
	}
	if !ps.isSelfPossibleRecipient(pssmsg, false) {
		t.Fatalf("isSelfPossibleRecipient false but %x == %x", remoteaddr[:8], localaddr[:8])
	}

	// all bytes same
	pssmsg.To = localaddr
	if !ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient false but %x == %x", remoteaddr, localaddr)
	}
	if !ps.isSelfPossibleRecipient(pssmsg, false) {
		t.Fatalf("isSelfPossibleRecipient false but %x == %x", remoteaddr[:8], localaddr[:8])
	}

}

// verify that node can be set as recipient regardless of explicit message address match if minimum one handler of a topic is explicitly set to allow it
// note that in these tests we use the raw capability on handlers for convenience
func TestAddressMatchProx(t *testing.T) {

	// recipient node address
	localAddr := network.RandomAddr().Over()
	localPotAddr := pot.NewAddressFromBytes(localAddr)

	// set up kademlia
	kadparams := network.NewKadParams()
	kad := network.NewKademlia(localAddr, kadparams)
	nnPeerCount := kad.MinBinSize
	peerCount := nnPeerCount + 2

	// set up pss
	privKey, err := ethCrypto.GenerateKey()
	pssp := NewParams().WithPrivateKey(privKey)
	ps, err := New(kad, pssp)
	// enqueue method now is blocking, so we need always somebody processing the outbox
	go func() {
		for slot := range ps.outbox.process {
			ps.outbox.free(slot)
		}
	}()
	if err != nil {
		t.Fatal(err.Error())
	}

	// create kademlia peers, so we have peers both inside and outside minproxlimit
	var peers []*network.Peer
	for i := 0; i < peerCount; i++ {
		rw := &p2p.MsgPipeRW{}
		ptpPeer := p2p.NewPeer(enode.ID{}, "362436 call me anytime", []p2p.Cap{})
		protoPeer := protocols.NewPeer(ptpPeer, rw, &protocols.Spec{})
		peerAddr := pot.RandomAddressAt(localPotAddr, i)
		bzzPeer := &network.BzzPeer{
			Peer: protoPeer,
			BzzAddr: &network.BzzAddr{
				OAddr: peerAddr.Bytes(),
				UAddr: []byte(fmt.Sprintf("%x", peerAddr[:])),
			},
		}
		peer := network.NewPeer(bzzPeer, kad)
		kad.On(peer)
		peers = append(peers, peer)
	}

	// TODO: create a test in the network package to make a table with n peers where n-m are proxpeers
	// meanwhile test regression for kademlia since we are compiling the test parameters from different packages
	var proxes int
	var conns int
	depth := kad.NeighbourhoodDepth()
	kad.EachConn(nil, peerCount, func(p *network.Peer, po int) bool {
		conns++
		if po >= depth {
			proxes++
		}
		return true
	})
	if proxes != nnPeerCount {
		t.Fatalf("expected %d proxpeers, have %d", nnPeerCount, proxes)
	} else if conns != peerCount {
		t.Fatalf("expected %d peers total, have %d", peerCount, proxes)
	}

	// remote address distances from localAddr to try and the expected outcomes if we use prox handler
	remoteDistances := []int{
		255,
		nnPeerCount + 1,
		nnPeerCount,
		nnPeerCount - 1,
		0,
	}
	expects := []bool{
		true,
		true,
		true,
		false,
		false,
	}

	// first the unit test on the method that calculates possible receipient using prox
	for i, distance := range remoteDistances {
		pssMsg := newPssMsg(&msgParams{})
		pssMsg.To = make([]byte, len(localAddr))
		copy(pssMsg.To, localAddr)
		var byteIdx = distance / 8
		pssMsg.To[byteIdx] ^= 1 << uint(7-(distance%8))
		log.Trace(fmt.Sprintf("addrmatch %v", bytes.Equal(pssMsg.To, localAddr)))
		if ps.isSelfPossibleRecipient(pssMsg, true) != expects[i] {
			t.Fatalf("expected distance %d to be %v", distance, expects[i])
		}
	}

	// we move up to higher level and test the actual message handler
	// for each distance check if we are possible recipient when prox variant is used is set

	// this handler will increment a counter for every message that gets passed to the handler
	var receives int
	rawHandlerFunc := func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		log.Trace("in allowraw handler")
		receives++
		return nil
	}

	// register it marking prox capability
	topic := BytesToTopic([]byte{0x2a})
	hndlrProxDereg := ps.Register(&topic, &handler{
		f: rawHandlerFunc,
		caps: &handlerCaps{
			raw:  true,
			prox: true,
		},
	})

	// test the distances
	var prevReceive int
	for i, distance := range remoteDistances {
		remotePotAddr := pot.RandomAddressAt(localPotAddr, distance)
		remoteAddr := remotePotAddr.Bytes()

		var data [32]byte
		rand.Read(data[:])
		pssMsg := newPssMsg(&msgParams{raw: true})
		pssMsg.To = remoteAddr
		pssMsg.Expire = uint32(time.Now().Unix() + 4200)
		pssMsg.Payload = data[:]
		pssMsg.Topic = topic

		log.Trace("withprox addrs", "local", localAddr, "remote", remoteAddr)
		ps.handle(context.TODO(), pssMsg)
		if (!expects[i] && prevReceive != receives) || (expects[i] && prevReceive == receives) {
			t.Fatalf("expected distance %d recipient %v when prox is set for handler", distance, expects[i])
		}
		prevReceive = receives
	}

	// now add a non prox-capable handler and test
	ps.Register(&topic, &handler{
		f: rawHandlerFunc,
		caps: &handlerCaps{
			raw: true,
		},
	})
	receives = 0
	prevReceive = 0
	for i, distance := range remoteDistances {
		remotePotAddr := pot.RandomAddressAt(localPotAddr, distance)
		remoteAddr := remotePotAddr.Bytes()

		var data [32]byte
		rand.Read(data[:])
		pssMsg := newPssMsg(&msgParams{raw: true})
		pssMsg.To = remoteAddr
		pssMsg.Expire = uint32(time.Now().Unix() + 4200)
		pssMsg.Payload = data[:]
		pssMsg.Topic = topic

		log.Trace("withprox addrs", "local", localAddr, "remote", remoteAddr)
		ps.handle(context.TODO(), pssMsg)
		if (!expects[i] && prevReceive != receives) || (expects[i] && prevReceive == receives) {
			t.Fatalf("expected distance %d recipient %v when prox is set for handler", distance, expects[i])
		}
		prevReceive = receives
	}

	// now deregister the prox capable handler, now none of the messages will be handled
	hndlrProxDereg()
	receives = 0

	for _, distance := range remoteDistances {
		remotePotAddr := pot.RandomAddressAt(localPotAddr, distance)
		remoteAddr := remotePotAddr.Bytes()

		pssMsg := newPssMsg(&msgParams{raw: true})
		pssMsg.To = remoteAddr
		pssMsg.Expire = uint32(time.Now().Unix() + 4200)
		pssMsg.Payload = []byte(remotePotAddr.String())
		pssMsg.Topic = topic

		log.Trace("noprox addrs", "local", localAddr, "remote", remoteAddr)
		ps.handle(context.TODO(), pssMsg)
		if receives != 0 {
			t.Fatalf("expected distance %d to not be recipient when prox is not set for handler", distance)
		}

	}
}

func TestMessageOutbox(t *testing.T) {
	// setup
	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err.Error())
	}

	addr := make([]byte, 32)
	addr[0] = 0x01
	ps := newTestPssStart(privkey, network.NewKademlia(addr, network.NewKadParams()), NewParams(), false)
	outboxCapacity := 2

	successC := make(chan struct{})
	forward := func(msg *PssMsg) error {
		successC <- struct{}{}
		return nil
	}
	ps.outbox = newOutbox(outboxCapacity, ps.quitC, forward)

	ps.Start(nil)
	defer ps.Stop()

	err = ps.enqueue(testRandomMessage())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	usedSlots := ps.outbox.len()
	if usedSlots != 1 {
		t.Fatalf("incorrect outbox length. expected 1, got %v", usedSlots)
	}
	t.Log("Message enqueued", "Outbox len", ps.outbox.len())

	select {
	case <-successC:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for success forward")
	}

	failed := make([]*PssMsg, 0)
	failedC := make(chan struct{})
	continueC := make(chan struct{})
	failedForward := func(msg *PssMsg) error {
		failed = append(failed, msg)
		failedC <- struct{}{}
		<-continueC
		return errors.New("Forced test error forwarding message")
	}

	ps.outbox.forward = failedForward

	err = ps.enqueue(testRandomMessage())
	if err != nil {
		t.Fatalf("Expected no error enqueing, got %v", err.Error())
	}

	select {
	case <-failedC:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for failing forward")
	}

	if len(failed) == 0 {
		t.Fatal("Incorrect number of failed messages, expected 1 got 0")
	}
	// The message will be retried once we send to continueC, so first, we change the forward function
	ps.outbox.forward = forward
	continueC <- struct{}{}
	select {
	case <-successC:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for second success forward")
	}

}

func TestOutboxFull(t *testing.T) {
	// setup
	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err.Error())
	}

	addr := make([]byte, 32)
	addr[0] = 0x01
	ps := newTestPssStart(privkey, network.NewKademlia(addr, network.NewKadParams()), NewParams(), false)
	defer ps.Stop()
	outboxCapacity := 2

	procChan := make(chan struct{})
	succesForward := func(msg *PssMsg) error {
		<-procChan
		log.Info("Message processed")
		return nil
	}
	ps.outbox = newOutbox(outboxCapacity, ps.quitC, succesForward)

	ps.Start(nil)

	err = ps.enqueue(testRandomMessage())
	if err != nil {
		t.Fatalf("expected no error enqueing first message, got %v", err)
	}
	err = ps.enqueue(testRandomMessage())
	if err != nil {
		t.Fatalf("expected no error enqueing second message, got %v", err)
	}
	//As we haven't signaled procChan, the messages are still in the outbox

	err = ps.enqueue(testRandomMessage())
	if err == nil {
		t.Fatalf("expected error enqueing third message, instead got nil")
	}
	procChan <- struct{}{}
	//There should be a slot again in the outbox
	select {
	case <-ps.outbox.slots:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for a free slot")
	}
}

// set and generate pubkeys and symkeys
func TestKeys(t *testing.T) {
	// make our key and init pss with it
	ourprivkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to retrieve 'our' private key")
	}
	theirprivkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to retrieve 'their' private key")
	}
	ps := newTestPss(ourprivkey, nil, nil)
	defer ps.Stop()

	// set up peer with mock address, mapped to mocked publicaddress and with mocked symkey
	addr := make(PssAddress, 32)
	copy(addr, network.RandomAddr().Over())
	outkey := network.RandomAddr().Over()
	topicobj := BytesToTopic([]byte("foo:42"))
	ps.SetPeerPublicKey(&theirprivkey.PublicKey, topicobj, addr)
	outkeyid, err := ps.SetSymmetricKey(outkey, topicobj, addr, false)
	if err != nil {
		t.Fatalf("failed to set 'our' outgoing symmetric key")
	}

	// make a symmetric key that we will send to peer for encrypting messages to us
	inkeyid, err := ps.GenerateSymmetricKey(topicobj, addr, true)
	if err != nil {
		t.Fatalf("failed to set 'our' incoming symmetric key")
	}

	// get the key back from crypto backend, check that it's still the same
	outkeyback, err := ps.Crypto.GetSymmetricKey(outkeyid)
	if err != nil {
		t.Fatalf(err.Error())
	}
	inkey, err := ps.Crypto.GetSymmetricKey(inkeyid)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !bytes.Equal(outkeyback, outkey) {
		t.Fatalf("passed outgoing symkey doesnt equal stored: %x / %x", outkey, outkeyback)
	}

	t.Logf("symout: %v", outkeyback)
	t.Logf("symin: %v", inkey)

	// check that the key is stored in the peerpool
	psp := ps.symKeyPool[inkeyid][topicobj]
	if !bytes.Equal(psp.address, addr) {
		t.Fatalf("inkey address does not match; %p != %p", psp.address, addr)
	}
}

// check that we can retrieve previously added public key entires per topic and peer
func TestGetPublickeyEntries(t *testing.T) {

	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	ps := newTestPss(privkey, nil, nil)
	defer ps.Stop()

	peeraddr := network.RandomAddr().Over()
	topicaddr := make(map[Topic]PssAddress)
	topicaddr[Topic{0x13}] = peeraddr
	topicaddr[Topic{0x2a}] = peeraddr[:16]
	topicaddr[Topic{0x02, 0x9a}] = []byte{}

	remoteprivkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	remotepubkeybytes := ps.Crypto.SerializePublicKey(&remoteprivkey.PublicKey)
	remotepubkeyhex := common.ToHex(remotepubkeybytes)

	pssapi := NewAPI(ps)

	for to, a := range topicaddr {
		err = pssapi.SetPeerPublicKey(remotepubkeybytes, to, a)
		if err != nil {
			t.Fatal(err)
		}
	}

	intopic, err := pssapi.GetPeerTopics(remotepubkeyhex)
	if err != nil {
		t.Fatal(err)
	}

OUTER:
	for _, tnew := range intopic {
		for torig, addr := range topicaddr {
			if bytes.Equal(torig[:], tnew[:]) {
				inaddr, err := pssapi.GetPeerAddress(remotepubkeyhex, torig)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(addr, inaddr) {
					t.Fatalf("Address mismatch for topic %x; got %x, expected %x", torig, inaddr, addr)
				}
				delete(topicaddr, torig)
				continue OUTER
			}
		}
		t.Fatalf("received topic %x did not match any existing topics", tnew)
	}

	if len(topicaddr) != 0 {
		t.Fatalf("%d topics were not matched", len(topicaddr))
	}
}

// forwarding should skip peers that do not have matching pss capabilities
func TestPeerCapabilityMismatch(t *testing.T) {

	// create privkey for forwarder node
	privkey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	// initialize kad
	baseaddr := network.RandomAddr()
	kad := network.NewKademlia((baseaddr).Over(), network.NewKadParams())
	rw := &p2p.MsgPipeRW{}

	// one peer has a mismatching version of pss
	wrongpssaddr := network.RandomAddr()
	wrongpsscap := p2p.Cap{
		Name:    protocolName,
		Version: 0,
	}
	nid := enode.ID{0x01}
	wrongpsspeer := network.NewPeer(&network.BzzPeer{
		Peer:    protocols.NewPeer(p2p.NewPeer(nid, common.ToHex(wrongpssaddr.Over()), []p2p.Cap{wrongpsscap}), rw, nil),
		BzzAddr: &network.BzzAddr{OAddr: wrongpssaddr.Over(), UAddr: nil},
	}, kad)

	// one peer doesn't even have pss (boo!)
	nopssaddr := network.RandomAddr()
	nopsscap := p2p.Cap{
		Name:    "nopss",
		Version: 1,
	}
	nid = enode.ID{0x02}
	nopsspeer := network.NewPeer(&network.BzzPeer{
		Peer:    protocols.NewPeer(p2p.NewPeer(nid, common.ToHex(nopssaddr.Over()), []p2p.Cap{nopsscap}), rw, nil),
		BzzAddr: &network.BzzAddr{OAddr: nopssaddr.Over(), UAddr: nil},
	}, kad)

	// add peers to kademlia and activate them
	// it's safe so don't check errors
	kad.Register(wrongpsspeer.BzzAddr)
	kad.On(wrongpsspeer)
	kad.Register(nopsspeer.BzzAddr)
	kad.On(nopsspeer)

	// create pss
	pssmsg := &PssMsg{
		To:      []byte{},
		Expire:  uint32(time.Now().Add(time.Second).Unix()),
		Payload: nil,
	}
	ps := newTestPss(privkey, kad, nil)
	defer ps.Stop()

	// run the forward
	// it is enough that it completes; trying to send to incapable peers would create segfault
	ps.forward(pssmsg)

}

// verifies that message handlers for raw messages only are invoked when minimum one handler for the topic exists in which raw messages are explicitly allowed
func TestRawAllow(t *testing.T) {

	// set up pss like so many times before
	privKey, err := ethCrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	baseAddr := network.RandomAddr()
	kad := network.NewKademlia((baseAddr).Over(), network.NewKadParams())
	ps := newTestPss(privKey, kad, nil)
	defer ps.Stop()
	topic := BytesToTopic([]byte{0x2a})

	// create handler innards that increments every time a message hits it
	var receives int
	rawHandlerFunc := func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		log.Trace("in allowraw handler")
		receives++
		return nil
	}

	// wrap this handler function with a handler without raw capability and register it
	hndlrNoRaw := &handler{
		f: rawHandlerFunc,
	}
	ps.Register(&topic, hndlrNoRaw)

	// test it with a raw message, should be poo-poo
	pssMsg := newPssMsg(&msgParams{
		raw: true,
	})
	pssMsg.To = baseAddr.OAddr
	pssMsg.Expire = uint32(time.Now().Unix() + 4200)
	pssMsg.Topic = topic
	pssMsg.Payload = nil
	ps.handle(context.TODO(), pssMsg)
	if receives > 0 {
		t.Fatalf("Expected handler not to be executed with raw cap off")
	}

	// now wrap the same handler function with raw capabilities and register it
	hndlrRaw := &handler{
		f: rawHandlerFunc,
		caps: &handlerCaps{
			raw: true,
		},
	}
	deregRawHandler := ps.Register(&topic, hndlrRaw)

	// should work now
	pssMsg.Payload = []byte("Raw Deal")
	ps.handle(context.TODO(), pssMsg)
	if receives == 0 {
		t.Fatalf("Expected handler to be executed with raw cap on")
	}

	// now deregister the raw capable handler
	prevReceives := receives
	deregRawHandler()

	// check that raw messages fail again
	pssMsg.Payload = []byte("Raw Trump")
	ps.handle(context.TODO(), pssMsg)
	if receives != prevReceives {
		t.Fatalf("Expected handler not to be executed when raw handler is retracted")
	}
}

// BELOW HERE ARE TESTS USING THE SIMULATION FRAMEWORK

// tests that the API layer can handle edge case values
func TestApi(t *testing.T) {
	clients, closeSimFunc, err := setupNetwork(2, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimFunc()

	topic := "0xdeadbeef"

	err = clients[0].Call(nil, "pss_sendRaw", "0x", topic, "0x666f6f")
	if err != nil {
		t.Fatal(err)
	}

	err = clients[0].Call(nil, "pss_sendRaw", "0xabcdef", topic, "0x")
	if err == nil {
		t.Fatal("expected error on empty msg")
	}

	overflowAddr := [33]byte{}
	err = clients[0].Call(nil, "pss_sendRaw", hexutil.Encode(overflowAddr[:]), topic, "0x666f6f")
	if err == nil {
		t.Fatal("expected error on send too big address")
	}
}

// verifies that nodes can send and receive raw (verbatim) messages
func TestSendRaw(t *testing.T) {
	t.Run("32", testSendRaw)
	t.Run("8", testSendRaw)
	t.Run("0", testSendRaw)
}

func testSendRaw(t *testing.T) {

	var addrsize int64
	var err error

	paramstring := strings.Split(t.Name(), "/")

	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("raw send test", "addrsize", addrsize)

	clients, closeSimFunc, err := setupNetwork(2, true)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimFunc()

	topic := "0xdeadbeef"

	var loaddrhex string
	err = clients[0].Call(&loaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddrhex = loaddrhex[:2+(addrsize*2)]
	var roaddrhex string
	err = clients[1].Call(&roaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddrhex = roaddrhex[:2+(addrsize*2)]

	time.Sleep(time.Millisecond * 500)

	// at this point we've verified that symkeys are saved and match on each peer
	// now try sending symmetrically encrypted message, both directions
	lmsgC := make(chan APIMsg)
	lctx, lcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer lcancel()
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", topic, true, false)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, rcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer rcancel()
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", topic, true, false)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// send and verify delivery
	lmsg := []byte("plugh")
	err = clients[1].Call(nil, "pss_sendRaw", loaddrhex, topic, hexutil.Encode(lmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-lmsgC:
		if !bytes.Equal(recvmsg.Msg, lmsg) {
			t.Fatalf("node 1 received payload mismatch: expected %v, got %v", lmsg, recvmsg)
		}
	case cerr := <-lctx.Done():
		t.Fatalf("test message (left) timed out: %v", cerr)
	}
	rmsg := []byte("xyzzy")
	err = clients[0].Call(nil, "pss_sendRaw", roaddrhex, topic, hexutil.Encode(rmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-rmsgC:
		if !bytes.Equal(recvmsg.Msg, rmsg) {
			t.Fatalf("node 2 received payload mismatch: expected %x, got %v", rmsg, recvmsg.Msg)
		}
	case cerr := <-rctx.Done():
		t.Fatalf("test message (right) timed out: %v", cerr)
	}
}

// send symmetrically encrypted message between two directly connected peers
func TestSendSym(t *testing.T) {
	t.Run("32", testSendSym)
	t.Run("8", testSendSym)
	t.Run("0", testSendSym)
}

func testSendSym(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("sym send test", "addrsize", addrsize)

	clients, closeSimFunc, err := setupNetwork(2, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimFunc()

	var topic string
	err = clients[0].Call(&topic, "pss_stringToTopic", "foo:42")
	if err != nil {
		t.Fatal(err)
	}

	var loaddrhex string
	err = clients[0].Call(&loaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddrhex = loaddrhex[:2+(addrsize*2)]
	var roaddrhex string
	err = clients[1].Call(&roaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddrhex = roaddrhex[:2+(addrsize*2)]

	// retrieve public key from pss instance
	// set this public key reciprocally
	var lpubkeyhex string
	err = clients[0].Call(&lpubkeyhex, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	var rpubkeyhex string
	err = clients[1].Call(&rpubkeyhex, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500)

	// at this point we've verified that symkeys are saved and match on each peer
	// now try sending symmetrically encrypted message, both directions
	lmsgC := make(chan APIMsg)
	lctx, lcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer lcancel()
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", topic, false, false)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, rcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer rcancel()
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", topic, false, false)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	lrecvkey := network.RandomAddr().Over()
	rrecvkey := network.RandomAddr().Over()

	var lkeyids [2]string
	var rkeyids [2]string

	// manually set reciprocal symkeys
	err = clients[0].Call(&lkeyids, "psstest_setSymKeys", rpubkeyhex, lrecvkey, rrecvkey, defaultSymKeySendLimit, topic, roaddrhex)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(&rkeyids, "psstest_setSymKeys", lpubkeyhex, rrecvkey, lrecvkey, defaultSymKeySendLimit, topic, loaddrhex)
	if err != nil {
		t.Fatal(err)
	}

	// send and verify delivery
	lmsg := []byte("plugh")
	err = clients[1].Call(nil, "pss_sendSym", rkeyids[1], topic, hexutil.Encode(lmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-lmsgC:
		if !bytes.Equal(recvmsg.Msg, lmsg) {
			t.Fatalf("node 1 received payload mismatch: expected %v, got %v", lmsg, recvmsg)
		}
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	rmsg := []byte("xyzzy")
	err = clients[0].Call(nil, "pss_sendSym", lkeyids[1], topic, hexutil.Encode(rmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-rmsgC:
		if !bytes.Equal(recvmsg.Msg, rmsg) {
			t.Fatalf("node 2 received payload mismatch: expected %x, got %v", rmsg, recvmsg.Msg)
		}
	case cerr := <-rctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
}

// send asymmetrically encrypted message between two directly connected peers
func TestSendAsym(t *testing.T) {
	t.Run("32", testSendAsym)
	t.Run("8", testSendAsym)
	t.Run("0", testSendAsym)
}

func testSendAsym(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("asym send test", "addrsize", addrsize)

	clients, closeSimFunc, err := setupNetwork(2, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimFunc()

	var topic string
	err = clients[0].Call(&topic, "pss_stringToTopic", "foo:42")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 250)

	var loaddrhex string
	err = clients[0].Call(&loaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddrhex = loaddrhex[:2+(addrsize*2)]
	var roaddrhex string
	err = clients[1].Call(&roaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddrhex = roaddrhex[:2+(addrsize*2)]

	// retrieve public key from pss instance
	// set this public key reciprocally
	var lpubkey string
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	var rpubkey string
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500) // replace with hive healthy code

	lmsgC := make(chan APIMsg)
	lctx, lcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer lcancel()
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", topic, false, false)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, rcancel := context.WithTimeout(context.Background(), time.Second*10)
	defer rcancel()
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", topic, false, false)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// store reciprocal public keys
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, topic, roaddrhex)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, topic, loaddrhex)
	if err != nil {
		t.Fatal(err)
	}

	// send and verify delivery
	rmsg := []byte("xyzzy")
	err = clients[0].Call(nil, "pss_sendAsym", rpubkey, topic, hexutil.Encode(rmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-rmsgC:
		if !bytes.Equal(recvmsg.Msg, rmsg) {
			t.Fatalf("node 2 received payload mismatch: expected %v, got %v", rmsg, recvmsg.Msg)
		}
	case cerr := <-rctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	lmsg := []byte("plugh")
	err = clients[1].Call(nil, "pss_sendAsym", lpubkey, topic, hexutil.Encode(lmsg))
	if err != nil {
		t.Fatal(err)
	}
	select {
	case recvmsg := <-lmsgC:
		if !bytes.Equal(recvmsg.Msg, lmsg) {
			t.Fatalf("node 1 received payload mismatch: expected %v, got %v", lmsg, recvmsg.Msg)
		}
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
}

type Job struct {
	Msg      []byte
	SendNode enode.ID
	RecvNode enode.ID
}

func worker(id int, jobs <-chan Job, rpcs map[enode.ID]*rpc.Client, pubkeys map[enode.ID]string, topic string) {
	for j := range jobs {
		rpcs[j.SendNode].Call(nil, "pss_sendAsym", pubkeys[j.RecvNode], topic, hexutil.Encode(j.Msg))
	}
}

func TestNetwork(t *testing.T) {
	t.Run("16/1000/4/sim", testNetwork)
}

// params in run name:
// nodes/recipientAddresses/addrbytes/adaptertype
// if adaptertype is exec uses execadapter, simadapter otherwise
func TestNetwork2000(t *testing.T) {
	if !*testutil.Longrunning {
		t.Skip("run with --longrunning flag to run extensive network tests")
	}
	t.Run("3/2000/4/sim", testNetwork)
	t.Run("4/2000/4/sim", testNetwork)
	t.Run("8/2000/4/sim", testNetwork)
	t.Run("16/2000/4/sim", testNetwork)
}

func TestNetwork5000(t *testing.T) {
	if !*testutil.Longrunning {
		t.Skip("run with --longrunning flag to run extensive network tests")
	}
	t.Run("3/5000/4/sim", testNetwork)
	t.Run("4/5000/4/sim", testNetwork)
	t.Run("8/5000/4/sim", testNetwork)
	t.Run("16/5000/4/sim", testNetwork)
}

func TestNetwork10000(t *testing.T) {
	if !*testutil.Longrunning {
		t.Skip("run with --longrunning flag to run extensive network tests")
	}
	t.Run("3/10000/4/sim", testNetwork)
	t.Run("4/10000/4/sim", testNetwork)
	t.Run("8/10000/4/sim", testNetwork)
}

func testNetwork(t *testing.T) {
	paramstring := strings.Split(t.Name(), "/")
	nodecount, _ := strconv.ParseInt(paramstring[1], 10, 0)
	msgcount, _ := strconv.ParseInt(paramstring[2], 10, 0)
	addrsize, _ := strconv.ParseInt(paramstring[3], 10, 0)
	adapter := paramstring[4]

	log.Info("network test", "nodecount", nodecount, "msgcount", msgcount, "addrhintsize", addrsize)

	nodes := make([]enode.ID, nodecount)
	bzzaddrs := make(map[enode.ID]string, nodecount)
	rpcs := make(map[enode.ID]*rpc.Client, nodecount)
	pubkeys := make(map[enode.ID]string, nodecount)

	sentmsgs := make([][]byte, msgcount)
	recvmsgs := make([]bool, msgcount)
	nodemsgcount := make(map[enode.ID]int, nodecount)

	trigger := make(chan enode.ID)

	var sim = &simulation.Simulation{}
	if adapter == "exec" {
		sim, _ = simulation.NewExec(newServices(false))
	} else if adapter == "tcp" || adapter == "sim" {
		sim = simulation.NewInProc(newServices(false))
	}
	defer sim.Close()

	net := sim.Net

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err := sim.UploadSnapshot(ctx, fmt.Sprintf("testdata/snapshot_%d.json", nodecount))
	if err != nil {
		//TODO: Fix p2p simulation framework to not crash when loading 32-nodes
		//t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	triggerChecks := func(trigger chan enode.ID, id enode.ID, rpcclient *rpc.Client, topic string) error {
		msgC := make(chan APIMsg)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		sub, err := rpcclient.Subscribe(ctx, "pss", msgC, "receive", topic, false, false)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			defer sub.Unsubscribe()
			for {
				select {
				case recvmsg := <-msgC:
					idx, _ := binary.Uvarint(recvmsg.Msg)
					if !recvmsgs[idx] {
						log.Debug("msg recv", "idx", idx, "id", id)
						recvmsgs[idx] = true
						trigger <- id
					}
				case <-sub.Err():
					return
				}
			}
		}()
		return nil
	}

	var topic string
	for i, nod := range net.GetNodes() {
		nodes[i] = nod.ID()
		rpcs[nodes[i]], err = nod.Client()
		if err != nil {
			t.Fatal(err)
		}
		if topic == "" {
			err = rpcs[nodes[i]].Call(&topic, "pss_stringToTopic", "foo:42")
			if err != nil {
				t.Fatal(err)
			}
		}
		var pubkey string
		err = rpcs[nodes[i]].Call(&pubkey, "pss_getPublicKey")
		if err != nil {
			t.Fatal(err)
		}
		pubkeys[nod.ID()] = pubkey
		var addrhex string
		err = rpcs[nodes[i]].Call(&addrhex, "pss_baseAddr")
		if err != nil {
			t.Fatal(err)
		}
		bzzaddrs[nodes[i]] = addrhex
		err = triggerChecks(trigger, nodes[i], rpcs[nodes[i]], topic)
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(1 * time.Second)

	// setup workers
	jobs := make(chan Job, 10)
	for w := 1; w <= 10; w++ {
		go worker(w, jobs, rpcs, pubkeys, topic)
	}

	time.Sleep(1 * time.Second)

	for i := 0; i < int(msgcount); i++ {
		sendnodeidx := rand.Intn(int(nodecount))
		recvnodeidx := rand.Intn(int(nodecount - 1))
		if recvnodeidx >= sendnodeidx {
			recvnodeidx++
		}
		nodemsgcount[nodes[recvnodeidx]]++
		sentmsgs[i] = make([]byte, 8)
		c := binary.PutUvarint(sentmsgs[i], uint64(i))
		if c == 0 {
			t.Fatal("0 byte message")
		}
		if err != nil {
			t.Fatal(err)
		}
		err = rpcs[nodes[sendnodeidx]].Call(nil, "pss_setPeerPublicKey", pubkeys[nodes[recvnodeidx]], topic, bzzaddrs[nodes[recvnodeidx]])
		if err != nil {
			t.Fatal(err)
		}

		jobs <- Job{
			Msg:      sentmsgs[i],
			SendNode: nodes[sendnodeidx],
			RecvNode: nodes[recvnodeidx],
		}
	}

	finalmsgcount := 0
outer:
	for i := 0; i < int(msgcount); i++ {
		select {
		case id := <-trigger:
			nodemsgcount[id]--
			finalmsgcount++
		case <-ctx.Done():
			log.Warn("timeout")
			break outer
		}
	}

	for i, msg := range recvmsgs {
		if !msg {
			log.Debug("missing message", "idx", i)
		}
	}
	t.Logf("%d of %d messages received", finalmsgcount, msgcount)

	if finalmsgcount != int(msgcount) {
		t.Fatalf("%d messages were not received", int(msgcount)-finalmsgcount)
	}

}

// check that in a network of a -> b -> c -> a
// a doesn't receive a sent message twice
func TestDeduplication(t *testing.T) {
	var err error

	clients, closeSimFunc, err := setupNetwork(3, false)
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimFunc()

	var addrsize = 32
	var loaddrhex string
	err = clients[0].Call(&loaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddrhex = loaddrhex[:2+(addrsize*2)]
	var roaddrhex string
	err = clients[1].Call(&roaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddrhex = roaddrhex[:2+(addrsize*2)]
	var xoaddrhex string
	err = clients[2].Call(&xoaddrhex, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 3 baseaddr fail: %v", err)
	}
	xoaddrhex = xoaddrhex[:2+(addrsize*2)]

	log.Info("peer", "l", loaddrhex, "r", roaddrhex, "x", xoaddrhex)

	var topic string
	err = clients[0].Call(&topic, "pss_stringToTopic", "foo:42")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 250)

	// retrieve public key from pss instance
	// set this public key reciprocally
	var rpubkey string
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get receivenode pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500) // replace with hive healthy code

	rmsgC := make(chan APIMsg)
	rctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", topic, false, false)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// store public key for recipient
	// zero-length address means forward to all
	// we have just two peers, they will be in proxbin, and will both receive
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, topic, "0x")
	if err != nil {
		t.Fatal(err)
	}

	// send and verify delivery
	rmsg := []byte("xyzzy")
	err = clients[0].Call(nil, "pss_sendAsym", rpubkey, topic, hexutil.Encode(rmsg))
	if err != nil {
		t.Fatal(err)
	}

	var receivedok bool
OUTER:
	for {
		select {
		case <-rmsgC:
			if receivedok {
				t.Fatalf("duplicate message received")
			}
			receivedok = true
		case <-rctx.Done():
			break OUTER
		}
	}
	if !receivedok {
		t.Fatalf("message did not arrive")
	}
}

// symmetric send performance with varying message sizes
func BenchmarkSymkeySend(b *testing.B) {
	b.Run(fmt.Sprintf("%d", 256), benchmarkSymKeySend)
	b.Run(fmt.Sprintf("%d", 1024), benchmarkSymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024), benchmarkSymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024*10), benchmarkSymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024*100), benchmarkSymKeySend)
}

func benchmarkSymKeySend(b *testing.B) {
	msgsizestring := strings.Split(b.Name(), "/")
	if len(msgsizestring) != 2 {
		b.Fatalf("benchmark called without msgsize param")
	}
	msgsize, err := strconv.ParseInt(msgsizestring[1], 10, 0)
	if err != nil {
		b.Fatalf("benchmark called with invalid msgsize param '%s': %v", msgsizestring[1], err)
	}
	privkey, err := ethCrypto.GenerateKey()
	ps := newTestPss(privkey, nil, nil)
	defer ps.Stop()
	msg := make([]byte, msgsize)
	rand.Read(msg)
	topic := BytesToTopic([]byte("foo"))
	to := make(PssAddress, 32)
	copy(to[:], network.RandomAddr().Over())
	symkeyid, err := ps.GenerateSymmetricKey(topic, to, true)
	if err != nil {
		b.Fatalf("could not generate symkey: %v", err)
	}
	symkey, err := ps.Crypto.GetSymmetricKey(symkeyid)
	if err != nil {
		b.Fatalf("could not retrieve symkey: %v", err)
	}
	ps.SetSymmetricKey(symkey, topic, to, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.SendSym(symkeyid, topic, msg)
	}
}

// asymmetric send performance with varying message sizes
func BenchmarkAsymkeySend(b *testing.B) {
	b.Run(fmt.Sprintf("%d", 256), benchmarkAsymKeySend)
	b.Run(fmt.Sprintf("%d", 1024), benchmarkAsymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024), benchmarkAsymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024*10), benchmarkAsymKeySend)
	b.Run(fmt.Sprintf("%d", 1024*1024*100), benchmarkAsymKeySend)
}

func benchmarkAsymKeySend(b *testing.B) {
	msgsizestring := strings.Split(b.Name(), "/")
	if len(msgsizestring) != 2 {
		b.Fatalf("benchmark called without msgsize param")
	}
	msgsize, err := strconv.ParseInt(msgsizestring[1], 10, 0)
	if err != nil {
		b.Fatalf("benchmark called with invalid msgsize param '%s': %v", msgsizestring[1], err)
	}
	privkey, err := ethCrypto.GenerateKey()
	ps := newTestPss(privkey, nil, nil)
	defer ps.Stop()
	msg := make([]byte, msgsize)
	rand.Read(msg)
	topic := BytesToTopic([]byte("foo"))
	to := make(PssAddress, 32)
	copy(to[:], network.RandomAddr().Over())
	ps.SetPeerPublicKey(&privkey.PublicKey, topic, to)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.SendAsym(common.ToHex(ps.Crypto.SerializePublicKey(&privkey.PublicKey)), topic, msg)
	}
}
func BenchmarkSymkeyBruteforceChangeaddr(b *testing.B) {
	b.Skip("Test doesn't work. Test messages are not valid, they need Control field")
	for i := 100; i < 100000; i = i * 10 {
		for j := 32; j < 10000; j = j * 8 {
			b.Run(fmt.Sprintf("%d/%d", i, j), benchmarkSymkeyBruteforceChangeaddr)
		}
		//b.Run(fmt.Sprintf("%d", i), benchmarkSymkeyBruteforceChangeaddr)
	}
}

// decrypt performance using symkey cache, worst case
// (decrypt key always last in cache)
func benchmarkSymkeyBruteforceChangeaddr(b *testing.B) {
	keycountstring := strings.Split(b.Name(), "/")
	cachesize := int64(0)
	var ps *Pss
	if len(keycountstring) < 2 {
		b.Fatalf("benchmark called without count param")
	}
	keycount, err := strconv.ParseInt(keycountstring[1], 10, 0)
	if err != nil {
		b.Fatalf("benchmark called with invalid count param '%s': %v", keycountstring[1], err)
	}
	if len(keycountstring) == 3 {
		cachesize, err = strconv.ParseInt(keycountstring[2], 10, 0)
		if err != nil {
			b.Fatalf("benchmark called with invalid cachesize '%s': %v", keycountstring[2], err)
		}
	}
	pssmsgs := make([]*PssMsg, 0, keycount)
	var keyid string
	privkey, err := ethCrypto.GenerateKey()
	if cachesize > 0 {
		ps = newTestPss(privkey, nil, &Params{SymKeyCacheCapacity: int(cachesize)})
	} else {
		ps = newTestPss(privkey, nil, nil)
	}
	defer ps.Stop()
	topic := BytesToTopic([]byte("foo"))
	for i := 0; i < int(keycount); i++ {
		to := make(PssAddress, 32)
		copy(to[:], network.RandomAddr().Over())
		keyid, err = ps.GenerateSymmetricKey(topic, to, true)
		if err != nil {
			b.Fatalf("cant generate symkey #%d: %v", i, err)
		}
		symkey, err := ps.Crypto.GetSymmetricKey(keyid)
		if err != nil {
			b.Fatalf("could not retrieve symkey %s: %v", keyid, err)
		}
		wparams := &crypto.WrapParams{
			SymmetricKey: symkey,
		}
		payload, err := ps.Crypto.Wrap([]byte("xyzzy"), wparams)
		if err != nil {
			b.Fatalf("could not generate envelope: %v", err)
		}
		ps.Register(&topic, &handler{
			f: noopHandlerFunc,
		})
		pssmsgs = append(pssmsgs, &PssMsg{
			To:      to,
			Topic:   topic,
			Payload: payload,
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ps.process(pssmsgs[len(pssmsgs)-(i%len(pssmsgs))-1], false, false); err != nil {
			b.Fatalf("pss processing failed: %v", err)
		}
	}
}

func BenchmarkSymkeyBruteforceSameaddr(b *testing.B) {
	b.Skip("Test doesn't work. Test messages are not valid, they need Control field")
	for i := 100; i < 100000; i = i * 10 {
		for j := 32; j < 10000; j = j * 8 {
			b.Run(fmt.Sprintf("%d/%d", i, j), benchmarkSymkeyBruteforceSameaddr)
		}
	}
}

// decrypt performance using symkey cache, best case
// (decrypt key always first in cache)
func benchmarkSymkeyBruteforceSameaddr(b *testing.B) {
	var keyid string
	var ps *Pss
	cachesize := int64(0)
	keycountstring := strings.Split(b.Name(), "/")
	if len(keycountstring) < 2 {
		b.Fatalf("benchmark called without count param")
	}
	keycount, err := strconv.ParseInt(keycountstring[1], 10, 0)
	if err != nil {
		b.Fatalf("benchmark called with invalid count param '%s': %v", keycountstring[1], err)
	}
	if len(keycountstring) == 3 {
		cachesize, err = strconv.ParseInt(keycountstring[2], 10, 0)
		if err != nil {
			b.Fatalf("benchmark called with invalid cachesize '%s': %v", keycountstring[2], err)
		}
	}
	addr := make([]PssAddress, keycount)
	privkey, err := ethCrypto.GenerateKey()
	if cachesize > 0 {
		ps = newTestPss(privkey, nil, &Params{SymKeyCacheCapacity: int(cachesize)})
	} else {
		ps = newTestPss(privkey, nil, nil)
	}
	defer ps.Stop()
	topic := BytesToTopic([]byte("foo"))
	for i := 0; i < int(keycount); i++ {
		copy(addr[i], network.RandomAddr().Over())
		keyid, err = ps.GenerateSymmetricKey(topic, addr[i], true)
		if err != nil {
			b.Fatalf("cant generate symkey #%d: %v", i, err)
		}

	}
	symkey, err := ps.Crypto.GetSymmetricKey(keyid)
	if err != nil {
		b.Fatalf("could not retrieve symkey %s: %v", keyid, err)
	}
	wparams := &crypto.WrapParams{
		SymmetricKey: symkey,
	}
	payload, err := ps.Crypto.Wrap([]byte("xyzzy"), wparams)
	if err != nil {
		b.Fatalf("could not generate envelope: %v", err)
	}
	ps.Register(&topic, &handler{
		f: noopHandlerFunc,
	})
	pssmsg := &PssMsg{
		To:      addr[len(addr)-1][:],
		Topic:   topic,
		Payload: payload,
	}
	for i := 0; i < b.N; i++ {
		if err := ps.process(pssmsg, false, false); err != nil {
			b.Fatalf("pss processing failed: %v", err)
		}
	}
}

func testRandomMessage() *PssMsg {
	addr := make([]byte, 32)
	addr[0] = 0x01
	msg := newPssMsg(&msgParams{})
	msg.To = addr
	msg.Expire = uint32(time.Now().Add(time.Second * 60).Unix())
	msg.Topic = [4]byte{}
	msg.Payload = []byte{0x66, 0x6f, 0x6f}
	return msg
}

// setup simulated network with bzz/discovery and pss services.
// connects nodes in a circle
// if allowRaw is set, omission of builtin pss encryption is enabled (see PssParams)
func setupNetwork(numnodes int, allowRaw bool) (clients []*rpc.Client, closeSimFunc func(), err error) {
	clients = make([]*rpc.Client, numnodes)
	if numnodes < 2 {
		return nil, nil, fmt.Errorf("minimum two nodes in network")
	}
	sim := simulation.NewInProc(newServices(allowRaw))
	closeSimFunc = sim.Close
	if numnodes == 2 {
		_, err = sim.AddNodesAndConnectChain(numnodes)

	} else {
		_, err = sim.AddNodesAndConnectRing(numnodes)
	}
	if err != nil {
		return nil, nil, err
	}
	nodes := sim.Net.GetNodes()
	for id, node := range nodes {
		client, err := node.Client()
		if err != nil {
			return nil, nil, fmt.Errorf("error getting the nodes clients")
		}
		clients[id] = client
	}
	return clients, closeSimFunc, nil
}

func newServices(allowRaw bool) map[string]simulation.ServiceFunc {
	stateStore := state.NewInmemoryStore()
	kademlias := make(map[enode.ID]*network.Kademlia)
	kademlia := func(id enode.ID, bzzKey []byte) *network.Kademlia {
		if k, ok := kademlias[id]; ok {
			return k
		}
		params := network.NewKadParams()
		params.NeighbourhoodSize = 2
		params.MaxBinSize = 3
		params.MinBinSize = 1
		params.MaxRetries = 1000
		params.RetryExponent = 2
		params.RetryInterval = 1000000
		kademlias[id] = network.NewKademlia(bzzKey, params)
		return kademlias[id]
	}
	return map[string]simulation.ServiceFunc{
		"bzz": func(ctx *adapters.ServiceContext, bucket *sync.Map) (s node.Service, cleanup func(), err error) {
			addr := network.NewAddr(ctx.Config.Node())
			bzzPrivateKey, err := simulation.BzzPrivateKeyFromConfig(ctx.Config)
			if err != nil {
				return nil, nil, err
			}
			addr.OAddr = network.PrivateKeyToBzzKey(bzzPrivateKey)
			bucket.Store(simulation.BucketKeyBzzPrivateKey, bzzPrivateKey)
			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			pskad := kademlia(ctx.Config.ID, addr.OAddr)
			bucket.Store(simulation.BucketKeyKademlia, pskad)
			return network.NewBzz(config, pskad, stateStore, nil, nil, nil, nil), nil, nil
		},
		protocolName: func(ctx *adapters.ServiceContext, bucket *sync.Map) (node.Service, func(), error) {
			// execadapter does not exec init()
			initTest()

			privkey, err := ethCrypto.GenerateKey()
			pssp := NewParams().WithPrivateKey(privkey)
			pssp.AllowRaw = allowRaw
			bzzPrivateKey, err := simulation.BzzPrivateKeyFromConfig(ctx.Config)
			if err != nil {
				return nil, nil, err
			}
			bzzKey := network.PrivateKeyToBzzKey(bzzPrivateKey)
			pskad := kademlia(ctx.Config.ID, bzzKey)
			bucket.Store(simulation.BucketKeyKademlia, pskad)
			ps, err := New(pskad, pssp)
			if err != nil {
				return nil, nil, err
			}
			ping := &Ping{
				OutC: make(chan bool),
				Pong: true,
			}
			p2pp := NewPingProtocol(ping)
			pp, err := RegisterProtocol(ps, &PingTopic, PingProtocol, p2pp, &ProtocolParams{Asymmetric: true})
			if err != nil {
				return nil, nil, err
			}
			if useHandshake {
				SetHandshakeController(ps, NewHandshakeParams())
			}
			cleanupFunc := ps.Register(&PingTopic, &handler{
				f: pp.Handle,
				caps: &handlerCaps{
					raw: true,
				},
			})
			ps.addAPI(rpc.API{
				Namespace: "psstest",
				Version:   "0.3",
				Service:   NewAPITest(ps),
				Public:    false,
			})
			pssprotocols[ctx.Config.ID.String()] = &protoCtrl{
				C:        ping.OutC,
				protocol: pp,
				run:      p2pp.Run,
			}

			return ps, cleanupFunc, nil
		},
	}
}

// New Test Pss that will be started
func newTestPss(privkey *ecdsa.PrivateKey, kad *network.Kademlia, ppextra *Params) *Pss {
	return newTestPssStart(privkey, kad, ppextra, true)
}

// New Test Pss but with a parameter to select if the pss process should start
func newTestPssStart(privkey *ecdsa.PrivateKey, kad *network.Kademlia, ppextra *Params, start bool) *Pss {
	nid := enode.PubkeyToIDV4(&privkey.PublicKey)
	// set up routing if kademlia is not passed to us
	if kad == nil {
		kp := network.NewKadParams()
		kp.NeighbourhoodSize = 3
		kad = network.NewKademlia(nid[:], kp)
	}

	// create pss
	pp := NewParams().WithPrivateKey(privkey)
	if ppextra != nil {
		pp.SymKeyCacheCapacity = ppextra.SymKeyCacheCapacity
	}
	ps, err := New(kad, pp)
	if err != nil {
		return nil
	}
	if start {
		err = ps.Start(nil)
		if err != nil {
			return nil
		}
	}

	return ps
}

// API calls for test/development use
type APITest struct {
	*Pss
}

func NewAPITest(ps *Pss) *APITest {
	return &APITest{Pss: ps}
}

func (apitest *APITest) SetSymKeys(pubkeyid string, recvsymkey []byte, sendsymkey []byte, limit uint16, topic Topic, to hexutil.Bytes) ([2]string, error) {

	recvsymkeyid, err := apitest.SetSymmetricKey(recvsymkey, topic, PssAddress(to), true)
	if err != nil {
		return [2]string{}, err
	}
	sendsymkeyid, err := apitest.SetSymmetricKey(sendsymkey, topic, PssAddress(to), false)
	if err != nil {
		return [2]string{}, err
	}
	return [2]string{recvsymkeyid, sendsymkeyid}, nil
}

func (apitest *APITest) Clean() (int, error) {
	return apitest.Pss.cleanKeys(), nil
}
