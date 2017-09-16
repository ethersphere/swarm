package pss

import (
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/simulations"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

const (
	pssServiceName = "pss"
	bzzServiceName = "bzz"
)

type protoCtrl struct {
	C        chan bool
	protocol *PssProtocol
	run      func(*p2p.Peer, p2p.MsgReadWriter) error
}

var (
	snapshotfile   string
	debugdebugflag = flag.Bool("vv", false, "veryverbose")
	debugflag      = flag.Bool("v", false, "verbose")
	w              *whisper.Whisper
	wapi           *whisper.PublicWhisperAPI
	// custom logging
	psslogmain   log.Logger
	pssprotocols map[string]*protoCtrl
)

var services = newServices()

func init() {

	flag.Parse()
	rand.Seed(time.Now().Unix())

	adapters.RegisterServices(services)

	loglevel := log.LvlInfo
	if *debugflag {
		loglevel = log.LvlDebug
	} else if *debugdebugflag {
		loglevel = log.LvlTrace
	}

	psslogmain = log.New("psslog", "*")
	hs := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	hf := log.LvlFilterHandler(loglevel, hs)
	h := log.CallerFileHandler(hf)
	log.Root().SetHandler(h)

	w = whisper.New()
	wapi = whisper.NewPublicWhisperAPI(w)

	pssprotocols = make(map[string]*protoCtrl)
}

// test if we can insert into cache, match items with cache and cache expiry
func TestCache(t *testing.T) {
	var err error
	to, _ := hex.DecodeString("08090a0b0c0d0e0f1011121314150001020304050607161718191a1b1c1d1e1f")
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)

	ps := NewTestPss(privkey, nil)
	pp := NewPssParams(privkey)
	data := []byte("foo")
	datatwo := []byte("bar")
	wparams := &whisper.MessageParams{
		TTL:      DefaultTTL,
		Src:      privkey,
		Dst:      &privkey.PublicKey,
		Topic:    PingTopic,
		WorkTime: defaultWhisperWorkTime,
		PoW:      defaultWhisperPoW,
		Payload:  data,
	}
	woutmsg, err := whisper.NewSentMessage(wparams)
	env, err := woutmsg.Wrap(wparams)
	msg := &PssMsg{
		Payload: env,
		To:      to,
	}
	wparams.Payload = datatwo
	woutmsg, err = whisper.NewSentMessage(wparams)
	envtwo, err := woutmsg.Wrap(wparams)
	msgtwo := &PssMsg{
		Payload: envtwo,
		To:      to,
	}

	digest, err := ps.storeMsg(msg)
	if err != nil {
		t.Fatalf("could not store cache msgone: %v", err)
	}
	digesttwo, err := ps.storeMsg(msgtwo)
	if err != nil {
		t.Fatalf("could not store cache msgtwo: %v", err)
	}

	if digest == digesttwo {
		t.Fatalf("different msgs return same hash: %d", digesttwo)
	}

	// check the cache
	err = ps.addFwdCache(digest)
	if err != nil {
		t.Fatalf("write to pss expire cache failed: %v", err)
	}

	if !ps.checkFwdCache(nil, digest) {
		t.Fatalf("message %v should have EXPIRE record in cache but checkCache returned false", msg)
	}

	if ps.checkFwdCache(nil, digesttwo) {
		t.Fatalf("message %v should NOT have EXPIRE record in cache but checkCache returned true", msgtwo)
	}

	time.Sleep(pp.CacheTTL)
	if ps.checkFwdCache(nil, digest) {
		t.Fatalf("message %v should have expired from cache but checkCache returned true", msg)
	}
}

