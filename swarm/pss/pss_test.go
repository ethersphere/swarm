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
	"github.com/ethereum/go-ethereum/pot"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/swarm/network"
	"github.com/ethereum/go-ethereum/swarm/storage"
	whisper "github.com/ethereum/go-ethereum/whisper/whisperv5"
)

const (
	pssServiceName = "pss"
	bzzServiceName = "bzz"
)

var (
	snapshotfile   string
	debugdebugflag = flag.Bool("vv", false, "veryverbose")
	debugflag      = flag.Bool("v", false, "verbose")
	w              *whisper.Whisper
	wapi           *whisper.PublicWhisperAPI

	// custom logging
	psslogmain log.Logger
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
}

func TestCache(t *testing.T) {
	var err error
	var potaddr pot.Address
	to, _ := hex.DecodeString("08090a0b0c0d0e0f1011121314150001020304050607161718191a1b1c1d1e1f")
	keys, err := wapi.NewKeyPair()
	privkey, err := w.GetPrivateKey(keys)

	ps := NewTestPss(privkey, nil)
	pp := NewPssParams(privkey)
	data := []byte("foo")
	datatwo := []byte("bar")
	fwdaddr := network.RandomAddr()
	copy(potaddr[:], fwdaddr.Over())
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

// tests:
// sets public key for peer
// set an outgoing symmetric key for peer
// generate own symmetric key for incoming message from peer
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
	inkeyid, err := ps.generateSymmetricKey(topic, addr, time.Second, true)
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
func TestKeysExchange(t *testing.T) {

	// set up two nodes directly connected
	// (we are not testing pss routing here)
	topic := whisper.BytesToTopic([]byte("foo:42"))
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	loaddr := make([]byte, 32)
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	roaddr := make([]byte, 32)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}

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
	var rkeysold []string

	for i := 0; i < 3; i++ {

		switch i {
		case 1:
			addrs[0] = loaddr[:8]
			addrs[1] = roaddr[:8]
			log.Info("Test partial address", "laddr", addrs[0], "raddr", addrs[1])
			break
		case 2:
			addrs[0] = []byte{}
			addrs[1] = []byte{}
			log.Info("Test empty address")
			break
		default:
			addrs[0] = loaddr
			addrs[1] = roaddr
			log.Info("Test full address", "laddr", addrs[0], "raddr", addrs[1])
		}

		err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, addrs[1])
		if err != nil {
			t.Fatal(err)
		}
		err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, addrs[0])
		if err != nil {
			t.Fatal(err)
		}

		// use api test method for generating and sending incoming symkey
		// the peer should save it, then generate and send back its own
		var symkeyid string
		err = clients[0].Call(&symkeyid, "pss_handshake", common.ToHex(rpubkey), hextopic, addrs[1])
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 2000) // wait for handshake complete, replace with sim expect logic

		// after the exchange, the key for receiving on node 1
		// should be the same as sending on node 2, and vice versa
		// check node 1 first
		lrecvkey := make([]byte, defaultSymKeyLength)
		err = clients[0].Call(&lrecvkey, "psstest_getSymKey", symkeyid)
		if err != nil {
			t.Fatal(err)
		}

		lsendkey := make([]byte, defaultSymKeyLength)
		err = clients[0].Call(&symkeyid, "pss_matchSymKey", symkeyid)
		if err != nil {
			t.Fatal(err)
		}
		err = clients[0].Call(&lsendkey, "psstest_getSymKey", symkeyid)

		// then check node 2
		var rkeys []string
		var rkeysnew []string
		var rkeysbytes [2][]byte
		err = clients[1].Call(&rkeys, "psstest_getSymKeys")
		if err != nil {
			t.Fatal(err)
		} else if len(rkeys) < i+1 {
			t.Fatalf("Wrong symkey count: Expected %d, got %d", i+1, len(rkeys))
		}
	nextkey:
		for i := 0; i < len(rkeys); i++ {
			for j := 0; j < len(rkeysold); j++ {
				if rkeys[i] == rkeysold[j] {
					continue nextkey
				}
			}
			rkeysnew = append(rkeysnew, rkeys[i])
			rkeysold = append(rkeysold, rkeys[i])
		}
		err = clients[1].Call(&rkeysbytes[0], "psstest_getSymKey", rkeysnew[0])
		if err != nil {
			t.Fatal(err)
		}
		err = clients[1].Call(&symkeyid, "pss_matchSymKey", rkeysnew[0])
		if err != nil {
			t.Fatal(err)
		}

		err = clients[1].Call(&rkeysbytes[1], "psstest_getSymKey", symkeyid)
		if err != nil {
			t.Fatal(err)
		}

		// check if they match
		// we do not know in which order they come from node 2 so we need to check both combinations
		if bytes.Equal(lrecvkey, rkeysbytes[1]) {
			if !bytes.Equal(lsendkey, rkeysbytes[0]) {
				t.Fatalf("Node 2 receive key does not match node 1 send key: %x != %x", rkeysbytes[0], lsendkey)
			}
		} else if bytes.Equal(lrecvkey, rkeysbytes[0]) {
			if !bytes.Equal(lsendkey, rkeysbytes[1]) {
				t.Fatalf("Node 2 receive key does not match node 1 send key: %x != %x", rkeysbytes[1], lsendkey)
			}
		} else {
			t.Fatalf("Node 1 receive key does not match any node 1 key: %x != %x | %x", lrecvkey, rkeysbytes[0], rkeysbytes[1])
		}
		t.Logf("#%d: left: %x / %x, right %x / %x", i, lrecvkey, lsendkey, rkeysbytes[0], rkeysbytes[1])
	}
}

