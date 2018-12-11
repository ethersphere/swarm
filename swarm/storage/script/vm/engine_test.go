package vm_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/ethereum/go-ethereum/crypto/sha3"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestEngineSig(t *testing.T) {
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

	s := vm.Script(spk)
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	err = e.Execute()
	if err != nil {
		t.Fatal(err)
	}

}

func TestEnginePow(t *testing.T) {

	targetCompact := []byte{30, 0xFF, 0xFF, 0xFF}

	sb := vm.NewScriptBuilder()
	sb.AddData(targetCompact)
	sb.AddOp(vm.OP_CHECKPOW)
	sb.EmbedData([]byte("some embedded data"))

	scriptKey, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("PAYLOAD")

	nonce := make([]byte, 8)
	n := (*uint64)(unsafe.Pointer(&nonce[0]))

	prepared, err := vm.PrepareScriptForSig(scriptKey)
	if err != nil {
		t.Fatal(err)
	}

	target, err := vm.Compact2Target(targetCompact)
	fmt.Println("target", target)
	if err != nil {
		t.Fatal(err)
	}

	for {
		hash := vm.CalcSignatureHash(nonce, prepared, payload)
		if vm.VerifyTarget(target, hash) {
			fmt.Println(hash)
			break
		}
		*n++
	}

	fmt.Println(nonce)

	sb = vm.NewScriptBuilder()
	sb.AddData(nonce)
	ssig, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	e, err := vm.NewEngine(scriptKey, ssig, payload, vm.ScriptFlags(0))
	if err != nil {
		t.Fatal(err)
	}

	dis, _ := e.DisasmScript(1)
	fmt.Println(dis)
	dis, _ = e.DisasmScript(0)
	fmt.Println(dis)

	s := vm.Script(scriptKey)
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	err = e.Execute()
	if err != nil {
		t.Fatal(err)
	}
}
