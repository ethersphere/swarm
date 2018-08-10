package mru

import (
	"testing"
)

func getTestUpdateHeader() *UpdateHeader {
	return &UpdateHeader{
		UpdateLookup: *getTestUpdateLookup(),
	}
}

func TestUpdateHeaderSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestUpdateHeader(), "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c190000000000000000")
}

func TestUpdateHeaderLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestUpdateHeader())
}
