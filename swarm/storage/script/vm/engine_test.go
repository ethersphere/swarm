package txscript_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/sha3"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestEngine(t *testing.T) {
	privKey, _ := crypto.HexToECDSA("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	address := crypto.PubkeyToAddress(privKey.PublicKey)

	sb := txscript.NewScriptBuilder()
	sb.AddData(address[:])
	sb.AddOp(txscript.OP_CHECKSIG)
	sb.AddData([]byte("some stuff"))
	sb.AddData([]byte("and more"))
	spk, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("PAYLOAD")
	hasher := sha3.NewKeccak256()
	hasher.Write(spk)
	hasher.Write(payload)
	digest := hasher.Sum(nil)
	sigBytes, err := crypto.Sign(digest, privKey)
	if err != nil {
		t.Fatal(err)
	}

	sb = txscript.NewScriptBuilder()
	sb.AddData(sigBytes)
	ssig, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	e, err := txscript.NewEngine(spk, ssig, payload, txscript.ScriptFlags(0))
	if err != nil {
		t.Fatal(err)
	}

	dis, _ := e.DisasmScript(1)
	fmt.Println(dis)
	dis, _ = e.DisasmScript(0)
	fmt.Println(dis)

	err = e.Execute()
	if err != nil {
		t.Fatal(err)
	}

}
