package message_test

import (
	"encoding/hex"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/pss/message"
)

type messageFixture struct {
	digest   string
	rlp      string
	stringer string
}

var messageFixtures = []messageFixture{{"4b34781cfa28a5ad653855567273675eabb8535461e57e4f4bfc81504d0a828d", "dd94fa12f92afbe00f8508d0e83bab9cf8cebf42e25e03808491273d4980", "PssMsg: Recipient: 0xfa12f92afbe00f8508d0e83bab9cf8cebf42e25e, Topic: 0x91273d49"},
	{"7f076bc036335b5d587d48c985d1b6ef8cd7015d6e484d0c7a72faddaa2aceaa", "e794210fc7bb818639ac48a4c6afa2f1581a8b9525e2000184ba78973d8aa84f7f80296fda3fd8df", "PssMsg: Recipient: 0x210fc7bb818639ac48a4c6afa2f1581a8b9525e2, Topic: 0xba78973d"},
	{"a3cb8298779bef44c33461f072c54391a39c09b7a726e55d60384d7484760559", "f194e2aadcd868ce028477f86e430140149b0300a9a5020284a6b46dd094f4b754a41bd4d5d11330e2924ff403c95bb84fa5", "PssMsg: Recipient: 0xe2aadcd868ce028477f86e430140149b0300a9a5, Topic: 0xa6b46dd0"},
	{"a82a894a753dffad41330dc1abbc85e5bc1791c393eba682eaf3cee56e6b0d9a", "f83b9460f9e0fa212bac5db82b22cee5272ee19a067256000384f013aa4b9e2fb3c9afcd593f3c5d3a96fecc1b7672562cc1b8828888269264bb976ed2", "PssMsg: Recipient: 0x60f9e0fa212bac5db82b22cee5272ee19a067256, Topic: 0xf013aa4b"},
	{"8ba6836253a10cf02e5031695ab39917e816b9677d53b4e4b2af5e439b05d362", "f845941dd4751f899d743d0780c9644375aae21132781803048426f57386a834dab59240ba3bcec68fd648a62ba94062413e5b5f89c0441b5809fff0a51dd1084e8f06fce30971", "PssMsg: Recipient: 0x1dd4751f899d743d0780c9644375aae211327818, Topic: 0x26f57386"},
}

func RandomArray(i, length int) []byte {
	source := rand.NewSource(int64(i))
	r := rand.New(source)
	b := make([]byte, length)
	for n := 0; n < length; n++ {
		b[n] = byte(r.Intn(256))
	}
	return b
}
func TestMessage(t *testing.T) {

	// generate some test messages deterministically
	for i, topicString := range someTopics {
		flags := message.Flags{
			Raw:       i&0x1 == 0,
			Symmetric: i&0x3 == 0,
		}

		msg := message.New(flags)
		msg.To = RandomArray(i, common.AddressLength)
		msg.Expire = uint32(i)
		msg.Topic = message.NewTopic([]byte(topicString))
		msg.Payload = RandomArray(i*9361, i*10)

		// test digest function:
		digest := msg.Digest()

		actual := hex.EncodeToString(digest[:])
		expected := messageFixtures[i].digest
		if expected != actual {
			t.Fatalf("Expected digest to be %s, got %s", expected, actual)
		}

		// test stringer:
		expected = messageFixtures[i].stringer
		actual = msg.String()
		if expected != actual {
			t.Fatalf("Expected stringer to return %s, got %s", expected, actual)
		}

		// Test RLP encoding:
		bytes, err := rlp.EncodeToBytes(&msg)
		if err != nil {
			t.Fatal(err)
		}

		expected = messageFixtures[i].rlp
		actual = hex.EncodeToString(bytes)
		if expected != actual {
			t.Fatalf("Expected RLP serialization to return %s, got %s", expected, actual)
		}

		// Test decoding:
		var msg2 message.Message
		err = rlp.DecodeBytes(bytes, &msg2)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(msg, &msg2) {
			t.Fatalf("Expected RLP decoding return %v, got %v", msg, &msg2)
		}
	}
}