// send symmetrically encrypted message between two directly connected peers
func TestSymSend(t *testing.T) {

	topic := whisper.BytesToTopic([]byte("foo:42"))
	hextopic := fmt.Sprintf("%x", topic)

	clients, err := setupNetwork(2)
	if err != nil {
		t.Fatal(err)
	}

	loaddr := make([]byte, 32)
	err = clients[0].Call(&loaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 1 baseaddr fail: %v", err)
	}
	roaddr := make([]byte, 32)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}

	time.Sleep(time.Millisecond * 500)

	var addrs [2][]byte

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
	for i := 0; i < 3; i++ {
		switch i {
		case 1:
			addrs[0] = loaddr[:8]
			addrs[1] = roaddr[:8]
			log.Info("Test partial address", "laddr", addrs[0], "raddr", addrs[1])
			break
		case 2:
			addrs[0] = []byte{}
			addrs[1] = []byte{}
			log.Info("Test empty address")
			break
		default:
			addrs[0] = loaddr
			addrs[1] = roaddr
			log.Info("Test full address", "laddr", addrs[0], "raddr", addrs[1])
		}

		lrecvkey := network.RandomAddr().Over()
		rrecvkey := network.RandomAddr().Over()
		var lkeyids [2]string
		var rkeyids [2]string

		err = clients[0].Call(&lkeyids, "psstest_setSymKeys", lrecvkey, rrecvkey, hextopic, roaddr)
		if err != nil {
			t.Fatal(err)
		}

		err = clients[1].Call(&rkeyids, "psstest_setSymKeys", rrecvkey, lrecvkey, hextopic, loaddr)
		if err != nil {
			t.Fatal(err)
		}

		lmsg := []byte("plugh")
		err = clients[1].Call(nil, "pss_sendSym", rkeyids[0], hextopic, lmsg)
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
		err = clients[0].Call(nil, "pss_sendSym", lkeyids[0], hextopic, rmsg)
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
}

// send asymmetrically encrypted message between two directly connected peers
func TestAsymSend(t *testing.T) {
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
	roaddr := make([]byte, 32)
	err = clients[1].Call(&roaddr, "pss_baseAddr")
	if err != nil {
		t.Fatalf("rpc get node 2 baseaddr fail: %v", err)
	}

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

	for i := 0; i < 3; i++ {
		switch i {
		case 1:
			addrs[0] = loaddr[:8]
			addrs[1] = roaddr[:8]
			log.Info("Test partial address", "laddr", addrs[0], "raddr", addrs[1])
			break
		case 2:
			addrs[0] = []byte{}
			addrs[1] = []byte{}
			log.Info("Test empty address")
			break
		default:
			addrs[0] = loaddr
			addrs[1] = roaddr
			log.Info("Test full address", "laddr", addrs[0], "raddr", addrs[1])
		}
		err = clients[0].Call(nil, "pss_setPeerPublicKey", rpubkey, hextopic, addrs[1])
		if err != nil {
			t.Fatal(err)
		}
		err = clients[1].Call(nil, "pss_setPeerPublicKey", lpubkey, hextopic, addrs[0])
		if err != nil {
			t.Fatal(err)
		}

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
}

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
	symkeyid, err := ps.generateSymmetricKey(topic, to, time.Second, true)
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
		keyid, err = ps.generateSymmetricKey(topic, to, time.Second, true)
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
		keyid, err = ps.generateSymmetricKey(topic, addr[i], time.Second, true)
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
				C: make(chan struct{}),
			}
			pp, err := RegisterPssProtocol(ps, &PingTopic, PingProtocol, NewPingProtocol(ping.PingHandler), 0x02)
			if err != nil {
				return nil, err
			}
			ps.Register(&PingTopic, pp.Handle)
			if err != nil {
				log.Error("Couldnt register pss protocol", "err", err)
				os.Exit(1)
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
