package message_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/epiclabs-io/ut"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/pss/internal/message"
)

func TestFlags(tx *testing.T) {
	t := ut.BeginTest(tx, false)
	defer t.FinishTest()

	bools := []bool{true, false}
	for _, r := range bools {
		for _, s := range bools {
			f := message.Flags{
				Symmetric: s,
				Raw:       r,
			}
			// Test encoding:
			bytes, err := rlp.EncodeToBytes(&f)
			t.Ok(err)
			t.EqualsKey(fmt.Sprintf("r=%t; s=%t", r, s), hex.EncodeToString(bytes))

			// Test decoding:

			var f2 message.Flags
			err = rlp.DecodeBytes(bytes, &f2)
			t.Ok(err)
			t.Equals(f, f2)
		}
	}

}

func TestFlagsErrors(tx *testing.T) {
	t := ut.BeginTest(tx, false)
	defer t.FinishTest()

	var f2 message.Flags
	err := rlp.DecodeBytes([]byte{0x82, 0xFF, 0xFF}, &f2)
	t.MustFailWith(err, message.ErrIncorrectFlagsFieldLength)
}
