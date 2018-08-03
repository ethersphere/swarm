package mru

import (
	"testing"
)

func getTestUpdateLookup() *UpdateLookup {
	return &UpdateLookup{
		view:    *getTestResourceView(),
		period:  79,
		version: 2010,
	}
}

func TestUpdateLookupUpdateAddr(t *testing.T) {
	ul := getTestUpdateLookup()
	updateAddr := ul.UpdateAddr()
	compareByteSliceToExpectedHex(t, "updateAddr", updateAddr, "0x0dbff62eef0075fc18a3b3b193cde32ac720f16652685dd86f92dfdd941ccb63")
}

func TestUpdateLookupSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestUpdateLookup(), "0x10dd205b00000000100e000000000000776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c4f000000da070000")
}

func TestUpdateLookupLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestUpdateLookup())
}
