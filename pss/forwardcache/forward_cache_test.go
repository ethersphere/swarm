package forwardcache

import (
	"context"
	"encoding/hex"
	"github.com/ethersphere/swarm/pss/crypto"
	"github.com/ethersphere/swarm/pss/message"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	var err error
	to, _ := hex.DecodeString("08090a0b0c0d0e0f1011121314150001020304050607161718191a1b1c1d1e1f")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cryptoBackend := crypto.New()
	keys, err := cryptoBackend.NewKeyPair(ctx)
	privkey, err := cryptoBackend.GetPrivateKey(keys)
	if err != nil {
		t.Fatal(err)
	}

	testCache := NewMockForwardCache(nil)

	data := []byte("foo")
	datatwo := []byte("bar")
	datathree := []byte("baz")
	wparams := &crypto.WrapParams{
		Src: privkey,
		Dst: &privkey.PublicKey,
	}
	env, err := cryptoBackend.WrapMessage(data, wparams)
	msg := &message.Message{
		Payload: env,
		To:      to,
		Topic:   message.Topic{},
	}
	envtwo, err := cryptoBackend.WrapMessage(datatwo, wparams)
	msgtwo := &message.Message{
		Payload: envtwo,
		To:      to,
		Topic:   message.Topic{},
	}
	envthree, err := cryptoBackend.WrapMessage(datathree, wparams)
	msgthree := &message.Message{
		Payload: envthree,
		To:      to,
		Topic:   message.Topic{},
	}

	digestone := msg.Digest()
	digesttwo := msgtwo.Digest()
	digestthree := msgthree.Digest()

	if digestone == digesttwo {
		t.Fatalf("different msgs return same hash: %d", digesttwo)
	}

	// check the cache
	err = testCache.AddFwdCache(msg)
	if err != nil {
		t.Fatalf("write to pss expire cache failed: %v", err)
	}

	if !testCache.CheckFwdCache(msg) {
		t.Fatalf("message %v should have EXPIRE record in cache but checkCache returned false", msg)
	}

	if testCache.CheckFwdCache(msgtwo) {
		t.Fatalf("message %v should NOT have EXPIRE record in cache but checkCache returned true", msgtwo)
	}

	time.Sleep(testCache.CacheTTL + 1*time.Second)
	err = testCache.AddFwdCache(msgthree)
	if err != nil {
		t.Fatalf("write to pss expire cache failed: %v", err)
	}

	if testCache.CheckFwdCache(msg) {
		t.Fatalf("message %v should have expired from cache but checkCache returned true", msg)
	}

	if _, ok := testCache.fwdCache[digestthree]; !ok {
		t.Fatalf("unexpired message should be in the cache: %v", digestthree)
	}

	if _, ok := testCache.fwdCache[digesttwo]; ok {
		t.Fatalf("expired message should have been cleared from the cache: %v", digesttwo)
	}
}
