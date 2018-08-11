package mru

import (
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

func getTestUpdateLookup() *UpdateLookup {
	return &UpdateLookup{
		View:  *getTestResourceView(),
		Epoch: lookup.GetFirstEpoch(1000),
	}
}

func TestUpdateLookupUpdateAddr(t *testing.T) {
	ul := getTestUpdateLookup()
	updateAddr := ul.UpdateAddr()
	compareByteSliceToExpectedHex(t, "updateAddr", updateAddr, "0x8b24583ec293e085f4c78aaee66d1bc5abfb8b4233304d14a349afa57af2a783")
}

func TestUpdateLookupSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestUpdateLookup(), "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781ce803000000000019")
}

func TestUpdateLookupLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestUpdateLookup())
}
