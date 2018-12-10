package vm_test

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

	sb := vm.NewScriptBuilder()
	sb.AddData(address[:])
	sb.AddOp(vm.OP_CHECKSIG)
	sb.EmbedData([]byte("some embedded data"))

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

	sb = vm.NewScriptBuilder()
	sb.AddData(sigBytes)
	ssig, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	e, err := vm.NewEngine(spk, ssig, payload, vm.ScriptFlags(0))
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
