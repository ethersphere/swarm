package message_test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethersphere/swarm/pss/message"
)

var bools = []bool{true, false}
var flagsFixture = map[string]string{
	"r=false; s=false": "00",
	"r=false; s=true":  "01",
	"r=true; s=false":  "02",
	"r=true; s=true":   "03",
}

func TestFlags(t *testing.T) {

	for _, r := range bools {
		for _, s := range bools {
			f := message.Flags{
				Symmetric: s,
				Raw:       r,
			}
			// Test encoding:
			bytes, err := rlp.EncodeToBytes(&f)
			if err != nil {
				t.Fatal(err)
			}
			expected := flagsFixture[fmt.Sprintf("r=%t; s=%t", r, s)]
			actual := hex.EncodeToString(bytes)
			if expected != actual {
				t.Fatalf("Expected RLP encoding of the flags to be %s, got %s", expected, actual)
			}

			// Test decoding:

			var f2 message.Flags
			err = rlp.DecodeBytes(bytes, &f2)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(f, f2) {
				t.Fatalf("Expected RLP decoding to return the same object. Got %v", f2)
			}
		}
	}

}

func TestFlagsErrors(t *testing.T) {
	var f2 message.Flags
	err := rlp.DecodeBytes([]byte{0x82, 0xFF, 0xFF}, &f2)
	if err != message.ErrIncorrectFlagsFieldLength {
		t.Fatalf("Expected an message.ErrIncorrectFlagsFieldLength error. Got %v", err)
	}
}
