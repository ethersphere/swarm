package mru

import (
	"testing"
)

func getTestResourceUpdate() *resourceUpdate {
	return &resourceUpdate{
		updateHeader: *getTestUpdateHeader(),
		data:         []byte("El que lee mucho y anda mucho, ve mucho y sabe mucho"),
	}
}

func TestResourceUpdateSerializer(t *testing.T) {
	testBinarySerializerRecovery(t, getTestResourceUpdate(), "0x10dd205b00000000100e000000000000776f726c64206e657773207265706f72742c20657665727920686f7572000000876a8936a7cd0b79ef0735ad0896c1afe278781c4f000000da070000456c20717565206c6565206d7563686f207920616e6461206d7563686f2c207665206d7563686f20792073616265206d7563686f")
}

func TestResourceUpdateLengthCheck(t *testing.T) {
	testBinarySerializerLengthCheck(t, getTestResourceUpdate())
	// Test fail if update is too big
	update := getTestResourceUpdate()
	update.data = make([]byte, maxUpdateDataLength+100)
	serialized := make([]byte, update.binaryLength())
	if err := update.binaryPut(serialized); err == nil {
		t.Fatal("Expected resourceUpdate.binaryPut to fail since update is too big")
	}

	// test fail if data is empty or nil
	update.data = nil
	serialized = make([]byte, update.binaryLength())
	if err := update.binaryPut(serialized); err == nil {
		t.Fatal("Expected resourceUpdate.binaryPut to fail since data is empty")
	}
}