// matching of address hints; whether a message could be or is for the node
func TestAddressMatch(t *testing.T) {

	localaddr := network.RandomAddr().Over()
	copy(localaddr[:8], []byte("deadbeef"))
	remoteaddr := []byte("feedbeef")
	kadparams := network.NewKadParams()
	kad := network.NewKademlia(localaddr, kadparams)
	keys, err := wapi.NewKeyPair()
	if err != nil {
		t.Fatalf("Could not generate private key: %v", err)
	}
	privkey, err := w.GetPrivateKey(keys)
	pssp := NewPssParams(privkey)
	ps := NewPss(kad, nil, pssp)

	pssmsg := &PssMsg{
		To:      remoteaddr,
		Payload: &whisper.Envelope{},
	}

	// differ from first byte
	if ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient true but %x != %x", remoteaddr, localaddr)
	}
	if ps.isSelfPossibleRecipient(pssmsg) {
		t.Fatalf("isSelfPossibleRecipient true but %x != %x", remoteaddr[:8], localaddr[:8])
	}

	// 8 first bytes same
	copy(remoteaddr[:4], localaddr[:4])
	if ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient true but %x != %x", remoteaddr, localaddr)
	}
	if !ps.isSelfPossibleRecipient(pssmsg) {
		t.Fatalf("isSelfPossibleRecipient false but %x == %x", remoteaddr[:8], localaddr[:8])
	}

	// all bytes same
	pssmsg.To = localaddr
	if !ps.isSelfRecipient(pssmsg) {
		t.Fatalf("isSelfRecipient false but %x == %x", remoteaddr, localaddr)
	}
	if !ps.isSelfPossibleRecipient(pssmsg) {
		t.Fatalf("isSelfPossibleRecipient false but %x == %x", remoteaddr[:8], localaddr[:8])
	}
}

// set and generate pubkeys and symkeys
func TestKeys(t *testing.T) {
	// make our key and init pss with it
	ourkeys, err := wapi.NewKeyPair()
	if err != nil {
		t.Fatalf("create 'our' key fail")
	}
	theirkeys, err := wapi.NewKeyPair()
	if err != nil {
		t.Fatalf("create 'their' key fail")
	}
	ourprivkey, err := w.GetPrivateKey(ourkeys)
	if err != nil {
		t.Fatalf("failed to retrieve 'our' private key")
	}
	theirprivkey, err := w.GetPrivateKey(theirkeys)
	if err != nil {
		t.Fatalf("failed to retrieve 'their' private key")
	}
	ps := NewTestPss(ourprivkey, nil)

	// set up peer with mock address, mapped to mocked publicaddress and with mocked symkey
	addr := network.RandomAddr().Over()
	outkey := network.RandomAddr().Over()
	topic := whisper.BytesToTopic([]byte("foo:42"))
	ps.SetPeerPublicKey(&theirprivkey.PublicKey, topic, addr)
	outkeyid, err := ps.SetSymmetricKey(outkey, topic, addr, 0, false)
	if err != nil {
		t.Fatalf("failed to set 'our' outgoing symmetric key")
	}

	// make a symmetric key that we will send to peer for encrypting messages to us
	inkeyid, err := ps.generateSymmetricKey(topic, addr, defaultSymKeySendLimit, true)
	if err != nil {
		t.Fatalf("failed to set 'our' incoming symmetric key")
	}

	// get the key back from whisper, check that it's still the same
	outkeyback, err := ps.w.GetSymKey(outkeyid)
	if err != nil {
		t.Fatalf(err.Error())
	}
	inkey, err := ps.w.GetSymKey(inkeyid)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !bytes.Equal(outkeyback, outkey) {
		t.Fatalf("passed outgoing symkey doesnt equal stored: %x / %x", outkey, outkeyback)
	}

	t.Logf("symout: %v", outkeyback)
	t.Logf("symin: %v", inkey)

	// check that the key is stored in the peerpool
	//psp := ps.symKeyPool[inkeyid][topic]
}

// asymmetrical key exchange between two directly connected peers
// full address, partial address (8 bytes) and empty address
func TestHandshake(t *testing.T) {
	t.Run("32", testHandshake)
	t.Run("8", testHandshake)
	t.Run("0", testHandshake)
}

