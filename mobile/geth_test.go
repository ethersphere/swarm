package geth

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/log"
)

const (
	pssPassword = "foo"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlDebug, log.StdoutHandler))
}

func TestPssNode(t *testing.T) {
	dir, err := ioutil.TempDir("", "geth-mobile-node")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ks := keystore.NewKeyStore(filepath.Join(dir, "keystore"), keystore.LightScryptN, keystore.LightScryptP)
	a, err := ks.NewAccount(pssPassword)
	if err != nil {
		t.Fatal(err)
	}

	cfg := NewNodeConfig()
	cfg.PssEnabled = true
	cfg.PssPassword = pssPassword
	cfg.PssAccount = a.Address.Hex()

	n, err := NewNode(dir, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	n.Start()
	time.Sleep(time.Second * 10)
}
