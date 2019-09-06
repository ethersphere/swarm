package message_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/epiclabs-io/ut"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/pss/message"
)

func TestMessage(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()

	// generate some test messages deterministically
	for i, topicString := range someTopics {
		flags := message.Flags{
			Raw:       i&0x1 == 0,
			Symmetric: i&0x3 == 0,
		}

		msg := message.New(flags)
		msg.To = ut.RandomArray(i, common.AddressLength)
		msg.Expire = uint32(i)
		msg.Topic = message.NewTopic([]byte(topicString))
		msg.Payload = ut.RandomArray(i*9361, i*10)

		// test digest function:
		digest := msg.Digest()
		t.EqualsKey(fmt.Sprintf("msg%d-digest", i), hex.EncodeToString(digest[:]))

		// test stringer:
		st := msg.String()
		t.EqualsKey(fmt.Sprintf("msg%d-string", i), st)

		// Test RLP encoding:
		bytes, err := rlp.EncodeToBytes(&msg)
		t.Ok(err)
		t.EqualsKey(fmt.Sprintf("msg%d-rlp", i), hex.EncodeToString(bytes))

		// Test decoding:
		var msg2 message.Message
		err = rlp.DecodeBytes(bytes, &msg2)
		t.Ok(err)
		t.Equals(msg, &msg2)

	}
}
