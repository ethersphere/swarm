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
	compareByteSliceToExpectedHex(t, "updateAddr", updateAddr, "0x5c66293944d0c8934ef161dda86f1b356301812ad49dcbcb086cadf79e4f745b")
}

func TestUpdateLookupSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestUpdateLookup(), "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c190000000000000000")
}

func TestUpdateLookupLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestUpdateLookup())
}
