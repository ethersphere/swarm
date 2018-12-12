// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package vm_test

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/storage/feed"
	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

// Signer generates signers with a deterministic private key for tests
func Signer(i int) feed.Signer {
	var bytes [32]byte
	binary.LittleEndian.PutUint64(bytes[:], uint64(i+1979))
	privKey, _ := crypto.ToECDSA(bytes[:])
	return feed.NewGenericSigner(privKey)
}

func TestEngineSig(t *testing.T) {

	// get a test signer:
	alice := Signer(1)

	// build a simple signing script. It will look like this:
	// DATA_20 <alice's address) CHECKSIG

	sb := vm.NewScriptBuilder()
	sb.AddData(alice.Address().Bytes())                   // add Alice's Ethereum address
	sb.AddOp(vm.OP_CHECKSIG)                              // CHECKSIG opcode
	sb.EmbedData([]byte("some embedded data in the key")) // this is some optional key metadata

	// retrieve the script out of the builder as a byte array
	spk, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("PAYLOAD") // some payload for our message/chunk

	// prepare script for signature. This removes certain opcodes
	preparedScript, err := vm.PrepareScriptForSig(spk)
	if err != nil {
		t.Fatal(err)
	}

	// calculate digest to sign:
	digest := common.BytesToHash(vm.CalcSignatureHash(nil, preparedScript, payload))

	// actually sign it and obtain a signature
	aliceSignature, err := alice.Sign(digest)

	// also generate a signature for the same content by an unauthorized user:
	rogueSignature, err := Signer(666).Sign(digest)
	if err != nil {
		t.Fatal(err)
	}

	signTest := func(expectedErrorCode vm.ErrorCode, signature feed.Signature) {

		// build the signature script. It will simply contain the 65-byte signature.
		sb := vm.NewScriptBuilder()

		sb.AddData(signature[:]) // add the signature

		ssig, err := sb.Script()
		if err != nil {
			t.Fatal(err)
		}

		e, err := vm.NewEngine(spk, ssig, payload, vm.ScriptFlags(0))
		if err != nil {
			t.Fatal(err)
		}

		err = e.Execute()
		if err != nil {
			if scriptError, ok := err.(vm.Error); ok {
				if scriptError.ErrorCode == expectedErrorCode {
					// this was expected, so no error.
					return
				} else {
					t.Fatalf("Script error %d: %s ", scriptError.ErrorCode, scriptError.Error())
				}
			} else {
				if expectedErrorCode == -1 {
					// expected error of other type, so ok.
					return
				}
			}
			t.Fatal(err)
		}
		if expectedErrorCode != 0 {
			t.Fatalf("Expected failure with error code %d, got nil error", expectedErrorCode)
		}
	}

	// should work with Alice's signature
	signTest(0, aliceSignature)

	// should fail with rogue signature
	signTest(vm.ErrEvalFalse, rogueSignature)

	// should fail with an invalid signature
	signTest(-1, feed.Signature{})

}

func TestEngineMultiSig(t *testing.T) {
	// test m of n multisig
	const numSignatures = 3 // signatures required
	const numSigners = 5    // how many people can sign

	// Build a 3 of 5 multisig script. It will look like this:
	// 3 DATA_20 <addr0> DATA_20 <addr1> DATA_20 <addr2> DATA_20 <addr3> DATA_20 <addr4> 5 CHECKMULTISIG

	sb := vm.NewScriptBuilder()

	sb.AddInt64(numSignatures)

	var signers []feed.Signer
	for i := 0; i < numSigners; i++ {
		signer := Signer(i)
		signers = append(signers, signer)
		sb.AddData(signer.Address().Bytes())
	}

	sb.AddInt64(numSigners)

	sb.AddOp(vm.OP_CHECKMULTISIG)
	sb.EmbedData([]byte("some embedded data in the key"))

	spk, err := sb.Script()
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte("PAYLOAD")
	preparedScript, err := vm.PrepareScriptForSig(spk)
	if err != nil {
		t.Fatal(err)
	}
	digest := common.BytesToHash(vm.CalcSignatureHash(nil, preparedScript, payload))

	rogueSignature, err := Signer(666).Sign(digest)
	if err != nil {
		t.Fatal(err)
	}

	var sigs [][]byte

	for i := 0; i < numSigners; i++ {
		signature, err := signers[i].Sign(digest)
		if err != nil {
			t.Fatal(err)
		}
		sigs = append(sigs, signature[:])
	}

	signTest := func(expectedErrorCode vm.ErrorCode, signatures ...[]byte) {
		sb := vm.NewScriptBuilder()

		for _, signature := range signatures {
			sb.AddData(signature[:])
		}

		ssig, err := sb.Script()
		if err != nil {
			t.Fatal(err)
		}

		e, err := vm.NewEngine(spk, ssig, payload, vm.ScriptFlags(0))
		if err != nil {
			t.Fatal(err)
		}

		err = e.Execute()
		if err != nil {
			if scriptError, ok := err.(vm.Error); ok {
				if scriptError.ErrorCode == expectedErrorCode {
					// this was expected, so no error.
					return
				} else {
					t.Fatalf("Script error %d: %s ", scriptError.ErrorCode, scriptError.Error())
				}
			} else {
				if expectedErrorCode == -1 {
					// expected error of other type, so ok.
					return
				}
			}
			t.Fatal(err)
		}
		if expectedErrorCode != 0 {
			t.Fatalf("Expected failure with error code %d, got nil error", expectedErrorCode)
		}
	}

	// Test it works with 3 signatures if they are in the right order:
	signTest(0, sigs[1], sigs[2], sigs[3])
	signTest(0, sigs[0], sigs[1], sigs[2])
	signTest(0, sigs[1], sigs[3], sigs[4])
	signTest(0, sigs[0], sigs[3], sigs[4])

	// In the wrong order does not work:
	signTest(vm.ErrEvalFalse, sigs[3], sigs[1], sigs[2])

	// Some wise guy signing three times:
	signTest(vm.ErrEvalFalse, sigs[4], sigs[4], sigs[4])

	// Two are good, one is not even a signature
	signTest(-1, sigs[1], sigs[2], make([]byte, 65)) // expect a signature recovery error

	// Two are good, one is a valid signature of someone not authorized in the key
	signTest(vm.ErrEvalFalse, sigs[1], sigs[2], rogueSignature[:])

	// Not enough signatures:
	signTest(vm.ErrInvalidStackOperation, sigs[3])

	// No signatures at all:
	signTest(vm.ErrInvalidStackOperation, nil)

	// Excess signatures:
	signTest(0, sigs...)

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

	err = e.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func TestScriptMarshallingUnmarshalling(t *testing.T) {
	st := `{
		"script": "OP_DATA_4 0x1effffff OP_CHECKPOW OP_EMBED 0x12 0x736f6d6520656d6265646465642064617461"
	}`

	var script vm.Script
	err := json.Unmarshal([]byte(st), &script)
	if err != nil {
		t.Fatal(err)
	}
	expectedString := `DATA_4 0x1effffff CHECKPOW EMBED 0x12 0x736f6d6520656d6265646465642064617461`
	if script.String() != expectedString {
		t.Fatalf("Expected %s, got %s", expectedString, script.String())
	}

	st = `{
		"binary": "0x041efffffff8f912736f6d6520656d6265646465642064617461"
	}`

	err = json.Unmarshal([]byte(st), &script)
	if err != nil {
		t.Fatal(err)
	}
	if script.String() != expectedString {
		t.Fatalf("Expected %s, got %s", expectedString, script.String())
	}
}
