package mru

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/mru/lookup"
)

func getTestUpdateLookup() *UpdateLookup {
	return &UpdateLookup{
		View:  *getTestResourceView(),
		Epoch: lookup.GetFirstEpoch(1000),
	}
}

func getTestLookupParams() *LookupParams {
	ul := getTestUpdateLookup()
	return &LookupParams{
		TimeLimit: 5000,
		View:      ul.View,
		Hint:      ul.Epoch,
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

// KV mocks a key value store
type KV map[string]string

func (kv KV) Get(key string) string {
	return kv[key]
}
func (kv KV) Set(key, value string) {
	kv[key] = value
}

func TestLookupParamsValues(t *testing.T) {
	var expected = KV{"hint.level": "25", "hint.time": "1000", "time": "5000", "topic": "0x776f726c64206e657773207265706f72742c20657665727920686f7572000000", "user": "0x876A8936A7Cd0b79Ef0735AD0896c1AFe278781c"}

	lp := getTestLookupParams()
	kv := make(KV)
	lp.ToValues(kv)

	if !reflect.DeepEqual(expected, kv) {
		expj, _ := json.Marshal(expected)
		gotj, _ := json.Marshal(kv)
		t.Fatalf("Expected kv to be %s, got %s", string(expj), string(gotj))
	}

	var lp2 LookupParams
	err := lp2.FromValues(kv, true)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(&lp2, lp) {
		t.Fatal("Expected recovered LookupParams to be the same")
	}

}