func testHandshake(t *testing.T) {

	// how much of the address we will use
	var addrsize int64
	var err error
	addrsizestring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(addrsizestring[1], 10, 0)

	// set up two nodes directly connected
	// (we are not testing pss routing here)
	topic := whisper.BytesToTopic([]byte("foo:42"))
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	var loaddr []byte
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddr = loaddr[:addrsize]
	var roaddr []byte
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddr = roaddr[:addrsize]
	log.Debug("addresses", "left", loaddr, "right", roaddr)

	// retrieve public key from pss instance
	// set this public key reciprocally
	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 1000) // replace with hive healthy code

	// give each node its peer's public key
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, roaddr)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, loaddr)
	if err != nil {
		t.Fatal(err)
	}

	// perform the handshake
	// after this each side will have defaultSymKeyBufferCapacity symkeys each for in- and outgoing messages:
	// L -> request 6 keys -> R
	// L <- send 6 keys, request 6 keys <- R
	// L -> send 6 keys -> R
	// the call will fill the array with symkeys L needs for sending to R
	var hsendsymkeyids []string
	err = clients[0].Call(&hsendsymkeyids, "pss_handshake", common.ToHex(rpubkey), hextopic, roaddr, true)
	if err != nil {
		t.Fatal(err)
	}

	// make sure the r-node gets its keys
	time.Sleep(time.Second)

	// check if we have 6 outgoing keys stored, and they match what was received from R
	var lsendsymkeyids []string
	err = clients[0].Call(&lsendsymkeyids, "pss_getSymmetricKeys", common.ToHex(rpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}
	m := 0
	for _, hid := range hsendsymkeyids {
		for _, lid := range lsendsymkeyids {
			if lid == hid {
				m++
			}
		}
	}
	if m != defaultSymKeyBufferCapacity {
		t.Fatalf("buffer size mismatch, expected %d, have %d: %v", defaultSymKeyBufferCapacity, m, lsendsymkeyids)
	}

	// check if in- and outgoing keys on l-node and r-node match up and are in opposite categories (l recv = r send, l send = r recv)
	var rsendsymkeyids []string
	err = clients[1].Call(&rsendsymkeyids, "pss_getSymmetricKeys", common.ToHex(lpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}

	// get outgoing symkeys in byte form from both sides
	var lsendsymkeys [][]byte
	for _, id := range lsendsymkeyids {
		var key []byte
		err = clients[0].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		lsendsymkeys = append(lsendsymkeys, key)
	}
	var rsendsymkeys [][]byte
	for _, id := range rsendsymkeyids {
		var key []byte
		err = clients[1].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		rsendsymkeys = append(rsendsymkeys, key)
	}

	// get ALL symkeys in byte form from both sides
	var lsymkeyids []string
	err = clients[0].Call(&lsymkeyids, "psstest_dumpSymKeys", common.ToHex(rpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}
	var lsymkeys [][]byte
	for _, id := range lsymkeyids {
		var key []byte
		err = clients[0].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		lsymkeys = append(lsymkeys, key)
	}
	var rsymkeyids []string
	err = clients[1].Call(&rsymkeyids, "psstest_dumpSymKeys", common.ToHex(lpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}
	var rsymkeys [][]byte
	for _, id := range rsymkeyids {
		var key []byte
		err = clients[1].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		rsymkeys = append(rsymkeys, key)
	}

	// filter out the recv keys on both sides
	var lrecvsymkeys [][]byte
louter:
	for _, key := range lsymkeys {
		for _, sendkey := range lsendsymkeys {
			if bytes.Equal(key, sendkey) {
				continue louter
			}
		}
		lrecvsymkeys = append(lrecvsymkeys, key)
	}
	var rrecvsymkeys [][]byte
router:
	for _, key := range rsymkeys {
		for _, sendkey := range rsendsymkeys {
			if bytes.Equal(key, sendkey) {
				continue router
			}
		}
		rrecvsymkeys = append(rrecvsymkeys, key)
	}
	m = 0
	for _, lkey := range lrecvsymkeys {
		for _, rkey := range rsendsymkeys {
			if bytes.Equal(lkey, rkey) {
				m++
			}
		}
	}
	if m != defaultSymKeyBufferCapacity {
		t.Fatalf("left recv buffer does not match right send buffer, expected %d, have %d", defaultSymKeyBufferCapacity, m)
	}

	m = 0
	for _, lkey := range lsendsymkeys {
		for _, rkey := range rrecvsymkeys {
			if bytes.Equal(lkey, rkey) {
				m++
			}
		}
	}
	if m != defaultSymKeyBufferCapacity {
		t.Fatalf("left send buffer does not match right recv buffer, expected %d, have %d", defaultSymKeyBufferCapacity, m)
	}

	// verify can't initiate handshake when buffer is full
	var moresymkeys []string
	err = clients[0].Call(&moresymkeys, "pss_handshake", common.ToHex(rpubkey), hextopic, roaddr, true)
	if err == nil {
		t.Fatal("expected handshake to fail due to full buffer but no error returned")
	}

	// expire two keys from left, reinitiate handshake
	// afterwards two of the keys should be different than the ones we stored before
	for i := 0; i < 2; i++ {
		err = clients[0].Call(nil, "psstest_depleteSymKey", lsendsymkeyids[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	err = clients[0].Call(&moresymkeys, "pss_handshake", common.ToHex(rpubkey), hextopic, roaddr, false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	var lnewsendsymkeyids []string
	err = clients[0].Call(&lnewsendsymkeyids, "pss_getSymmetricKeys", common.ToHex(rpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}
	m = 0
	for _, id := range lnewsendsymkeyids {
		var key []byte
		err = clients[0].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		for _, oldkey := range lsendsymkeys {
			if bytes.Equal(oldkey, key) {
				m++
			}
		}
	}
	if m != defaultSymKeyBufferCapacity-2 {
		t.Fatalf("Left buffer mismatch after re-handshake, expected %d different keys in batch, have %d", defaultSymKeyBufferCapacity-2, m)
	}

	// expire two keys from both sides, reinitiate handshake
	// same difference as before, but now on both sides
	for i := 0; i < 2; i++ {
		err = clients[0].Call(nil, "psstest_depleteSymKey", lnewsendsymkeyids[i])
		if err != nil {
			t.Fatal(err)
		}
		err = clients[1].Call(nil, "psstest_depleteSymKey", rsendsymkeyids[i])
		if err != nil {
			t.Fatal(err)
		}
	}
	err = clients[0].Call(&moresymkeys, "pss_handshake", common.ToHex(rpubkey), hextopic, roaddr, false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	var rnewsendsymkeyids []string
	err = clients[1].Call(&rnewsendsymkeyids, "pss_getSymmetricKeys", common.ToHex(lpubkey), hextopic)
	if err != nil {
		t.Fatal(err)
	}
	m = 0
	for _, id := range rnewsendsymkeyids {
		var key []byte
		err = clients[1].Call(&key, "pss_getSymmetricKey", id)
		if err != nil {
			t.Fatal(err)
		}
		for _, oldkey := range rsendsymkeys {
			if bytes.Equal(oldkey, key) {
				m++
			}
		}
	}
	if m != defaultSymKeyBufferCapacity-2 {
		t.Fatalf("Right buffer mismatch after re-handshake, expected %d different keys in batch, have %d", defaultSymKeyBufferCapacity-2, m)
	}

	// expire one symkey, and attempt to send with it. should fail
	err = clients[1].Call(nil, "psstest_depleteSymKey", rnewsendsymkeyids[0])
	if err != nil {
		t.Fatal(err)
	}

	err = clients[1].Call(nil, "pss_sendSym", rnewsendsymkeyids[0], hextopic, []byte("foo"))
	if err == nil {
		t.Fatal("expected sym send fail on depleted key, but got no error")
	}
}

// send symmetrically encrypted message between two directly connected peers
func TestSymSend(t *testing.T) {
	t.Run("32", testSymSend)
	t.Run("8", testSymSend)
	t.Run("0", testSymSend)
}

func testSymSend(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("sym send test", "addrsize", addrsize)

	topic := whisper.BytesToTopic([]byte("foo:42"))
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	loaddr := make([]byte, addrsize)
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddr = loaddr[:addrsize]
	roaddr := make([]byte, addrsize)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddr = roaddr[:addrsize]

	// retrieve public key from pss instance
	// set this public key reciprocally
	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500)

	// at this point we've verified that symkeys are saved and match on each peer
	// now try sending symmetrically encrypted message, both directions
	lmsgC := make(chan APIMsg)
	lctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", hextopic)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", hextopic)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	lrecvkey := network.RandomAddr().Over()
	rrecvkey := network.RandomAddr().Over()

	var lkeyids [2]string
	var rkeyids [2]string

	// manually set reciprocal symkeys
	err = clients[0].Call(&lkeyids, "psstest_setSymKeys", common.ToHex(rpubkey), lrecvkey, rrecvkey, defaultSymKeySendLimit, hextopic, roaddr)
	if err != nil {
		t.Fatal(err)
	}

	err = clients[1].Call(&rkeyids, "psstest_setSymKeys", common.ToHex(lpubkey), rrecvkey, lrecvkey, defaultSymKeySendLimit, hextopic, loaddr)
	if err != nil {
		t.Fatal(err)
	}

	// send and verify delivery
	lmsg := []byte("plugh")
	err = clients[1].Call(nil, "pss_sendSym", rkeyids[1], hextopic, lmsg)
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
	err = clients[0].Call(nil, "pss_sendSym", lkeyids[1], hextopic, rmsg)
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
}

// send asymmetrically encrypted message between two directly connected peers
func TestAsymSend(t *testing.T) {
	t.Run("32", testAsymSend)
	t.Run("8", testAsymSend)
	t.Run("0", testAsymSend)
}

func testAsymSend(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("asym send test", "addrsize", addrsize)

	topic := whisper.BytesToTopic([]byte("foo:42"))
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 250)

	loaddr := make([]byte, 32)
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddr = loaddr[:addrsize]
	roaddr := make([]byte, 32)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddr = roaddr[:addrsize]

	// retrieve public key from pss instance
	// set this public key reciprocally
	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500) // replace with hive healthy code

	var addrs [2][]byte

	lmsgC := make(chan APIMsg)
	lctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", hextopic)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", hextopic)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// store reciprocal public keys
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, addrs[1])
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, addrs[0])
	if err != nil {
		t.Fatal(err)
	}

	// send and verify delivery
	rmsg := []byte("xyzzy")
	err = clients[0].Call(nil, "pss_sendAsym", common.ToHex(rpubkey), hextopic, rmsg)
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
	err = clients[1].Call(nil, "pss_sendAsym", common.ToHex(lpubkey), hextopic, lmsg)
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

// simple ping pong protocol test for the pss devp2p emulation
func TestProtocol(t *testing.T) {
	t.Run("32", testProtocol)
	t.Run("8", testProtocol)
	t.Run("0", testProtocol)
}

func testProtocol(t *testing.T) {

	// address hint size
	var addrsize int64
	var err error
	paramstring := strings.Split(t.Name(), "/")
	addrsize, _ = strconv.ParseInt(paramstring[1], 10, 0)
	log.Info("protocol test", "addrsize", addrsize)

	hextopic := fmt.Sprintf("%x", PingTopic)
	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	var loaddr []byte
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	loaddr = loaddr[:addrsize]
	var roaddr []byte
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}
	roaddr = roaddr[:addrsize]
	lnodeinfo := &p2p.NodeInfo{}
	err = clients[0].Call(&lnodeinfo, "admin_nodeInfo")
	if err != nil {
		t.Fatalf("rpc nodeinfo node 11 fail: %v", err)
	}

	lpubkey := make([]byte, 32)
	err = clients[0].Call(&lpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 1 pubkey fail: %v", err)
	}
	rpubkey := make([]byte, 32)
	err = clients[1].Call(&rpubkey, "pss_getPublicKey")
	if err != nil {
		t.Fatalf("rpc get node 2 pubkey fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500) // replace with hive healthy code

	lmsgC := make(chan APIMsg)
	lctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	lsub, err := clients[0].Subscribe(lctx, "pss", lmsgC, "receive", hextopic)
	log.Trace("lsub", "id", lsub)
	defer lsub.Unsubscribe()
	rmsgC := make(chan APIMsg)
	rctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	rsub, err := clients[1].Subscribe(rctx, "pss", rmsgC, "receive", hextopic)
	log.Trace("rsub", "id", rsub)
	defer rsub.Unsubscribe()

	// set reciprocal public keys
	err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, roaddr)
	if err != nil {
		t.Fatal(err)
	}
	err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, loaddr)
	if err != nil {
		t.Fatal(err)
	}

	// add right peer's public key as protocol peer on left
	nid, _ := discover.HexID("0x00") // this hack is needed to satisfy the p2p method
	p := p2p.NewPeer(nid, fmt.Sprintf("%x", loaddr), []p2p.Cap{})
	pssprotocols[lnodeinfo.ID].protocol.AddPeer(p, pssprotocols[lnodeinfo.ID].run, PingTopic, true, common.ToHex(rpubkey))

	// sends ping asym, checks delivery
	pssprotocols[lnodeinfo.ID].C <- false
	select {
	case <-lmsgC:
		log.Debug("lnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	select {
	case <-rmsgC:
		log.Debug("rnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}

	// sends ping asym, checks delivery
	pssprotocols[lnodeinfo.ID].C <- false
	select {
	case <-lmsgC:
		log.Debug("lnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
	}
	select {
	case <-rmsgC:
		log.Debug("rnode ok")
	case cerr := <-lctx.Done():
		t.Fatalf("test message timed out: %v", cerr)
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
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)
	ps := NewTestPss(privkey, nil)
	msg := make([]byte, msgsize)
	rand.Read(msg)
	topic := whisper.BytesToTopic([]byte("foo"))
	to := network.RandomAddr().Over()
	symkeyid, err := ps.generateSymmetricKey(topic, to, defaultSymKeySendLimit, true)
	if err != nil {
		b.Fatalf("could not generate symkey: %v", err)
	}
	symkey, err := ps.w.GetSymKey(symkeyid)
	if err != nil {
		b.Fatalf("could not retreive symkey: %v", err)
	}
	ps.SetSymmetricKey(symkey, topic, to, 0, false)

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
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)
	ps := NewTestPss(privkey, nil)
	msg := make([]byte, msgsize)
	rand.Read(msg)
	topic := whisper.BytesToTopic([]byte("foo"))
	to := network.RandomAddr().Over()
	ps.SetPeerPublicKey(&privkey.PublicKey, topic, to)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.SendAsym(common.ToHex(crypto.FromECDSAPub(&privkey.PublicKey)), topic, msg)
	}
}
func BenchmarkSymkeyBruteforceChangeaddr(b *testing.B) {
	for i := 100; i < 100000; i = i * 10 {
		for j := 32; j < 10000; j = j * 8 {
			b.Run(fmt.Sprintf("%d/%d", i, j), benchmarkSymkeyBruteforceChangeaddr)
		}
		//b.Run(fmt.Sprintf("%d", i), benchmarkSymkeyBruteforceChangeaddr)
	}
}

// decrypt performance using symkey cache, worst case
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
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)
	if cachesize > 0 {
		ps = NewTestPss(privkey, &PssParams{SymKeyCacheCapacity: int(cachesize)})
	} else {
		ps = NewTestPss(privkey, nil)
	}
	topic := whisper.BytesToTopic([]byte("foo"))
	for i := 0; i < int(keycount); i++ {
		to := network.RandomAddr().Over()
		keyid, err = ps.generateSymmetricKey(topic, to, defaultSymKeySendLimit, true)
		if err != nil {
			b.Fatalf("cant generate symkey #%d: %v", i, err)
		}
		symkey, err := ps.w.GetSymKey(keyid)
		if err != nil {
			b.Fatalf("could not retreive symkey %s: %v", keyid, err)
		}
		wparams := &whisper.MessageParams{
			TTL:      DefaultTTL,
			KeySym:   symkey,
			Topic:    topic,
			WorkTime: defaultWhisperWorkTime,
			PoW:      defaultWhisperPoW,
			Payload:  []byte("xyzzy"),
			Padding:  []byte("1234567890abcdef"),
		}
		woutmsg, err := whisper.NewSentMessage(wparams)
		if err != nil {
			b.Fatalf("could not create whisper message: %v", err)
		}
		env, err := woutmsg.Wrap(wparams)
		if err != nil {
			b.Fatalf("could not generate whisper envelope: %v", err)
		}
		ps.Register(&topic, func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
			return nil
		})
		pssmsgs = append(pssmsgs, &PssMsg{
			To:      to,
			Payload: env,
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ps.process(pssmsgs[len(pssmsgs)-(i%len(pssmsgs))-1])
		if err != nil {
			b.Fatalf("pss processing failed: %v", err)
		}
	}
}

func BenchmarkSymkeyBruteforceSameaddr(b *testing.B) {
	for i := 100; i < 100000; i = i * 10 {
		for j := 32; j < 10000; j = j * 8 {
			b.Run(fmt.Sprintf("%d/%d", i, j), benchmarkSymkeyBruteforceSameaddr)
		}
	}
}

// decrypt performance using symkey cache, best case
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
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)
	if cachesize > 0 {
		ps = NewTestPss(privkey, &PssParams{SymKeyCacheCapacity: int(cachesize)})
	} else {
		ps = NewTestPss(privkey, nil)
	}
	topic := whisper.BytesToTopic([]byte("foo"))
	for i := 0; i < int(keycount); i++ {
		addr[i] = network.RandomAddr().Over()
		keyid, err = ps.generateSymmetricKey(topic, addr[i], defaultSymKeySendLimit, true)
		if err != nil {
			b.Fatalf("cant generate symkey #%d: %v", i, err)
		}

	}
	symkey, err := ps.w.GetSymKey(keyid)
	if err != nil {
		b.Fatalf("could not retreive symkey %s: %v", keyid, err)
	}
	wparams := &whisper.MessageParams{
		TTL:      DefaultTTL,
		KeySym:   symkey,
		Topic:    topic,
		WorkTime: defaultWhisperWorkTime,
		PoW:      defaultWhisperPoW,
		Payload:  []byte("xyzzy"),
		Padding:  []byte("1234567890abcdef"),
	}
	woutmsg, err := whisper.NewSentMessage(wparams)
	if err != nil {
		b.Fatalf("could not create whisper message: %v", err)
	}
	env, err := woutmsg.Wrap(wparams)
	if err != nil {
		b.Fatalf("could not generate whisper envelope: %v", err)
	}
	ps.Register(&topic, func(msg []byte, p *p2p.Peer, asymmetric bool, keyid string) error {
		return nil
	})
	pssmsg := &PssMsg{
		To:      addr[len(addr)-1][:],
		Payload: env,
	}
	for i := 0; i < b.N; i++ {
		err := ps.process(pssmsg)
		if err != nil {
			b.Fatalf("pss processing failed: %v", err)
		}
	}
}

// setup simulated network and connect nodes in circle
func setupNetwork(numnodes int) (clients []*rpc.Client, err error) {
	nodes := make([]*simulations.Node, numnodes)
	clients = make([]*rpc.Client, numnodes)
	if numnodes < 2 {
		return nil, fmt.Errorf("Minimum two nodes in network")
	}
	adapter := adapters.NewSimAdapter(services)
	net := simulations.NewNetwork(adapter, &simulations.NetworkConfig{
		ID:             "0",
		DefaultService: "bzz",
	})
	for i := 0; i < numnodes; i++ {
		nodes[i], err = net.NewNodeWithConfig(&adapters.NodeConfig{
			Services: []string{"bzz", "pss"},
		})
		if err != nil {
			return nil, fmt.Errorf("error creating node 1: %v", err)
		}
		err = net.Start(nodes[i].ID())
		if err != nil {
			return nil, fmt.Errorf("error starting node 1: %v", err)
		}
		if i > 0 {
			err = net.Connect(nodes[i].ID(), nodes[i-1].ID())
			if err != nil {
				return nil, fmt.Errorf("error connecting nodes: %v", err)
			}
		}
		clients[i], err = nodes[i].Client()
		if err != nil {
			return nil, fmt.Errorf("create node 1 rpc client fail: %v", err)
		}
	}
	if numnodes > 2 {
		err = net.Connect(nodes[0].ID(), nodes[len(nodes)-1].ID())
		if err != nil {
			return nil, fmt.Errorf("error connecting first and last nodes")
		}
	}
	return clients, nil
}

func newServices() adapters.Services {
	stateStore := adapters.NewSimStateStore()
	kademlias := make(map[discover.NodeID]*network.Kademlia)
	kademlia := func(id discover.NodeID) *network.Kademlia {
		if k, ok := kademlias[id]; ok {
			return k
		}
		addr := network.NewAddrFromNodeID(id)
		params := network.NewKadParams()
		params.MinProxBinSize = 2
		params.MaxBinSize = 3
		params.MinBinSize = 1
		params.MaxRetries = 1000
		params.RetryExponent = 2
		params.RetryInterval = 1000000
		kademlias[id] = network.NewKademlia(addr.Over(), params)
		return kademlias[id]
	}
	return adapters.Services{
		"pss": func(ctx *adapters.ServiceContext) (node.Service, error) {
			cachedir, err := ioutil.TempDir("", "pss-cache")
			if err != nil {
				return nil, fmt.Errorf("create pss cache tmpdir failed", "error", err)
			}
			dpa, err := storage.NewLocalDPA(cachedir)
			if err != nil {
				return nil, fmt.Errorf("local dpa creation failed", "error", err)
			}

			keys, err := wapi.NewKeyPair()
			privkey, err := w.GetPrivateKey(keys)
			pssp := NewPssParams(privkey)
			pskad := kademlia(ctx.Config.ID)
			ps := NewPss(pskad, dpa, pssp)

			ping := &Ping{
				OutC: make(chan bool),
				Pong: true,
			}
			p2pp := NewPingProtocol(ping.OutC, ping.PingHandler)
			pp, err := RegisterPssProtocol(ps, &PingTopic, PingProtocol, p2pp, &PssProtocolOptions{Asymmetric: true})
			if err != nil {
				return nil, err
			}
			ps.Register(&PingTopic, pp.Handle)
			if err != nil {
				log.Error("Couldnt register pss protocol", "err", err)
				os.Exit(1)
			}
			pssprotocols[ctx.Config.ID.String()] = &protoCtrl{
				C:        ping.OutC,
				protocol: pp,
				run:      p2pp.Run,
			}
			return ps, nil
		},
		"bzz": func(ctx *adapters.ServiceContext) (node.Service, error) {
			addr := network.NewAddrFromNodeID(ctx.Config.ID)
			hp := network.NewHiveParams()
			hp.Discovery = false
			config := &network.BzzConfig{
				OverlayAddr:  addr.Over(),
				UnderlayAddr: addr.Under(),
				HiveParams:   hp,
			}
			return network.NewBzz(config, kademlia(ctx.Config.ID), stateStore), nil
		},
	}
}
