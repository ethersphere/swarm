package mru

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func getTestUpdateLookup() *UpdateLookup {
	rootAddr, _ := hexutil.Decode("0xDEADC0DEDEADC0DEDEADC0DEDEADC0DEDEADC0DEDEADC0DEDEADC0DEDEADC0DE")
	return &UpdateLookup{
		period:   79,
		version:  2010,
		rootAddr: rootAddr,
	}
}

func compareUpdateLookup(a, b *UpdateLookup) bool {
	return a.version == b.version &&
		a.period == b.period &&
		bytes.Equal(a.rootAddr, b.rootAddr)
}

func TestUpdateLookupUpdateAddr(t *testing.T) {
	ul := getTestUpdateLookup()
	updateAddr := ul.UpdateAddr()
	compareByteSliceToExpectedHex(t, "updateAddr", updateAddr, "0xd8aece4abcee948c3d2cc726966eb96c8d89a214c86a703bf9824e216f5089f3")
}

func TestUpdateLookupSerializer(t *testing.T) {
	serializedUpdateLookup := make([]byte, updateLookupLength)
	ul := getTestUpdateLookup()
	if err := ul.binaryPut(serializedUpdateLookup); err != nil {
		t.Fatal(err)
	}
	compareByteSliceToExpectedHex(t, "serializedUpdateLookup", serializedUpdateLookup, "0x4f000000da070000deadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0de")

	// set receiving slice to the wrong size
	serializedUpdateLookup = make([]byte, updateLookupLength+7)
	if err := ul.binaryPut(serializedUpdateLookup); err == nil {
		t.Fatalf("Expected UpdateLookup.binaryPut to fail when receiving slice has a length != %d", updateLookupLength)
	}

	// set rootAddr to an invalid length
	ul.rootAddr = []byte{1, 2, 3, 4}
	serializedUpdateLookup = make([]byte, updateLookupLength)
	if err := ul.binaryPut(serializedUpdateLookup); err == nil {
		t.Fatal("Expected UpdateLookup.binaryPut to fail when rootAddr is not of the correct size")
	}
}

func TestUpdateLookupDeserializer(t *testing.T) {
	serializedUpdateLookup, _ := hexutil.Decode("0x4f000000da070000deadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0dedeadc0de")
	var recoveredUpdateLookup UpdateLookup
	if err := recoveredUpdateLookup.binaryGet(serializedUpdateLookup); err != nil {
		t.Fatal(err)
	}
	originalUpdateLookup := *getTestUpdateLookup()
	if !compareUpdateLookup(&originalUpdateLookup, &recoveredUpdateLookup) {
		t.Fatalf("Expected recovered UpdateLookup to match")
	}

	// set source slice to the wrong size
	serializedUpdateLookup = make([]byte, updateLookupLength+4)
	if err := recoveredUpdateLookup.binaryGet(serializedUpdateLookup); err == nil {
		t.Fatalf("Expected UpdateLookup.binaryGet to fail when source slice has a length != %d", updateLookupLength)
	}
}

func TestUpdateLookupSerializeDeserialize(t *testing.T) {
	serializedUpdateLookup := make([]byte, updateLookupLength)
	originalUpdateLookup := getTestUpdateLookup()
	if err := originalUpdateLookup.binaryPut(serializedUpdateLookup); err != nil {
		t.Fatal(err)
	}
	var recoveredUpdateLookup UpdateLookup
	if err := recoveredUpdateLookup.binaryGet(serializedUpdateLookup); err != nil {
		t.Fatal(err)
	}
	if !compareUpdateLookup(originalUpdateLookup, &recoveredUpdateLookup) {
		t.Fatalf("Expected recovered UpdateLookup to match")
	}
}
