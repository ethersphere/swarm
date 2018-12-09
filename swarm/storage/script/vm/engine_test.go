package txscript_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/swarm/storage/script/vm"
)

func TestEngine(t *testing.T) {

	spk := []byte{txscript.OP_2, txscript.OP_2, txscript.OP_EQUAL}
	ssig := []byte{}

	e, err := txscript.NewEngine(spk, ssig, txscript.ScriptFlags(0))
	if err != nil {
		t.Fatal(err)
	}

	err = e.Execute()
	if err != nil {
		t.Fatal(err)
	}

}
